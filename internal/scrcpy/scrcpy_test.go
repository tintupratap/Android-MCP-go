package scrcpy

import (
	"context"
	"fmt"
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

func TestProfileResolutionAndDegradation(t *testing.T) {
	sysCaps := SystemCapabilities{
		Binary: BinaryCapabilities{
			SupportsRenderDriver: true,
			SupportsVideoCodec:   true,
			SupportsAudioSource:  true,
			SupportedFlags: map[string]bool{
				"--render-driver": true,
				"--video-codec":   true,
				"--audio-source":  true,
				"--no-audio":      true,
			},
		},
		Host: HostCapabilities{
			OS:             "darwin",
			Arch:           "arm64",
			NativeRenderer: "metal",
		},
		Device: DeviceCapabilities{
			Model:  "SOG09",
			IsWiFi: true,
		},
	}

	profile := ResolveOptimalProfile(nil, sysCaps, "192.168.1.3:5555", "Test Window")
	if profile.VideoCodec != "h265" || profile.Renderer != "metal" || profile.Bitrate != "4M" {
		t.Fatalf("unexpected optimal profile resolution: %+v", profile)
	}

	// Degrade Step 1: drop render-driver
	p1, ok1 := DegradeProfile(profile, sysCaps, fmt.Errorf("metal render driver unsupported"))
	if !ok1 || p1.Renderer != "" {
		t.Fatalf("expected step 1 degradation dropping renderer: %+v", p1)
	}

	// Degrade Step 2: fallback H.265 to H.264
	p2, ok2 := DegradeProfile(p1, sysCaps, fmt.Errorf("h265 encoder unsupported"))
	if !ok2 || p2.VideoCodec != "h264" {
		t.Fatalf("expected step 2 degradation switching to h264: %+v", p2)
	}

	// Degrade Step 3: disable audio
	p3, ok3 := DegradeProfile(p2, sysCaps, fmt.Errorf("audio output error"))
	if !ok3 || p3.Audio != "disabled" {
		t.Fatalf("expected step 3 degradation disabling audio: %+v", p3)
	}

	// Degrade Step 4: minimal safe profile
	p4, ok4 := DegradeProfile(p3, sysCaps, fmt.Errorf("generic option failed"))
	if !ok4 || p4.Name != "minimal_safe" || p4.Optimized {
		t.Fatalf("expected step 4 minimal safe profile: %+v", p4)
	}
}

func TestLiveViewManagerStateMachine(t *testing.T) {
	mgr, err := NewManager(nil)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	lvmDisabled := NewLiveViewManager(mgr, nil, nil, nil, true)
	if state, _, _, _, _ := lvmDisabled.GetStateInfo(); state != StateDisabled {
		t.Fatalf("expected StateDisabled when noScrcpyFlag is true, got %s", state)
	}

	lvm := NewLiveViewManager(mgr, nil, nil, nil, false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lvm.StartBackground(ctx)

	state, _, _, _, _ := lvm.GetStateInfo()
	if state != StateInitializing && state != StateWaitingForDevice {
		t.Fatalf("expected StateInitializing or StateWaitingForDevice, got %s", state)
	}

	lvm.Stop()
	if stateEnd, _, _, _, _ := lvm.GetStateInfo(); stateEnd != StateStopped {
		t.Fatalf("expected StateStopped after Stop(), got %s", stateEnd)
	}
}

func TestEnsureLiveViewClosedByUser(t *testing.T) {
	mgr, err := NewManager(nil)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	resolver := func(ctx context.Context) (string, string, error) {
		return "192.168.1.3:5555", "SOG09", nil
	}

	lvm := NewLiveViewManager(mgr, resolver, nil, nil, false)
	lvm.SetState(StateClosedByUser, "manual window close")

	state, _, _, _, _ := lvm.GetStateInfo()
	if state != StateClosedByUser {
		t.Fatalf("expected StateClosedByUser, got %s", state)
	}
}
