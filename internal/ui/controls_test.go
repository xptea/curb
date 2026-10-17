package view

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSettingsNavigationAndTheme(t *testing.T) {
	state, render := newTestView()
	tt := ui.NewTester(render, 900, 640)
	if err := tt.Click("Settings"); err != nil {
		t.Fatal(err)
	}
	if !state.SettingsOpen || !tt.HasText("Default save location") {
		t.Fatal("settings page did not open")
	}
	if state.SettingsDestination != state.Destination {
		t.Fatal("settings page did not receive the saved destination")
	}
	if err := tt.Click("Switch to light mode"); err != nil {
		t.Fatal(err)
	}
	if !state.LightMode {
		t.Fatal("theme callback did not update shared state")
	}
	if err := tt.Click("Back to downloads"); err != nil {
		t.Fatal(err)
	}
	if state.SettingsOpen || !tt.HasText("Paste a link to download") {
		t.Fatal("back button did not restore downloads")
	}
}

func TestMediaAndQualityControls(t *testing.T) {
	state, render := newTestView()
	tt := ui.NewTester(render, 900, 640)
	before, ok := tt.Find("Video quality")
	if !ok {
		t.Fatal("video quality control is missing")
	}
	for _, label := range []string{"Video quality", "720p"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	after, _ := tt.Find("Video quality")
	if state.Quality != "720p" || after.W >= before.W {
		t.Fatalf("quality selection did not update its content-sized trigger: %q, %.1f -> %.1f", state.Quality, before.W, after.W)
	}
	if math.Abs(float64(after.X+after.W-before.X-before.W)) > 1 {
		t.Fatal("quality control lost its right alignment")
	}
	for _, label := range []string{"Audio", "Audio quality", "192 kbps"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	if state.Format != "Audio" || state.AudioQuality != "192 kbps" {
		t.Fatalf("audio controls did not update shared state: %q, %q", state.Format, state.AudioQuality)
	}
}

func TestAutoCopySettingsAreDraftUntilSaved(t *testing.T) {
	state, render := newTestView()
	state.AutoCopyDownloads = true
	tt := ui.NewTester(render, 900, 640)
	if err := tt.Click("Settings"); err != nil {
		t.Fatal(err)
	}
	if !state.SettingsAutoCopyDownloads {
		t.Fatal("clipboard switch did not reflect the saved preference")
	}
	if err := tt.Click("Auto-copy downloaded files"); err != nil {
		t.Fatal(err)
	}
	if state.SettingsAutoCopyDownloads || !state.AutoCopyDownloads {
		t.Fatal("editing a draft changed the saved preference")
	}
	for _, label := range []string{"Cancel", "Settings"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	if !state.SettingsAutoCopyDownloads {
		t.Fatal("cancel did not discard the draft clipboard preference")
	}
}

func TestOutputFormatSettingsCancelAndReopen(t *testing.T) {
	state, render := newTestView()
	state.VideoOutputFormat, state.AudioOutputFormat = "MP4", "MP3"
	tt := ui.NewTester(render, 900, 640)
	for _, label := range []string{"Settings", "Video output format", "Original", "Audio output format", "FLAC"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	if state.SettingsVideoOutputFormat != "Original" || state.SettingsAudioOutputFormat != "FLAC" {
		t.Fatal("output selectors did not update draft settings")
	}
	if state.VideoOutputFormat != "MP4" || state.AudioOutputFormat != "MP3" {
		t.Fatal("draft output formats were applied before saving")
	}
	for _, label := range []string{"Cancel", "Settings"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
	}
	if state.SettingsVideoOutputFormat != "MP4" || state.SettingsAudioOutputFormat != "MP3" {
		t.Fatal("cancel did not discard output format edits")
	}
}

func TestSourceAndLosslessAudioDoNotOfferBitrate(t *testing.T) {
	for _, format := range []string{"Original", "WAV", "FLAC"} {
		state, render := newTestView()
		state.Format, state.AudioOutputFormat = "Audio", format
		tt := ui.NewTester(render, 900, 640)
		if _, found := tt.Find("Audio quality"); found {
			t.Fatalf("%s should not offer a lossy bitrate selector", format)
		}
	}
}
