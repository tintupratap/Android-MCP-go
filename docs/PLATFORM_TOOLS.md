# Self-Contained Android SDK Platform-Tools Management

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Overview

`Android-MCP-go` automatically manages its own Android SDK Platform-Tools (`adb`, `fastboot`) so that users do not need to manually install ADB or configure system paths.

The managed Platform-Tools directory is stored at:
```text
~/.android-mcp/platform-tools/
```

---

## Official Download Source Policy

All Platform-Tools downloads originate **exclusively** from official Google Android developer repositories:

- **macOS**: `https://dl.google.com/android/repository/platform-tools-latest-darwin.zip`
- **Linux**: `https://dl.google.com/android/repository/platform-tools-latest-linux.zip`
- **Windows**: `https://dl.google.com/android/repository/platform-tools-latest-windows.zip`

`Android-MCP-go` strictly refuses downloads from third-party mirrors, user-provided URLs, or unofficial binary hosts.

---

## Security & Installation Mechanics

1. **Atomic Installation**: Downloads archive into temporary directory `~/.android-mcp/platform-tools.download/`, extracts contents, verifies binary execution (`adb version`), and atomically swaps into place (`~/.android-mcp/platform-tools/`). If an error occurs, previous working installations remain intact.
2. **Zip Slip Protection**: Zip entries are strictly validated during extraction to ensure file paths cannot escape target destination directories.
3. **Executable Permission Normalization**: Unix file modes are normalized to `0755` for `adb` and `fastboot`, and `0644` for support files.

---

## Resolution Precedence

```text
1. Explicit ANDROID_MCP_ADB environment variable override
       ↓
2. Managed ~/.android-mcp/platform-tools/adb
       ↓
3. System PATH adb / ~/.scrcpy/adb / Android SDK
       ↓
4. Automatic official Google Platform-Tools download
```

---

## CLI Management Commands

```bash
# Check Platform-Tools status
android-mcp platform-tools status

# Force update Platform-Tools to latest official release
android-mcp platform-tools update
```
