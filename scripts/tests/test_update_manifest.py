import unittest

from update_manifest import build_manifest


class UpdateManifestTests(unittest.TestCase):
    def release(self, version="0.2.0"):
        name = f"Curb-{version}-x86_64.AppImage"
        return {"tag_name": f"v{version}", "draft": False, "prerelease": False, "published_at": "2026-10-16T00:00:00Z",
                "assets": [{"name": name, "state": "uploaded", "size": 12345, "digest": "sha256:" + "a" * 64,
                            "browser_download_url": f"https://github.com/xptea/curb/releases/download/v{version}/{name}"}]}

    def test_uses_only_published_platform_assets(self):
        manifest = build_manifest(self.release())
        self.assertEqual(manifest["version"], "0.2.0")
        self.assertEqual(list(manifest["platforms"]), ["linux-amd64"])
        self.assertEqual(manifest["platforms"]["linux-amd64"]["sha256"], "a" * 64)

    def test_rejects_missing_digest_wrong_repository_and_drafts(self):
        for change in ("digest", "url", "draft", "empty"):
            release = self.release()
            if change == "digest": release["assets"][0]["digest"] = ""
            if change == "url": release["assets"][0]["browser_download_url"] = "https://example.com/app"
            if change == "draft": release["draft"] = True
            if change == "empty": release["assets"] = []
            with self.subTest(change=change), self.assertRaises(ValueError):
                build_manifest(release)

    def test_older_release_cannot_downgrade_the_feed(self):
        previous = build_manifest(self.release("0.10.0"))
        self.assertEqual(build_manifest(self.release("0.9.0"), previous), previous)
