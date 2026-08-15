package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"android-mcp-go/internal/logging"
)

// Config represents persistent device state in ~/.android-mcp/android-mcp.json
type Config struct {
	LastIP                   string    `json:"last_ip,omitempty"`
	DeviceSerial             string    `json:"device_serial,omitempty"`
	DeviceModel              string    `json:"device_model,omitempty"`
	Port                     int       `json:"port,omitempty"`
	Connection               string    `json:"connection,omitempty"` // "auto", "usb", "wifi"
	LastSeen                 time.Time `json:"last_seen,omitempty"`
	LastSuccessfulConnection time.Time `json:"last_successful_connection,omitempty"`
	WiFiEnabled              bool      `json:"wifi_enabled"`
	UBSSBootstrapEnabled     bool      `json:"usb_bootstrap_enabled"`
}

// ScrcpyConfig represents external state from ~/.scrcpy/scrcpy.json
type ScrcpyConfig struct {
	LastIP       string `json:"last_ip,omitempty"`
	DeviceSerial string `json:"device_serial,omitempty"`
	DeviceModel  string `json:"device_model,omitempty"`
	Port         int    `json:"port,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Port:                 5555,
		Connection:           "auto",
		WiFiEnabled:          true,
		UBSSBootstrapEnabled: true,
	}
}

// GetAndroidMCPDir returns ~/.android-mcp
func GetAndroidMCPDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	dir := filepath.Join(home, ".android-mcp")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config dir %s: %w", dir, err)
	}
	return dir, nil
}

// GetAndroidMCPConfigPath returns ~/.android-mcp/android-mcp.json
func GetAndroidMCPConfigPath() (string, error) {
	dir, err := GetAndroidMCPDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "android-mcp.json"), nil
}

// GetScrcpyConfigPath returns ~/.scrcpy/scrcpy.json
func GetScrcpyConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".scrcpy", "scrcpy.json"), nil
}

// LoadConfig loads ~/.android-mcp/android-mcp.json or returns default if not present/corrupted
func LoadConfig() (*Config, error) {
	path, err := GetAndroidMCPConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	return LoadConfigFromPath(path)
}

// LoadConfigFromPath loads config from specified path
func LoadConfigFromPath(path string) (*Config, error) {
	cfg := DefaultConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		logging.Debugf("No existing android-mcp config at %s", path)
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logging.Warnf("Failed to read config file %s: %v", path, err)
		return cfg, nil
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		logging.Warnf("Malformed json in config %s: %v. Using defaults.", path, err)
		// Preserve corrupted file as backup
		backupPath := path + ".bak"
		_ = os.WriteFile(backupPath, data, 0600)
		logging.Warnf("Backed up malformed config to %s", backupPath)
		return DefaultConfig(), nil
	}

	if cfg.Port == 0 {
		cfg.Port = 5555
	}
	if cfg.Connection == "" {
		cfg.Connection = "auto"
	}

	logging.Infof("Loaded persistent device state from %s (last_ip=%s, device_serial=%s)", path, cfg.LastIP, cfg.DeviceSerial)
	return cfg, nil
}

// SaveConfig atomically saves configuration to ~/.android-mcp/android-mcp.json
func SaveConfig(cfg *Config) error {
	path, err := GetAndroidMCPConfigPath()
	if err != nil {
		return err
	}
	return SaveConfigToPath(path, cfg)
}

// SaveConfigToPath atomically saves config to specified path
func SaveConfigToPath(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "android-mcp-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file for config: %w", err)
	}
	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write config to temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to fsync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp config file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to atomically rename config file: %w", err)
	}

	logging.Infof("State saved to %s", path)
	return nil
}

// LoadScrcpyConfig loads external state from ~/.scrcpy/scrcpy.json
func LoadScrcpyConfig() (*ScrcpyConfig, error) {
	path, err := GetScrcpyConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadScrcpyConfigFromPath(path)
}

// LoadScrcpyConfigFromPath loads scrcpy config from specified path
func LoadScrcpyConfigFromPath(path string) (*ScrcpyConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logging.Warnf("Failed to read scrcpy config %s: %v", path, err)
		return nil, nil
	}

	var cfg ScrcpyConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		logging.Warnf("Malformed json in scrcpy config %s: %v", path, err)
		return nil, nil
	}

	if cfg.Port == 0 {
		cfg.Port = 5555
	}

	logging.Infof("Read external discovery state from %s (last_ip=%s, serial=%s)", path, cfg.LastIP, cfg.DeviceSerial)
	return &cfg, nil
}
