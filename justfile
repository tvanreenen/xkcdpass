default:
  @just --list --unsorted

test:
  GOCACHE={{justfile_directory()}}/.gocache go test ./...

build:
  GOCACHE={{justfile_directory()}}/.gocache go build ./cmd/xkcdpass

run *args:
  GOCACHE={{justfile_directory()}}/.gocache go run ./cmd/xkcdpass -- {{args}}

release version:
  ./scripts/release.sh {{version}}

clean:
  rm -rf dist .gocache xkcdpass
