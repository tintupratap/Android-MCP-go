# Android-MCP Porting Analysis: Python & scrcpy-wireless-go to Native Go

## 1. Overview & Objectives

This document presents the detailed architectural and functional analysis of **Android-MCP-py** and **scrcpy-wireless-go**, laying the technical foundation for **Android-MCP-go**. 

The goal of **Android-MCP-go** is to build a high-performance, single-binary, production-grade Go implementation of the Model Context Protocol (MCP) server for Android devices. Key enhancements over the Python baseline include:
- Native wireless ADB bootstrap (USB → TCP/IP WiFi auto-bootstrap & IP discovery).
- Persistent state management via `~/.android-mcp/android-mcp.json`.
- Integration/discovery fallback using `~/.scrcpy/scrcpy.json`.
- Platform-native notifications (`terminal-notifier` on macOS, `notify-send` on Linux).
- Lazy device resolution (server starts immediately even without connected devices).
- Zero runtime external dependencies (no Python runtime, no uiautomator2 daemon required).

---

## 2. Android-MCP-py Analysis

### 2.1 Architecture & Control Flow
- **Framework**: `FastMCP` (Python MCP protocol wrapper).
- **Entrypoint**: `android_mcp.__main__:main`. Parses CLI flags (`--device`, `--connection`, `--wifi`, `--usb`), evaluates environment variables (`ANDROID_MCP_DEVICE`, `ANDROID_MCP_CONNECTION`, `ANDROID_MCP_HOST`), and registers 14 MCP tools.
- **Service Layer**:
  - `Mobile` (`android_mcp.mobile.service`): Wraps ADB process invocations (`adb devices`, `adb connect`, `adb mdns services`) and `uiautomator2` device instance.
  - `Tree` (`android_mcp.tree.service`): Parses Android XML UI dumps, filters interactive elements, computes center coordinates, formats text tables (`tabulate`), and annotates screenshots (`PIL`).
- **Data Models**:
  - `ElementNode` (`name`, `class_name`, `coordinates`, `bounding_box`, `resource_id`)
  - `BoundingBox` (`x1`, `y1`, `x2`, `y2`)
  - `CenterCord` (`x`, `y`)
  - `TreeState` (`interactive_elements`)
  - `MobileState` (`tree_state`, `screenshot`)

### 2.2 Compatibility Tool Matrix

| Tool Name | Arguments | Description & Schema | Python Implementation | Planned Go Implementation |
|---|---|---|---|---|
| `ListDevices` | None | List available ADB devices | `Mobile.list_devices()` | `adb devices -l` parser |
| `ConnectDevice` | `serial: str` | Connect to device by serial/host:port | `Mobile.adb_connect()` + `mobile.connect()` | ADB TCP connect & DeviceManager update |
| `Device` | `action: "list"|"connect"|"disconnect"`, `serial?: str` | Unified device management | Delegates to List/Connect/Disconnect | Multi-action dispatcher |
| `Click` | `x: int`, `y: int` | Tap coordinate | `device.click(x, y)` | `adb shell input tap x y` |
| `ClickBySelector` | `text?`, `resourceId?`, `className?`, `description?`, `index?`, `timeout?` | Find element by selector & tap | Dump XML, match nodes, tap center | Dump XML, node evaluation, `input tap x y` |
| `Snapshot` | `use_vision: bool`, `use_annotation: bool` | Dump UI tree + optional annotated image | `dump_hierarchy()` + `screencap` + PIL draw | XML parser + `screencap -p` + Go `image/draw` |
| `LongClick` | `x: int`, `y: int` | Long press coordinate | `device.long_click(x, y)` | `adb shell input swipe x y x y 1000` |
| `Swipe` | `x1: int`, `y1: int`, `x2: int`, `y2: int` | Swipe gesture | `device.swipe(x1, y1, x2, y2)` | `adb shell input swipe x1 y1 x2 y2 300` |
| `Type` | `text: str`, `x: int`, `y: int`, `clear?: bool` | Focus/tap & enter text | `set_fastinput_ime` + `send_keys` | `input tap` + `input text` / ADB keyevents |
| `Drag` | `x1: int`, `y1: int`, `x2: int`, `y2: int` | Drag and drop gesture | `device.drag(x1, y1, x2, y2)` | `adb shell input swipe x1 y1 x2 y2 800` |
| `Press` | `button: str` | Hardware button press | `device.press(button)` | `adb shell input keyevent <KEYCODE>` |
| `Notification` | None | Open notification shade | `device.open_notification()` | `adb shell cmd statusbar expand-notifications` |
| `Wait` | `duration: int` | Pause execution | `device.sleep(duration)` | Go `time.Sleep(time.Duration(duration) * time.Second)` |
| `WaitForElement` | `text?`, `resourceId?`, `className?`, `description?`, `timeout?` | Wait for UI element to appear | Polling `wait(timeout)` | Retry loop dumping XML until selector match or timeout |

### 2.3 Dependencies & Go Counterparts
- `fastmcp` → Standard JSON-RPC 2.0 stdio MCP server in Go.
- `uiautomator2` → Pure ADB command execution (`exec-out screencap`, `uiautomator dump`, `input`).
- `pillow` → Standard library `image`, `image/draw`, `image/png`, `golang.org/x/image/font`.
- `tabulate` → Standard Go string formatting / clean tabular alignment.

---

## 3. scrcpy-wireless-go Analysis

### 3.1 Key Responsibilities & Capabilities
`scrcpy-wireless-go` is a reference CLI utility in Go that automates ADB over WiFi setup for `scrcpy`.
- **Path Resolution**: Discovers ADB and scrcpy binaries in PATH and `~/.scrcpy/`.
- **Config Persistence**: Reads/writes `~/.scrcpy/scrcpy.json`.
- **USB Detection & Device Parsing**: Parses `adb devices -l` output to identify connected USB physical devices, ignoring emulators and existing TCP/IP serials.
- **TCP/IP Bootstrap**: Executes `adb -s <serial> tcpip 5555`.
- **IP Address Discovery**: Multi-strategy extraction:
  1. `adb shell ip -4 addr show` (looks for `wlan`, `ap`, `eth` interfaces).
  2. `adb shell ip route` (extracts `src <IP>`).
  3. `adb shell getprop dhcp.wlan0.ipaddress`.
  4. Fallback search on any non-loopback, non-rmnet interface.
- **Connection Verification**: Executes `adb connect <IP>:5555` and checks `adb devices` output to confirm state is `"device"`.

---

## 4. Android-MCP-go Architecture Design

### 4.1 Package Structure

```text
Android-MCP-go/
├── cmd/
│   └── android-mcp/
│       └── main.go
├── internal/
│   ├── adb/            # Low-level ADB command execution & output parsing
│   ├── config/         # Persistent state (~/.android-mcp/android-mcp.json) & scrcpy.json reader
│   ├── device/         # Device models & selection logic
│   ├── discovery/      # IP discovery & wireless bootstrap engine
│   ├── logging/        # Structured logger
│   ├── mcp/            # MCP server protocol, JSON-RPC 2.0 stdio transport, tool registry
│   ├── notification/   # OS-native notification abstraction (macOS/Linux/fallback)
│   ├── server/         # Core server orchestrator binding DeviceManager & MCP
│   ├── ui/             # XML tree parser, element selector, bounding boxes & image drawing
│   └── wireless/       # Wireless state machine & reconnect manager
├── docs/
│   └── PORTING_ANALYSIS.md
├── go.mod
└── go.sum
```

### 4.2 Connection Priority Pipeline

```text
                     ┌───────────────────────┐
                     │  android-mcp starts   │
                     └───────────┬───────────┘
                                 │
                                 ▼
                     Read CLI Flags / Env Vars
                                 │
                                 ▼
                     Read android-mcp.json
                                 │
                                 ▼
                       Read scrcpy.json
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │ WiFi IP available?      │
                    └────┬───────────────┬────┘
                         │ YES           │ NO
                         ▼               ▼
                   Try WiFi ADB   ┌───────────────────────────┐
                         │        │ USB Serial available?     │
                         │        └─────┬───────────────┬─────┘
                         │              │ YES           │ NO
                         │              ▼               ▼
                         │       Bootstrap WiFi    Auto Discovery
                         │              │               │
                         └──────────────┴───────────────┘
                                        │
                                        ▼
                             Verify & Save State
```

---

## 5. Compatibility & Risk Mitigation

1. **Lazy Resolution**: Server initialization MUST NOT fail if no device is connected. Device resolution occurs when tools are executed or explicitly requested.
2. **Atomic Writes**: `android-mcp.json` is written via temporary file write and atomic rename (`os.Rename`), preventing corruption.
3. **Graceful Fallbacks**: If `terminal-notifier` or `notify-send` fails, log warning and continue; connection must never fail due to notification issues.
4. **Shell Injection Safety**: All ADB commands use `exec.CommandContext` with direct argument slices—never `sh -c` or concatenated string commands.
5. **Multi-device Safety**: Selection respects explicit flags > environment vars > persisted serial/IP > scrcpy state > physical over emulator auto-pick.
