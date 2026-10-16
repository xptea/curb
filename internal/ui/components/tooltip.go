package components

import (
	"github.com/egoist/mygo/ui"

	"curb/internal/ui/style"
)

func Tooltip(c *ui.Context, anchor ui.Element, text string) {
	t := c.Theme()
	anchor.Description(text)
	ui.TooltipBase(c, anchor, func(tip ui.Element) {
		tip.MaxWidth(320).Padding(5, 8).Radius(style.Radius).
			Background(t.Text).TextColor(t.Background).FontSize(t.FontSize-1).
			Shadow(0, 2, 8, 0, ui.RGBA(0, 0, 0, .2))
		tip.Children(func() { ui.Text(c, text) })
		tip.AttachTo(anchor, ui.AnchorTop, ui.AnchorBottom).Margin(0, 0, 6, 0)
	})
}
