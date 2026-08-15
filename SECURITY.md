# Android-MCP-go Security Policy & Architecture

## 1. Overview & Trust Model

`Android-MCP-go` acts as a high-privilege bridge between AI assistants (via the MCP protocol over stdio) and connected Android hardware (via ADB). 

### Trust Domains

1. **Host Environment (Trusted)**: Machine running `android-mcp-go`.
2. **MCP Client (Semi-Trusted)**: Claude Desktop, Cursor IDE, or AI Agent communicating over stdin/stdout.
3. **Android Device (Untrusted Input Source)**: UI dumps, package names, files, network data, and ADB outputs.

---

## 2. Command Execution & Injection Prevention

### Strict Isolation Rule
`Android-MCP-go` NEVER invokes `sh -c` or shell string concatenation to execute ADB or system commands.

All command invocations use `exec.CommandContext(ctx, binary, args...)` with explicit, unescaped argument slices:

```go
// SAFE: Direct argument array passing
cmd := exec.CommandContext(ctx, adbPath, "-s", serial, "shell", "am", "force-stop", packageName)
```

This completely eliminates command injection vectors when handling parameters such as package names, paths, selectors, or text inputs.

---

## 3. Host vs Android Execution Isolation

To prevent accidental host execution of Android commands:
- **Host Execution**: Restricted strictly to ADB binary discovery (`exec.LookPath("adb")`) and local notification binaries (`terminal-notifier`, `notify-send`).
- **Android Execution**: ALL device operations route explicitly through `adb -s <serial> shell ...` or ADB protocol subcommands (`push`, `pull`, `screencap`).

---

## 4. Filesystem Access Controls

- File transfer operations (`file_push`, `file_pull`) enforce path validation.
- Without Android root access, ADB operations are restricted by standard Android SELinux permissions to accessible storage partitions (`/sdcard`, `/data/local/tmp`).
- Attempts to read restricted `/data/data/` directories on non-rooted devices fail gracefully returning permission errors—never hanging or compromising host security.

---

## 5. Persistent State Integrity

`~/.android-mcp/android-mcp.json` is protected using atomic file writes:
1. Data written to temporary file `android-mcp-*.tmp`.
2. `fsync()` called to ensure disk persistence.
3. Atomic `os.Rename()` into place.

If `android-mcp.json` or `~/.scrcpy/scrcpy.json` is corrupted or tampered with, `Android-MCP-go` backs up the corrupted file to `.bak`, logs a warning, and falls back cleanly to default safe settings.

---

## 6. Multi-Device Safety

To prevent accidental execution of commands on the wrong Android device when multiple phones or emulators are connected:
- Device selection follows strict precedence:
  `Explicit CLI Flag > Environment Variable > Persisted Serial/IP > Scrcpy State > Physical Auto-Pick`
- If no deterministic selection can be made, `Android-MCP-go` refuses auto-selection and prompts the caller to specify `--device` explicitly.
