package update

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const product = "dev.curb.app"
const markerName = ".curb-install.json"

var ErrUnpackaged = errors.New("self-updates require the Linux AppImage, a macOS app bundle, or a Curb portable Windows release")

type Installation struct {
	Kind       string `json:"kind"`
	Target     string `json:"target"`
	Executable string `json:"executable"`
	Platform   string `json:"platform"`
}

type installMarker struct {
	Product  string   `json:"product"`
	Version  string   `json:"version"`
	Platform string   `json:"platform"`
	Files    []string `json:"files"`
}

// DetectInstallation never treats a development directory or Downloads folder
// as a replaceable application directory.
func DetectInstallation() (Installation, error) {
	executable, err := os.Executable()
	if err != nil {
		return Installation{}, err
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	platform := runtime.GOOS + "-" + runtime.GOARCH
	switch runtime.GOOS {
	case "linux":
		image := os.Getenv("APPIMAGE")
		if image == "" || !filepath.IsAbs(image) {
			return Installation{}, ErrUnpackaged
		}
		if resolved, err := filepath.EvalSymlinks(image); err == nil {
			image = resolved
		}
		if err := validateAppImage(image, runtime.GOARCH); err != nil {
			return Installation{}, err
		}
		return Installation{Kind: "appimage", Target: image, Platform: platform}, nil
	case "darwin":
		bundle := filepath.Dir(filepath.Dir(filepath.Dir(executable)))
		if !strings.HasSuffix(bundle, ".app") || filepath.Base(filepath.Dir(executable)) != "MacOS" {
			return Installation{}, ErrUnpackaged
		}
		relative, err := filepath.Rel(bundle, executable)
		if err != nil {
			return Installation{}, err
		}
		return Installation{Kind: "bundle", Target: bundle, Executable: relative, Platform: platform}, nil
	case "windows":
		root := filepath.Dir(executable)
		if _, err := readMarker(root, platform, ""); err != nil {
			return Installation{}, ErrUnpackaged
		}
		return Installation{Kind: "directory", Target: root, Executable: filepath.Base(executable), Platform: platform}, nil
	}
	return Installation{}, ErrUnpackaged
}

// Plan lives in a private staging directory next to the installation. The
// helper applies it only after the original process exits.
type Plan struct {
	Installation
	Work      string   `json:"work"`
	Staged    string   `json:"staged"`
	Version   string   `json:"version"`
	Token     string   `json:"token"`
	ParentPID int      `json:"parent_pid"`
	Arguments []string `json:"arguments,omitempty"`
	committed bool
}

// Prepare downloads and verifies a release without touching the current app.
func (c *Client) Prepare(ctx context.Context, release Release, installation Installation, progress func(int64, int64)) (_ *Plan, result error) {
	if err := release.validate(); err != nil {
		return nil, err
	}
	if release.Platform != installation.Platform && !(release.Platform == "darwin-universal" && strings.HasPrefix(installation.Platform, "darwin-")) {
		return nil, errors.New("update is for another platform")
	}
	work, err := os.MkdirTemp(filepath.Dir(installation.Target), ".curb-update-")
	if err != nil {
		return nil, fmt.Errorf("move Curb to a writable folder to update: %w", err)
	}
	defer func() {
		if result != nil {
			os.RemoveAll(work)
		}
	}()
	archive := filepath.Join(work, "download")
	if err := c.download(ctx, release, archive, progress); err != nil {
		return nil, err
	}
	staged := archive
	switch installation.Kind {
	case "appimage":
		if release.Asset.Format != "appimage" {
			return nil, errors.New("expected an AppImage update")
		}
		if err := validateAppImage(archive, strings.TrimPrefix(installation.Platform, "linux-")); err != nil {
			return nil, err
		}
		if err := os.Chmod(archive, 0o755); err != nil {
			return nil, err
		}
	case "directory", "bundle":
		unpacked := filepath.Join(work, "unpacked")
		if err := extractArchive(ctx, archive, unpacked, release.Asset.Format); err != nil {
			return nil, err
		}
		entries, err := os.ReadDir(unpacked)
		if err != nil || len(entries) != 1 || !entries[0].IsDir() {
			return nil, errors.New("update must contain exactly one app folder")
		}
		staged = filepath.Join(unpacked, entries[0].Name())
		if installation.Kind == "bundle" {
			if !strings.HasSuffix(staged, ".app") {
				return nil, errors.New("update contains no macOS app bundle")
			}
			if err := validateBundle(staged, release.Version); err != nil {
				return nil, err
			}
		} else if _, err := readMarker(staged, release.Platform, release.Version); err != nil {
			return nil, err
		}
		info, err := os.Stat(filepath.Join(staged, installation.Executable))
		if err != nil || !info.Mode().IsRegular() {
			return nil, errors.New("updated app executable is missing")
		}
		os.Remove(archive)
	default:
		return nil, ErrUnpackaged
	}
	var token [24]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, err
	}
	return &Plan{Installation: installation, Work: work, Staged: staged, Version: release.Version,
		Token: hex.EncodeToString(token[:]), ParentPID: os.Getpid(), Arguments: os.Args[1:]}, nil
}

func (c *Client) download(ctx context.Context, release Release, destination string, progress func(int64, int64)) error {
	response, err := c.get(ctx, release.URL())
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.ContentLength >= 0 && response.ContentLength != release.Asset.Size {
		return errors.New("release size differs from the update manifest")
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	sum := sha256.New()
	writer := &progressWriter{writer: io.MultiWriter(file, sum), total: release.Asset.Size, notify: progress}
	n, err := io.Copy(writer, io.LimitReader(response.Body, release.Asset.Size+1))
	if err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if n != release.Asset.Size {
		return errors.New("update download was incomplete")
	}
	if !strings.EqualFold(hex.EncodeToString(sum.Sum(nil)), release.Asset.SHA256) {
		return errors.New("update checksum verification failed")
	}
	if progress != nil {
		progress(n, n)
	}
	return nil
}

type progressWriter struct {
	writer            io.Writer
	total, downloaded int64
	last              time.Time
	notify            func(int64, int64)
}

func (w *progressWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	w.downloaded += int64(n)
	if w.notify != nil && time.Since(w.last) >= 100*time.Millisecond {
		w.last = time.Now()
		w.notify(w.downloaded, w.total)
	}
	return n, err
}

func validateAppImage(path, architecture string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var header [20]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		return errors.New("invalid AppImage header")
	}
	machine := uint16(header[18]) | uint16(header[19])<<8
	if string(header[:4]) != "\x7fELF" || string(header[8:11]) != "AI\x02" || header[5] != 1 ||
		!(architecture == "amd64" && machine == 62 || architecture == "arm64" && machine == 183) {
		return errors.New("download is not an AppImage for this architecture")
	}
	return nil
}

func readMarker(root, platform, version string) (*installMarker, error) {
	file, err := os.Open(filepath.Join(root, markerName))
	if err != nil {
		return nil, errors.New("not a managed Curb release")
	}
	defer file.Close()
	var marker installMarker
	err = json.NewDecoder(io.LimitReader(file, 256<<10)).Decode(&marker)
	if err != nil || marker.Product != product || marker.Platform != platform || version != "" && marker.Version != version {
		return nil, errors.New("release identity does not match this update")
	}
	// A marker does not grant ownership of unrelated files. Refuse a folder
	// with additions instead of deleting personal files during replacement.
	owned := make(map[string]bool)
	for _, name := range marker.Files {
		if _, err := archivePath(root, name); err != nil {
			return nil, err
		}
		owned[strings.ToLower(name)] = true
	}
	if len(owned) == 0 || !owned[markerName] {
		return nil, errors.New("release has no managed-file inventory")
	}
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if !owned[strings.ToLower(filepath.ToSlash(relative))] {
			return errors.New("extra files are present in the app folder; move them out before updating")
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("managed Windows releases cannot contain symlinks")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &marker, nil
}

func validateBundle(root, version string) error {
	file, err := os.Open(filepath.Join(root, "Contents", "Info.plist"))
	if err != nil {
		return errors.New("macOS app metadata is missing")
	}
	defer file.Close()
	decoder := xml.NewDecoder(io.LimitReader(file, 64<<10))
	values := map[string]string{}
	key := ""
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if start, ok := token.(xml.StartElement); ok && (start.Name.Local == "key" || start.Name.Local == "string") {
			var value string
			if err := decoder.DecodeElement(&value, &start); err != nil {
				return err
			}
			if start.Name.Local == "key" {
				key = value
			} else if key != "" {
				values[key], key = value, ""
			}
		}
	}
	if values["CFBundleIdentifier"] != product || values["CFBundleShortVersionString"] != version {
		return errors.New("macOS bundle identity or version does not match this update")
	}
	return nil
}
