# Android-MCP-go Changelog

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

All notable changes to **Android-MCP-go** are documented in this file. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [0.4.0] - 2026-08-15

### Added
- **100% Self-Contained Runtime Architecture**:
  - Eliminated all runtime dependencies on host Android SDKs (`ANDROID_HOME`, `ANDROID_SDK_ROOT`), system `adb`, system `fastboot`, system `scrcpy`, or host package managers.
  - Centralized path resolver `RuntimePaths` supporting custom `ANDROID_MCP_HOME` root override.
  - Added unit test `TestEnvironmentIsolation` verifying zero host SDK path leakage.
- **Managed `scrcpy` & Live Display Mirroring (`internal/scrcpy`)**:
  - Dynamically resolves latest official GitHub Releases (`Genymobile/scrcpy`).
  - SHA-256 integrity verification, Tar/Zip Slip archive security protection, and atomic installation under `~/.android-mcp/scrcpy/`.
  - Non-blocking auto-launch of display mirror window (`scrcpy -s <serial>`) upon device connection.
  - Duplicate process protection per device serial and clean SIGTERM exit handling.
- **Machine-Readable Skills System (`internal/skills`)**:
  - Automatically installs 10 machine-readable skill domain JSON manifests under `~/.android-mcp/skills/`.
  - Subcommands: `android-mcp skills list` and `android-mcp skills install`.
- **One-Line Installer Script (`install.sh`)**:
  - Automated one-command installation of binary, Platform-Tools, `scrcpy`, and skills manifests via `curl -fsSL ... | bash`.
- **Legal & Contribution Suite**:
  - `LICENSE` (MIT), `CONTRIBUTING.md`, `CREDITS.md`, `docs/TESTING_REPORT.md`, `docs/SELF_CONTAINED.md`.

---

## [0.3.0] - 2026-08-15

### Added
- **Self-Contained Platform-Tools Management (`internal/platformtools`)**:
  - Downloads official Google Platform-Tools (`dl.google.com`) under `~/.android-mcp/platform-tools/`.
  - Zip Slip security validation and atomic installation via `~/.android-mcp/.staging/`.
- **Debug Activity Desktop Notifications (`--debug`)**:
  - Rate-limited async desktop alert queue with action correlation IDs and parameter credential redaction.

---

## [0.2.0] - 2026-08-15

### Added
- **Diagnostic Health Suite & Core CLI Commands**:
  - `android-mcp doctor` comprehensive diagnostic health report.
  - `android-mcp status` quick operational check.
- **Expanded Tool Capabilities**:
  - `list_apps`, `launch_app`, `stop_app`, `file_push`, `file_pull`, `shell_exec`.

---

## [0.1.0] - 2026-08-15

### Added
- Initial production Go implementation of Android MCP server.
- JSON-RPC 2.0 stdio server transport with 14 core tools.
- Automatic USB $\to$ WiFi ADB wireless bootstrap engine.
- Persistent state schema `~/.android-mcp/android-mcp.json` written via atomic file replacement.
