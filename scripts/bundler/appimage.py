"""Assemble and compress a native AppDir with the official type-2 runtime."""
import hashlib
import json
import os
import platform
import shutil
import subprocess

from .lean_ffmpeg import RUNTIME
from .network import asset, download, release
from .paths import CACHE, ROOT, VERSION


def packaging_tool(repo, tag, name):
    record = CACHE / f"{name}-{tag}.json"
    if record.exists():
        info = json.loads(record.read_text())
    else:
        info = asset(release(repo, tag), name)
        if not info.get("digest", "").startswith("sha256:"):
            raise RuntimeError(f"No published checksum for {name}")
        record.write_text(json.dumps(info))
    executable = download(info["browser_download_url"], name, info["digest"].removeprefix("sha256:"))
    executable.chmod(0o755)
    return executable, info


def build_appimage(portable):
    if not (RUNTIME / "gio").is_dir() or not (RUNTIME / "immodules.cache").is_file():
        raise SystemExit("AppImage runtime is missing or outdated; run make tools first.")
    arch = platform.machine()
    pins = json.loads((ROOT / "packaging/appimage.json").read_text())
    packager, packager_info = packaging_tool("AppImage/appimagetool", pins["appimagetool"], f"appimagetool-{arch}.AppImage")
    runtime, runtime_info = packaging_tool("AppImage/type2-runtime", pins["runtime"], f"runtime-{arch}")
    appdir = ROOT / "build" / "Curb.AppDir"
    if appdir.exists():
        shutil.rmtree(appdir)
    binary = appdir / "usr/bin"
    binary.mkdir(parents=True)
    shutil.copy2(portable / "curb", binary / "curb")
    shutil.copytree(portable / "tools", binary / "tools", symlinks=True)
    shutil.copytree(RUNTIME / "lib", appdir / "usr/lib", symlinks=True)
    shutil.copytree(RUNTIME / "fonts", appdir / "usr/share/fonts")
    shutil.copytree(RUNTIME / "gio", appdir / "usr/lib/gio/modules")
    shutil.copytree(RUNTIME / "schemas", appdir / "usr/share/glib-2.0/schemas")
    for name in ("immodules.cache", "pixbuf.cache"):
        shutil.copy2(RUNTIME / name, appdir / "usr/lib" / name)
    docs = appdir / "usr/share/doc/curb"
    docs.mkdir(parents=True)
    for name in ("LICENSE", "THIRD_PARTY_NOTICES.md", "README.md"):
        shutil.copy2(ROOT / name, docs / name)
    shutil.copytree(ROOT / "licenses", docs / "licenses")
    shutil.copytree(ROOT / "docs", docs / "docs")
    shutil.copytree(RUNTIME / "licenses", docs / "licenses/linux-runtime")
    shutil.copy2(RUNTIME / "manifest.json", docs / "linux-runtime.json")
    runtime_license = download(
        f"https://raw.githubusercontent.com/AppImage/type2-runtime/{pins['runtime']}/LICENSE",
        f"appimage-runtime-{pins['runtime']}-LICENSE",
    )
    shutil.copy2(runtime_license, docs / "appimage-runtime-LICENSE")
    metadata = {"appimagetool": packager_info, "runtime": runtime_info}
    (docs / "appimage-tools.json").write_text(json.dumps(metadata, indent=2) + "\n")
    desktop = (ROOT / "packaging/linux/Curb.desktop").read_text().replace("Exec=env GDK_BACKEND=wayland,x11 curb", "Exec=curb")
    (appdir / "dev.curb.app.desktop").write_text(desktop)
    shutil.copy2(ROOT / "resources/curb.svg", appdir / "curb.svg")
    (appdir / ".DirIcon").symlink_to("curb.svg")
    shutil.copy2(ROOT / "packaging/linux/AppRun", appdir / "AppRun")
    (appdir / "AppRun").chmod(0o755)
    fontconfig = appdir / "usr/etc/fonts"
    fontconfig.mkdir(parents=True)
    shutil.copy2(ROOT / "packaging/linux/fonts.conf", fontconfig / "fonts.conf")
    output = ROOT / "dist" / f"Curb-{VERSION}-{arch}.AppImage"
    output.unlink(missing_ok=True)
    env = dict(os.environ, ARCH=arch, APPIMAGELAUNCHER_DISABLE="1")
    subprocess.run([str(packager), "--appimage-extract-and-run", "--no-appstream",
                    "--runtime-file", str(runtime), "--comp", "zstd",
                    "--mksquashfs-opt", "-Xcompression-level", "--mksquashfs-opt", "19",
                    "--mksquashfs-opt", "-processors", "--mksquashfs-opt", "4",
                    str(appdir), str(output)], env=env, check=True)
    output.chmod(0o755)
    digest = hashlib.sha256(output.read_bytes()).hexdigest()
    output.with_suffix(".AppImage.sha256").write_text(f"{digest}  {output.name}\n")
    print(f"Ready: {output} ({output.stat().st_size / 1024**2:.1f} MiB)", flush=True)
    return output
