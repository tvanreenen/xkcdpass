#!/usr/bin/env bash

# Keep source-side maintainer commands on the same strict v-prefixed SemVer grammar
# used by the distribution and tap workflows.
xkcdpass_release_semver='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*))?(\+([0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*))?$'

validate_release_version() {
  local version="${1:-}"
  [[ "${version}" =~ ${xkcdpass_release_semver} ]]
}
