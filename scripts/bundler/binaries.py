"""Remove development symbols from bundled Linux shared libraries."""
import shutil
import subprocess


def strip_shared_libraries(root):
    tool = shutil.which("strip")
    if tool is None:
        raise RuntimeError("Install binutils (strip) to prepare the compact Linux tools")
    count, before, after = 0, 0, 0
    for path in sorted(root.rglob("*")):
        if path.is_symlink() or not path.is_file():
            continue
        if not (path.name.endswith(".so") or ".so." in path.name):
            continue
        with path.open("rb") as source:
            if source.read(4) != b"\x7fELF":
                continue
        size = path.stat().st_size
        subprocess.run([tool, "--strip-unneeded", str(path)], check=True)
        before += size
        after += path.stat().st_size
        count += 1
    # Never strip the PyInstaller launcher: its appended Python archive must
    # remain intact. Exported symbols needed by shared libraries are retained.
    return {"method": "strip --strip-unneeded (ELF shared libraries only)",
            "libraries": count, "before_bytes": before, "after_bytes": after}
