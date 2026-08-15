package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/device"
	"github.com/tintupratap/Android-MCP-go/internal/platformtools"
	"github.com/tintupratap/Android-MCP-go/internal/scrcpy"
)

type DoctorReport struct {
	ADBPath           string
	ADBVersion        string
	ADBServerRunning  bool
	PlatformToolsPath string
	PTManaged         bool
	PTSource          string
	ScrcpyPath        string
	ScrcpyBinary      string
	ScrcpyInstalled   bool
	ScrcpyVersion     string
	ScrcpyRunning     bool
	MCPConfigPath     string
	MCPConfigStatus   string
	ScrcpyConfigPath  string
	ScrcpyStatus      string
	USBDevices        []string
	WiFiDevices       []string
	SelectedDevice    *device.Device
	DeviceModel       string
	DeviceSerial      string
	Endpoint          string
	TerminalNotifier  bool
	NotifySend        bool
	ToolCount         int
	IsHealthy         bool
}

func RunDoctor(ctx context.Context, dm device.DeviceManager, client *adb.Client, ptMgr *platformtools.DefaultManager, scrcpyMgr *scrcpy.Manager) *DoctorReport {
	if client == nil {
		client = adb.NewClient("")
	}

	rep := &DoctorReport{
		ADBPath:   client.ADBPath(),
		IsHealthy: true,
		ToolCount: 23,
	}

	if ptMgr != nil {
		rep.PlatformToolsPath = ptMgr.Path()
		rep.PTManaged = ptMgr.IsInstalled()
		url, _ := platformtools.ResolveOfficialURL("darwin")
		rep.PTSource = url
	}

	if scrcpyMgr != nil {
		rep.ScrcpyPath = scrcpyMgr.Path()
		rep.ScrcpyBinary = scrcpyMgr.BinaryPath()
		rep.ScrcpyInstalled = scrcpyMgr.IsInstalled()
		if rep.ScrcpyInstalled {
			ver, _ := scrcpyMgr.GetVersion(ctx)
			rep.ScrcpyVersion = ver
		}
	}

	// 1. ADB Version & Server Check
	verCmd := exec.CommandContext(ctx, client.ADBPath(), "version")
	out, err := verCmd.Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 0 {
			rep.ADBVersion = strings.TrimSpace(lines[0])
		}
		rep.ADBServerRunning = true
	} else {
		rep.ADBVersion = "Unavailable"
		rep.ADBServerRunning = false
		rep.IsHealthy = false
	}

	// 2. Configuration files
	mcpPath, err := config.GetAndroidMCPConfigPath()
	if err == nil {
		rep.MCPConfigPath = mcpPath
		if _, err := os.Stat(mcpPath); err == nil {
			rep.MCPConfigStatus = "OK (~/.android-mcp/android-mcp.json)"
		} else {
			rep.MCPConfigStatus = "Not Created Yet"
		}
	} else {
		rep.MCPConfigStatus = "Error"
	}

	scrcpyPath, err := config.GetScrcpyConfigPath()
	if err == nil {
		if _, err := os.Stat(scrcpyPath); err == nil {
			rep.ScrcpyConfigPath = scrcpyPath
			rep.ScrcpyStatus = "Legacy File Found (Imported into android-mcp.json)"
		} else {
			rep.ScrcpyStatus = "None (Independent)"
		}
	}

	// 3. Device Discovery
	if dm != nil {
		devs, err := dm.List(ctx)
		if err == nil {
			for _, d := range devs {
				if d.State == "device" {
					if d.IsWiFi() {
						rep.WiFiDevices = append(rep.WiFiDevices, d.Serial)
					} else if d.IsUSB() {
						rep.USBDevices = append(rep.USBDevices, d.Serial)
					}
				}
			}
		}

		// Try resolving active target
		dev, err := dm.Resolve(ctx)
		if err == nil && dev != nil {
			rep.SelectedDevice = dev
			rep.DeviceModel = dev.Model
			rep.DeviceSerial = dev.Serial
			rep.Endpoint = dev.Serial
			if scrcpyMgr != nil {
				rep.ScrcpyRunning = scrcpyMgr.IsRunning(dev.Serial)
			}
		}
	}

	// 4. Notifications
	if _, err := exec.LookPath("terminal-notifier"); err == nil {
		rep.TerminalNotifier = true
	}
	if _, err := exec.LookPath("notify-send"); err == nil {
		rep.NotifySend = true
	}

	return rep
}

func (r *DoctorReport) Format() string {
	var sb strings.Builder

	sb.WriteString("Android-MCP-Go Doctor\n")
	sb.WriteString("=====================\n\n")

	sb.WriteString("ADB:\n")
	sb.WriteString(fmt.Sprintf("  Binary:  %s\n", r.ADBPath))
	sb.WriteString(fmt.Sprintf("  Version: %s\n", r.ADBVersion))
	serverState := "running"
	if !r.ADBServerRunning {
		serverState = "NOT running"
	}
	sb.WriteString(fmt.Sprintf("  Server:  %s\n\n", serverState))

	sb.WriteString("Platform-Tools:\n")
	managedStr := "no"
	if r.PTManaged {
		managedStr = "yes (installed)"
	}
	sb.WriteString(fmt.Sprintf("  Managed: %s\n", managedStr))
	if r.PlatformToolsPath != "" {
		sb.WriteString(fmt.Sprintf("  Path:    %s\n", r.PlatformToolsPath))
	}
	sb.WriteString("  Source:  Official Android/Google\n\n")

	sb.WriteString("scrcpy Screen Mirror:\n")
	scrcpyState := "no"
	if r.ScrcpyInstalled {
		scrcpyState = "yes (installed)"
	}
	sb.WriteString(fmt.Sprintf("  Installed: %s\n", scrcpyState))
	sb.WriteString(fmt.Sprintf("  Binary:    %s\n", r.ScrcpyBinary))
	if r.ScrcpyVersion != "" {
		sb.WriteString(fmt.Sprintf("  Version:   %s\n", r.ScrcpyVersion))
	}
	activeMirror := "no"
	if r.ScrcpyRunning {
		activeMirror = "yes (active display window)"
	}
	sb.WriteString(fmt.Sprintf("  Mirror:    %s\n\n", activeMirror))

	sb.WriteString("Configuration:\n")
	sb.WriteString(fmt.Sprintf("  android-mcp.json: %s (%s)\n", r.MCPConfigStatus, r.MCPConfigPath))
	sb.WriteString(fmt.Sprintf("  scrcpy.json:      %s (%s)\n\n", r.ScrcpyStatus, r.ScrcpyConfigPath))

	sb.WriteString("Devices:\n")
	usbStr := fmt.Sprintf("%d detected", len(r.USBDevices))
	if len(r.USBDevices) == 0 {
		usbStr = "none detected"
	}
	wifiStr := fmt.Sprintf("%d connected", len(r.WiFiDevices))
	if len(r.WiFiDevices) == 0 {
		wifiStr = "none connected"
	}
	sb.WriteString(fmt.Sprintf("  USB:  %s\n", usbStr))
	sb.WriteString(fmt.Sprintf("  WiFi: %s\n\n", wifiStr))

	sb.WriteString("Selected Device:\n")
	if r.SelectedDevice != nil {
		sb.WriteString(fmt.Sprintf("  Model:    %s\n", r.DeviceModel))
		sb.WriteString(fmt.Sprintf("  Serial:   %s\n", r.DeviceSerial))
		sb.WriteString(fmt.Sprintf("  Endpoint: %s\n\n", r.Endpoint))
	} else {
		sb.WriteString("  No target device currently resolved (Lazy Mode)\n\n")
	}

	sb.WriteString("Notifications:\n")
	tnStr := "available"
	if !r.TerminalNotifier {
		tnStr = "unavailable"
	}
	nsStr := "available"
	if !r.NotifySend {
		nsStr = "unavailable"
	}
	sb.WriteString(fmt.Sprintf("  terminal-notifier (macOS): %s\n", tnStr))
	sb.WriteString(fmt.Sprintf("  notify-send (Linux):       %s\n\n", nsStr))

	sb.WriteString("MCP:\n")
	sb.WriteString("  Transport: stdio\n")
	sb.WriteString(fmt.Sprintf("  Tools: %d registered\n\n", r.ToolCount))

	statusStr := "HEALTHY"
	if !r.IsHealthy {
		statusStr = "UNHEALTHY"
	} else if r.SelectedDevice == nil {
		statusStr = "READY (Lazy Mode - No device currently active)"
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", statusStr))

	return sb.String()
}

func RunStatus(ctx context.Context, dm device.DeviceManager, scrcpyMgr *scrcpy.Manager) (string, int) {
	if dm == nil {
		return "Android-MCP: Error initializing DeviceManager", 1
	}

	dev, err := dm.Resolve(ctx)
	if err != nil || dev == nil {
		return fmt.Sprintf("Android-MCP\nStatus: LAZY / NO DEVICE\nError: %v", err), 1
	}

	mirrorState := "Inactive"
	if scrcpyMgr != nil && scrcpyMgr.IsRunning(dev.Serial) {
		mirrorState = "Active"
	}

	var sb strings.Builder
	sb.WriteString("Android-MCP\n")
	sb.WriteString(fmt.Sprintf("Device:   %s\n", dev.Model))
	sb.WriteString(fmt.Sprintf("Mode:     %s\n", strings.ToUpper(dev.Connection)))
	sb.WriteString(fmt.Sprintf("Endpoint: %s\n", dev.Serial))
	sb.WriteString("ADB:      Connected\n")
	sb.WriteString(fmt.Sprintf("scrcpy:   %s\n", mirrorState))
	sb.WriteString("Status:   READY\n")

	return sb.String(), 0
}
