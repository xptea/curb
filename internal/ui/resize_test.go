package view

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestDownloadFormSurvivesResize(t *testing.T) {
	state, render := newTestView()
	tt := ui.NewTester(render, 900, 640)
	for _, size := range [][2]int{{1280, 800}, {900, 640}, {640, 540}, {1920, 1080}, {900, 640}} {
		state.EdgeToEdge = size[0] > 900
		tt.SetSize(size[0], size[1])
		input, ok := tt.Find("Media URL")
		if !ok || input.W < 100 || input.H < 30 || math.Abs(float64(input.Y+input.H/2)-float64(size[1])/2) > 30 {
			t.Fatalf("form missing or misplaced after resize to %v: %+v", size, input)
		}
		// The URL field's surface must actually be painted, not just present
		// in the layout tree after the native surface changes size.
		pixel := tt.Image().RGBAAt(int(input.X+input.W/2), int(input.Y+input.H-2))
		if pixel.R != 0x20 || pixel.G != 0x20 || pixel.B != 0x20 {
			t.Fatalf("form not painted after resize to %v: %+v at %+v", size, pixel, input)
		}
	}
}
