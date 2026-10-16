package view

import (
	"github.com/egoist/mygo/ui"
	"math"
	"testing"
)

func newTestView() (*State, func(*ui.Context)) {
	state := &State{Version: "test", Destination: "/downloads", Format: "Video", Quality: "Best available", AudioQuality: "Best quality", Progress: -1}
	return state, func(c *ui.Context) { Render(c, state, Actions{}) }
}

func TestTitleStaysAtWindowCenter(t *testing.T) {
	for _, width := range []int{640, 900, 1280} {
		for _, bar := range []ui.TitleBar{{Height: 46, Right: 140}, {Height: 46, Left: 100}, {Height: 46, Left: 40, Right: 140}, {Height: 46}} {
			_, render := newTestView()
			tt := ui.NewTester(render, width, 640)
			tt.SetTitleBar(bar)
			r, ok := tt.Find("Application title")
			if !ok {
				t.Fatal("title missing")
			}
			if delta := math.Abs(float64(r.X+r.W/2) - float64(width)/2); delta > 1 {
				t.Errorf("width=%d controls=%+v: title offset %.2fpx", width, bar, delta)
			}
			input, ok := tt.Find("Media URL")
			if !ok || input.X < 0 || input.X+input.W > float32(width) {
				t.Errorf("input clipped at width=%d: %+v", width, input)
			}
		}
	}
}

func TestPageTitlesAndFooterAlignment(t *testing.T) {
	for _, width := range []int{640, 900, 1280} {
		for _, light := range []bool{false, true} {
			for _, page := range []string{"downloads", "settings", "sites"} {
				state, render := newTestView()
				state.LightMode = light
				state.SettingsOpen, state.SupportedSitesOpen = page == "settings", page == "sites"
				tt := ui.NewTester(render, width, 640)
				if page != "downloads" {
					title, ok := tt.Find("Page title")
					if !ok || math.Abs(float64(title.X+title.W/2)-float64(width)/2) > 1 {
						t.Fatalf("%s title is not at the window center at width %d: %+v", page, width, title)
					}
				}
				version, vok := tt.Find("App version")
				sites, sok := tt.Find("Supported sites")
				if !vok || !sok {
					t.Fatal("footer items missing")
				}
				left, right := version.X, float32(width)-sites.X-sites.W
				versionBottom, sitesBottom := 640-version.Y-version.H, 640-sites.Y-sites.H
				for _, delta := range []float32{left - right, left - versionBottom, right - sitesBottom, version.Y + version.H/2 - sites.Y - sites.H/2} {
					if math.Abs(float64(delta)) > 1 {
						t.Fatalf("%s footer is uneven at width %d: version=%+v, sites=%+v", page, width, version, sites)
					}
				}
			}
		}
	}
}
