#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:?VERSION must be set, for example VERSION=0.1.0}"
repository="${REPOSITORY:?REPOSITORY must be set, for example REPOSITORY=OWNER/REPOSITORY}"
dist="${DIST:-dist/devserver}"
release_url="${RELEASE_URL:-https://github.com/${repository}/releases/download/devserver-${version}}"
release_date="${RELEASE_DATE:-$(date -u +%Y-%m-%d)}"
manifest="${MANIFEST:-devserver-manifest.json}"

if [[ ! "$version" =~ ^[0-9]+(\.[0-9]+)*$ ]]; then
  printf 'VERSION must be numeric, for example 0.1.0\n' >&2
  exit 1
fi

rm -rf "$dist"
mkdir -p "$dist"

packages_json="$dist/packages.json"
printf '[\n' >"$packages_json"
first=true

build_target() {
  local goos=$1
  local goarch=$2
  local suffix=$3
  local binary="edgeworkers-devserver-${goos}-${goarch}${suffix}"
  local checksum

  printf 'Building %s\n' "$binary"
  GOOS="$goos" GOARCH="$goarch" go build -ldflags "-X main.version=${version}" -o "$dist/$binary" ./cmd/edgeworkersdevserver

  checksum="$(shasum -a 256 "$dist/$binary")"
  checksum="${checksum%% *}"

  if [[ "$first" == true ]]; then
    first=false
  else
    printf ',\n' >>"$packages_json"
  fi
  printf '  {\n    "os": "%s",\n    "arch": "%s",\n    "url": "%s/%s",\n    "sha256": "%s"\n  }' \
    "$goos" "$goarch" "$release_url" "$binary" "$checksum" >>"$packages_json"
}

build_target darwin amd64 ""
build_target darwin arm64 ""
build_target linux amd64 ""
build_target linux arm64 ""
build_target windows amd64 .exe
build_target windows arm64 .exe

printf '\n]\n' >>"$packages_json"

go run ./build/package/rotate-manifest \
  --manifest "$manifest" \
  --version "$version" \
  --release-date "$release_date" \
  --packages-file "$packages_json"
