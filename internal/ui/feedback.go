package view

import (
	"github.com/egoist/mygo/ui"

	"curb/internal/ui/style"
)

func (a *screen) feedback(c *ui.Context) {
	if a.Failure != "" {
		ui.Text(c, a.Failure).Label("Download error").FontSize(12).TextColor(c.Theme().Danger).MaxHeight(66)
		return
	}
	if a.Busy {
		ui.Progress(c, a.Progress).Height(3).Label("Download progress")
		ui.Row(c).Children(func() {
			ui.Text(c, a.Status).SingleLine().Grow(1).FontSize(11).TextColor(c.Theme().TextMuted)
			ui.Text(c, a.Detail).FontSize(11).TextColor(c.Theme().TextMuted)
		})
		return
	}
	if a.SavedPath != "" {
		ui.Row(c).Gap(7).Children(func() {
			ui.Icon(c, style.CheckIcon).Size(15, 15).TextColor(c.Theme().Success)
			ui.Text(c, "Saved to your computer").FontSize(12).TextColor(c.Theme().Success).Grow(1)
			if ui.Button(c, "Show file").Height(26).Padding(3, 8).FontSize(11).Clicked() {
				a.actions.ShowFile(a.SavedPath)
			}
		})
		return
	}
	if a.Status == "Download canceled" {
		ui.Text(c, a.Status).TextAlign(ui.Center).FontSize(12).TextColor(c.Theme().TextMuted)
	}
}
