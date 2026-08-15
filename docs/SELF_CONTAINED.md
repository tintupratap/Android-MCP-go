# Android-MCP-go Self-Contained Runtime Architecture

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Overview

`Android-MCP-go` is built as a **100% self-contained Android control stack**.

It operates with **ZERO** runtime dependencies on:
- Android SDK / Android Studio
- Environment variables (`ANDROID_HOME`, `ANDROID_SDK_ROOT`)
- System `adb` or `fastboot`
- System `scrcpy`
- Package managers (Homebrew, MacPorts, apt, pacman)
- Legacy directories (`~/.scrcpy`, `scrcpy-wireless-go`)

The single, authoritative managed root directory for all binaries, configuration, downloads, logs, and cache is:
```text
~/.android-mcp/
```

(Or configurable via `ANDROID_MCP_HOME`).

---

## Directory Layout

```text
~/.android-mcp/
├── android-mcp.json             # Unified persistent state & preferences
├── platform-tools/              # Managed Google Android SDK Platform-Tools
│   ├── adb                      # Authoritative ADB executable
│   ├── fastboot                 # Managed fastboot binary
│   └── ...
├── scrcpy/                      # Managed Genymobile scrcpy bundle
│   ├── scrcpy                   # Authoritative scrcpy executable
│   ├── scrcpy-server            # Supporting server jar
│   └── ...                      # Supporting SDL/FFmpeg libraries
├── .downloads/                  # Temporary archive download directory
├── .staging/                    # Temporary extraction and verification staging
├── cache/                       # Cached runtime artifacts
└── logs/                        # Logging output
```

---

## Component Management

### 1. Managed Platform-Tools (`internal/platformtools`)
- Downloaded directly from official Google Android HTTPS repositories (`dl.google.com/android/repository/platform-tools-latest-*.zip`).
- Extracted securely with Zip Slip path traversal protection.
- Atomically installed into `~/.android-mcp/platform-tools/`.

### 2. Managed scrcpy (`internal/scrcpy`)
- Dynamically queries official GitHub Releases API (`api.github.com/repos/Genymobile/scrcpy/releases/latest`).
- Resolves release assets based on `runtime.GOOS` and `runtime.GOARCH` (`darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`, `windows/386`).
- Verifies SHA-256 digests if available in release checksum assets.
- Validates executable by running `scrcpy --version` before atomic directory replacement into `~/.android-mcp/scrcpy/`.

---

## Clean-Machine & Portable Execution

- On a fresh machine without Android Studio, Android SDK, Go, Python, or system package managers installed:
  1. Compile or download `android-mcp`.
  2. Run `android-mcp doctor`.
  3. `Android-MCP-go` automatically downloads and installs managed Platform-Tools and `scrcpy`.
  4. Plug in physical USB Android device or connect over WiFi.
  5. `Android-MCP-go` executes automatic USB $\to$ WiFi ADB bootstrap and launches `scrcpy` live display mirror.
