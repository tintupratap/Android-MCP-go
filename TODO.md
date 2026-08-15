# Android-MCP-go TODO Checklist

## 1. Managed `scrcpy` & Live Screen Mirroring System
- [x] Implement `internal/scrcpy` package.
- [x] Managed directory under `~/.android-mcp/scrcpy/`.
- [x] Dynamic GitHub Release resolution (`api.github.com/repos/Genymobile/scrcpy/releases/latest`).
- [x] Host OS & architecture asset resolver (`darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`, `windows/386`).
- [x] SHA-256 checksum verification against official published checksum assets.
- [x] Tar.gz and Zip archive extraction with Zip/Tar Slip path traversal protection.
- [x] Atomic directory installation via `~/.android-mcp/.downloads/` and `~/.android-mcp/.staging/scrcpy/`.
- [x] Non-blocking automatic launch of `scrcpy` live display mirror window upon device connection.
- [x] Duplicate window protection per device serial.
- [x] CLI `android-mcp scrcpy status|update|reinstall|start|stop`.

## 2. Unified Configuration & Independence
- [x] Unified JSON schema in `~/.android-mcp/android-mcp.json`.
- [x] One-time migration (`PerformOneTimeMigration`) importing legacy `~/.scrcpy/scrcpy.json` parameters.
- [x] Total elimination of runtime dependency on `~/.scrcpy/scrcpy.json` and `scrcpy-wireless-go`.
- [x] Atomic write persistence using temporary file replacement.

## 3. Self-Contained Platform-Tools Management
- [x] Implement `internal/platformtools` package.
- [x] Managed directory under `~/.android-mcp/platform-tools/`.
- [x] Official Android/Google URLs only (`dl.google.com/android/repository/platform-tools-latest-*.zip`).
- [x] Zip Slip security validation during extraction.
- [x] Atomic directory installation via `platform-tools.download/`.
- [x] CLI `android-mcp platform-tools status|update|reinstall`.

## 4. Debug Activity Notification System
- [x] Throttled async activity notifier (`ActivityNotifier`).
- [x] `--debug` flag triggers desktop activity notifications for AI actions.
- [x] Automatic redaction of sensitive parameters (passwords, tokens, secrets).
- [x] Rate-limiting queue (default 250ms interval) to prevent desktop notification spam.
- [x] Unique action correlation IDs (`ACTION 8f4c2d`).

## 5. Verification & Testing
- [x] 100% pass rate on `go test ./...`.
- [x] 100% pass rate on race detector `go test -race ./...` across all 14 packages.
- [x] Physical hardware verification on Sony Xperia SOG09 across all 23 registered MCP tools via `python3 e2e_test.py`.

## 6. GitHub Publication & Documentation
- [x] `docs/PLATFORM_TOOLS.md`
- [x] `docs/NOTIFICATIONS.md`
- [x] `docs/SCRCPY.md`
- [x] `docs/CONFIGURATION_MIGRATION.md`
- [x] `docs/RELEASE_MANAGEMENT.md`
- [x] `README.md`, `SKILLS.md`, `ARCHITECTURE.md`, `CONFIGURATION.md`, `SECURITY.md`, `DEVELOPMENT.md`, `ROADMAP.md`, `CHANGELOG.md`.
