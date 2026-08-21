#!/usr/bin/env bash

set -euo pipefail

fail() {
  echo "$1" >&2
  exit 1
}

if [[ $# -ne 3 ]]; then
  fail "usage: $0 <version> <target-commit> <artifact-directory>"
fi

version="$1"
target_sha="$2"
dist_dir="$3"
repository="${GITHUB_REPOSITORY:-}"
api_header="X-GitHub-Api-Version: 2026-03-10"
release_note_marker='<!-- xkcdpass distribution workflow -->'
semver='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*))?(\+([0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*))?$'

[[ "${version}" =~ ${semver} ]] ||
  fail "release version must be a valid v-prefixed Semantic Version"
[[ "${target_sha}" =~ ^[0-9a-f]{40}$ ]] || fail "invalid target commit: ${target_sha}"
[[ "${repository}" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] ||
  fail "invalid GITHUB_REPOSITORY: ${repository}"
[[ -d "${dist_dir}" ]] || fail "artifact directory not found: ${dist_dir}"
command -v gh >/dev/null 2>&1 || fail "gh CLI is required"
command -v jq >/dev/null 2>&1 || fail "jq is required"

expected_assets=(
  "xkcdpass_${version}_darwin_arm64.tar.gz"
  "xkcdpass_${version}_linux_amd64.tar.gz"
  "checksums.txt"
)

unsorted_actual_assets=()
shopt -s nullglob dotglob
for path in "${dist_dir}"/*; do
  [[ -f "${path}" && ! -L "${path}" ]] ||
    fail "artifact directory contains a non-file entry: ${path}"
  unsorted_actual_assets+=("$(basename "${path}")")
done
shopt -u nullglob dotglob
actual_assets=()
while IFS= read -r asset; do
  actual_assets+=("${asset}")
done < <(printf '%s\n' "${unsorted_actual_assets[@]}" | LC_ALL=C sort)
sorted_expected_assets=()
while IFS= read -r asset; do
  sorted_expected_assets+=("${asset}")
done < <(printf '%s\n' "${expected_assets[@]}" | LC_ALL=C sort)
[[ "${actual_assets[*]}" == "${sorted_expected_assets[*]}" ]] ||
  fail "artifact directory does not contain exactly the expected release assets"

(
  cd "${dist_dir}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum --check --strict checksums.txt
  else
    shasum -a 256 --check checksums.txt
  fi
)

checksum_assets=()
while IFS= read -r asset; do
  checksum_assets+=("${asset}")
done < <(awk '{print $2}' "${dist_dir}/checksums.txt" | LC_ALL=C sort)
expected_archives=("${expected_assets[0]}" "${expected_assets[1]}")
sorted_expected_archives=()
while IFS= read -r asset; do
  sorted_expected_archives+=("${asset}")
done < <(printf '%s\n' "${expected_archives[@]}" | LC_ALL=C sort)
[[ "${checksum_assets[*]}" == "${sorted_expected_archives[*]}" ]] ||
  fail "checksums.txt does not cover exactly the expected release archives"

asset_digest() {
  local path="$1"
  local digest

  if command -v sha256sum >/dev/null 2>&1; then
    digest="$(sha256sum "${path}" | awk '{print $1}')"
  else
    digest="$(shasum -a 256 "${path}" | awk '{print $1}')"
  fi
  printf 'sha256:%s\n' "${digest}"
}

runner_temp="${RUNNER_TEMP:-/tmp}"
release_json="${runner_temp}/xkcdpass-release.json"
releases_json="${runner_temp}/xkcdpass-releases.json"
tag_json="${runner_temp}/xkcdpass-tag.json"
api_error="${runner_temp}/xkcdpass-api-error.txt"
release_exists=false
tag_exists=false

fetch_optional() {
  local endpoint="$1"
  local output="$2"

  if gh api -H "${api_header}" "${endpoint}" > "${output}" 2> "${api_error}"; then
    return 0
  fi
  if grep -q '(HTTP 404)' "${api_error}"; then
    : > "${output}"
    return 4
  fi
  cat "${api_error}" >&2
  return 1
}

fetch_release() {
  gh api -H "${api_header}" --paginate --slurp \
    "repos/${repository}/releases?per_page=100" > "${releases_json}"
  jq --arg version "${version}" \
    '[.[][] | select(.tag_name == $version)]' "${releases_json}" > "${release_json}"

  local release_count
  release_count="$(jq 'length' "${release_json}")"
  if [[ "${release_count}" == "1" ]]; then
    jq '.[0]' "${release_json}" > "${release_json}.single"
    mv "${release_json}.single" "${release_json}"
    release_exists=true
  elif [[ "${release_count}" == "0" ]]; then
    release_exists=false
  else
    fail "multiple releases claim version ${version}"
  fi
}

fetch_tag() {
  if fetch_optional "repos/${repository}/git/ref/tags/${version}" "${tag_json}"; then
    tag_exists=true
  else
    local status=$?
    [[ ${status} -eq 4 ]] || return "${status}"
    tag_exists=false
  fi
}

resolve_tag_commit() {
  local object_sha object_type annotated_json depth

  object_sha="$(jq -er '.object.sha' "${tag_json}")"
  object_type="$(jq -er '.object.type' "${tag_json}")"
  depth=0
  while [[ "${object_type}" == "tag" ]]; do
    ((depth += 1))
    [[ ${depth} -le 5 ]] || fail "tag ${version} has too many annotation levels"
    annotated_json="${runner_temp}/xkcdpass-annotated-tag-${depth}.json"
    gh api -H "${api_header}" "repos/${repository}/git/tags/${object_sha}" > "${annotated_json}"
    object_sha="$(jq -er '.object.sha' "${annotated_json}")"
    object_type="$(jq -er '.object.type' "${annotated_json}")"
  done

  [[ "${object_type}" == "commit" ]] || fail "tag ${version} does not resolve to a commit"
  printf '%s\n' "${object_sha}"
}

is_prerelease=false
version_without_build="${version%%+*}"
if [[ "${version_without_build}" == *-* ]]; then
  is_prerelease=true
fi

verify_draft_metadata() {
  [[ "$(jq -r '.draft' "${release_json}")" == "true" ]] ||
    fail "release ${version} is already published; published releases are immutable"
  [[ "$(jq -r '.tag_name' "${release_json}")" == "${version}" ]] ||
    fail "draft release tag does not match ${version}"
  [[ "$(jq -r '.target_commitish' "${release_json}")" == "${target_sha}" ]] ||
    fail "draft release ${version} targets a different commit"
  [[ "$(jq -r '.name' "${release_json}")" == "${version}" ]] ||
    fail "draft release ${version} has a mismatched title"
  [[ "$(jq -r '.body' "${release_json}")" == "${release_note_marker}"* ]] ||
    fail "draft release ${version} was not created with the expected generated notes"
  [[ "$(jq -r '.prerelease' "${release_json}")" == "${is_prerelease}" ]] ||
    fail "draft release ${version} has a mismatched prerelease setting"
}

verify_no_unexpected_remote_assets() {
  local remote_asset expected found

  while IFS= read -r remote_asset; do
    found=false
    for expected in "${expected_assets[@]}"; do
      if [[ "${remote_asset}" == "${expected}" ]]; then
        found=true
        break
      fi
    done
    [[ "${found}" == "true" ]] ||
      fail "draft release ${version} contains unexpected asset: ${remote_asset}"
  done < <(jq -r '.assets[].name' "${release_json}")
}

verify_complete_remote_assets() {
  local asset state digest local_digest

  remote_assets=()
  while IFS= read -r asset; do
    remote_assets+=("${asset}")
  done < <(jq -r '.assets[].name' "${release_json}" | LC_ALL=C sort)
  [[ "${remote_assets[*]}" == "${sorted_expected_assets[*]}" ]] ||
    fail "draft release ${version} does not contain exactly the expected assets"

  for asset in "${expected_assets[@]}"; do
    state="$(jq -r --arg name "${asset}" '.assets[] | select(.name == $name) | .state' "${release_json}")"
    digest="$(jq -r --arg name "${asset}" '.assets[] | select(.name == $name) | .digest // ""' "${release_json}")"
    local_digest="$(asset_digest "${dist_dir}/${asset}")"
    [[ "${state}" == "uploaded" ]] || fail "release asset ${asset} is not fully uploaded"
    [[ "${digest}" == "${local_digest}" ]] ||
      fail "release asset ${asset} does not match this workflow run"
  done
}

# Complete every local and remote preflight check before creating or changing release state.
fetch_release
fetch_tag

if [[ "${tag_exists}" == "true" ]]; then
  [[ "$(resolve_tag_commit)" == "${target_sha}" ]] ||
    fail "tag ${version} targets a different commit"
fi

if [[ "${release_exists}" == "true" ]]; then
  verify_draft_metadata
  verify_no_unexpected_remote_assets
else
  create_args=(release create "${version}" --draft --title "${version}" --notes "${release_note_marker}" --generate-notes --target "${target_sha}")
  if [[ "${tag_exists}" == "true" ]]; then
    create_args+=(--verify-tag)
  fi
  if [[ "${is_prerelease}" == "true" ]]; then
    create_args+=(--prerelease)
  fi
  gh "${create_args[@]}"

  fetch_release
  fetch_tag
  [[ "${release_exists}" == "true" ]] || fail "GitHub did not create the expected draft release"
  verify_draft_metadata
  if [[ "${tag_exists}" == "true" ]]; then
    [[ "$(resolve_tag_commit)" == "${target_sha}" ]] ||
      fail "new tag ${version} does not target ${target_sha}"
  fi
fi

for asset in "${expected_assets[@]}"; do
  local_digest="$(asset_digest "${dist_dir}/${asset}")"
  remote_digest="$(jq -r --arg name "${asset}" '.assets[] | select(.name == $name) | .digest // ""' "${release_json}")"
  if [[ "${remote_digest}" == "${local_digest}" ]]; then
    continue
  fi
  gh release upload "${version}" "${dist_dir}/${asset}" --clobber
done

# Publication is the only irreversible transition. Recheck all invariants immediately before it.
fetch_release
fetch_tag
verify_draft_metadata
if [[ "${tag_exists}" == "true" ]]; then
  [[ "$(resolve_tag_commit)" == "${target_sha}" ]] ||
    fail "tag ${version} changed before publication"
fi
verify_complete_remote_assets

gh release edit "${version}" --draft=false

fetch_release
fetch_tag
[[ "$(jq -r '.draft' "${release_json}")" == "false" ]] ||
  fail "release ${version} was not published"
[[ "${tag_exists}" == "true" ]] || fail "published release ${version} has no tag"
[[ "$(resolve_tag_commit)" == "${target_sha}" ]] ||
  fail "published tag ${version} does not target ${target_sha}"
verify_complete_remote_assets
