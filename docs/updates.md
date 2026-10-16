# Desktop updates

Curb checks `https://raw.githubusercontent.com/xptea/curb/main/version.json` in the background at startup and every six hours. Clicking the version checks again. It never downloads an update in the background without a click.

When a newer stable release exists for the running OS/CPU, the version becomes a green **Update to v…** button with a download icon. Clicking it shows a spinner and percentage, downloads the matching GitHub Release asset, verifies its exact size and SHA-256, and stages it next to the installation. Updating waits until media downloads and folder selection finish. Errors appear above the version button; a failed download leaves the running app untouched.

The staged helper waits for the original process to exit, replaces the installation, and opens the new version. The old files are kept until the new app acknowledges opening its window. A launch failure restores and reopens the previous version. After success, the helper removes its backup and staging files. Windows uses a short-lived cleanup process because a running executable cannot delete itself. The installation path stays the same so existing shortcuts continue to work; settings and downloaded media live outside it.

## Supported installations

- Linux: the actual AppImage identified by the runtime's `APPIMAGE` environment variable. The mounted executable is never patched. Relaunch supports extraction without FUSE.
- macOS: a writable `.app` bundle, replaced as a unit, including its native download tools and resources. Updates use a `.tar.gz` containing one `Curb.app`; relative internal bundle symlinks and executable permissions are preserved.
- Windows: a writable portable Curb directory containing its managed-install marker and file inventory. Updates use a `.zip` containing one `Curb` folder, including `Curb.exe` and its native tools. Updates refuse a folder containing extra files, preserving unrelated user data. The updater does not replace an arbitrary directory containing a standalone development executable.

Read-only installs and development binaries cannot install updates. Move a release into a user-writable location first. Windows/macOS compile and archive behavior are checked here, but full native OS update tests are still required before distributing those releases. Do not include Linux tool binaries in either platform's package.

## Release process

Normal branch pushes do not run CI. The separate **CI** workflow runs only for pull requests or a manual request. **Build release** runs when a stable `vMAJOR.MINOR.PATCH` tag is pushed, when a stable GitHub Release is published, or when it is manually dispatched with an existing tag.

Push the workflow/code changes first, then tag the commit you want to ship:

```sh
git tag v0.1.1
git push origin v0.1.1
```

The tag is the release version. Each isolated runner updates `mygo.json` to that version before compiling; you do not need to hand-edit the build version first. `version.json` keeps describing the previous published release until the new builds succeed.

**Build release** runs four native jobs:

| Target | Runner | Release asset |
| --- | --- | --- |
| Linux x64 | Ubuntu 24.04, Ubuntu 22.04 runtime container | `Curb-<version>-x86_64.AppImage` |
| Windows x64 | Windows 2025, MSYS2 UCRT64 | `Curb-<version>-windows-amd64.zip` |
| macOS Intel | macOS 15 Intel | `Curb-<version>-darwin-amd64.tar.gz` |
| macOS Apple Silicon | macOS 15 ARM64 | `Curb-<version>-darwin-arm64.tar.gz` |

Each builds its own lean shared FFmpeg from the checksum-pinned source, prepares platform-specific yt-dlp/QuickJS, bundles dependency libraries and notices, and runs tests with the packaged toolchain. Mac/Windows also launch the real app and require its first-frame startup receipt. Mac release builds currently require macOS 15 or newer, matching their build/dependency baseline.

Only after **all four jobs pass**, the publisher creates or completes the GitHub Release, uploads the packages/checksums, publishes it, and updates `version.json` on `main`. It also advances `mygo.json` on `main` if its version is older. A failed platform leaves the update feed unchanged. Published binaries are never silently overwritten with different bytes. A duplicate tag/release event skips rebuilding a release that already has all four packages and can repair a missed feed commit.

Publishing needs repository `contents: write`. Branch protection must allow the metadata commit, or you must apply the generated feed manually. The build jobs and normal CI remain read-only. The separate **Publish update feed** workflow is a manual repair tool; it no longer races a newly published release while builds are still running. The release job updates the feed directly because releases created using GitHub's job token do not trigger another workflow.

Mac builds are ad-hoc signed; Developer ID signing/notarization is not configured. Windows builds are unsigned portable packages. Trusted distribution signing needs the appropriate certificates and signing setup before a public launch. The workflow contains no signing credentials.

## Manual builds and feed repair

`make appimage` still builds the native Linux asset locally. On Mac/Windows, install the build dependencies listed in `release.yml`, then:

```sh
python3 scripts/release.py prepare --tag v0.1.1
python3 scripts/release.py build-native --platform darwin-arm64
# On Windows, in the configured MSYS2 UCRT64 shell:
python scripts/release.py build-native --platform windows-amd64
```

These commands change the local build version/resources and prepare native tools. They do not publish anything. `scripts/package_desktop_update.py` can separately archive an existing native bundle or portable app. It preserves a Mac bundle's signature and adds the Windows managed-file inventory; generated installers are excluded from the portable ZIP.

To repair the feed for an already-published release, manually run **Publish update feed** with its tag, or:

```sh
gh api repos/xptea/curb/releases/tags/v0.1.1 > build/release.json
python3 scripts/update_manifest.py --release-json build/release.json
```

The initial feed has no platform assets until the first release is published. Never invent checksums or advertise assets before uploading them. Reuse neither version numbers nor release tags for different builds. Feed/downloads use HTTPS and stay on GitHub; checksums come from the repository-controlled raw feed. Repository/release write access therefore also controls the software trusted by installed apps.

## Validation

`make check` includes platform selection/version comparison, checksum/truncation rejection, extraction traversal/link checks, subprocess handoff, rollback/reopen, and cleanup tests. Network tests use local fixtures and never replace the development checkout. Test real release-to-release updates on each supported OS, including a failed download, an unwritable installation, and a failed new launch, before publishing to users.
