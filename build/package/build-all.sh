#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:?VERSION must be set, for example VERSION=1.0.0}"
repository="${REPOSITORY:?REPOSITORY must be set, for example REPOSITORY=OWNER/REPOSITORY}"
name="${NAME:?NAME must be set, for example NAME=edgeworkers}"
command_path="${COMMAND_PATH:?COMMAND_PATH must be set, for example COMMAND_PATH=./cmd/edgeworkers}"
dist="${DIST:-dist/${name}}"
release_url="${RELEASE_URL:-https://github.com/${repository}/releases/download/${name}-${version}}"
packages=()

case "$name" in
  edgeworkers)
    description="Manage Akamai EdgeWorkers."
    ;;
  edgekv)
    description="Manage Akamai EdgeKV."
    ;;
  *) description="Manage Akamai ${name}." ;;
esac
ldflags="-X main.version=${version}"

if [[ ! "$version" =~ ^[0-9]+(\.[0-9]+)*$ ]]; then
  printf 'VERSION must be numeric, for example 1.0.0\n' >&2
  exit 1
fi

rm -rf "$dist"
mkdir -p "$dist/.spin"

build_target() {
  local goos=$1
  local goarch=$2
  local akamai_os=$3
  local spin_os=$4
  local spin_arch=$5
  local suffix=$6
  local akamai_binary="akamai-${name}-${akamai_os}-${goarch}${suffix}"
  local spin_binary="${name}${suffix}"
  local archive="${name}-${version}-${spin_os}-${spin_arch}.tar.gz"
  local package_dir="$dist/.spin/${spin_os}-${spin_arch}"
  local checksum

  printf 'Building %s\n' "$akamai_binary"
  GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags "$ldflags" -o "$dist/$akamai_binary" "$command_path"

  mkdir -p "$package_dir"
  cp "$dist/$akamai_binary" "$package_dir/$spin_binary"
  tar -C "$package_dir" -czf "$dist/$archive" "$spin_binary"

  checksum="$(shasum -a 256 "$dist/$archive")"
  checksum="${checksum%% *}"
  packages+=("$spin_os|$spin_arch|$archive|$checksum")
}

build_target linux amd64 linux linux amd64 ""
build_target linux arm64 linux linux aarch64 ""
build_target darwin amd64 mac macos amd64 ""
build_target darwin arm64 mac macos aarch64 ""
build_target windows amd64 windows windows amd64 .exe
build_target windows arm64 windows windows aarch64 .exe

{
  printf '{\n'
  printf '  "name": "%s",\n' "$name"
  printf '  "description": "%s",\n' "$description"
  printf '  "homepage": "https://github.com/%s",\n' "$repository"
  printf '  "version": "%s",\n' "$version"
  printf '  "spinCompatibility": ">=1.0",\n'
  printf '  "license": "Apache-2.0",\n'
  printf '  "packages": [\n'

  for index in "${!packages[@]}"; do
    IFS='|' read -r spin_os spin_arch archive checksum <<<"${packages[$index]}"
    printf '    {\n'
    printf '      "os": "%s",\n' "$spin_os"
    printf '      "arch": "%s",\n' "$spin_arch"
    printf '      "url": "%s/%s",\n' "$release_url" "$archive"
    printf '      "sha256": "%s"\n' "$checksum"
    if ((index + 1 < ${#packages[@]})); then
      printf '    },\n'
    else
      printf '    }\n'
    fi
  done

  printf '  ]\n'
  printf '}\n'
} >"$dist/${name}.json"

rm -rf "$dist/.spin"
