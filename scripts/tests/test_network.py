"""Verified download caches must work offline and reject corrupted inputs."""
import hashlib
import io
import tempfile
import unittest
import urllib.error
from pathlib import Path
from unittest.mock import patch

from bundler import network


class DownloadCacheTests(unittest.TestCase):
    def test_temporary_network_failure_is_retried(self):
        with patch.object(network.urllib.request, "urlopen", side_effect=[urllib.error.URLError(TimeoutError()), io.BytesIO(b"retried response")]) as request, \
                patch.object(network.time, "sleep"):
            self.assertEqual(network.fetch("https://example.invalid/metadata"), b"retried response")
            self.assertEqual(request.call_count, 2)

    def test_missing_resource_is_not_retried(self):
        error = urllib.error.HTTPError("https://example.invalid/missing", 404, "Not found", {}, None)
        with patch.object(network.urllib.request, "urlopen", side_effect=error) as request, \
                patch.object(network.time, "sleep") as sleep:
            with self.assertRaises(urllib.error.HTTPError):
                network.fetch(error.url)
            error.close()
            self.assertEqual(request.call_count, 1)
            sleep.assert_not_called()

    def test_release_api_checksum_avoids_a_separate_asset_download(self):
        info = {"assets": [{"name": "yt-dlp.exe", "digest": "sha256:" + "a" * 64}]}
        with patch.object(network, "fetch", side_effect=AssertionError("unnecessary checksum download")):
            self.assertEqual(network.asset_checksum(info, "yt-dlp.exe"), "a" * 64)

    def test_interrupted_download_restarts_and_promotes_only_verified_bytes(self):
        class Interrupted(io.BytesIO):
            def read(self, size=-1):
                if self.tell():
                    raise ConnectionResetError("interrupted download")
                return super().read(3)

        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            data = b"complete verified download"
            with patch.object(network, "CACHE", root), \
                    patch.object(network.urllib.request, "urlopen", side_effect=[Interrupted(b"partial"), io.BytesIO(data)]) as request, \
                    patch.object(network.time, "sleep"):
                path = network.download("https://example.invalid/tool", "tool", hashlib.sha256(data).hexdigest())
            self.assertEqual(path.read_bytes(), data)
            self.assertEqual(request.call_count, 2)
            self.assertFalse((root / "tool.partial").exists())

    def test_verified_cached_asset_needs_no_network(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            data = b"verified cached build input"
            path = root / "tool"
            path.write_bytes(data)
            digest = hashlib.sha256(data).hexdigest()
            with patch.object(network, "CACHE", root), \
                    patch.object(network.urllib.request, "urlopen", side_effect=AssertionError("unexpected network request")):
                self.assertEqual(network.download("https://example.invalid/tool", "tool", digest), path)

    def test_corrupt_cached_asset_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            path = root / "tool"
            path.write_bytes(b"corrupted input")
            with patch.object(network, "CACHE", root):
                with self.assertRaisesRegex(RuntimeError, "Checksum mismatch"):
                    network.download("https://example.invalid/tool", "tool", hashlib.sha256(b"expected input").hexdigest())
            self.assertFalse(path.exists())
