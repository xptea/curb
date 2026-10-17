package download

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishKeepsExistingDownloads(t *testing.T) {
	source := filepath.Join(t.TempDir(), "media.mp4")
	if err := os.WriteFile(source, []byte("new download"), 0600); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	existing := filepath.Join(destination, "media.mp4")
	if err := os.WriteFile(existing, []byte("previous download"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"media (1).mp4", "media (2).mp4"} {
		path, err := publishOutput(context.Background(), source, destination)
		if err != nil || filepath.Base(path) != name {
			t.Fatalf("safe filename not selected: %s %v", path, err)
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != "new download" {
			t.Fatal("published file differs from final converted file")
		}
	}
	data, err := os.ReadFile(existing)
	if err != nil || string(data) != "previous download" {
		t.Fatal("existing download was overwritten")
	}
}

func TestCopyFallbackPreservesExistingFileAndCleansCanceledCopy(t *testing.T) {
	source, target := filepath.Join(t.TempDir(), "source.mp3"), filepath.Join(t.TempDir(), "target.mp3")
	if err := os.WriteFile(source, []byte("audio"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := copyOutput(context.Background(), source, target); err != nil {
		t.Fatal(err)
	}
	if err := copyOutput(context.Background(), source, target); !errors.Is(err, os.ErrExist) {
		t.Fatal("copy must not replace existing files")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	partial := filepath.Join(t.TempDir(), "canceled.mp3")
	if err := copyOutput(ctx, source, partial); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel was ignored: %v", err)
	}
	if _, err := os.Stat(partial); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("canceled copy left a partial file")
	}
}
