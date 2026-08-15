package scrcpy

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
)

type BinaryCapabilities struct {
	Version            string
	SupportsRenderDriver bool
	SupportsVideoCodec   bool
	SupportsVideoEncoder bool
	SupportsMaxFPS       bool
	SupportsAudioCodec   bool
	SupportsAudioSource  bool
	SupportsDisplayID    bool
	SupportsCamera       bool
	SupportedFlags       map[string]bool
}

type DeviceCapabilities struct {
	Manufacturer string
	Model        string
	SDKVersion   string
	IsWiFi       bool
	Encoders     []string
}

type HostCapabilities struct {
	OS            string
	Arch          string
	NativeRenderer string
}

type SystemCapabilities struct {
	Binary BinaryCapabilities
	Host   HostCapabilities
	Device DeviceCapabilities
}

var (
	capCacheMu sync.Mutex
	capCache   *BinaryCapabilities
	capBinPath string
)

func DetectBinaryCapabilities(ctx context.Context, scrcpyBin string) BinaryCapabilities {
	capCacheMu.Lock()
	defer capCacheMu.Unlock()

	if capCache != nil && capBinPath == scrcpyBin {
		return *capCache
	}

	caps := BinaryCapabilities{
		SupportedFlags: make(map[string]bool),
	}

	cmd := exec.CommandContext(ctx, scrcpyBin, "--help")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	_ = cmd.Run()
	helpText := out.String()

	flags := []string{
		"--render-driver",
		"--video-codec",
		"--video-encoder",
		"--max-fps",
		"--audio-codec",
		"--audio-source",
		"--display-id",
		"--camera-id",
		"--stay-awake",
		"--turn-screen-off",
		"--always-on-top",
	}

	for _, flag := range flags {
		if strings.Contains(helpText, flag) {
			caps.SupportedFlags[flag] = true
		}
	}

	caps.SupportsRenderDriver = caps.SupportedFlags["--render-driver"]
	caps.SupportsVideoCodec = caps.SupportedFlags["--video-codec"]
	caps.SupportsVideoEncoder = caps.SupportedFlags["--video-encoder"]
	caps.SupportsMaxFPS = caps.SupportedFlags["--max-fps"]
	caps.SupportsAudioCodec = caps.SupportedFlags["--audio-codec"]
	caps.SupportsAudioSource = caps.SupportedFlags["--audio-source"]
	caps.SupportsDisplayID = caps.SupportedFlags["--display-id"]
	caps.SupportsCamera = caps.SupportedFlags["--camera-id"]

	// Get version string
	verCmd := exec.CommandContext(ctx, scrcpyBin, "--version")
	var verOut bytes.Buffer
	verCmd.Stdout = &verOut
	if err := verCmd.Run(); err == nil {
		lines := strings.Split(verOut.String(), "\n")
		if len(lines) > 0 {
			caps.Version = strings.TrimSpace(lines[0])
		}
	}

	capCache = &caps
	capBinPath = scrcpyBin

	return caps
}

func DetectHostCapabilities() HostCapabilities {
	nativeRenderer := "opengl"
	if runtime.GOOS == "darwin" {
		nativeRenderer = "metal"
	} else if runtime.GOOS == "windows" {
		nativeRenderer = "direct3d"
	}

	return HostCapabilities{
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		NativeRenderer: nativeRenderer,
	}
}

func DetectDeviceCapabilities(ctx context.Context, adbClient *adb.Client, serial string) DeviceCapabilities {
	caps := DeviceCapabilities{
		IsWiFi: strings.Contains(serial, ":"),
	}

	if adbClient == nil || serial == "" {
		return caps
	}

	if model, err := adbClient.ExecuteShell(ctx, serial, "getprop", "ro.product.model"); err == nil {
		caps.Model = strings.TrimSpace(model)
	}
	if manu, err := adbClient.ExecuteShell(ctx, serial, "getprop", "ro.product.manufacturer"); err == nil {
		caps.Manufacturer = strings.TrimSpace(manu)
	}
	if sdk, err := adbClient.ExecuteShell(ctx, serial, "getprop", "ro.build.version.sdk"); err == nil {
		caps.SDKVersion = strings.TrimSpace(sdk)
	}

	logging.Debugf("Detected device capabilities for %s: model=%s, sdk=%s, wifi=%v", serial, caps.Model, caps.SDKVersion, caps.IsWiFi)
	return caps
}
