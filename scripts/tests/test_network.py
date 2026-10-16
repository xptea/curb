"""Verified download caches must work offline and reject corrupted inputs."""
import hashlib
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from bundler import network


class DownloadCacheTests(unittest.TestCase):
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
