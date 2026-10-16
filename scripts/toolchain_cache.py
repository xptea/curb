"""Reuse prepared native tools across release tags through GitHub artifacts."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import posixpath
import shutil
import subprocess
import tarfile
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]
REPOSITORY = "xptea/curb"
ARCHIVE = ROOT / "build/toolchain-cache/toolchain.tar.gz"


def cache_key(platform):
    files = [ROOT / "packaging/tools.json", Path(__file__)]
    for directory in ("packaging/linux", "packaging/native", "scripts/bundler"):
        files.extend(path for path in (ROOT / directory).glob("*") if path.suffix in (".py", ".sh", ".Dockerfile"))
    digest = hashlib.sha256(platform.encode())
    for path in sorted(files):
        digest.update(path.relative_to(ROOT).as_posix().encode())
        digest.update(path.read_bytes())
    return f"curb-toolchain-v1-{platform}-{digest.hexdigest()}"


def api(path):
    result = subprocess.check_output(["gh", "api", f"repos/{REPOSITORY}/{path}"], text=True)
    return json.loads(result)


def trusted_run(run):
    return (run.get("event") in ("push", "release", "workflow_dispatch")
            and run.get("head_repository", {}).get("full_name") == REPOSITORY
            and run.get("path", "").split("@", 1)[0] == ".github/workflows/release.yml")


def extract(archive):
    with tarfile.open(archive) as data:
        for member in data.getmembers():
            path = Path(member.name)
            def allowed(candidate):
                return (not candidate.is_absolute() and ".." not in candidate.parts
                        and (candidate.parts[:1] == ("tools",) or candidate.parts[:2] == ("build", "linux-runtime")))
            if not allowed(path):
                raise ValueError("Unexpected toolchain cache path")
            if member.issym() or member.islnk():
                link = member.linkname if member.islnk() else posixpath.join(member.name.rsplit("/", 1)[0], member.linkname)
                if not allowed(Path(posixpath.normpath(link))):
                    raise ValueError("Unexpected toolchain cache link")
        data.extractall(ROOT, filter="data")


def restore(platform):
    key = cache_key(platform)
    hit = False
    # Tag-scoped Actions caches cannot be shared with the next release tag.
    # Repository artifacts can, and are saved before the app build starts.
    artifacts = api(f"actions/artifacts?name={key}&per_page=20")["artifacts"]
    for artifact in sorted(artifacts, key=lambda item: item["id"], reverse=True):
        if artifact["expired"] or artifact["name"] != key:
            continue
        run = api(f"actions/runs/{artifact['workflow_run']['id']}")
        if not trusted_run(run):
            continue
        with tempfile.TemporaryDirectory(prefix="curb-tools-") as temporary:
            download = Path(temporary) / "artifact.zip"
            with download.open("wb") as output:
                subprocess.run(["gh", "api", "--allow-escape-sequences",
                                f"repos/{REPOSITORY}/actions/artifacts/{artifact['id']}/zip"], stdout=output, check=True)
            expected = artifact.get("digest", "")
            if not expected.startswith("sha256:"):
                raise ValueError("Toolchain artifact has no published checksum")
            with download.open("rb") as source:
                if "sha256:" + hashlib.file_digest(source, "sha256").hexdigest() != expected:
                    raise ValueError("Toolchain artifact checksum mismatch")
            with zipfile.ZipFile(download) as files:
                with files.open("toolchain.tar.gz") as source, (Path(temporary) / "tools.tar.gz").open("wb") as target:
                    shutil.copyfileobj(source, target)
            extract(Path(temporary) / "tools.tar.gz")
        hit = True
        print(f"Reused {platform} tools from release run {run['id']}; skipping FFmpeg compilation.", flush=True)
        break
    if not hit:
        print(f"No matching {platform} toolchain artifact; building it once for reuse.", flush=True)
    if os.environ.get("GITHUB_OUTPUT"):
        with open(os.environ["GITHUB_OUTPUT"], "a") as output:
            output.write(f"key={key}\nhit={str(hit).lower()}\n")


def pack(platform):
    ARCHIVE.parent.mkdir(parents=True, exist_ok=True)
    with tarfile.open(ARCHIVE, "w:gz", compresslevel=3) as archive:
        archive.add(ROOT / "tools", arcname="tools")
        if platform.startswith("linux-"):
            archive.add(ROOT / "build/linux-runtime", arcname="build/linux-runtime")
    print(f"Prepared reusable toolchain: {ARCHIVE}", flush=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("restore", "pack"))
    parser.add_argument("--platform", required=True)
    arguments = parser.parse_args()
    (restore if arguments.command == "restore" else pack)(arguments.platform)
