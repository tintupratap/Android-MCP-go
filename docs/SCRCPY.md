# Managed scrcpy & Live Screen Mirroring System

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Overview

`Android-MCP-go` includes built-in managed `scrcpy` display mirroring integration so that users can continuously observe their physical or emulated Android device in a real-time GUI window while AI agents operate the device via MCP tools.

Managed `scrcpy` binaries are installed under:
```text
~/.android-mcp/scrcpy/
```

---

## Official Source Policy

All `scrcpy` releases originate exclusively from the official upstream project:
```text
https://github.com/Genymobile/scrcpy
```

`Android-MCP-go` fetches releases dynamically using the official GitHub Releases API (`https://api.github.com/repos/Genymobile/scrcpy/releases/latest`), rejecting third-party binary hosts, unofficial mirrors, or arbitrary user URLs.

---

## Automatic Live Mirroring Workflow

```text
USB Connected
     │
     ▼
Detect Android Device
     │
     ▼
Discover WiFi IP & Enable ADB TCP/IP (port 5555)
     │
     ▼
Connect WiFi ADB & Verify Transport
     │
     ▼
Launch scrcpy Window targeting exact device endpoint:
scrcpy -s 192.168.1.3:5555 --window-title "Android-MCP — SOG09 (192.168.1.3:5555)"
     │
     ▼
USB Disconnected -> WiFi & Live Display Window Continue Seamlessly!
```

---

## Integration with `~/.scrcpy/scrcpy.json`

`Android-MCP-go` reuses existing `scrcpy.json` preferences if present without overwriting them:
- `video_codec` (e.g. `h265`, `h264`, `av1`)
- `video_bitrate` (e.g. `4M`, `8M`)
- `display_id` (e.g. `0`)
- `audio_source` (e.g. `playback`)
- `stay_awake` (`true`)

---

## Security & Archive Controls

1. **Zip Slip Protection**: Zip entries are strictly validated during extraction to ensure files cannot traverse outside target destination directories.
2. **Atomic Installation**: Archive extracts into temporary directory `~/.android-mcp/scrcpy.download/`, verifies `scrcpy --version` binary execution, and atomically swaps into place.
3. **Privacy Protections**: Screenshots and notifications sent to desktop displays redact passwords, tokens, and secrets.

---

## CLI Commands

```bash
# Check scrcpy status
android-mcp scrcpy status

# Download / update scrcpy from official GitHub Releases
android-mcp scrcpy update

# Manually launch screen mirror window for connected device
android-mcp scrcpy start

# Stop active scrcpy screen mirror windows
android-mcp scrcpy stop
```

---

## Configuration & Environment Variables

| Variable | Description | Default |
|---|---|---|
| `ANDROID_MCP_SCRCPY` | Enable/disable auto-launching screen mirror (`true`/`false`) | `true` |
| `ANDROID_MCP_SCRCPY_PATH` | Explicit path to `scrcpy` executable | Auto-detected |
| `ANDROID_MCP_SCRCPY_ARGS` | Extra arguments to pass to `scrcpy` | `""` |
