#!/bin/bash
set -e

VERSION=${1:-"1.0.0-rc.1"}
BUILD_DIR=".release/rc"
ARTIFACTS_DIR="$BUILD_DIR/artifacts"

echo "Building Libcode $VERSION"

# Clean and create directories
rm -rf "$BUILD_DIR"
mkdir -p "$ARTIFACTS_DIR"

# Build for all platforms
echo "Building for all platforms..."
mkdir -p bin

GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$VERSION" -o bin/libcode-darwin-arm64 ./cmd/libcode
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o bin/libcode-darwin-amd64 ./cmd/libcode
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o bin/libcode-linux-amd64 ./cmd/libcode
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o bin/libcode-windows-amd64.exe ./cmd/libcode

# Copy artifacts
cp bin/libcode-* "$ARTIFACTS_DIR/"

# Create checksums
cd "$ARTIFACTS_DIR"
shasum * > SHA256SUMS 2>/dev/null || sha256sum * > SHA256SUMS
cd -

echo ""
echo "Build complete!"
echo "Version: $VERSION"
echo "Artifacts: $ARTIFACTS_DIR"
ls -lh "$ARTIFACTS_DIR"
