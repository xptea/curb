package components

import (
	"time"

	"github.com/egoist/mygo/ui"

	"curb/internal/ui/style"
)

func MediaTypeToggle(c *ui.Context, format *string, disabled bool) {
	selected := 0
	if *format == "Audio" {
		selected = 1
	}
	seg := ui.SegmentedBase(c.Key("download-type"), &selected, 2)
	track := seg.Track.Label("Download type").Width(178).Height(34).Shrink(0).Padding(3).Gap(2).
		Radius(style.Radius).Background(c.Theme().Surface).Border(1, c.Theme().Border).Disabled(disabled)
	offset := track.Animate("mode-thumb", float32(selected)*86, 140*time.Millisecond)
	thumb := c.Theme().SurfaceHover
	track.Draw(func(p *ui.Painter, r ui.Rect) {
		p.Fill(ui.Rect{X: r.X + 4 + offset, Y: r.Y + 4, W: 84, H: 26}, thumb, style.Radius)
	}).Children(func() {
		for i, option := range []struct {
			name   string
			symbol *ui.SVG
		}{{"Video", style.VideoIcon}, {"Audio", style.AudioIcon}} {
			tint := c.Theme().TextMuted
			if selected == i {
				tint = c.Theme().Text
			}
			segment := seg.Segment(i).Label(option.name).Size(84, 26).Gap(6).Radius(style.Radius).Disabled(disabled).
				Children(func() {
					ui.Icon(c, option.symbol).Size(15, 15).TextColor(tint)
					ui.Text(c, option.name).FontSize(12).TextColor(tint)
				})
			Tooltip(c, segment, option.name)
		}
	})
	if selected == 1 {
		*format = "Audio"
	} else {
		*format = "Video"
	}
}
