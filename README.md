# Curb

A minimal native video and audio downloader. Paste a link, choose video or MP3 audio and a quality, then save the file locally.

Curb uses [MyGo](https://mygo.egoist.dev/) for its Go UI and [yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org/), and [QuickJS-NG](https://github.com/quickjs-ng/quickjs) for downloads and media processing. It uses no webview or frontend server.

## Features

- Centered URL input with a video/audio switch and a quality selector.
- Dark and light themes, custom window controls on Linux and Windows, and native macOS traffic lights.
- Full-page settings for the default save folder.
- Searchable supported-site list from the installed yt-dlp, available offline.
- Download progress, cancellation, and a shortcut to show the saved file.
- A Linux AppImage with GTK and download tools included.
- Update checks with a green download button when a newer desktop release is available.

Linux is the packaged and manually tested platform. macOS and Windows are checked for compilation; their release packaging and runtime behavior are not yet validated.

## Development

The separate React + Tailwind landing page lives in [website/](website/README.md). It builds to static HTML, CSS, and JavaScript for Cloudflare Workers Static Assets. It is independent of the native desktop app and its release builds.

Install Go **1.27.1+**, Python **3.12+** and Docker for AppImage packaging, and GTK 3 with its graphics and font libraries for development on Linux. On Ubuntu 24.04:

```sh
sudo apt-get install libgtk-3-0t64 libepoxy0 fonts-dejavu-core binutils
```

From the repository root:

```sh
go mod download
make tools
make dev
```

`make tools` prepares verified yt-dlp and QuickJS-NG downloads and builds a small FFmpeg toolchain in an Ubuntu 22.04 container, placing generated resources in the ignored `tools/` directory. You can instead use installed `yt-dlp`, `ffmpeg`, and a supported JavaScript runtime from `PATH` during development. QuickJS-NG is preferred; Deno and Node are supported fallbacks.

`make dev` prefers native Wayland on Linux with X11 as a fallback. Use `make dev BACKEND=x11` to test X11 explicitly. MyGo automatically uses hardware GPU rendering when available.

Downloads initially go to the operating system's Downloads folder. Change this in Settings or set `CURB_DOWNLOAD_DIR=/absolute/folder` for a development run. Filenames come from the media title, with the file extension added automatically.

Enable **Automatically copy downloaded files** in Settings to put the finished video or audio file on the system clipboard, ready to paste. This is off by default and remembered between launches. A clipboard error does not affect the saved download.

## Build and check

```sh
make check             # formatting, vet, Go/Python tests, packaging syntax
make build             # build/curb
make test-integration  # real local downloads using prepared tools
make appimage          # prepare tools and build the Linux AppImage
make bundle            # alias for the AppImage build
make bundle-local      # rebuild with prepared tools; first run may cache packaging assets
```

The integration tests serve a small synthetic media fixture over loopback HTTP. They do not need a public video URL, site account, or production video encoder.

To use a Linux build, make `dist/Curb-0.1.0-x86_64.AppImage` executable and run it. GTK and the download tools are included. No Go, Python installation, Docker, or Debian package is required on the target machine. The runtime baseline is Ubuntu 22.04 / glibc 2.35; see [packaging](docs/packaging.md) for compatibility and size choices.

## Repository layout

```text
main.go                  Embedded version metadata and executable entry point
internal/
  app/                   Window lifetime, download actions, settings coordination
  download/              Tool discovery, yt-dlp arguments, progress, subprocesses
  platform/              OS-specific native window setup
  settings/              JSON preferences and atomic persistence
  ui/                    Shell, title bar, download and settings screens
    components/          Shared buttons, tooltips, media switch, quality selector
    style/               Theme colors, corner radius, SVG icons
packaging/
  linux/                 AppImage launcher, GTK baseline, FFmpeg recipe, desktop entry
  appimage.json          AppImage packager/runtime release pins
  tools.json             Download-tool release pins
scripts/
  bundle.py              Packaging CLI
  bundler/               Verified tools, small FFmpeg build, AppImage assembly
resources/               App icon
licenses/                Preserved dependency license texts
.github/                 PR checks, native tag releases, and issue/PR templates
```

See [architecture](docs/architecture.md), [development](docs/development.md), and [packaging](docs/packaging.md) for implementation details. Contributions are welcome; start with [CONTRIBUTING.md](CONTRIBUTING.md).

Push a stable tag such as `v0.1.1` to build Linux, Windows, and Intel/Apple Silicon Mac releases. Ordinary branch pushes do not run CI. See [desktop updates](docs/updates.md) for the native build jobs, GitHub Raw version feed, restart/rollback behavior, and publishing workflow.

## Supported sites and responsible use

Open **Supported sites** in the bottom-right corner for the available handlers in your installed yt-dlp. The list includes separate video, channel, and playlist handlers, and hides entries marked as currently broken. It does not guarantee that every listed site or content type works. Curb downloads individual items; listing a playlist handler does not enable batch playlist downloads.

yt-dlp also maintains a [public supported-sites list](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md). That list tracks upstream development and can differ from the bundled version. Some unlisted links work through generic or embedded media extraction.

Only download when permitted by the rights holder, applicable law, and the site's terms. Users are responsible for their use and any account restrictions. Curb is not affiliated with, endorsed by, or sponsored by the listed platforms. Availability and access are controlled by those platforms.

For example, [YouTube's terms](https://www.youtube.com/static?template=terms) restrict downloading and automated access except where authorized, and allow account suspension or termination for certain breaches. Curb does not guarantee protection from account restrictions. This notice does not grant downloading rights or override a platform's terms.

## License

Curb's own code and assets are available under the [MIT license](LICENSE). Dependencies and bundled executables retain their own licenses; see [third-party notices](THIRD_PARTY_NOTICES.md).
