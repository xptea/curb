# Third-party notices

The MIT license at the repository root applies to Curb's own code and assets. It does not replace the licenses of dependencies, system libraries, or downloaded executables.

## App artwork

The cloud/download mark in `resources/curb.svg` was supplied as `download-cloud-svgrepo-com.svg`, whose metadata credits SVG Repo (https://www.svgrepo.com/). Curb adds the white rounded background, padding, and dark fill; `resources/icon.png` is a rendering of this composition. The original mark is third-party artwork, separate from Curb's MIT code. The supplied SVG does not contain a license declaration; verify the original asset's license before publishing it.

## Go dependencies

These dependencies are compiled into Curb. Their original license texts, including copyright notices, are preserved in `licenses/dependencies/` and copied into portable distributions.

| Dependency | License | Preserved text |
| --- | --- | --- |
| [MyGo](https://github.com/egoist/mygo) | MIT | `mygo.txt` |
| Ebitengine shader compiler included in MyGo | Apache-2.0 | `mygo-shader.txt` |
| [purego](https://github.com/ebitengine/purego) | Apache-2.0 | `purego.txt` |
| [go-text/typesetting](https://github.com/go-text/typesetting) | Unlicense OR BSD-3-Clause | `typesetting.txt` |
| HarfBuzz code in go-text/typesetting | Upstream notices in license file | `harfbuzz.txt` |
| [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) | BSD-3-Clause | `x-image.txt` |
| [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) | BSD-3-Clause | `x-text.txt` |

Exact module versions are recorded in `go.mod` and `go.sum`. Refresh these preserved texts when changing the corresponding dependencies.

MyGo's shader compiler includes code copyright 2020–2026 The Ebiten Authors and The Ebitengine Authors, derived from Ebitengine's `internal/shader` and `internal/shaderir` packages at commit `16f18b0` and modified upstream by MyGo. Curb uses these packages without further changes.

## Separate bundled executables

Portable Linux distributions include yt-dlp, FFmpeg/ffprobe, and QuickJS-NG as separate executables invoked by Curb. Their notices and source information are generated under `tools/licenses/`, and their resolved releases/checksums are recorded in `tools/manifest.json`.

- [yt-dlp](https://github.com/yt-dlp/yt-dlp): retain its license and standalone binary's third-party notices.
- [FFmpeg](https://ffmpeg.org/): built from checksum-pinned official source with a small shared-library configuration and GPL-licensed x264 for H.264 output. This FFmpeg build uses GPL-2.0-or-later; preserve its GPL notice, x264's notices, configuration log, and corresponding source/build materials for the exact distributed build. Curb invokes FFmpeg as a separate executable; Curb's own code remains MIT.
- [QuickJS-NG](https://github.com/quickjs-ng/quickjs): retain its MIT license and notices.

The AppImage includes GTK and dependency libraries from an Ubuntu 22.04 build container, a DejaVu fallback font, and the official AppImage type-2 runtime. Their copyright notices, license texts, and version manifests are included under `usr/share/doc/curb/` inside the AppImage. GPU drivers and glibc remain supplied by the host. See [packaging](docs/packaging.md) for redistribution and source/build requirements.
