"""Build the development and latest release documentation with Hugo."""

import argparse
import html
import io
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile
from collections.abc import Iterable
from pathlib import Path
from typing import Literal
from urllib.parse import urlsplit

from check import check_site

ROOT = Path(__file__).resolve().parents[2]
SITE = Path(__file__).resolve().parent
HUGO = ("go", "-C", str(SITE), "tool", "hugo")
STABLE_TAG = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)")


def git(repo: Path, *arguments: str) -> bytes:
    return subprocess.check_output(["git", "-C", str(repo), *arguments])


def hugo(*arguments: str) -> None:
    # Share Go's module cache instead of downloading the tool's dependencies again.
    environment = os.environ.copy()
    environment["GOMODCACHE"] = subprocess.check_output(
        ["go", "-C", str(SITE), "env", "GOMODCACHE"], text=True
    ).strip()
    subprocess.run([*HUGO, *arguments], env=environment, check=True)


def select_latest(tags: Iterable[str]) -> str | None:
    versions = [
        (tuple(map(int, match.groups())), tag)
        for tag in tags
        if (match := STABLE_TAG.fullmatch(tag))
    ]
    return max(versions)[1] if versions else None


def latest_tag(repo: Path) -> str | None:
    return select_latest(git(repo, "tag", "--list").decode().splitlines())


def normalize_base_url(value: str) -> str:
    url = urlsplit(value)
    if url.scheme not in ("http", "https") or not url.netloc or url.query or url.fragment:
        raise ValueError("The base URL must be an absolute HTTP(S) URL without a query or fragment")
    return value.rstrip("/") + "/"


def snapshot(repo: Path, ref: str, destination: Path) -> Path:
    archive = git(repo, "archive", "--format=tar", ref, "docs/")
    with tarfile.open(fileobj=io.BytesIO(archive)) as source:
        source.extractall(destination, filter="data")
    return destination / "docs"


def configuration(
    path: Path,
    content: Path,
    base_url: str,
    version: str,
    tag: str | None,
    ref: str,
    preview: bool = False,
) -> str:
    if not (content / "index.md").is_file():
        raise ValueError(f"Documentation index missing: {content / 'index.md'}")
    config = {
        "params": {
            "DocsVersion": version,
            "DocsLatestTag": tag or "",
            "DocsBaseURL": base_url,
            "DocsRef": ref,
            "DocsPreview": preview,
            "DocsMainURL": base_url if preview else base_url + "main/",
        },
        "module": {
            "mounts": [
                {"source": str(content), "target": "content", "files": ["! index.md"]},
                {"source": str(content / "index.md"), "target": "content/_index.md"},
            ]
        },
    }
    path.write_text(json.dumps(config), encoding="utf-8")
    return f"{SITE / 'hugo.toml'},{path}"


def redirect(destination: str) -> str:
    url = html.escape(destination, quote=True)
    return (
        f'<!doctype html><html lang="en"><head><meta charset="utf-8">'
        f'<title>xobis documentation</title><meta http-equiv="refresh" content="0;url={url}">'
        f'<link rel="canonical" href="{url}"></head><body>'
        f'<p><a href="{url}">Open the documentation</a></p></body></html>\n'
    )


def build_site(repo: Path, output: Path, base_url: str) -> None:
    base_url = normalize_base_url(base_url)
    tag = latest_tag(repo)
    # Replace only output produced by this tool, and only after both builds pass.
    if output.exists() and any(output.iterdir()) and not (output / ".xobis-docs").is_file():
        raise ValueError(f"Output directory contains unrelated files: {output}")
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="xobis-docs-", dir=output.parent) as temporary:
        stage = Path(temporary)
        public = stage / "public"
        variants = [("main", repo / "docs", "Development", "main")]
        if tag:
            content = snapshot(repo, tag, stage / "release")
            variants.append(("latest", content, tag, tag))
        for name, content, version, ref in variants:
            config = configuration(stage / f"{name}.json", content, base_url, version, tag, ref)
            hugo(
                "--source",
                str(SITE),
                "--config",
                config,
                "--baseURL",
                base_url + name + "/",
                "--destination",
                str(public / name),
                "--minify",
                "--panicOnWarning",
            )
        public.joinpath("index.html").write_text(
            redirect(base_url + ("latest/" if tag else "main/"))
        )
        public.joinpath(".xobis-docs").touch()
        check_site(public, base_url)
        if output.exists():
            shutil.rmtree(output)
        shutil.move(public, output)
    print(f"Documentation built in {output}; latest release: {tag or 'none'}", flush=True)


def serve_site(repo: Path, port: int) -> None:
    # Native mounts let Hugo watch the original Markdown, including index.md.
    with tempfile.TemporaryDirectory(prefix="xobis-docs-preview-") as temporary:
        config = configuration(
            Path(temporary) / "preview.json",
            repo / "docs",
            f"http://localhost:{port}/",
            "Development",
            None,
            "main",
            preview=True,
        )
        hugo(
            "server",
            "--source",
            str(SITE),
            "--config",
            config,
            "--baseURL",
            f"http://localhost:{port}/",
            "--port",
            str(port),
            "--disableFastRender",
            "--renderToMemory",
        )


class Arguments(argparse.Namespace):
    command: Literal["build", "serve"]
    base_url: str
    output: Path
    port: int


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("build", "serve"))
    parser.add_argument("--base-url", default="https://pkgshed.github.io/xobis/")
    parser.add_argument("--output", type=Path, default=ROOT / "public")
    parser.add_argument("--port", type=int, default=1313)
    args = parser.parse_args(namespace=Arguments())
    try:
        if args.command == "build":
            build_site(ROOT, args.output.resolve(), args.base_url)
        else:
            serve_site(ROOT, args.port)
    except KeyboardInterrupt:
        return 0
    except (ValueError, OSError, subprocess.CalledProcessError) as error:
        print(f"Documentation build failed: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
