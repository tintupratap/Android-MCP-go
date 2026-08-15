package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type RuntimePaths struct {
	Root          string
	Config        string
	PlatformTools string
	ADB           string
	Scrcpy        string
	ScrcpyBinary  string
	Downloads     string
	Staging       string
	Logs          string
	Cache         string
}

func GetRuntimePaths() (*RuntimePaths, error) {
	var root string
	if envRoot := os.Getenv("ANDROID_MCP_HOME"); envRoot != "" {
		root = envRoot
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		root = filepath.Join(home, ".android-mcp")
	}

	adbBinName := "adb"
	scrcpyBinName := "scrcpy"
	if runtime.GOOS == "windows" {
		adbBinName = "adb.exe"
		scrcpyBinName = "scrcpy.exe"
	}

	ptDir := filepath.Join(root, "platform-tools")
	scrcpyDir := filepath.Join(root, "scrcpy")

	// If scrcpy is extracted in subfolder (e.g. scrcpy-macos-x86_64-v4.1/scrcpy), resolve subfolder binary
	scrcpyBin := filepath.Join(scrcpyDir, scrcpyBinName)
	if _, err := os.Stat(scrcpyBin); err != nil {
		var foundBin string
		_ = filepath.Walk(scrcpyDir, func(p string, info os.FileInfo, err error) error {
			if err == nil && info.Name() == scrcpyBinName && !info.IsDir() {
				foundBin = p
				return filepath.SkipAll
			}
			return nil
		})
		if foundBin != "" {
			scrcpyBin = foundBin
		}
	}

	return &RuntimePaths{
		Root:          root,
		Config:        filepath.Join(root, "android-mcp.json"),
		PlatformTools: ptDir,
		ADB:           filepath.Join(ptDir, adbBinName),
		Scrcpy:        scrcpyDir,
		ScrcpyBinary:  scrcpyBin,
		Downloads:     filepath.Join(root, ".downloads"),
		Staging:       filepath.Join(root, ".staging"),
		Logs:          filepath.Join(root, "logs"),
		Cache:         filepath.Join(root, "cache"),
	}, nil
}
