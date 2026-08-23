#!/usr/bin/env bash

set -euo pipefail

: "${MOCK_GH_STATE_DIR:?MOCK_GH_STATE_DIR is required}"
: "${MOCK_GH_LOG:?MOCK_GH_LOG is required}"

release_file="${MOCK_GH_STATE_DIR}/release.json"
tag_file="${MOCK_GH_STATE_DIR}/tag.json"

fail() {
  echo "fake gh: $1" >&2
  exit 1
}

file_digest() {
  local path="$1"
  local digest

  if command -v sha256sum >/dev/null 2>&1; then
    digest="$(sha256sum "${path}" | awk '{print $1}')"
  else
    digest="$(shasum -a 256 "${path}" | awk '{print $1}')"
  fi
  printf 'sha256:%s\n' "${digest}"
}

replace_release() {
  local replacement="$1"
  mv "${replacement}" "${release_file}"
}

if [[ "${1:-}" == "api" ]]; then
  shift
  [[ "${1:-}" == "-H" && -n "${2:-}" ]] || fail "expected API version header"
  shift 2

  if [[ "${1:-}" == "--paginate" ]]; then
    [[ "${2:-}" == "--slurp" && $# -eq 3 ]] || fail "unexpected paginated API arguments"
    endpoint="$3"
    [[ "${endpoint}" == repos/*/releases\?per_page=100 ]] || fail "unexpected endpoint: ${endpoint}"
    if [[ -f "${release_file}" ]]; then
      jq -s '[.]' "${release_file}"
    else
      printf '[[]]\n'
    fi
    exit 0
  fi

  [[ $# -eq 1 ]] || fail "unexpected API arguments"
  endpoint="$1"
  case "${endpoint}" in
    repos/*/git/ref/tags/*)
      if [[ -f "${tag_file}" ]]; then
        cat "${tag_file}"
      else
        echo "gh: Not Found (HTTP 404)" >&2
        exit 1
      fi
      ;;
    repos/*/git/tags/*)
      object_sha="${endpoint##*/}"
      annotated_file="${MOCK_GH_STATE_DIR}/annotated-${object_sha}.json"
      [[ -f "${annotated_file}" ]] || fail "annotated tag not found: ${object_sha}"
      cat "${annotated_file}"
      ;;
    *)
      fail "unexpected endpoint: ${endpoint}"
      ;;
  esac
  exit 0
fi

if [[ "${1:-}" == "release" && "${2:-}" == "create" ]]; then
  [[ $# -ge 3 ]] || fail "missing release tag"
  version="$3"
  shift 3
  draft=false
  generate_notes=false
  prerelease=false
  verify_tag=false
  title=""
  notes=""
  target=""

  while (($# > 0)); do
    case "$1" in
      --draft)
        draft=true
        ;;
      --generate-notes)
        generate_notes=true
        ;;
      --prerelease)
        prerelease=true
        ;;
      --verify-tag)
        verify_tag=true
        ;;
      --title)
        shift
        title="${1:-}"
        ;;
      --notes)
        shift
        notes="${1:-}"
        ;;
      --target)
        shift
        target="${1:-}"
        ;;
      *)
        fail "unexpected release create argument: $1"
        ;;
    esac
    shift
  done

  [[ "${draft}" == "true" ]] || fail "release must be created as a draft"
  [[ "${generate_notes}" == "true" ]] || fail "generated notes are required"
  [[ -n "${title}" && -n "${notes}" && -n "${target}" ]] || fail "release metadata is incomplete"
  [[ ! -f "${release_file}" ]] || fail "release already exists"
  if [[ "${verify_tag}" == "true" ]]; then
    [[ -f "${tag_file}" ]] || fail "--verify-tag used without an existing tag"
  elif [[ ! -f "${tag_file}" ]]; then
    jq -n --arg sha "${target}" '{object: {type: "commit", sha: $sha}}' > "${tag_file}"
  fi

  jq -n \
    --arg tag "${version}" \
    --arg target "${target}" \
    --arg title "${title}" \
    --arg body "${notes}" \
    --argjson prerelease "${prerelease}" \
    '{tag_name: $tag, target_commitish: $target, name: $title, body: $body,
      draft: true, prerelease: $prerelease, assets: []}' > "${release_file}"
  printf 'create\t%s\t%s\n' "${version}" "${target}" >> "${MOCK_GH_LOG}"
  exit 0
fi

if [[ "${1:-}" == "release" && "${2:-}" == "upload" ]]; then
  [[ $# -eq 5 && "$5" == "--clobber" ]] || fail "unexpected release upload arguments"
  version="$3"
  asset_path="$4"
  asset_name="$(basename "${asset_path}")"
  [[ -f "${release_file}" ]] || fail "release does not exist"
  [[ "$(jq -r '.tag_name' "${release_file}")" == "${version}" ]] || fail "release tag mismatch"
  [[ "$(jq -r '.draft' "${release_file}")" == "true" ]] || fail "cannot upload to a published release"

  if [[ "${MOCK_GH_FAIL_UPLOAD:-}" == "${asset_name}" ]]; then
    printf 'upload-failed\t%s\n' "${asset_name}" >> "${MOCK_GH_LOG}"
    exit 1
  fi

  digest="$(file_digest "${asset_path}")"
  replacement="$(mktemp "${MOCK_GH_STATE_DIR}/release.XXXXXX")"
  jq --arg name "${asset_name}" --arg digest "${digest}" \
    '.assets = ([.assets[] | select(.name != $name)] +
      [{name: $name, state: "uploaded", digest: $digest}])' \
    "${release_file}" > "${replacement}"
  replace_release "${replacement}"
  printf 'upload\t%s\n' "${asset_name}" >> "${MOCK_GH_LOG}"
  exit 0
fi

if [[ "${1:-}" == "release" && "${2:-}" == "edit" ]]; then
  [[ $# -eq 4 && "$4" == "--draft=false" ]] || fail "unexpected release edit arguments"
  version="$3"
  [[ -f "${release_file}" ]] || fail "release does not exist"
  [[ "$(jq -r '.tag_name' "${release_file}")" == "${version}" ]] || fail "release tag mismatch"
  [[ "$(jq -r '.draft' "${release_file}")" == "true" ]] || fail "release is already published"

  replacement="$(mktemp "${MOCK_GH_STATE_DIR}/release.XXXXXX")"
  jq '.draft = false' "${release_file}" > "${replacement}"
  replace_release "${replacement}"
  printf 'publish\t%s\n' "${version}" >> "${MOCK_GH_LOG}"
  exit 0
fi

fail "unexpected command: $*"
