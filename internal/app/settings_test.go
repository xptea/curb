package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSavedDestinationSurvivesRestart(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("CURB_DOWNLOAD_DIR", "")
	a := newApplication("test")
	a.SettingsDestination = filepath.Join(t.TempDir(), "My downloads")
	a.saveSettings()
	if a.SettingsError != "" {
		t.Fatal(a.SettingsError)
	}
	b := newApplication("test")
	if b.Destination != a.SettingsDestination {
		t.Fatal("saved defaults did not survive restart")
	}
	if _, err := os.Stat(b.settingsPath); err != nil {
		t.Fatal(err)
	}
}

func TestAutoCopyPreferenceSurvivesRestart(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("CURB_DOWNLOAD_DIR", "")
	a := newApplication("test")
	if a.AutoCopyDownloads {
		t.Fatal("clipboard copying must be opt-in")
	}
	for _, enabled := range []bool{true, false} {
		a.SettingsDestination = a.Destination
		a.SettingsAutoCopyDownloads = enabled
		a.saveSettings()
		if a.SettingsError != "" {
			t.Fatal(a.SettingsError)
		}
		b := newApplication("test")
		if b.AutoCopyDownloads != enabled || a.AutoCopyDownloads != enabled {
			t.Fatal("clipboard preference did not survive restart")
		}
	}
}
