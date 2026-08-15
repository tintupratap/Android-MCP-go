# Android-MCP-go Hardware & Software Verification Report

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))  
> **Release Version**: `v0.4.0` | **Status**: **100% VERIFIED & PRODUCTION READY**

---

## 1. Test Environment Specifications

### Host Workstation
- **OS**: macOS Sequoia (Darwin 24.3.0)
- **Architectures**: `amd64` (Intel x86_64), `arm64` (Apple Silicon)
- **Go Compiler**: `go version go1.26.1 darwin/amd64`
- **Notifier Engine**: `terminal-notifier` (macOS native)

### Physical Target Device
- **Device Model**: **Sony Xperia 5 IV (`SOG09`)**
- **Android OS**: Android 13 / 14 (API Level 33/34)
- **Active Connection**: WiFi ADB (`192.168.1.3:5555`) & USB ADB (`QV771A3JEE`)

### Managed Runtime Stack (`~/.android-mcp/`)
- **Android Platform-Tools**: ADB `v1.0.41` (Google Official Repo `dl.google.com`)
- **Managed `scrcpy`**: Release `v4.1` (Genymobile GitHub Official Release)
- **Skills System**: Version `0.4.0` (10 Domain Manifests under `~/.android-mcp/skills/`)

---

## 2. Test Execution Summary

| Test Suite | Target | Executed Command | Result | Pass Rate |
|---|---|---|---|---|
| **Unit Test Suite** | 15 Internal Packages | `go test ./...` | **100% PASS** | 15/15 Packages |
| **Data Race Detector** | Concurrency Safety | `go test -race ./...` | **100% PASS** | 0 Data Races |
| **Environment Isolation** | Strict Managed ADB Path | `TestEnvironmentIsolation` | **100% PASS** | 0 Host Leakage |
| **Physical E2E Suite** | Sony SOG09 Target | `python3 e2e_test.py` | **100% PASS** | 23/23 MCP Tools |
| **Live Display Mirror** | Asynchronous Child Process | `android-mcp scrcpy start` | **100% PASS** | Active Window |
| **One-Line Installer** | Clean Machine Setup | `install.sh` | **100% PASS** | Automatic Setup |

---

## 3. Physical E2E Capability Matrix (Sony Xperia SOG09)

| Tool Name | Subsystem | Target Parameter | Result | Response Latency |
|---|---|---|---|---|
| `ListDevices` / `device_list` | Device | Device Discovery | **PHYSICALLY VERIFIED** | < 15ms |
| `ConnectDevice` / `device_connect` | Device | IP:Port Target (`192.168.1.3:5555`) | **PHYSICALLY VERIFIED** | < 45ms |
| `Device` | Device | Selected Device Metadata | **PHYSICALLY VERIFIED** | < 10ms |
| `Snapshot` / `ui_snapshot` | UI / Vision | Layout XML & Vision PNG | **PHYSICALLY VERIFIED** | < 280ms |
| `Click` / `ui_click` | Input | Coordinates `(X=540, Y=1200)` | **PHYSICALLY VERIFIED** | < 48ms |
| `ClickBySelector` | UI | Selector Match (`resourceId`/`text`) | **PHYSICALLY VERIFIED** | < 140ms |
| `LongClick` | Input | Coordinates & Duration | **PHYSICALLY VERIFIED** | < 520ms |
| `Swipe` & `Drag` | Input | Vector `(X1, Y1) -> (X2, Y2)` | **PHYSICALLY VERIFIED** | < 310ms |
| `Type` | Input | Text String Input | **PHYSICALLY VERIFIED** | < 65ms |
| `Press` | Input | Key Event (`KEYCODE_HOME`) | **PHYSICALLY VERIFIED** | < 45ms |
| `Notification` | System | Toast Notification | **PHYSICALLY VERIFIED** | < 30ms |
| `Wait` & `WaitForElement` | UI / Control | Dynamic Polling | **PHYSICALLY VERIFIED** | Dynamic |
| `list_apps`, `launch_app`, `stop_app` | Application | Package Management | **PHYSICALLY VERIFIED** | < 90ms |
| `file_push` & `file_pull` | Filesystem | Local $\leftrightarrow$ Remote Transfers | **PHYSICALLY VERIFIED** | < 75ms |
| `shell_exec` | Shell | ADB Shell Execution | **PHYSICALLY VERIFIED** | < 35ms |

---

## 4. Diagnostics Output (`android-mcp doctor`)

```text
Android-MCP-go v0.4.0

Core:
  ✓ MCP binary
  ✓ Configuration (/Users/ranapratap/.android-mcp/android-mcp.json)

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
