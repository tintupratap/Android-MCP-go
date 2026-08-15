# Android-MCP-go

[![Go Reference](https://pkg.go.dev/badge/github.com/tintupratap/Android-MCP-go.svg)](https://pkg.go.dev/github.com/tintupratap/Android-MCP-go)
[![CI Status](https://github.com/tintupratap/Android-MCP-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tintupratap/Android-MCP-go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Author](https://img.shields.io/badge/Author-Ranapratap-blue.svg)](mailto:tintupratap@gmail.com)
[![Tools: 23](https://img.shields.io/badge/MCP%20Tools-23%20Registered-brightgreen.svg)](#-supported-mcp-tools--capabilities)

**Android-MCP-go** is a high-performance, single-binary, production-grade Go implementation of the Model Context Protocol (MCP) server for Android devices.

It enables AI assistants (**Claude Desktop**, **Cursor IDE**, **Windsurf**, **AGY**, **Custom Agents**) to inspect, control, and automate Android smartphones and emulators over standard stdio JSON-RPC 2.0.

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

- **⚡ Native Go Runtime**: Zero Python overhead, instant startup (< 5ms), minimal footprint, single cross-platform binary.
- **🖥️ Managed scrcpy Live Mirroring**: Automatically manages `scrcpy` from official GitHub releases (`Genymobile/scrcpy`) under `~/.android-mcp/scrcpy/` and launches a live GUI display window upon device connection. Observe the physical device in real time while AI agents automate actions!
- **📦 Self-Contained Platform-Tools**: Automatically downloads, extracts, and manages official Google Android SDK Platform-Tools (`adb`, `fastboot`) under `~/.android-mcp/platform-tools/`. No manual ADB installation required!
- **📡 Automatic USB → WiFi Bootstrap**: Connect via USB once; `Android-MCP-go` automatically discovers the device's WiFi IP address, switches ADB to TCP/IP mode (`port 5555`), verifies connection integrity, and persists state. USB can then be unplugged!
- **💾 Atomic State Persistence**: Maintains connected device history in `~/.android-mcp/android-mcp.json` using atomic temporary file writes to prevent state corruption.
- **🔍 scrcpy Integration**: Reads external device state and video/audio preferences from `~/.scrcpy/scrcpy.json`.
- **🩺 Diagnostic Health Suite**: Includes `android-mcp doctor`, `android-mcp status`, `android-mcp platform-tools`, and `android-mcp scrcpy` subcommands.
- **🔔 Debug Activity Notifications**: Real-time desktop alerts for AI-agent actions (`--debug`) with rate-limiting anti-spam queues and sensitive data redaction.
- **💤 Lazy Device Resolution**: Server boots instantly even when no Android device is connected. Device resolution occurs when tools are invoked.
- **🖼️ Visual Vision Engine**: Generates UI layout tables and visually annotated PNG screenshots with bounding boxes and index badges.
- **🔒 Race-Free & Secure**: 100% test coverage with Go `-race` detector, strict argument slice isolation (no `sh -c`), Zip Slip archive security validation.

---

## ⚙️ MCP Client Configuration

Add `android-mcp` to your MCP client configuration:

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

### Enable Debug Notifications (`--debug`)

```json
{
  "mcpServers": {
    "android-mcp": {
      "command": "android-mcp",
      "args": [
        "--debug"
      ]
    }
  }
}
```

### Cursor IDE / Windsurf / VSCode

```json
{
  "mcpServers": {
    "android-mcp": {
      "command": "/usr/local/bin/android-mcp"
    }
  }
}
```

---

## 💻 CLI Commands & Subcommands

### System Doctor (`doctor`)

Run a comprehensive diagnostic check:

```bash
android-mcp doctor
```

Output:

```text
Android-MCP-Go Doctor
=====================

ADB:
  Binary:  /Users/ranapratap/.android-mcp/platform-tools/adb
  Version: Android Debug Bridge version 1.0.41
  Server:  running

Platform-Tools:
  Managed: yes (installed)
  Path:    /Users/ranapratap/.android-mcp/platform-tools
  Source:  Official Android/Google

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

### Operational Status (`status`)

Returns concise status with exit code `0` (Ready) or `1` (Not Ready):

```bash
android-mcp status
```

### Managed Platform-Tools Commands (`platform-tools`)

```bash
# Check platform-tools status
android-mcp platform-tools status

# Force update platform-tools to latest official Google release
android-mcp platform-tools update
```

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `ANDROID_MCP_ADB` | Explicit path to `adb` executable | `~/.android-mcp/platform-tools/adb` |
| `ANDROID_MCP_DEVICE` | Target device serial or `host:port` | Auto-detected |
| `ANDROID_MCP_CONNECTION` | Connection mode (`auto`, `usb`, `wifi`) | `auto` |
| `ANDROID_MCP_HOST` | WiFi IP or host | Auto-discovered |
| `LOG_LEVEL` | Logging level (`info`, `debug`) | `info` |

---

## 🛠️ Supported MCP Tools & Capabilities (23 Registered Tools)

| Tool Name | Aliases | Description | Read-Only |
|---|---|---|---|
| `ListDevices` | `device_list` | List available ADB devices (USB, WiFi, Emulators) | Yes |
| `ConnectDevice` | `device_connect` | Connect to ADB device by serial or `IP:port` | No |
| `Device` | — | Unified device manager (`list`, `connect`, `disconnect`) | No |
| `Snapshot` | `ui_snapshot` | Dump UI layout table (+ visual annotated PNG if `use_vision=True`) | Yes |
| `Click` | `ui_click` | Tap screen coordinate `(x, y)` | No |
| `ClickBySelector` | `ui_click_selector` | Locate element by selector (`text`, `resourceId`, `className`) & tap | No |
| `LongClick` | — | Long click screen coordinate `(x, y)` | No |
| `Swipe` | — | Swipe between `(x1, y1)` and `(x2, y2)` | No |
| `Drag` | — | Drag and drop gesture | No |
| `Type` | — | Type text at `(x, y)` coordinate | No |
| `Press` | — | Send keyevents (`home`, `back`, `power`, `volume_up`, `volume_down`, `enter`) | No |
| `Notification` | — | Pull down notification shade | No |
| `Wait` | — | Pause execution for `duration` seconds | No |
| `WaitForElement` | — | Polling wait for UI element to appear | Yes |
| `list_apps` | — | List installed packages (`third_party_only: bool`) | Yes |
| `launch_app` | — | Launch application package via `am start`/`monkey` | No |
| `stop_app` | — | Force-stop application package via `am force-stop` | No |
| `file_push` | — | Transfer host file to Android storage path | No |
| `file_pull` | — | Transfer Android file to host machine path | Yes |
| `shell_exec` | — | Run shell command returning `{ stdout, stderr, exit_code, duration_ms }` | No |

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
│          │                 │                  │                                         │
│          ▼                 ▼                  ▼                                         │
│   ┌──────────────────────────────────────────────────┐                                  │
│   │ internal/platformtools (Managed Platform-Tools)  │                                  │
│   └──────────────────────────────────────────────────┘                                  │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📜 Documentation

- [SKILLS.md](SKILLS.md) — Human and AI capability map for all 23 MCP tools.
- [ARCHITECTURE.md](ARCHITECTURE.md) — Package layout and internal design.
- [CONFIGURATION.md](CONFIGURATION.md) — Persistent state schema & discovery hierarchy.
- [SECURITY.md](SECURITY.md) — Download source policies, Zip Slip protection & argument safety.
- [DEVELOPMENT.md](DEVELOPMENT.md) — Development workflow, testing, benchmarks, and E2E suite.
- [docs/PLATFORM_TOOLS.md](docs/PLATFORM_TOOLS.md) — Self-contained platform-tools manager documentation.
- [docs/NOTIFICATIONS.md](docs/NOTIFICATIONS.md) — Desktop notification engine & `--debug` activity system.
- [CHANGELOG.md](CHANGELOG.md) — Version release notes.

---

## 👤 Author & Maintainer

- **Author**: Ranapratap
- **Email**: [tintupratap@gmail.com](mailto:tintupratap@gmail.com)
- **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
