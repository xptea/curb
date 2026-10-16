package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestArchiveRejectsTraversalAndEscapingSymlinks(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", "Curb/../../escape", "Curb\\escape", "Curb/file:stream"} {
		root := t.TempDir()
		file, _ := os.Create(filepath.Join(root, "update.zip"))
		archive := zip.NewWriter(file)
		entry, _ := archive.Create(name)
		entry.Write([]byte("payload"))
		archive.Close()
		file.Close()
		if err := extractArchive(context.Background(), file.Name(), filepath.Join(root, "stage"), "zip"); err == nil {
			t.Fatalf("accepted unsafe path %s", name)
		}
	}
	root := t.TempDir()
	file, _ := os.Create(filepath.Join(root, "update.tar.gz"))
	gz := gzip.NewWriter(file)
	archive := tar.NewWriter(gz)
	archive.WriteHeader(&tar.Header{Name: "Curb.app/link", Linkname: "../../outside", Typeflag: tar.TypeSymlink, Mode: 0o777})
	archive.Close()
	gz.Close()
	file.Close()
	if err := extractArchive(context.Background(), file.Name(), filepath.Join(root, "stage"), "tar.gz"); err == nil {
		t.Fatal("accepted escaping symlink")
	}
}

func TestBundleExtractionPreservesExecutableAndInternalLinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("macOS bundle symlinks are exercised on Unix runners")
	}
	root := t.TempDir()
	file, _ := os.Create(filepath.Join(root, "update.tar.gz"))
	gz := gzip.NewWriter(file)
	archive := tar.NewWriter(gz)
	for _, entry := range []struct {
		name, body, link string
		kind             byte
		mode             int64
	}{
		{"Curb.app/Contents/MacOS/Curb", "binary", "", tar.TypeReg, 0o755},
		{"Curb.app/Contents/MacOS/current", "", "Curb", tar.TypeSymlink, 0o777},
	} {
		archive.WriteHeader(&tar.Header{Name: entry.name, Size: int64(len(entry.body)), Typeflag: entry.kind, Mode: entry.mode, Linkname: entry.link})
		archive.Write([]byte(entry.body))
	}
	archive.Close()
	gz.Close()
	file.Close()
	stage := filepath.Join(root, "stage")
	if err := extractArchive(context.Background(), file.Name(), stage, "tar.gz"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(stage, "Curb.app/Contents/MacOS/Curb"))
	if err != nil || info.Mode()&0o111 == 0 {
		t.Fatal("lost executable permissions", err)
	}
	if link, err := os.Readlink(filepath.Join(stage, "Curb.app/Contents/MacOS/current")); err != nil || link != "Curb" {
		t.Fatal("lost bundle symlink", link, err)
	}
}
