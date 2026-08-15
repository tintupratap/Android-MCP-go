# Contributing to Android-MCP-go

Thank you for your interest in contributing to **Android-MCP-go**!

This project is a production-grade, 100% self-contained Model Context Protocol (MCP) server for Android devices built in Go. We welcome bug reports, feature suggestions, documentation improvements, and pull requests.

---

## 1. Development Principles & Core Directives

When submitting contributions, please strictly adhere to the following architectural principles:

1. **Zero External Runtime Dependency**:
   - The project MUST remain 100% self-contained under `~/.android-mcp/` (or `$ANDROID_MCP_HOME`).
   - NEVER add runtime dependencies on host Android SDKs (`ANDROID_HOME`, `ANDROID_SDK_ROOT`), system `adb`, system `scrcpy`, or host package managers (Homebrew, MacPorts, apt).
2. **Command Execution Security**:
   - Executing subcommands MUST use `exec.CommandContext(ctx, binary, args...)` with direct string argument slices.
   - **NEVER** pass raw shell strings via `sh -c` or `cmd /c`.
3. **Data Race Safety**:
   - All state mutations (device registration, process lifecycles, configuration updates) MUST be thread-safe.
   - All PRs must pass `go test -race ./...` with zero race warnings.
4. **Data Privacy & Redaction**:
   - Notifications and logs under `--debug` MUST redact sensitive information (passwords, OTPs, auth tokens).

---

## 2. Setting Up Your Development Environment

### Prerequisites
- **Go**: Version 1.22+ (used only for building from source).
- **Python**: Version 3.10+ (used for running the `e2e_test.py` harness).
- **Android Hardware** (Optional but recommended): Physical Android device connected over USB or WiFi ADB.

### Clone and Build

```bash
git clone https://github.com/tintupratap/Android-MCP-go.git
cd Android-MCP-go

# Build binary locally
go build -o android-mcp ./cmd/android-mcp

# Run diagnostic check
./android-mcp doctor
```

---

## 3. Testing Guidelines

Before opening a pull request, run the complete verification suite:

```bash
# 1. Run unit tests
go test ./...

# 2. Run data race detector across all packages
go test -race ./...

# 3. Run static code analysis
go vet ./...

# 4. Run physical / simulated E2E test suite (if device is connected)
python3 e2e_test.py
```

---

## 4. Submitting a Pull Request (PR)

1. **Fork the repository** on GitHub.
2. **Create a topic branch**: `git checkout -b feature/my-new-feature`
3. **Commit your changes**: Write clear, descriptive commit messages.
4. **Run verification**: Ensure `go test -race ./...` passes cleanly.
5. **Push to your fork**: `git push origin feature/my-new-feature`
6. **Open a Pull Request** against `main` on [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go).

---

## 5. Reporting Issues & Feature Requests

Please use the GitHub Issue Tracker to report bugs or request features:
- **Bug Reports**: Include host OS (`darwin/amd64`, `darwin/arm64`, `linux/amd64`, `windows/amd64`), Android device model, steps to reproduce, and output of `android-mcp doctor`.
- **Security Vulnerabilities**: Report sensitive security issues directly to author Ranapratap at [tintupratap@gmail.com](mailto:tintupratap@gmail.com).

Thank you for helping make `Android-MCP-go` better!
