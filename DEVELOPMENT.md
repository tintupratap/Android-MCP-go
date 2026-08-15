# Android-MCP-go Development Guide

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Project Structure

```text
Android-MCP-go/
├── cmd/
│   └── android-mcp/        # Main executable entry point & CLI flag parsing
├── internal/
│   ├── adb/                # ADB runner, device parser, IP extraction logic
│   ├── config/             # Atomic JSON persistence & scrcpy reader
│   ├── device/             # DeviceManager, state machine, preference resolver
│   ├── discovery/          # Multi-strategy IP discovery & TCP/IP bootstrap engine
│   ├── logging/            # Structured level logger
│   ├── mcp/                # JSON-RPC 2.0 stdio MCP server & 14 tool handlers
│   ├── notification/       # macOS & Linux desktop notifier integration
│   └── ui/                 # XML UI hierarchy parser, selectors, visual annotation
├── docs/
│   └── PORTING_ANALYSIS.md # Detailed archaeology & compatibility analysis
├── e2e_test.py             # Physical hardware end-to-end test script
├── ARCHITECTURE.md         # System architecture specification
├── CONFIGURATION.md        # Persistent config documentation
├── ROADMAP.md              # Project milestones & roadmap
├── TODO.md                 # Technical task checklist
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
└── README.md               # User manual & quickstart
```

---

## Building & Testing

### Build Executable

```bash
go build -o android-mcp ./cmd/android-mcp
```

### Run Unit Tests

```bash
go test -v ./...
```

### Run Integration & Physical Verification Tests

```bash
python3 e2e_test.py
```

See [docs/TESTING_REPORT.md](docs/TESTING_REPORT.md) for full physical test specifications and benchmarks on Sony Xperia SOG09 hardware.

---

## Adding New MCP Tools

To add a new tool to `Android-MCP-go`:
1. Open `internal/mcp/server.go`.
2. Define the `Tool` schema in `s.registerTools()` (specifying `Name`, `Description`, `InputSchema`, and `Annotations`).
3. Add the handler closure implementing the execution logic using `s.requireDevice(ctx)` or low-level ADB commands.
4. Add unit/integration tests in `internal/mcp/mcp_test.go` and `e2e_test.py`.
