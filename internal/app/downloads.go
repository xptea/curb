package app

import (
	"context"
	"errors"

	"curb/internal/download"
)

func (a *application) startDownload() {
	if a.Busy || a.Updating {
		return
	}
	link, err := download.ValidateURL(a.URL)
	if err != nil {
		a.Failure = err.Error()
		return
	}
	if a.tools.Downloader == "" {
		a.Failure = "The download engine is starting. Try again in a moment."
		return
	}
	if a.tools.FFmpeg == "" {
		a.Failure = "FFmpeg is missing. Prepare the bundled tools first."
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel, a.Busy, a.Failure, a.SavedPath = cancel, true, "", ""
	a.CopiedToClipboard, a.ClipboardError = false, ""
	a.Status, a.Detail, a.Progress = "Connecting to your link…", "", -1
	request := download.Request{URL: link, Destination: a.Destination, Format: a.Format, Quality: a.Quality, AudioQuality: a.AudioQuality}
	win := a.win
	go func() {
		path, err := download.Run(ctx, a.tools, request, func(u download.Update) {
			win.Update(func() {
				if u.Path == "" {
					a.Status, a.Detail, a.Progress = u.Status, u.Detail, u.Progress
				}
			})
		})
		cancel()
		win.Update(func() {
			a.finishDownload(path, err)
		})
	}()
}

// finishDownload runs on the window's main thread after the final media file
// exists. Offer a native file reference, never the URL, path text, or file bytes.
func (a *application) finishDownload(path string, err error) {
	a.Busy, a.cancel = false, nil
	a.CopiedToClipboard, a.ClipboardError = false, ""
	if errors.Is(err, context.Canceled) {
		a.Status = "Download canceled"
		return
	}
	if err != nil {
		a.Failure = err.Error()
		return
	}
	a.SavedPath, a.Status, a.Progress = path, "Saved", 1
	if a.AutoCopyDownloads {
		if err := a.copyFile(path); err != nil {
			a.ClipboardError = err.Error()
		} else {
			a.CopiedToClipboard = true
		}
	}
}

func (a *application) canDownload() bool {
	if a.Busy || a.Updating || a.tools.Downloader == "" || a.tools.FFmpeg == "" {
		return false
	}
	_, err := download.ValidateURL(a.URL)
	return err == nil
}

func (a *application) cancelDownload() {
	if a.cancel != nil {
		a.cancel()
		a.Status = "Canceling…"
	}
}
