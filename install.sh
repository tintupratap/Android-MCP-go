#!/usr/bin/env bash
# ==============================================================================
# Android-MCP-go Installer
# Repository: https://github.com/tintupratap/Android-MCP-go
# Author: Ranapratap (tintupratap@gmail.com)
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/tintupratap/Android-MCP-go/main/install.sh | bash
# ==============================================================================

set -e

BOLD="\033[1m"
GREEN="\033[32m"
CYAN="\033[36m"
YELLOW="\033[33m"
RED="\033[31m"
RESET="\033[0m"

log() {
    printf "${CYAN}${BOLD}[Android-MCP]${RESET} %s\n" "$1"
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

# 3. Build or Download Binary
INSTALLED=0

if command -v go >/dev/null 2>&1; then
    log "Found Go compiler ($(go version)). Building from source..."
    TMP_DIR="$(mktemp -d)"
    trap 'rm -rf "$TMP_DIR"' EXIT

    git clone --depth 1 https://github.com/tintupratap/Android-MCP-go.git "$TMP_DIR/repo" 2>/dev/null || true
    if [ -d "$TMP_DIR/repo" ]; then
        cd "$TMP_DIR/repo"
        go build -o "$BINARY_TARGET" ./cmd/android-mcp
        INSTALLED=1
    fi
fi

if [ "$INSTALLED" -eq 0 ]; then
    log "Downloading prebuilt release from GitHub..."
    RELEASE_URL="https://github.com/tintupratap/Android-MCP-go/releases/latest/download/android-mcp-${OS}-${ARCH}"
    if curl -fsSL "$RELEASE_URL" -o "$BINARY_TARGET" 2>/dev/null; then
        chmod +x "$BINARY_TARGET"
        INSTALLED=1
    fi
fi

if [ "$INSTALLED" -eq 0 ]; then
    error "Failed to install android-mcp. Ensure Go is installed or check internet connection."
    exit 1
fi

chmod +x "$BINARY_TARGET"
success "Binary installed to ${BINARY_TARGET}"

# 4. Ensure Platform-Tools & scrcpy Display Mirror
log "Ensuring official Android Platform-Tools..."
"$BINARY_TARGET" platform-tools update >/dev/null 2>&1 || true

log "Ensuring official scrcpy display mirror..."
"$BINARY_TARGET" scrcpy update >/dev/null 2>&1 || true

# 5. Verification & Health Check
log "Running installation health check..."
if command -v android-mcp >/dev/null 2>&1 || [ -x "$BINARY_TARGET" ]; then
    "$BINARY_TARGET" doctor
    echo ""
    success "Android-MCP-go installation complete and verified!"
    log "Configuration guide: https://github.com/tintupratap/Android-MCP-go#quick-start"
fi
