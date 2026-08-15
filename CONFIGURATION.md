# Android-MCP-go Configuration Guide

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Configuration Storage Location

`Android-MCP-go` stores all state and preferences strictly under:
```text
~/.android-mcp/android-mcp.json
```

All runtime dependencies on `~/.scrcpy/scrcpy.json` and external tools have been removed.

---

## Unified Schema Reference (`android-mcp.json`)

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
    "video_source": "display",
    "display_id": "0",
    "video_codec": "h265",
    "video_bitrate": "4M",
    "audio_source": "playback",
    "audio_codec": "opus",
    "audio_bitrate": "128K",
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
  },
  "migration": {
    "scrcpy_wireless_imported": true
  }
}
```

### Key Sections

- **`device`**: Most recently verified Android endpoint parameters (`last_ip`, `serial`, `model`, `port`, `connection`).
- **`scrcpy`**: User video, audio, rendering driver, and display mirroring preferences.
- **`platform_tools`**: Installation metadata for managed Google Android SDK Platform-Tools.
- **`managed_scrcpy`**: Release metadata for managed Genymobile `scrcpy` binary installation.
- **`notifications`**: Desktop notification settings.
- **`migration`**: Tracks one-time migration status for importing legacy `~/.scrcpy/scrcpy.json` values.

---

## Precedence Hierarchy

```text
CLI Arguments (--device, --connection)
       ↓
Environment Variables (ANDROID_MCP_DEVICE, ANDROID_MCP_CONNECTION, ANDROID_MCP_HOST)
       ↓
Unified Persistent State (~/.android-mcp/android-mcp.json)
       ↓
Automatic Device Discovery (Physical USB -> WiFi Bootstrap -> Physical WiFi -> Emulator)
```

---

## Atomic File Persistence

All configuration file writes use **atomic temporary file replacement**:
1. Marshal JSON in memory.
2. Write content to temporary file `~/.android-mcp/android-mcp.json.tmp`.
3. Atomically replace destination via `os.Rename`.

This guarantees that power interrupts or unexpected terminations never leave a corrupt zero-byte configuration file.
