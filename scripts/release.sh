#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

version="$1"
if [[ "${version}" != v* ]]; then
  echo "version must start with v, for example v0.1.0" >&2
  exit 1
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist_dir="${repo_root}/dist"
build_dir="${dist_dir}/build"
binary_name="xkcdpass"

mkdir -p "${build_dir}"
rm -rf "${build_dir:?}/"*
rm -f "${dist_dir}/${binary_name}_${version}_"*.tar.gz "${dist_dir}/checksums.txt"

export GOCACHE="${repo_root}/.gocache"

platforms=(
  "darwin arm64"
  "linux amd64"
)

for platform in "${platforms[@]}"; do
  read -r goos goarch <<<"${platform}"

  target_dir="${build_dir}/${binary_name}_${version}_${goos}_${goarch}"
  mkdir -p "${target_dir}"

  GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 \
    go build \
      -ldflags="-s -w -X main.version=${version}" \
      -o "${target_dir}/${binary_name}" \
      ./cmd/xkcdpass

  tar -C "${build_dir}" -czf "${dist_dir}/${binary_name}_${version}_${goos}_${goarch}.tar.gz" \
    "${binary_name}_${version}_${goos}_${goarch}"
done

(
  cd "${dist_dir}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum ./*.tar.gz > checksums.txt
  else
    shasum -a 256 ./*.tar.gz > checksums.txt
  fi
)

echo "release artifacts written to ${dist_dir}"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required to create a draft GitHub release" >&2
  exit 1
fi

if gh release view "${version}" >/dev/null 2>&1; then
  echo "GitHub release ${version} already exists; refusing to overwrite" >&2
  exit 1
fi

echo "creating draft GitHub release ${version}"
gh release create "${version}" \
  "${dist_dir}/${binary_name}_${version}_"*.tar.gz \
  "${dist_dir}/checksums.txt" \
  --draft \
  --title "${version}" \
  --generate-notes
