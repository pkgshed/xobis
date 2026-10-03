"""Check rendered documentation links, anchors, assets, and search indexes."""

import json
import re
from html.parser import HTMLParser
from pathlib import Path
from typing import TypedDict
from urllib.parse import unquote, urljoin, urlsplit


class SearchEntry(TypedDict):
    href: str


class Page(HTMLParser):
    def __init__(self, content: str) -> None:
        super().__init__()
        self.ids: set[str] = set()
        self.links: list[str] = []
        self.feed(content)

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        attributes = dict(attrs)
        identifier = attributes.get("id")
        if identifier:
            self.ids.add(identifier)
        name = attributes.get("name")
        if tag == "a" and name:
            self.ids.add(name)
        for attribute in ("href", "src", "poster"):
            value = attributes.get(attribute)
            if value:
                self.links.append(value)
        if tag == "meta" and (attributes.get("http-equiv") or "").lower() == "refresh":
            _, separator, url = (attributes.get("content") or "").partition("url=")
            if separator:
                self.links.append(url.strip("'\""))


def check_site(root: Path, base_url: str) -> None:
    root = Path(root).resolve()
    base = urlsplit(base_url)
    pages = {file: Page(file.read_text(encoding="utf-8")) for file in root.rglob("*.html")}
    errors: set[str] = set()

    def check(source: Path, link: str) -> None:
        source_url = urljoin(base_url, source.relative_to(root).as_posix())
        url = urlsplit(urljoin(source_url, link))
        if (url.scheme, url.netloc) != (base.scheme, base.netloc):
            return
        if not url.path.startswith(base.path):
            errors.add(f"{source.relative_to(root)}: link outside the site: {link}")
            return
        target = (root / unquote(url.path[len(base.path) :])).resolve()
        if not target.is_relative_to(root):
            errors.add(f"{source.relative_to(root)}: link outside the output: {link}")
            return
        if target.is_dir():
            target /= "index.html"
        if not target.is_file():
            errors.add(f"{source.relative_to(root)}: missing destination: {link}")
        elif url.fragment and target in pages and unquote(url.fragment) not in pages[target].ids:
            errors.add(f"{source.relative_to(root)}: missing anchor: {link}")

    for source, page in pages.items():
        for link in page.links:
            check(source, link)
    # The search script loads its JSON index and engine dynamically.
    for source in root.rglob("*.search*.js"):
        for link in re.findall(r"""["']([^"']+\.(?:json|js))["']""", source.read_text()):
            check(source, link)
    for source in root.rglob("*.search-data*.json"):
        prefix = base.path + source.parent.relative_to(root).as_posix() + "/"
        entries: list[SearchEntry] = json.loads(source.read_text())
        for entry in entries:
            link = entry["href"]
            if not link.startswith(prefix):
                errors.add(
                    f"{source.relative_to(root)}: search crosses documentation versions: {link}"
                )
            check(source, link)
    if errors:
        raise ValueError("Broken documentation links:\n" + "\n".join(sorted(errors)))
    print(f"Checked {len(pages)} HTML pages and their local links, assets, and search indexes")
