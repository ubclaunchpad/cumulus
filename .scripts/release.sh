#!/bin/sh
# This script cross-compiles binaries for various platforms

# Download our release binary builder
go get -u github.com/mitchellh/gox

# Specify platforms and release version
PLATFORMS="linux/amd64 linux/386 darwin/386 windows/amd64 windows/386"
RELEASE=$(git describe --tags)
echo "Building release $RELEASE"

# Build Inertia Go binaries for specified platforms
gox -output="cumulus.$(git describe --tags).{{.OS}}.{{.Arch}}" \
    -osarch="$PLATFORMS" \
