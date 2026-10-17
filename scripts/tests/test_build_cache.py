"""Build-cache regression checks; no Docker, compiler, or network required."""
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import bundle
from bundler import lean_ffmpeg


class BuildCacheTests(unittest.TestCase):
    def test_restored_toolchain_can_package_without_an_existing_download_directory(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            tools = root / "tools"
            tools.mkdir()
            for name in ("yt-dlp", "ffmpeg", "ffprobe", "qjs", "manifest.json"):
                (tools / name).write_text("restored tool")
            cache = root / "build/downloads"

            def package(format):
                self.assertTrue(cache.is_dir())
                self.assertEqual(format, "appimage")

            with patch.object(bundle, "CACHE", cache), patch.object(bundle, "TOOLS", tools), \
                    patch.object(bundle.platform, "system", return_value="Linux"), \
                    patch.object(bundle.platform, "machine", return_value="x86_64"), \
                    patch("sys.argv", ["bundle.py", "--skip-download"]), \
                    patch.object(bundle, "prepare_tools", side_effect=AssertionError("unnecessary tool rebuild")), \
                    patch.object(bundle, "bundle", side_effect=package) as build:
                bundle.main()
                build.assert_called_once_with("appimage")

    def test_cache_refreshes_when_inputs_change_or_outputs_are_missing(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            runtime = root / "build/linux-runtime"
            output = root / "build/lean-ffmpeg"
            collector = root / "scripts/bundler/linux_runtime.py"
            for path in (root / "packaging/linux/build-ffmpeg.sh", root / "packaging/linux/runtime.Dockerfile", collector):
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("initial build input")
            (root / "build/ffmpeg-test").mkdir(parents=True)
            versions = {"ffmpeg": "test", "ffmpeg-source-sha256": "original-digest"}
            builds = []

            def build(*command):
                builds.append(command[0])
                if command[0] == "sh":
                    for path in (output / "bin/ffmpeg", output / "bin/ffprobe", output / "FFmpeg-LICENSE", output / "config.log"):
                        path.parent.mkdir(parents=True, exist_ok=True)
                        path.write_text("built output")
                    (output / "lib").mkdir()
                else:
                    runtime.mkdir(parents=True, exist_ok=True)
                    (runtime / "manifest.json").write_text("{}")
                    (runtime / "immodules.cache").write_text("cache")
                    (runtime / "gio").mkdir(exist_ok=True)

            with patch.object(lean_ffmpeg, "ROOT", root), patch.object(lean_ffmpeg, "RUNTIME", runtime), \
                    patch.object(lean_ffmpeg, "TOOL_VERSIONS", versions), \
                    patch.object(lean_ffmpeg, "docker_run", side_effect=build), \
                    patch.object(lean_ffmpeg.subprocess, "run"), \
                    patch.object(lean_ffmpeg, "download", return_value=root / "archive.tar.xz"):
                self.assertEqual(lean_ffmpeg.prepare(), output)
                self.assertEqual(builds, ["sh", "python3"])
                lean_ffmpeg.prepare()
                self.assertEqual(builds, ["sh", "python3"])

                collector.write_text("updated runtime collector")
                lean_ffmpeg.prepare()
                self.assertEqual(builds.count("sh"), 2)

                versions["ffmpeg-source-sha256"] = "new-source-digest"
                lean_ffmpeg.prepare()
                self.assertEqual(builds.count("sh"), 3)

                (output / "bin/ffprobe").unlink()
                lean_ffmpeg.prepare()
                self.assertEqual(builds.count("sh"), 4)
