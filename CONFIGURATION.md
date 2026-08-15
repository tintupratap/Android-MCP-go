# Android-MCP-go Configuration Guide

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Configuration File Locations

- **Primary State**: `~/.android-mcp/android-mcp.json`
- **External Discovery Source**: `~/.scrcpy/scrcpy.json`

---

## Schema Reference (`android-mcp.json`)

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

### Fields

- `last_ip` (*string*): Most recently verified WiFi IPv4 address.
- `device_serial` (*string*): Physical USB device serial number.
- `device_model` (*string*): Android product model (e.g. `SOG09`, `Pixel 7`).
- `port` (*int*): ADB TCP port (default `5555`).
- `connection` (*string*): Connection mode (`wifi`, `usb`, `auto`).
- `last_seen` (*timestamp*): ISO 8601 timestamp of last resolution.
- `last_successful_connection` (*timestamp*): ISO 8601 timestamp of last verified ADB communication.

---

## Config Precedence Hierarchy

```text
CLI Arguments (--device, --wifi, --usb, --connection)
       ↓
Environment Variables (ANDROID_MCP_DEVICE, ANDROID_MCP_CONNECTION, ANDROID_MCP_HOST)
       ↓
Persistent State (~/.android-mcp/android-mcp.json)
       ↓
External State (~/.scrcpy/scrcpy.json)
       ↓
Automatic Device Discovery (Physical USB -> WiFi Bootstrap -> Physical WiFi -> Emulator)
```

---

## Error Handling & Recovery

- If `~/.android-mcp/android-mcp.json` contains malformed JSON, a backup copy is saved as `~/.android-mcp/android-mcp.json.bak`, logging a warning, and continuing with default settings.
- File updates are written atomically (write to temporary file, `fsync`, and atomic `os.Rename`), avoiding partial writes or file corruption during power failures.
