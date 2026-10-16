package components

import (
	"github.com/egoist/mygo/ui"

	"curb/internal/ui/style"
)

func QualitySelect(c *ui.Context, selected *string, options []string, label, hint string, disabled bool) {
	t := c.Theme()
	sel := ui.SelectBase(c, selected)
	button := sel.Trigger.Label(label).Height(34).Padding(6, 12).Gap(10).
		Radius(style.Radius).Border(1, t.Border).Background(t.Surface).TextColor(t.Text).
		FontSize(12).Justify(ui.SpaceBetween).Disabled(disabled)
	if !disabled {
		if button.Pressed() {
			button.Background(t.SurfacePressed)
		} else if button.Hovered() {
			button.Background(t.SurfaceHover)
		}
	}
	button.Children(func() {
		ui.Text(c, *selected).SingleLine()
		ui.Icon(c, style.ChevronIcon).Size(10, 10).Shrink(0).TextColor(t.TextMuted)
	})
	Tooltip(c, button, hint)
	sel.Popup(func(panel ui.Element) {
		panel.AttachTo(button, ui.AnchorBottomEnd, ui.AnchorTopEnd).Margin(4, 0, 0, 0).Padding(4).Radius(style.Radius).
			Background(t.Background).Border(1, t.Border).
			Shadow(0, 6, 20, 0, ui.RGBA(0, 0, 0, .18))
		for _, option := range options {
			item := sel.Item(option).Padding(6, 10).Radius(style.Radius)
			if item.Highlighted() {
				item.Background(t.Accent).TextColor(t.AccentText)
			} else if option == *selected {
				item.Background(t.Surface)
			}
			item.Children(func() { ui.Text(c, option).SingleLine() })
		}
	})
}
