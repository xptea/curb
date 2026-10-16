package view

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestVersionButtonChecksInstallsAndBlocksRepeatClicks(t *testing.T) {
	for _, light := range []bool{false, true} {
		state, _ := newTestView()
		state.LightMode = light
		checks, installs := 0, 0
		tester := ui.NewTester(func(c *ui.Context) {
			Render(c, state, Actions{CheckUpdates: func() { checks++ }, InstallUpdate: func() { installs++; state.Updating = true }})
		}, 900, 640)
		if err := tester.Click("App version"); err != nil || checks != 1 {
			t.Fatal("version did not request an update check", err)
		}
		state.UpdateVersion = "0.2.0"
		tester.Frame()
		if !tester.HasText("Update to v0.2.0") {
			t.Fatal("new version was not offered")
		}
		button, _ := tester.Find("App version")
		pixel := tester.Image().RGBAAt(int(button.X+3), int(button.Y+button.H/2))
		if pixel.G <= pixel.R || pixel.G <= pixel.B {
			t.Fatalf("update button is not green: %+v", pixel)
		}
		state.Busy = true
		tester.Frame()
		_ = tester.Click("App version")
		if installs != 0 {
			t.Fatal("update interrupted a media download")
		}
		state.Busy = false
		tester.Frame()
		if err := tester.Click("App version"); err != nil || installs != 1 {
			t.Fatal("update button did not start installation", err)
		}
		if _, ok := tester.Find("Update progress"); !ok {
			t.Fatal("installation spinner is missing")
		}
		_ = tester.Click("App version")
		if installs != 1 {
			t.Fatal("repeated click started another installation")
		}
	}
}
