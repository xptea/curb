"""Build the small shared FFmpeg configuration used by Curb's AppImage."""
import hashlib
import os
import platform
import shutil
import subprocess
import tarfile

from .network import download
from .paths import ROOT, TOOL_VERSIONS

IMAGE = "curb-appimage-runtime:22.04"
RUNTIME = ROOT / "build" / "linux-runtime"


def docker_run(*command):
    subprocess.run(["docker", "run", "--rm", "--user", f"{os.getuid()}:{os.getgid()}",
                    "-v", f"{ROOT}:/work", "-w", "/work", IMAGE, *command], check=True)


def prepare():
    version = TOOL_VERSIONS["ffmpeg"]
    script = ROOT / "packaging" / "linux" / "build-ffmpeg.sh"
    dockerfile = ROOT / "packaging/linux/runtime.Dockerfile"
    collector = ROOT / "scripts/bundler/linux_runtime.py"
    digest = hashlib.sha256(script.read_bytes() + dockerfile.read_bytes() + collector.read_bytes()
                            + version.encode() + TOOL_VERSIONS["ffmpeg-source-sha256"].encode()
                            + platform.machine().encode()).hexdigest()
    output = ROOT / "build" / "lean-ffmpeg"
    stamp = output / "curb-build-id"
    complete = all(path.exists() for path in (output / "bin/ffmpeg", output / "bin/ffprobe", output / "lib",
                                              output / "FFmpeg-LICENSE", output / "config.log",
                                              RUNTIME / "manifest.json", RUNTIME / "immodules.cache", RUNTIME / "gio"))
    if complete and stamp.exists() and stamp.read_text() == digest:
        return output
    subprocess.run(["docker", "build", "-f", str(ROOT / "packaging/linux/runtime.Dockerfile"),
                    "-t", IMAGE, str(ROOT)], check=True)
    archive = download(f"https://ffmpeg.org/releases/ffmpeg-{version}.tar.xz",
                       f"ffmpeg-{version}.tar.xz", TOOL_VERSIONS["ffmpeg-source-sha256"])
    source = ROOT / "build" / f"ffmpeg-{version}"
    if not source.exists():
        with tarfile.open(archive) as data:
            data.extractall(ROOT / "build", filter="data")
    if output.exists():
        shutil.rmtree(output)
    docker_run("sh", "/work/packaging/linux/build-ffmpeg.sh",
               f"/work/build/ffmpeg-{version}", "/work/build/lean-ffmpeg")
    docker_run("python3", "scripts/bundler/linux_runtime.py", "/work/build/linux-runtime")
    stamp.write_text(digest)
    return output
