# Android-MCP-go Changelog

## [0.4.0] - 2026-08-15

### Added
- **Managed `scrcpy` & Automatic Live Screen Mirroring (`internal/scrcpy`)**:
  - Automatically manages `scrcpy` under `~/.android-mcp/scrcpy/` from official GitHub Releases (`https://github.com/Genymobile/scrcpy/releases`).
  - Automatically launches `scrcpy` GUI display window targeting the connected Android device serial (`scrcpy -s <serial> --window-title "Android-MCP — <model> (<serial>)"`) upon device connection (USB or WiFi ADB).
  - Integrates with `~/.scrcpy/scrcpy.json` preferences (`video_codec`, `video_bitrate`, `display_id`, `audio_source`, `stay_awake`, `render_driver`).
  - Duplicate window protection: checks running process per serial before launching to prevent multiple windows.
  - Non-blocking background launch ensures server startup and MCP tool calls are never blocked.
  - Subcommands: `android-mcp scrcpy status|update|start|stop|restart`.
- **Documentation**:
  - Added `docs/SCRCPY.md` detailing managed scrcpy architecture, workflow, safety features, and CLI commands.

---

## [0.3.0] - 2026-08-15

### Added
- **Self-Contained Platform-Tools Management (`internal/platformtools`)**:
  - Automatically manages ADB and fastboot under `~/.android-mcp/platform-tools/`.
  - Downloads exclusively from official Google HTTPS endpoints (`dl.google.com/android/repository/platform-tools-latest-*.zip`).
  - Implements Zip Slip security validation during extraction.
  - Performs atomic directory replacement via `platform-tools.download/`.
  - CLI management commands: `android-mcp platform-tools status|update|reinstall`.
- **Debug Activity Notification System (`--debug`)**:
  - Real-time desktop activity notifications for AI actions when `--debug` is active (e.g., `AI: Clicked "Login"`, `AI: Launched com.example.app`).
  - Rate-limited async notification queue (`ANDROID_MCP_DEBUG_NOTIFY_INTERVAL`, default 250ms) to prevent notification spam.
  - Automatic redaction of sensitive parameters (passwords, tokens, secrets).
  - Action correlation IDs (`ACTION 8f4c2d`) linking desktop alerts to structured debug logs.
- **Documentation**:
  - `docs/PLATFORM_TOOLS.md`: Detailed platform-tools architecture and security model.
  - `docs/NOTIFICATIONS.md`: Desktop notification hierarchy and debug activity system.

---

## [0.2.0] - 2026-08-15

### Added
- **Skills System & Manifest**:
  - `SKILLS.md`: Comprehensive capability map detailing status, arguments, and requirements.
  - `skills/`: Machine-readable capability domain manifests (`manifest.json`, `device.json`, `ui.json`, `screenshot.json`, `input.json`, `applications.json`, `filesystem.json`, `shell.json`, `automation.json`).
- **Service Layer Architecture (`internal/service`)**:
  - Clean separation: `Android Capability -> Go Service -> MCP Adapter -> MCP Tool`.
- **System Engineering & Health Diagnostics**:
  - `android-mcp doctor`: Diagnostic CLI report checking ADB version, config files, device connections, notification backends, and tool count.
  - `android-mcp status`: Concise status report returning exit code 0 when ready/connected.
- **Expanded MCP Capabilities**: `list_apps`, `launch_app`, `stop_app`, `file_push`, `file_pull`, `shell_exec`.

---

## [0.1.0] - 2026-08-15

### Added
- Initial native Go port of Android-MCP-py.
- Stdio JSON-RPC 2.0 MCP protocol transport.
- Automatic USB → WiFi ADB wireless bootstrap engine.
- Persistent state management in `~/.android-mcp/android-mcp.json` with atomic writes.
- External discovery integration with `~/.scrcpy/scrcpy.json`.
- Desktop notifications (`terminal-notifier` on macOS, `notify-send` on Linux).
- Lazy device resolution allowing MCP server boot without connected hardware.
- Physical hardware verification on Sony Xperia (`SOG09`).
