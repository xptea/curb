"""Prepare pinned download tools and lean shared FFmpeg on macOS/Windows runners."""
import hashlib
import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import tarfile

from .network import asset, asset_checksum, download, fetch, release
from .paths import ROOT, CACHE, TOOLS, TOOL_VERSIONS
from .native_runtime import collect_macos, collect_windows, command


def native_build_id(target):
    inputs = ("packaging/tools.json", "packaging/native/build-ffmpeg.sh", "scripts/bundler/native_tools.py",
              "scripts/bundler/native_runtime.py", "scripts/bundler/network.py")
    return hashlib.sha256(target.encode() + b"".join((ROOT / path).read_bytes() for path in inputs)).hexdigest()


def ffmpeg_shell(windows):
    if not windows:
        return "/bin/bash"
    # CreateProcess can find System32/bash.exe (the WSL launcher) before
    # searching PATH. Resolve MSYS2's actual Bash and launch it by full path.
    shell = Path(command("cygpath", "-w", "/usr/bin/bash.exe"))
    if not shell.is_absolute() or not shell.is_file():
        raise RuntimeError(f"MSYS2 Bash is missing: {shell}; run this build in the MSYS2 UCRT64 shell")
    return str(shell)


def prepare_native(target):
    expected = "windows-amd64" if os.name == "nt" else "darwin-" + ("arm64" if platform.machine() == "arm64" else "amd64")
    if target != expected or not (os.name == "nt" or platform.system() == "Darwin"):
        raise RuntimeError(f"Build {target} on its matching native runner")
    stamp = TOOLS / "curb-build-id"
    suffix = ".exe" if target.startswith("windows-") else ""
    build_id = native_build_id(target)
    required = [TOOLS / (name + suffix) for name in ("yt-dlp", "qjs", "ffmpeg", "ffprobe")]
    required.extend((TOOLS / "manifest.json", TOOLS / "licenses/ffmpeg-LICENSE"))
    if stamp.is_file() and stamp.read_text() == build_id and all(path.is_file() for path in required):
        print(f"Reusing prepared {target} download tools and FFmpeg.", flush=True)
        return
    CACHE.mkdir(parents=True, exist_ok=True)
    if TOOLS.exists():
        shutil.rmtree(TOOLS)
    TOOLS.mkdir()
    notices = TOOLS / "licenses"
    notices.mkdir()
    windows = target.startswith("windows-")
    shell = ffmpeg_shell(windows)
    suffix = ".exe" if windows else ""
    metadata = {}

    info = release("yt-dlp/yt-dlp", TOOL_VERSIONS["yt-dlp"])
    name = "yt-dlp.exe" if windows else "yt-dlp_macos"
    digest = asset_checksum(info, name)
    shutil.copy2(download(asset(info, name)["browser_download_url"], name, digest), TOOLS / ("yt-dlp" + suffix))
    metadata["yt-dlp"] = {"version": info["tag_name"], "sha256": digest, "source": info["html_url"]}

    info = release("quickjs-ng/quickjs", TOOL_VERSIONS["quickjs-ng"])
    name = "qjs-windows-x86_64.exe" if windows else "qjs-darwin-" + ("arm64" if target.endswith("arm64") else "x86_64")
    item = asset(info, name)
    if not item.get("digest", "").startswith("sha256:"):
        raise RuntimeError("QuickJS release has no SHA-256 digest")
    digest = item["digest"][7:]
    shutil.copy2(download(item["browser_download_url"], name, digest), TOOLS / ("qjs" + suffix))
    metadata["quickjs-ng"] = {"version": info["tag_name"], "sha256": digest, "source": info["html_url"]}

    version = TOOL_VERSIONS["ffmpeg"]
    archive = download(f"https://ffmpeg.org/releases/ffmpeg-{version}.tar.xz", f"ffmpeg-{version}.tar.xz", TOOL_VERSIONS["ffmpeg-source-sha256"])
    source = ROOT / "build" / f"ffmpeg-{version}"
    if not source.exists():
        with tarfile.open(archive) as data:
            data.extractall(ROOT / "build", filter="data")
    prefix = ROOT / "build/native-ffmpeg"
    # Configure logs are shipped with the binaries. Build tools do not need
    # the release/API credentials used by the downloader above.
    environment = {key: value for key, value in os.environ.items()
                   if not key.endswith(("_TOKEN", "_PASSWORD", "_KEY")) and not key.startswith("ACTIONS_")}
    if not windows:
        brew = command("brew", "--prefix")
        environment["CURB_FFMPEG_CFLAGS"] = f"-Os -I{brew}/include"
        environment["CURB_FFMPEG_LDFLAGS"] = f"-L{brew}/lib -Wl,-headerpad_max_install_names"
        environment["PKG_CONFIG_PATH"] = ":".join(command("brew", "--prefix", package) + "/lib/pkgconfig" for package in ("gnutls", "lame", "x264")) + ":" + environment.get("PKG_CONFIG_PATH", "")
        environment["MACOSX_DEPLOYMENT_TARGET"] = "15.0"
    subprocess.run([shell, "packaging/native/build-ffmpeg.sh", source.as_posix(), prefix.as_posix()], cwd=ROOT, env=environment, check=True)
    for name in ("ffmpeg", "ffprobe"):
        shutil.copy2(prefix / "bin" / (name + suffix), TOOLS / (name + suffix))
    for name, output in (("FFmpeg-LICENSE", "ffmpeg-LICENSE"), ("config.log", "ffmpeg-config.log")):
        shutil.copy2(prefix / name, notices / output)
    libraries = collect_windows(prefix, TOOLS) if windows else collect_macos(prefix, TOOLS)
    metadata["ffmpeg"] = {"version": version, "source": f"https://ffmpeg.org/releases/ffmpeg-{version}.tar.xz", "sha256": TOOL_VERSIONS["ffmpeg-source-sha256"], "configuration": "packaging/native/build-ffmpeg.sh", "libraries": libraries}
    for repo, tag, name, output in (
        ("yt-dlp/yt-dlp", TOOL_VERSIONS["yt-dlp"], "LICENSE", "yt-dlp-LICENSE"),
        ("yt-dlp/yt-dlp", TOOL_VERSIONS["yt-dlp"], "THIRD_PARTY_LICENSES.txt", "yt-dlp-THIRD_PARTY_LICENSES.txt"),
        ("quickjs-ng/quickjs", TOOL_VERSIONS["quickjs-ng"], "LICENSE", "quickjs-LICENSE"),
    ):
        (notices / output).write_bytes(fetch(f"https://raw.githubusercontent.com/{repo}/{tag}/{name}"))
    (notices / "SOURCES.txt").write_text(json.dumps(metadata, indent=2) + "\nNative dependency package versions/licenses are preserved beside this file.\n")
    (TOOLS / "manifest.json").write_text(json.dumps(metadata, indent=2) + "\n")
    for name, argument in (("yt-dlp", "--version"), ("qjs", "--version"), ("ffmpeg", "-version"), ("ffprobe", "-version")):
        binary = TOOLS / (name + suffix)
        binary.chmod(0o755)
        subprocess.run([str(binary), argument], cwd=ROOT, check=True)
    stamp.write_text(build_id)
