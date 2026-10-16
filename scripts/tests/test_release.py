import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import release
from bundler.native_runtime import collect_macos, collect_windows, macos_imports, windows_imports


class ReleaseTests(unittest.TestCase):
    def published(self):
        tag = "v0.2.0"
        return {"tag_name": tag, "published_at": "2026-10-16T00:00:00Z", "draft": False, "prerelease": False,
                "assets": [{"name": name, "state": "uploaded", "size": 123, "digest": "sha256:" + "a" * 64,
                            "browser_download_url": f"https://github.com/xptea/curb/releases/download/{tag}/{name}"}
                           for name in release.asset_names("0.2.0")]}

    def test_tag_controls_build_version_without_advertising_an_update(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "mygo.json").write_text('{"name":"Curb","version":"0.1.0"}')
            (root / "version.json").write_text('{"schema":1,"version":"0.1.0","platforms":{}}')
            before = (root / "version.json").read_bytes()
            with patch.object(release, "ROOT", root):
                self.assertEqual(release.prepare("v0.2.0"), "0.2.0")
            self.assertEqual(json.loads((root / "mygo.json").read_text())["version"], "0.2.0")
            self.assertEqual((root / "version.json").read_bytes(), before)
        for invalid in ("main", "0.2.0", "v1.0.0-beta", "v01.0.0", "v1.0.0/other"):
            with self.subTest(tag=invalid), self.assertRaises(ValueError):
                release.tag_version(invalid)

    def test_release_is_complete_only_after_all_platforms_are_published(self):
        info = self.published()
        with patch.object(release, "github_release", return_value=info):
            self.assertTrue(release.complete("v0.2.0"))
            info["assets"].pop()
            self.assertFalse(release.complete("v0.2.0"))
            info["assets"] = []
            self.assertFalse(release.complete("v0.2.0"))

    def test_no_release_is_created_for_partial_builds(self):
        with tempfile.TemporaryDirectory() as temporary, \
                patch.object(release, "github_release", return_value=None), \
                patch.object(release, "run") as run:
            with self.assertRaisesRegex(ValueError, "all four"):
                release.publish("v0.2.0", Path(temporary))
            run.assert_not_called()

    def test_complete_release_updates_feed_and_build_version_together(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "version.json").write_text('{"schema":1,"version":"0.1.0","platforms":{}}')
            (root / "mygo.json").write_text('{"version":"0.1.0"}')
            with patch.object(release, "ROOT", root), patch.object(release, "github_release", return_value=self.published()):
                release.publish("v0.2.0", None)
            feed = json.loads((root / "version.json").read_text())
            self.assertEqual(set(feed["platforms"]), set(release.PLATFORMS))
            self.assertEqual(feed["version"], "0.2.0")
            self.assertEqual(json.loads((root / "mygo.json").read_text())["version"], "0.2.0")

    def test_published_release_bytes_cannot_be_overwritten(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            for name in release.asset_names("0.2.0"):
                (root / name).write_bytes(b"different release build")
            with patch.object(release, "github_release", return_value=self.published()), patch.object(release, "run") as run:
                with self.assertRaisesRegex(ValueError, "different bytes"):
                    release.publish("v0.2.0", root)
                run.assert_not_called()


class NativeRuntimeTests(unittest.TestCase):
    def test_macos_universal_headers_are_not_dylib_dependencies(self):
        output = ("/tmp/yt-dlp (architecture x86_64):\n"
                  "\t/usr/lib/libSystem.B.dylib (compatibility version 1.0.0, current version 1.0.0)\n"
                  "/tmp/yt-dlp (architecture arm64):\n"
                  "\t/usr/lib/libSystem.B.dylib (compatibility version 1.0.0, current version 1.0.0)\n")
        self.assertEqual(macos_imports(output), ["/usr/lib/libSystem.B.dylib"])

    def test_macos_keeps_the_universal_pyinstaller_launcher_unchanged(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            tools = root / "tools"
            (tools / "licenses").mkdir(parents=True)
            for name in ("ffmpeg", "ffprobe", "qjs", "yt-dlp"):
                (tools / name).write_bytes(b"native binary fixture")

            def output(*args):
                if args[:2] == ("brew", "--prefix"): return str(root)
                if args[:2] == ("brew", "info"): return "{}"
                path = args[-1]
                return (f"{path} (architecture x86_64):\n"
                        "\t/usr/lib/libSystem.B.dylib (compatibility version 1.0.0)\n"
                        f"{path} (architecture arm64):\n"
                        "\t/usr/lib/libSystem.B.dylib (compatibility version 1.0.0)\n")

            with patch("bundler.native_runtime.command", side_effect=output), \
                    patch("bundler.native_runtime.subprocess.run") as run:
                self.assertEqual(collect_macos(root / "prefix", tools), [])
            self.assertFalse(any(call.args[0][-1] == str(tools / "yt-dlp") for call in run.call_args_list))

    def test_import_parsers(self):
        self.assertEqual(windows_imports(" DLL Name: KERNEL32.dll\n DLL Name: avcodec.dll"), ["KERNEL32.dll", "avcodec.dll"])
        self.assertEqual(macos_imports("ffmpeg:\n /usr/lib/libSystem.B.dylib (compatibility version 1.0.0)\n @rpath/libavcodec.dylib (compatibility version 1.0.0)"),
                         ["/usr/lib/libSystem.B.dylib", "@rpath/libavcodec.dylib"])

    def test_windows_collects_transitive_dependencies_and_skips_system_dlls(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            prefix, tools, mingw, system = [root / name for name in ("ffmpeg", "tools", "ucrt64", "Windows")]
            for path in (prefix / "bin", tools / "licenses", mingw / "bin", system / "System32"):
                path.mkdir(parents=True)
            for path in (prefix / "bin/avcodec.dll", mingw / "bin/libgnutls.dll", system / "System32/KERNEL32.dll", tools / "ffmpeg.exe", tools / "ffprobe.exe"):
                path.write_bytes(b"native binary fixture")

            def output(*args):
                if args[0] == "cygpath": return str(mingw)
                if args[0] == "pacman": return "gnutls 1.0"
                name = Path(args[-1]).name
                dependency = "libgnutls.dll" if name == "avcodec.dll" else "KERNEL32.dll" if name == "libgnutls.dll" else "avcodec.dll"
                return " DLL Name: " + dependency

            with patch.dict("os.environ", {"SystemRoot": str(system), "MINGW_PREFIX": "/ucrt64"}), \
                    patch("bundler.native_runtime.command", side_effect=output):
                names = collect_windows(prefix, tools)
            self.assertEqual(names, ["avcodec.dll", "libgnutls.dll"])
            self.assertTrue((tools / "libgnutls.dll").is_file())
            self.assertFalse((tools / "KERNEL32.dll").exists())
