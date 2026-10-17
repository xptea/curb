package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveReplacesPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "curb", "settings.json")
	for _, destination := range []string{"/downloads/first", "/downloads/with spaces"} {
		if err := Save(path, Preferences{Destination: destination}); err != nil {
			t.Fatal(err)
		}
		got, err := Load(path)
		if err != nil || got.Destination != destination {
			t.Fatalf("settings did not round trip: %+v, %v", got, err)
		}
	}
	leftovers, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".settings-*"))
	if err != nil || len(leftovers) != 0 {
		t.Fatalf("temporary settings files left behind: %v, %v", leftovers, err)
	}
}

func TestLoadMalformedPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("malformed settings should return an error for the app to handle")
	}
}

func TestExistingPreferencesLeaveClipboardUnchanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"destination":"/downloads"}`), 0600); err != nil {
		t.Fatal(err)
	}
	preferences, err := Load(path)
	if err != nil || preferences.AutoCopyDownloads || preferences.Destination != "/downloads" {
		t.Fatalf("legacy preferences changed behavior: %+v, %v", preferences, err)
	}
}
