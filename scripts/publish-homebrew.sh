#!/usr/bin/env bash

set -euo pipefail

fail() {
  echo "$1" >&2
  exit 1
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=release-version.sh
source "${script_dir}/release-version.sh"

if [[ $# -ne 1 ]]; then
  fail "usage: $0 <v-prefixed-version>"
fi

version="$1"
validate_release_version "${version}" ||
  fail "version must be a valid v-prefixed Semantic Version"
command -v gh >/dev/null 2>&1 || fail "gh CLI is required"

release_details=""
if ! release_details="$(
  gh release view "${version}" \
    --repo tvanreenen/xkcdpass \
    --json isDraft,tagName,url \
    --template '{{.tagName}}|{{.isDraft}}|{{.url}}{{"\n"}}'
)"; then
  fail "could not confirm published xkcdpass release ${version}"
fi

IFS='|' read -r release_tag release_is_draft release_url <<< "${release_details}"
[[ "${release_tag}" == "${version}" ]] ||
  fail "GitHub returned release ${release_tag}, expected ${version}"
[[ "${release_is_draft}" == "false" ]] ||
  fail "xkcdpass release ${version} is still a draft"

echo "Confirmed published source release: ${release_url}"
gh workflow run publish-package.yml \
  --repo tvanreenen/homebrew-tap \
  --ref main \
  -f package=xkcdpass \
  -f version="${version}"

echo "Dispatched the Homebrew tap update for xkcdpass ${version}."
echo "Follow it with: gh run list --repo tvanreenen/homebrew-tap --workflow publish-package.yml"
