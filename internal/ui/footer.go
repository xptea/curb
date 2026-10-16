package view

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"curb/internal/ui/components"
	"curb/internal/ui/style"
)

func (a *screen) footer(c *ui.Context) {
	// Both items share a row and equal padding. Their outer bounds sit the
	// same distance from the bottom and their respective side of the window.
	ui.Row(c).Absolute().Left(8).Right(8).Bottom(8).AlignItems(ui.Center).Justify(ui.SpaceBetween).Children(func() {
		a.versionButton(c)
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

func (a *screen) versionButton(c *ui.Context) {
	theme := c.Theme()
	available := a.UpdateVersion != ""
	label := "v" + a.Version
	if available {
		label = "Update to v" + a.UpdateVersion
	}
	if a.Updating {
		label = fmt.Sprintf("Updating… %.0f%%", a.UpdateProgress*100)
	}
	button := ui.ButtonBase(c).Label("App version").Padding(8).Gap(6).
		Radius(style.Radius).TextColor(theme.TextMuted).
		Disabled(a.UpdateChecking || available && !a.Updating && (a.Busy || a.Choosing))
	if available || a.Updating {
		button.Background(ui.Hex("#23834a")).TextColor(ui.Hex("#ffffff"))
		if button.Pressed() {
			button.Background(ui.Hex("#185c34"))
		} else if button.Hovered() {
			button.Background(ui.Hex("#1d713f"))
		}
	} else if button.Hovered() {
		button.Background(theme.SurfaceHover).TextColor(theme.Text)
	}
	button.OnClick(func() {
		if a.Updating || a.UpdateChecking {
			return
		}
		if available {
			if a.actions.InstallUpdate != nil {
				a.actions.InstallUpdate()
			}
		} else if a.actions.CheckUpdates != nil {
			a.actions.CheckUpdates()
		}
	}).Children(func() {
		if a.Updating || a.UpdateChecking {
			color := theme.TextMuted
			if a.Updating {
				color = ui.Hex("#ffffff")
			}
			ui.Spinner(c).Size(13, 13).TextColor(color).Label("Update progress")
		} else if available {
			ui.Icon(c, style.DownloadIcon).Size(13, 13)
		}
		ui.Text(c, label).FontSize(11)
	})
	hint := "Check for updates"
	if available {
		hint = "Download, install, and restart Curb"
	}
	if available && (a.Busy || a.Choosing) {
		hint = "Finish the current download or folder selection before updating"
	}
	if a.UpdateError != "" {
		hint = a.UpdateError
	}
	components.Tooltip(c, button, hint)
	if a.UpdateError != "" {
		ui.Text(c, a.UpdateError).Label("Update error").Absolute().Left(16).Bottom(49).
			MaxWidth(440).FontSize(11).TextColor(theme.Danger)
	}
}
