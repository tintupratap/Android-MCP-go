# Self-Contained Android SDK Platform-Tools Management

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

`Android-MCP-go` automatically manages its own official Google Android SDK Platform-Tools (`adb`, `fastboot`) under `~/.android-mcp/platform-tools/` so users do not need to install an Android SDK or system ADB manually.

---

## 1. Official Google Distribution Source Policy

Platform-Tools archives originate **exclusively** from official Google Android developer repositories:

- **macOS**: `https://dl.google.com/android/repository/platform-tools-latest-darwin.zip`
- **Linux**: `https://dl.google.com/android/repository/platform-tools-latest-linux.zip`
- **Windows**: `https://dl.google.com/android/repository/platform-tools-latest-windows.zip`

Third-party mirrors or unverified binary hosts are strictly rejected.

---

## 2. Security & Installation Mechanics

1. **Zip Slip Protection**: Archive extraction routines inspect header paths to guarantee files cannot traverse outside target staging directories (`~/.android-mcp/.staging/`).
2. **Execution Verification**: Executes `adb version` check inside staging before performing atomic replacements.
3. **Atomic Replacement**: Swaps validated packages into `~/.android-mcp/platform-tools/`. If an error occurs, existing working installations remain untouched.
4. **Permissions**: Normalizes executable file permissions to `0755`.

---

## 3. Path Resolver Precedence

1. **`ANDROID_MCP_ADB` Environment Override** (Development testing)
2. **Authoritative Managed Path**: `~/.android-mcp/platform-tools/adb` (or `$ANDROID_MCP_HOME/platform-tools/adb`)

---

## 4. CLI Management Commands

```bash
# Check managed Platform-Tools installation status
android-mcp platform-tools status

# Force update Platform-Tools to latest official Google release
android-mcp platform-tools update
```
