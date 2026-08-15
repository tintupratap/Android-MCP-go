package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/logging"
)

type DeviceConfig struct {
	LastIP     string `json:"last_ip,omitempty"`
	Serial     string `json:"serial,omitempty"`
	Model      string `json:"model,omitempty"`
	Port       int    `json:"port,omitempty"`
	Connection string `json:"connection,omitempty"` // "auto", "usb", "wifi"
}

type ScrcpyPreferences struct {
	Enabled            bool   `json:"enabled"`
	AutoStart          bool   `json:"auto_start"`
	VideoSource        string `json:"video_source,omitempty"`
	DisplayID          string `json:"display_id,omitempty"`
	CameraID           string `json:"camera_id,omitempty"`
	CameraSize         string `json:"camera_size,omitempty"`
	CameraFPS          string `json:"camera_fps,omitempty"`
	CameraHighSpeed    bool   `json:"camera_high_speed,omitempty"`
	VideoCodec         string `json:"video_codec,omitempty"`
	VideoEncoder       string `json:"video_encoder,omitempty"`
	VideoBitrate       string `json:"video_bitrate,omitempty"`
	MaxResolutionSize int    `json:"max_resolution_size,omitempty"`
	AudioSource        string `json:"audio_source,omitempty"`
	AudioCodec         string `json:"audio_codec,omitempty"`
	AudioEncoder       string `json:"audio_encoder,omitempty"`
	AudioBitrate       string `json:"audio_bitrate,omitempty"`
	TurnScreenOff      bool   `json:"turn_screen_off,omitempty"`
	StayAwake          bool   `json:"stay_awake,omitempty"`
	PowerOffOnClose    bool   `json:"power_off_on_close,omitempty"`
	RenderDriver       string `json:"render_driver,omitempty"`
}

type PlatformToolsConfig struct {
	Managed     bool      `json:"managed"`
	Path        string    `json:"path,omitempty"`
	Version     string    `json:"version,omitempty"`
	Source      string    `json:"source,omitempty"`
	InstalledAt time.Time `json:"installed_at,omitempty"`
}

type ManagedScrcpyConfig struct {
	Managed     bool      `json:"managed"`
	Path        string    `json:"path,omitempty"`
	Version     string    `json:"version,omitempty"`
	Release     string    `json:"release,omitempty"`
	Source      string    `json:"source,omitempty"`
	InstalledAt time.Time `json:"installed_at,omitempty"`
}

type NotificationsConfig struct {
	Enabled bool `json:"enabled"`
}

type MigrationInfo struct {
	ScrcpyWirelessImported bool      `json:"scrcpy_wireless_imported"`
	ImportedAt             time.Time `json:"imported_at,omitempty"`
}

// Config represents persistent device state and preferences in ~/.android-mcp/android-mcp.json
type Config struct {
	Version       int                  `json:"version"`
	Device        DeviceConfig         `json:"device"`
	Scrcpy        ScrcpyPreferences    `json:"scrcpy"`
	PlatformTools *PlatformToolsConfig `json:"platform_tools,omitempty"`
	ManagedScrcpy *ManagedScrcpyConfig `json:"managed_scrcpy,omitempty"`
	Notifications NotificationsConfig  `json:"notifications"`
	Migration     *MigrationInfo       `json:"migration,omitempty"`

	// Legacy backward compatibility fields (auto-synced with Device struct)
	LastIP                   string    `json:"last_ip,omitempty"`
	DeviceSerial             string    `json:"device_serial,omitempty"`
	DeviceModel              string    `json:"device_model,omitempty"`
	Port                     int       `json:"port,omitempty"`
	Connection               string    `json:"connection,omitempty"`
	LastSeen                 time.Time `json:"last_seen,omitempty"`
	LastSuccessfulConnection time.Time `json:"last_successful_connection,omitempty"`
	WiFiEnabled              bool      `json:"wifi_enabled"`
	UBSSBootstrapEnabled     bool      `json:"usb_bootstrap_enabled"`
}

// LegacyScrcpyImport is used ONLY for one-time migration from legacy ~/.scrcpy/scrcpy.json
type LegacyScrcpyImport struct {
	LastIP       string `json:"last_ip"`
	DeviceSerial string `json:"device_serial"`
	DeviceModel  string `json:"device_model"`
	Port         int    `json:"port"`
	VideoSource  string `json:"video_source"`
	DisplayID    string `json:"display_id"`
	VideoCodec   string `json:"video_codec"`
	VideoBitrate string `json:"video_bitrate"`
	AudioSource  string `json:"audio_source"`
	AudioCodec   string `json:"audio_codec"`
	AudioBitrate string `json:"audio_bitrate"`
	StayAwake    bool   `json:"stay_awake"`
	RenderDriver string `json:"render_driver"`
}

func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Device: DeviceConfig{
			Port:       5555,
			Connection: "auto",
		},
		Scrcpy: ScrcpyPreferences{
			Enabled:      true,
			AutoStart:    true,
			VideoSource:  "display",
			VideoCodec:   "h265",
			VideoBitrate: "4M",
			AudioSource:  "playback",
			AudioCodec:   "opus",
			AudioBitrate: "128K",
			StayAwake:    true,
		},
		Notifications: NotificationsConfig{
			Enabled: true,
		},
		Port:                 5555,
		Connection:           "auto",
		WiFiEnabled:          true,
		UBSSBootstrapEnabled: true,
	}
}

func GetAndroidMCPConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".android-mcp"), nil
}

func GetAndroidMCPConfigPath() (string, error) {
	dir, err := GetAndroidMCPConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "android-mcp.json"), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetAndroidMCPConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			_ = PerformOneTimeMigration(cfg)
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from config file %s: %w", configPath, err)
	}

	// Sync legacy flat fields to Device struct
	if cfg.Device.LastIP == "" && cfg.LastIP != "" {
		cfg.Device.LastIP = cfg.LastIP
	}
	if cfg.Device.Serial == "" && cfg.DeviceSerial != "" {
		cfg.Device.Serial = cfg.DeviceSerial
	}
	if cfg.Device.Model == "" && cfg.DeviceModel != "" {
		cfg.Device.Model = cfg.DeviceModel
	}
	if cfg.Device.Port == 0 && cfg.Port != 0 {
		cfg.Device.Port = cfg.Port
	}
	if cfg.Device.Connection == "" && cfg.Connection != "" {
		cfg.Device.Connection = cfg.Connection
	}

	// Perform one-time migration if legacy file exists and has not been imported yet
	if cfg.Migration == nil || !cfg.Migration.ScrcpyWirelessImported {
		if PerformOneTimeMigration(cfg) {
			_ = SaveConfig(cfg)
		}
	}

	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	configPath, err := GetAndroidMCPConfigPath()
	if err != nil {
		return err
	}

	// Sync Device struct back to flat fields for legacy code compatibility
	cfg.LastIP = cfg.Device.LastIP
	cfg.DeviceSerial = cfg.Device.Serial
	cfg.DeviceModel = cfg.Device.Model
	cfg.Port = cfg.Device.Port
	cfg.Connection = cfg.Device.Connection

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	tmpFile := configPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file %s: %w", tmpFile, err)
	}

	if err := os.Rename(tmpFile, configPath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to atomically replace config file %s: %w", configPath, err)
	}

	logging.Debugf("Saved state to %s", configPath)
	return nil
}

// PerformOneTimeMigration imports old ~/.scrcpy/scrcpy.json into android-mcp.json if present
func PerformOneTimeMigration(cfg *Config) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	legacyFile := filepath.Join(home, ".scrcpy", "scrcpy.json")
	data, err := os.ReadFile(legacyFile)
	if err != nil {
		return false
	}

	var legacy LegacyScrcpyImport
	if err := json.Unmarshal(data, &legacy); err != nil {
		return false
	}

	logging.Infof("Migrating legacy state from %s into ~/.android-mcp/android-mcp.json...", legacyFile)

	if legacy.LastIP != "" && cfg.Device.LastIP == "" {
		cfg.Device.LastIP = legacy.LastIP
	}
	if legacy.DeviceSerial != "" && cfg.Device.Serial == "" {
		cfg.Device.Serial = legacy.DeviceSerial
	}
	if legacy.DeviceModel != "" && cfg.Device.Model == "" {
		cfg.Device.Model = legacy.DeviceModel
	}
	if legacy.Port > 0 && cfg.Device.Port == 0 {
		cfg.Device.Port = legacy.Port
	}
	if legacy.VideoCodec != "" {
		cfg.Scrcpy.VideoCodec = legacy.VideoCodec
	}
	if legacy.VideoBitrate != "" {
		cfg.Scrcpy.VideoBitrate = legacy.VideoBitrate
	}
	if legacy.DisplayID != "" {
		cfg.Scrcpy.DisplayID = legacy.DisplayID
	}
	if legacy.AudioSource != "" {
		cfg.Scrcpy.AudioSource = legacy.AudioSource
	}
	if legacy.AudioCodec != "" {
		cfg.Scrcpy.AudioCodec = legacy.AudioCodec
	}
	if legacy.AudioBitrate != "" {
		cfg.Scrcpy.AudioBitrate = legacy.AudioBitrate
	}
	if legacy.StayAwake {
		cfg.Scrcpy.StayAwake = legacy.StayAwake
	}
	if legacy.RenderDriver != "" {
		cfg.Scrcpy.RenderDriver = legacy.RenderDriver
	}

	cfg.Migration = &MigrationInfo{
		ScrcpyWirelessImported: true,
		ImportedAt:             time.Now(),
	}

	return true
}

// Deprecated: ScrcpyConfig represents legacy state struct
type ScrcpyConfig struct {
	LastIP       string `json:"last_ip,omitempty"`
	DeviceSerial string `json:"device_serial,omitempty"`
	DeviceModel  string `json:"device_model,omitempty"`
	Port         int    `json:"port,omitempty"`
	VideoSource  string `json:"video_source,omitempty"`
	DisplayID    string `json:"display_id,omitempty"`
	VideoCodec   string `json:"video_codec,omitempty"`
	VideoBitrate string `json:"video_bitrate,omitempty"`
	AudioSource  string `json:"audio_source,omitempty"`
	AudioCodec   string `json:"audio_codec,omitempty"`
	AudioBitrate string `json:"audio_bitrate,omitempty"`
	StayAwake    bool   `json:"stay_awake,omitempty"`
	RenderDriver string `json:"render_driver,omitempty"`
}

// Deprecated: GetScrcpyConfigPath is retained ONLY for diagnostic reporting
func GetScrcpyConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".scrcpy", "scrcpy.json"), nil
}

// Deprecated: LoadScrcpyConfig returns state from android-mcp.json
func LoadScrcpyConfig() (*ScrcpyConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return &ScrcpyConfig{
		LastIP:       cfg.Device.LastIP,
		DeviceSerial: cfg.Device.Serial,
		DeviceModel:  cfg.Device.Model,
		Port:         cfg.Device.Port,
		VideoCodec:   cfg.Scrcpy.VideoCodec,
		VideoBitrate: cfg.Scrcpy.VideoBitrate,
		DisplayID:    cfg.Scrcpy.DisplayID,
		AudioSource:  cfg.Scrcpy.AudioSource,
		AudioCodec:   cfg.Scrcpy.AudioCodec,
		AudioBitrate: cfg.Scrcpy.AudioBitrate,
		StayAwake:    cfg.Scrcpy.StayAwake,
		RenderDriver: cfg.Scrcpy.RenderDriver,
	}, nil
}
