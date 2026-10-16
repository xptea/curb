# Linux AppImage packaging

The default Linux release is one executable AppImage. Build dependencies are Go 1.27.1+, Python 3.12+, and Docker. None is required on the target desktop.

```sh
make tools         # verified tools and small FFmpeg build
make appimage      # build AppImage; also available as make bundle
make bundle-local  # reuse prepared tools and cached packaging assets
```

The output is `dist/Curb-<version>-<architecture>.AppImage` with a SHA-256 file. Make it executable and run it. The x86_64 build is the tested release target; native aarch64 packaging is available but not validated. The runtime is collected from an Ubuntu 22.04 container, with glibc 2.35 as the baseline. The host supplies glibc, the display server, and GPU drivers. Compatibility still needs checking on each release target.

## Size choices

Curb's UI is native Go. It includes no webview or JavaScript frontend engine. yt-dlp's official onedir ZIP is verified and unpacked before packaging so SquashFS compresses it once. Development/debug symbols are stripped from ELF shared libraries, retaining exported runtime symbols and every dependency. The PyInstaller executable and its appended Python archive are left intact. Install binutils on the build host for this step. The upstream archive digest and stripping method are recorded in the tool manifest. No Python installation is needed, and downloads do not extract an embedded Python runtime on every invocation.

FFmpeg and ffprobe share a small set of libraries built from the pinned official source. `packaging/linux/build-ffmpeg.sh` preserves stream copying/remuxing, common audio decoders, HTTPS input, and LAME MP3 encoding. It excludes video encoders, capture devices, hardware acceleration, and general video filters. Curb downloads video without transcoding. Unusual audio codecs outside the configured decoders may fail to convert.

QuickJS-NG runs JavaScript needed by yt-dlp's site extraction. GTK, its runtime dependencies, and a fallback font are included from the older build container. Their package/source versions and notices are preserved. The official AppImage runtime and packager are release-pinned in `packaging/appimage.json` and verified using published SHA-256 digests. AppImage compression uses Zstandard.

Generated resources live in `tools/`, `build/downloads/`, `build/lean-ffmpeg/`, `build/linux-runtime/`, `build/Curb.AppDir/`, and `dist/`. None belongs in source commits. An optional legacy tar staging output is available with `python3 scripts/bundle.py --format tar`; it requires system GTK and media dependencies and is not the default release format.

## Validation and release

Run `make check` and `make test-integration`, then test the actual AppImage from outside the repository. Check video stream merging, MP3 conversion, cancellation, resizing, fullscreen, settings, and the site list. Integration tests use a checked-in synthetic media fixture, without requiring a video encoder or public video URL.

The standard `--appimage-extract` and `--appimage-extract-and-run` options are available; the latter works without FUSE mounting. `AppRun` prefers Wayland with X11 fallback, sets production mode, and uses the bundled GTK/font resources. Settings and downloads stay outside the read-only image. Shared libraries can be replaced in an extracted AppImage.

Preserve notices and provide corresponding source/build materials for the exact LGPL/GPL components as required by their licenses before publishing a release. Links alone do not replace those obligations. The FFmpeg source digest and full configure log are preserved; the Linux manifest lists exact source-package versions. The app's own code remains MIT. No build script publishes releases automatically.

## Windows and macOS

The app icon is configured through `icon` in `mygo.json`. MyGo converts the master PNG into a multi-resolution `AppIcon.icns` referenced by the macOS bundle's `Info.plist`, and embeds the corresponding icon group in Windows executables and installers. Plain `go build` sets the running window's icon but does not create macOS bundles or Windows executable resources; use MyGo's packaging command for those:

```sh
go run github.com/egoist/mygo/cmd/mygo build -platform darwin/universal -skip-dmg
go run github.com/egoist/mygo/cmd/mygo build -platform windows/amd64
```

These commands must use tools prepared for the target platform. The shared `tools/` resource in the current config is prepared by the Linux bundler; replace it with native resources before making macOS/Windows releases. The icon/bundle checks can omit this resource to validate branding without incorrectly shipping Linux executables. macOS disk-image creation and signing require macOS; Windows installer creation requires NSIS.

MyGo's native event loop and window events are used. Windows tool discovery recognizes `.exe`; macOS resources are found in `Contents/Resources/tools`. macOS retains its Dock app after the window closes and reopens the window on activation; Cmd+Q quits normally. Native download folders are used, and cancellation includes child tools on each desktop platform.

Cross-compilation covers Windows x64 and macOS Intel/Apple Silicon. This does not verify runtime behavior. Those platforms still need their own tool binaries, packaging, signing/notarization as appropriate, and tests on the actual operating systems. Linux tools must not be copied into those releases.
