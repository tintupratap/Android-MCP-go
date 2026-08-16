#!/usr/bin/env bash
# ==============================================================================
# Android-MCP-go Installer
# Repository: https://github.com/tintupratap/Android-MCP-go
# Releases: https://github.com/tintupratap/Android-MCP-go/releases
# Author: Ranapratap (tintupratap@gmail.com)
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/tintupratap/Android-MCP-go/main/install.sh | bash
#   VERSION=v0.5.0 curl -fsSL https://raw.githubusercontent.com/tintupratap/Android-MCP-go/main/install.sh | bash
# ==============================================================================

set -e

BOLD="\033[1m"
GREEN="\033[32m"
CYAN="\033[36m"
YELLOW="\033[33m"
RED="\033[31m"
RESET="\033[0m"

log() {
    printf "\033[36m\033[1m[Android-MCP]\033[0m %s\n" "$1"
}

success() {
    printf "\033[32m\033[1m✓ %s\033[0m\n" "$1"
}

warn() {
    printf "\033[33m\033[1m⚠️  %s\033[0m\n" "$1"
}

error() {
    printf "\033[31m\033[1m❌ %s\033[0m\n" "$1"
}

log "Starting Android-MCP-go Installation..."

# 1. Detect OS & Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    darwin)
        OS="darwin"
        ;;
    linux)
        OS="linux"
        ;;
    *)
        error "Unsupported operating system: $OS"
        exit 1
        ;;
esac

log "Detected Environment: OS=${OS}, ARCH=${ARCH}"

# 2. Determine Installation Path
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *)
            warn "Adding $INSTALL_DIR to PATH recommended."
            ;;
    esac
fi

BINARY_TARGET="${INSTALL_DIR}/android-mcp"
INSTALLED=0

# 3. Resolve Release Tag (Supports Pre-releases, Releases, and Git Tags)
if [ -n "$VERSION" ]; then
    TAG="$VERSION"
else
    log "Querying latest release/pre-release tag from GitHub..."
    TAG=$(curl -fsSL -H "User-Agent: Android-MCP-Installer" "https://api.github.com/repos/tintupratap/Android-MCP-go/releases" 2>/dev/null | grep -o '"tag_name": *"[^"]*"' | head -n 1 | cut -d'"' -f4 || true)
    
    if [ -z "$TAG" ]; then
        TAG=$(curl -fsSL -H "User-Agent: Android-MCP-Installer" "https://api.github.com/repos/tintupratap/Android-MCP-go/tags" 2>/dev/null | grep -o '"name": *"[^"]*"' | head -n 1 | cut -d'"' -f4 || true)
    fi

    if [ -z "$TAG" ]; then
        TAG=$(curl -fsSL "https://github.com/tintupratap/Android-MCP-go/releases" 2>/dev/null | grep -o '/tintupratap/Android-MCP-go/releases/tag/[^"]*' | head -n 1 | cut -d'/' -f6 || true)
    fi
fi

TMP_BIN="$(mktemp)"
trap 'rm -f "$TMP_BIN"' EXIT

if [ -n "$TAG" ]; then
    log "Targeting GitHub Release Tag: ${TAG}"
    
    # Try downloading standalone binary asset for resolved tag
    RELEASE_URL="https://github.com/tintupratap/Android-MCP-go/releases/download/${TAG}/android-mcp-${OS}-${ARCH}"
    if curl -cL --fail --silent --show-error "$RELEASE_URL" -o "$TMP_BIN" 2>/dev/null; then
        chmod +x "$TMP_BIN"
        mv "$TMP_BIN" "$BINARY_TARGET"
        INSTALLED=1
        success "Downloaded prebuilt release binary (${TAG}) from GitHub Releases!"
    else
        warn "Direct release binary download failed for ${TAG}. Attempting release tarball extraction..."
        TMP_DIR="$(mktemp -d)"
        trap 'rm -rf "$TMP_DIR"' EXIT

        # Try downloading release tarball archive for resolved tag
        TAR_URLS=(
            "https://github.com/tintupratap/Android-MCP-go/releases/download/${TAG}/android-mcp-${TAG}-${OS}-${ARCH}.tar.gz"
            "https://github.com/tintupratap/Android-MCP-go/releases/download/${TAG}/android-mcp-v0.5.0-${OS}-${ARCH}.tar.gz"
        )

        for TAR_URL in "${TAR_URLS[@]}"; do
            if curl -cL --fail --silent "$TAR_URL" -o "$TMP_DIR/archive.tar.gz" 2>/dev/null; then
                tar -xzf "$TMP_DIR/archive.tar.gz" -C "$TMP_DIR"
                EXTRACTED_BIN="$(find "$TMP_DIR" -type f -name "android-mcp" | head -n 1)"
                if [ -f "$EXTRACTED_BIN" ]; then
                    chmod +x "$EXTRACTED_BIN"
                    mv "$EXTRACTED_BIN" "$BINARY_TARGET"
                    INSTALLED=1
                    success "Extracted prebuilt release binary from tarball archive!"
                    break
                fi
            fi
        done
    fi
fi

# Fallback download attempt via releases/latest endpoint if tag resolution failed
if [ "$INSTALLED" -eq 0 ]; then
    log "Attempting fallback download from GitHub Releases latest endpoint..."
    LATEST_URL="https://github.com/tintupratap/Android-MCP-go/releases/latest/download/android-mcp-${OS}-${ARCH}"
    if curl -cL --fail --silent --show-error "$LATEST_URL" -o "$TMP_BIN" 2>/dev/null; then
        chmod +x "$TMP_BIN"
        mv "$TMP_BIN" "$BINARY_TARGET"
        INSTALLED=1
        success "Downloaded prebuilt release binary from GitHub Releases latest!"
    fi
fi

# 4. Fallback: Build from Source if Go Compiler is Installed
if [ "$INSTALLED" -eq 0 ]; then
    if command -v go >/dev/null 2>&1; then
        warn "Release download unavailable. Falling back to building from source using Go ($(go version))..."
        TMP_REPO="$(mktemp -d)"
        trap 'rm -rf "$TMP_REPO"' EXIT

        if git clone --depth 1 https://github.com/tintupratap/Android-MCP-go.git "$TMP_REPO/repo" 2>/dev/null; then
            cd "$TMP_REPO/repo"
            go build -ldflags="-s -w -X main.Version=0.5.0" -o "$BINARY_TARGET" ./cmd/android-mcp
            chmod +x "$BINARY_TARGET"
            INSTALLED=1
            success "Built and installed binary from source!"
        fi
    fi
fi

if [ "$INSTALLED" -eq 0 ]; then
    error "Failed to install android-mcp. Check network connection or releases page: https://github.com/tintupratap/Android-MCP-go/releases"
    exit 1
fi

chmod +x "$BINARY_TARGET"
success "Binary installed to ${BINARY_TARGET}"

# 5. Ensure Platform-Tools, scrcpy Display Mirror & Skills
log "Ensuring official Android Platform-Tools..."
"$BINARY_TARGET" platform-tools update >/dev/null 2>&1 || true

log "Ensuring official scrcpy display mirror..."
"$BINARY_TARGET" scrcpy update >/dev/null 2>&1 || true

log "Ensuring machine-readable skills manifests..."
"$BINARY_TARGET" skills install >/dev/null 2>&1 || true

# 6. Verification & Health Check
log "Running installation health check..."
if command -v android-mcp >/dev/null 2>&1 || [ -x "$BINARY_TARGET" ]; then
    "$BINARY_TARGET" doctor
    echo ""
    success "Android-MCP-go installation complete and verified!"
    log "Releases page: https://github.com/tintupratap/Android-MCP-go/releases"
fi
