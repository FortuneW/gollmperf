#!/bin/bash

# Build script for gollmperf
# This script builds the binary and optionally packages it with config and example files

set -e  # Exit on any error

# Check if we should package
PACK=false
if [ "$1" = "pack" ]; then
    PACK=true
fi

# Get git tag/description
GIT_TAG=$(git describe --tags 2>/dev/null || echo "untagged")

# Get the architecture
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Convert architecture names to common formats
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="armv7"
        ;;
esac

# Convert OS names to common formats
case $OS in
    linux)
        OS="linux"
        ;;
    darwin)
        OS="darwin"
        ;;
    mingw*|cygwin*|msys*)
        OS="windows"
        ;;
esac

# Package name
PKG_NAME="gollmperf-${GIT_TAG}-${OS}-${ARCH}.tar.gz"

# Create releases directory if it doesn't exist
mkdir -p releases

# Build the binary without CGO
echo "Building gollmperf for ${OS}/${ARCH}..."
CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -o gollmperf .

# If not packaging, exit here
if [ "$PACK" = false ]; then
    echo "Build completed successfully! Binary created: gollmperf"
    exit 0
fi

# Package name

# Create a temporary directory for packaging
TMP_DIR="releases/tmp"
rm -rf $TMP_DIR
mkdir -p $TMP_DIR/gollmperf

# Copy binary
echo "Copying binary..."
cp gollmperf $TMP_DIR/gollmperf/

# Copy config files
echo "Copying config files..."
mkdir -p $TMP_DIR/gollmperf/configs
cp configs/example.yaml $TMP_DIR/gollmperf/configs/

# Copy example files
echo "Copying example files..."
mkdir -p $TMP_DIR/gollmperf/examples
cp examples/test_cases.jsonl $TMP_DIR/gollmperf/examples/

# Copy README files if they exist
if [ -f README.md ]; then
    echo "Copying README files..."
    cp README.md $TMP_DIR/gollmperf/
    mkdir -p $TMP_DIR/gollmperf/docs/assets/logos
    cp ./docs/assets/logos/logo1.png $TMP_DIR/gollmperf/docs/assets/logos/
    mkdir -p $TMP_DIR/gollmperf/docs/assets/reporter_case
    cp ./docs/assets/reporter_case/reporter_case.png $TMP_DIR/gollmperf/docs/assets/reporter_case/
fi
if [ -f README_zh.md ]; then
    cp README_zh.md $TMP_DIR/gollmperf/
fi

# Copy and update LICENSE if it exists
if [ -f LICENSE ]; then
    echo "Copying and updating LICENSE..."
    CURRENT_YEAR=$(date +%Y)
    sed "s/Copyright \[2025\]/Copyright ${CURRENT_YEAR}/g" LICENSE > $TMP_DIR/gollmperf/LICENSE
fi

# Create the package
echo "Creating package ${PKG_NAME}..."
cd $TMP_DIR
tar -czf ../${PKG_NAME} gollmperf
cd ../..

# Clean up temporary directory
rm -rf $TMP_DIR

# Show package information
echo "Package created: releases/${PKG_NAME}"
echo "Package contents:"
tar -tzf releases/${PKG_NAME}

echo "Build completed successfully!"