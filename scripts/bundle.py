#!/usr/bin/env python3
"""Prepare download tools and build a portable Linux Curb distribution."""
import argparse
import platform

from bundler.build import bundle
from bundler.paths import CACHE, TOOLS
from bundler.tools import prepare_tools


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--skip-download", action="store_true", help="Reuse already prepared tools")
    parser.add_argument("--prepare-only", action="store_true", help="Prepare tools without building a distribution")
    parser.add_argument("--format", choices=("appimage", "tar"), default="appimage", help="Distribution format (default: AppImage)")
    args = parser.parse_args()
    if platform.system() != "Linux" or platform.machine() not in ("x86_64", "aarch64"):
        parser.error("This bundle script targets Linux x86_64 and aarch64")
    # Restored toolchains do not include transient AppImage packager downloads.
    CACHE.mkdir(parents=True, exist_ok=True)
    if not args.skip_download:
        prepare_tools()
    required = ("yt-dlp", "ffmpeg", "ffprobe", "qjs", "manifest.json")
    if any(not (TOOLS / name).is_file() for name in required):
        parser.error("Bundled tools are missing; run without --skip-download")
    if not args.prepare_only:
        bundle(args.format)


if __name__ == "__main__":
    main()
