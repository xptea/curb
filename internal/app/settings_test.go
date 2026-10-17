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

func TestOutputFormatsSurviveRestart(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("CURB_DOWNLOAD_DIR", "")
	a := newApplication("test")
	if a.VideoOutputFormat != "MP4" || a.AudioOutputFormat != "MP3" {
		t.Fatal("expected friendly defaults for new installations")
	}
	for _, formats := range [][2]string{{"Original", "Original"}, {"MKV", "FLAC"}, {"MP4", "M4A"}} {
		a.SettingsDestination = a.Destination
		a.SettingsVideoOutputFormat, a.SettingsAudioOutputFormat = formats[0], formats[1]
		a.saveSettings()
		if a.SettingsError != "" {
			t.Fatal(a.SettingsError)
		}
		b := newApplication("test")
		if b.VideoOutputFormat != formats[0] || b.AudioOutputFormat != formats[1] {
			t.Fatalf("output formats did not survive restart: %s / %s", b.VideoOutputFormat, b.AudioOutputFormat)
		}
	}
}
