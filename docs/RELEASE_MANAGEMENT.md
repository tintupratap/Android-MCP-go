# Dynamic scrcpy Release Management & Resolution

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Overview

`Android-MCP-go` automatically manages its `scrcpy` binary installation under:
```text
~/.android-mcp/scrcpy/
```

All release downloads originate exclusively from the official upstream project:
```text
https://github.com/Genymobile/scrcpy
```

---

## Dynamic Release Resolution Pipeline

```text
Fetch GitHub Latest Release API (api.github.com/repos/Genymobile/scrcpy/releases/latest)
                              │
                              ▼
        Filter out Drafts and Prereleases
                              │
                              ▼
   Match Host Platform & Architecture (runtime.GOOS & runtime.GOARCH)
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
  Verify SHA-256 Checksum (if published in release assets)
                              │
                              ▼
  Safely Extract to ~/.android-mcp/.staging/scrcpy/
  (Tar.gz & Zip with Zip Slip path traversal validation)
                              │
                              ▼
  Verify Executable & Run `scrcpy --version`
                              │
                              ▼
  Atomic Replacement to ~/.android-mcp/scrcpy/
```

---

## CLI Management Subcommands

```bash
# Check installed scrcpy version and running status
android-mcp scrcpy status

# Query GitHub Releases API & update to latest official release
android-mcp scrcpy update

# Force reinstall scrcpy
android-mcp scrcpy reinstall
```
