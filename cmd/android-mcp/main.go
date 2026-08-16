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
	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/device"
	"github.com/tintupratap/Android-MCP-go/internal/doctor"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
	"github.com/tintupratap/Android-MCP-go/internal/mcp"
	"github.com/tintupratap/Android-MCP-go/internal/notification"
	"github.com/tintupratap/Android-MCP-go/internal/platformtools"
	"github.com/tintupratap/Android-MCP-go/internal/scrcpy"
	"github.com/tintupratap/Android-MCP-go/internal/skills"
)

var Version = "0.5.0"

func main() {
	// Check for subcommand arguments first (e.g. android-mcp doctor, android-mcp status, android-mcp platform-tools, android-mcp scrcpy, android-mcp skills)
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
		case "scrcpy":
			runScrcpyCmd(os.Args[2:])
			os.Exit(0)
		case "skills":
			runSkillsCmd(os.Args[2:])
			os.Exit(0)
		}
	}

	var (
		deviceFlag        string
		connectionFlag    string
		wifiFlag          string
		usbFlag           string
		noScrcpyFlag      bool
		noRelaunchFlag    bool
		noAlwaysOnTopFlag bool
		quietFlag         bool
		debugFlag         bool
		showVersion       bool
		showDoctor        bool
		showStatus        bool
	)

	flag.StringVar(&deviceFlag, "device", "", "ADB device serial or host:port")
	flag.StringVar(&connectionFlag, "connection", "", "Preferred device connection type (auto, usb, wifi)")
	flag.StringVar(&wifiFlag, "wifi", "", "Use WiFi ADB. Accepts HOST or HOST:PORT (defaults to port 5555)")
	flag.StringVar(&usbFlag, "usb", "", "Use USB ADB. Optionally provide a specific device serial")
	flag.BoolVar(&noScrcpyFlag, "no-scrcpy", false, "Disable automatic scrcpy live view display window startup")
	flag.BoolVar(&noRelaunchFlag, "no-scrcpy-relaunch", false, "Disable tool-call scrcpy auto-relaunch after user window close")
	flag.BoolVar(&noAlwaysOnTopFlag, "no-always-on-top", false, "Disable --always-on-top window flag when launching scrcpy")
	flag.BoolVar(&quietFlag, "quiet", false, "Disable desktop notifications")
	flag.BoolVar(&debugFlag, "debug", false, "Enable verbose debug logging and activity desktop notifications")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.BoolVar(&showDoctor, "doctor", false, "Run doctor diagnostic check and exit")
	flag.BoolVar(&showStatus, "status", false, "Run status check and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("Android-MCP-go v%s\n", Version)
		return
	}

	if showDoctor {
		runDoctorCmd()
		return
	}

	if showStatus {
		runStatusCmd()
		return
	}

	appCfg, errLoad := config.LoadConfig()
	if errLoad == nil && appCfg != nil {
		var updated bool
		if noScrcpyFlag {
			appCfg.Scrcpy.Enabled = false
			appCfg.Scrcpy.AutoStart = false
			updated = true
		}
		if noRelaunchFlag {
			appCfg.Scrcpy.AutoRelaunchOnToolCall = false
			updated = true
		}
		if noAlwaysOnTopFlag {
			appCfg.Scrcpy.DisableAlwaysOnTop = true
			updated = true
		}
		if updated {
			_ = config.SaveConfig(appCfg)
		}
	}

	var notifier notification.Notifier
	notifLevel := notification.LevelNormal

	if quietFlag || os.Getenv("ANDROID_MCP_QUIET") == "true" {
		notifier = notification.NewNoopNotifier()
		notifLevel = notification.LevelSilent
	} else if debugFlag || os.Getenv("LOG_LEVEL") == "debug" {
		logging.SetLevel(logging.LevelDebug)
		notifier = notification.NewNotifier()
		notifLevel = notification.LevelDebug
	} else {
		logging.SetLevel(logging.LevelInfo)
		notifier = notification.NewNotifier()
		notifLevel = notification.LevelNormal
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

	// Ensure scrcpy is present if enabled
	scrcpyMgr, errScrcpy := scrcpy.NewManager(notifier)
	if errScrcpy == nil && os.Getenv("ANDROID_MCP_SCRCPY") != "false" && !noScrcpyFlag {
		if !scrcpyMgr.IsInstalled() {
			_, errEnsureScrcpy := scrcpyMgr.Ensure(ctx)
			if errEnsureScrcpy != nil {
				logging.Warnf("scrcpy ensure warning: %v.", errEnsureScrcpy)
			}
		}
	}

	pref := resolvePreference(deviceFlag, connectionFlag, wifiFlag, usbFlag)

	logging.Infof("Android-MCP-go v%s starting (connection=%s, source=%s, debug=%v)...", Version, pref.Connection, pref.Source, debugFlag)

	adbClient := adb.NewClient("")
	deviceMgr := device.NewDeviceManager(adbClient, notifier, pref, scrcpyMgr)

	resolverFunc := func(ctx context.Context) (string, string, error) {
		dev, err := deviceMgr.Resolve(ctx)
		if err != nil || dev == nil {
			return "", "", err
		}
		return dev.Serial, dev.Model, nil
	}

	liveViewMgr := scrcpy.NewLiveViewManager(scrcpyMgr, resolverFunc, adbClient, notifier, noScrcpyFlag)
	liveViewMgr.StartBackground(ctx)
	defer liveViewMgr.Stop()

	server := mcp.NewServer(deviceMgr, adbClient, os.Stdin, os.Stdout, actNotifier)
	server.SetLiveViewManager(liveViewMgr)

	if err := server.ListenAndServe(ctx); err != nil && ctx.Err() == nil {
		logging.Errorf("Server error: %v", err)
		os.Exit(1)
	}

	scrcpyMgr.StopAll()
	logging.Infof("Android-MCP server shut down cleanly.")
}

func runDoctorCmd() {
	ctx := context.Background()
	notifier := notification.NewNotifier()

	ptMgr, _ := platformtools.NewManager(notifier)
	if ptMgr != nil {
		_, _ = ptMgr.Ensure(ctx)
		_ = os.Setenv("ANDROID_MCP_ADB", ptMgr.ADBPath())
	}
	scrcpyMgr, _ := scrcpy.NewManager(notifier)
	if scrcpyMgr != nil && !scrcpyMgr.IsInstalled() {
		_, _ = scrcpyMgr.Ensure(ctx)
	}

	adbClient := adb.NewClient("")
	deviceMgr := device.NewDeviceManager(adbClient, notifier, device.DevicePreference{}, scrcpyMgr)

	rep := doctor.RunDoctor(ctx, deviceMgr, adbClient, ptMgr, scrcpyMgr)
	fmt.Print(rep.Format())
}

func runStatusCmd() {
	ctx := context.Background()
	notifier := notification.NewNotifier()

	ptMgr, _ := platformtools.NewManager(notifier)
	if ptMgr != nil && ptMgr.IsInstalled() {
		_ = os.Setenv("ANDROID_MCP_ADB", ptMgr.ADBPath())
	}
	scrcpyMgr, _ := scrcpy.NewManager(notifier)

	adbClient := adb.NewClient("")
	deviceMgr := device.NewDeviceManager(adbClient, notifier, device.DevicePreference{}, scrcpyMgr)

	out, code := doctor.RunStatus(ctx, deviceMgr, scrcpyMgr)
	fmt.Println(out)
	if code != 0 {
		os.Exit(code)
	}
}

func runScrcpyCmd(args []string) {
	ctx := context.Background()
	notifier := notification.NewNotifier()
	scrcpyMgr, err := scrcpy.NewManager(notifier)
	if err != nil {
		fmt.Printf("Error initializing scrcpy manager: %v\n", err)
		os.Exit(1)
	}

	action := "status"
	if len(args) > 0 {
		action = strings.ToLower(args[0])
	}

	switch action {
	case "capabilities":
		binCaps := scrcpy.DetectBinaryCapabilities(ctx, scrcpyMgr.BinaryPath())
		hostCaps := scrcpy.DetectHostCapabilities()
		adbClient := adb.NewClient("")
		devCaps := scrcpy.DetectDeviceCapabilities(ctx, adbClient, "")

		fmt.Println("scrcpy Subsystem Capabilities")
		fmt.Println("=============================")
		fmt.Printf("Binary Path:      %s\n", scrcpyMgr.BinaryPath())
		fmt.Printf("Version:          %s\n", binCaps.Version)
		fmt.Printf("Host Platform:    %s/%s\n", hostCaps.OS, hostCaps.Arch)
		fmt.Printf("Native Renderer:  %s\n", hostCaps.NativeRenderer)
		fmt.Println("Supported Flags:")
		fmt.Printf("  --render-driver: %v\n", binCaps.SupportsRenderDriver)
		fmt.Printf("  --video-codec:   %v\n", binCaps.SupportsVideoCodec)
		fmt.Printf("  --video-encoder: %v\n", binCaps.SupportsVideoEncoder)
		fmt.Printf("  --audio-source:  %v\n", binCaps.SupportsAudioSource)
		fmt.Printf("  --display-id:    %v\n", binCaps.SupportsDisplayID)
		if devCaps.Model != "" {
			fmt.Printf("Target Device:    %s (%s, SDK %s)\n", devCaps.Model, devCaps.Manufacturer, devCaps.SDKVersion)
		}
	case "profile":
		adbClient := adb.NewClient("")
		sysCaps := scrcpy.SystemCapabilities{
			Binary: scrcpy.DetectBinaryCapabilities(ctx, scrcpyMgr.BinaryPath()),
			Host:   scrcpy.DetectHostCapabilities(),
			Device: scrcpy.DetectDeviceCapabilities(ctx, adbClient, ""),
		}
		appCfg, _ := config.LoadConfig()
		var prefs *config.ScrcpyPreferences
		if appCfg != nil {
			prefs = &appCfg.Scrcpy
		}
		profile := scrcpy.ResolveOptimalProfile(prefs, sysCaps, "192.168.1.3:5555", "Android-MCP")
		fmt.Println("scrcpy Resolved Optimal Profile")
		fmt.Println("===============================")
		fmt.Printf("Profile Name:     %s\n", profile.Name)
		fmt.Printf("Video Codec:      %s\n", profile.VideoCodec)
		fmt.Printf("Renderer:         %s\n", profile.Renderer)
		fmt.Printf("Audio Mode:       %s\n", profile.Audio)
		fmt.Printf("Bitrate:          %s\n", profile.Bitrate)
		fmt.Printf("Optimized:        %v\n", profile.Optimized)
		fmt.Printf("Command Args:     %s\n", strings.Join(profile.Args, " "))
	case "status":
		fmt.Printf("scrcpy Path:       %s\n", scrcpyMgr.Path())
		fmt.Printf("scrcpy Binary:     %s\n", scrcpyMgr.BinaryPath())
		fmt.Printf("Installed:         %v\n", scrcpyMgr.IsInstalled())
		if scrcpyMgr.IsInstalled() {
			ver, _ := scrcpyMgr.GetVersion(ctx)
			fmt.Printf("Version:           %s\n", ver)
		}
	case "update", "reinstall":
		fmt.Println("Downloading official scrcpy release...")
		binPath, err := scrcpyMgr.Ensure(ctx)
		if err != nil {
			fmt.Printf("❌ Failed to update scrcpy: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ scrcpy ready at %s\n", binPath)
	case "start", "launch":
		adbClient := adb.NewClient("")
		devs, _ := adbClient.ListDevices(ctx)
		if len(devs) == 0 {
			fmt.Println("❌ No connected Android device found.")
			os.Exit(1)
		}
		serial := devs[0].Serial
		fmt.Printf("Launching scrcpy screen mirror for device %s...\n", serial)
		if err := scrcpyMgr.LaunchWithFallback(ctx, adbClient, serial, fmt.Sprintf("Android-MCP — %s", serial)); err != nil {
			fmt.Printf("❌ Launch failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Screen mirror window launched!")
	case "stop":
		scrcpyMgr.StopAll()
		fmt.Println("✓ Stopped active scrcpy screen mirror processes.")
	default:
		fmt.Printf("Unknown scrcpy subcommand: %s. Use capabilities, profile, status, update, start, or stop.\n", action)
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

func runSkillsCmd(args []string) {
	ctx := context.Background()
	skillsMgr, err := skills.NewManager()
	if err != nil {
		fmt.Printf("Error initializing skills manager: %v\n", err)
		os.Exit(1)
	}

	action := "list"
	if len(args) > 0 {
		action = strings.ToLower(args[0])
	}

	switch action {
	case "install", "update", "ensure":
		fmt.Println("Ensuring Android-MCP-go machine-readable skills...")
		if err := skillsMgr.EnsureSkills(ctx); err != nil {
			fmt.Printf("❌ Failed to install skills: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Skills ready at %s\n", skillsMgr.SkillsDir())
	case "list", "status":
		_ = skillsMgr.EnsureSkills(ctx)
		fmt.Print(skillsMgr.ListSkills())
	default:
		fmt.Printf("Unknown skills subcommand: %s. Use list or install.\n", action)
	}
}
