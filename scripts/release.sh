#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist_dir="${repo_root}/dist"
build_dir="${dist_dir}/build"
binary_name="xkcdpass"
tap_repo="${HOME}/Code/homebrew-tap"
tap_formula_relpath="Formula/xkcdpass.rb"
tap_formula_path="${tap_repo}/${tap_formula_relpath}"
tap_target_os="darwin"
tap_target_arch="arm64"
tap_tarball=""

usage() {
  cat <<EOF
usage: $0 <build|publish|tap|all> <version>

commands:
  build    Build release archives and checksums.
  publish  Publish GitHub release from existing dist artifacts.
  tap      Update and push Homebrew tap formula.
  all      Run build, publish, and tap in sequence.
EOF
}

validate_version() {
  local version="$1"
  if [[ "${version}" != v* ]]; then
    echo "version must start with v, for example v0.1.0" >&2
    exit 1
  fi
}

ensure_tarball_name() {
  local version="$1"
  tap_tarball="${binary_name}_${version}_${tap_target_os}_${tap_target_arch}.tar.gz"
}

build_release() {
  local version="$1"

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
}

publish_release() {
  local version="$1"
  ensure_tarball_name "${version}"

  if ! command -v gh >/dev/null 2>&1; then
    echo "gh CLI is required to create a GitHub release" >&2
    exit 1
  fi

  if [[ ! -f "${dist_dir}/${binary_name}_${version}_darwin_arm64.tar.gz" ]] || [[ ! -f "${dist_dir}/${binary_name}_${version}_linux_amd64.tar.gz" ]] || [[ ! -f "${dist_dir}/checksums.txt" ]]; then
    echo "missing release artifacts in ${dist_dir}; run '$0 build ${version}' first" >&2
    exit 1
  fi

  if gh release view "${version}" >/dev/null 2>&1; then
    echo "GitHub release ${version} already exists; skipping publish step"
    return 0
  fi

  echo "creating GitHub release ${version}"
  gh release create "${version}" \
    "${dist_dir}/${binary_name}_${version}_"*.tar.gz \
    "${dist_dir}/checksums.txt" \
    --title "${version}" \
    --generate-notes
}

update_tap() {
  local version="$1"
  local origin_url owner_repo tap_tarball_path tap_sha256 tap_url
  ensure_tarball_name "${version}"

  tap_tarball_path="${dist_dir}/${tap_tarball}"
  if [[ ! -f "${tap_tarball_path}" ]]; then
    echo "expected tap target tarball not found: ${tap_tarball_path}" >&2
    echo "run '$0 build ${version}' first" >&2
    exit 1
  fi

  tap_sha256="$(shasum -a 256 "${tap_tarball_path}" | awk '{print $1}')"
  origin_url="$(git -C "${repo_root}" config --get remote.origin.url || true)"
  if [[ -z "${origin_url}" ]]; then
    echo "remote.origin.url is not set in ${repo_root}; cannot derive GitHub release URL for tap update" >&2
    exit 1
  fi

  owner_repo="$(printf '%s\n' "${origin_url}" | sed -E 's#^git@github\.com:##; s#^https://github\.com/##; s#\.git$##')"
  if [[ "${owner_repo}" != */* ]]; then
    echo "could not parse owner/repo from origin URL: ${origin_url}" >&2
    exit 1
  fi

  if [[ ! -d "${tap_repo}" ]]; then
    echo "tap repo not found at ${tap_repo}" >&2
    exit 1
  fi

  if [[ ! -f "${tap_formula_path}" ]]; then
    echo "formula not found at ${tap_formula_path}" >&2
    echo "create the formula first, then rerun release" >&2
    exit 1
  fi

  tap_url="https://github.com/${owner_repo}/releases/download/${version}/${tap_tarball}"
  sed -E -i '' "s#^([[:space:]]*url[[:space:]]+\").*(\")#\1${tap_url}\2#" "${tap_formula_path}"
  sed -E -i '' "s#^([[:space:]]*sha256[[:space:]]+\").*(\")#\1${tap_sha256}\2#" "${tap_formula_path}"

  git -C "${tap_repo}" add "${tap_formula_relpath}"
  if git -C "${tap_repo}" diff --cached --quiet; then
    echo "tap formula unchanged; skipping tap commit and push"
  else
    git -C "${tap_repo}" commit -m "${binary_name} ${version}"
    git -C "${tap_repo}" push
    echo "updated tap formula in ${tap_repo}/${tap_formula_relpath}"
  fi
}

if [[ $# -ne 2 ]]; then
  usage >&2
  exit 1
fi

command="$1"
version="$2"
validate_version "${version}"

case "${command}" in
  build)
    build_release "${version}"
    ;;
  publish)
    publish_release "${version}"
    ;;
  tap)
    update_tap "${version}"
    ;;
  all)
    build_release "${version}"
    publish_release "${version}"
    update_tap "${version}"
    ;;
  *)
    usage >&2
    exit 1
    ;;
esac
