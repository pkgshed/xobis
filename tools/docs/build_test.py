"""Exercise release selection, real Hugo output, and broken link detection."""

import tempfile
import unittest
from pathlib import Path

from build import build_site, git, select_latest
from check import check_site

BASE_URL = "https://example.org/xobis/"


class SelectionTests(unittest.TestCase):
    def test_selects_highest_stable_version_in_any_order(self) -> None:
        self.assertEqual(select_latest(["v0.10.0", "v0.9.0", "v0.8.1", "v1.0.0-rc.1"]), "v0.10.0")
        self.assertEqual(select_latest(["v2.0.0", "v10.0.0", "v9.20.30"]), "v10.0.0")

    def test_no_eligible_tag(self) -> None:
        self.assertIsNone(select_latest([]))
        self.assertIsNone(
            select_latest(["v1.0.0-beta.1", "v1.0.0+build", "other", "v01.0.0", "1.0.0"])
        )


class LinkTests(unittest.TestCase):
    def test_local_links_and_assets_with_project_prefix(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "main").mkdir()
            (root / "main/index.html").write_text(
                '<h1 id="hello">Hello</h1><a href="#hello">Anchor</a>'
                '<a href="https://external.example/">External</a>'
                '<a href="mailto:someone@example.org">Email</a><script src="app.js"></script>'
            )
            (root / "main/app.js").touch()
            check_site(root, BASE_URL)

    def test_missing_page_anchor_and_asset_are_reported(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "index.html").write_text(
                '<a href="missing.md">Page</a><a href="#missing">Anchor</a><img src="missing.svg">'
            )
            with self.assertRaises(ValueError) as error:
                check_site(root, BASE_URL)
            self.assertIn("missing.md", str(error.exception))
            self.assertIn("missing anchor", str(error.exception))
            self.assertIn("missing.svg", str(error.exception))

    def test_search_cannot_mix_versions(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "main").mkdir()
            (root / "latest").mkdir()
            (root / "latest/index.html").touch()
            (root / "main/en.search-data.json").write_text('[{"href":"/xobis/latest/"}]')
            with self.assertRaisesRegex(ValueError, "search crosses documentation versions"):
                check_site(root, BASE_URL)


class HugoTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory(prefix="xobis-docs-test-")
        self.addCleanup(self.temporary.cleanup)
        self.repo = Path(self.temporary.name) / "repo"
        self.repo.mkdir()
        git(self.repo, "init", "--quiet", "--initial-branch=main")
        (self.repo / "docs").mkdir()
        (self.repo / "docs/index.md").write_text(
            "---\ntitle: Documentation\nweight: 1\n---\n# Documentation\n[Topic](topic.md#anchor)\n"
        )
        self.topic = self.repo / "docs/topic.md"
        self.write_topic("Released marker.")
        self.commit()
        self.output = Path(self.temporary.name) / "public"

    def write_topic(self, content: str) -> None:
        self.topic.write_text(
            "---\ntitle: Topic\nweight: 10\n---\n# Topic\n## Anchor\n"
            + content
            + "\n[Home](index.md#documentation)\n"
        )

    def commit(self) -> None:
        git(self.repo, "add", "docs")
        git(
            self.repo,
            "-c",
            "user.name=Docs test",
            "-c",
            "user.email=docs@example.org",
            "commit",
            "--quiet",
            "--message=Documentation fixture",
        )

    def test_no_release_builds_development_and_redirect(self) -> None:
        git(self.repo, "tag", "v1.0.0-rc.1")
        build_site(self.repo, self.output, BASE_URL)
        home = (self.output / "main/index.html").read_text()
        self.assertIn("No release yet", home)
        self.assertNotIn("Latest release (", home)
        self.assertIn("/xobis/main/topic/#anchor", home)
        self.assertIn("/xobis/main/", (self.output / "index.html").read_text())
        self.assertFalse((self.output / "latest").exists())

    def test_release_uses_tagged_content_and_rebuild_removes_stale_release(self) -> None:
        git(self.repo, "tag", "v0.9.0")
        git(self.repo, "tag", "v0.10.0")
        self.write_topic("Development marker.")
        (self.repo / "docs/new.md").write_text("---\ntitle: New\n---\n# Only on main\n")
        self.commit()
        git(self.repo, "tag", "v0.8.1")
        git(self.repo, "tag", "v1.0.0-rc.1")
        build_site(self.repo, self.output, BASE_URL)
        release = (self.output / "latest/topic/index.html").read_text()
        development = (self.output / "main/topic/index.html").read_text()
        self.assertIn("Released marker.", release)
        self.assertNotIn("Development marker.", release)
        self.assertIn("Development marker.", development)
        self.assertIn("Latest release (v0.10.0)", development)
        self.assertNotIn("Only on main", release)
        self.assertFalse((self.output / "latest/new").exists())
        self.assertIn("/xobis/latest/", (self.output / "index.html").read_text())
        for tag in ("v0.9.0", "v0.10.0", "v0.8.1"):
            git(self.repo, "tag", "--delete", tag)
        build_site(self.repo, self.output, BASE_URL)
        self.assertFalse((self.output / "latest").exists())
        self.assertIn("/xobis/main/", (self.output / "index.html").read_text())


if __name__ == "__main__":
    unittest.main()
