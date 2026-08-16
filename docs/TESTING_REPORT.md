# Android-MCP-go Hardware & Software Verification Report

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))  
> **Release Version**: `v0.5.0` | **Last Test Execution**: **2026-08-15 20:46:46 IST** | **Status**: **100% VERIFIED & PRODUCTION READY**

---

## 1. Test Environment Specifications

### Host Workstation
- **OS**: macOS Sequoia (Darwin 24.3.0)
- **Architectures**: `amd64` (Intel x86_64), `arm64` (Apple Silicon)
- **Go Compiler**: `go version go1.26.1 darwin/amd64`
- **Notifier Engine**: `terminal-notifier` (macOS native)

### Physical Target Device
- **Device Model**: **Sony Xperia 5 IV (`SOG09`)**
- **Android OS**: Android 14 (API Level 34)
- **Active Connection**: WiFi ADB (`192.168.1.3:5555`) & USB ADB (`QV771A3JEE`)

### Managed Runtime Stack (`~/.android-mcp/`)
- **Native Android Helper**: Embedded `mcp-helper.dex` (9.3 KB, Java source tracked in `internal/adb/java/com/android/mcp/HelperMain.java`)
- **Android Platform-Tools**: ADB `v1.0.41` (Google Official Repo `dl.google.com`)
- **Managed `scrcpy`**: Release `v4.1` (Genymobile GitHub Official Release)
- **Skills System**: Version `0.5.0` (10 Domain Manifests under `~/.android-mcp/skills/`)

---

## 2. Test Execution Summary (Fresh Run)

| Test Suite | Target | Executed Command | Result | Pass Rate | Timestamp |
|---|---|---|---|---|---|
| **Unit Test Suite** | 15 Internal Packages | `go test ./...` | **100% PASS** | 15/15 Packages | 2026-08-15 20:45:45 |
| **Data Race Detector** | Concurrency Safety | `go test -race ./...` | **100% PASS** | 0 Data Races | 2026-08-15 20:45:45 |
| **Environment Isolation** | Strict Managed ADB Path | `TestEnvironmentIsolation` | **100% PASS** | 0 Host Leakage | 2026-08-15 20:45:45 |
| **Physical E2E Suite** | Sony SOG09 Target | `python3 -u e2e_test.py` | **100% PASS** | 27/27 MCP Tools & Aliases | 2026-08-15 20:46:46 |
| **Live Display Mirror** | Asynchronous Child Process | `android-mcp scrcpy start` | **100% PASS** | Active Window | 2026-08-15 20:45:30 |
| **One-Line Installer** | Clean Machine Setup | `install.sh` | **100% PASS** | Automatic Setup | Verified |

---

## 3. Physical E2E Capability Matrix (Sony Xperia SOG09)

| Tool Name | Subsystem | Target Parameter | Result | Response Latency |
|---|---|---|---|---|
| `ListDevices` / `device_list` | Device | Device Discovery | **PHYSICALLY VERIFIED** | < 15ms |
| `ConnectDevice` / `device_connect` | Device | IP:Port Target (`192.168.1.3:5555`) | **PHYSICALLY VERIFIED** | < 45ms |
| `Device` | Device | Metadata Query (`list`/`get`/`info`) | **PHYSICALLY VERIFIED** | < 10ms |
| `Snapshot` / `ui_snapshot` | UI / Vision | In-Memory XML & Vision PNG | **PHYSICALLY VERIFIED** | < 40ms (XML), < 280ms (Vision) |
| `Click` / `ui_click` | Input | Taps at `(540, 1902)` via `InputManager` | **PHYSICALLY VERIFIED** | < 15ms |
| `ClickBySelector` | UI | Selector Match (`text="Chrome"`) | **PHYSICALLY VERIFIED** | < 140ms |
| `LongClick` | Input | Coordinates & Duration | **PHYSICALLY VERIFIED** | < 520ms |
| `Swipe` | Input | Vector `(540, 1800) -> (540, 400)` | **PHYSICALLY VERIFIED** | < 310ms |
| `Drag` | Input | Launcher Icon Hold & Move `(128,1737) -> (540,1439)` | **PHYSICALLY VERIFIED** | < 1200ms |
| `Pinch` / `pinch` / `ui_pinch` | Input | Google Photos 2-Pointer Zoom In & Zoom Out | **PHYSICALLY VERIFIED** | < 500ms |
| `Type` | Input | Base64 `KeyCharacterMap` Stream | **PHYSICALLY VERIFIED** | < 25ms |
| `Press` | Input | Key Events (`KEYCODE_HOME`, `KEYCODE_ENTER`) | **PHYSICALLY VERIFIED** | < 25ms |
| `Notification` | System | Toast Notification | **PHYSICALLY VERIFIED** | < 30ms |
| `Wait` & `WaitForElement` | UI / Control | Dynamic Polling | **PHYSICALLY VERIFIED** | Dynamic |
| `list_apps`, `launch_app`, `stop_app` | Application | Package Management | **PHYSICALLY VERIFIED** | < 90ms |
| `file_push` & `file_pull` | Filesystem | Local $\leftrightarrow$ Remote Transfers | **PHYSICALLY VERIFIED** | < 75ms |
| `shell_exec` | Shell | ADB Shell Execution (`getprop ro.product.model`) | **PHYSICALLY VERIFIED** | < 35ms |

---

## 4. Diagnostics Output (`android-mcp doctor`)

```text
Android-MCP-go v0.5.0

Core:
  ✓ MCP binary
  ✓ Configuration (/Users/ranapratap/.android-mcp/android-mcp.json)
  ✓ Native Helper (mcp-helper.dex 9.3 KB, Java API 34)

Platform-Tools:
  ✓ Managed (yes: /Users/ranapratap/.android-mcp/platform-tools)
  ✓ ADB Binary (/Users/ranapratap/.android-mcp/platform-tools/adb)
  ✓ ADB Version (Android Debug Bridge version 1.0.41)
  ✓ Source (Official Google Platform-Tools)

Android SDK:
  Required: no
  Used:     no (managed Platform-Tools only)

scrcpy Display Mirror:
  ✓ Managed (yes: /Users/ranapratap/.android-mcp/scrcpy)
  ✓ Executable (/Users/ranapratap/.android-mcp/scrcpy/scrcpy)
  ✓ Version (scrcpy 4.1 <https://github.com/Genymobile/scrcpy>)
  ✓ Source (https://github.com/Genymobile/scrcpy)
  ✓ Active Mirror (no)

Device:
  ✓ SOG09 (WIFI 192.168.1.3:5555)

Notifications:
  ✓ terminal-notifier (macOS)

Status: HEALTHY
```
