# Android-MCP-go

**Android-MCP-go** is a high-performance, single-binary, production-grade Go implementation of the Model Context Protocol (MCP) server for Android operating systems.

It allows AI assistants (Claude Desktop, Cursor, Custom Agents, etc.) to control and inspect Android devices natively over standard stdio JSON-RPC 2.0.

---

## Key Features

- **🚀 Native Go Implementation**: Fast startup, zero Python runtime overhead, light memory footprint, single binary deployment.
- **📡 Automatic Wireless ADB Bootstrap**: Plug in a physical device via USB once; `Android-MCP-go` automatically discovers the device's WiFi IP address, switches ADB to TCP/IP mode (`port 5555`), verifies the wireless link, and persists state. USB can then be unplugged!
- **💾 Persistent State Management**: Remembers connected devices in `~/.android-mcp/android-mcp.json` using atomic file writes to prevent corruption.
- **🔍 scrcpy Integration**: Reads external discovery state from `~/.scrcpy/scrcpy.json` if available.
- **🔔 Desktop Notifications**: Displays OS notifications on successful wireless bootstrap (`terminal-notifier` on macOS, `notify-send` on Linux).
- **💤 Lazy Device Resolution**: The MCP server starts instantly even when no Android device is currently connected. Device resolution occurs when tools are invoked.
- **🖼️ Vision & Bounding Boxes**: Generates UI layout tables and visually annotated PNG screenshots with bounding boxes and indexed badges.
- **🛡️ Physical Device Preference**: Automatically prioritizes real physical hardware over emulators.

---

## Quick Start

### 1. Installation

Build from source:

```bash
git clone https://github.com/android-mcp/android-mcp-go.git
cd android-mcp-go
go build -o android-mcp ./cmd/android-mcp
```

Or install using `go install`:

```bash
go install android-mcp-go/cmd/android-mcp@latest
```

### 2. MCP Server Configuration

Add to your MCP client configuration file (e.g. `claude_desktop_config.json` or Cursor MCP settings):

```json
{
  "mcpServers": {
    "android-mcp": {
      "command": "android-mcp"
    }
  }
}
```

*Note: You do NOT need to hardcode IP addresses in the MCP client config. Dynamic device discovery handles connection automatically.*

---

## Usage & CLI Flags

```bash
# Default mode (uses persistent state ~/.android-mcp/android-mcp.json & auto discovery)
android-mcp

# Connect to explicit USB serial
android-mcp --usb QV771A3JEE

# Connect to explicit WiFi IP / host:port
android-mcp --wifi 192.168.1.3

# Connect to explicit device target
android-mcp --device 192.168.1.3:5555

# Verbose debug logging
android-mcp --debug
```

### Environment Variables

| Variable | Description |
|---|---|
| `ANDROID_MCP_DEVICE` | Explicit device serial or `host:port` |
| `ANDROID_MCP_CONNECTION` | Connection preference: `auto`, `usb`, `wifi` |
| `ANDROID_MCP_HOST` | WiFi IP or host |
| `LOG_LEVEL` | Set to `debug` for verbose logs |

---

## Supported MCP Tools (14 Tools)

1. `ListDevices`: List connected ADB devices.
2. `ConnectDevice`: Connect to device by serial or IP:port.
3. `Device`: Unified device manager (action: list, connect, disconnect).
4. `Click`: Tap screen coordinate (x, y).
5. `ClickBySelector`: Tap element matching text, resourceId, className, or description.
6. `Snapshot`: Get UI hierarchy text table + optional visually annotated screenshot PNG.
7. `LongClick`: Long press screen coordinate (x, y).
8. `Swipe`: Swipe between screen coordinates (x1, y1) to (x2, y2).
9. `Type`: Focus & type text at coordinate (x, y) with optional clear.
10. `Drag`: Drag & drop gesture.
11. `Press`: Press hardware key (home, back, power, volume_up, volume_down, enter).
12. `Notification`: Expand notification shade.
13. `Wait`: Pause execution for N seconds.
14. `WaitForElement`: Polling wait until UI element appears.

---

## Workflow: USB → WiFi Auto-Bootstrap

```text
Connect USB
     │
     ▼
Detect USB Device (Serial/Model)
     │
     ▼
Discover Device IP (wlan/ip route/dhcp)
     │
     ▼
Enable ADB TCP/IP Mode (port 5555)
     │
     ▼
Connect to <IP>:5555
     │
     ▼
Verify Connection (getprop ro.product.model)
     │
     ▼
Persist State to ~/.android-mcp/android-mcp.json
     │
     ▼
Desktop Notification: "USB can now be disconnected"
     │
     ▼
Unplug USB & Continue Over WiFi
```

---

## License

MIT License.
