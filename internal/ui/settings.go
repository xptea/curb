package view

import (
	"curb/internal/media"
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
	ui.Column(c).Gap(8).Children(func() {
		ui.Row(c).Gap(16).Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Gap(7).Children(func() {
				ui.Text(c, "Default video format").FontSize(13).FontWeight(500)
				ui.Select(c, &a.SettingsVideoOutputFormat, media.VideoOutputs).Label("Video output format").FillWidth().Height(42).FontSize(13)
			})
			ui.Column(c).Grow(1).MinWidth(0).Gap(7).Children(func() {
				ui.Text(c, "Default audio format").FontSize(13).FontWeight(500)
				ui.Select(c, &a.SettingsAudioOutputFormat, media.AudioOutputs).Label("Audio output format").FillWidth().Height(42).FontSize(13)
			})
		})
		ui.Text(c, "Original keeps the source format. Other options convert when needed.").FontSize(12).TextColor(c.Theme().TextMuted)
	})
	ui.Row(c).AlignItems(ui.Center).Gap(16).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(5).Children(func() {
			ui.Text(c, "Automatically copy downloaded files").FontSize(13).FontWeight(500)
			ui.Text(c, "Copy the finished file to your clipboard, ready to paste.").FontSize(12).TextColor(c.Theme().TextMuted)
		})
		ui.Switch(c, &a.SettingsAutoCopyDownloads).Label("Auto-copy downloaded files").Shrink(0)
	})
	if a.SettingsError != "" {
		ui.Text(c, a.SettingsError).FontSize(12).TextColor(c.Theme().Danger)
	}
	ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
		ui.Button(c, "Cancel").OnClick(func() { a.SettingsOpen = false })
		ui.PrimaryButton(c, "Save defaults").OnClick(a.actions.SaveSettings)
	})
}
