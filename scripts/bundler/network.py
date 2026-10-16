"""Download and verify published release assets."""
import hashlib
import http.client
import json
import shutil
import time
import urllib.error
import urllib.request
import os
from urllib.parse import urlparse

from .paths import CACHE, HEADERS


def retry_download(operation, label):
    for attempt in range(4):
        try:
            return operation()
        except (urllib.error.URLError, TimeoutError, ConnectionError, http.client.IncompleteRead) as error:
            if isinstance(error, urllib.error.HTTPError) and error.code not in (408, 429, 500, 502, 503, 504):
                raise
            if attempt == 3:
                raise
            print(f"Network interruption fetching {label}; retrying ({attempt + 1}/3)…", flush=True)
            time.sleep(2 ** attempt)


def fetch(url):
    headers = dict(HEADERS)
    if urlparse(url).hostname == "api.github.com" and os.environ.get("GH_TOKEN"):
        headers["Authorization"] = "Bearer " + os.environ["GH_TOKEN"]
    def request():
        with urllib.request.urlopen(urllib.request.Request(url, headers=headers), timeout=60) as response:
            return response.read()
    return retry_download(request, urlparse(url).hostname)


def release(repo, tag="latest"):
    suffix = "latest" if tag == "latest" else "tags/" + tag
    return json.loads(fetch(f"https://api.github.com/repos/{repo}/releases/{suffix}"))


def download(url, name, expected=None):
    destination = CACHE / name
    if not destination.exists():
        print(f"Downloading {name}…", flush=True)
        temporary = destination.with_suffix(destination.suffix + ".partial")
        def transfer():
            try:
                request = urllib.request.Request(url, headers={"User-Agent": HEADERS["User-Agent"]})
                with urllib.request.urlopen(request, timeout=120) as response:
                    with temporary.open("wb") as output:
                        shutil.copyfileobj(response, output)
                if expected and hashlib.sha256(temporary.read_bytes()).hexdigest() != expected:
                    raise RuntimeError(f"Checksum mismatch for {name}; removed invalid download")
                temporary.replace(destination)
            finally:
                temporary.unlink(missing_ok=True)
        retry_download(transfer, name)
    if expected:
        with destination.open("rb") as data:
            actual = hashlib.file_digest(data, "sha256").hexdigest()
        if actual != expected:
            destination.unlink()
            raise RuntimeError(f"Checksum mismatch for {name}; removed invalid download")
    return destination


def asset(release_info, name):
    return next(a for a in release_info["assets"] if a["name"] == name)


def asset_checksum(release_info, name):
    digest = asset(release_info, name).get("digest", "")
    if digest.startswith("sha256:"):
        value = digest.removeprefix("sha256:")
        if len(value) == 64 and all(character in "0123456789abcdef" for character in value):
            return value
        raise RuntimeError(f"Invalid published checksum for {name}")
    return checksum(fetch(asset(release_info, "SHA2-256SUMS")["browser_download_url"]).decode(), name)


def checksum(text, filename):
    for line in text.splitlines():
        parts = line.split()
        if len(parts) == 2 and parts[1].lstrip("*").removeprefix("./") == filename:
            return parts[0]
    raise RuntimeError(f"No published checksum for {filename}")
