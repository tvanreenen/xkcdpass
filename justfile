export GOCACHE := justfile_directory() / ".gocache"

default:
    @just --list --unsorted

test:
    go test ./...

build:
    go build ./cmd/xkcdpass

[positional-arguments]
[script]
run *args:
    if [ "${1:-}" = "--" ]; then
      shift
    fi

    go run ./cmd/xkcdpass "$@"

release version:
    ./scripts/release.sh all {{ version }}

release-build version:
    ./scripts/release.sh build {{ version }}

release-publish version:
    ./scripts/release.sh publish {{ version }}

release-tap version:
    ./scripts/release.sh tap {{ version }}

clean:
    rm -rf dist .gocache xkcdpass
