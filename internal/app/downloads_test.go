package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestCompletedDownloadsCopyFinalFile(t *testing.T) {
	for _, extension := range []string{".mp4", ".mp3"} {
		t.Run(extension, func(t *testing.T) {
			a := &application{}
			a.AutoCopyDownloads = true
			a.Busy = true
			path := filepath.Join(t.TempDir(), "My media"+extension)
			var copied []string
			a.copyFile = func(path string) error { copied = append(copied, path); return nil }
			before := time.Now()
			a.finishDownload(path, nil)
			if a.SavedUntil.Before(before.Add(10*time.Second)) || a.SavedUntil.After(time.Now().Add(10*time.Second)) {
				t.Fatal("saved confirmation was not scheduled for ten seconds")
			}
			if len(copied) != 1 || copied[0] != path || !a.CopiedToClipboard || a.SavedPath != path || a.Busy || a.Progress != 1 {
				t.Fatalf("final file was not copied after success: paths=%v state=%+v", copied, a.State)
			}
		})
	}
}

func TestDownloadDoesNotCopyWhenDisabledCanceledOrFailed(t *testing.T) {
	for _, tc := range []struct {
		name    string
		enabled bool
		err     error
	}{
		{name: "disabled"},
		{name: "canceled", enabled: true, err: context.Canceled},
		{name: "failed", enabled: true, err: errors.New("download failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &application{}
			a.AutoCopyDownloads = tc.enabled
			a.CopiedToClipboard = true
			a.copyFile = func(string) error { t.Fatal("clipboard changed unexpectedly"); return nil }
			a.finishDownload("/downloads/media.mp4", tc.err)
			if a.CopiedToClipboard || a.ClipboardError != "" {
				t.Fatalf("stale clipboard status: %+v", a.State)
			}
		})
	}
}

func TestClipboardFailureKeepsSuccessfulDownload(t *testing.T) {
	a := &application{}
	a.AutoCopyDownloads = true
	a.copyFile = func(string) error { return errors.New("clipboard unavailable") }
	path := filepath.Join(t.TempDir(), "video.mp4")
	a.finishDownload(path, nil)
	if a.Failure != "" || a.SavedPath != path || a.Status != "Saved" || a.Progress != 1 || a.CopiedToClipboard || a.ClipboardError != "clipboard unavailable" {
		t.Fatalf("clipboard failure hid a successful download: %+v", a.State)
	}
}
