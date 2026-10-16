package view

import (
	"github.com/egoist/mygo/ui"
)

func (a *screen) settingsView(c *ui.Context) {
	ui.Column(c).Gap(7).Children(func() {
		ui.Text(c, "Default save location").FontSize(13).FontWeight(500)
		ui.Row(c).Gap(8).Children(func() {
			ui.TextInput(c, &a.SettingsDestination).Label("Default save location").Grow(1).MinWidth(0).Height(42).FontSize(13)
			ui.Button(c, "Choose…").Height(42).FontSize(13).Disabled(a.Choosing).OnClick(a.actions.ChooseFolder)
		})
	})
	if a.SettingsError != "" {
		ui.Text(c, a.SettingsError).FontSize(12).TextColor(c.Theme().Danger)
	}
	ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
		ui.Button(c, "Cancel").OnClick(func() { a.SettingsOpen = false })
		ui.PrimaryButton(c, "Save defaults").OnClick(a.actions.SaveSettings)
	})
}
