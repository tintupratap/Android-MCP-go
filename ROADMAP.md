# Android-MCP-go Roadmap

## Phase 1 — Analysis & Architecture ✅
- [x] Archaeology of Android-MCP-py and scrcpy-wireless-go.
- [x] Porting Analysis (`docs/PORTING_ANALYSIS.md`).
- [x] Architecture Specification (`ARCHITECTURE.md`).

## Phase 2 — Core Foundations & Packages ✅
- [x] Go module initialization (`go.mod`).
- [x] `internal/config`: Persistent state schema (`android-mcp.json`) & atomic file operations.
- [x] `internal/adb`: Safe ADB command runner (`exec.CommandContext`).
- [x] `internal/notification`: Platform-aware notifications (`terminal-notifier` / `notify-send`).
- [x] `internal/discovery`: Multi-strategy IP discovery & TCP/IP bootstrap engine.
- [x] `internal/device`: `DeviceManager` & state machine.

## Phase 3 — UI & Interaction Engine ✅
- [x] `internal/ui`: Android UI hierarchy XML parser (`encoding/xml`), element selector matching, bounding box math, tabular formatter, visual screenshot annotation engine (`image/draw`).

## Phase 4 — MCP Protocol & Server ✅
- [x] `internal/mcp`: JSON-RPC 2.0 stdio server protocol implementation.
- [x] Port all 23 MCP tools and aliases (`ListDevices`, `ConnectDevice`, `Device`, `Click`, `ClickBySelector`, `Snapshot`, `LongClick`, `Swipe`, `Type`, `Drag`, `Press`, `Notification`, `Wait`, `WaitForElement`, `list_apps`, `launch_app`, `stop_app`, `file_push`, `file_pull`, `shell_exec`).
- [x] CLI flags and environment variables integration.
- [x] Lazy device resolution (server boots cleanly without connected devices).

## Phase 5 — Self-Contained Platform-Tools Management ✅
- [x] `internal/platformtools`: Automatic download, extraction, Zip Slip protection, and atomic installation of official Google Android SDK Platform-Tools under `~/.android-mcp/platform-tools/`.
- [x] Subcommands `android-mcp platform-tools status|update|reinstall`.

## Phase 6 — Debug Activity Desktop Notification Engine ✅
- [x] `ActivityNotifier`: Real-time desktop alerts for AI-agent actions (`--debug`) with rate-limiting anti-spam queues, action correlation IDs, and secret redaction.

## Phase 7 — Managed scrcpy & Live Display Mirroring ✅
- [x] `internal/scrcpy`: Dynamic GitHub Release resolution (`api.github.com/repos/Genymobile/scrcpy/releases/latest`), host OS/arch asset resolver (`darwin`, `linux`, `windows`), SHA-256 checksum verification, Tar/Zip Slip protection, atomic installation under `~/.android-mcp/scrcpy/`, and non-blocking auto-launch of `scrcpy` live display mirror window upon device connection.
- [x] Subcommands `android-mcp scrcpy status|update|reinstall|start|stop`.

## Phase 8 — Unified Configuration & Independence ✅
- [x] Unified JSON schema in `~/.android-mcp/android-mcp.json`.
- [x] Automatic one-time migration (`PerformOneTimeMigration`) importing legacy `~/.scrcpy/scrcpy.json` parameters.
- [x] Elimination of runtime dependency on `~/.scrcpy/scrcpy.json` and external projects.

## P0.5 — Fully Self-Contained Runtime ✅
- [x] Centralized portable root resolver (`internal/config/paths.go`, `RuntimePaths`).
- [x] Support `ANDROID_MCP_HOME` environment variable for root path override.
- [x] Total elimination of searches for `ANDROID_HOME`, `ANDROID_SDK_ROOT`, system `adb`, system `fastboot`, Homebrew, and system `scrcpy`.
- [x] Enforce `ADB=~/.android-mcp/platform-tools/adb` when launching `scrcpy`.
- [x] Environment isolation unit test (`TestEnvironmentIsolation`) passing 100%.
- [x] Doctor diagnostic report formatted with `Android SDK: Required: no, Used: no`.
- [x] Published `docs/SELF_CONTAINED.md`.

## Phase 9 — Verification & GitHub Release ✅
- [x] Unit tests passing 100% across all packages (`go test ./...`).
- [x] Data race detector passing 100% (`go test -race ./...`).
- [x] Physical device verification against connected Xperia (`SOG09`) phone (`192.168.1.3:5555`) covering all 23 MCP tools and live `scrcpy` screen mirroring.
- [x] Complete documentation suite published.
