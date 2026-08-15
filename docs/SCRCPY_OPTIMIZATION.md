# Adaptive scrcpy Optimization & Degradation Architecture

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

`Android-MCP-go` features an **adaptive, capability-aware `scrcpy` launch engine** with a **strict single-instance invariant** (max 1 window) and **always-on-top display mode** (`--always-on-top`). Rather than using hardcoded CLI arguments, it dynamically probes host, device, and `scrcpy` capabilities to select the optimal video codec, renderer, bitrate, and audio configuration—falling back gracefully if optional features fail.

---

## 1. Profile Resolution Pipeline

```text
Detect scrcpy Binary Flags (`scrcpy --help`)
                     │
                     ▼
Detect Host OS & Native Graphics (`darwin` -> Metal, `windows` -> Direct3D)
                     │
                     ▼
Detect Device Capabilities & Network Mode (WiFi ADB 4M vs USB ADB 8M)
                     │
                     ▼
       Resolve Optimal Profile
                     │
                     ▼
              Launch scrcpy
                     │
         Success? ───┼───> [Done] Live Window Active
                     │
                     ▼ (On Failure)
        Progressive Degradation:
   1. Drop --render-driver (Try standard renderer)
   2. Fallback H.265 -> H.264
   3. Drop Audio (--no-audio)
   4. Minimal Safe Profile (--window-title & -s <serial>)
```

---

## 2. Progressive Degradation & Fallback Steps

| Step | Profile Name | Changes Applied | Goal |
|---|---|---|---|
| **0** | `auto` / `optimized` | Full platform acceleration (e.g. H.265, Metal, Audio, WiFi-tuned bitrate) | Maximum Performance |
| **1** | `reduced_optimized` | Drop `--render-driver` parameter | Handle unsupported host renderers |
| **2** | `h264_fallback` | Fallback from `--video-codec h265` to `--video-codec h264` | Handle unsupported hardware encoders |
| **3** | `no_audio_fallback` | Drop `--audio-source` / append `--no-audio` | Handle audio pipeline failures |
| **4** | `minimal_safe` | Minimal argument set (`--window-title`, `-s <serial>`) | Guaranteed baseline mirror |

---

## 3. Configuration & Overrides (`~/.android-mcp/android-mcp.json`)

Users can rely on `auto` defaults or explicitly override settings:

```json
{
  "scrcpy": {
    "enabled": true,
    "auto_start": true,
    "profile": "auto",
    "optimization": "maximum",
    "video_codec": "auto",
    "render_driver": "auto",
    "audio": "auto",
    "video_bitrate": "auto"
  }
}
```

*Note: Explicit non-`auto` user settings (e.g. `video_codec: "h264"`) are strictly respected and never overridden during initial profile resolution.*

---

## 4. CLI Capability & Profile Diagnostics

```bash
# Display detected scrcpy flags, host platform, and device capabilities
android-mcp scrcpy capabilities

# Display resolved optimal profile and exact CLI argument array
android-mcp scrcpy profile

# Check managed scrcpy status
android-mcp scrcpy status
```
