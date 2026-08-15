package device

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/discovery"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
	"github.com/tintupratap/Android-MCP-go/internal/notification"
	"github.com/tintupratap/Android-MCP-go/internal/scrcpy"
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
	scrcpyMgr    *scrcpy.Manager
	preference   DevicePreference
	activeDevice *Device
	state        ConnectionState
}

func NewDeviceManager(client *adb.Client, notifier notification.Notifier, pref DevicePreference, scrcpyMgr *scrcpy.Manager) *DefaultDeviceManager {
	if client == nil {
		client = adb.NewClient("")
	}
	if notifier == nil {
		notifier = notification.NewNotifier()
	}
	if scrcpyMgr == nil {
		scrcpyMgr, _ = scrcpy.NewManager(notifier)
	}

	bs := discovery.NewWirelessBootstrapper(client, notifier)

	return &DefaultDeviceManager{
		adbClient:    client,
		notifier:     notifier,
		bootstrapper: bs,
		scrcpyMgr:    scrcpyMgr,
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

	cfg, errCfg := config.LoadConfig()
	if errCfg == nil {
		if connType == "wifi" {
			parts := strings.Split(target, ":")
			cfg.Device.LastIP = parts[0]
		}
		cfg.Device.Serial = target
		cfg.Device.Model = model
		cfg.Device.Connection = connType
		cfg.LastSeen = time.Now()
		cfg.LastSuccessfulConnection = time.Now()
		_ = config.SaveConfig(cfg)
	}

	if m.scrcpyMgr != nil && os.Getenv("ANDROID_MCP_SCRCPY") != "false" {
		go func(ser, mod string) {
			title := fmt.Sprintf("Android-MCP — %s (%s)", mod, ser)
			_ = m.scrcpyMgr.Launch(context.Background(), ser, title)
		}(target, model)
	}

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
	if err == nil {
		if cfg.Device.LastIP != "" {
			port := cfg.Device.Port
			if port <= 0 {
				port = 5555
			}
			wifiAddr := adb.FormatWiFiSerial(cfg.Device.LastIP, port)
			logging.Infof("Attempting to connect to cached WiFi device %s...", wifiAddr)
			dev, err := m.connectLocked(ctx, wifiAddr)
			if err == nil {
				cfg.LastSeen = time.Now()
				cfg.LastSuccessfulConnection = time.Now()
				_ = config.SaveConfig(cfg)
				return dev, nil
			}
			logging.Warnf("Cached WiFi device %s unavailable", wifiAddr)
		} else if cfg.Device.Serial != "" {
			logging.Infof("Attempting to connect to cached USB serial device %s...", cfg.Device.Serial)
			dev, err := m.connectLocked(ctx, cfg.Device.Serial)
			if err == nil {
				return dev, nil
			}
			logging.Warnf("Cached USB serial device %s unavailable", cfg.Device.Serial)
		}
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
