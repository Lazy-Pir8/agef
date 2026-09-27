#!/usr/bin/env bash
# ==============================================================================
# Script to build RPM package for agef
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${SCRIPT_DIR}/dist"
BUILD_DIR="${SCRIPT_DIR}/build/rpm"

echo "Building agef binary..."
mkdir -p "${SCRIPT_DIR}/bin"
CGO_ENABLED=0 go build -ldflags="-s -w" -o "${SCRIPT_DIR}/bin/agef" ./cmd/agef

echo "Setting up RPM build environment..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"/{BUILD,RPMS,SOURCES,SPECS,SRPMS,BUILDROOT}
mkdir -p "$DIST_DIR"

# Copy binary to SOURCES directory where spec expects it
cp "${SCRIPT_DIR}/bin/agef" "${BUILD_DIR}/SOURCES/agef"

# Build RPM
echo "Building RPM package..."
rpmbuild -bb \
  --define "_topdir ${BUILD_DIR}" \
  "${SCRIPT_DIR}/packaging/agef.spec"

# Copy generated RPMs to dist/
cp "${BUILD_DIR}/RPMS"/x86_64/*.rpm "$DIST_DIR/"

echo "RPM built successfully:"
ls -lh "${DIST_DIR}"/*.rpm

