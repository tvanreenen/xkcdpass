export GOCACHE := justfile_directory() / ".gocache"

default:
    @just --list --unsorted

test:
    go test ./...
    ./scripts/test-release-commands.sh
    ./scripts/test-publish-release.sh

build:
    go build ./cmd/xkcdpass

# Publish the source release through the environment-gated distribution workflow.
release version:
    ./scripts/release.sh {{quote(version)}}

# Dispatch the tap-owned Homebrew update after the source release is published.
publish-homebrew version:
    ./scripts/publish-homebrew.sh {{quote(version)}}

[positional-arguments]
[script]
run *args:
    if [ "${1:-}" = "--" ]; then
      shift
    fi

    go run ./cmd/xkcdpass "$@"

clean:
    rm -rf dist .gocache xkcdpass
