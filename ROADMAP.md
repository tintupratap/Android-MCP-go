# Android-MCP-go Milestone Roadmap

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

---

## Completed Milestones

### Phase 1 — Analysis & Architecture ✅
- [x] Archaeology of Android-MCP-py and scrcpy-wireless-go.
- [x] Porting analysis (`docs/PORTING_ANALYSIS.md`) and architecture specification (`ARCHITECTURE.md`).

### Phase 2 — Core Foundations & Data Safety ✅
- [x] Portable path resolver (`RuntimePaths`) supporting `ANDROID_MCP_HOME`.
- [x] Atomic JSON persistence (`~/.android-mcp/android-mcp.json`).
- [x] Direct slice ADB command execution engine (`internal/adb`).
- [x] Wireless multi-strategy IP discovery & TCP/IP bootstrap engine (`internal/discovery`).

### Phase 3 — UI Automation & Vision Engine ✅
- [x] XML UI hierarchy parser, selector matcher, and bounding box math.
- [x] Visually annotated PNG screenshot generator with index badges (`internal/ui`).

### Phase 4 — MCP Stdio Protocol & 23 Registered Tools ✅
- [x] JSON-RPC 2.0 stdio server implementation (`internal/mcp`).
- [x] Registered all 23 MCP tools and aliases.
- [x] Lazy device resolution (instant server boot without connected devices).

### Phase 5 — Self-Contained Platform-Tools Management ✅
- [x] Managed Google Platform-Tools installer under `~/.android-mcp/platform-tools/`.
- [x] Zip Slip security checks and atomic installation.

### Phase 6 — Debug Activity Notification Engine ✅
- [x] Rate-limited desktop alert queue with action correlation IDs and credential redaction (`--debug`).

### Phase 7 — Managed `scrcpy` & Live Screen Mirroring ✅
- [x] Dynamic GitHub Release asset resolver (`Genymobile/scrcpy`).
- [x] SHA-256 integrity verification, Tar/Zip Slip protection, and atomic installation under `~/.android-mcp/scrcpy/`.
- [x] Non-blocking `scrcpy` display window manager with duplicate window protection and clean SIGTERM exit handling.

### Phase 8 — Unified State & Complete Project Independence ✅
- [x] Single state store `~/.android-mcp/android-mcp.json`.
- [x] Automated one-time migration for legacy `~/.scrcpy/scrcpy.json`.
- [x] Complete removal of runtime dependency on external projects.

### Phase 9 — 100% Self-Contained Runtime & Skills Integration ✅
- [x] Zero runtime dependency on `ANDROID_HOME`, `ANDROID_SDK_ROOT`, system `adb`, or system `scrcpy`.
- [x] Automated installation of 10 machine-readable skill domain manifests under `~/.android-mcp/skills/`.
- [x] CLI subcommands (`doctor`, `status`, `platform-tools`, `scrcpy`, `skills`).

### P0.6 — Adaptive Platform-Optimized Live View ✅
- [x] Capability probing engine (`DetectBinaryCapabilities`, `DetectHostCapabilities`, `DetectDeviceCapabilities`).
- [x] Progressive degradation fallback pipeline (`Optimized` $\to$ `Reduced` $\to$ `H.264 Fallback` $\to$ `No Audio` $\to$ `Minimal Safe`).
- [x] CLI diagnostic subcommands (`android-mcp scrcpy capabilities`, `android-mcp scrcpy profile`).
- [x] Hardware acceleration defaults (Metal renderer on macOS, WiFi-conscious 4M bitrate vs USB 8M bitrate).
- [x] Published `docs/SCRCPY_OPTIMIZATION.md`.

### P0.7 — Automatic Live View Startup ✅
- [x] Non-blocking `LiveViewManager` background thread starting automatically upon server boot (`StartBackground`).
- [x] Zero MCP tool call dependency to launch `scrcpy` window.
- [x] Auto-reconnect & crash recovery state machine (`WAITING_FOR_DEVICE`, `PREPARING_SCRCPY`, `STARTING_SCRCPY`, `RUNNING`).
- [x] Added `--no-scrcpy` CLI override flag.

### P0.8 — Persistent Visual Observability ✅
- [x] Tool-call observer guard (`EnsureLiveView`) restoring `scrcpy` BEFORE executing device actions (`Click`, `Swipe`, `Type`, `Snapshot`, `Press`, etc.).
- [x] Manual close state tracking (`CLOSED_BY_USER`) preventing background restart storms.
- [x] Singleflight mutex protection for concurrent tool call relaunches.
- [x] Added `--no-scrcpy-relaunch` CLI override flag and `auto_relaunch_on_tool_call` configuration.

### P0.9 — Single-Instance Live View Enforcement ✅
- [x] Eliminated all duplicate launch paths across `DeviceManager` and `MCP` server.
- [x] Monotonic generation tracking (`generation uint64`) preventing stale process exit callbacks from modifying active state.
- [x] Idempotent `EnsureRunning` startup gate using `sync.Cond` and double-checking under lock.
- [x] OS-level process file lock (`~/.android-mcp/scrcpy.lock`) preventing multiple server instances from spawning duplicate windows.
- [x] 100% unit & data race test pass rate (`go test -race ./...`).
- [x] Physical hardware E2E verification on Sony Xperia `SOG09` (`192.168.1.3:5555`).

---

## Future Enhancement Goals (Post-v0.4.0)

- [ ] WebRTC / Browser-based live device stream server option.
- [ ] Multi-device concurrent UI automation orchestration.
