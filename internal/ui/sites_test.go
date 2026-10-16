package view

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSupportedSitesNavigationAndSearch(t *testing.T) {
	state, _ := newTestView()
	state.SupportedSites = []string{"Bandcamp", "vimeo", "youtube", "youtube:playlist"}
	loads := 0
	tt := ui.NewTester(func(c *ui.Context) {
		Render(c, state, Actions{LoadSupportedSites: func() { loads++ }})
	}, 640, 540)
	if err := tt.Click("Supported sites"); err != nil {
		t.Fatal(err)
	}
	if !state.SupportedSitesOpen || loads != 1 || !tt.HasText("Bandcamp") {
		t.Fatal("supported-sites page did not open")
	}
	if err := tt.Click("Search supported sites"); err != nil {
		t.Fatal(err)
	}
	tt.Type("YoUtUbE")
	if !tt.HasText("youtube:playlist") || tt.HasText("Bandcamp") || !tt.HasText("2 of 4 handlers") {
		t.Fatal("search did not filter handlers case-insensitively")
	}
	if err := tt.Click("View yt-dlp’s public list"); err != nil {
		t.Fatal(err)
	}
	if urls := tt.OpenedURLs(); len(urls) != 1 || urls[0] != supportedSitesURL {
		t.Fatalf("incorrect upstream list URL: %v", urls)
	}
	if err := tt.Click("Settings"); err != nil {
		t.Fatal(err)
	}
	if state.SupportedSitesOpen || !state.SettingsOpen || !tt.HasText("Default save location") {
		t.Fatal("settings did not replace the supported-sites page")
	}
	if err := tt.Click("Back to downloads"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Paste a link to download") {
		t.Fatal("downloads did not return")
	}
}

func TestSupportedSitesListIsVirtualAndReachesEnd(t *testing.T) {
	state, render := newTestView()
	state.SupportedSitesOpen = true
	for i := 0; i < 2000; i++ {
		state.SupportedSites = append(state.SupportedSites, fmt.Sprintf("Site %04d", i))
	}
	tt := ui.NewTester(render, 640, 540)
	r, ok := tt.Find("Supported site handlers")
	if !ok || r.H < 50 || r.Y+r.H > 500 {
		t.Fatalf("list does not fit the minimum window: %+v", r)
	}
	if tt.HasText("Site 1999") || !tt.HasText("Site 0000") {
		t.Fatal("list built offscreen rows")
	}
	tt.Scroll(r.X+20, r.Y+20, 0, 100000)
	if !tt.HasText("Site 1999") {
		t.Fatal("last handler cannot be reached")
	}
	if err := tt.Click("Search supported sites"); err != nil {
		t.Fatal(err)
	}
	tt.Type("0000")
	if !tt.HasText("Site 0000") {
		t.Fatal("search did not reset list position")
	}
}
