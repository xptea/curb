"""Packaging regressions seen on clean release runners."""
import hashlib
import io
import os
import tarfile
import tempfile
import unittest
import zipfile
from pathlib import Path
from unittest.mock import patch

from bundler.linux_runtime import collect_tool_libraries
from bundler import native_tools
import toolchain_cache


class LinuxToolchainTests(unittest.TestCase):
    def test_ffmpeg_bundles_transitive_runtime_libraries_without_host_copies(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            tools, runtime = root / "tools", root / "runtime"
            for directory in (tools / "bin", tools / "lib", runtime / "lib"):
                directory.mkdir(parents=True)
            for path in (tools / "lib/libavcodec.so", runtime / "lib/libmp3lame.so.0", runtime / "lib/libgnutls.so.30", runtime / "lib/libnettle.so.8"):
                path.write_bytes(b"native library fixture")

            def output(args, **kwargs):
                name = Path(args[-1]).name
                imports = {"ffmpeg": ["libavcodec.so", "libmp3lame.so.0"], "ffprobe": ["libavcodec.so"],
                           "libavcodec.so": ["libgnutls.so.30"], "libgnutls.so.30": ["libnettle.so.8"],
                           "libnettle.so.8": ["libc.so.6", "ld-linux-x86-64.so.2"], "libmp3lame.so.0": ["libm.so.6"]}
                return "\n".join("  NEEDED " + item for item in imports[name])

            with patch("bundler.linux_runtime.subprocess.check_output", side_effect=output):
                copied = collect_tool_libraries(tools, runtime)
            self.assertEqual(copied, ["libgnutls.so.30", "libmp3lame.so.0", "libnettle.so.8"])
            self.assertTrue((tools / "lib/libnettle.so.8").is_file())
            self.assertFalse((tools / "lib/libc.so.6").exists())


class ToolchainReuseTests(unittest.TestCase):
    def test_cache_restores_matching_tools_from_a_different_release_tag(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            payload = io.BytesIO()
            with tarfile.open(fileobj=payload, mode="w:gz") as archive:
                info = tarfile.TarInfo("tools/ffmpeg")
                info.mode = 0o755
                content = b"verified cached native ffmpeg"
                info.size = len(content)
                archive.addfile(info, io.BytesIO(content))
            download = io.BytesIO()
            with zipfile.ZipFile(download, "w") as files:
                files.writestr("toolchain.tar.gz", payload.getvalue())
            data = download.getvalue()
            artifact = {"id": 12, "name": "matching-recipe", "expired": False,
                        "digest": "sha256:" + hashlib.sha256(data).hexdigest(), "workflow_run": {"id": 10}}
            run = {"id": 10, "head_branch": "v0.1.1", "event": "push", "head_repository": {"full_name": "xptea/curb"},
                   "path": ".github/workflows/release.yml"}

            def transfer(*args, **kwargs):
                kwargs["stdout"].write(data)

            with patch.object(toolchain_cache, "ROOT", root), \
                    patch.object(toolchain_cache, "cache_key", return_value="matching-recipe"), \
                    patch.object(toolchain_cache, "api", side_effect=[{"artifacts": [artifact]}, run]), \
                    patch.object(toolchain_cache.subprocess, "run", side_effect=transfer), \
                    patch.dict("os.environ", {"GITHUB_OUTPUT": str(root / "outputs"), "GITHUB_REF_NAME": "v0.1.2"}):
                toolchain_cache.restore("linux-amd64")
            self.assertEqual((root / "tools/ffmpeg").read_bytes(), content)
            # Windows uses PE executables, not Unix execute permission bits.
            if os.name != "nt":
                self.assertTrue((root / "tools/ffmpeg").stat().st_mode & 0o111)
            self.assertIn("hit=true", (root / "outputs").read_text())

    def test_cache_recipe_key_changes_with_build_configuration(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            files = ("packaging/tools.json", "packaging/native/build-ffmpeg.sh", "scripts/toolchain_cache.py", "scripts/bundler/native_runtime.py")
            for name in files:
                path = root / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("original input")
            with patch.object(toolchain_cache, "ROOT", root), patch.object(toolchain_cache, "__file__", str(root / "scripts/toolchain_cache.py")):
                original = toolchain_cache.cache_key("darwin-arm64")
                self.assertNotEqual(original, toolchain_cache.cache_key("darwin-amd64"))
                (root / "packaging/native/build-ffmpeg.sh").write_text("changed encoder configuration")
                self.assertNotEqual(original, toolchain_cache.cache_key("darwin-arm64"))

    def test_native_cache_hit_skips_compiler_and_network(self):
        with tempfile.TemporaryDirectory() as temporary:
            tools = Path(temporary)
            for name in ("ffmpeg", "ffprobe", "qjs", "yt-dlp", "manifest.json", "licenses/ffmpeg-LICENSE"):
                path = tools / name
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("cached output")
            (tools / "curb-build-id").write_text("matching-inputs")
            with patch.object(native_tools, "TOOLS", tools), \
                    patch.object(native_tools.platform, "system", return_value="Darwin"), \
                    patch.object(native_tools.platform, "machine", return_value="arm64"), \
                    patch.object(native_tools, "native_build_id", return_value="matching-inputs"), \
                    patch.object(native_tools, "release", side_effect=AssertionError("unexpected download")), \
                    patch.object(native_tools.subprocess, "run", side_effect=AssertionError("unexpected compiler")):
                if native_tools.os.name == "nt":
                    self.skipTest("macOS cache fixture")
                native_tools.prepare_native("darwin-arm64")

    def test_only_release_workflow_artifacts_are_trusted(self):
        run = {"event": "push", "head_repository": {"full_name": "xptea/curb"}, "path": ".github/workflows/release.yml"}
        self.assertTrue(toolchain_cache.trusted_run(run))
        self.assertFalse(toolchain_cache.trusted_run(dict(run, event="pull_request")))
        self.assertFalse(toolchain_cache.trusted_run(dict(run, head_repository={"full_name": "other/curb"})))
        self.assertFalse(toolchain_cache.trusted_run(dict(run, path=".github/workflows/ci.yml")))


class WindowsBuildShellTests(unittest.TestCase):
    def test_windows_uses_explicit_msys2_bash_instead_of_wsl(self):
        with tempfile.TemporaryDirectory(prefix="MSYS2 runtime ") as temporary:
            shell = Path(temporary) / "usr/bin/bash.exe"
            shell.parent.mkdir(parents=True)
            shell.write_bytes(b"MSYS2 executable fixture")
            with patch.object(native_tools, "command", return_value=str(shell)) as command, \
                    patch.object(native_tools.shutil, "which", return_value="C:/Windows/System32/bash.exe"):
                self.assertEqual(native_tools.ffmpeg_shell(True), str(shell))
                command.assert_called_once_with("cygpath", "-w", "/usr/bin/bash.exe")

    def test_missing_msys2_bash_does_not_fall_back_to_wsl(self):
        with tempfile.TemporaryDirectory() as temporary:
            shell = Path(temporary) / "missing/bash.exe"
            with patch.object(native_tools, "command", return_value=str(shell)):
                with self.assertRaisesRegex(RuntimeError, "MSYS2 Bash is missing"):
                    native_tools.ffmpeg_shell(True)

