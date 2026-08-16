#!/usr/bin/env bash
# ==============================================================================
# Multi-Platform Release Build Script for Android-MCP-go
# Repository: https://github.com/tintupratap/Android-MCP-go
# Author: Ranapratap (tintupratap@gmail.com)
#
# Usage:
#   ./build_releases.sh [vX.Y.Z]
#   VERSION=v0.5.0 ./build_releases.sh
# ==============================================================================

set -eo pipefail

BOLD="\033[1m"
GREEN="\033[32m"
CYAN="\033[36m"
YELLOW="\033[33m"
RED="\033[31m"
RESET="\033[0m"

log() {
    printf "${CYAN}${BOLD}[Release Builder]${RESET} %s\n" "$1"
}

success() {
    printf "${GREEN}${BOLD}✓ %s${RESET}\n" "$1"
}

warn() {
    printf "${YELLOW}${BOLD}⚠️  %s${RESET}\n" "$1"
}

error() {
    printf "${RED}${BOLD}❌ %s${RESET}\n" "$1"
}

# 1. Resolve Release Version
VERSION="${1:-${VERSION}}"
if [ -z "$VERSION" ]; then
    if git describe --tags --always >/dev/null 2>&1; then
        VERSION="$(git describe --tags --always)"
    else
        VERSION="v0.5.0"
    fi
fi

# Ensure version prefix
if [[ "$VERSION" != v* ]]; then
    TAG_VERSION="v${VERSION}"
else
    TAG_VERSION="${VERSION}"
fi
RAW_VERSION="${TAG_VERSION#v}"

log "Building Multi-Platform Release Distribution: ${BOLD}${TAG_VERSION}${RESET}"

# 2. Setup Build Output Directory
DIST_DIR="dist"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

# 3. Define Matrix Targets (GOOS/GOARCH)
TARGETS=(
    "darwin/arm64"
    "darwin/amd64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
)

LDFLAGS="-s -w -X main.Version=${RAW_VERSION}"

# Common documentation files to include in release archives
DOC_FILES=("README.md" "LICENSE" "CHANGELOG.md" "install.sh")

# 4. Build Loop
for TARGET in "${TARGETS[@]}"; do
    GOOS="${TARGET%/*}"
    GOARCH="${TARGET#*/}"
    
    EXT=""
    if [ "$GOOS" == "windows" ]; then
        EXT=".exe"
    fi

    RAW_BIN_NAME="android-mcp-${GOOS}-${GOARCH}${EXT}"
    RAW_BIN_PATH="${DIST_DIR}/${RAW_BIN_NAME}"

    log "Cross-compiling for ${BOLD}${GOOS}/${GOARCH}${RESET}..."
    
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -trimpath \
        -ldflags "$LDFLAGS" \
        -o "$RAW_BIN_PATH" \
        ./cmd/android-mcp

    success "Built binary: ${RAW_BIN_NAME} ($(du -h "$RAW_BIN_PATH" | cut -f1))"

    # Package into tar.gz or zip archive
    ARCHIVE_STEM="android-mcp-${TAG_VERSION}-${GOOS}-${GOARCH}"
    STAGE_DIR="${DIST_DIR}/${ARCHIVE_STEM}"
    mkdir -p "$STAGE_DIR"

    cp "$RAW_BIN_PATH" "${STAGE_DIR}/android-mcp${EXT}"
    for DOC in "${DOC_FILES[@]}"; do
        if [ -f "$DOC" ]; then
            cp "$DOC" "$STAGE_DIR/"
        fi
    done

    if [ "$GOOS" == "windows" ]; then
        ARCHIVE_NAME="${ARCHIVE_STEM}.zip"
        (cd "$DIST_DIR" && zip -q -r "$ARCHIVE_NAME" "$ARCHIVE_STEM")
    else
        ARCHIVE_NAME="${ARCHIVE_STEM}.tar.gz"
        (cd "$DIST_DIR" && tar -czf "$ARCHIVE_NAME" "$ARCHIVE_STEM")
    fi

    rm -rf "$STAGE_DIR"
    success "Packaged archive: ${ARCHIVE_NAME}"
done

# 5. Generate SHA-256 Checksums
log "Generating SHA-256 Checksums..."
CDIR="$(pwd)"
cd "$DIST_DIR"

CHECKSUM_FILE="checksums.txt"
rm -f "$CHECKSUM_FILE"

if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 * > "$CHECKSUM_FILE"
elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum * > "$CHECKSUM_FILE"
else
    warn "Neither shasum nor sha256sum found. Skipping checksum generation."
fi

cd "$CDIR"

if [ -f "${DIST_DIR}/${CHECKSUM_FILE}" ]; then
    success "Created ${DIST_DIR}/${CHECKSUM_FILE}"
fi

# 6. Build Summary & GitHub Release Guide
echo ""
echo -e "${GREEN}${BOLD}==============================================================================${RESET}"
echo -e "${GREEN}${BOLD}  Multi-Platform Release Build Complete: ${TAG_VERSION}${RESET}"
echo -e "${GREEN}${BOLD}==============================================================================${RESET}"
echo ""
log "Generated Artifacts in ${BOLD}${DIST_DIR}/${RESET}:"
ls -lh "$DIST_DIR"
echo ""
log "To create a GitHub release automatically using GitHub CLI (gh):"
echo -e "${CYAN}  gh release create ${TAG_VERSION} dist/* --title \"Release ${TAG_VERSION}\" --notes-file CHANGELOG.md${RESET}"
echo ""
log "Alternatively, upload all files in ${BOLD}dist/${RESET} manually at:"
echo -e "${CYAN}  https://github.com/tintupratap/Android-MCP-go/releases/new${RESET}"
echo ""
