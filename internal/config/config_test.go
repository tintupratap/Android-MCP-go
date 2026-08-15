package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigLoadSaveAtomic(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "android-mcp.json")

	// Test load missing file
	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("expected no error loading missing config, got: %v", err)
	}
	if cfg.Port != 5555 || cfg.Connection != "auto" {
		t.Fatalf("expected default config, got %+v", cfg)
	}

	// Test save
	now := time.Now().Truncate(time.Second)
	cfg.LastIP = "192.168.1.50"
	cfg.DeviceSerial = "TEST_SERIAL_123"
	cfg.DeviceModel = "TestPhone"
	cfg.Port = 5555
	cfg.Connection = "wifi"
	cfg.LastSeen = now
	cfg.LastSuccessfulConnection = now

	if err := SaveConfigToPath(configPath, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config file was not created")
	}

	// Test reload
	reloaded, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if reloaded.LastIP != "192.168.1.50" || reloaded.DeviceSerial != "TEST_SERIAL_123" || reloaded.Connection != "wifi" {
		t.Fatalf("reloaded config mismatch: %+v", reloaded)
	}
}

func TestMalformedConfigRecovery(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "android-mcp.json")

	// Write invalid JSON
	if err := os.WriteFile(configPath, []byte("{invalid json..."), 0644); err != nil {
		t.Fatalf("failed to write invalid json: %v", err)
	}

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("expected graceful fallback, got error: %v", err)
	}
	if cfg.Port != 5555 {
		t.Fatalf("expected default config on malformed file, got %+v", cfg)
	}

	// Check backup file exists
	backupPath := configPath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("expected backup file %s to be created", backupPath)
	}
}

func TestScrcpyConfigLoad(t *testing.T) {
	tempDir := t.TempDir()
	scrcpyPath := filepath.Join(tempDir, "scrcpy.json")

	// Missing file
	scrcpyCfg, err := LoadScrcpyConfigFromPath(scrcpyPath)
	if err != nil || scrcpyCfg != nil {
		t.Fatalf("expected nil scrcpy config for missing file, got %+v, err=%v", scrcpyCfg, err)
	}

	// Write valid scrcpy.json
	data := []byte(`{
		"last_ip": "192.168.1.100",
		"device_serial": "SCRCPY_SERIAL",
		"device_model": "SOG09",
		"port": 5555
	}`)
	if err := os.WriteFile(scrcpyPath, data, 0644); err != nil {
		t.Fatalf("failed to write scrcpy json: %v", err)
	}

	scrcpyCfg, err = LoadScrcpyConfigFromPath(scrcpyPath)
	if err != nil || scrcpyCfg == nil {
		t.Fatalf("failed to load scrcpy config: %v", err)
	}
	if scrcpyCfg.LastIP != "192.168.1.100" || scrcpyCfg.DeviceSerial != "SCRCPY_SERIAL" {
		t.Fatalf("unexpected scrcpy config contents: %+v", scrcpyCfg)
	}
}
