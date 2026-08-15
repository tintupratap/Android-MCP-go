# Android-MCP-go

[![Go Reference](https://pkg.go.dev/badge/github.com/tintupratap/Android-MCP-go.svg)](https://pkg.go.dev/github.com/tintupratap/Android-MCP-go)
[![CI Status](https://github.com/tintupratap/Android-MCP-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tintupratap/Android-MCP-go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Author](https://img.shields.io/badge/Author-Ranapratap-blue.svg)](mailto:tintupratap@gmail.com)

**Android-MCP-go** is a high-performance, single-binary, production-grade Go implementation of the Model Context Protocol (MCP) server for Android devices.

It allows AI assistants (**Claude Desktop**, **Cursor IDE**, **Windsurf**, **Custom Agents**) to inspect, control, and automate Android hardware over standard stdio JSON-RPC 2.0.

---

## ⚡ Quick One-Line Installation

Install `android-mcp` automatically on **macOS** or **Linux** with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/tintupratap/Android-MCP-go/main/install.sh | bash
```

### Alternative Installation Methods

#### Go Install (Go 1.22+)

```bash
go install github.com/tintupratap/Android-MCP-go/cmd/android-mcp@latest
```

#### Build From Source

```bash
git clone https://github.com/tintupratap/Android-MCP-go.git
cd Android-MCP-go
go build -o android-mcp ./cmd/android-mcp
```

---

## 🚀 Key Features

- **⚡ Native Go Architecture**: Zero Python runtime overhead, instant server startup, low memory footprint, single binary deployment.
- **📡 Automatic USB → WiFi Bootstrap**: Connect via USB once; `Android-MCP-go` automatically discovers the device's WiFi IP address, switches ADB to TCP/IP mode (`port 5555`), verifies connection integrity, and persists state. USB can then be unplugged!
- **💾 Atomic State Persistence**: Maintains connected device history in `~/.android-mcp/android-mcp.json` using atomic temporary file writes to prevent state corruption.
- **🔍 scrcpy Integration**: Reads external device state from `~/.scrcpy/scrcpy.json` if present.
- **🩺 Diagnostic Health Suite**: Includes `android-mcp doctor` and `android-mcp status` commands for environment troubleshooting.
- **🔔 Native Desktop Notifications**: Alerts when wireless connection succeeds so USB can be disconnected safely (`terminal-notifier` on macOS, `notify-send` on Linux).
- **💤 Lazy Device Resolution**: MCP server boots instantly even when no Android device is connected. Device resolution occurs when tools are invoked.
- **🖼️ Screen Vision & Visual Annotations**: Generates UI layout tables and visually annotated PNG screenshots with bounding boxes and index badges.
- **🔒 Race-Free & Secure**: Tested with Go `-race` detector, strict argument array isolation (no shell string evaluation).

---

## ⚙️ MCP Client Configuration

Add `android-mcp` to your MCP client configuration file:

### Claude Desktop (`claude_desktop_config.json`)

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "android-mcp": {
      "command": "android-mcp"
    }
  }
}
```

### Cursor IDE / Windsurf / Custom Agents

Add to your MCP server settings:

```json
{
  "mcpServers": {
    "android-mcp": {
      "command": "/usr/local/bin/android-mcp"
    }
  }
}
```

*Note: Dynamic device discovery handles IP/serial connection automatically. You do NOT need to hardcode dynamic IP addresses in static MCP client configs.*

---

## 💻 CLI Usage & Commands

### Health Check Doctor

Run a complete system diagnostic check:

```bash
android-mcp doctor
```

Output sample:

```text
Android-MCP-Go Doctor
=====================

ADB:
  Binary: /Users/ranapratap/.scrcpy/adb
  Version: Android Debug Bridge version 1.0.41
  Server: running

Configuration:
  android-mcp.json: OK (~/.android-mcp/android-mcp.json)
  scrcpy.json:      OK (~/.scrcpy/scrcpy.json)

Devices:
  USB:  none detected
  WiFi: 1 connected

Selected Device:
  Model:    SOG09
  Serial:   192.168.1.3:5555
  Endpoint: 192.168.1.3:5555

Notifications:
  terminal-notifier (macOS): available
  notify-send (Linux):       unavailable

MCP:
  Transport: stdio
  Tools: 23 registered

Status: HEALTHY
```

### Quick Status

Check operational readiness:

```bash
android-mcp status
```

### CLI Arguments & Options

```bash
# Default mode (uses persistent state & auto discovery)
android-mcp

# Explicitly target a USB device serial
android-mcp --usb QV771A3JEE

# Explicitly target a WiFi IP or HOST:PORT
android-mcp --wifi 192.168.1.3

# Explicitly target specific device target
android-mcp --device 192.168.1.3:5555

# Verbose debug logging
android-mcp --debug
```

### Environment Variables

| Variable | Description |
|---|---|
| `ANDROID_MCP_DEVICE` | Explicit device serial or `host:port` |
| `ANDROID_MCP_CONNECTION` | Preferred connection type: `auto`, `usb`, `wifi` |
| `ANDROID_MCP_HOST` | WiFi host or IP address |
| `LOG_LEVEL` | Set to `debug` for verbose logs |

---

## 🛠️ Supported MCP Tools & Capabilities (23 Registered Tools)

| Tool Name | Description | Read-Only |
|---|---|---|
| `ListDevices` / `device_list` | List available ADB devices (USB, WiFi, Emulators) | Yes |
| `ConnectDevice` / `device_connect` | Connect to an ADB device by serial number or IP:port | No |
| `Device` | Unified device manager (`list`, `connect`, `disconnect`) | No |
| `Click` / `ui_click` | Tap screen coordinate `(x, y)` | No |
| `ClickBySelector` / `ui_click_selector` | Locate element by selector (`text`, `resourceId`, `className`, `description`) & tap | No |
| `Snapshot` / `ui_snapshot` | Return UI hierarchy table (+ visual annotated screenshot PNG if `use_vision=True`) | Yes |
| `LongClick` | Long click screen coordinate `(x, y)` | No |
| `Swipe` | Swipe between coordinates `(x1, y1)` and `(x2, y2)` | No |
| `Type` | Focus & type text at `(x, y)` with clear option | No |
| `Drag` | Drag and drop gesture | No |
| `Press` | Send keyevents (`home`, `back`, `power`, `volume_up`, `volume_down`, `enter`) | No |
| `Notification` | Open notification shade | No |
| `Wait` | Pause execution for `duration` seconds | No |
| `WaitForElement` | Polling wait for dynamic element to appear on screen | Yes |
| `list_apps` | List installed application packages (`third_party_only: bool`) | Yes |
| `launch_app` | Launch application package via `am start`/`monkey` | No |
| `stop_app` | Force-stop application package via `am force-stop` | No |
| `file_push` | Transfer local file to Android storage path | No |
| `file_pull` | Transfer remote Android file to local machine path | Yes |
| `shell_exec` | Run structured shell command returning `{ stdout, stderr, exit_code, duration_ms }` | No |

---

## 🏗️ Architecture

```text
                               ┌───────────────────────────┐
                               │        MCP Client         │
                               │ (Claude / Cursor / IDE)   │
                               └─────────────┬─────────────┘
                                             │ stdio (JSON-RPC 2.0)
                                             ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ Android-MCP-go Server                                                                   │
│                                                                                         │
│   ┌─────────────────────────────────────────────────────────────────────────────────┐   │
│   │ internal/mcp (JSON-RPC Protocol Server & 23 Tool Handlers)                      │   │
│   └────────────────────────┬────────────────────────────────────────────────────────┘   │
│                            │                                                            │
│                            ▼                                                            │
│   ┌─────────────────────────────────────────────────────────────────────────────────┐   │
│   │ internal/service (DeviceService, InputService, UIService, AppService, etc.)     │   │
│   └────────────────────────┬────────────────────────────────────────────────────────┘   │
│                            │ Lazy Device Access                                         │
│                            ▼                                                            │
│   ┌─────────────────────────────────────────────────────────────────────────────────┐   │
│   │ internal/device (DeviceManager Orchestrator & State Machine)                    │   │
│   └──────┬─────────────────┬──────────────────┬──────────────────┬──────────────────┘   │
│          │                 │                  │                  │                      │
│          ▼                 ▼                  ▼                  ▼                      │
│   ┌──────────────┐  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐              │
│   │ internal/    │  │ internal/    │   │ internal/    │   │ internal/    │              │
│   │ config       │  │ discovery    │   │ adb          │   │ ui           │              │
│   └──────────────┘  └──────────────┘   └──────────────┘   └──────────────┘              │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📜 Documentation

- [SKILLS.md](SKILLS.md) — Complete human and AI capability map.
- [ARCHITECTURE.md](ARCHITECTURE.md) — Deep architectural design and package layout.
- [CONFIGURATION.md](CONFIGURATION.md) — Persistent state schema & resolution hierarchy.
- [SECURITY.md](SECURITY.md) — Trust boundary models & command injection prevention.
- [DEVELOPMENT.md](DEVELOPMENT.md) — Development setup & tool extension guide.
- [CHANGELOG.md](CHANGELOG.md) — Release notes and version history.

---

## 👤 Author & Maintainer

- **Author**: Ranapratap
- **Email**: [tintupratap@gmail.com](mailto:tintupratap@gmail.com)
- **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
