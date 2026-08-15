# Android-MCP-go Security Policy & Architecture

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

`Android-MCP-go` acts as a high-privilege bridge between AI agents (via standard stdio JSON-RPC 2.0) and physical Android hardware (via ADB). This document details our security boundaries, archive extraction validation, command execution rules, and data privacy safeguards.

---

## 1. Security Architecture & Threat Model

```text
┌───────────────────────────┐         stdio (JSON-RPC 2.0)        ┌───────────────────────────┐
│        AI Agent           │ ──────────────────────────────────> │     Android-MCP-go        │
│ (Claude / Cursor / IDE)   │ <────────────────────────────────── │     (Host Boundary)       │
└───────────────────────────┘                                     └─────────────┬─────────────┘
                                                                                │ Managed ADB
                                                                                ▼
                                                                  ┌───────────────────────────┐
                                                                  │      Android Device       │
                                                                  │ (SELinux Sandbox Boundary)│
                                                                  └───────────────────────────┘
```

---

## 2. Security Safeguards

### A. Official HTTPS Download Verification
- **Platform-Tools**: Downloaded exclusively from official Google HTTPS endpoints (`https://dl.google.com/android/repository/platform-tools-latest-*.zip`).
- **Managed `scrcpy`**: Downloaded exclusively from official GitHub Releases (`https://github.com/Genymobile/scrcpy`).
- **Integrity Verification**: SHA-256 digests are calculated and checked against published release checksum assets before extraction.

### B. Zip & Tar Slip Path Traversal Protection
Archive extraction routines (`extractZipSecurely` & `extractTarGzSecurely`) inspect every header entry path before extraction to guarantee entries cannot traverse outside extraction target directories:

```go
filePath := filepath.Join(destDir, header.Name)
if !strings.HasPrefix(filepath.Clean(filePath), filepath.Clean(destDir)) {
    return fmt.Errorf("Zip/Tar Slip path traversal attempt detected: %s", header.Name)
}
```

### C. Direct Slice Execution (No `sh -c`)
`Android-MCP-go` NEVER invokes `sh -c`, `cmd /c`, or raw shell string evaluation. All commands use `exec.CommandContext(ctx, binary, args...)` with explicit, unescaped string slices to prevent command injection.

### D. Sensitive Data Redaction
Real-time desktop notifications under `--debug` automatically redact sensitive parameters (passwords, OTPs, authentication tokens, secret keys) before sending system alerts.

### E. Atomic Persistence Security
Persistent state updates to `~/.android-mcp/android-mcp.json` are written to temporary files (`android-mcp.json.tmp.*`) and flushed before performing atomic filesystem renames, preventing partial state corruption.

---

## 3. Reporting Security Vulnerabilities

Please report security issues directly to author Ranapratap at **[tintupratap@gmail.com](mailto:tintupratap@gmail.com)**. Reports will receive initial response within 24 hours.
