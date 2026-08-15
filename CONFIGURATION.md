# Android-MCP-go Configuration Guide

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

`Android-MCP-go` maintains unified state, device history, and mirroring preferences in a single JSON file:

```text
~/.android-mcp/android-mcp.json
```

(Configurable via `ANDROID_MCP_HOME`).

---

## 1. Unified JSON Schema Reference

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

---

## 2. Configuration Parameters

| Section | Field | Description | Default |
|---|---|---|---|
| `device` | `last_ip` | Last verified WiFi IP address | Auto-discovered |
| `device` | `serial` | Last connected device serial | Auto-discovered |
| `device` | `connection` | Preferred connection type (`wifi`, `usb`, `auto`) | `auto` |
| `scrcpy` | `enabled` | Enable managed scrcpy support | `true` |
| `scrcpy` | `auto_start` | Automatically launch display window on device connection | `true` |
| `scrcpy` | `video_codec` | Video encoding codec (`h264`, `h265`, `av1`) | `h265` |
| `scrcpy` | `video_bitrate` | Stream bit rate | `4M` |
| `scrcpy` | `stay_awake` | Keep device screen awake while mirroring | `true` |

---

## 3. Precedence Hierarchy

1. **Explicit CLI Flags**: `--device 192.168.1.3:5555 --connection wifi`
2. **Environment Variables**: `ANDROID_MCP_DEVICE`, `ANDROID_MCP_HOST`
3. **Persistent State**: `~/.android-mcp/android-mcp.json` (`device.last_ip`)
4. **Auto-Pick Discovery**: First available active ADB device.

---

## 4. Atomic Persistence & One-Time Migration

- **Atomic Writes**: Saved using temporary file writes (`android-mcp.json.tmp.*`) followed by atomic rename to prevent corruption.
- **One-Time Migration**: On first boot, if legacy `~/.scrcpy/scrcpy.json` is detected, preferences are automatically imported into `~/.android-mcp/android-mcp.json`.
