"""Native release builds and publishing for stable vMAJOR.MINOR.PATCH tags."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import secrets
import subprocess
import sys
import tempfile
import time

from update_manifest import build_manifest, version_tuple

ROOT = Path(__file__).resolve().parents[1]
REPOSITORY = "xptea/curb"
PLATFORMS = ("linux-amd64", "windows-amd64", "darwin-amd64", "darwin-arm64")


def tag_version(tag):
    if not tag.startswith("v"):
        raise ValueError("Use a stable tag such as v0.1.1")
    version_tuple(tag[1:])
    return tag[1:]


def asset_names(version):
    return [f"Curb-{version}-x86_64.AppImage", f"Curb-{version}-windows-amd64.zip",
            f"Curb-{version}-darwin-amd64.tar.gz", f"Curb-{version}-darwin-arm64.tar.gz"]


def run(*command, **kwargs):
    return subprocess.run(command, cwd=ROOT, check=True, **kwargs)


def github_release(tag):
    result = subprocess.run(["gh", "api", f"repos/{REPOSITORY}/releases/tags/{tag}"], text=True, capture_output=True)
    if result.returncode:
        if "404" in result.stderr:
            # The by-tag endpoint omits drafts. The publisher must find an
            # existing draft to upload assets, including one it just created.
            listing = subprocess.run(["gh", "api", "--paginate", "--slurp",
                                      f"repos/{REPOSITORY}/releases?per_page=100"], text=True, capture_output=True)
            if listing.returncode:
                raise RuntimeError(listing.stderr)
            return next((item for page in json.loads(listing.stdout) for item in page if item["tag_name"] == tag), None)
        raise RuntimeError(result.stderr)
    return json.loads(result.stdout)


def complete(tag):
    version = tag_version(tag)
    release = github_release(tag)
    if release is None or release.get("draft") or release.get("prerelease"):
        return False
    try:
        manifest = build_manifest(release)
    except ValueError:
        return False
    return set(PLATFORMS).issubset(manifest["platforms"])


def prepare(tag):
    version = tag_version(tag)
    config = json.loads((ROOT / "mygo.json").read_text())
    config["version"] = version
    (ROOT / "mygo.json").write_text(json.dumps(config, indent=2) + "\n")
    return version


def native_path(path, platform):
    value = str(path.resolve())
    # MSYS2's Python can return forward slashes even on Windows. Go's update
    # plan deliberately requires filepath.Clean's canonical Windows spelling.
    return value.replace("/", "\\") if platform.startswith("windows-") else value


def smoke_window(executable, target, version, platform):
    # A startup receipt is written by the real application's first drawn frame.
    with tempfile.TemporaryDirectory(prefix=".curb-update-", dir=target.parent) as temporary:
        work = Path(temporary)
        token = secrets.token_hex(24)
        plan = {"kind": "bundle" if platform.startswith("darwin-") else "directory", "target": native_path(target, platform),
                "executable": str(executable.relative_to(target)), "platform": platform, "work": native_path(work, platform),
                "staged": native_path(work / "unused", platform), "version": version, "token": token, "parent_pid": os.getpid()}
        path = work / "plan.json"
        path.write_text(json.dumps(plan))
        environment = dict(os.environ, CURB_UPDATE_PLAN=native_path(path, platform), MYGO_GPU="0")
        log = tempfile.TemporaryFile(mode="w+b")
        process = subprocess.Popen([str(executable.resolve())], cwd=target.parent, env=environment,
                                   stdin=subprocess.DEVNULL, stdout=log, stderr=subprocess.STDOUT)
        try:
            deadline = time.monotonic() + 45
            while time.monotonic() < deadline:
                receipt = work / "app-ready"
                if receipt.exists() and receipt.read_text() == token:
                    return
                if process.poll() is not None:
                    raise RuntimeError(f"Native app exited before drawing its first frame (exit {process.returncode})")
                time.sleep(0.1)
            raise RuntimeError("Native app did not acknowledge its first frame")
        finally:
            process.terminate()
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
            log.seek(0)
            diagnostics = log.read().decode("utf-8", errors="replace").strip()
            if diagnostics:
                print("Native startup log:\n" + diagnostics[-8000:], flush=True)
            log.close()


def build_native(platform):
    from bundler.native_tools import prepare_native
    prepare_native(platform)
    config = json.loads((ROOT / "mygo.json").read_text())
    config["resources"] = ["tools", "LICENSE", "THIRD_PARTY_NOTICES.md", "licenses", "docs"]
    if platform.startswith("darwin-"):
        config.setdefault("macos", {})["minimumSystemVersion"] = "15.0"
    (ROOT / "mygo.json").write_text(json.dumps(config, indent=2) + "\n")
    suffix = ".exe" if platform.startswith("windows-") else ""
    cli = ROOT / "build" / ("mygo-cli" + suffix)
    run("go", "build", "-buildvcs=false", "-o", str(cli), "github.com/egoist/mygo/cmd/mygo")
    goos, arch = platform.split("-", 1)
    run(str(cli), "build", "-platform", goos + "/" + arch, "-skip-dmg", "-skip-notarize")
    output = ROOT / "build/mygo" / platform
    target = output / "Curb.app" if goos == "darwin" else output
    executable = target / "Contents/MacOS/Curb" if goos == "darwin" else target / "Curb.exe"
    smoke_window(executable, target, config["version"], platform)
    tools = target / "Contents/Resources/tools" if goos == "darwin" else output / "tools"
    environment = dict(os.environ, CURB_INTEGRATION="1", CURB_INTEGRATION_TOOLS=str(tools.resolve()))
    run("go", "test", "-buildvcs=false", "-race", "-count=1", "./internal/download", env=environment)
    run(sys.executable, "scripts/package_desktop_update.py", "--platform", platform, "--source", str(target))
    name = next(name for name in asset_names(config["version"]) if platform in name)
    path = ROOT / "dist" / name
    digest = hashlib.sha256(path.read_bytes()).hexdigest()
    path.with_name(path.name + ".sha256").write_text(f"{digest}  {path.name}\n", newline="\n")


def publish(tag, assets):
    version = tag_version(tag)
    names = asset_names(version)
    release = github_release(tag)
    if assets is not None:
        files = [assets / name for name in names]
        if not all(path.is_file() for path in files):
            raise ValueError("Refusing a partial release: all four platform assets are required")
        if release is None:
            run("gh", "release", "create", tag, "--repo", REPOSITORY, "--verify-tag", "--draft", "--title", f"Curb {version}", "--generate-notes")
            release = github_release(tag)
            if release is None:
                raise RuntimeError("Created draft release is not yet visible; retry publication without rebuilding")
        existing = {asset["name"]: asset for asset in release["assets"]}
        for path in files:
            digest = "sha256:" + hashlib.sha256(path.read_bytes()).hexdigest()
            previous = existing.get(path.name)
            if previous and not release["draft"]:
                if previous.get("digest") != digest:
                    raise ValueError(f"Published asset {path.name} has different bytes; use a new version")
                continue
            command = ["gh", "release", "upload", tag, str(path), "--repo", REPOSITORY]
            if release["draft"]:
                command.append("--clobber")
            run(*command)
            checksum = path.with_name(path.name + ".sha256")
            if checksum.exists():
                run("gh", "release", "upload", tag, str(checksum), "--repo", REPOSITORY, "--clobber")
        previous = json.loads((ROOT / "version.json").read_text())
        latest = version_tuple(version) >= version_tuple(previous["version"])
        run("gh", "release", "edit", tag, "--repo", REPOSITORY, "--draft=false", "--prerelease=false", "--latest=" + str(latest).lower())
    for attempt in range(6):
        release = github_release(tag)
        try:
            manifest = build_manifest(release, json.loads((ROOT / "version.json").read_text()))
            if set(PLATFORMS).issubset(build_manifest(release)["platforms"]):
                break
        except ValueError:
            if attempt == 5:
                raise
        time.sleep(3)
    else:
        raise ValueError("Published release is missing platform assets")
    (ROOT / "version.json").write_text(json.dumps(manifest, indent=2) + "\n")
    config = json.loads((ROOT / "mygo.json").read_text())
    if version_tuple(version) > version_tuple(config["version"]):
        config["version"] = version
        (ROOT / "mygo.json").write_text(json.dumps(config, indent=2) + "\n")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    for name in ("prepare", "complete", "publish"):
        child = sub.add_parser(name)
        child.add_argument("--tag", required=True)
        if name == "publish":
            child.add_argument("--assets", type=Path)
    child = sub.add_parser("build-native")
    child.add_argument("--platform", choices=PLATFORMS[1:], required=True)
    child = sub.add_parser("prepare-tools")
    child.add_argument("--platform", choices=PLATFORMS[1:], required=True)
    args = parser.parse_args()
    if args.command == "prepare":
        print(prepare(args.tag))
    elif args.command == "complete":
        result = str(complete(args.tag)).lower()
        print(result)
        if os.environ.get("GITHUB_OUTPUT"):
            with open(os.environ["GITHUB_OUTPUT"], "a") as file:
                file.write(f"complete={result}\n")
    elif args.command == "build-native":
        build_native(args.platform)
    elif args.command == "prepare-tools":
        from bundler.native_tools import prepare_native
        prepare_native(args.platform)
    else:
        publish(args.tag, args.assets)


if __name__ == "__main__":
    main()
