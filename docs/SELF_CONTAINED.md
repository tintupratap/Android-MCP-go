# Android-MCP-go Self-Contained Runtime Architecture

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

`Android-MCP-go` operates as a **100% self-contained Android control stack** with **zero runtime dependencies** on host Android SDKs (`ANDROID_HOME`, `ANDROID_SDK_ROOT`), system `adb`, system `scrcpy`, or host package managers (Homebrew, MacPorts, apt, pacman).

All managed binaries, state persistence, skill manifests, downloads, and logs live strictly inside:

```text
~/.android-mcp/
```

(Configurable at runtime via `ANDROID_MCP_HOME`).

---

## 1. Directory Tree & Single Source of Truth

```text
~/.android-mcp/
├── android-mcp.json         # Unified persistent state & preferences
├── platform-tools/          # Managed Google Platform-Tools (adb, fastboot)
├── scrcpy/                  # Managed Genymobile scrcpy bundle & executable
├── skills/                  # Machine-readable skill domain manifests
├── .downloads/              # Temporary download staging area
├── .staging/                # Temporary archive extraction staging
├── cache/                   # Runtime cache
└── logs/                    # Operational logs
```

---

## 2. Component Pipeline & Safety Guarantees

### A. Managed Platform-Tools (`internal/platformtools`)
- Downloaded from official Google HTTPS endpoints (`dl.google.com/android/repository/platform-tools-latest-*.zip`).
- Validated with Zip Slip path traversal protection.
- Verified via `adb version` in staging before performing atomic replacements into `~/.android-mcp/platform-tools/`.

### B. Managed `scrcpy` (`internal/scrcpy`)
- Dynamically queries official GitHub Releases API (`api.github.com/repos/Genymobile/scrcpy/releases/latest`).
- Maps host `GOOS`/`GOARCH` to official assets (`darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`, `windows/386`).
- Verifies SHA-256 digests and performs `scrcpy --version` staging validation before atomic replacements.
- Launches background screen mirror child processes passing `ADB=~/.android-mcp/platform-tools/adb` in environment.

---

## 3. Clean-Machine & Portable Execution

On a fresh host machine without Android Studio, Android SDK, Go, Python, or package managers installed:
1. Run `curl -fsSL https://raw.githubusercontent.com/tintupratap/Android-MCP-go/main/install.sh | bash`.
2. `Android-MCP-go` automatically downloads and installs managed Platform-Tools, managed `scrcpy`, and skill manifests.
3. Connect an Android device via USB or WiFi. `Android-MCP-go` handles USB $\to$ WiFi bootstrap and opens a live display mirror window automatically.
