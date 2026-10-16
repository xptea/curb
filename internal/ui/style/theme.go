package style

import (
	"github.com/egoist/mygo/ui"
)

// Radius is the common corner radius in device-independent pixels.
const Radius = 12

func Theme(light bool) *ui.Theme {
	t := ui.DarkTheme()
	if light {
		t = ui.LightTheme()
		t.Background, t.Surface = ui.Hex("#fafafa"), ui.Hex("#ffffff")
		t.SurfaceHover, t.SurfacePressed = ui.Hex("#eeeeee"), ui.Hex("#e4e4e7")
		t.Border, t.Text, t.TextMuted = ui.Hex("#e4e4e7"), ui.Hex("#18181b"), ui.Hex("#71717a")
		t.Accent, t.AccentHover, t.AccentPressed, t.AccentText = t.Text, ui.Hex("#303030"), ui.Hex("#454545"), t.Surface
		t.Focus, t.Selection = ui.Hex("#a1a1aa"), ui.Hex("#e4e4e7")
		t.Success, t.Danger = ui.Hex("#237744"), ui.Hex("#b42318")
	} else {
		t.Background, t.Surface = ui.Hex("#181818"), ui.Hex("#202020")
		t.SurfaceHover, t.SurfacePressed = ui.Hex("#292929"), ui.Hex("#343434")
		t.Border, t.Text, t.TextMuted = ui.Hex("#343434"), ui.Hex("#f4f4f5"), ui.Hex("#a1a1aa")
		t.Accent, t.AccentHover, t.AccentPressed, t.AccentText = t.Text, ui.Hex("#dedee2"), ui.Hex("#c4c4cb"), t.Background
		t.Focus, t.Selection = ui.Hex("#71717a"), ui.Hex("#3f3f46")
		t.Success, t.Danger = ui.Hex("#a3e6bf"), ui.Hex("#f0a1a1")
	}
	t.Radius, t.FontSize = Radius, 13
	return t
}
