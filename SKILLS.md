# Android-MCP-go Capability & Skills Reference

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

This document defines the capability map for human operators and AI agents interacting with **Android-MCP-go**.

Status Vocabulary:
- `PHYSICALLY VERIFIED`: Code complete, race tested (`go test -race ./...`), and verified on physical Android hardware (Sony Xperia `SOG09`).
- `TESTED`: Covered by automated unit and integration tests.
- `IMPLEMENTED`: Code complete and functionally working.

---

## 1. Device Management & Discovery

### List Devices
- **MCP Tools**: `ListDevices`, `device_list`
- **Description**: Enumerates connected USB physical devices, WiFi TCP/IP endpoints, and emulators.
- **Status**: `PHYSICALLY VERIFIED`

### Connect Device
- **MCP Tools**: `ConnectDevice`, `device_connect`
- **Description**: Connects to target WiFi device (`IP:Port`) or selects a USB serial.
- **Status**: `PHYSICALLY VERIFIED`

### Automatic USB → WiFi Bootstrap
- **Subsystem**: `DeviceManager` / `WirelessBootstrapper`
- **Description**: Detects USB device, resolves device WiFi IP address, switches ADB to TCP/IP `port 5555`, verifies link integrity, and persists connection state (`~/.android-mcp/android-mcp.json`).
- **Status**: `PHYSICALLY VERIFIED`

### Lazy Device Resolution
- **Subsystem**: Core Lifecycle
- **Description**: Allows instant MCP server startup without requiring a connected device upfront; defers resolution until tool execution.
- **Status**: `PHYSICALLY VERIFIED`

---

## 2. UI Inspection & Automation

### Layout Snapshot (XML & Vision)
- **MCP Tools**: `Snapshot`, `ui_snapshot`
- **Description**: Dumps UI element hierarchy XML. When `use_vision: true`, returns a visually annotated PNG screenshot with bounding boxes and element index badges.
- **Status**: `PHYSICALLY VERIFIED`

### Click Coordinate
- **MCP Tools**: `Click`, `ui_click`
- **Description**: Taps screen at specified `(x, y)` coordinate.
- **Status**: `PHYSICALLY VERIFIED`

### Click By Selector
- **MCP Tool**: `ClickBySelector`
- **Description**: Taps UI element matching `resourceId`, `text`, `contentDescription`, or `className`.
- **Status**: `PHYSICALLY VERIFIED`

### Long Click
- **MCP Tool**: `LongClick`
- **Description**: Performs long-press gesture at `(x, y)` coordinate with custom duration.
- **Status**: `PHYSICALLY VERIFIED`

### Multi-Touch Pinch Zoom
- **MCP Tools**: `Pinch`, `pinch`, `ui_pinch`
- **Description**: Performs 2-pointer multi-touch pinch gesture (`MotionEvent.obtain` multi-pointer engine in `mcp-helper.dex`) for live photo and canvas zoom in, zoom out, and scaling.
- **Status**: `PHYSICALLY VERIFIED`

### Swipe & Drag
- **MCP Tools**: `Swipe`, `Drag`
- **Description**: Performs touch swipe or stationary touch-down hold drag (800ms hold in `mcp-helper.dex`) from `(x1, y1)` to `(x2, y2)`.
- **Status**: `PHYSICALLY VERIFIED`

### Type Text
- **MCP Tool**: `Type`
- **Description**: Types text string into currently focused input control using high-speed Base64 `KeyCharacterMap` event stream in `mcp-helper.dex`.
- **Status**: `PHYSICALLY VERIFIED`

### Hardware Key Press
- **MCP Tool**: `Press`
- **Description**: Sends key code events (`KEYCODE_HOME`, `KEYCODE_BACK`, `KEYCODE_ENTER`, etc.).
- **Status**: `PHYSICALLY VERIFIED`

### Wait & WaitForElement
- **MCP Tools**: `Wait`, `WaitForElement`
- **Description**: Pauses execution or polls UI hierarchy until a target selector appears.
- **Status**: `PHYSICALLY VERIFIED`

---

## 3. Application Lifecycle Management

### List Applications
- **MCP Tool**: `list_apps`
- **Description**: Lists installed application packages on device (`third_party_only: bool`).
- **Status**: `PHYSICALLY VERIFIED`

### Launch Application
- **MCP Tool**: `launch_app`
- **Description**: Launches application package or main activity.
- **Status**: `PHYSICALLY VERIFIED`

### Stop Application
- **MCP Tool**: `stop_app`
- **Description**: Force-stops running application package.
- **Status**: `PHYSICALLY VERIFIED`

---

## 4. Filesystem & Shell Control

### File Push & Pull
- **MCP Tools**: `file_push`, `file_pull`
- **Description**: Transfers files between host machine and Android device storage.
- **Status**: `PHYSICALLY VERIFIED`

### Shell Execution
- **MCP Tool**: `shell_exec`
- **Description**: Executes arbitrary ADB shell command with context timeout and returns `stdout`, `stderr`, and `exit_code`.
- **Status**: `PHYSICALLY VERIFIED`

---

## 5. System Subsystems

### Managed Platform-Tools (`adb`)
- **CLI Commands**: `android-mcp platform-tools [status|update]`
- **Description**: Self-contained Platform-Tools manager under `~/.android-mcp/platform-tools/` downloading official Google releases with Zip Slip protection.
- **Status**: `PHYSICALLY VERIFIED`

### Managed `scrcpy` Live Screen Mirror
- **CLI Commands**: `android-mcp scrcpy [status|update|start|stop|capabilities|profile]`
- **Description**: Self-contained display mirroring engine under `~/.android-mcp/scrcpy/` with **always-on-top** display mode (`--always-on-top`), **single-instance invariant** (max 1 window), adaptive platform-optimized profile resolution (Metal on macOS, H.265/H.264), and automatic tool-call relaunching (`EnsureLiveView`).
- **Status**: `PHYSICALLY VERIFIED`

### Machine-Readable Skills Manifests
- **CLI Commands**: `android-mcp skills [list|install]`
- **Description**: Manages 10 domain capability manifests under `~/.android-mcp/skills/`.
- **Status**: `PHYSICALLY VERIFIED`
