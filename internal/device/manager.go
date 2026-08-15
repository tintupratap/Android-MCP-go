package device

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/discovery"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
	"github.com/tintupratap/Android-MCP-go/internal/notification"
)

type ConnectionState string

const (
	StateNoDevice          ConnectionState = "NoDevice"
	StateUSBDetected       ConnectionState = "USBDetected"
	StateBootstrappingWiFi ConnectionState = "BootstrappingWiFi"
	StateWiFiConnected     ConnectionState = "WiFiConnected"
	StateUSBConnected      ConnectionState = "USBConnected"
	StateBootstrapFailed   ConnectionState = "BootstrapFailed"
)

type Device struct {
	Serial     string
	Model      string
	Connection string // "wifi", "usb", "emulator"
}

type DevicePreference struct {
	Connection string // "auto", "usb", "wifi"
	Serial     string
	Source     string
}

type DeviceManager interface {
	Resolve(ctx context.Context) (*Device, error)
	List(ctx context.Context) ([]adb.ADBDevice, error)
	Connect(ctx context.Context, serial string) (*Device, error)
	Disconnect(ctx context.Context) error
	GetActiveDevice() *Device
	SetPreference(pref DevicePreference)
	State() ConnectionState
}

type DefaultDeviceManager struct {
	mu           sync.Mutex
	adbClient    *adb.Client
	bootstrapper *discovery.WirelessBootstrapper
	notifier     notification.Notifier
	preference   DevicePreference
	activeDevice *Device
	state        ConnectionState
}

func NewDeviceManager(client *adb.Client, notifier notification.Notifier, pref DevicePreference) *DefaultDeviceManager {
	if client == nil {
		client = adb.NewClient("")
	}
	if notifier == nil {
		notifier = notification.NewNotifier()
	}
	return &DefaultDeviceManager{
		adbClient:    client,
		bootstrapper: discovery.NewWirelessBootstrapper(client, notifier),
		notifier:     notifier,
		preference:   pref,
		state:        StateNoDevice,
	}
}

func (m *DefaultDeviceManager) SetPreference(pref DevicePreference) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.preference = pref
}

func (m *DefaultDeviceManager) GetActiveDevice() *Device {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activeDevice
}

func (m *DefaultDeviceManager) State() ConnectionState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

func (m *DefaultDeviceManager) List(ctx context.Context) ([]adb.ADBDevice, error) {
	return m.adbClient.ListDevices(ctx)
}

func (m *DefaultDeviceManager) Disconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeDevice != nil && strings.Contains(m.activeDevice.Serial, ":") {
		_ = m.adbClient.Disconnect(ctx, m.activeDevice.Serial)
	}
	m.activeDevice = nil
	m.state = StateNoDevice
	return nil
}

func (m *DefaultDeviceManager) Connect(ctx context.Context, serial string) (*Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.connectLocked(ctx, serial)
}

func (m *DefaultDeviceManager) connectLocked(ctx context.Context, serial string) (*Device, error) {
	target := serial
	if strings.Contains(target, ":") {
		target = adb.FormatWiFiSerial(target, 5555)
		ok, err := m.adbClient.Connect(ctx, target)
		if err != nil || !ok {
			return nil, fmt.Errorf("failed to connect to WiFi device %s: %v", target, err)
		}
	}

	model, _ := m.adbClient.GetProp(ctx, target, "ro.product.model")
	if model == "" {
		model = target
	}

	connType := "usb"
	if strings.Contains(target, ":") {
		connType = "wifi"
		m.state = StateWiFiConnected
	} else if strings.HasPrefix(target, "emulator-") {
		connType = "emulator"
		m.state = StateUSBConnected
	} else {
		m.state = StateUSBConnected
	}

	m.activeDevice = &Device{
		Serial:     target,
		Model:      model,
		Connection: connType,
	}

	logging.Infof("Connected to target device: %s (%s, mode=%s)", model, target, connType)
	return m.activeDevice, nil
}

func (m *DefaultDeviceManager) Resolve(ctx context.Context) (*Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if existing active device is still online
	if m.activeDevice != nil {
		devices, err := m.adbClient.ListDevices(ctx)
		if err == nil {
			for _, d := range devices {
				if d.Serial == m.activeDevice.Serial && d.State == "device" {
					return m.activeDevice, nil
				}
			}
		}
		logging.Warnf("Active device %s disconnected or offline. Re-resolving...", m.activeDevice.Serial)
		m.activeDevice = nil
		m.state = StateNoDevice
	}

	// Priority 1: Explicit preference (CLI flags or env vars)
	if m.preference.Serial != "" {
		target := m.preference.Serial
		if m.preference.Connection == "wifi" || strings.Contains(target, ":") {
			target = adb.FormatWiFiSerial(target, 5555)
		}
		dev, err := m.connectLocked(ctx, target)
		if err == nil {
			return dev, nil
		}
		logging.Warnf("Explicit device target %s connection failed: %v", target, err)
	}

	// Priority 2: Persistent state ~/.android-mcp/android-mcp.json
	cfg, err := config.LoadConfig()
	if err == nil && cfg.LastIP != "" {
		wifiAddr := adb.FormatWiFiSerial(cfg.LastIP, cfg.Port)
		logging.Infof("Attempting to connect to cached android-mcp WiFi device %s...", wifiAddr)
		dev, err := m.connectLocked(ctx, wifiAddr)
		if err == nil {
			// Update last_seen
			cfg.LastSeen = time.Now()
			cfg.LastSuccessfulConnection = time.Now()
			_ = config.SaveConfig(cfg)
			return dev, nil
		}
		logging.Warnf("Cached android-mcp WiFi device %s unavailable", wifiAddr)
	}

	// Priority 3: External scrcpy state ~/.scrcpy/scrcpy.json
	scrcpyCfg, err := config.LoadScrcpyConfig()
	if err == nil && scrcpyCfg != nil && scrcpyCfg.LastIP != "" {
		scrcpyAddr := adb.FormatWiFiSerial(scrcpyCfg.LastIP, scrcpyCfg.Port)
		logging.Infof("Attempting to connect to cached scrcpy WiFi device %s...", scrcpyAddr)
		dev, err := m.connectLocked(ctx, scrcpyAddr)
		if err == nil {
			return dev, nil
		}
		logging.Warnf("Cached scrcpy WiFi device %s unavailable", scrcpyAddr)
	}

	// Priority 4: Auto-pick connected ADB devices
	devices, err := m.adbClient.ListDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list ADB devices: %w", err)
	}

	var onlineDevs []adb.ADBDevice
	for _, d := range devices {
		if d.State == "device" {
			onlineDevs = append(onlineDevs, d)
		}
	}

	if len(onlineDevs) == 0 {
		m.state = StateNoDevice
		return nil, fmt.Errorf("no online Android devices found. Connect USB or start WiFi ADB")
	}

	// 4a. Check if any WiFi device is already online
	for _, d := range onlineDevs {
		if d.IsWiFi() {
			return m.connectLocked(ctx, d.Serial)
		}
	}

	// 4b. Check for physical USB devices for wireless bootstrap
	var usbPhysical []adb.ADBDevice
	for _, d := range onlineDevs {
		if d.IsUSB() {
			usbPhysical = append(usbPhysical, d)
		}
	}

	if len(usbPhysical) > 0 {
		targetUSB := usbPhysical[0]
		m.state = StateUSBDetected

		// Attempt wireless bootstrap USB -> WiFi
		m.state = StateBootstrappingWiFi
		res, err := m.bootstrapper.BootstrapUSBToWiFi(ctx, targetUSB, 5555)
		if err == nil && res != nil {
			dev, err := m.connectLocked(ctx, res.WiFiAddress)
			if err == nil {
				return dev, nil
			}
		}

		// Fall back to direct USB connection if WiFi bootstrap fails
		m.state = StateBootstrapFailed
		logging.Warnf("WiFi bootstrap failed (%v). Falling back to direct USB transport.", err)
		return m.connectLocked(ctx, targetUSB.Serial)
	}

	// 4c. Fall back to emulator if present
	for _, d := range onlineDevs {
		if d.IsEmulator() {
			return m.connectLocked(ctx, d.Serial)
		}
	}

	m.state = StateNoDevice
	return nil, fmt.Errorf("no usable Android devices resolved")
}
