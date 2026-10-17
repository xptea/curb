package view

import (
	"runtime"

	"github.com/egoist/mygo/ui"

	"curb/internal/media"
	"curb/internal/ui/components"
	"curb/internal/ui/style"
)

func (a *screen) titleBar(c *ui.Context) {
	bar := c.TitleBar()
	height := max(float32(bar.Height), 46)
	inset := (height - 32) / 2
	// Equal reserved widths keep the title at the WINDOW midpoint even
	// when GTK puts all decoration buttons on just one side.
	reserve := max(float32(bar.Left), float32(bar.Right)) + 12
	ui.Row(c).Height(height).Shrink(0).FillWidth().Padding(0, reserve).
		Justify(ui.Center).DragWindow().Children(func() {
		ui.Text(c, "Curb").Label("Application title").FontSize(13).FontWeight(600).TextColor(c.Theme().TextMuted)
		label, symbol := "Switch to light mode", style.SunIcon
		if a.LightMode {
			label, symbol = "Switch to dark mode", style.MoonIcon
		}
		tools := ui.Row(c).Gap(4).Absolute().Top(inset)
		if runtime.GOOS == "darwin" {
			tools.Right(inset + bar.Right)
		} else {
			tools.Left(inset + bar.Left)
		}
		tools.Children(func() {
			components.IconButton(c, "Settings", style.SettingsIcon).OnClick(func() {
				a.SupportedSitesOpen = false
				a.SettingsOpen = !a.SettingsOpen
				if a.SettingsOpen {
					a.SettingsDestination, a.SettingsError = a.Destination, ""
					a.SettingsAutoCopyDownloads = a.AutoCopyDownloads
					a.SettingsVideoOutputFormat = media.VideoOutput(a.VideoOutputFormat)
					a.SettingsAudioOutputFormat = media.AudioOutput(a.AudioOutputFormat)
				}
			})
			components.IconButton(c, label, symbol).OnClick(a.toggleTheme)
		})
		if runtime.GOOS != "darwin" {
			ui.Row(c).Gap(4).Absolute().Top(inset).Right(inset).Children(func() {
				components.IconButton(c, "Minimize", style.MinimizeIcon).OnClick(a.actions.Minimize)
				label, symbol := "Maximize", style.MaximizeIcon
				if a.EdgeToEdge {
					label, symbol = "Restore window", style.RestoreIcon
				}
				components.IconButton(c, label, symbol).OnClick(a.actions.ToggleMaximize)
				components.IconButton(c, "Close", style.CloseIcon).OnClick(a.actions.Close)
			})
		}
	})
}

func (a *screen) toggleTheme() {
	a.LightMode = !a.LightMode
}
