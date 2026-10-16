package components

import (
	"github.com/egoist/mygo/ui"

	"curb/internal/ui/style"
)

func BackButton(c *ui.Context, onClick func()) {
	t := c.Theme()
	button := ui.ButtonBase(c).Label("Back to downloads").Height(32).Padding(0, 8).
		Margin(0, 0, 0, -8).Gap(7).AlignSelf(ui.Start).Radius(style.Radius).TextColor(t.TextMuted)
	if button.Pressed() {
		button.Background(t.SurfacePressed).TextColor(t.Text)
	} else if button.Hovered() {
		button.Background(t.SurfaceHover).TextColor(t.Text)
	}
	button.OnClick(onClick).Children(func() {
		ui.Icon(c, style.BackIcon).Size(16, 16)
		ui.Text(c, "Back").FontSize(12)
	})
}
