package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/device"
	"github.com/tintupratap/Android-MCP-go/internal/doctor"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
	"github.com/tintupratap/Android-MCP-go/internal/mcp"
	"github.com/tintupratap/Android-MCP-go/internal/notification"
	"github.com/tintupratap/Android-MCP-go/internal/platformtools"
)

const Version = "0.2.0"

func main() {
	// Check for subcommand arguments first (e.g. android-mcp doctor, android-mcp status, android-mcp platform-tools)
	if len(os.Args) > 1 {
		subcmd := strings.ToLower(os.Args[1])
		switch subcmd {
		case "doctor":
			runDoctorCmd()
			os.Exit(0)
		case "status":
			runStatusCmd()
			return
		case "platform-tools":
			runPlatformToolsCmd(os.Args[2:])
			os.Exit(0)
		}
	}

	var (
		deviceFlag     string
		connectionFlag string
		wifiFlag       string
		usbFlag        string
		debugFlag      bool
		showVersion    bool
		showDoctor     bool
		showStatus     bool
	)

	flag.StringVar(&deviceFlag, "device", "", "ADB device serial or host:port")
	flag.StringVar(&connectionFlag, "connection", "", "Preferred device connection type (auto, usb, wifi)")
	flag.StringVar(&wifiFlag, "wifi", "", "Use WiFi ADB. Accepts HOST or HOST:PORT (defaults to port 5555)")
	flag.StringVar(&usbFlag, "usb", "", "Use USB ADB. Optionally provide a specific device serial")
	flag.BoolVar(&debugFlag, "debug", false, "Enable verbose debug logging and activity desktop notifications")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.BoolVar(&showDoctor, "doctor", false, "Run diagnostic health check")
	flag.BoolVar(&showStatus, "status", false, "Show concise device status and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("Android-MCP-go v%s\n", Version)
		os.Exit(0)
	}

	if showDoctor {
		runDoctorCmd()
		os.Exit(0)
	}

	if showStatus {
		runStatusCmd()
		return
	}

	notifLevel := notification.LevelNormal
	if debugFlag || os.Getenv("LOG_LEVEL") == "debug" {
		logging.SetLevel(logging.LevelDebug)
		notifLevel = notification.LevelDebug
	} else {
		logging.SetLevel(logging.LevelInfo)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	notifier := notification.NewNotifier()
	actNotifier := notification.NewActivityNotifier(notifier, notifLevel)

	// Ensure Platform-Tools are present
	ptMgr, err := platformtools.NewManager(notifier)
	if err == nil {
		adbPath, errEnsure := ptMgr.Ensure(ctx)
		if errEnsure != nil {
			logging.Warnf("Platform-Tools ensure warning: %v. Falling back to system ADB.", errEnsure)
		} else {
			_ = os.Setenv("ANDROID_MCP_ADB", adbPath)
		}
	}

	pref := resolvePreference(deviceFlag, connectionFlag, wifiFlag, usbFlag)

	logging.Infof("Android-MCP-go v%s starting (connection=%s, source=%s, debug=%v)...", Version, pref.Connection, pref.Source, debugFlag)

	adbClient := adb.NewClient("")
	deviceMgr := device.NewDeviceManager(adbClient, notifier, pref)

	server := mcp.NewServer(deviceMgr, adbClient, os.Stdin, os.Stdout, actNotifier)

	if err := server.ListenAndServe(ctx); err != nil && ctx.Err() == nil {
		logging.Errorf("Server error: %v", err)
		os.Exit(1)
	}

	logging.Infof("Android-MCP server shut down cleanly.")
}

func runDoctorCmd() {
	ctx := context.Background()
	notifier := notification.NewNotifier()

	ptMgr, _ := platformtools.NewManager(notifier)
	if ptMgr != nil && ptMgr.IsInstalled() {
		_ = os.Setenv("ANDROID_MCP_ADB", ptMgr.ADBPath())
	}

	adbClient := adb.NewClient("")
	deviceMgr := device.NewDeviceManager(adbClient, notifier, device.DevicePreference{})

	rep := doctor.RunDoctor(ctx, deviceMgr, adbClient, ptMgr)
	fmt.Print(rep.Format())
}

func runStatusCmd() {
	ctx := context.Background()
	notifier := notification.NewNotifier()

	ptMgr, _ := platformtools.NewManager(notifier)
	if ptMgr != nil && ptMgr.IsInstalled() {
		_ = os.Setenv("ANDROID_MCP_ADB", ptMgr.ADBPath())
	}

	adbClient := adb.NewClient("")
	deviceMgr := device.NewDeviceManager(adbClient, notifier, device.DevicePreference{})

	out, code := doctor.RunStatus(ctx, deviceMgr)
	fmt.Println(out)
	if code != 0 {
		os.Exit(code)
	}
}

func runPlatformToolsCmd(args []string) {
	ctx := context.Background()
	notifier := notification.NewNotifier()
	ptMgr, err := platformtools.NewManager(notifier)
	if err != nil {
		fmt.Printf("Error initializing Platform-Tools manager: %v\n", err)
		os.Exit(1)
	}

	action := "status"
	if len(args) > 0 {
		action = strings.ToLower(args[0])
	}

	switch action {
	case "status":
		fmt.Printf("Platform-Tools Path: %s\n", ptMgr.Path())
		fmt.Printf("ADB Binary Path:     %s\n", ptMgr.ADBPath())
		fmt.Printf("Installed:           %v\n", ptMgr.IsInstalled())
		if ptMgr.IsInstalled() {
			fmt.Printf("ADB Version:         %s\n", ptMgr.ADBPath())
		}
	case "update", "reinstall":
		fmt.Println("Downloading and updating official Android Platform-Tools...")
		adbPath, err := ptMgr.Update(ctx)
		if err != nil {
			fmt.Printf("❌ Failed to update Platform-Tools: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Platform-Tools updated successfully! Binary: %s\n", adbPath)
	default:
		fmt.Printf("Unknown platform-tools command: %s. Use status, update, or reinstall.\n", action)
	}
}

func resolvePreference(deviceFlag, connectionFlag, wifiFlag, usbFlag string) device.DevicePreference {
	envDevice := cleanEnv("ANDROID_MCP_DEVICE")
	envConn := normalizeConn(cleanEnv("ANDROID_MCP_CONNECTION"))
	envHost := cleanEnv("ANDROID_MCP_HOST")

	// CLI --wifi
	if isFlagPassed("wifi") || wifiFlag != "" {
		host := wifiFlag
		if host == "" {
			host = envHost
		}
		serial := adb.FormatWiFiSerial(host, 5555)
		return device.DevicePreference{
			Connection: "wifi",
			Serial:     serial,
			Source:     "--wifi",
		}
	}

	// CLI --usb
	if isFlagPassed("usb") || usbFlag != "" {
		return device.DevicePreference{
			Connection: "usb",
			Serial:     strings.TrimSpace(usbFlag),
			Source:     "--usb",
		}
	}

	// CLI --device
	if deviceFlag != "" {
		conn := normalizeConn(connectionFlag)
		if conn == "auto" {
			conn = "auto"
		}
		return device.DevicePreference{
			Connection: conn,
			Serial:     strings.TrimSpace(deviceFlag),
			Source:     "--device",
		}
	}

	// Env ANDROID_MCP_DEVICE
	if envDevice != "" {
		return device.DevicePreference{
			Connection: envConn,
			Serial:     envDevice,
			Source:     "ANDROID_MCP_DEVICE",
		}
	}

	// Env ANDROID_MCP_CONNECTION=wifi or ANDROID_MCP_HOST
	if envConn == "wifi" || envHost != "" {
		serial := adb.FormatWiFiSerial(envHost, 5555)
		return device.DevicePreference{
			Connection: "wifi",
			Serial:     serial,
			Source:     "ANDROID_MCP_CONNECTION/ANDROID_MCP_HOST",
		}
	}

	// CLI --connection fallback
	conn := normalizeConn(connectionFlag)
	if conn == "auto" {
		conn = envConn
	}

	return device.DevicePreference{
		Connection: conn,
		Serial:     "",
		Source:     "auto-detect",
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func cleanEnv(name string) string {
	val := os.Getenv(name)
	return strings.TrimSpace(val)
}

func normalizeConn(val string) string {
	val = strings.ToLower(strings.TrimSpace(val))
	switch val {
	case "usb", "wifi":
		return val
	default:
		return "auto"
	}
}
