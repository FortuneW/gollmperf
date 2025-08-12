#!/bin/bash

# Build script for gollmperf
# This script builds the binary and optionally packages it with config and example files

set -e  # Exit on any error

# Default values
PACK=false
RELEASE=false

# Parse command line arguments
case "$1" in
    pack)
        PACK=true
        ;;
    release)
        RELEASE=true
        ;;
    *)
        # Default behavior - just build
        ;;
esac

# Get git tag/description
GIT_TAG=$(git describe --tags 2>/dev/null || echo "untagged")

# Function to copy assets to package directory
copy_assets() {
    local target_dir=$1
    
    # Copy config files
    echo "Copying config files..."
    mkdir -p $target_dir/configs
    cp configs/example.yaml $target_dir/configs/
    
    # Copy example files
    echo "Copying example files..."
    mkdir -p $target_dir/examples
    cp examples/test_cases.jsonl $target_dir/examples/
    
    # Copy README files if they exist
    if [ -f README.md ]; then
        echo "Copying README files..."
        cp README.md $target_dir/
        mkdir -p $target_dir/docs/assets/logos
        cp ./docs/assets/logos/logo1.png $target_dir/docs/assets/logos/
        mkdir -p $target_dir/docs/assets/reporter_case
        cp ./docs/assets/reporter_case/reporter_case.png $target_dir/docs/assets/reporter_case/
    fi
    if [ -f README_zh.md ]; then
        cp README_zh.md $target_dir/
    fi
    
    # Copy and update LICENSE if it exists
    if [ -f LICENSE ]; then
        echo "Copying and updating LICENSE..."
        CURRENT_YEAR=$(date +%Y)
        sed "s/Copyright \[2025\]/Copyright ${CURRENT_YEAR}/g" LICENSE > $target_dir/LICENSE
    fi
}

# Function to create package
create_package() {
    local pkg_dir=$1
    local pkg_name=$2
    local pkg_os=$3
    
    echo "Creating package ${pkg_name}..."
    cd $pkg_dir
    if [ "$pkg_os" = "windows" ]; then
        zip -r "../${pkg_name}" gollmperf
    else
        tar -czf "../${pkg_name}" gollmperf
    fi
    cd ../..
}

# Function to show package contents
show_package_contents() {
    local pkg_name=$1
    local pkg_os=$2
    
    echo "Package contents:"
    if [ "$pkg_os" = "windows" ]; then
        unzip -l releases/${pkg_name}
    else
        tar -tzf releases/${pkg_name}
    fi
}

# Function to build and package for a specific platform
build_and_package() {
    local target_os=$1
    local target_arch=$2
    local is_release=$3
    
    echo "Building gollmperf for ${target_os}/${target_arch}..."
    
    # Set binary name
    binary_name="gollmperf"
    if [ "$target_os" = "windows" ]; then
        binary_name="gollmperf.exe"
    fi
    
    go mod tidy
    # Build the binary without CGO
    if [ "$is_release" = "true" ]; then
        CGO_ENABLED=0 GOOS=$target_os GOARCH=$target_arch go build -o "releases/$binary_name" .
    else
        CGO_ENABLED=0 GOOS=$target_os GOARCH=$target_arch go build -o gollmperf .
    fi
    
    # If not packaging, exit here (only for non-release builds)
    if [ "$PACK" = false ] && [ "$RELEASE" = false ] && [ "$is_release" = "false" ]; then
        return 0
    fi
    
    # Create a temporary directory for packaging
    TMP_DIR="releases/tmp"
    rm -rf $TMP_DIR
    mkdir -p $TMP_DIR/gollmperf
    
    # Copy binary
    echo "Copying binary..."
    if [ "$is_release" = "true" ]; then
        cp "releases/$binary_name" $TMP_DIR/gollmperf/
    else
        cp gollmperf $TMP_DIR/gollmperf/
    fi
    
    # Copy assets
    copy_assets $TMP_DIR/gollmperf
    
    # Create the package
    PKG_NAME="gollmperf-${GIT_TAG}-${target_os}-${target_arch}"
    if [ "$target_os" = "windows" ]; then
        PKG_NAME="${PKG_NAME}.zip"
    else
        PKG_NAME="${PKG_NAME}.tar.gz"
    fi
    
    create_package $TMP_DIR "$PKG_NAME" "$target_os"
    
    # Clean up temporary directory and binary (for release builds)
    rm -rf $TMP_DIR
    if [ "$is_release" = "true" ]; then
        rm -f "releases/$binary_name"
    fi
    
    # Show package information
    echo "Package created: releases/${PKG_NAME}"
    show_package_contents "$PKG_NAME" "$target_os"
}

# If building for release, build for all platforms
if [ "$RELEASE" = true ]; then
    # Define platforms to build for release
    platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")
    
    # Create releases directory if it doesn't exist
    mkdir -p releases
    
    for platform in "${platforms[@]}"; do
        IFS="/" read -r -a parts <<< "$platform"
        os="${parts[0]}"
        arch="${parts[1]}"
        
        build_and_package "$os" "$arch" "true"
    done
    
    echo "All release packages created successfully!"
    exit 0
fi

# Get the architecture and OS, allowing override from environment variables
ARCH=${ARCH:-$(uname -m)}
OS=${OS:-$(uname -s | tr '[:upper:]' '[:lower:]')}

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
    windows|mingw*|cygwin*|msys*)
        OS="windows"
        ;;
esac

# Build for single platform
build_and_package "$OS" "$ARCH" "false"

echo "Build completed successfully! Binary created: gollmperf"