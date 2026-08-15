# Dynamic Release Management & Resolution Guide

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

`Android-MCP-go` automatically manages Platform-Tools and `scrcpy` releases dynamically without requiring hardcoded URLs or manual binary updates.

---

## 1. Dynamic Release Resolution Pipeline

```text
Query GitHub API (api.github.com/repos/Genymobile/scrcpy/releases/latest)
                               │
                               ▼
        Filter Out Drafts & Prereleases
                               │
                               ▼
    Match Host Platform & Arch (runtime.GOOS & runtime.GOARCH)
    ├── darwin/arm64  -> scrcpy-macos-aarch64-v*.tar.gz
    ├── darwin/amd64  -> scrcpy-macos-x86_64-v*.tar.gz
    ├── linux/amd64   -> scrcpy-linux-x86_64-v*.tar.gz
    ├── windows/amd64 -> scrcpy-win64-v*.zip
    └── windows/386   -> scrcpy-win32-v*.zip
                               │
                               ▼
   Download Archive to ~/.android-mcp/.downloads/
                               │
                               ▼
   Verify SHA-256 Digest Against Official Release Checksums
                               │
                               ▼
   Extract Safely to ~/.android-mcp/.staging/scrcpy/
   (Zip & Tar Slip Path Traversal Validation)
                               │
                               ▼
   Verify Executable via `scrcpy --version` Staging Check
                               │
                               ▼
   Atomic Directory Replacement to ~/.android-mcp/scrcpy/
```

---

## 2. Platform-Tools Management Pipeline

Google Platform-Tools archives (`platform-tools-latest-*.zip`) are fetched from `dl.google.com`, validated with Zip Slip checks, tested via `adb version` in staging, and atomically installed into `~/.android-mcp/platform-tools/`.

---

## 3. CLI Release Management Commands

```bash
# Check installed component status
android-mcp platform-tools status
android-mcp scrcpy status

# Force update to latest official releases
android-mcp platform-tools update
android-mcp scrcpy update
```
