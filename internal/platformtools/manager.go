package platformtools

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
	"github.com/tintupratap/Android-MCP-go/internal/notification"
)

const (
	OfficialDarwinURL  = "https://dl.google.com/android/repository/platform-tools-latest-darwin.zip"
	OfficialLinuxURL   = "https://dl.google.com/android/repository/platform-tools-latest-linux.zip"
	OfficialWindowsURL = "https://dl.google.com/android/repository/platform-tools-latest-windows.zip"
)

type PlatformToolsMetadata struct {
	Managed     bool      `json:"managed"`
	Path        string    `json:"path"`
	Version     string    `json:"version"`
	Source      string    `json:"source"`
	InstalledAt time.Time `json:"installed_at"`
}

type DefaultManager struct {
	mu       sync.Mutex
	baseDir  string
	notifier notification.Notifier
	metadata *PlatformToolsMetadata
}

func NewManager(notifier notification.Notifier) (*DefaultManager, error) {
	if notifier == nil {
		notifier = notification.NewNotifier()
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine user home directory: %w", err)
	}
	baseDir := filepath.Join(home, ".android-mcp")

	return &DefaultManager{
		baseDir:  baseDir,
		notifier: notifier,
	}, nil
}

func (m *DefaultManager) Path() string {
	return filepath.Join(m.baseDir, "platform-tools")
}

func (m *DefaultManager) ADBPath() string {
	binName := "adb"
	if runtime.GOOS == "windows" {
		binName = "adb.exe"
	}
	return filepath.Join(m.Path(), binName)
}

func ResolveOfficialURL(goos string) (string, error) {
	switch goos {
	case "darwin":
		return OfficialDarwinURL, nil
	case "linux":
		return OfficialLinuxURL, nil
	case "windows":
		return OfficialWindowsURL, nil
	default:
		return "", fmt.Errorf("unsupported operating system for official platform-tools: %s", goos)
	}
}

func (m *DefaultManager) IsInstalled() bool {
	adbBinary := m.ADBPath()
	info, err := os.Stat(adbBinary)
	if err != nil || info.IsDir() {
		return false
	}
	return true
}

func (m *DefaultManager) Ensure(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Check explicit ANDROID_MCP_ADB override
	if envADB := os.Getenv("ANDROID_MCP_ADB"); envADB != "" {
		if _, err := os.Stat(envADB); err == nil {
			logging.Infof("Using explicit ANDROID_MCP_ADB override: %s", envADB)
			return envADB, nil
		}
	}

	// 2. Check local managed platform-tools
	if m.IsInstalled() {
		adbPath := m.ADBPath()
		ver, _ := m.getVersionLocked(ctx, adbPath)
		logging.Debugf("Found managed Platform-Tools at %s (version: %s)", adbPath, ver)
		return adbPath, nil
	}

	// 3. Download and install managed platform-tools
	logging.Infof("Platform-Tools not found in ~/.android-mcp/platform-tools. Triggering automatic download...")
	return m.installLocked(ctx)
}

func (m *DefaultManager) Update(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logging.Infof("Updating managed Platform-Tools...")
	return m.installLocked(ctx)
}

func (m *DefaultManager) installLocked(ctx context.Context) (string, error) {
	url, err := ResolveOfficialURL(runtime.GOOS)
	if err != nil {
		return "", err
	}

	_ = m.notifier.Notify("Android-MCP", "Android Platform-Tools not found. Downloading official Google ADB package...")

	tmpDir := filepath.Join(m.baseDir, "platform-tools.download")
	_ = os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "platform-tools.zip")

	logging.Infof("Downloading official Platform-Tools from %s...", url)
	if err := downloadFile(ctx, url, zipPath); err != nil {
		_ = m.notifier.Notify("Android-MCP", "Failed to download Android Platform-Tools. See logs for details.")
		return "", fmt.Errorf("failed to download official platform-tools: %w", err)
	}

	logging.Infof("Extracting Platform-Tools package securely...")
	extractTarget := filepath.Join(tmpDir, "extracted")
	if err := extractZipSecurely(zipPath, extractTarget); err != nil {
		_ = m.notifier.Notify("Android-MCP", "Failed to extract Android Platform-Tools archive.")
		return "", fmt.Errorf("failed to extract platform-tools zip: %w", err)
	}

	// Locate extracted platform-tools directory (zip typically contains root "platform-tools" directory)
	sourceDir := filepath.Join(extractTarget, "platform-tools")
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		sourceDir = extractTarget
	}

	binName := "adb"
	if runtime.GOOS == "windows" {
		binName = "adb.exe"
	}
	candidateADB := filepath.Join(sourceDir, binName)
	if _, err := os.Stat(candidateADB); os.IsNotExist(err) {
		return "", fmt.Errorf("downloaded platform-tools does not contain valid adb binary at %s", candidateADB)
	}

	// Ensure executable permissions on unix
	if runtime.GOOS != "windows" {
		_ = os.Chmod(candidateADB, 0755)
		if fastbootPath := filepath.Join(sourceDir, "fastboot"); func() bool { _, e := os.Stat(fastbootPath); return e == nil }() {
			_ = os.Chmod(fastbootPath, 0755)
		}
	}

	// Test candidate ADB binary execution
	ver, err := m.getVersionLocked(ctx, candidateADB)
	if err != nil || ver == "" {
		return "", fmt.Errorf("candidate adb binary execution check failed: %v", err)
	}

	// Atomic replacement: target directory
	targetDir := m.Path()
	backupDir := filepath.Join(m.baseDir, "platform-tools.old")
	_ = os.RemoveAll(backupDir)

	if _, err := os.Stat(targetDir); err == nil {
		_ = os.Rename(targetDir, backupDir)
	}

	if err := os.Rename(sourceDir, targetDir); err != nil {
		// Roll back backup if rename failed
		if _, errOld := os.Stat(backupDir); errOld == nil {
			_ = os.Rename(backupDir, targetDir)
		}
		return "", fmt.Errorf("failed to atomically install platform-tools: %w", err)
	}
	_ = os.RemoveAll(backupDir)

	finalADB := m.ADBPath()

	// Update persistent state in android-mcp.json
	cfg, errLoad := config.LoadConfig()
	if errLoad == nil {
		cfg.PlatformTools = &config.PlatformToolsConfig{
			Managed:     true,
			Path:        targetDir,
			Version:     ver,
			Source:      url,
			InstalledAt: time.Now(),
		}
		_ = config.SaveConfig(cfg)
	}

	logging.Infof("Platform-Tools v%s installed successfully at %s", ver, targetDir)
	_ = m.notifier.Notify("Android-MCP", fmt.Sprintf("Android Platform-Tools v%s installed successfully. ADB is ready.", ver))

	return finalADB, nil
}

func (m *DefaultManager) getVersionLocked(ctx context.Context, adbPath string) (string, error) {
	cmd := exec.CommandContext(ctx, adbPath, "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Android Debug Bridge version") {
			return strings.TrimPrefix(line, "Android Debug Bridge version "), nil
		}
		if strings.Contains(line, "Version") {
			return line, nil
		}
	}
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "unknown", nil
}

func downloadFile(ctx context.Context, url string, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %d downloading %s", resp.StatusCode, url)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	var total int64 = resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)
	lastMilestone := -1

	for {
		n, rErr := resp.Body.Read(buf)
		if n > 0 {
			_, wErr := out.Write(buf[:n])
			if wErr != nil {
				return wErr
			}
			downloaded += int64(n)
			if total > 0 {
				percent := int(float64(downloaded) / float64(total) * 100)
				milestone := percent / 25
				if milestone != lastMilestone {
					lastMilestone = milestone
					logging.Debugf("Downloading Platform-Tools progress: %d%% (%d / %d bytes)", percent, downloaded, total)
				}
			}
		}
		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			return rErr
		}
	}
	return nil
}

func extractZipSecurely(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Zip Slip vulnerability protection: ensure target path stays inside destDir
		filePath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(filePath), filepath.Clean(destDir)) {
			return fmt.Errorf("illegal file path in zip archive (Zip Slip): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
