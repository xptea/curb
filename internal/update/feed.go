// Package update checks Curb's release feed and installs verified desktop releases.
package update

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	FeedURL            = "https://raw.githubusercontent.com/xptea/curb/main/version.json"
	ReleaseURL         = "https://github.com/xptea/curb/releases/download/"
	maxAssetSize int64 = 1 << 30
)

type Asset struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Format string `json:"format"`
}

type Manifest struct {
	Schema    int              `json:"schema"`
	Version   string           `json:"version"`
	Platforms map[string]Asset `json:"platforms"`
}

type Release struct {
	Version  string
	Platform string
	Asset    Asset
}

func (r Release) URL() string {
	return ReleaseURL + "v" + r.Version + "/" + url.PathEscape(r.Asset.Name)
}

type Client struct {
	Current  string
	Platform string
	HTTP     *http.Client
}

func NewClient(version string) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 20 * time.Second
	return &Client{Current: version, Platform: runtime.GOOS + "-" + runtime.GOARCH,
		HTTP: &http.Client{Transport: transport, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many update redirects")
			}
			host := req.URL.Hostname()
			if req.URL.Scheme != "https" || !(host == "github.com" || host == "raw.githubusercontent.com" || strings.HasSuffix(host, ".githubusercontent.com")) {
				return errors.New("update redirected outside GitHub")
			}
			return nil
		}}}
}

func (c *Client) get(ctx context.Context, address string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Curb/"+c.Current)
	req.Header.Set("Cache-Control", "no-cache")
	response, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("GitHub returned HTTP %d", response.StatusCode)
	}
	return response, nil
}

// Check returns only a newer stable release that has an asset for this machine.
// No asset is downloaded until the user clicks the update button.
func (c *Client) Check(ctx context.Context) (*Release, error) {
	response, err := c.get(ctx, FeedURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, (64<<10)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 64<<10 {
		return nil, errors.New("update manifest is too large")
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid update manifest: %w", err)
	}
	if manifest.Schema != 1 {
		return nil, errors.New("unsupported update manifest")
	}
	newer, err := newerVersion(manifest.Version, c.Current)
	if err != nil || !newer {
		return nil, err
	}
	platform := c.Platform
	asset, ok := manifest.Platforms[platform]
	if !ok && strings.HasPrefix(platform, "darwin-") {
		platform = "darwin-universal"
		asset, ok = manifest.Platforms[platform]
	}
	if !ok {
		return nil, nil
	}
	release := &Release{Version: manifest.Version, Platform: platform, Asset: asset}
	if err := release.validate(); err != nil {
		return nil, err
	}
	return release, nil
}

func (r Release) validate() error {
	if _, err := parseVersion(r.Version); err != nil {
		return err
	}
	if r.Asset.Name == "" || r.Asset.Name != path.Base(r.Asset.Name) || strings.ContainsAny(r.Asset.Name, "\\?#%\x00\r\n") {
		return errors.New("invalid release filename")
	}
	sum, err := hex.DecodeString(r.Asset.SHA256)
	if err != nil || len(sum) != 32 {
		return errors.New("invalid update checksum")
	}
	if r.Asset.Size <= 0 || r.Asset.Size > maxAssetSize {
		return errors.New("invalid update size")
	}
	switch {
	case strings.HasPrefix(r.Platform, "linux-") && r.Asset.Format == "appimage" && strings.HasSuffix(r.Asset.Name, ".AppImage"):
	case strings.HasPrefix(r.Platform, "windows-") && r.Asset.Format == "zip" && strings.HasSuffix(r.Asset.Name, ".zip"):
	case strings.HasPrefix(r.Platform, "darwin-") && r.Asset.Format == "tar.gz" && strings.HasSuffix(r.Asset.Name, ".tar.gz"):
	default:
		return errors.New("release format does not match this platform")
	}
	return nil
}

func parseVersion(version string) ([3]uint64, error) {
	var result [3]uint64
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return result, errors.New("expected a stable version such as 0.1.0")
	}
	for i, part := range parts {
		if part == "" || len(part) > 1 && part[0] == '0' {
			return result, errors.New("invalid version number")
		}
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return result, errors.New("invalid version number")
			}
		}
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return result, errors.New("invalid version number")
		}
		result[i] = n
	}
	return result, nil
}

func newerVersion(candidate, current string) (bool, error) {
	a, err := parseVersion(candidate)
	if err != nil {
		return false, err
	}
	b, err := parseVersion(current)
	if err != nil {
		return false, err
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i], nil
		}
	}
	return false, nil
}
