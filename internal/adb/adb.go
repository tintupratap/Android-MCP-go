package adb

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
)

type ADBDevice struct {
	Serial      string
	State       string
	Product     string
	Model       string
	Device      string
	TransportID string
}

func (d ADBDevice) IsUSB() bool {
	return d.State == "device" && !strings.Contains(d.Serial, ":") && !strings.HasPrefix(d.Serial, "emulator-")
}

func (d ADBDevice) IsWiFi() bool {
	return d.State == "device" && strings.Contains(d.Serial, ":")
}

func (d ADBDevice) IsEmulator() bool {
	return strings.HasPrefix(d.Serial, "emulator-")
}

type Client struct {
	adbPath string
}

func NewClient(adbPath string) *Client {
	if adbPath == "" {
		adbPath = FindADBPath()
	}
	return &Client{adbPath: adbPath}
}

func (c *Client) ADBPath() string {
	return c.adbPath
}

func FindADBPath() string {
	if envPath := os.Getenv("ANDROID_MCP_ADB"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}
	if envPath := os.Getenv("ADB"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	paths, err := config.GetRuntimePaths()
	if err == nil && paths != nil {
		return paths.ADB
	}

	home, err := os.UserHomeDir()
	if err == nil {
		managedADB := filepath.Join(home, ".android-mcp", "platform-tools", "adb")
		if runtime.GOOS == "windows" {
			managedADB += ".exe"
		}
		return managedADB
	}

	return filepath.Join(".android-mcp", "platform-tools", "adb")
}

func (c *Client) runCmd(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, c.adbPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outStr := strings.TrimSpace(stdout.String())
	errStr := strings.TrimSpace(stderr.String())

	if err != nil {
		fullErr := outStr
		if errStr != "" {
			if fullErr != "" {
				fullErr += " | " + errStr
			} else {
				fullErr = errStr
			}
		}
		if fullErr == "" {
			fullErr = err.Error()
		}
		return "", fmt.Errorf("adb %s failed: %s", strings.Join(args, " "), fullErr)
	}
	return outStr, nil
}

func (c *Client) ListDevices(ctx context.Context) ([]ADBDevice, error) {
	out, err := c.runCmd(ctx, "devices", "-l")
	if err != nil {
		// Attempt mDNS discovery fallback
		_ = c.MDNSConnectPeers(ctx)
		out, err = c.runCmd(ctx, "devices", "-l")
		if err != nil {
			return nil, err
		}
	}
	devices := ParseADBDevices(out)
	if len(devices) == 0 {
		_ = c.MDNSConnectPeers(ctx)
		if out2, err2 := c.runCmd(ctx, "devices", "-l"); err2 == nil {
			devices = ParseADBDevices(out2)
		}
	}
	return devices, nil
}

func ParseADBDevices(output string) []ADBDevice {
	var devices []ADBDevice
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		dev := ADBDevice{
			Serial: parts[0],
			State:  parts[1],
		}
		for _, part := range parts[2:] {
			if strings.HasPrefix(part, "product:") {
				dev.Product = strings.TrimPrefix(part, "product:")
			} else if strings.HasPrefix(part, "model:") {
				dev.Model = strings.TrimPrefix(part, "model:")
			} else if strings.HasPrefix(part, "device:") {
				dev.Device = strings.TrimPrefix(part, "device:")
			} else if strings.HasPrefix(part, "transport_id:") {
				dev.TransportID = strings.TrimPrefix(part, "transport_id:")
			}
		}
		devices = append(devices, dev)
	}
	return devices
}

func (c *Client) Connect(ctx context.Context, addr string) (bool, error) {
	_ = c.Disconnect(ctx, addr)

	out, err := c.runCmd(ctx, "connect", addr)
	if err != nil {
		return false, err
	}

	lowered := strings.ToLower(out)
	if strings.Contains(lowered, "connected to") || strings.Contains(lowered, "already connected to") {
		time.Sleep(500 * time.Millisecond)
		devices, err := c.ListDevices(ctx)
		if err == nil {
			for _, d := range devices {
				if d.Serial == addr && d.State == "device" {
					return true, nil
				}
			}
		}
	}
	return false, fmt.Errorf("adb connect returned: %s", out)
}

func (c *Client) Disconnect(ctx context.Context, addr string) error {
	_, err := c.runCmd(ctx, "disconnect", addr)
	return err
}

func (c *Client) EnableTCPIP(ctx context.Context, serial string, port int) error {
	if port <= 0 {
		port = 5555
	}
	portStr := fmt.Sprintf("%d", port)
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "tcpip", portStr)
	_, err := c.runCmd(ctx, args...)
	return err
}

func (c *Client) GetProp(ctx context.Context, serial, prop string) (string, error) {
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "shell", "getprop", prop)
	out, err := c.runCmd(ctx, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (c *Client) GetDeviceIP(ctx context.Context, serial string) (string, error) {
	// Strategy 1: ip -4 addr show
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	addrArgs := append(append([]string{}, args...), "shell", "ip", "-4", "addr", "show")
	if out, err := c.runCmd(ctx, addrArgs...); err == nil {
		if ip := ParseIPFromAddrShow(out); ip != "" {
			return ip, nil
		}
	}

	// Strategy 2: ip route
	routeArgs := append(append([]string{}, args...), "shell", "ip", "route")
	if out, err := c.runCmd(ctx, routeArgs...); err == nil {
		if ip := ParseIPFromRoute(out); ip != "" {
			return ip, nil
		}
	}

	// Strategy 3: getprop dhcp.wlan0.ipaddress
	propArgs := append(append([]string{}, args...), "shell", "getprop", "dhcp.wlan0.ipaddress")
	if out, err := c.runCmd(ctx, propArgs...); err == nil {
		ip := strings.TrimSpace(out)
		if ip != "" && isValidPrivateIPv4(ip) {
			return ip, nil
		}
	}

	// Strategy 4: Fallback ip -4 addr show scanning any interface
	if out, err := c.runCmd(ctx, addrArgs...); err == nil {
		if ip := ParseIPFromFallbackAddrShow(out); ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("could not find a valid wireless IP address on device %s", serial)
}

func (c *Client) ExecuteShell(ctx context.Context, serial string, shellArgs ...string) (string, error) {
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "shell")
	args = append(args, shellArgs...)
	return c.runCmd(ctx, args...)
}

func (c *Client) PushFile(ctx context.Context, serial, localPath, remotePath string) error {
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "push", localPath, remotePath)
	_, err := c.runCmd(ctx, args...)
	return err
}

func (c *Client) PullFile(ctx context.Context, serial, remotePath, localPath string) error {
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "pull", remotePath, localPath)
	_, err := c.runCmd(ctx, args...)
	return err
}

func (c *Client) DumpHierarchy(ctx context.Context, serial string) (string, error) {
	// Dump to /data/local/tmp/uidump.xml and cat content
	dumpCmd := "uiautomator dump /data/local/tmp/uidump.xml > /dev/null 2>&1 && cat /data/local/tmp/uidump.xml"
	out, err := c.ExecuteShell(ctx, serial, dumpCmd)
	if err != nil || !strings.Contains(out, "<?xml") && !strings.Contains(out, "<node") {
		// Fallback to dump stdout directly if supported
		out2, err2 := c.ExecuteShell(ctx, serial, "uiautomator", "dump", "/dev/tty")
		if err2 == nil && (strings.Contains(out2, "<?xml") || strings.Contains(out2, "<node")) {
			return out2, nil
		}
		if err != nil {
			return "", err
		}
	}
	return out, nil
}

func (c *Client) Screencap(ctx context.Context, serial string) ([]byte, error) {
	args := []string{}
	if serial != "" {
		args = append(args, "-s", serial)
	}
	args = append(args, "exec-out", "screencap", "-p")

	cmd := exec.CommandContext(ctx, c.adbPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("adb screencap failed: %v | %s", err, stderr.String())
	}
	data := stdout.Bytes()
	if len(data) == 0 {
		return nil, fmt.Errorf("screencap returned empty buffer")
	}
	return data, nil
}

func (c *Client) Tap(ctx context.Context, serial string, x, y int) error {
	_, err := c.ExecuteShell(ctx, serial, "input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
	return err
}

func (c *Client) Swipe(ctx context.Context, serial string, x1, y1, x2, y2, durationMs int) error {
	if durationMs <= 0 {
		durationMs = 300
	}
	_, err := c.ExecuteShell(ctx, serial, "input", "swipe",
		fmt.Sprintf("%d", x1), fmt.Sprintf("%d", y1),
		fmt.Sprintf("%d", x2), fmt.Sprintf("%d", y2),
		fmt.Sprintf("%d", durationMs))
	return err
}

func (c *Client) SendKeys(ctx context.Context, serial string, text string) error {
	// Escape special characters for adb shell input text
	escaped := strings.ReplaceAll(text, " ", "%s")
	escaped = strings.ReplaceAll(escaped, "&", "\\&")
	escaped = strings.ReplaceAll(escaped, "<", "\\<")
	escaped = strings.ReplaceAll(escaped, ">", "\\>")
	escaped = strings.ReplaceAll(escaped, "(", "\\(")
	escaped = strings.ReplaceAll(escaped, ")", "\\)")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "'", "\\'")

	_, err := c.ExecuteShell(ctx, serial, "input", "text", escaped)
	return err
}

func (c *Client) KeyEvent(ctx context.Context, serial string, keycode string) error {
	keyMap := map[string]string{
		"home":        "KEYCODE_HOME",
		"back":        "KEYCODE_BACK",
		"power":       "KEYCODE_POWER",
		"volume_up":   "KEYCODE_VOLUME_UP",
		"volume_down": "KEYCODE_VOLUME_DOWN",
		"menu":        "KEYCODE_MENU",
		"enter":       "KEYCODE_ENTER",
	}

	code := keycode
	if mapped, ok := keyMap[strings.ToLower(keycode)]; ok {
		code = mapped
	}
	if !strings.HasPrefix(code, "KEYCODE_") && !strings.HasPrefix(code, "3") && !strings.HasPrefix(code, "4") {
		code = "KEYCODE_" + strings.ToUpper(code)
	}

	_, err := c.ExecuteShell(ctx, serial, "input", "keyevent", code)
	return err
}

func (c *Client) OpenNotification(ctx context.Context, serial string) error {
	_, err := c.ExecuteShell(ctx, serial, "cmd", "statusbar", "expand-notifications")
	if err != nil {
		_, err = c.ExecuteShell(ctx, serial, "input", "keyevent", "KEYCODE_NOTIFICATION")
	}
	return err
}

func (c *Client) MDNSConnectPeers(ctx context.Context) error {
	out, err := c.runCmd(ctx, "mdns", "services")
	if err != nil {
		return err
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "_adb-tls-connect") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				ipPort := fields[len(fields)-1]
				logging.Debugf("Connecting to mDNS TLS peer: %s", ipPort)
				_, _ = c.runCmd(ctx, "connect", ipPort)
			}
		}
	}
	return nil
}
