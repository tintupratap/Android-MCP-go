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
- [x] 100% unit & data race test pass rate (`go test -race ./...`).
- [x] Physical hardware E2E verification on Sony Xperia `SOG09` (`192.168.1.3:5555`).

---

## Future Enhancement Goals (Post-v0.4.0)

- [ ] WebRTC / Browser-based live device stream server option.
- [ ] Multi-device concurrent UI automation orchestration.
