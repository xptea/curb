package app

import (
	"log"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"curb/internal/download"
	"curb/internal/platform"
)

func Run(version string, icon []byte) error {
	platform.PrepareApplication()
	mygo.App.SetName("Curb")
	mygo.App.SetVersion(version)
	app := newApplication(version)
	openWindow := func() {
		if app.win != nil && !app.win.IsDestroyed() {
			app.win.Show()
			app.win.Focus()
			return
		}
		app.maximized, app.fullScreen, app.EdgeToEdge = false, false, false
		app.Busy, app.SitesLoading, app.Choosing = false, false, false
		app.win = mygo.NewWindow(mygo.WindowOptions{
			Title: "Curb", Width: 900, Height: 640, MinWidth: 640, MinHeight: 540,
			TitleBarStyle: mygo.TitleBarHidden, TitleBarHeight: 46,
			Frameless:   runtime.GOOS != "darwin",
			Transparent: true, Content: ui.View(app.view),
		})
		platform.ConfigureWindow(app.win)
		if runtime.GOOS != "darwin" {
			if err := app.win.SetIcon(icon); err != nil {
				log.Printf("Window icon: %v", err)
			}
		}
		// GTK emits window-state-event before its own maximized property
		// changes. Use the event itself rather than querying the old state.
		app.win.OnMaximize(func() { app.maximized = true; app.syncWindowState() })
		app.win.OnUnmaximize(func() { app.maximized = false; app.syncWindowState() })
		app.win.OnEnterFullScreen(func() { app.fullScreen = true; app.syncWindowState() })
		app.win.OnLeaveFullScreen(func() { app.fullScreen = false; app.syncWindowState() })
		app.win.OnClosed(func() {
			if app.sitesCancel != nil {
				app.sitesCancel()
			}
			if app.cancel != nil {
				app.cancel()
			}
			app.cancel, app.sitesCancel = nil, nil
		})
		win := app.win
		go func() {
			tools, err := download.LocateTools()
			win.Update(func() {
				app.tools = tools
				if err != nil {
					app.Failure = err.Error()
				}
			})
		}()
	}
	mygo.App.WhenReady(func() {
		if runtime.GOOS == "darwin" {
			if err := mygo.App.Dock.SetIcon(icon); err != nil {
				log.Printf("Dock icon: %v", err)
			}
		}
		openWindow()
	})
	if runtime.GOOS == "darwin" {
		// Keep the Dock app alive after its window closes; Cmd+Q still quits.
		mygo.App.OnWindowAllClosed(func() {})
		mygo.App.OnActivate(func(hasVisibleWindows bool) {
			if !hasVisibleWindows {
				openWindow()
			}
		})
	}
	return mygo.App.Run()
}

func (a *application) syncWindowState() {
	a.EdgeToEdge = a.maximized || a.fullScreen
	if a.win != nil {
		a.win.Invalidate()
	}
}

func (a *application) toggleMaximize() {
	if a.win == nil {
		return
	}
	if a.fullScreen {
		a.win.SetFullScreen(false)
	} else {
		a.win.ToggleMaximize()
	}
}
