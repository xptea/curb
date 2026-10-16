"""Create the GitHub Raw update feed from an already-published release.

Uses GitHub's asset sizes and SHA-256 digests. It never advertises a draft,
prerelease, missing platform asset, or an older version than the existing feed.
"""
import argparse
import json
from pathlib import Path
import re

REPOSITORY = "xptea/curb"


def version_tuple(version):
    if not re.fullmatch(r"(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)", version):
        raise ValueError("Use a stable version such as 0.1.1")
    return tuple(map(int, version.split(".")))


def build_manifest(release, previous=None):
    if release.get("draft") or release.get("prerelease") or not release.get("published_at"):
        raise ValueError("Only a published stable release can become an update")
    tag = release["tag_name"]
    if not tag.startswith("v"):
        raise ValueError("Release tags must start with v, such as v0.1.1")
    version = tag[1:]
    current = version_tuple(version)
    if previous and version_tuple(previous["version"]) > current:
        return previous
    platforms = {}
    names = {
        f"Curb-{version}-x86_64.AppImage": ("linux-amd64", "appimage"),
        f"Curb-{version}-aarch64.AppImage": ("linux-arm64", "appimage"),
    }
    for platform in ("windows-amd64", "windows-arm64", "darwin-amd64", "darwin-arm64", "darwin-universal"):
        extension = "zip" if platform.startswith("windows-") else "tar.gz"
        names[f"Curb-{version}-{platform}.{extension}"] = (platform, extension)
    for asset in release.get("assets", []):
        name = asset["name"]
        if name not in names:
            continue
        platform, format_ = names[name]
        digest = asset.get("digest", "")
        if not re.fullmatch(r"sha256:[0-9a-fA-F]{64}", digest):
            raise ValueError(f"Missing published SHA-256 digest for {name}")
        expected_url = f"https://github.com/{REPOSITORY}/releases/download/{tag}/{name}"
        if asset.get("browser_download_url") != expected_url or asset.get("state") != "uploaded":
            raise ValueError(f"Invalid release asset URL or state: {name}")
        size = asset.get("size", 0)
        if not isinstance(size, int) or not 0 < size <= 1 << 30:
            raise ValueError(f"Invalid release asset size: {name}")
        platforms[platform] = {"name": name, "sha256": digest[7:].lower(), "size": size, "format": format_}
    if not platforms:
        raise ValueError("The release contains no supported desktop update assets")
    return {"schema": 1, "version": version, "platforms": platforms}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--release-json", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=Path("version.json"))
    args = parser.parse_args()
    previous = json.loads(args.output.read_text()) if args.output.exists() else None
    manifest = build_manifest(json.loads(args.release_json.read_text()), previous)
    args.output.write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Update feed: {manifest['version']} ({', '.join(manifest['platforms'])})")


if __name__ == "__main__":
    main()
