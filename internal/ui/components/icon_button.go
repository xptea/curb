package components

import (
	"github.com/egoist/mygo/ui"

	"curb/internal/ui/style"
)

func IconButton(c *ui.Context, label string, symbol *ui.SVG) ui.Element {
	button := ui.ButtonBase(c).Label(label).Size(32, 32).Radius(style.Radius).TextColor(c.Theme().TextMuted)
	if button.Hovered() {
		button.Background(c.Theme().SurfaceHover)
	}
	Tooltip(c, button, label)
	return button.Children(func() { ui.Icon(c, symbol).Size(18, 18).TextColor(c.Theme().TextMuted) })
}
