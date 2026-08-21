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

clean:
    rm -rf dist .gocache xkcdpass
