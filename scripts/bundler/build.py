"""Compile Curb and assemble the portable Linux distribution."""
import platform
import shutil
import subprocess
import tarfile

from .paths import ROOT, TOOLS


def bundle(format="appimage"):
    out = ROOT / "dist" / "Curb"
    out.mkdir(parents=True, exist_ok=True)
    subprocess.run(
        ["go", "build", "-buildvcs=false", "-trimpath", "-ldflags=-s -w",
         "-o", str(out / "curb"), "."],
        cwd=ROOT,
        check=True,
    )
    # Recreate tool resources so retired runtimes cannot survive a rebuild.
    if (out / "tools").exists():
        shutil.rmtree(out / "tools")
    shutil.copytree(TOOLS, out / "tools", symlinks=True)
    shutil.copy2(ROOT / "resources" / "curb.svg", out / "curb.svg")
    for name in ("Curb.desktop", "install.sh"):
        shutil.copy2(ROOT / "packaging" / "linux" / name, out / name)
    for name in ("README.md", "CONTRIBUTING.md", "LICENSE", "THIRD_PARTY_NOTICES.md"):
        shutil.copy2(ROOT / name, out / name)
    for name in ("docs", "licenses"):
        # Dependency licenses can be read-only copies from Go's module cache.
        # Recreate generated resources so repeated bundles also replace them.
        if (out / name).exists():
            shutil.rmtree(out / name)
        shutil.copytree(ROOT / name, out / name)
    (out / "install.sh").chmod(0o755)
    if format == "appimage":
        from .appimage import build_appimage
        return build_appimage(out)
    arch = "x86_64" if platform.machine() == "x86_64" else "aarch64"
    archive = ROOT / "dist" / f"Curb-linux-{arch}.tar.gz"
    print("Packaging portable app…", flush=True)
    with tarfile.open(archive, "w:gz", compresslevel=3) as output:
        output.add(out, arcname="Curb")
    print(f"Ready: {archive}", flush=True)
