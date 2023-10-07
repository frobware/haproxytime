#!/usr/bin/env bash

set -e

new_version=""
commit_flag=0
release_flag=0
create_branch=0

while [ "$1" != "" ]; do
    case $1 in
        --version) shift; new_version=$1 ;;
        --commit) commit_flag=1 ;;
        --release) release_flag=1 ;;
        --branch) create_branch=1 ;;
        *) echo "Unknown flag $1"; exit 1 ;;
    esac
    shift
done

if [ -n "$new_version" ]; then
    version=$new_version
else
    version=$(git describe --tags --match 'v*' --abbrev=8 --always --long --dirty | cut -c 2-)
fi

if [ $release_flag -eq 1 ]; then
    date=$(date '+%Y-%m-%d')
else
    date=$(date --iso-8601=seconds)
fi

goversion=$(go version)

if [ $create_branch -eq 1 ]; then
    git checkout -b "release-${version}"
fi

mkdir -p build

cat <<EOF > build/build.go
package build

// Automatically generated. DO NOT EDIT.

var (
	Version   = "${version}"
	Date      = "${date}"
	GoVersion = "${goversion}"
)
EOF

gofmt -w ./build/build.go

cat <<EOF > version.nix
# Automatically generated. DO NOT EDIT.
{ version = "${version}"; date = "${date}"; goversion = "${goversion}"; }
EOF

if [ $commit_flag -eq 1 ]; then
    git add build/build.go version.nix
    git commit -m "Update build.go and version.nix for version ${version}"
    git tag "v${version}"
fi
