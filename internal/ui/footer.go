package view

import (
	"github.com/egoist/mygo/ui"

	"curb/internal/ui/style"
)

func (a *screen) footer(c *ui.Context) {
	// Both items share a row and equal padding. Their outer bounds sit the
	// same distance from the bottom and their respective side of the window.
	ui.Row(c).Absolute().Left(8).Right(8).Bottom(8).AlignItems(ui.Center).Justify(ui.SpaceBetween).Children(func() {
		ui.Text(c, "v"+a.Version).Label("App version").FontSize(11).Padding(8).TextColor(c.Theme().TextMuted)
		button := ui.ButtonBase(c).Label("Supported sites").Padding(8).
			Radius(style.Radius).TextColor(c.Theme().TextMuted)
		if button.Pressed() {
			button.Background(c.Theme().SurfacePressed).TextColor(c.Theme().Text)
		} else if button.Hovered() {
			button.Background(c.Theme().SurfaceHover).TextColor(c.Theme().Text)
		}
		button.OnClick(func() {
			a.SupportedSitesOpen, a.SettingsOpen = true, false
			if a.actions.LoadSupportedSites != nil {
				a.actions.LoadSupportedSites()
			}
		}).Children(func() { ui.Text(c, "Supported sites").FontSize(11) })
	})
}
