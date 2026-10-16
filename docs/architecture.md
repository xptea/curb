# Architecture

Curb has one native MyGo window. The executable embeds `mygo.json` and passes its version to `app.Run`; version metadata is not duplicated in UI code.

## Package boundaries

```text
main
  └─ app
      ├─ ui ── components ── style
      ├─ download
      ├─ settings
      └─ platform
```

`internal/app` owns the window, cancellation context, discovered tools, and application state. It connects `ui.Actions` callbacks to downloads, settings persistence, file-manager integration, and native window actions.

`internal/ui` receives `State` and `Actions`. It renders the shell, title bar, download form, feedback, and settings page. It does not read preferences, discover binaries, run subprocesses, or query native window state. Presentation-only changes such as the selected theme or settings page update the shared UI state on MyGo's main thread.

`internal/ui/components` holds reusable controls. `internal/ui/style` is the single home for colors, the shared corner radius, and SVG icons. Use these instead of introducing per-screen copies.

`internal/download` has no MyGo dependency. It discovers tools, validates HTTP(S) URLs, assembles yt-dlp arguments, parses structured output, and runs the download process. Process cancellation is selected by Go's OS-specific filenames/build tags. Linux/macOS cancel the process group; Windows terminates the helper tree so FFmpeg and the JavaScript runtime do not survive a canceled download.

`internal/settings` reads and atomically replaces the per-user JSON preferences file. Only the destination folder is persisted. Malformed or unavailable settings fall back to the app's defaults; errors while saving are shown in the settings screen.

`internal/platform` contains native window appearance setup. Linux makes the GTK backing transparent once; MyGo and the desktop compositor draw the rounded window. No window region is rebuilt during resize.

## Threading and state

MyGo builds the UI on its main thread. Tool discovery, downloads, and native folder dialogs run outside the UI build. Their callbacks change application state through `win.Update`, which returns to the main thread and requests a frame. Download progress callbacks are serialized by the downloader.

Window maximize and fullscreen events update state directly. GTK can emit these events before its queried maximized property changes, so the UI follows the event rather than polling during every frame. Restoring from fullscreen explicitly exits fullscreen.

## Downloads

Video uses the best available media, optionally capped at the selected height. Separate video/audio streams are merged into MKV; a single stream keeps its format. Audio is converted to MP3 at the selected quality. Playlists are disabled, existing files are preserved, and partial files are retained for possible resume. The output template uses the media title and its extension.

Tools are resolved beside the executable, in macOS bundle resources, in the source module's development `tools/` folder, and then `PATH`. Windows filenames use `.exe`. AppImage tooling preserves the shared FFmpeg libraries and yt-dlp's onedir layout.

## Tests

Tests live beside the code they exercise. `internal/app` covers controller wiring and persisted defaults; `internal/ui` covers title placement, resize rendering, navigation, themes, and media/quality controls. `internal/download` covers validation, arguments, progress, and opt-in integration tests using a synthetic media fixture. Set `CURB_INTEGRATION_TOOLS` to test an extracted AppImage's exact toolchain. UI tests use MyGo's headless tester; a desktop is not required for the default suite.
