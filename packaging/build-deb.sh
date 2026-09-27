#!/usr/bin/env bash
# ==============================================================================
# Script to build Debian/Ubuntu (.deb) package for agef
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${SCRIPT_DIR}/dist"
BUILD_DIR="${SCRIPT_DIR}/build/deb"
VERSION="1.0.0"

echo "Building agef binary..."
mkdir -p "${SCRIPT_DIR}/bin"
CGO_ENABLED=0 go build -ldflags="-s -w" -o "${SCRIPT_DIR}/bin/agef" ./cmd/agef

echo "Setting up DEB build environment..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"/{data/usr/bin,control}
mkdir -p "$DIST_DIR"

# 1. debian-binary
echo "2.0" > "$BUILD_DIR/debian-binary"

# 2. control file
cat << EOF > "$BUILD_DIR/control/control"
Package: agef
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Open Source Contributor <contributor@example.com>
Description: High-performance recursive folder encryption and decryption using age
 agef is a fast, standalone folder encryption and decryption tool
 powered by the age encryption library. It allows mass encryption and decryption
 of entire folder trees while preserving directory hierarchy and timestamps,
 creating in-place encrypted and decrypted folders.
EOF

# 3. Payload
install -m 0755 "${SCRIPT_DIR}/bin/agef" "$BUILD_DIR/data/usr/bin/agef"

# 4. Tar control and data
(cd "$BUILD_DIR/control" && tar --numeric-owner --group=0 --owner=0 -czf "$BUILD_DIR/control.tar.gz" .)
(cd "$BUILD_DIR/data" && tar --numeric-owner --group=0 --owner=0 -czf "$BUILD_DIR/data.tar.gz" .)

# 5. Pack deb
DEB_NAME="agef_${VERSION}_amd64.deb"
(cd "$BUILD_DIR" && ar rcs "${DIST_DIR}/${DEB_NAME}" debian-binary control.tar.gz data.tar.gz)

# Symlink generic release name
cp "${DIST_DIR}/${DEB_NAME}" "${DIST_DIR}/agef.deb"

echo "DEB built successfully:"
ls -lh "${DIST_DIR}"/*.deb
