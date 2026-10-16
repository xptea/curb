"""Download and verify published release assets."""
import hashlib
import json
import shutil
import urllib.request
import os
from urllib.parse import urlparse

from .paths import CACHE, HEADERS


def fetch(url):
    headers = dict(HEADERS)
    if urlparse(url).hostname == "api.github.com" and os.environ.get("GH_TOKEN"):
        headers["Authorization"] = "Bearer " + os.environ["GH_TOKEN"]
    with urllib.request.urlopen(urllib.request.Request(url, headers=headers), timeout=60) as response:
        return response.read()


def release(repo, tag="latest"):
    suffix = "latest" if tag == "latest" else "tags/" + tag
    return json.loads(fetch(f"https://api.github.com/repos/{repo}/releases/{suffix}"))


def download(url, name, expected=None):
    destination = CACHE / name
    if not destination.exists():
        print(f"Downloading {name}…", flush=True)
        temporary = destination.with_suffix(destination.suffix + ".partial")
        request = urllib.request.Request(url, headers={"User-Agent": HEADERS["User-Agent"]})
        with urllib.request.urlopen(request, timeout=120) as response:
            with temporary.open("wb") as output:
                shutil.copyfileobj(response, output)
        temporary.replace(destination)
    if expected:
        with destination.open("rb") as data:
            actual = hashlib.file_digest(data, "sha256").hexdigest()
        if actual != expected:
            destination.unlink()
            raise RuntimeError(f"Checksum mismatch for {name}; removed invalid download")
    return destination


def asset(release_info, name):
    return next(a for a in release_info["assets"] if a["name"] == name)


def checksum(text, filename):
    for line in text.splitlines():
        parts = line.split()
        if len(parts) == 2 and parts[1].lstrip("*").removeprefix("./") == filename:
            return parts[0]
    raise RuntimeError(f"No published checksum for {filename}")
