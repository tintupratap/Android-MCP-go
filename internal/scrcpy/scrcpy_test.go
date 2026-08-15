package scrcpy

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tintupratap/Android-MCP-go/internal/config"
)

func TestScrcpyManagerBinaryPath(t *testing.T) {
	mgr, err := NewManager(nil)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	bin := mgr.BinaryPath()
	if len(bin) == 0 {
		t.Fatalf("expected binary path")
	}

	if mgr.IsInstalled() {
		ctx := context.Background()
		found, err := mgr.Ensure(ctx)
		if err != nil || found == "" {
			t.Fatalf("Ensure failed on installed scrcpy: %v", err)
		}

		ver, err := mgr.GetVersion(ctx)
		if err != nil || len(ver) == 0 {
			t.Fatalf("GetVersion failed: %v", err)
		}
	}
}

func TestResolveAssetForPlatform(t *testing.T) {
	mockRel := &GitHubRelease{
		TagName:    "v4.1",
		Prerelease: false,
		Draft:      false,
		Assets: []GitHubReleaseAsset{
			{Name: "scrcpy-win64-v4.1.zip", BrowserDownloadURL: "https://example.com/scrcpy-win64-v4.1.zip"},
			{Name: "scrcpy-linux-x86_64-v4.1.tar.gz", BrowserDownloadURL: "https://example.com/scrcpy-linux-x86_64-v4.1.tar.gz"},
			{Name: "scrcpy-macos-aarch64-v4.1.tar.gz", BrowserDownloadURL: "https://example.com/scrcpy-macos-aarch64-v4.1.tar.gz"},
		},
	}

	assetWin, err := ResolveAssetForPlatform(mockRel, "windows", "amd64")
	if err != nil || assetWin.Name != "scrcpy-win64-v4.1.zip" {
		t.Fatalf("unexpected windows resolution: %+v, err=%v", assetWin, err)
	}

	assetLin, err := ResolveAssetForPlatform(mockRel, "linux", "amd64")
	if err != nil || assetLin.Name != "scrcpy-linux-x86_64-v4.1.tar.gz" {
		t.Fatalf("unexpected linux resolution: %+v, err=%v", assetLin, err)
	}

	assetMac, err := ResolveAssetForPlatform(mockRel, "darwin", "arm64")
	if err != nil || assetMac.Name != "scrcpy-macos-aarch64-v4.1.tar.gz" {
		t.Fatalf("unexpected macos resolution: %+v, err=%v", assetMac, err)
	}

	_, errUnsup := ResolveAssetForPlatform(mockRel, "openbsd", "amd64")
	if errUnsup == nil {
		t.Fatalf("expected error for unsupported openbsd platform")
	}
}

func TestBuildArgs(t *testing.T) {
	cfg := &config.ScrcpyPreferences{
		VideoCodec:   "h265",
		VideoBitrate: "4M",
		DisplayID:    "0",
		AudioSource:  "playback",
		StayAwake:    true,
	}

	args := BuildArgs(cfg, "192.168.1.3:5555", "Test Title")
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "--window-title Test Title") {
		t.Fatalf("missing window title in args: %s", joined)
	}
	if !strings.Contains(joined, "-s 192.168.1.3:5555") {
		t.Fatalf("missing target serial in args: %s", joined)
	}
	if !strings.Contains(joined, "--video-codec h265") || !strings.Contains(joined, "--stay-awake") {
		t.Fatalf("missing scrcpy options in args: %s", joined)
	}
}

func TestScrcpyLaunchMissingBinary(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("PATH", tempHome)
	t.Setenv("ANDROID_MCP_SCRCPY_PATH", filepath.Join(tempHome, "nonexistent_scrcpy"))

	mgr, err := NewManager(nil)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()
	err = mgr.Launch(ctx, "mock_device", "Test Title")
	if err == nil {
		t.Fatalf("expected error launching missing binary, got nil")
	}
}
