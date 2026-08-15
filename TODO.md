# Android-MCP-go TODO Checklist

## Self-Contained Platform-Tools Management
- [x] Implement `internal/platformtools` package.
- [x] Managed directory under `~/.android-mcp/platform-tools/`.
- [x] Official Android/Google URLs only (`dl.google.com/android/repository/platform-tools-latest-*.zip`).
- [x] OS & Architecture detection (`darwin`, `linux`, `windows`).
- [x] Zip Slip security validation during extraction.
- [x] Atomic directory installation via `platform-tools.download/`.
- [x] Installation & progress desktop notifications.
- [x] Automatic fallback and resolution hierarchy (`ANDROID_MCP_ADB` > managed path > system ADB).
- [x] CLI `android-mcp platform-tools status|update|reinstall`.

## Debug Activity Notification System
- [x] Throttled async activity notifier (`ActivityNotifier`).
- [x] `--debug` flag triggers desktop activity notifications for AI actions.
- [x] Redaction of sensitive parameters (passwords, tokens, secrets).
- [x] Rate-limiting queue (default 250ms interval) to prevent desktop notification spam.
- [x] Unique action IDs (`ACTION 8f4c2d`) linking desktop alerts to debug logs.

## Reliability & Performance
- [x] Offline execution when Platform-Tools are already installed.
- [x] Graceful error message when missing and offline.
- [x] Zero race conditions (`go test -race ./...` passing 100%).
- [x] Physical verification on Sony Xperia (`SOG09`).

## Documentation
- [x] `docs/PLATFORM_TOOLS.md`
- [x] `docs/NOTIFICATIONS.md`
- [x] `README.md`, `SECURITY.md`, `CONFIGURATION.md`, `DEVELOPMENT.md`, `SKILLS.md`, `ROADMAP.md`, `CHANGELOG.md`.
