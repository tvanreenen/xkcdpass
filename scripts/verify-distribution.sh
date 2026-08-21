#!/usr/bin/env bash

set -euo pipefail

fail() {
  echo "$1" >&2
  exit 1
}

if [[ $# -ne 1 ]]; then
  fail "usage: $0 <version>"
fi

version="$1"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist_dir="${repo_root}/dist"
binary_name="xkcdpass"
archives=(
  "${binary_name}_${version}_darwin_arm64.tar.gz"
  "${binary_name}_${version}_linux_amd64.tar.gz"
)

for archive in "${archives[@]}"; do
  archive_base="${archive%.tar.gz}"
  archive_path="${dist_dir}/${archive}"
  [[ -f "${archive_path}" ]] || fail "missing distribution archive: ${archive_path}"

  actual_entries="$(tar -tzf "${archive_path}" | LC_ALL=C sort)"
  expected_entries="${archive_base}/${binary_name}"
  [[ "${actual_entries}" == "${expected_entries}" ]] ||
    fail "unexpected archive contents for ${archive}: ${actual_entries}"
done

[[ -f "${dist_dir}/checksums.txt" ]] ||
  fail "missing distribution checksum file: ${dist_dir}/checksums.txt"

(
  cd "${dist_dir}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum --check checksums.txt
  else
    shasum -a 256 --check checksums.txt
  fi
)

actual_checksum_entries="$(awk '{print $2}' "${dist_dir}/checksums.txt" | LC_ALL=C sort)"
expected_checksum_entries="$(printf '%s\n' "${archives[@]}" | LC_ALL=C sort)"
[[ "${actual_checksum_entries}" == "${expected_checksum_entries}" ]] ||
  fail "checksums.txt does not cover exactly the supported platform archives"

verify_dir="$(mktemp -d)"
trap 'rm -rf "${verify_dir}"' EXIT
for archive in "${archives[@]}"; do
  tar -C "${verify_dir}" -xzf "${dist_dir}/${archive}"
done

darwin_binary="${verify_dir}/${binary_name}_${version}_darwin_arm64/${binary_name}"
linux_binary="${verify_dir}/${binary_name}_${version}_linux_amd64/${binary_name}"
darwin_description="$(file "${darwin_binary}")"
linux_description="$(file "${linux_binary}")"
[[ "${darwin_description}" == *"Mach-O 64-bit"* && "${darwin_description}" == *"arm64"* ]] ||
  fail "unexpected darwin/arm64 binary format: ${darwin_description}"
[[ "${linux_description}" == *"ELF 64-bit LSB executable, x86-64"* ]] ||
  fail "unexpected linux/amd64 binary format: ${linux_description}"

case "$(go env GOOS)/$(go env GOARCH)" in
  darwin/arm64)
    smoke_binary="${darwin_binary}"
    ;;
  linux/amd64)
    smoke_binary="${linux_binary}"
    ;;
  *)
    echo "artifact execution smoke checks are not supported on this host"
    exit 0
    ;;
esac

actual_version="$("${smoke_binary}" --version)"
[[ "${actual_version}" == "${version}" ]] ||
  fail "version smoke check returned ${actual_version}, expected ${version}"

generated="$("${smoke_binary}" --words 3 --separator /)"
[[ "${generated}" =~ ^[a-z]+/[a-z]+/[a-z]+$ ]] ||
  fail "generation smoke check returned unexpected output: ${generated}"
