package components

import "github.com/egoist/mygo/ui"

// PageHeader reserves equal space on each side so navigation and actions
// never shift the title away from the window's center.
func PageHeader(c *ui.Context, title string, onBack func(), actions func()) {
	ui.Row(c).Height(32).Shrink(0).Padding(0, 100).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		ui.Text(c, title).Label("Page title").FontSize(22).FontWeight(600).SingleLine()
		ui.Box(c).Absolute().Left(0).Top(0).Children(func() { BackButton(c, onBack) })
		if actions != nil {
			ui.Row(c).Absolute().Right(0).Top(0).Height(32).AlignItems(ui.Center).Children(actions)
		}
	})
}
