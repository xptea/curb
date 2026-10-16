package app

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/egoist/mygo"

	"curb/internal/update"
)

func (a *application) beginUpdateChecks() {
	a.UpdateChecking, a.Updating = false, false
	ctx, cancel := context.WithCancel(context.Background())
	a.updateLoopCancel = cancel
	a.checkUpdates(false)
	win := a.win
	go func() {
		tick := time.NewTicker(6 * time.Hour)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				win.Update(func() { a.checkUpdates(false) })
			}
		}
	}()
}

func (a *application) stopUpdateChecks() {
	for _, cancel := range []context.CancelFunc{a.updateLoopCancel, a.updateCheckCancel, a.updateInstallCancel} {
		if cancel != nil {
			cancel()
		}
	}
}

func (a *application) checkUpdates(manual bool) {
	if a.UpdateChecking || a.Updating {
		return
	}
	a.UpdateChecking = true
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	a.updateCheckCancel = cancel
	win := a.win
	go func() {
		release, err := a.updateClient.Check(ctx)
		cancel()
		win.Update(func() {
			a.UpdateChecking, a.updateCheckCancel = false, nil
			if err != nil {
				log.Printf("Update check: %v", err)
				if manual {
					a.UpdateError = "Could not check for updates. Click the version to retry."
				}
				return
			}
			a.availableUpdate, a.UpdateVersion = release, ""
			if manual {
				a.UpdateError = ""
			}
			if release != nil {
				a.UpdateVersion = release.Version
			}
		})
	}()
}

func (a *application) installUpdate() {
	if a.Updating || a.Busy || a.Choosing || a.availableUpdate == nil {
		return
	}
	installation, err := update.DetectInstallation()
	if err != nil {
		a.UpdateError = err.Error()
		return
	}
	a.Updating, a.UpdateProgress, a.UpdateError = true, 0, ""
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	a.updateInstallCancel = cancel
	win, release := a.win, *a.availableUpdate
	go func() {
		plan, err := a.updateClient.Prepare(ctx, release, installation, func(done, total int64) {
			win.Update(func() { a.UpdateProgress = float64(done) / float64(total) })
		})
		if err == nil {
			if ctx.Err() != nil {
				err = ctx.Err()
			} else {
				err = update.StartHelper(plan)
			}
			if err != nil {
				_ = os.RemoveAll(plan.Work)
			}
		}
		cancel()
		win.Update(func() {
			a.updateInstallCancel = nil
			if err != nil {
				a.Updating, a.UpdateError = false, err.Error()
				return
			}
			// The helper is running and waiting for this process to exit.
			// It replaces the app, opens the new version, and removes the old one.
			mygo.App.Quit()
		})
	}()
}
