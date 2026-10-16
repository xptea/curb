package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Extraction never follows archive symlinks, accepts absolute paths, or writes
// outside staging. Limits also bound disk use from compressed archives.
func archivePath(root, name string) (string, error) {
	name = strings.TrimSuffix(name, "/")
	if name == "" || path.Clean(name) != name || !filepath.IsLocal(filepath.FromSlash(name)) || strings.ContainsAny(name, "\\:\x00") {
		return "", errors.New("unsafe path in update archive")
	}
	return filepath.Join(root, filepath.FromSlash(name)), nil
}

func extractArchive(ctx context.Context, source, root, format string) error {
	if err := os.Mkdir(root, 0o700); err != nil {
		return err
	}
	var expanded int64
	count := 0
	write := func(name string, mode os.FileMode, size int64, data io.Reader) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		count++
		expanded += size
		if size < 0 || count > 20000 || expanded > 2<<30 {
			return errors.New("update archive exceeds extraction limits")
		}
		destination, err := archivePath(root, name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		permission := os.FileMode(0o644)
		if mode&0o111 != 0 {
			permission = 0o755
		}
		file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, permission)
		if err != nil {
			return err
		}
		n, err := io.CopyN(file, data, size)
		if err == nil {
			var extra [1]byte
			if read, endErr := data.Read(extra[:]); read != 0 || endErr != io.EOF {
				err = errors.New("invalid archive file length or checksum")
			}
		}
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil || n != size {
			return errors.New("incomplete file in update archive")
		}
		return nil
	}
	if format == "zip" {
		archive, err := zip.OpenReader(source)
		if err != nil {
			return err
		}
		defer archive.Close()
		if len(archive.File) > 20000 {
			return errors.New("too many files in update archive")
		}
		for _, entry := range archive.File {
			destination, err := archivePath(root, entry.Name)
			if err != nil {
				return err
			}
			if entry.FileInfo().IsDir() {
				if err := os.MkdirAll(destination, 0o755); err != nil {
					return err
				}
				continue
			}
			if !entry.Mode().IsRegular() || entry.UncompressedSize64 > 2<<30 {
				return errors.New("unsupported file in update ZIP")
			}
			reader, err := entry.Open()
			if err != nil {
				return err
			}
			err = write(entry.Name, entry.Mode(), int64(entry.UncompressedSize64), reader)
			reader.Close()
			if err != nil {
				return err
			}
		}
		return nil
	}
	if format != "tar.gz" {
		return errors.New("unsupported update archive format")
	}
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	archive := tar.NewReader(gz)
	type link struct{ name, target string }
	var links []link
	for {
		entry, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		destination, err := archivePath(root, entry.Name)
		if err != nil {
			return err
		}
		switch entry.Typeflag {
		case tar.TypeDir:
			count++
			if count > 20000 {
				return errors.New("too many archive entries")
			}
			if err := os.MkdirAll(destination, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := write(entry.Name, os.FileMode(entry.Mode), entry.Size, archive); err != nil {
				return err
			}
		case tar.TypeSymlink:
			count++
			if count > 20000 || strings.ContainsAny(entry.Linkname, "\\:\x00") || path.IsAbs(entry.Linkname) {
				return errors.New("unsafe archive link")
			}
			target := path.Clean(path.Join(path.Dir(entry.Name), entry.Linkname))
			if _, err := archivePath(root, target); err != nil {
				return err
			}
			// Bundle links must stay inside their own .app directory.
			if strings.Split(target, "/")[0] != strings.Split(entry.Name, "/")[0] {
				return errors.New("archive link escapes app bundle")
			}
			links = append(links, link{entry.Name, entry.Linkname})
		default:
			return errors.New("unsupported entry in update archive")
		}
	}
	// Drain to validate gzip's trailer and checksum before installation.
	if n, err := io.Copy(io.Discard, io.LimitReader(gz, (1<<20)+1)); err != nil {
		return err
	} else if n > 1<<20 {
		return errors.New("unexpected data after update archive")
	}
	for _, link := range links {
		destination, _ := archivePath(root, link.name)
		for parent := filepath.Dir(destination); parent != root; parent = filepath.Dir(parent) {
			info, err := os.Lstat(parent)
			if err != nil || !info.IsDir() {
				return errors.New("unsafe archive link parent")
			}
		}
		if err := os.Symlink(link.target, destination); err != nil {
			return err
		}
	}
	return nil
}
