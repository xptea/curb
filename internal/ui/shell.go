package view

import (
	"github.com/egoist/mygo/ui"

	"curb/internal/ui/components"
	"curb/internal/ui/style"
)

// Render builds one frame from state; MyGo owns layout and rendering.
func Render(c *ui.Context, state *State, actions Actions) {
	s := screen{State: state, actions: actions}
	s.render(c)
}

func (a *screen) render(c *ui.Context) {
	if a.SavedPath != "" && !a.SavedUntil.IsZero() {
		if remaining := a.SavedUntil.Sub(c.Now()); remaining > 0 {
			c.After(remaining)
		} else {
			a.ClearSavedFeedback()
		}
	}
	c.SetTheme(style.Theme(a.LightMode))
	c.Root().Background(ui.Transparent)
	radius := float32(style.Radius)
	if a.EdgeToEdge {
		radius = 0
	}
	ui.Box(c).Fill().Background(c.Theme().Background).Radius(radius).Children(func() {
		a.titleBar(c)
		if a.SupportedSitesOpen {
			a.supportedSitesView(c)
		} else if a.SettingsOpen {
			ui.Scroll(c.Key("settings-page")).Grow(1).Padding(20, 32, 60, 32).Children(func() {
				ui.Column(c).Width(620).MaxWidthPercent(100).Margin(0, ui.Auto).Gap(24).Children(func() {
					components.PageHeader(c, "Settings", func() { a.SettingsOpen = false }, nil)
					a.settingsView(c)
				})
			})
		} else {
			ui.Column(c).Grow(1).MinHeight(0).Center().Children(func() {
				ui.Column(c).Width(620).MaxWidthPercent(84).Children(func() { a.downloadForm(c) })
			})
			// Balance the title bar so the URL stays at the window's center.
			ui.Box(c).Height(max(c.TitleBar().Height, 46)).Shrink(0)
		}
		a.footer(c)
	})
}
