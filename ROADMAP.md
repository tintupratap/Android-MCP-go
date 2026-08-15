# Android-MCP-go Roadmap

## Phase 1 — Analysis & Architecture ✅
- [x] Archaeology of Android-MCP-py and scrcpy-wireless-go.
- [x] Porting Analysis (`docs/PORTING_ANALYSIS.md`).
- [x] Architecture Specification (`ARCHITECTURE.md`).

## Phase 2 — Core Foundations & Packages ✅
- [x] Go module initialization (`go.mod`).
- [x] `internal/config`: Persistent state schema (`android-mcp.json`) & atomic file operations + `scrcpy.json` reader.
- [x] `internal/adb`: Safe ADB command runner (`exec.CommandContext`).
- [x] `internal/notification`: Platform-aware notifications (`terminal-notifier` / `notify-send`).
- [x] `internal/discovery`: Multi-strategy IP discovery & TCP/IP bootstrap engine.
- [x] `internal/device`: `DeviceManager` & state machine.

## Phase 3 — UI & Interaction Engine ✅
- [x] `internal/ui`: Android UI hierarchy XML parser (`encoding/xml`), element selector matching, bounding box math, tabular formatter, visual screenshot annotation engine (`image/draw`).

## Phase 4 — MCP Protocol & Server ✅
- [x] `internal/mcp`: JSON-RPC 2.0 stdio server protocol implementation.
- [x] Port all 14 MCP tools (`ListDevices`, `ConnectDevice`, `Device`, `Click`, `ClickBySelector`, `Snapshot`, `LongClick`, `Swipe`, `Type`, `Drag`, `Press`, `Notification`, `Wait`, `WaitForElement`).
- [x] CLI flags and environment variables integration.
- [x] Lazy device resolution (server boots cleanly without connected devices).

## Phase 5 — Testing & Quality Assurance ✅
- [x] Unit tests for config, adb parsing, IP discovery, state machine, XML UI parser, notification, and MCP tool serialization (100% passing across all packages).
- [x] Physical device verification against connected Xperia (`SOG09`) phone (`192.168.1.3:5555`) covering JSON-RPC stdio protocol, `ListDevices`, text layout `Snapshot`, and annotated visual image `Snapshot`.

## Phase 6 — Documentation & Release ✅
- [x] `README.md`, `ARCHITECTURE.md`, `CONFIGURATION.md`, `DEVELOPMENT.md`, `ROADMAP.md`, `TODO.md`, `docs/PORTING_ANALYSIS.md`.
- [x] Clean compilation and end-to-end verification.
