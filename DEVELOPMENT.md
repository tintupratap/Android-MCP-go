# Android-MCP-go Development Guide

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

This guide outlines codebase structure, building, testing procedures, and guidelines for extending `Android-MCP-go`.

---

## 1. Project Directory Layout

```text
Android-MCP-go/
├── cmd/android-mcp/          # Main executable entry point & CLI flag/subcommand parser
├── internal/
│   ├── adb/                  # ADB client, device parser, shell & file transfer logic
│   ├── config/               # Portable path resolver (RuntimePaths) & atomic JSON config
│   ├── device/               # DeviceManager orchestrator & state machine
│   ├── discovery/            # Wireless IP discovery & TCP/IP bootstrap engine
│   ├── doctor/               # Comprehensive diagnostic health engine
│   ├── logging/              # Level-based logger
│   ├── mcp/                  # JSON-RPC 2.0 stdio transport & 23 MCP tool handlers
│   ├── notification/         # Desktop notifier & --debug activity rate-limited alert queue
│   ├── platformtools/        # Managed Google Platform-Tools installer with Zip Slip checks
│   ├── scrcpy/               # Managed scrcpy GitHub release installer & display mirror manager
│   ├── service/              # Decoupled service adapters mapping Android tools
│   ├── skills/               # Machine-readable skill domain manifest manager
│   └── ui/                   # UI hierarchy XML parser & visual PNG annotation engine
├── docs/                     # Technical documentation & testing reports
├── skills/                   # Domain capability JSON manifests (v0.4.0)
├── e2e_test.py               # E2E JSON-RPC physical hardware verification harness
├── install.sh                # One-line curl installation script
├── go.mod                    # Module definition (github.com/tintupratap/Android-MCP-go)
└── README.md                 # Project homepage & manual
```

---

## 2. Build & Test Commands

```bash
# 1. Build local binary
go build -o android-mcp ./cmd/android-mcp

# 2. Run unit tests across all 15 packages
go test ./...

# 3. Run data race detector across all packages
go test -race ./...

# 4. Run static code analysis
go vet ./...

# 5. Run physical hardware E2E test suite (requires connected Android device)
python3 e2e_test.py
```

See [docs/TESTING_REPORT.md](docs/TESTING_REPORT.md) for physical testing specifications and benchmarks on Sony Xperia `SOG09` hardware.

---

## 3. How to Add a New MCP Tool

1. **Define Service Method**: Add method to corresponding service in `internal/service/` (e.g. `DeviceService`, `UIService`).
2. **Register Tool Schema**: Add tool definition to `RegisterTools()` in `internal/mcp/server.go`.
3. **Add JSON-RPC Handler**: Implement RPC handler in `internal/mcp/server.go` with parameter validation.
4. **Update Manifests & Tests**: Update `skills/` JSON domain manifests, add unit test in `internal/mcp/mcp_test.go`, and add step in `e2e_test.py`.
