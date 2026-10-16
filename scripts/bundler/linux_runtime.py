"""Collect the GTK runtime from an older supported Linux build baseline."""
import argparse
import json
from pathlib import Path
import re
import shutil
import subprocess


SYSTEM_LIBRARIES = {"libc.so.6", "libm.so.6", "libdl.so.2", "librt.so.1",
                    "libpthread.so.0", "libresolv.so.2", "libgcc_s.so.1", "libstdc++.so.6",
                    "ld-linux-x86-64.so.2", "ld-linux-aarch64.so.1"}


def collect_tool_libraries(tools, runtime):
    """Make the FFmpeg tools runnable without libraries installed on the host."""
    libraries = tools / "lib"
    queue = [tools / "bin/ffmpeg", tools / "bin/ffprobe"]
    visited = set()
    copied = []
    while queue:
        binary = queue.pop()
        output = subprocess.check_output(["objdump", "-p", str(binary)], text=True)
        for name in re.findall(r"^\s*NEEDED\s+(\S+)", output, re.MULTILINE):
            if name in visited or name in SYSTEM_LIBRARIES:
                continue
            visited.add(name)
            target = libraries / name
            if not target.is_file():
                source = runtime / "lib" / name
                if not source.is_file():
                    raise RuntimeError(f"Missing bundled FFmpeg library: {name}")
                shutil.copy2(source, target)
                copied.append(name)
            queue.append(target)
    return sorted(copied)


def collect(destination):
    destination = Path(destination)
    if destination.exists():
        shutil.rmtree(destination)
    libraries = destination / "lib"
    notices = destination / "licenses"
    libraries.mkdir(parents=True)
    notices.mkdir()
    packages = set()
    roots = [next(Path("/usr/lib").glob(f"*-linux-gnu/{name}"))
             for name in ("libgtk-3.so.0", "libepoxy.so.0", "libmp3lame.so.0", "libgnutls.so.30")]
    module_sets = {
        "gio": list(Path("/usr/lib").glob("*-linux-gnu/gio/modules/*.so")),
        "immodules": list(Path("/usr/lib").glob("*-linux-gnu/gtk-3.0/*/immodules/*.so")),
        "pixbuf": list(Path("/usr/lib").glob("*-linux-gnu/gdk-pixbuf-2.0/*/loaders/*.so")),
    }
    for modules in module_sets.values():
        roots.extend(modules)
    files = {str(root) for root in roots}
    for root in roots:
        output = subprocess.check_output(["ldd", str(root)], text=True)
        if "not found" in output:
            raise RuntimeError(output)
        files.update(re.findall(r"=> (/\S+)", output))
    # Use the host's libc and GPU driver stack, never a newer build-host copy.
    for filename in sorted(files):
        path = Path(filename)
        if path.name in SYSTEM_LIBRARIES:
            continue
        shutil.copy2(path, libraries / path.name)
        # Resolve /lib -> /usr/lib differences in Debian package ownership.
        owner = subprocess.run(["dpkg-query", "-S", filename], text=True, capture_output=True)
        if owner.returncode:
            owner = subprocess.run(["dpkg-query", "-S", str(path.resolve())], text=True, capture_output=True)
        if owner.returncode:
            raise RuntimeError(f"Cannot identify runtime library owner: {path}")
        packages.update(line.rsplit(": ", 1)[0] for line in owner.stdout.splitlines())
    # Module paths in these caches must resolve inside the mounted AppImage.
    # dlopen finds the flattened module names through AppRun's LD_LIBRARY_PATH.
    for kind, pattern in (("immodules", "*-linux-gnu/gtk-3.0/*/immodules.cache"),
                          ("pixbuf", "*-linux-gnu/gdk-pixbuf-2.0/*/loaders.cache")):
        cache = next(Path("/usr/lib").glob(pattern))
        content = re.sub(r'"/[^"\n]+/([^/"\n]+\.so)"', r'"\1"', cache.read_text())
        (destination / f"{kind}.cache").write_text(content)
    gio = destination / "gio"
    gio.mkdir()
    for module in module_sets["gio"]:
        shutil.copy2(module, gio / module.name)
    cache = next(Path("/usr/lib").glob("*-linux-gnu/gio/modules/giomodule.cache"))
    shutil.copy2(cache, gio)
    shutil.copytree("/usr/share/glib-2.0/schemas", destination / "schemas")
    packages.add("libgtk-3-common")
    # MyGo can use this font even on desktops with no font package installed.
    fonts = destination / "fonts"
    fonts.mkdir()
    shutil.copy2("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", fonts)
    packages.add("fonts-dejavu-core")
    manifest = []
    for package in sorted(packages):
        record = subprocess.check_output(["dpkg-query", "-W", "-f=${binary:Package}\t${Version}\t${source:Package}\t${source:Version}", package], text=True).split("\t")
        manifest.append(dict(zip(("package", "version", "source", "source_version"), record)))
        copyright_file = Path("/usr/share/doc") / package.split(":")[0] / "copyright"
        if copyright_file.exists():
            shutil.copy2(copyright_file, notices / f"{package.replace(':', '_')}.copyright")
    shutil.copytree("/usr/share/common-licenses", notices / "common-licenses")
    (destination / "manifest.json").write_text(json.dumps({"baseline": "Ubuntu 22.04", "packages": manifest}, indent=2) + "\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("destination")
    collect(parser.parse_args().destination)
