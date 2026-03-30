set shell := ["zsh", "-cu"]

default:
  @just --list

test:
  GOCACHE={{justfile_directory()}}/.gocache go test ./...

build:
  GOCACHE={{justfile_directory()}}/.gocache go build ./cmd/xkcdpass

run *args:
  GOCACHE={{justfile_directory()}}/.gocache go run ./cmd/xkcdpass -- {{args}}

release version:
  ./scripts/release.sh {{version}}
