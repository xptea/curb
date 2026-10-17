package app

import (
	"context"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"curb/internal/download"
	view "curb/internal/ui"
	"curb/internal/update"
)

// application owns native window lifetime and mutates UI state on MyGo's
// main thread. Long-running work goes through the download package.
type application struct {
	view.State
	win                   *mygo.Window
	actions               view.Actions
	tools                 download.Tools
	cancel                context.CancelFunc
	copyFile              func(string) error
	sitesCancel           context.CancelFunc
	maximized, fullScreen bool
	settingsPath          string
	updateClient          *update.Client
	availableUpdate       *update.Release
	updateLoopCancel      context.CancelFunc
	updateCheckCancel     context.CancelFunc
	updateInstallCancel   context.CancelFunc
}

func newApplication(version string) *application {
	home, _ := os.UserHomeDir()
	destination := filepath.Join(home, "Downloads")
	if path, err := mygo.App.Path(mygo.PathDownloads); err == nil && path != "" {
		destination = path
	}
	if p := os.Getenv("CURB_DOWNLOAD_DIR"); p != "" {
		destination = p
	}
	a := &application{State: view.State{Version: version, Destination: destination, Format: "Video", Quality: "Best available", AudioQuality: "Best quality", Status: "Ready when you are", Progress: -1}}
	a.loadSettings()
	a.copyFile = func(path string) error { return mygo.Clipboard.WriteFiles(path) }
	a.updateClient = update.NewClient(version)
	a.UpdateError = os.Getenv("CURB_UPDATE_ERROR")
	os.Unsetenv("CURB_UPDATE_ERROR")
	if p := os.Getenv("CURB_DOWNLOAD_DIR"); p != "" {
		a.Destination = p
	}
	a.actions = view.Actions{
		Download: a.startDownload, Cancel: a.cancelDownload,
		ChooseFolder: a.chooseFolder, SaveSettings: a.saveSettings,
		Minimize: func() {
			if a.win != nil {
				a.win.Minimize()
			}
		},
		ToggleMaximize: a.toggleMaximize,
		Close: func() {
			if a.win != nil {
				a.win.Close()
			}
		},
		ShowFile:           mygo.Shell.ShowItemInFolder,
		LoadSupportedSites: a.loadSupportedSites,
		CheckUpdates:       func() { a.checkUpdates(true) }, InstallUpdate: a.installUpdate,
	}
	return a
}

func (a *application) view(c *ui.Context) {
	a.CanDownload = a.canDownload()
	view.Render(c, &a.State, a.actions)
}
