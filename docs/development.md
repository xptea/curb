# Development

Run commands from the repository root. The root `main.go` stays small so `go run .`, `make dev`, and MyGo's configured entry point continue to work.

## Common commands

| Command | Purpose |
| --- | --- |
| `make dev` | Run the app, preferring Wayland on Linux |
| `make dev BACKEND=x11` | Run with GTK's X11 backend |
| `make fmt` | Format Go source |
| `make check` | Check formatting, vet, Go race tests, Python packaging tests, and script syntax |
| `make test-integration` | Test local HTTP/HTTPS downloads, MP3 conversion, merging, and cancellation |
| `make tools` | Prepare verified download tools and build the small FFmpeg/runtime resources |
| `make build` | Compile the application to `build/curb` |
| `make appimage` | Prepare tools and build the Linux AppImage |
| `make bundle-local` | Rebuild the AppImage using existing tools and cached packaging assets |
| `make clean` | Remove generated apps/distributions, retaining tool caches |

`GO` and `PYTHON` can override the build-tool commands, for example `make check GO=/path/to/go`.

The first AppImage build requires network access, Python 3.12+, Docker access, and binutils (`strip`). After tools and packaging assets have been prepared, `make bundle-local` can rebuild offline. The ordinary app build and headless Go tests do not require Docker or downloaded media tools. See [packaging](packaging.md) for the build baseline and platform status.

## UI work

Screens belong in `internal/ui`; shared controls belong in `internal/ui/components`. Update `internal/ui/style` for app-wide colors, corner radii, or icons. Keep native window operations in the application/platform layers and filesystem or subprocess work outside UI building.

After visible changes, check dark and light modes, the settings page, a narrow window, maximize/restore, fullscreen/restore, and live resizing. The title must stay centered against the entire window, even when decorations occupy only one side.

## App icon

`resources/curb.svg` is the editable logo: the supplied cloud/download mark centered on a white rounded tile with transparent corners. `resources/icon.png` is its checked-in 1024×1024 rendering. After editing the SVG, run `python3 scripts/render_icon.py` with librsvg and Cairo installed (`librsvg2-2` and `libcairo2` on Linux). Normal builds use the committed PNG and need no rendering tools.

The PNG is embedded for Linux/Windows window icons and the macOS Dock, including development runs. `mygo.json` selects the same PNG for MyGo's macOS `.icns` and Windows executable/installer resources. The Linux AppImage and desktop installer use the SVG, with `dev.curb.app` as the desktop/window identity.

## Local configuration

Settings are stored in `curb/settings.json` inside the directory returned by Go's `os.UserConfigDir`. On Linux this is normally `$XDG_CONFIG_HOME/curb/settings.json` or `~/.config/curb/settings.json`. `CURB_DOWNLOAD_DIR` overrides the initial destination for the current run.

To keep development settings separate, use a temporary `XDG_CONFIG_HOME` on Linux. Do not commit personal settings, sample downloads, credentials, or generated executables.

## Graphics diagnostics

`MYGO_FRAME_STATS=1 make dev` logs slow frames and identifies the rendering path. `MYGO_GPU=0 make dev` exercises the software fallback. Let MyGo detect hardware acceleration during normal use; forcing OpenGL on a software driver is not evidence of GPU acceleration.

## Dependency updates

Go dependencies are declared in `go.mod`. After an update, run the checks, refresh the corresponding preserved license texts in `licenses/dependencies`, and check `THIRD_PARTY_NOTICES.md`. Tool release pins are separate, in `packaging/tools.json`.
