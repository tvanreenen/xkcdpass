# Contributing

## Development setup

This project uses `just` for the common local workflows and keeps Go caches inside the repo when invoked through those commands.

```sh
just test
just build
just run -- --words 6
```

If you prefer running Go commands directly, use repo-local caches to avoid polluting global state and to keep behavior consistent with the `just` tasks:

```sh
GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./...
GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go build ./cmd/xkcdpass
```

## Tests

Regular tests run with `go test ./...` and include validation of the embedded EFF wordlist. That means CI still checks:

- the list contains exactly 7,776 entries
- entries are lowercase
- entries do not contain surrounding or internal whitespace
- entries are unique

These checks run as normal tests; you do not need to do anything special to enable them.

## Benchmarks and timing

Benchmarks are opt-in and do not run under a normal `go test ./...`.

To run the benchmark coverage for the startup-sensitive paths:

```sh
GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test -bench . -benchmem ./internal/wordlist ./internal/generator
```

This runs:

- `BenchmarkWords`
- `BenchmarkValidate`
- `BenchmarkGenerate4Words`

To measure end-to-end CLI startup time locally:

```sh
GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go build -o xkcdpass ./cmd/xkcdpass
time sh -c 'i=0; while [ $i -lt 500 ]; do ./xkcdpass >/dev/null; i=$((i+1)); done'
```

The timing loop is only for manual performance checks. It is not part of the normal test suite or CI.

## Releases

Release automation is handled through the existing `just` targets.

Use the full release command for the normal end-to-end flow:

```sh
just release v0.1.0
```

That command runs the three release stages in order:

- `just release-build v0.1.0`
- `just release-publish v0.1.0`
- `just release-tap v0.1.0`

Use the individual stage commands only when a later stage fails or when you need to rerun one part of the pipeline without repeating the earlier successful stages:

```sh
just release-build v0.1.0
just release-publish v0.1.0
just release-tap v0.1.0
```
