package view

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestSavedConfirmationExpiresAfterTenSeconds(t *testing.T) {
	state, render := newTestView()
	tt := ui.NewTester(render, 900, 640)
	state.SavedPath, state.Status = "/downloads/video.mp4", "Saved"
	state.SavedUntil = tt.Now().Add(10 * time.Second)
	tt.Frame()
	if !tt.HasText("Saved to your computer") {
		t.Fatal("saved confirmation is missing")
	}
	tt.Advance(9 * time.Second)
	if !tt.HasText("Saved to your computer") {
		t.Fatal("confirmation disappeared before ten seconds")
	}
	tt.Advance(time.Second)
	if tt.HasText("Saved to your computer") || state.SavedPath != "" || !state.SavedUntil.IsZero() {
		t.Fatal("confirmation did not expire without user input")
	}
	if _, found := tt.Find("Show file"); found {
		t.Fatal("expired confirmation still shows its action")
	}
}

func TestEditingURLClearsSavedConfirmationImmediately(t *testing.T) {
	state, render := newTestView()
	tt := ui.NewTester(render, 900, 640)
	state.SavedPath, state.Status = "/downloads/audio.mp3", "Saved"
	state.CopiedToClipboard = true
	state.SavedUntil = tt.Now().Add(10 * time.Second)
	tt.Frame()
	if !tt.HasText("Saved and copied to clipboard") {
		t.Fatal("clipboard confirmation is missing")
	}
	if err := tt.Click("Media URL"); err != nil {
		t.Fatal(err)
	}
	tt.Type("https://example.com/new-video")
	if state.SavedPath != "" || state.CopiedToClipboard || state.ClipboardError != "" || !state.SavedUntil.IsZero() {
		t.Fatal("URL edit left stale saved feedback")
	}
	if tt.HasText("Saved and copied to clipboard") {
		t.Fatal("confirmation remained after editing the URL")
	}
}

func TestEarlierDeadlineDoesNotDismissNewConfirmation(t *testing.T) {
	state, render := newTestView()
	tt := ui.NewTester(render, 900, 640)
	state.SavedPath, state.Status = "/downloads/first.mp4", "Saved"
	state.SavedUntil = tt.Now().Add(10 * time.Second)
	tt.Frame()
	tt.Advance(9 * time.Second)
	state.SavedPath = "/downloads/second.mp4"
	state.SavedUntil = tt.Now().Add(10 * time.Second)
	tt.Frame()
	tt.Advance(time.Second)
	if state.SavedPath != "/downloads/second.mp4" || !tt.HasText("Saved to your computer") {
		t.Fatal("the previous confirmation's deadline dismissed the new one")
	}
	tt.Advance(9 * time.Second)
	if state.SavedPath != "" {
		t.Fatal("new confirmation did not receive its own ten seconds")
	}
}
