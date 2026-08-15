# Android-MCP-go Technical Checklist

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

---

## 1. Core Architecture & Self-Contained Stack
- [x] Portable path resolver (`RuntimePaths`) with `ANDROID_MCP_HOME` override.
- [x] Strict elimination of fallback to host SDK (`ANDROID_HOME`, `ANDROID_SDK_ROOT`, system `adb`, `/usr/bin/adb`).
- [x] `adb.Client` strictly executes `~/.android-mcp/platform-tools/adb`.
- [x] `scrcpy.Manager` strictly executes `~/.android-mcp/scrcpy/scrcpy` with `ADB=~/.android-mcp/platform-tools/adb`.
- [x] Environment isolation test (`TestEnvironmentIsolation`) passing 100%.

## 2. Managed Platform-Tools (`internal/platformtools`)
- [x] Official Google HTTPS download sources only (`dl.google.com`).
- [x] Zip Slip path traversal security validation.
- [x] Atomic directory installation via `~/.android-mcp/.staging/`.
- [x] CLI `android-mcp platform-tools status|update`.

## 3. Managed `scrcpy` & Adaptive Live View Engine (`internal/scrcpy`)
- [x] Official GitHub Release asset resolver (`Genymobile/scrcpy`).
- [x] Host OS/arch asset mapper (`darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`, `windows/386`).
- [x] SHA-256 integrity verification against release checksum assets.
- [x] Capability probing engine (`DetectBinaryCapabilities`, `DetectHostCapabilities`, `DetectDeviceCapabilities`).
- [x] Progressive degradation fallback pipeline (`Optimized` $\to$ `Reduced` $\to$ `H.264 Fallback` $\to$ `No Audio` $\to$ `Minimal Safe`).
- [x] CLI `android-mcp scrcpy capabilities` & `android-mcp scrcpy profile`.
- [x] Duplicate window protection per device serial.
- [x] Non-blocking launch of `scrcpy` live display mirror window upon device connection.

## 4. Skills & Capability Manifests (`internal/skills`)
- [x] Installed under `~/.android-mcp/skills/` (10 domain JSON manifests).
- [x] CLI `android-mcp skills list` and `android-mcp skills install`.
- [x] Embedded fallback manifests for offline/first-install bootstrap.

## 5. Diagnostic Health Suite & Notifications
- [x] `android-mcp doctor` comprehensive diagnostic health report.
- [x] `android-mcp status` fast operational check (Exit 0 / 1).
- [x] Throttled `--debug` activity notification engine with rate-limiting and credential redaction.

## 6. Verification & Quality Assurance
- [x] 100% unit test pass rate (`go test ./...`).
- [x] 100% data race detector pass rate (`go test -race ./...`) across all 15 packages.
- [x] Physical hardware E2E verification on Sony Xperia `SOG09` (`192.168.1.3:5555`) across all 23 MCP tools (`python3 e2e_test.py`).
- [x] Clean one-line installer (`install.sh`) verified.
