"""Collect native FFmpeg dependencies without relying on the target's Homebrew/MSYS2."""
from pathlib import Path
import os
import re
import shutil
import subprocess


def command(*arguments):
    return subprocess.check_output(arguments, text=True).strip()


def windows_imports(output):
    return re.findall(r"DLL Name:\s*(\S+)", output)


def macos_imports(output):
    return [line.strip().split(" (", 1)[0] for line in output.splitlines()[1:] if " (" in line]


def collect_windows(prefix, tools):
    mingw = Path(command("cygpath", "-w", os.environ["MINGW_PREFIX"]))
    system = Path(os.environ["SystemRoot"]) / "System32"
    queue = sorted(tools.glob("*.exe"))
    copied = set()
    while queue:
        binary = queue.pop()
        for name in windows_imports(command("objdump", "-p", str(binary))):
            key = name.lower()
            if key in copied or key.startswith(("api-ms-", "ext-ms-")):
                continue
            source = next((root / name for root in (prefix / "bin", mingw / "bin") if (root / name).is_file()), None)
            if source is None:
                if (system / name).is_file():
                    continue
                raise RuntimeError(f"Missing native DLL: {name}")
            target = tools / name
            shutil.copy2(source, target)
            copied.add(key)
            queue.append(target)
    licenses = mingw / "share/licenses"
    if licenses.exists():
        shutil.copytree(licenses, tools / "licenses/msys2", dirs_exist_ok=True)
    (tools / "licenses/msys2-packages.txt").write_text(command("pacman", "-Q") + "\n")
    return sorted(copied)


def collect_macos(prefix, tools):
    brew = Path(command("brew", "--prefix"))
    libdir = tools / "lib"
    libdir.mkdir()
    queue = [(tools / "ffmpeg", prefix / "bin/ffmpeg"), (tools / "ffprobe", prefix / "bin/ffprobe"),
             (tools / "qjs", tools / "qjs"), (tools / "yt-dlp", tools / "yt-dlp")]
    copied = {}
    packages = set()
    while queue:
        binary, origin = queue.pop()
        for dependency in macos_imports(command("otool", "-L", str(binary))):
            if dependency.startswith(("/usr/lib/", "/System/Library/")):
                continue
            if binary.name == "yt-dlp":
                raise RuntimeError("Refusing to rewrite the PyInstaller launcher; it has an unexpected external dependency")
            if dependency.startswith("@loader_path/"):
                source = origin.parent / dependency.removeprefix("@loader_path/")
            elif dependency.startswith("@rpath/"):
                name = dependency.removeprefix("@rpath/")
                roots = [origin.parent, prefix / "lib", brew / "lib", *sorted((brew / "opt").glob("*/lib"))]
                source = next((root / name for root in roots if (root / name).is_file()), None)
            else:
                source = Path(dependency)
            if source is None or not source.is_file():
                raise RuntimeError(f"Missing native dylib: {dependency}")
            source = source.resolve()
            name = source.name
            if name in copied and copied[name] != source:
                raise RuntimeError(f"Conflicting native libraries: {name}")
            if name not in copied:
                copied[name] = source
                target = libdir / name
                shutil.copy2(source, target)
                target.chmod(0o755)
                queue.append((target, source))
                if "/Cellar/" in str(source):
                    parts = str(source).split("/Cellar/", 1)[1].split("/")
                    package = Path(str(source).split("/Cellar/", 1)[0]) / "Cellar" / parts[0] / parts[1]
                    packages.add(package)
            relative = "@loader_path/" + ("lib/" if binary.parent == tools else "") + name
            subprocess.run(["install_name_tool", "-change", dependency, relative, str(binary)], check=True)
        if binary.parent == libdir:
            subprocess.run(["install_name_tool", "-id", "@rpath/" + binary.name, str(binary)], check=True)
        # PyInstaller carries its Python archive after the native launcher.
        if binary.name != "yt-dlp":
            subprocess.run(["strip", "-x", str(binary)], check=True)
            # Relocation/stripping invalidates old signatures. MyGo preserves
            # pre-signed resource binaries, so repair these before bundling.
            subprocess.run(["codesign", "--force", "--sign", "-", str(binary)], check=True)
    for package in packages:
        destination = tools / "licenses/homebrew" / package.parent.name
        destination.mkdir(parents=True, exist_ok=True)
        for file in package.rglob("*"):
            if file.is_file() and (file.name.startswith(("LICENSE", "COPYING", "AUTHORS", "NOTICE")) or file.name == "INSTALL_RECEIPT.json"):
                target = destination / file.relative_to(package)
                target.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(file, target)
    (tools / "licenses/homebrew-packages.json").write_text(command("brew", "info", "--json=v2", "--installed") + "\n")
    return sorted(copied)
