// Package view renders Curb's native UI. It depends on state and callbacks,
// never on tool discovery, subprocesses, or settings files.
package view

import "time"

type State struct {
	Version                                              string
	URL, Destination, Format, Quality, AudioQuality      string
	Busy, Choosing, CanDownload                          bool
	LightMode, EdgeToEdge, SettingsOpen                  bool
	SettingsDestination, SettingsError                   string
	AutoCopyDownloads, SettingsAutoCopyDownloads         bool
	VideoOutputFormat, AudioOutputFormat                 string
	SettingsVideoOutputFormat, SettingsAudioOutputFormat string
	CopiedToClipboard                                    bool
	ClipboardError                                       string
	SavedUntil                                           time.Time
	Status, Detail, Failure, SavedPath                   string
	Progress                                             float64
	SupportedSitesOpen, SitesLoading                     bool
	SupportedSites                                       []string
	SitesQuery, SitesError                               string
	UpdateVersion, UpdateError                           string
	UpdateChecking, Updating                             bool
	UpdateProgress                                       float64
}

// ClearSavedFeedback dismisses the confirmation without changing the saved
// file or the system clipboard.
func (s *State) ClearSavedFeedback() {
	s.SavedPath, s.ClipboardError = "", ""
	s.CopiedToClipboard = false
	s.SavedUntil = time.Time{}
	if !s.Busy && s.Status == "Saved" {
		s.Status, s.Detail, s.Progress = "Ready when you are", "", -1
	}
}

// Actions connects user interaction to the application controller.
type Actions struct {
	Download, Cancel, ChooseFolder, SaveSettings func()
	Minimize, ToggleMaximize, Close              func()
	ShowFile                                     func(string)
	LoadSupportedSites                           func()
	CheckUpdates, InstallUpdate                  func()
}

type screen struct {
	*State
	actions Actions
}
