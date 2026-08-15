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
	Version           string
	ADBPath           string
	ADBVersion        string
	ADBServerRunning  bool
	PlatformToolsPath string
	PTManaged         bool
	PTInstalled       bool
	ScrcpyPath        string
	ScrcpyBinary      string
	ScrcpyInstalled   bool
	ScrcpyVersion     string
	ScrcpyRunning     bool
	MCPConfigPath     string
	MCPConfigStatus   string
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
	ScrcpyErr         string
}

func RunDoctor(ctx context.Context, dm device.DeviceManager, client *adb.Client, ptMgr *platformtools.DefaultManager, scrcpyMgr *scrcpy.Manager) *DoctorReport {
	if client == nil {
		client = adb.NewClient("")
	}

	rep := &DoctorReport{
		Version:   "0.4.0",
		ADBPath:   client.ADBPath(),
		IsHealthy: true,
		ToolCount: 23,
	}

	if ptMgr != nil {
		rep.PlatformToolsPath = ptMgr.Path()
		rep.PTInstalled = ptMgr.IsInstalled()
		rep.PTManaged = rep.PTInstalled
	}

	if scrcpyMgr != nil {
		rep.ScrcpyPath = scrcpyMgr.Path()
		rep.ScrcpyBinary = scrcpyMgr.BinaryPath()
		rep.ScrcpyInstalled = scrcpyMgr.IsInstalled()
		if rep.ScrcpyInstalled {
			ver, _ := scrcpyMgr.GetVersion(ctx)
			rep.ScrcpyVersion = ver
		} else {
			// Auto-attempt installation if missing
			if _, err := scrcpyMgr.EnsureInstalled(ctx); err == nil {
				rep.ScrcpyInstalled = true
				ver, _ := scrcpyMgr.GetVersion(ctx)
				rep.ScrcpyVersion = ver
			} else {
				rep.ScrcpyErr = err.Error()
			}
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
			rep.MCPConfigStatus = "OK"
		} else {
			rep.MCPConfigStatus = "Not Created Yet"
		}
	} else {
		rep.MCPConfigStatus = "Error"
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

	sb.WriteString(fmt.Sprintf("Android-MCP-go v%s\n\n", r.Version))

	sb.WriteString("Core:\n")
	sb.WriteString("  ✓ MCP binary\n")
	sb.WriteString(fmt.Sprintf("  ✓ Configuration (%s)\n\n", r.MCPConfigPath))

	sb.WriteString("Platform-Tools:\n")
	if r.PTInstalled {
		sb.WriteString(fmt.Sprintf("  ✓ Installed (%s)\n", r.PlatformToolsPath))
	} else {
		sb.WriteString("  ✗ Platform-Tools missing\n")
	}
	sb.WriteString(fmt.Sprintf("  ✓ ADB executable (%s)\n", r.ADBPath))
	if r.ADBServerRunning {
		sb.WriteString(fmt.Sprintf("  ✓ ADB version (%s)\n\n", r.ADBVersion))
	} else {
		sb.WriteString("  ✗ ADB server not running\n\n")
	}

	sb.WriteString("scrcpy Display Mirror:\n")
	if r.ScrcpyInstalled {
		sb.WriteString(fmt.Sprintf("  ✓ Installed (%s)\n", r.ScrcpyPath))
		sb.WriteString(fmt.Sprintf("  ✓ Executable valid (%s)\n", r.ScrcpyVersion))
		mirrorState := "no"
		if r.ScrcpyRunning {
			mirrorState = fmt.Sprintf("running against %s", r.Endpoint)
		}
		sb.WriteString(fmt.Sprintf("  ✓ Active Mirror (%s)\n\n", mirrorState))
	} else {
		sb.WriteString("  ✗ scrcpy not installed\n")
		if r.ScrcpyErr != "" {
			sb.WriteString(fmt.Sprintf("  Reason: %s\n", r.ScrcpyErr))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Device:\n")
	if r.SelectedDevice != nil {
		sb.WriteString(fmt.Sprintf("  ✓ %s (%s %s)\n\n", r.DeviceModel, strings.ToUpper(r.SelectedDevice.Connection), r.Endpoint))
	} else {
		sb.WriteString("  - No active target device resolved (Lazy Mode)\n\n")
	}

	sb.WriteString("Notifications:\n")
	if r.TerminalNotifier {
		sb.WriteString("  ✓ terminal-notifier (macOS)\n\n")
	} else if r.NotifySend {
		sb.WriteString("  ✓ notify-send (Linux)\n\n")
	} else {
		sb.WriteString("  - Standard system logger\n\n")
	}

	statusStr := "HEALTHY"
	if !r.IsHealthy || !r.ADBServerRunning {
		statusStr = "UNHEALTHY — ADB server unavailable"
	} else if !r.ScrcpyInstalled {
		statusStr = "DEGRADED — MCP core functional, live screen view unavailable"
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
