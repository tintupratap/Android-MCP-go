# Unified State Schema & Configuration Migration

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Overview

`Android-MCP-go` is a 100% independent server and maintains its persistent state strictly under:
```text
~/.android-mcp/android-mcp.json
```

All external runtime dependencies on `~/.scrcpy/scrcpy.json` and the old `scrcpy-wireless-go` project have been completely eliminated.

---

## Unified JSON Schema (`~/.android-mcp/android-mcp.json`)

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
    "camera_id": "0",
    "camera_size": "default",
    "camera_fps": "default",
    "camera_high_speed": false,
    "video_codec": "h265",
    "video_encoder": "OMX.qcom.video.encoder.hevc",
    "video_bitrate": "4M",
    "max_resolution_size": 0,
    "audio_source": "playback",
    "audio_codec": "opus",
    "audio_encoder": "c2.android.opus.encoder",
    "audio_bitrate": "128K",
    "turn_screen_off": false,
    "stay_awake": true,
    "power_off_on_close": false,
    "render_driver": "metal"
  },
  "platform_tools": {
    "managed": true,
    "path": "~/.android-mcp/platform-tools",
    "version": "1.0.41",
    "source": "official-google",
    "installed_at": "2026-08-15T01:00:00Z"
  },
  "managed_scrcpy": {
    "managed": true,
    "path": "~/.android-mcp/scrcpy",
    "version": "v4.1",
    "release": "v4.1",
    "source": "https://github.com/Genymobile/scrcpy",
    "installed_at": "2026-08-15T01:00:00Z"
  },
  "notifications": {
    "enabled": true
  },
  "migration": {
    "scrcpy_wireless_imported": true,
    "imported_at": "2026-08-15T01:00:00Z"
  }
}
```

---

## One-Time Migration Logic

When `Android-MCP-go` starts:
1. It checks if `~/.scrcpy/scrcpy.json` exists on the system.
2. If present and `migration.scrcpy_wireless_imported` is `false` or missing, it reads the old values (`last_ip`, `device_serial`, `device_model`, `port`, `video_codec`, `video_bitrate`, `display_id`, `audio_source`, `stay_awake`, `render_driver`) and merges them into `~/.android-mcp/android-mcp.json`.
3. It sets `migration.scrcpy_wireless_imported = true` and writes `~/.android-mcp/android-mcp.json` atomically.
4. After migration, `Android-MCP-go` **never** reads or writes `~/.scrcpy/scrcpy.json` again.

---

## Device Connection Precedence

When resolving an Android device:
1. **Explicit CLI / Environment**: `--device` or `ANDROID_MCP_DEVICE`.
2. **`android-mcp.json` Cached State**:
   - If `device.last_ip` exists $\to$ Prefer WiFi ADB (`<last_ip>:5555`).
   - If `device.serial` exists $\to$ Prefer USB device serial.
3. **Auto-Pick ADB Device**: Enumerates online devices via `adb devices`.
4. **USB Wireless Bootstrap**: If USB device is connected, automatically triggers WiFi ADB TCP/IP bootstrap.
