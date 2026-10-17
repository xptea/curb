package app

import (
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"

	"curb/internal/media"
	"curb/internal/settings"
)

func (a *application) loadSettings() {
	path, err := settings.Path()
	if err != nil {
		return
	}
	a.settingsPath = path
	preferences, err := settings.Load(path)
	if err != nil {
		return
	}
	a.AutoCopyDownloads = preferences.AutoCopyDownloads
	a.VideoOutputFormat = media.VideoOutput(preferences.VideoOutputFormat)
	a.AudioOutputFormat = media.AudioOutput(preferences.AudioOutputFormat)
	if filepath.IsAbs(preferences.Destination) {
		a.Destination = preferences.Destination
	}
}

func (a *application) saveSettings() {
	destination := strings.TrimSpace(a.SettingsDestination)
	if !filepath.IsAbs(destination) {
		a.SettingsError = "Choose an absolute folder path."
		return
	}
	if a.settingsPath == "" {
		a.SettingsError = "Your settings folder is unavailable."
		return
	}
	videoOutput := media.VideoOutput(a.SettingsVideoOutputFormat)
	audioOutput := media.AudioOutput(a.SettingsAudioOutputFormat)
	if err := settings.Save(a.settingsPath, settings.Preferences{Destination: destination, AutoCopyDownloads: a.SettingsAutoCopyDownloads, VideoOutputFormat: videoOutput, AudioOutputFormat: audioOutput}); err != nil {
		a.SettingsError = err.Error()
		return
	}
	a.Destination = destination
	a.AutoCopyDownloads = a.SettingsAutoCopyDownloads
	a.VideoOutputFormat, a.AudioOutputFormat = videoOutput, audioOutput
	a.SettingsOpen = false
}

func (a *application) chooseFolder() {
	a.Choosing = true
	initial := a.Destination
	if a.SettingsOpen {
		initial = a.SettingsDestination
	}
	win := a.win
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: win, Title: "Choose download folder", DefaultPath: initial, Directory: true, ButtonLabel: "Save here"})
		win.Update(func() {
			a.Choosing = false
			if err != nil {
				a.SettingsError = err.Error()
				return
			}
			if len(paths) > 0 {
				a.SettingsDestination = paths[0]
			}
		})
	}()
}
