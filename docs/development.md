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

Commit the resulting module and checksum changes together. Keep this module exclusive to golangci-lint and do not upgrade its transitive dependencies individually. `just tidy` and `just tidy-check` handle all three module files, running inside the linter's directory to preserve its isolation.
