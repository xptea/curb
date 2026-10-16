package app

import (
	"context"
	"time"

	"curb/internal/download"
)

func (a *application) loadSupportedSites() {
	if a.SitesLoading || len(a.SupportedSites) > 0 {
		return
	}
	a.SitesLoading, a.SitesError = true, ""
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	a.sitesCancel = cancel
	win := a.win
	go func() {
		defer cancel()
		// Discover independently: startup tool discovery may still be running.
		// A missing FFmpeg or JS runtime does not prevent listing extractors.
		tools, _ := download.LocateTools()
		sites, err := download.SupportedSites(ctx, tools)
		win.Update(func() {
			a.SitesLoading, a.sitesCancel = false, nil
			if err != nil {
				a.SitesError = err.Error()
				return
			}
			a.SupportedSites = sites
		})
	}()
}
