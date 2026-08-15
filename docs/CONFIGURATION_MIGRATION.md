# Unified State Schema & Migration Guide

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

`Android-MCP-go` manages all state, device history, mirroring preferences, and managed component metadata inside:

```text
~/.android-mcp/android-mcp.json
```

All runtime dependencies on `~/.scrcpy` and `scrcpy-wireless-go` have been removed.

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
    "video_codec": "h265",
    "video_bitrate": "4M",
    "audio_source": "playback",
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
  "migration": {
    "scrcpy_wireless_imported": true
  }
}
```

---

## 2. Automatic Migration Engine

- **Trigger**: Upon first load of `android-mcp`, if `~/.scrcpy/scrcpy.json` exists and `"scrcpy_wireless_imported"` is false, `PerformOneTimeMigration` reads existing preferences and imports them into `~/.android-mcp/android-mcp.json`.
- **Atomic Operations**: All state writes flush to `android-mcp.json.tmp.*` files before atomic rename, preventing data loss.
