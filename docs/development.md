---
title: Development
weight: 70
---

# Development

This document is targetted at developers who wish to expand on the `xobis`-package directly.

Use stable Go and `just` for development; the library supports Go 1.24 and later. Run `just` to list available tasks.

`just quality` checks formatting, dependencies, lint findings, and vulnerabilities
both locally and in CI. `just test` runs the tests; `just race` adds race detection. `just check` runs all three.

Pass Go flags to tests and benchmarks, or select a fuzz target:

```sh
just test -run '^TestParseReading$'
just bench -benchtime=1x
just fuzz FuzzParseReading 30s 4
```

Regenerate enum and presence methods with `just generate`. Development tools are pinned in `tools/tools.mod`; golangci-lint is pinned independently in `tools/golangci-lint/go.mod` and `go.sum`.

Run `just lint` locally. It checks the library and its tests using the linters explicitly enabled in `.golangci.yml`: govet, staticcheck, unused, errcheck, errorlint, and testifylint. Tests are structured following the arrange-act-assert pattern, using the `testify`-framework with `require` for prerequisites and `assert` for independent outcomes. 
Exact JSON comparisons are retained for serialization tests. The linter also allows the cursor's intentional direct comparisons with `io.EOF`. `just govulncheck` remains a separate vulnerability check, and `just vet` is available for a standalone Go vet run.

To update golangci-lint, select an exact release and run, for example:

```sh
go -C tools/golangci-lint get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
go -C tools/golangci-lint mod tidy
just check
```

Commit the resulting module and checksum changes together. Keep this module exclusive to golangci-lint and do not upgrade its transitive dependencies individually. `just tidy` and `just tidy-check` handle the library and both Go development tool modules, running inside the linter's directory to preserve its isolation.

## Documentation

The project documentation is located in the `docs/` directory, written in Markdown and rendered with Hugo and the Hugo Book theme into a static page. The uv tool and Go 1.27 or later is required to be installed and to work on the website. Hugo and the theme are pinned in a separate module under `tools/docs/`. The Python scripts invoke `go -C tools/docs tool hugo`, so Go downloads and compiles the pinned Hugo version on the first run and caches it for later runs.

The Python scripts have their own uv project in `tools/docs/` to improve reproducibility of the process. Local tasks and CI use `--locked` to keep the tooling consistent.

Hugo resolves public modules through the Go module proxy. When loading modules, Hugo changes `GOPATH` for its internal Go commands to a directory in its own cache. Unless `GOMODCACHE` is explicitly set, Go derives the module cache location from `GOPATH`. This would give Hugo a separate cache and cause it to download dependencies already fetched when compiling the Hugo tool.

The Python scripts therefore run `go -C tools/docs env GOMODCACHE` and pass that path explicitly in Hugo's environment. This preserves any existing `GOMODCACHE` override and keeps Hugo's module loading in the same cache as the Go tool. CI caches these modules and compiled tools, so later runs reuse both downloads and build results. Sharing the cache is a performance optimization; the site also builds correctly with separate caches.

```sh
just docs-serve   # Preview with live reload at http://localhost:1313/.
just docs-build   # Build public/main/ and, when available, public/latest/.
just docs-test    # Test version selection, release snapshots, and link checks.
just docs-quality # Check lint, formatting, annotations, and types with Ruff and ty.
just docs-fmt     # Sort Python imports and apply Ruff formatting.
```

To update the Python tooling, run `uv lock --project tools/docs --upgrade` to upgrade all dependencies within the project's version constraints, then `just docs-quality`, `just docs-test`, and `just docs-build`. Commit `pyproject.toml` and `uv.lock` changes together.

`docs-build` uses the working tree for development documentation and the highest stable `vMAJOR.MINOR.PATCH` tag for release documentation. Prereleases and other tag formats are ignored. Before the first stable tag, it builds development documentation only. The site root redirects to the release documentation when available, otherwise to development documentation. Each version has its own navigation and search index.

Existing relative Markdown links, including heading anchors, work on GitHub and the website. Add `title` and `weight` front matter to new pages to set their sidebar title and order. Hugo mounts `docs/index.md` as its homepage without renaming the source file. Website configuration and templates always come from the working tree; release content comes from its tag.

The build checks rendered local links, heading anchors, assets, and search indexes before writing the final output. Generated files in `public/` are ignored by Git. To override the default project URL, use `just docs-build --base-url https://example.org/xobis/`.

### Updating the Toolchain

To update Hugo deliberately, run `go -C tools/docs get -tool github.com/gohugoio/hugo@VERSION`. To update the theme, run `go -C tools/docs tool hugo mod get github.com/alex-shpak/hugo-book@VERSION`. Then run `just docs-quality`, `just docs-test`, and `just docs-build`, and commit `tools/docs/go.mod` and `go.sum` together. Documentation module dependencies are managed separately from `just tidy`.
