# Android-MCP-go System Architecture

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Overview

`Android-MCP-go` is designed as a **decoupled, layered, 100% self-contained Go application**. All binary management, state persistence, discovery, and execution happen strictly under `~/.android-mcp/` (or `$ANDROID_MCP_HOME`).

---

## 1. System Layout & Data Flow

```text
                               ┌───────────────────────────┐
                               │        MCP Client         │
                               │ (Claude / Cursor / IDE)   │
                               └─────────────┬─────────────┘
                                             │ stdio (JSON-RPC 2.0)
                                             ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│ Android-MCP-go Server (~/.android-mcp/)                                                     │
│                                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │ internal/mcp (JSON-RPC Protocol Server & 23 Tool Handlers)                         │   │
│   └────────────────────────┬────────────────────────────────────────────────────────────┘   │
│                            │                                                                │
│                            ▼                                                                │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │ internal/service (DeviceService, InputService, UIService, AppService, etc.)         │   │
│   └────────────────────────┬────────────────────────────────────────────────────────────┘   │
│                            │ Lazy Device Access                                             │
│                            ▼                                                                │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐   │
│   │ internal/device (DeviceManager Orchestrator & State Machine)                        │   │
│   └──────┬─────────────────┬──────────────────┬──────────────────┬──────────────────────┘   │
│          │                 │                  │                  │                          │
│          ▼                 ▼                  ▼                  ▼                          │
│   ┌──────────────┐  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐                  │
│   │ internal/    │  │ internal/    │   │ internal/    │   │ internal/    │                  │
│   │ config       │  │ discovery    │   │ adb          │   │ ui           │                  │
│   └──────────────┘  └──────────────┘   └──────────────┘   └──────────────┘                  │
│          │                 │                  │                  │                          │
│          ▼                 ▼                  ▼                  ▼                          │
│   ┌────────────────────────────────────────────────────────────────────────┐                │
│   │ internal/platformtools (Managed SDK Platform-Tools / adb)              │                │
│   │ internal/scrcpy        (Managed scrcpy Release & Display Window)       │                │
│   │ internal/skills        (Machine-Readable Capability Manifests)         │                │
│   └────────────────────────────────────────────────────────────────────────┘                │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Package Responsibilities

| Package | Purpose & Responsibility |
|---|---|
| `cmd/android-mcp/` | Executable entry point. Handles CLI flags (`--device`, `--debug`), subcommands (`doctor`, `status`, `platform-tools`, `scrcpy`, `skills`), and stdio protocol server lifecycle. |
| `internal/config/` | Portable path resolver (`RuntimePaths`) and unified JSON configuration (`android-mcp.json`) written via atomic file replacement. |
| `internal/adb/` | Safe ADB client executing direct argument slices (`exec.CommandContext`). Performs shell commands, screencaps, and file transfers using managed `adb`. |
| `internal/platformtools/` | Downloads, verifies, extracts, and manages official Google Platform-Tools under `~/.android-mcp/platform-tools/` with Zip Slip security checks. |
| `internal/scrcpy/` | Resolves official GitHub releases (`Genymobile/scrcpy`), verifies SHA-256 digests, and manages non-blocking display mirror child processes. |
| `internal/discovery/` | USB detection, IP resolution (`ip addr`, `ip route`), TCP/IP mode switching (`port 5555`), and wireless connection verification. |
| `internal/device/` | `DeviceManager` state machine and precedence hierarchy (`android-mcp.json` > Auto-pick ADB device). |
| `internal/service/` | Adapter layer mapping Android operations (`DeviceService`, `InputService`, `UIService`, `AppService`, `FileService`, `ShellService`) to MCP handlers. |
| `internal/ui/` | Parses `uiautomator` XML hierarchies, computes element bounds, and generates visual annotated PNG screenshots. |
| `internal/skills/` | Installs and parses 10 machine-readable skill JSON domain manifests under `~/.android-mcp/skills/`. |
| `internal/notification/` | Cross-platform desktop notification engine with rate-limiting and sensitive credential redaction. |
| `internal/mcp/` | Stdio JSON-RPC 2.0 transport handler registering all 23 MCP tools and aliases. |

---

## 3. Directory Layout (`~/.android-mcp/`)

```text
~/.android-mcp/
├── android-mcp.json         # Unified persistent state & preferences
├── platform-tools/          # Managed Google Platform-Tools (adb, fastboot)
├── scrcpy/                  # Managed Genymobile scrcpy bundle & executable
├── skills/                  # Machine-readable capability manifests
├── .downloads/              # Temporary download staging area
├── .staging/                # Temporary archive extraction staging
├── cache/                   # Runtime cache
└── logs/                    # Operational logs
```
