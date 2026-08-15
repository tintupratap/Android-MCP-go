# Android-MCP-go Skills & Capability Map

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

This document outlines the full capability map for **Android-MCP-go**, serving as a comprehensive reference for humans and AI agents.

Status Vocabulary:
- `PHYSICALLY VERIFIED`: Implemented, unit tested, race verified (`go test -race`), and physically verified on physical Android hardware (e.g. Sony Xperia SOG09).
- `TESTED`: Implemented and covered by automated unit/integration tests.
- `IMPLEMENTED`: Code complete and functionally working.
- `PLANNED`: Designed and scheduled for future milestones.
- `EXPERIMENTAL`: Early release feature under active iteration.
- `UNSUPPORTED`: Explicitly out of scope or blocked by platform security constraints.

---

## 1. Device Management

### Skill: List Devices
- **Description**: Query connected ADB devices (USB physical, WiFi TCP/IP, and emulators).
- **MCP tool**: `ListDevices` / `device_list`
- **Arguments**: None
- **Return value**: Text table mapping serials to ADB states (`device`, `offline`, `unauthorized`)
- **Requires device**: No
- **Requires root**: No
- **Supported Android versions**: Android 4.4+ (API 19+)
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/adb/adb_test.go`, `e2e_test.py`

### Skill: Connect Device
- **Description**: Connect to a remote device over TCP/IP or explicitly select a USB serial.
- **MCP tool**: `ConnectDevice` / `device_connect`
- **Arguments**: `serial: string`
- **Return value**: Success/failure confirmation message
- **Requires device**: No
- **Requires root**: No
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/device/device_test.go`, `e2e_test.py`

### Skill: Automatic USB → WiFi Bootstrap
- **Description**: Automatically detect USB physical device, discover device WiFi IP address, switch ADB to TCP/IP port 5555, verify link, notify user, and persist state.
- **MCP tool**: Internal `DeviceManager` / `WirelessBootstrapper`
- **Arguments**: `usbDev: ADBDevice`, `targetPort: int`
- **Return value**: `BootstrapResult` (WiFi IP, serial, model)
- **Requires device**: Yes (USB initially)
- **Requires root**: No
- **Supported Android versions**: Android 5.0+ (API 21+)
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/discovery/bootstrap_test.go`

### Skill: Lazy Device Resolution
- **Description**: Deferred device connection—allows MCP server startup without requiring a connected device.
- **MCP tool**: Server core lifecycle / `RequireDevice()`
- **Arguments**: None
- **Return value**: `*Device` pointer or formatted configuration help error
- **Requires device**: No (server start); Yes (tool execution)
- **Requires root**: No
- **Supported Android versions**: All
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/mcp/mcp_test.go`

---

## 2. UI Automation

### Skill: Click Coordinate
- **Description**: Tap specific screen coordinate `(x, y)`.
- **MCP tool**: `Click` / `ui_click`
- **Arguments**: `x: int`, `y: int`
- **Return value**: Confirmation string `Clicked on (x,y)`
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/adb/adb_test.go`, `e2e_test.py`

### Skill: Selector-Based Click
- **Description**: Locate UI element matching text, resourceId, className, or description and click its center coordinate.
- **MCP tool**: `ClickBySelector` / `ui_click_selector`
- **Arguments**: `text?`, `resourceId?`, `className?`, `description?`, `index?`, `timeout?`
- **Return value**: Target element match info or timeout error string
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 5.0+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/ui/ui_test.go`, `e2e_test.py`

### Skill: Swipe & Drag
- **Description**: Perform touch swipe or drag-and-drop gesture between coordinates `(x1,y1)` and `(x2,y2)`.
- **MCP tool**: `Swipe`, `Drag` / `ui_swipe`, `ui_drag`
- **Arguments**: `x1: int`, `y1: int`, `x2: int`, `y2: int`
- **Return value**: Gesture status string
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/adb/adb_test.go`, `e2e_test.py`

### Skill: Key & Hardware Button Press
- **Description**: Send Android keyevents (`KEYCODE_HOME`, `KEYCODE_BACK`, `KEYCODE_POWER`, etc.).
- **MCP tool**: `Press` / `input_press`
- **Arguments**: `button: string`
- **Return value**: Keypress confirmation string
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/adb/adb_test.go`, `e2e_test.py`

### Skill: Type Text
- **Description**: Enter text at specific coordinates with optional field clearing.
- **MCP tool**: `Type` / `input_type`
- **Arguments**: `text: string`, `x: int`, `y: int`, `clear?: bool`
- **Return value**: Type action confirmation
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/adb/adb_test.go`, `e2e_test.py`

### Skill: Element Wait
- **Description**: Poll UI hierarchy until matching element appears or timeout expires.
- **MCP tool**: `WaitForElement` / `ui_wait_element`
- **Arguments**: `text?`, `resourceId?`, `className?`, `description?`, `timeout?`
- **Return value**: Found element metadata or timeout error
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 5.0+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/ui/ui_test.go`, `e2e_test.py`

---

## 3. Screen & Vision

### Skill: UI Layout Snapshot
- **Description**: Dump XML hierarchy, parse interactive nodes, and return clean text alignment table.
- **MCP tool**: `Snapshot` / `ui_snapshot`
- **Arguments**: `use_vision: false`, `use_annotation: true`
- **Return value**: Aligned UI tree text block
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 5.0+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/ui/ui_test.go`, `e2e_test.py`

### Skill: Visual Annotated Screenshot
- **Description**: Capture PNG screencap, render bounding box rectangles and indexed badge labels over interactive elements.
- **MCP tool**: `Snapshot` / `ui_snapshot`
- **Arguments**: `use_vision: true`, `use_annotation: true`
- **Return value**: Array containing UI text table + Base64 PNG image
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 5.0+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/ui/ui_test.go`, `e2e_test.py`

---

## 4. Application Management

### Skill: List Installed Packages
- **Description**: List installed applications on device with filtering options.
- **MCP tool**: `list_apps`
- **Arguments**: `third_party_only?: bool`
- **Return value**: Text list of package names
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/service/service_test.go`, `e2e_test.py`

### Skill: Launch Application
- **Description**: Launch package or activity via ADB `monkey` or `am start`.
- **MCP tool**: `launch_app`
- **Arguments**: `package_name: string`
- **Return value**: Launch confirmation string
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/service/service_test.go`, `e2e_test.py`

### Skill: Stop Application
- **Description**: Force-stop application package via `am force-stop`.
- **MCP tool**: `stop_app`
- **Arguments**: `package_name: string`
- **Return value**: Stop confirmation string
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/service/service_test.go`, `e2e_test.py`

---

## 5. Filesystem Operations

### Skill: Push File
- **Description**: Transfer host file to Android filesystem target path.
- **MCP tool**: `file_push`
- **Arguments**: `local_path: string`, `remote_path: string`
- **Return value**: Transfer status string
- **Requires device**: Yes
- **Requires root**: No (for `/sdcard` or `/data/local/tmp`)
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/service/service_test.go`, `e2e_test.py`

### Skill: Pull File
- **Description**: Transfer Android file to host machine target path.
- **MCP tool**: `file_pull`
- **Arguments**: `remote_path: string`, `local_path: string`
- **Return value**: Transfer status string
- **Requires device**: Yes
- **Requires root**: No (readable locations)
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/service/service_test.go`, `e2e_test.py`

---

## 6. Shell & System Operations

### Skill: Execute Android Shell Command
- **Description**: Safely run structured shell command on Android target with context timeout.
- **MCP tool**: `shell_exec`
- **Arguments**: `command: string`, `timeout_seconds?: int`
- **Return value**: JSON struct `{ stdout, stderr, exit_code, duration_ms }`
- **Requires device**: Yes
- **Requires root**: No
- **Supported Android versions**: Android 4.4+
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/service/service_test.go`, `e2e_test.py`

---

## 7. Diagnostics & System Engineering

### Skill: Health Doctor
- **Description**: Complete diagnostic health check for ADB, Platform-Tools, managed scrcpy, config files, devices, notification backends, and MCP server.
- **MCP tool**: CLI command `android-mcp doctor`
- **Arguments**: None
- **Return value**: Formatted health report text
- **Requires device**: No
- **Requires root**: No
- **Supported Android versions**: N/A
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/doctor/doctor_test.go`

### Skill: System Status
- **Description**: Concise 1-line or 5-line operational status check with exit codes.
- **MCP tool**: CLI command `android-mcp status`
- **Arguments**: None
- **Return value**: Status text block
- **Requires device**: No
- **Requires root**: No
- **Supported Android versions**: N/A
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/doctor/doctor_test.go`

---

## 8. Managed Dependencies & Live Display Mirroring

### Skill: Managed Platform-Tools Management
- **Description**: Self-contained download, extraction, Zip Slip protection, and atomic installation of official Google Android SDK Platform-Tools (`adb`, `fastboot`) under `~/.android-mcp/platform-tools/`.
- **MCP tool**: CLI subcommand `android-mcp platform-tools status|update|reinstall`
- **Arguments**: Action subcommand
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/platformtools/platformtools_test.go`

### Skill: Managed scrcpy & Automatic Live Mirroring
- **Description**: Dynamic GitHub Release resolution, SHA-256 verification, Tar/Zip Slip protection, atomic installation under `~/.android-mcp/scrcpy/`, and non-blocking auto-launch of `scrcpy` live screen mirror window upon device connection.
- **MCP tool**: CLI subcommand `android-mcp scrcpy status|update|reinstall|start|stop`
- **Arguments**: Action subcommand
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/scrcpy/scrcpy_test.go`

### Skill: Unified State & One-Time Migration
- **Description**: Unified state storage under `~/.android-mcp/android-mcp.json` with atomic temporary file writes. Automatically imports legacy `~/.scrcpy/scrcpy.json` parameters into `android-mcp.json` on first load.
- **MCP tool**: Internal `config` package
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/config/config_test.go`

### Skill: Debug Activity Desktop Notifications
- **Description**: Throttled desktop alerts for AI-agent actions (`--debug`) with rate-limiting anti-spam queues, action correlation IDs, and secret redaction.
- **MCP tool**: CLI flag `--debug` / `internal/notification`
- **Status**: `PHYSICALLY VERIFIED`
- **Tests**: `internal/notification/activity_test.go`
