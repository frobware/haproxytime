#!/usr/bin/env bash

set -eu

version_info=$(nix-instantiate --eval --json -E '(import ./version.nix)')
version=$(echo "$version_info" | jq -r '.version')
date=$(echo "$version_info" | jq -r '.date')
goversion=$(echo "$version_info" | jq -r '.goversion')

# This is just for visual verification.
echo "Version: $version"
echo "Date: $date"
echo "Go Version: $goversion"

git tag -a "v${version}" -m "Release v${version}"
