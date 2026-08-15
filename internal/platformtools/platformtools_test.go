package platformtools

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveOfficialURL(t *testing.T) {
	darwinURL, err := ResolveOfficialURL("darwin")
	if err != nil || darwinURL != OfficialDarwinURL {
		t.Fatalf("unexpected darwin url: %s, err=%v", darwinURL, err)
	}

	linuxURL, err := ResolveOfficialURL("linux")
	if err != nil || linuxURL != OfficialLinuxURL {
		t.Fatalf("unexpected linux url: %s, err=%v", linuxURL, err)
	}

	windowsURL, err := ResolveOfficialURL("windows")
	if err != nil || windowsURL != OfficialWindowsURL {
		t.Fatalf("unexpected windows url: %s, err=%v", windowsURL, err)
	}

	_, err = ResolveOfficialURL("unsupported_os")
	if err == nil {
		t.Fatalf("expected error for unsupported OS")
	}
}

func TestManagerPaths(t *testing.T) {
	mgr, err := NewManager(nil)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	if !strings.HasSuffix(mgr.Path(), filepath.Join(".android-mcp", "platform-tools")) {
		t.Fatalf("unexpected path: %s", mgr.Path())
	}

	binName := "adb"
	if runtime.GOOS == "windows" {
		binName = "adb.exe"
	}
	if !strings.HasSuffix(mgr.ADBPath(), binName) {
		t.Fatalf("unexpected adb path: %s", mgr.ADBPath())
	}
}

func TestZipSlipProtection(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "malicious.zip")

	// Create malicious zip archive with path traversal entry
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	f, err := zw.Create("../../evil.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	_, _ = f.Write([]byte("malicious content"))
	_ = zw.Close()

	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}

	extractTarget := filepath.Join(tempDir, "extract")
	err = extractZipSecurely(zipPath, extractTarget)
	if err == nil {
		t.Fatalf("expected Zip Slip security error, got nil")
	}
	if !strings.Contains(err.Error(), "Zip Slip") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEnsureExistingADB(t *testing.T) {
	mgr, err := NewManager(nil)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Mock ANDROID_MCP_ADB to pointing to valid binary (e.g. system adb or mock)
	if mgr.IsInstalled() {
		ctx := context.Background()
		found, err := mgr.Ensure(ctx)
		if err != nil || found == "" {
			t.Fatalf("Ensure failed on existing installation: %v", err)
		}
	}
}
