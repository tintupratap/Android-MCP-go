# Android-MCP-go Architecture

## Overview

`Android-MCP-go` is a native Go implementation of the Android Model Context Protocol (MCP) server. It exposes mobile automation and device inspection capabilities to AI assistants via the standard MCP stdio protocol, featuring automatic wireless ADB bootstrapping, persistent device state management, and desktop notification integration.

---

## 1. System Architecture

```text
                                ┌───────────────────────────┐
                                │       MCP Client          │
                                │ (Claude / Cursor / IDE)   │
                                └─────────────┬─────────────┘
                                              │ stdio (JSON-RPC 2.0)
                                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│ Android-MCP-go Server                                                                    │
│                                                                                          │
│   ┌──────────────────────────────────────────────────────────────────────────────────┐   │
│   │ mcp/ (JSON-RPC Protocol Server & 14 Tool Handlers)                               │   │
│   └────────────────────────┬─────────────────────────────────────────────────────────┘   │
│                            │ Lazy Device Access                                          │
│                            ▼                                                             │
│   ┌──────────────────────────────────────────────────────────────────────────────────┐   │
│   │ device/ (DeviceManager Orchestrator & State Machine)                             │   │
│   └──────┬─────────────────┬──────────────────┬──────────────────┬───────────────────┘   │
│          │                 │                  │                  │                       │
│          ▼                 ▼                  ▼                  ▼                       │
│   ┌──────────────┐  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐               │
│   │ config/      │  │ discovery/   │   │ adb/         │   │ ui/          │               │
│   │ Persistent   │  │ IP & Wireless│   │ ADB Command  │   │ XML Parser & │               │
│   │ State        │  │ Bootstrap    │   │ Runner       │   │ Visual Annot.│               │
│   └──────────────┘  └──────────────┘   └──────────────┘   └──────────────┘               │
│          │                 │                  │                                          │
└──────────┼─────────────────┼──────────────────┼──────────────────────────────────────────┘
           │                 │                  │
           ▼                 ▼                  ▼
┌────────────────────┐ ┌──────────┐ ┌────────────────────────┐
│ ~/.android-mcp/    │ │ macOS /  │ │ Connected Android      │
│ android-mcp.json   │ │ Linux    │ │ Device (USB or Wi-Fi)  │
│ & scrcpy.json      │ │ Notifier │ └────────────────────────┘
└────────────────────┘ └──────────┘
```

---

## 2. Core Package Responsibilities

### `cmd/android-mcp/`
Main executable entry point. Parses CLI flags (`--device`, `--wifi`, `--usb`, `--connection`), initializes logging, boots the MCP stdio server, and wires the `DeviceManager`.

### `internal/config/`
Manages persistent state in `~/.android-mcp/android-mcp.json` using atomic file writes (write to temp file + `os.Rename`). Also handles reading discovery data from `~/.scrcpy/scrcpy.json` if available.

### `internal/adb/`
Provides safe, structured execution of ADB commands via `exec.CommandContext`. Responsible for executing `devices -l`, `connect`, `disconnect`, `tcpip`, `shell input`, `shell uiautomator dump`, and `exec-out screencap`.

### `internal/discovery/`
Contains wireless bootstrap engine:
- USB device detection.
- Multi-strategy IP address discovery (`ip -4 addr show`, `ip route`, `dhcp.wlan0.ipaddress`).
- TCP/IP mode setup (`adb tcpip 5555`).
- Wireless connection verification.

### `internal/device/`
Defines `DeviceManager` interface and explicit state machine (`NoDevice`, `USBDetected`, `BootstrappingWiFi`, `VerifyingWiFi`, `WiFiConnected`, `Reconnecting`). Handles connection priority (CLI > Env > android-mcp.json > scrcpy.json > Auto-pick physical device over emulator).

### `internal/mcp/`
Implements the JSON-RPC 2.0 stdio MCP transport protocol. Registers all 14 tools (`ListDevices`, `ConnectDevice`, `Device`, `Click`, `ClickBySelector`, `Snapshot`, `LongClick`, `Swipe`, `Type`, `Drag`, `Press`, `Notification`, `Wait`, `WaitForElement`).

### `internal/ui/`
Parses UI hierarchy XML returned by `uiautomator dump`. Extracts interactive elements, builds layout bounding boxes, formats clean text tables, computes element center coordinates, and draws visual annotations on PNG screenshots.

### `internal/notification/`
Cross-platform desktop notifier interface. Invokes `terminal-notifier` on macOS, `notify-send` on Linux, or logs silently on unsupported platforms. Failures are non-fatal.

---

## 3. Persistent Configuration Schema (`android-mcp.json`)

```json
{
  "last_ip": "192.168.1.3",
  "device_serial": "QV771A3JEE",
  "device_model": "SOG09",
  "port": 5555,
  "connection": "wifi",
  "last_seen": "2026-08-15T05:35:46Z",
  "last_successful_connection": "2026-08-15T05:35:46Z",
  "wifi_enabled": true,
  "usb_bootstrap_enabled": true
}
```

---

## 4. Connection Priority & Resolution Logic

1. **CLI Flags**: `--device`, `--wifi`, `--usb`, `--connection`.
2. **Environment Variables**: `ANDROID_MCP_DEVICE`, `ANDROID_MCP_CONNECTION`, `ANDROID_MCP_HOST`.
3. **Android-MCP State**: `~/.android-mcp/android-mcp.json`.
4. **scrcpy State**: `~/.scrcpy/scrcpy.json`.
5. **Auto Discovery**: Scans connected USB/WiFi devices and prefers physical devices over emulators.
