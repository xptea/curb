"""Prepare verified yt-dlp, FFmpeg, and QuickJS-NG tool resources."""
import json
import platform
import shutil
import subprocess
import zipfile

from .network import asset, checksum, download, fetch, release
from .binaries import strip_shared_libraries
from .paths import CACHE, MYGO_VERSION, TOOLS, TOOL_VERSIONS


def _prepare_ytdlp():
    info = release("yt-dlp/yt-dlp", TOOL_VERSIONS["yt-dlp"])
    name = "yt-dlp_linux.zip" if platform.machine() == "x86_64" else "yt-dlp_linux_aarch64.zip"
    sums_url = asset(info, "SHA2-256SUMS")["browser_download_url"]
    digest = checksum(fetch(sums_url).decode(), name)
    binary = download(asset(info, name)["browser_download_url"], name, digest)
    runtime = TOOLS / "yt-dlp-runtime"
    if runtime.exists():
        shutil.rmtree(runtime)
    runtime.mkdir()
    with zipfile.ZipFile(binary) as archive:
        for member in archive.infolist():
            target = runtime / member.filename
            if not target.resolve().is_relative_to(runtime.resolve()):
                raise RuntimeError("Unsafe yt-dlp archive path")
            if member.is_dir():
                target.mkdir(parents=True, exist_ok=True)
            else:
                target.parent.mkdir(parents=True, exist_ok=True)
                with archive.open(member) as source, target.open("wb") as output:
                    shutil.copyfileobj(source, output)
                target.chmod((member.external_attr >> 16) & 0o777 or 0o644)
    executable = "yt-dlp_linux" if platform.machine() == "x86_64" else "yt-dlp_linux_aarch64"
    (runtime / executable).chmod(0o755)
    (TOOLS / "yt-dlp").unlink(missing_ok=True)
    (TOOLS / "yt-dlp").symlink_to(f"yt-dlp-runtime/{executable}")
    stripping = strip_shared_libraries(runtime)
    return {"version": info["tag_name"], "sha256": digest, "source": info["html_url"], "layout": "onedir", "stripping": stripping}


def _prepare_ffmpeg(notices):
    from .lean_ffmpeg import prepare
    source = prepare()
    for notice in notices.glob("ffmpeg-*"):
        notice.unlink()
    for directory in ("bin", "lib"):
        if (TOOLS / directory).exists():
            shutil.rmtree(TOOLS / directory)
        shutil.copytree(source / directory, TOOLS / directory, symlinks=True,
                        ignore=shutil.ignore_patterns("pkgconfig", "*.a"))
    for executable in ("ffmpeg", "ffprobe"):
        launcher = TOOLS / executable
        launcher.unlink(missing_ok=True)
        launcher.write_text(
            '#!/bin/sh\nset -eu\n'
            'tool_dir=$(CDPATH= cd -- "${0%/*}" && pwd)\n'
            'export LD_LIBRARY_PATH="$tool_dir/lib${CURB_RUNTIME_LIBDIR:+:$CURB_RUNTIME_LIBDIR}${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"\n'
            f'exec "$tool_dir/bin/{executable}" "$@"\n'
        )
        launcher.chmod(0o755)
    shutil.copy2(source / "FFmpeg-LICENSE", notices / "ffmpeg-LICENSE")
    shutil.copy2(source / "config.log", notices / "ffmpeg-config.log")
    return {
        "sha256": TOOL_VERSIONS["ffmpeg-source-sha256"],
        "source": f"https://ffmpeg.org/releases/ffmpeg-{TOOL_VERSIONS['ffmpeg']}.tar.xz",
        "configuration": "packaging/linux/build-ffmpeg.sh",
        "layout": "lean-shared",
    }


def _prepare_quickjs():
    info = release("quickjs-ng/quickjs", TOOL_VERSIONS["quickjs-ng"])
    architecture = "x86_64" if platform.machine() == "x86_64" else "aarch64"
    name = f"qjs-linux-{architecture}"
    binary_asset = asset(info, name)
    digest = binary_asset.get("digest", "")
    if not digest.startswith("sha256:"):
        raise RuntimeError("QuickJS release asset has no published SHA-256 digest")
    digest = digest.removeprefix("sha256:")
    binary = download(binary_asset["browser_download_url"], name, digest)
    shutil.copy2(binary, TOOLS / "qjs")
    return {"version": info["tag_name"], "sha256": digest, "source": info["html_url"]}


def _write_notices(notices, metadata):
    yt_version = metadata["yt-dlp"]["version"]
    qjs_version = metadata["quickjs-ng"]["version"]
    sources = [
        ("yt-dlp/yt-dlp", yt_version, "LICENSE", "yt-dlp-LICENSE"),
        ("yt-dlp/yt-dlp", yt_version, "THIRD_PARTY_LICENSES.txt", "yt-dlp-THIRD_PARTY_LICENSES.txt"),
        ("quickjs-ng/quickjs", qjs_version, "LICENSE", "quickjs-LICENSE"),
        ("egoist/mygo", MYGO_VERSION, "LICENSE", "mygo-LICENSE"),
    ]
    for repo, ref, name, filename in sources:
        url = f"https://raw.githubusercontent.com/{repo}/{ref}/{name}"
        (notices / filename).write_bytes(fetch(url))
    (notices / "GPL-3.0.txt").write_bytes(fetch("https://www.gnu.org/licenses/gpl-3.0.txt"))
    (notices / "SOURCES.txt").write_text(
        "Bundled executables are separate programs invoked by Curb.\n"
        f"yt-dlp: https://github.com/yt-dlp/yt-dlp/tree/{yt_version}\n"
        f"QuickJS-NG: https://github.com/quickjs-ng/quickjs/tree/{qjs_version}\n"
        f"FFmpeg source: {metadata['ffmpeg']['source']}\n"
        "Build configuration and exact FFmpeg revision: run tools/ffmpeg -version\n"
        "FFmpeg build recipe: packaging/linux/build-ffmpeg.sh in Curb's source\n"
        "FFmpeg upstream source: https://git.ffmpeg.org/ffmpeg.git\n"
    )


def prepare_tools():
    TOOLS.mkdir(exist_ok=True)
    CACHE.mkdir(parents=True, exist_ok=True)
    notices = TOOLS / "licenses"
    notices.mkdir(exist_ok=True)
    metadata = {
        "yt-dlp": _prepare_ytdlp(),
        "ffmpeg": _prepare_ffmpeg(notices),
        "quickjs-ng": _prepare_quickjs(),
    }
    versions = {}
    for name in ("yt-dlp", "ffmpeg", "ffprobe", "qjs"):
        (TOOLS / name).chmod(0o755)
        argument = {"yt-dlp": "--version", "qjs": "--version"}.get(name, "-version")
        output = subprocess.check_output([str(TOOLS / name), argument], text=True)
        versions[name] = output.splitlines()[0]
    metadata["ffmpeg"]["version"] = versions["ffmpeg"]
    _write_notices(notices, metadata)
    (TOOLS / "manifest.json").write_text(json.dumps(metadata, indent=2) + "\n")
    # Remove the retired runtime when updating an existing development bundle.
    (TOOLS / "deno").unlink(missing_ok=True)
    (notices / "deno-LICENSE.md").unlink(missing_ok=True)
