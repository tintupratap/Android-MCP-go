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
│   ├── adb/                  # ADB client, mcp-helper.dex, shell & file transfer logic
│   ├── config/               # Portable path resolver (RuntimePaths) & atomic JSON config
│   ├── device/               # DeviceManager orchestrator & state machine
│   ├── discovery/            # Wireless IP discovery & TCP/IP bootstrap engine
│   ├── doctor/               # Comprehensive diagnostic health engine
│   ├── logging/              # Level-based logger
│   ├── mcp/                  # JSON-RPC 2.0 stdio transport & 25 MCP tool handlers
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
# 1. Install via Go package manager
go install github.com/tintupratap/Android-MCP-go/cmd/android-mcp@latest

# 2. Build local Go binary (pre-bundled with internal/adb/mcp-helper.dex)
go build -o android-mcp ./cmd/android-mcp

# 3. (Optional) Recompile embedded Android DEX helper from Java source:
# mkdir -p build/out
# javac -cp $ANDROID_HOME/platforms/android-34/android.jar -d build/out internal/adb/java/com/android/mcp/HelperMain.java
# $ANDROID_HOME/build-tools/34.0.0/d8 --output ./ build/out/com/android/mcp/HelperMain.class
# mv classes.dex internal/adb/mcp-helper.dex

# 4. Run unit tests across all 15 packages
go test ./...

# 5. Run data race detector across all packages
go test -race ./...

# 6. Run static code analysis
go vet ./...

# 7. Run physical hardware E2E test suite (requires connected Android device)
python3 e2e_test.py

# 8. Build multi-platform release distribution archives & checksums for GitHub Releases
./build_releases.sh v0.5.0
```

See [docs/TESTING_REPORT.md](docs/TESTING_REPORT.md) for physical testing specifications and benchmarks on Sony Xperia `SOG09` hardware.

---

## 3. How to Add a New MCP Tool

1. **Define Service Method**: Add method to corresponding service in `internal/service/` (e.g. `DeviceService`, `UIService`).
2. **Register Tool Schema**: Add tool definition to `RegisterTools()` in `internal/mcp/server.go`.
3. **Add JSON-RPC Handler**: Implement RPC handler in `internal/mcp/server.go` with parameter validation.
4. **Update Manifests & Tests**: Update `skills/` JSON domain manifests, add unit test in `internal/mcp/mcp_test.go`, and add step in `e2e_test.py`.
