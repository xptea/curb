"""Package an existing MyGo macOS bundle or Windows portable app for updates.

Run on the target OS after preparing its own native download tools. This
script does not compile, sign, publish, or modify the supplied application.
"""
import argparse
import json
from pathlib import Path
import plistlib
import shutil
import tarfile
import tempfile
import zipfile

from update_manifest import version_tuple


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--platform", choices=["windows-amd64", "windows-arm64", "darwin-amd64", "darwin-arm64", "darwin-universal"], required=True)
    parser.add_argument("--source", type=Path, required=True, help="MyGo output folder on Windows; Curb.app on macOS")
    parser.add_argument("--out", type=Path, default=Path("dist"))
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    version = json.loads((root / "mygo.json").read_text())["version"]
    version_tuple(version)
    args.out.mkdir(parents=True, exist_ok=True)
    if args.platform.startswith("darwin-"):
        with (args.source / "Contents/Info.plist").open("rb") as file:
            plist = plistlib.load(file)
        if plist.get("CFBundleIdentifier") != "dev.curb.app" or plist.get("CFBundleShortVersionString") != version:
            raise ValueError("Bundle identity/version differs from mygo.json")
        output = args.out / f"Curb-{version}-{args.platform}.tar.gz"
        with tarfile.open(output, "w:gz", dereference=False) as archive:
            archive.add(args.source, arcname="Curb.app")
    else:
        if not (args.source / "Curb.exe").is_file():
            raise ValueError("Expected the MyGo output folder containing Curb.exe")
        output = args.out / f"Curb-{version}-{args.platform}.zip"
        with tempfile.TemporaryDirectory() as temporary:
            app = Path(temporary) / "Curb"
            shutil.copytree(args.source, app, ignore=shutil.ignore_patterns("Curb Setup *.exe"))
            files = sorted(path.relative_to(app).as_posix() for path in app.rglob("*"))
            if ".curb-install.json" not in files:
                files.append(".curb-install.json")
            marker = {"product": "dev.curb.app", "version": version, "platform": args.platform, "files": sorted(files)}
            (app / ".curb-install.json").write_text(json.dumps(marker) + "\n")
            with zipfile.ZipFile(output, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
                for path in sorted(app.rglob("*")):
                    if path.is_file():
                        archive.write(path, path.relative_to(app.parent).as_posix())
    print(output)


if __name__ == "__main__":
    main()
