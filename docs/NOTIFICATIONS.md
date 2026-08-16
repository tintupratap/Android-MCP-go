# Desktop Notification & Debug Activity System

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

`Android-MCP-go` provides a non-intrusive desktop notification system that alerts users to major connection milestones and real-time AI agent actions without notification spam.

---

## 1. Cross-Platform Backends

- **macOS**: `terminal-notifier` (Primary rich notification backend) with `osascript` AppleScript fallback.
- **Linux**: `notify-send` (`libnotify`).
- **Resilience**: If no desktop notifier is installed, operations continue silently without error.

---

## 2. Notification Modes & Activity Queue

- **Normal Mode (Default)**: Emits notifications only for major lifecycle milestones:
  - Dependency download / installation completion.
  - USB $\to$ WiFi ADB bootstrap completion (`USB can be disconnected`).
  - Connection error alerts.
- **Debug Mode (`--debug`)**: Emits real-time alerts for AI-agent actions (`AI: Clicked Element`, `AI: Launched App`, `AI: Captured Screenshot`).
- **Quiet Mode (`--quiet` / `ANDROID_MCP_QUIET=true`)**: Completely disables all desktop notifications using `notification.NewNoopNotifier()`.
- **Rate Limiting**: Async alert queue rate-limits alerts (default 250ms interval) to eliminate desktop notification spam.

---

## 3. Parameter Redaction & Action Correlation

- **Data Redaction**: Sensitive parameter fields (passwords, auth tokens, secret keys) are automatically sanitized before sending desktop notifications.
- **Correlation IDs**: Each action generates a unique correlation ID (e.g. `ACTION 8f4c2d`) linking desktop alerts to detailed structured debug logs.
