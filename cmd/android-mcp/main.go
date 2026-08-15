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
)

const Version = "0.2.0"

func main() {
	// Check for subcommand arguments first (e.g. android-mcp doctor, android-mcp status)
	if len(os.Args) > 1 {
		subcmd := strings.ToLower(os.Args[1])
		if subcmd == "doctor" {
			runDoctorCmd()
			os.Exit(0)
		} else if subcmd == "status" {
			runStatusCmd()
			return
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
	flag.BoolVar(&debugFlag, "debug", false, "Enable verbose debug logging")
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

	if debugFlag || os.Getenv("LOG_LEVEL") == "debug" {
		logging.SetLevel(logging.LevelDebug)
	} else {
		logging.SetLevel(logging.LevelInfo)
	}

	pref := resolvePreference(deviceFlag, connectionFlag, wifiFlag, usbFlag)

	logging.Infof("Android-MCP-go v%s starting (connection=%s, source=%s)...", Version, pref.Connection, pref.Source)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	adbClient := adb.NewClient("")
	notifier := notification.NewNotifier()
	deviceMgr := device.NewDeviceManager(adbClient, notifier, pref)

	server := mcp.NewServer(deviceMgr, adbClient, os.Stdin, os.Stdout)

	if err := server.ListenAndServe(ctx); err != nil && ctx.Err() == nil {
		logging.Errorf("Server error: %v", err)
		os.Exit(1)
	}

	logging.Infof("Android-MCP server shut down cleanly.")
}

func runDoctorCmd() {
	ctx := context.Background()
	adbClient := adb.NewClient("")
	notifier := notification.NewNotifier()
	deviceMgr := device.NewDeviceManager(adbClient, notifier, device.DevicePreference{})

	rep := doctor.RunDoctor(ctx, deviceMgr, adbClient)
	fmt.Print(rep.Format())
}

func runStatusCmd() {
	ctx := context.Background()
	adbClient := adb.NewClient("")
	notifier := notification.NewNotifier()
	deviceMgr := device.NewDeviceManager(adbClient, notifier, device.DevicePreference{})

	out, code := doctor.RunStatus(ctx, deviceMgr)
	fmt.Println(out)
	if code != 0 {
		os.Exit(code)
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
