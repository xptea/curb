package view

import (
	"github.com/egoist/mygo/ui"

	"curb/internal/media"
	"curb/internal/ui/components"
	"curb/internal/ui/style"
)

func (a *screen) downloadForm(c *ui.Context) {
	ui.Text(c, "Paste a link to download").FontSize(22).FontWeight(500).
		TextAlign(ui.Center).FillWidth().Margin(0, 0, 22, 0)
	ui.Row(c).Height(54).Padding(6, 8, 6, 18).Gap(10).Radius(style.Radius).
		Border(1, c.Theme().Border).Background(c.Theme().Surface).Children(func() {
		ui.TextInputBase(c.Key("media-url"), &a.URL).Grow(1).Height(40).Padding(11, 0).
			FontSize(14).Placeholder("Paste a video or audio link…").Label("Media URL").Disabled(a.Busy).
			OnChange(func() {
				a.Failure = ""
				a.ClearSavedFeedback()
			}).OnSubmit(a.actions.Download)
		if a.Busy {
			if ui.Button(c, "Cancel").Height(34).Padding(6, 10).FontSize(12).
				Border(0, ui.Transparent).Background(ui.Transparent).Clicked() {
				a.actions.Cancel()
			}
		} else {
			ready := a.CanDownload
			fill, tint := ui.Transparent, c.Theme().TextMuted
			if ready {
				fill, tint = c.Theme().Accent, c.Theme().AccentText
			}
			button := ui.ButtonBase(c).Label("Download").Size(34, 34).
				Padding(8).Radius(style.Radius).Background(fill).TextColor(tint).Disabled(!ready)
			components.Tooltip(c, button, "Download · Enter")
			if ready && button.Hovered() {
				button.Background(c.Theme().AccentHover)
			}
			if !ready {
				button.Opacity(.4)
			}
			button.OnClick(a.actions.Download).Children(func() {
				ui.Icon(c, style.ArrowIcon).Size(18, 18).TextColor(tint)
			})
		}
	})
	ui.Row(c).Margin(12, 0, 0, 0).FillWidth().Gap(8).Children(func() {
		components.MediaTypeToggle(c, &a.Format, a.Busy)
		ui.Spacer(c)
		if a.Format == "Audio" {
			if media.AudioHasBitrate(a.AudioOutputFormat) {
				components.QualitySelect(c, &a.AudioQuality, []string{"Best quality", "320 kbps", "256 kbps", "192 kbps", "128 kbps"},
					"Audio quality", media.AudioOutput(a.AudioOutputFormat)+" audio quality", a.Busy)
			} else {
				label := "Lossless"
				if media.AudioOutput(a.AudioOutputFormat) == media.Original {
					label = "Source quality"
				}
				ui.Text(c, label).FontSize(12).TextColor(c.Theme().TextMuted).Margin(8, 0)
			}
		} else {
			components.QualitySelect(c, &a.Quality, []string{"Best available", "2160p", "1440p", "1080p", "720p", "480p", "360p"},
				"Video quality", "Highest available quality at or below this resolution", a.Busy)
		}
	})
	ui.Column(c).Absolute().Top(163).FillWidth().Gap(10).Children(func() { a.feedback(c) })
}
