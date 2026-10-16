// Package view renders Curb's native UI. It depends on state and callbacks,
// never on tool discovery, subprocesses, or settings files.
package view

type State struct {
	Version                                         string
	URL, Destination, Format, Quality, AudioQuality string
	Busy, Choosing, CanDownload                     bool
	LightMode, EdgeToEdge, SettingsOpen             bool
	SettingsDestination, SettingsError              string
	Status, Detail, Failure, SavedPath              string
	Progress                                        float64
	SupportedSitesOpen, SitesLoading                bool
	SupportedSites                                  []string
	SitesQuery, SitesError                          string
	UpdateVersion, UpdateError                      string
	UpdateChecking, Updating                        bool
	UpdateProgress                                  float64
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
