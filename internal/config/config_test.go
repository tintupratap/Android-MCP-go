package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load default config: %v", err)
	}

	if cfg.Device.Port != 5555 {
		t.Fatalf("expected default port 5555, got %d", cfg.Device.Port)
	}

	cfg.Device.LastIP = "192.168.1.50"
	cfg.Device.Model = "Sony SOG09"
	cfg.Scrcpy.VideoCodec = "h265"

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to re-load config: %v", err)
	}

	if loaded.Device.LastIP != "192.168.1.50" || loaded.Device.Model != "Sony SOG09" {
		t.Fatalf("unexpected loaded device values: %+v", loaded.Device)
	}
	if loaded.Scrcpy.VideoCodec != "h265" {
		t.Fatalf("unexpected loaded scrcpy values: %+v", loaded.Scrcpy)
	}
}

func TestOneTimeMigration(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Create legacy ~/.scrcpy/scrcpy.json
	scrcpyDir := filepath.Join(tempHome, ".scrcpy")
	if err := os.MkdirAll(scrcpyDir, 0755); err != nil {
		t.Fatalf("failed to create legacy dir: %v", err)
	}

	legacyData := LegacyScrcpyImport{
		LastIP:       "192.168.1.100",
		DeviceSerial: "QV771A3JEE",
		DeviceModel:  "SOG09",
		Port:         5555,
		VideoCodec:   "hevc",
		VideoBitrate: "8M",
		StayAwake:    true,
	}

	data, err := json.Marshal(legacyData)
	if err != nil {
		t.Fatalf("failed to marshal legacy data: %v", err)
	}

	if err := os.WriteFile(filepath.Join(scrcpyDir, "scrcpy.json"), data, 0644); err != nil {
		t.Fatalf("failed to write legacy file: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config during migration test: %v", err)
	}

	if cfg.Migration == nil || !cfg.Migration.ScrcpyWirelessImported {
		t.Fatalf("expected migration to be marked complete")
	}

	if cfg.Device.LastIP != "192.168.1.100" || cfg.Device.Model != "SOG09" {
		t.Fatalf("unexpected migrated device values: %+v", cfg.Device)
	}
	if cfg.Scrcpy.VideoBitrate != "8M" {
		t.Fatalf("unexpected migrated scrcpy bitrate: %s", cfg.Scrcpy.VideoBitrate)
	}
}
