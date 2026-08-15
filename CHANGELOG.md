# Android-MCP-go Changelog

## [0.2.0] - 2026-08-15

### Added
- **Skills System & Manifest**:
  - `SKILLS.md`: Comprehensive capability map detailing status, arguments, and requirements for all Android-MCP skills.
  - `skills/`: Machine-readable capability domain manifests (`manifest.json`, `device.json`, `ui.json`, `screenshot.json`, `input.json`, `applications.json`, `filesystem.json`, `shell.json`, `automation.json`).
- **Service Layer Architecture (`internal/service`)**:
  - Clean separation: `Android Capability -> Go Service -> MCP Adapter -> MCP Tool`.
  - `DeviceService`, `InputService`, `ScreenService`, `UIService`, `AppService`, `FileService`, `ShellService`.
- **System Engineering & Health Diagnostics**:
  - `android-mcp doctor`: Detailed diagnostic CLI report checking ADB version, server status, configuration files, device connections, notification backends, and MCP tool count.
  - `android-mcp status`: Concise 5-line status report with exit codes (0 for healthy/ready, 1 for disconnected/error).
- **Expanded MCP Capabilities & Tools**:
  - `list_apps`: List installed application packages with third-party filtering.
  - `launch_app`: Launch application package on device via `monkey`/`am start`.
  - `stop_app`: Force-stop application package via `am force-stop`.
  - `file_push`: Transfer host file to Android storage.
  - `file_pull`: Transfer Android file to host machine.
  - `shell_exec`: Execute structured Android shell commands with context timeout, returning `{ stdout, stderr, exit_code, duration_ms }`.
- **Security & Race Hardening**:
  - `SECURITY.md`: Trust models, argument slice execution isolation (no `sh -c`), and host/device execution boundary separation.
  - Passed `go test -race ./...` with zero data races across all packages.
- **Performance Benchmarks (`bench/`)**:
  - Micro-benchmarks for XML UI parsing (~49µs/op), selector searching (~49µs/op), tabular formatting (~2.6µs/op), ADB parsing (~1.2µs/op), and PNG screenshot visual annotation (~12.8ms/op).

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
