# Desktop Notification & Debug Activity System

> **Repository**: [https://github.com/tintupratap/Android-MCP-go](https://github.com/tintupratap/Android-MCP-go)  
> **Author**: Ranapratap ([tintupratap@gmail.com](mailto:tintupratap@gmail.com))

## Overview

`Android-MCP-go` includes a multi-tiered notification system designed to provide desktop awareness of lifecycle events and AI-agent actions without desktop notification spam.

---

## Notification Backends

### macOS
1. `terminal-notifier` (Primary - rich native desktop notifications)
2. `osascript` (Fallback - built-in AppleScript notifications)
3. Silent fallback (if both unavailable)

### Linux
1. `notify-send` (libnotify)
2. Silent fallback (if unavailable)

*Note: Notification failures never break core server operations or device connectivity.*

---

## Notification Modes & Levels

- **Normal Mode (Default)**: Notifies only major lifecycle milestones:
  - Platform-Tools downloading & installation completion.
  - Wireless USB → WiFi ADB bootstrap success (`USB can now be disconnected`).
  - Critical connection failures.
- **Debug Mode (`--debug`)**: Generates real-time, rate-limited notifications for meaningful AI-agent actions (e.g. `AI: Clicked "Login"`, `AI: Launched com.example.app`, `AI: Captured Screenshot`).
- **Silent Mode**: Disables desktop notifications.

---

## Redaction & Security

In `--debug` mode, activity notifications automatically sanitize sensitive text parameters (e.g., passwords, API tokens, sensitive shell command parameters) before sending to desktop notification displays.

---

## Action Correlation & Forensic Logging

Each action receives a unique random action ID (e.g., `ACTION 8f4c2d`), correlating high-level desktop alerts with detailed structured debug log entries.
