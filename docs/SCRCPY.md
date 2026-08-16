# Managed scrcpy & Live Display Mirroring System

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

`Android-MCP-go` automatically manages official `scrcpy` releases to launch real-time GUI screen mirror windows whenever an Android device connects, allowing human users to continuously observe AI agent operations.

```text
~/.android-mcp/scrcpy/
```

---

## 1. Official GitHub Releases Pipeline

- **Source**: Official upstream repository [`Genymobile/scrcpy`](https://github.com/Genymobile/scrcpy).
- **Dynamic Resolver**: Queries `https://api.github.com/repos/Genymobile/scrcpy/releases/latest` to resolve latest stable releases (ignoring drafts/prereleases).
- **Host Resolution**:
  - `darwin/arm64`: macOS Apple Silicon bundle (`scrcpy-macos-aarch64-v*.tar.gz`)
  - `darwin/amd64`: macOS Intel bundle (`scrcpy-macos-x86_64-v*.tar.gz`)
  - `linux/amd64`: Linux x86_64 bundle (`scrcpy-linux-x86_64-v*.tar.gz`)
  - `windows/amd64`: Windows 64-bit package (`scrcpy-win64-v*.zip`)
  - `windows/386`: Windows 32-bit package (`scrcpy-win32-v*.zip`)
- **Integrity & Security**: Verifies SHA-256 digests against published release checksum assets, enforces Zip/Tar Slip path traversal protection, runs `scrcpy --version` verification in staging, and atomically installs.

---

## 2. Process Manager & Process Lifecycle

- **Child Process Execution**: Spawned asynchronously via Go's `exec.Command` with `ADB=~/.android-mcp/platform-tools/adb` passed in `cmd.Env`.
- **Targeting**: Explicitly targets the resolved device serial (`scrcpy -s <serial> --window-title "Android-MCP — <model> (<serial>)"`).
- **Window Flags**: Includes `--always-on-top` by default so the mirror window stays visible. Use `--no-always-on-top` CLI flag to launch standard window mode.
- **Disabling Live View**: Use `--no-scrcpy` CLI flag to disable display mirror startup entirely.
- **Duplicate Protection**: Thread-safe process table prevents spawning multiple windows per serial on repeated tool calls.
- **Clean Exit**: Terminating `android-mcp` sends `SIGTERM` to active `scrcpy` processes, preventing orphan/zombie GUI windows.

---

## 3. CLI Management Commands

```bash
# Query managed scrcpy installation status & version
android-mcp scrcpy status

# Force update scrcpy to latest official release
android-mcp scrcpy update

# Manually launch screen mirror window for connected device
android-mcp scrcpy start

# Stop active screen mirror windows
android-mcp scrcpy stop
```
