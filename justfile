set positional-arguments

# List available tasks.
default:
    @just --list

# Run quality checks, tests, and race detection.
check: quality test race

# Check formatting, dependencies, static analysis, and vulnerabilities.
quality: fmt-check tidy-check lint govulncheck

# Build the library; additional arguments are passed to go build.
build *args:
    go build "$@" ./...

# Run the test suite; accepts additional go test flags.
test *args:
    go test "$@" ./...

# Run the test suite with the race detector; accepts additional go test flags.
race *args:
    go test -race "$@" ./...

# Report test coverage; accepts additional go test flags.
coverage *args:
    go test -race -cover "$@" ./...

# Run Go's static analysis; accepts additional go vet flags.
vet *args:
    go vet "$@" ./...

# Run the pinned linters against the library; accepts golangci-lint run flags.
lint *args:
    go tool -modfile=tools/golangci-lint/go.mod golangci-lint run "$@" ./...

# Check for reachable vulnerabilities; accepts additional flags.
govulncheck *args:
    go tool -modfile=tools/tools.mod govulncheck "$@" ./...

# Format Go sources and update imports.
fmt:
    go tool -modfile=tools/tools.mod goimports -w .

# Check Go formatting and imports without modifying files.
fmt-check:
    #!/usr/bin/env sh
    set -eu
    unformatted=$(go tool -modfile=tools/tools.mod goimports -l .)
    if [ -n "$unformatted" ]; then
        printf '%s\n' "$unformatted"
        exit 1
    fi

# Regenerate enum and presence methods with the pinned tools.
generate:
    go generate ./...

# Run benchmarks with allocation statistics; accepts additional go test flags.
bench *args:
    go test -run '^$' -bench . -benchmem "$@" ./...

# Fuzz one named target for the given duration with the given worker count.
fuzz target='FuzzParse' duration='30s' parallel='4':
    go test -run '^$' -fuzz "^${1}$" -fuzztime "$2" -parallel "$3" .

# Run every fuzz target sequentially; duration applies to each target.
fuzz-all duration='30s' parallel='4': (fuzz 'FuzzParse' duration parallel) (fuzz 'FuzzParseHex' duration parallel) (fuzz 'FuzzParseReading' duration parallel) (fuzz 'FuzzParsePatternFrom' duration parallel) (fuzz 'FuzzParseReadingFrom' duration parallel)

# Tidy dependencies for the library and the pinned development tools.
tidy:
    go mod tidy
    go mod tidy -modfile=tools/tools.mod
    go -C tools/golangci-lint mod tidy

# Check dependency files without modifying them.
tidy-check:
    go mod tidy -diff
    go mod tidy -diff -modfile=tools/tools.mod
    go -C tools/golangci-lint mod tidy -diff
