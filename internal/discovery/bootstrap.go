package discovery

import (
	"context"
	"fmt"
	"time"

	"android-mcp-go/internal/adb"
	"android-mcp-go/internal/config"
	"android-mcp-go/internal/logging"
	"android-mcp-go/internal/notification"
)

type WirelessBootstrapper struct {
	adbClient *adb.Client
	notifier  notification.Notifier
}

func NewWirelessBootstrapper(client *adb.Client, notifier notification.Notifier) *WirelessBootstrapper {
	if notifier == nil {
		notifier = notification.NewNotifier()
	}
	return &WirelessBootstrapper{
		adbClient: client,
		notifier:  notifier,
	}
}

type BootstrapResult struct {
	WiFiAddress  string
	DeviceSerial string
	DeviceModel  string
	Port         int
}

// BootstrapUSBToWiFi converts a USB connected device to a verified WiFi ADB connection
func (b *WirelessBootstrapper) BootstrapUSBToWiFi(ctx context.Context, usbDev adb.ADBDevice, targetPort int) (*BootstrapResult, error) {
	if targetPort <= 0 {
		targetPort = 5555
	}

	logging.Infof("Found USB device: %s (%s)", usbDev.Serial, usbDev.Model)

	// Step 1: Discover device IP before/after enabling TCP/IP
	ip, err := b.adbClient.GetDeviceIP(ctx, usbDev.Serial)
	if err != nil {
		return nil, fmt.Errorf("failed to discover device IP: %w", err)
	}
	logging.Infof("Discovered WiFi address: %s", ip)

	// Step 2: Enable TCP/IP on device
	logging.Infof("Enabling ADB TCP/IP mode on port %d...", targetPort)
	if err := b.adbClient.EnableTCPIP(ctx, usbDev.Serial, targetPort); err != nil {
		return nil, fmt.Errorf("failed to enable ADB TCP/IP: %w", err)
	}

	// Step 3: Wait for adbd to restart in TCP mode
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(2500 * time.Millisecond):
	}

	// Step 4: Connect to IP:port
	wifiAddr := adb.FormatWiFiSerial(ip, targetPort)
	logging.Infof("Connecting to %s...", wifiAddr)
	ok, err := b.adbClient.Connect(ctx, wifiAddr)
	if err != nil || !ok {
		return nil, fmt.Errorf("failed to connect to wireless ADB at %s: %v", wifiAddr, err)
	}

	// Step 5: Verify wireless connection with lightweight getprop command
	model, err := b.adbClient.GetProp(ctx, wifiAddr, "ro.product.model")
	if err != nil || model == "" {
		model = usbDev.Model
	}
	if model == "" {
		model = "Android Device"
	}

	logging.Infof("WiFi ADB connection verified for %s (%s)", model, wifiAddr)

	// Step 6: Update persistent state
	cfg, _ := config.LoadConfig()
	cfg.LastIP = ip
	cfg.DeviceSerial = usbDev.Serial
	cfg.DeviceModel = model
	cfg.Port = targetPort
	cfg.Connection = "wifi"
	cfg.LastSeen = time.Now()
	cfg.LastSuccessfulConnection = time.Now()

	if err := config.SaveConfig(cfg); err != nil {
		logging.Warnf("Failed to persist state: %v", err)
	} else {
		logging.Infof("State saved to ~/.android-mcp/android-mcp.json")
	}

	// Step 7: Send desktop notification to user
	notifMsg := fmt.Sprintf("%s is connected over WiFi at %s. USB can be disconnected.", model, wifiAddr)
	_ = b.notifier.Notify("Android-MCP", notifMsg)
	logging.Infof("USB can now be disconnected.")

	return &BootstrapResult{
		WiFiAddress:  wifiAddr,
		DeviceSerial: usbDev.Serial,
		DeviceModel:  model,
		Port:         targetPort,
	}, nil
}
