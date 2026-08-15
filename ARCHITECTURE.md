# Android-MCP-go Architecture

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Overview

`Android-MCP-go` is a native Go implementation of the Android Model Context Protocol (MCP) server. It exposes mobile automation, screen vision, and device inspection capabilities to AI assistants via standard stdio JSON-RPC 2.0, featuring self-contained Platform-Tools management, managed `scrcpy` display mirroring, automatic wireless ADB bootstrapping, unified state persistence, and debug desktop notifications.

---

## 1. System Architecture Diagram

```text
                               ┌───────────────────────────┐
                               │        MCP Client         │
                               │ (Claude / Cursor / IDE)   │
                               └─────────────┬─────────────┘
                                             │ stdio (JSON-RPC 2.0)
                                             ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│ Android-MCP-go Server                                                                       │
│                                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │ internal/mcp (JSON-RPC Protocol Server & 23 Tool Handlers)                         │   │
│   └────────────────────────┬────────────────────────────────────────────────────────────┘   │
│                            │                                                                │
│                            ▼                                                                │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │ internal/service (DeviceService, InputService, UIService, AppService, etc.)         │   │
│   └────────────────────────┬────────────────────────────────────────────────────────────┘   │
│                            │ Lazy Device Access                                             │
│                            ▼                                                                │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │ internal/device (DeviceManager Orchestrator & State Machine)                        │   │
│   └──────┬─────────────────┬──────────────────┬──────────────────┬──────────────────────┘   │
│          │                 │                  │                  │                          │
│          ▼                 ▼                  ▼                  ▼                          │
│   ┌──────────────┐  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐                  │
│   │ internal/    │  │ internal/    │   │ internal/    │   │ internal/    │                  │
│   │ config       │  │ discovery    │   │ adb          │   │ ui           │                  │
│   └──────────────┘  └──────────────┘   └──────────────┘   └──────────────┘                  │
│          │                 │                  │                  │                          │
│          ▼                 ▼                  ▼                  ▼                          │
│   ┌────────────────────────────────────────────────────────────────────────┐                │
│   │ internal/platformtools (Managed SDK Platform-Tools / adb)              │                │
│   │ internal/scrcpy        (Managed scrcpy Release & Display Mirror)       │                │
│   └────────────────────────────────────────────────────────────────────────┘                │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Core Package Responsibilities

### `cmd/android-mcp/`
Main executable entry point. Parses CLI flags (`--device`, `--connection`, `--debug`), handles subcommands (`doctor`, `status`, `platform-tools`, `scrcpy`), boots the MCP stdio server, and wires the `DeviceManager`.

### `internal/config/`
Manages persistent state in `~/.android-mcp/android-mcp.json` using atomic temporary file replacement. Implements one-time migration (`PerformOneTimeMigration`) for importing legacy `~/.scrcpy/scrcpy.json` parameters.

### `internal/adb/`
Provides safe, structured execution of ADB commands via direct argument slice execution (`exec.CommandContext`). Manages connections, shell commands, file transfers (`PushFile`, `PullFile`), screencaps, and property queries.

### `internal/platformtools/`
Self-contained Platform-Tools manager. Automatically downloads, verifies, and installs official Google Android SDK Platform-Tools under `~/.android-mcp/platform-tools/` with Zip Slip protection.

### `internal/scrcpy/`
Managed `scrcpy` engine. Queries official GitHub Releases API (`Genymobile/scrcpy`), resolves OS/arch release assets, verifies SHA-256 digests, safely extracts archives, and launches non-blocking screen mirror display windows.

### `internal/discovery/`
Wireless ADB bootstrap engine:
- USB device detection.
- Multi-strategy IP address discovery (`ip -4 addr show`, `ip route`, `dhcp.wlan0.ipaddress`).
- TCP/IP mode setup (`adb tcpip 5555`).
- Wireless connection verification.

### `internal/device/`
Defines `DeviceManager` interface and state machine (`NoDevice`, `USBDetected`, `BootstrappingWiFi`, `WiFiConnected`, `USBConnected`). Enforces precedence logic (`android-mcp.json` > Auto-pick).

### `internal/service/`
Decoupled service layer mapping Android capabilities to MCP adapters:
- `DeviceService` (`ListDevices`, `ConnectDevice`, `Device`)
- `InputService` (`Click`, `LongClick`, `Swipe`, `Drag`, `Type`, `Press`)
- `UIService` (`Snapshot`, `ClickBySelector`, `WaitForElement`)
- `AppService` (`list_apps`, `launch_app`, `stop_app`)
- `FileService` (`file_push`, `file_pull`)
- `ShellService` (`shell_exec`)

### `internal/mcp/`
JSON-RPC 2.0 stdio transport protocol handler. Registers all 23 MCP tools and aliases.

### `internal/ui/`
Parses UI hierarchy XML from `uiautomator dump`. Extracts interactive elements, builds bounding boxes, computes element center coordinates, and draws visual annotations on PNG screenshots.

### `internal/notification/`
Cross-platform desktop notifier interface with rate-limited `--debug` activity notification engine (`ActivityNotifier`) and automatic secret redaction.

---

## 3. Persistent Configuration Schema (`~/.android-mcp/android-mcp.json`)

```json
{
  "version": 1,
  "device": {
    "last_ip": "192.168.1.3",
    "serial": "QV771A3JEE",
    "model": "SOG09",
    "port": 5555,
    "connection": "wifi"
  },
  "scrcpy": {
    "enabled": true,
    "auto_start": true,
    "video_codec": "h265",
    "video_bitrate": "4M",
    "audio_source": "playback",
    "stay_awake": true,
    "render_driver": "metal"
  },
  "platform_tools": {
    "managed": true,
    "path": "~/.android-mcp/platform-tools",
    "version": "1.0.41",
    "source": "official-google"
  },
  "managed_scrcpy": {
    "managed": true,
    "path": "~/.android-mcp/scrcpy",
    "release": "v4.1",
    "source": "https://github.com/Genymobile/scrcpy"
  },
  "notifications": {
    "enabled": true
  }
}
```
