default:
  @just --list --unsorted

test:
  GOCACHE={{justfile_directory()}}/.gocache go test ./...

build:
  GOCACHE={{justfile_directory()}}/.gocache go build ./cmd/xkcdpass

run *args:
  GOCACHE={{justfile_directory()}}/.gocache go run ./cmd/xkcdpass -- {{args}}

release version:
  ./scripts/release.sh all {{version}}

release-build version:
  ./scripts/release.sh build {{version}}

release-publish version:
  ./scripts/release.sh publish {{version}}

release-tap version:
  ./scripts/release.sh tap {{version}}

clean:
  rm -rf dist .gocache xkcdpass
