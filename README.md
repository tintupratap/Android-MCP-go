# Android-MCP-go

[![Go Reference](https://pkg.go.dev/badge/github.com/tintupratap/Android-MCP-go.svg)](https://pkg.go.dev/github.com/tintupratap/Android-MCP-go)
[![CI Status](https://github.com/tintupratap/Android-MCP-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tintupratap/Android-MCP-go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Author](https://img.shields.io/badge/Author-Ranapratap-blue.svg)](mailto:tintupratap@gmail.com)
[![Tools: 25](https://img.shields.io/badge/MCP%20Tools-25%20Registered-brightgreen.svg)](#-supported-mcp-tools--capabilities)

**Android-MCP-go** is a high-performance, 100% self-contained Go server implementing the Model Context Protocol (MCP) for Android devices.

It empowers AI assistants (**Claude Desktop**, **Cursor IDE**, **Windsurf**, **AGY**, **Custom Agents**) to inspect, control, and automate physical Android smartphones and emulators via standard stdio JSON-RPC 2.0.

---

## ⚡ Installation Methods

### Option 1: 1-Line Automated Installer (Recommended)

Install `android-mcp` automatically on **macOS** or **Linux** with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/tintupratap/Android-MCP-go/main/install.sh | bash
```

### Option 2: Install via `go install`

If you have Go installed on your system, you can install directly via Go Package Manager:

```bash
go install github.com/tintupratap/Android-MCP-go/cmd/android-mcp@latest
```

### Option 3: Build From Source

```bash
git clone https://github.com/tintupratap/Android-MCP-go.git
cd Android-MCP-go

# (Optional) Recompile the embedded Android DEX helper (mcp-helper.dex) from Java source:
# Requires Android SDK platform android.jar and d8 build-tool:
# mkdir -p build/out
# javac -cp $ANDROID_HOME/platforms/android-34/android.jar -d build/out internal/adb/java/com/android/mcp/HelperMain.java
# $ANDROID_HOME/build-tools/34.0.0/d8 --output ./ build/out/com/android/mcp/HelperMain.class
# mv classes.dex internal/adb/mcp-helper.dex

# Build the self-contained Go binary (pre-bundled with internal/adb/mcp-helper.dex):
go build -o android-mcp ./cmd/android-mcp
sudo cp android-mcp /usr/local/bin/
```

The installation automatically initializes:
1. `android-mcp` binary available in your system `PATH`.
2. Official **Google Android SDK Platform-Tools** (`adb`) into `~/.android-mcp/platform-tools/`.
3. Official **Genymobile `scrcpy`** display mirror into `~/.android-mcp/scrcpy/`.
4. Embedded **Native Android Helper (`mcp-helper.dex`)** compiled with Android API 34.
5. Machine-readable **Skill Manifests** into `~/.android-mcp/skills/`.

---

## 🚀 Highlights

- **⚡ Native Embedded Android Engine (`mcp-helper.dex`)**: Bundles a 9.3 KB embedded DEX helper (`//go:embed mcp-helper.dex`) for high-speed hardware touch injection, multi-touch gestures, and in-memory UI dumps.
- **🤏 Native Multi-Touch Pinch Zoom (`Pinch`, `pinch`, `ui_pinch`)**: 2-pointer `MotionEvent` engine for live photo and canvas pinch-to-zoom (zoom in, zoom out, scaling).
- **✋ Stationary Touch-Down Hold Drag (`Drag`)**: Injects `ACTION_DOWN` with a 800ms stationary hold to trigger Android's View long-press listener for 100% reliable launcher icon and item dragging.
- **🔤 Base64 Unicode Text Engine (`Type`)**: Injects raw `KeyEvents` directly via `KeyCharacterMap`, bypassing shell character escaping bugs across special characters, emojis, and non-ASCII text.
- **🚀 Ultra-Fast In-Memory UI Dumper (< 40ms)**: Direct in-memory tree traversal using `mcp-helper.dex` without `/sdcard/` filesystem disk I/O.
- **⚡ Instant & Native**: Compiled Go binary with startup < 5ms, zero Python overhead, and minimal memory footprint.
- 🖥️ **Managed `scrcpy` & Single-Instance Live View**: Auto-installs official `scrcpy` releases with a **strict single-instance invariant** (0 or 1 window, never 2+). Automatically opens on boot without requiring tool calls.
- 👁️ **Persistent Visual Observability**: If you manually close the `scrcpy` window, the MCP server and ADB remain connected. Upon the next device tool call, `Android-MCP-go` automatically restores `scrcpy` **before executing the action** so you can continuously observe AI behavior.
- **📦 100% Self-Contained**: Manages its own Platform-Tools (`adb`) and `scrcpy` binaries inside `~/.android-mcp/`. Zero dependency on host Android SDK, `ANDROID_HOME`, system ADB, or package managers.
- **📡 Wireless Auto-Bootstrap**: Plug in USB once; `Android-MCP-go` automatically discovers the device's WiFi IP, switches ADB to TCP/IP mode (`port 5555`), verifies connection integrity, and persists state. USB can be disconnected immediately.
- **🖼️ Vision & Annotated Screenshots**: Generates structured XML UI element trees and visually annotated PNG screenshots with bounding boxes and element index badges.
- **🩺 Comprehensive CLI Tooling**: Includes `doctor`, `status`, `platform-tools`, `scrcpy`, and `skills` subcommands.
- **🔒 Race-Free & Secure**: 100% concurrency safety (`go test -race ./...`), direct argument slice execution (no `sh -c`), and Zip/Tar Slip archive security validation.

---

## ⚙️ Client Integration Setup

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
      "args": ["--debug"]
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

## 💻 CLI Commands & Diagnostics

```bash
# Run comprehensive diagnostic health check
android-mcp doctor

# Check quick operational status (Exit 0 = Ready, Exit 1 = Error)
android-mcp status

# Inspect or re-sync machine-readable capability skills
android-mcp skills list
android-mcp skills install

# Manage platform-tools
android-mcp platform-tools status
android-mcp platform-tools update

# Manage scrcpy live display mirror
android-mcp scrcpy status
android-mcp scrcpy update
android-mcp scrcpy start
android-mcp scrcpy stop
```

---

## 🛠️ Supported MCP Tools & Capabilities (25 Tools & Aliases)

| Tool Name | Aliases | Description | Read-Only |
|---|---|---|---|
| `ListDevices` | `device_list` | List connected USB, WiFi, and emulator devices | Yes |
| `ConnectDevice` | `device_connect` | Connect to WiFi device target (`IP:Port`) or auto-resolve physical target | No |
| `Device` | — | Query active device metadata, model, serial, and connection mode (`action`: `list`, `connect`, `disconnect`, `get`, `info`) | Yes |
| `Snapshot` | `ui_snapshot` | Fast in-memory UI hierarchy XML or annotated vision PNG (`use_vision: true`) | Yes |
| `Click` | `ui_click` | Perform high-speed hardware screen tap at `(x, y)` via `InputManager` | No |
| `ClickBySelector` | — | Click UI element matching text, resource ID, accessibility ID, or XPath | No |
| `LongClick` | — | Perform long-press gesture at `(x, y)` with custom duration | No |
| `Swipe` | — | Perform swipe gesture from `(x1, y1)` to `(x2, y2)` | No |
| `Drag` | — | Stationary touch-down hold (800ms) + smooth drag gesture across workspace & views | No |
| `Pinch` | `pinch`, `ui_pinch` | Multi-touch 2-pointer pinch gesture for live zoom-in, zoom-out, and scaling | No |
| `Type` | — | High-speed Base64 character stream input supporting Emojis, Unicode, and special characters | No |
| `Press` | — | Trigger hardware key events (`KEYCODE_HOME`, `KEYCODE_BACK`, `KEYCODE_ENTER`, etc.) | No |
| `Notification` | — | Display toast notification on Android screen | No |
| `Wait` | — | Pause execution for specified duration (seconds) | Yes |
| `WaitForElement` | — | Poll UI hierarchy until specified element selector appears | Yes |
| `list_apps` | — | List installed application packages (`third_party_only: bool`) | Yes |
| `launch_app` | — | Launch application package via package name or intent | No |
| `stop_app` | — | Force-stop running application package | No |
| `file_push` | — | Transfer host file to Android storage path | No |
| `file_pull` | — | Transfer Android file to host machine path | Yes |
| `shell_exec` | — | Run arbitrary ADB shell command returning stdout, stderr, and exit code | No |

---

## 🏗️ System Architecture

```text
                               ┌───────────────────────────┐
                               │        MCP Client         │
                               │ (Claude / Cursor / IDE)   │
                               └─────────────┬─────────────┘
                                             │ stdio (JSON-RPC 2.0)
                                             ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ Android-MCP-go Server (~/.android-mcp/)                                                 │
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
│   ┌───────────────────────────────┐   ┌───────────────────────────────┐                 │
│   │ Managed Platform-Tools (adb)  │   │ Managed scrcpy (Display Window)│                 │
│   └───────────────────────────────┘   └───────────────────────────────┘                 │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📜 Documentation & Contributing

- [SKILLS.md](SKILLS.md) — Human and AI capability map across all 23 MCP tools.
- [ARCHITECTURE.md](ARCHITECTURE.md) — Package layout, internal design, and subsystem flow.
- [CONFIGURATION.md](CONFIGURATION.md) — Persistent state schema & discovery hierarchy.
- [SECURITY.md](SECURITY.md) — Download source policies, Zip/Tar Slip protection & security audit.
- [DEVELOPMENT.md](DEVELOPMENT.md) — Development workflow, testing, benchmarks, and E2E suite.
- [CONTRIBUTING.md](CONTRIBUTING.md) — Guidelines for contributing and pull requests.
- [CREDITS.md](CREDITS.md) — Open source licenses, third-party software credits & acknowledgments.
- [docs/TESTING_REPORT.md](docs/TESTING_REPORT.md) — Comprehensive hardware & software testing report.
- [docs/SELF_CONTAINED.md](docs/SELF_CONTAINED.md) — 100% self-contained runtime architecture guide.
- [docs/SCRCPY.md](docs/SCRCPY.md) — Managed scrcpy & live display mirror documentation.
- [docs/SCRCPY_OPTIMIZATION.md](docs/SCRCPY_OPTIMIZATION.md) — Adaptive scrcpy optimization & progressive degradation architecture.
- [docs/PLATFORM_TOOLS.md](docs/PLATFORM_TOOLS.md) — Self-contained platform-tools manager documentation.
- [docs/NOTIFICATIONS.md](docs/NOTIFICATIONS.md) — Desktop notification engine & `--debug` activity alerts.
- [docs/CONFIGURATION_MIGRATION.md](docs/CONFIGURATION_MIGRATION.md) — Unified state schema & one-time migration guide.
- [CHANGELOG.md](CHANGELOG.md) — Version release history.

---

## 👏 Credits & Special Thanks

- **Romain Vimont ([@rom1v](https://github.com/rom1v)) & Genymobile** for [`scrcpy`](https://github.com/Genymobile/scrcpy).
- **Jeomon George ([@jeo_geo_alukka](https://github.com/Jeomon)) & CursorTouch** for [`Android-MCP`](https://github.com/CursorTouch/Android-MCP), the inspiration.
- **The Android Open Source Project (AOSP) / Google** for the Android Debug Bridge (ADB) and Platform-Tools.
- **Anthropic** for the Model Context Protocol (MCP) open specification.
- **The Go Authors** for the Go language compiler and runtime.

---

## 🏷️ Trademarks & Disclaimers

Android, Android Studio, Google, and Google Play are trademarks of **Google LLC**. Sony and Xperia are trademarks of **Sony Group Corporation**. All other trademarks belong to their respective owners. This project is an independent open-source tool and is not affiliated with or endorsed by Google LLC or Sony Group Corporation.

---

## 👤 Author & Maintainer

- **Author**: Ranapratap
- **Email**: [tintupratap@gmail.com](mailto:tintupratap@gmail.com)
- **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)

---

## 📄 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.
