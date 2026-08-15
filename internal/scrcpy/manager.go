package scrcpy

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
	"github.com/tintupratap/Android-MCP-go/internal/notification"
)

const (
	OfficialRepoURL    = "https://github.com/Genymobile/scrcpy"
	OfficialReleaseAPI = "https://api.github.com/repos/Genymobile/scrcpy/releases/latest"
)

type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type GitHubRelease struct {
	TagName    string               `json:"tag_name"`
	Name       string               `json:"name"`
	Prerelease bool                 `json:"prerelease"`
	Draft      bool                 `json:"draft"`
	Assets     []GitHubReleaseAsset `json:"assets"`
}

type ResolvedAsset struct {
	Name        string
	URL         string
	ChecksumURL string
	Version     string
	Size        int64
}

type Manager struct {
	mu           sync.Mutex
	baseDir      string
	notifier     notification.Notifier
	activeProcs  map[string]*exec.Cmd
	restartCount map[string]int
	procsMu      sync.Mutex
}

func NewManager(notifier notification.Notifier) (*Manager, error) {
	if notifier == nil {
		notifier = notification.NewNotifier()
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to determine user home directory: %w", err)
	}
	baseDir := filepath.Join(home, ".android-mcp")

	return &Manager{
		baseDir:      baseDir,
		notifier:     notifier,
		activeProcs:  make(map[string]*exec.Cmd),
		restartCount: make(map[string]int),
	}, nil
}

func (m *Manager) Path() string {
	return filepath.Join(m.baseDir, "scrcpy")
}

func (m *Manager) DownloadsDir() string {
	return filepath.Join(m.baseDir, ".downloads")
}

func (m *Manager) StagingDir() string {
	return filepath.Join(m.baseDir, ".staging", "scrcpy")
}

func (m *Manager) BinaryPath() string {
	binName := "scrcpy"
	if runtime.GOOS == "windows" {
		binName = "scrcpy.exe"
	}

	// 1. Managed path ~/.android-mcp/scrcpy/scrcpy (or subfolder)
	managedPath := m.Path()
	var foundBin string
	_ = filepath.Walk(managedPath, func(p string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == binName && !info.IsDir() {
			foundBin = p
			return filepath.SkipAll
		}
		return nil
	})
	if foundBin != "" {
		return foundBin
	}

	return filepath.Join(managedPath, binName)
}

func (m *Manager) IsInstalled() bool {
	binPath := m.BinaryPath()
	info, err := os.Stat(binPath)
	if err != nil || info.IsDir() {
		return false
	}
	return true
}

func (m *Manager) GetVersion(ctx context.Context) (string, error) {
	binPath := m.BinaryPath()
	if !m.IsInstalled() {
		return "not installed", fmt.Errorf("scrcpy binary not found at %s", binPath)
	}

	cmd := exec.CommandContext(ctx, binPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "unknown", err
	}

	lines := strings.Split(out.String(), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "unknown", nil
}

func (m *Manager) EnsureInstalled(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if envBin := os.Getenv("ANDROID_MCP_SCRCPY_PATH"); envBin != "" {
		if _, err := os.Stat(envBin); err == nil {
			return envBin, nil
		}
	}

	if m.IsInstalled() {
		return m.BinaryPath(), nil
	}

	return m.DownloadAndInstallLocked(ctx)
}

func (m *Manager) Ensure(ctx context.Context) (string, error) {
	return m.EnsureInstalled(ctx)
}

func FetchLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", OfficialReleaseAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Android-MCP-go")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var rel GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	if rel.Draft || rel.Prerelease {
		return nil, fmt.Errorf("latest release %s is a draft or prerelease", rel.TagName)
	}
	return &rel, nil
}

func ResolveAssetForPlatform(rel *GitHubRelease, goos, goarch string) (*ResolvedAsset, error) {
	if rel == nil {
		return nil, fmt.Errorf("release object is nil")
	}

	var chosenAsset *GitHubReleaseAsset
	var checksumURL string

	for _, asset := range rel.Assets {
		name := strings.ToLower(asset.Name)

		if strings.Contains(name, "sha256") || strings.Contains(name, "checksum") {
			checksumURL = asset.BrowserDownloadURL
			continue
		}

		isArchive := strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz")
		if !isArchive {
			continue
		}

		switch goos {
		case "darwin":
			if strings.Contains(name, "macos") || strings.Contains(name, "mac") || strings.Contains(name, "darwin") {
				if goarch == "arm64" && (strings.Contains(name, "aarch64") || strings.Contains(name, "arm64") || !strings.Contains(name, "x86_64")) {
					chosenAsset = &asset
				} else if goarch == "amd64" && (strings.Contains(name, "x86_64") || strings.Contains(name, "amd64")) {
					chosenAsset = &asset
				}
			}
		case "linux":
			if strings.Contains(name, "linux") {
				if goarch == "amd64" && (strings.Contains(name, "x86_64") || strings.Contains(name, "amd64")) {
					chosenAsset = &asset
				}
			}
		case "windows":
			if goarch == "386" && (strings.Contains(name, "win32") || strings.Contains(name, "x86")) {
				chosenAsset = &asset
			} else if goarch == "amd64" && (strings.Contains(name, "win64") || strings.Contains(name, "x86_64") || strings.Contains(name, "amd64")) {
				chosenAsset = &asset
			}
		}

		if chosenAsset != nil {
			break
		}
	}

	if chosenAsset == nil {
		return nil, fmt.Errorf("no official upstream scrcpy binary release available for %s/%s in release %s", goos, goarch, rel.TagName)
	}

	return &ResolvedAsset{
		Name:        chosenAsset.Name,
		URL:         chosenAsset.BrowserDownloadURL,
		ChecksumURL: checksumURL,
		Version:     rel.TagName,
		Size:        chosenAsset.Size,
	}, nil
}

func (m *Manager) DownloadAndInstallLocked(ctx context.Context) (string, error) {
	rel, err := FetchLatestRelease(ctx)
	if err != nil {
		logging.Warnf("Failed to fetch latest scrcpy release from GitHub API: %v", err)
		_ = m.notifier.Notify("Android-MCP", "Failed to resolve official scrcpy release.")
		return "", fmt.Errorf("scrcpy official release resolution failed: %w", err)
	}

	asset, err := ResolveAssetForPlatform(rel, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		logging.Warnf("scrcpy platform resolution error: %v", err)
		_ = m.notifier.Notify("Android-MCP", fmt.Sprintf("scrcpy not supported for %s/%s", runtime.GOOS, runtime.GOARCH))
		return "", err
	}

	_ = m.notifier.Notify("Android-MCP", fmt.Sprintf("Downloading official scrcpy %s release...", asset.Version))
	logging.Infof("Downloading official scrcpy %s release (%s) for %s/%s...", asset.Version, asset.Name, runtime.GOOS, runtime.GOARCH)

	downloadsDir := m.DownloadsDir()
	_ = os.MkdirAll(downloadsDir, 0755)
	archivePath := filepath.Join(downloadsDir, asset.Name)
	defer os.Remove(archivePath)

	if err := downloadFileWithProgress(ctx, asset.URL, archivePath, asset.Size); err != nil {
		_ = m.notifier.Notify("Android-MCP", "Failed to download scrcpy.")
		return "", fmt.Errorf("failed to download %s: %w", asset.URL, err)
	}

	if asset.ChecksumURL != "" {
		if err := verifyChecksum(ctx, archivePath, asset.Name, asset.ChecksumURL); err != nil {
			logging.Warnf("Checksum verification note: %v", err)
		} else {
			logging.Infof("SHA-256 checksum verified for %s", asset.Name)
		}
	}

	stagingDir := m.StagingDir()
	_ = os.RemoveAll(stagingDir)
	_ = os.MkdirAll(stagingDir, 0755)
	defer os.RemoveAll(stagingDir)

	if strings.HasSuffix(asset.Name, ".zip") {
		if err := extractZipSecurely(archivePath, stagingDir); err != nil {
			return "", fmt.Errorf("failed to extract zip archive: %w", err)
		}
	} else if strings.HasSuffix(asset.Name, ".tar.gz") || strings.HasSuffix(asset.Name, ".tgz") {
		if err := extractTarGzSecurely(archivePath, stagingDir); err != nil {
			return "", fmt.Errorf("failed to extract tar.gz archive: %w", err)
		}
	} else {
		return "", fmt.Errorf("unsupported archive format: %s", asset.Name)
	}

	binName := "scrcpy"
	if runtime.GOOS == "windows" {
		binName = "scrcpy.exe"
	}
	var stagedBin string
	_ = filepath.Walk(stagingDir, func(p string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == binName && !info.IsDir() {
			stagedBin = p
			return filepath.SkipAll
		}
		return nil
	})

	if stagedBin == "" {
		return "", fmt.Errorf("extracted scrcpy package does not contain %s executable", binName)
	}

	_ = os.Chmod(stagedBin, 0755)

	verCmd := exec.CommandContext(ctx, stagedBin, "--version")
	if err := verCmd.Run(); err != nil {
		return "", fmt.Errorf("scrcpy executable --version check failed: %w", err)
	}

	targetDir := m.Path()
	_ = os.RemoveAll(targetDir)

	sourceDir := filepath.Dir(stagedBin)
	if err := os.Rename(sourceDir, targetDir); err != nil {
		return "", fmt.Errorf("failed to install scrcpy to %s: %w", targetDir, err)
	}

	finalBin := m.BinaryPath()

	cfg, errLoad := config.LoadConfig()
	if errLoad == nil {
		cfg.ManagedScrcpy = &config.ManagedScrcpyConfig{
			Managed:     true,
			Path:        targetDir,
			Version:     asset.Version,
			Release:     rel.TagName,
			Source:      OfficialRepoURL,
			InstalledAt: time.Now(),
		}
		_ = config.SaveConfig(cfg)
	}

	logging.Infof("scrcpy %s installed successfully at %s", asset.Version, targetDir)
	_ = m.notifier.Notify("Android-MCP", fmt.Sprintf("scrcpy %s installed successfully.", asset.Version))
	return finalBin, nil
}

func BuildArgs(cfg *config.ScrcpyPreferences, serial string, title string) []string {
	if title == "" {
		title = fmt.Sprintf("Android-MCP — %s", serial)
	}

	args := []string{"--window-title", title}
	if serial != "" {
		args = append(args, "-s", serial)
	}

	if cfg != nil {
		if cfg.VideoCodec != "" {
			args = append(args, "--video-codec", cfg.VideoCodec)
		}
		if cfg.VideoBitrate != "" {
			args = append(args, "--video-bit-rate", cfg.VideoBitrate)
		}
		if cfg.DisplayID != "" {
			args = append(args, "--display-id", cfg.DisplayID)
		}
		if cfg.AudioSource != "" {
			args = append(args, "--audio-source", cfg.AudioSource)
		}
		if cfg.StayAwake {
			args = append(args, "--stay-awake")
		}
		if cfg.TurnScreenOff {
			args = append(args, "--turn-screen-off")
		}
		if cfg.RenderDriver != "" {
			args = append(args, "--render-driver", cfg.RenderDriver)
		}
	}

	if extraArgs := os.Getenv("ANDROID_MCP_SCRCPY_ARGS"); extraArgs != "" {
		args = append(args, strings.Fields(extraArgs)...)
	}

	return args
}

func (m *Manager) Launch(ctx context.Context, serial string, title string) error {
	m.procsMu.Lock()
	if cmd, ok := m.activeProcs[serial]; ok && cmd != nil && cmd.Process != nil {
		if err := cmd.Process.Signal(syscall.Signal(0)); err == nil {
			m.procsMu.Unlock()
			logging.Debugf("scrcpy screen mirror process already active for device %s", serial)
			return nil
		}
	}
	m.procsMu.Unlock()

	binPath := m.BinaryPath()
	if _, err := os.Stat(binPath); err != nil {
		logging.Warnf("Cannot launch scrcpy: binary not found at %s", binPath)
		return fmt.Errorf("scrcpy binary not found at %s", binPath)
	}

	appCfg, errCfg := config.LoadConfig()
	var prefs *config.ScrcpyPreferences
	if errCfg == nil && appCfg != nil {
		prefs = &appCfg.Scrcpy
	}

	args := BuildArgs(prefs, serial, title)

	logging.Infof("Launching scrcpy screen mirror for device %s: %s %s", serial, binPath, strings.Join(args, " "))

	cmd := exec.Command(binPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		logging.Warnf("Failed to launch scrcpy process: %v", err)
		return fmt.Errorf("failed to launch scrcpy: %w", err)
	}

	m.procsMu.Lock()
	m.activeProcs[serial] = cmd
	m.procsMu.Unlock()

	go func() {
		_ = cmd.Wait()
		m.procsMu.Lock()
		delete(m.activeProcs, serial)
		m.procsMu.Unlock()
		logging.Debugf("scrcpy screen mirror process for device %s exited", serial)
	}()

	_ = m.notifier.Notify("Android-MCP", fmt.Sprintf("Live screen mirror active for %s", serial))
	return nil
}

func (m *Manager) IsRunning(serial string) bool {
	m.procsMu.Lock()
	defer m.procsMu.Unlock()
	cmd, ok := m.activeProcs[serial]
	if !ok || cmd == nil || cmd.Process == nil {
		return false
	}
	return cmd.Process.Signal(syscall.Signal(0)) == nil
}

func (m *Manager) Stop(serial string) {
	m.procsMu.Lock()
	defer m.procsMu.Unlock()

	if cmd, ok := m.activeProcs[serial]; ok && cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
		delete(m.activeProcs, serial)
	}
}

func (m *Manager) StopAll() {
	m.procsMu.Lock()
	defer m.procsMu.Unlock()

	for serial, cmd := range m.activeProcs {
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		delete(m.activeProcs, serial)
	}
}

type progressWriter struct {
	total      int64
	downloaded int64
	lastReport int
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)
	if pw.total > 0 {
		pct := int((float64(pw.downloaded) / float64(pw.total)) * 100)
		if pct >= pw.lastReport+25 {
			pw.lastReport = pct
			logging.Infof("scrcpy download progress: %d%%", pct)
		}
	}
	return n, nil
}

func downloadFileWithProgress(ctx context.Context, url string, dest string, totalSize int64) error {
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
		return fmt.Errorf("HTTP %d downloading %s", resp.StatusCode, url)
	}

	if totalSize <= 0 && resp.ContentLength > 0 {
		totalSize = resp.ContentLength
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	pw := &progressWriter{total: totalSize}
	writer := io.MultiWriter(out, pw)

	_, err = io.Copy(writer, resp.Body)
	if err == nil {
		logging.Infof("scrcpy download complete (100%%)")
	}
	return err
}

func verifyChecksum(ctx context.Context, archivePath, assetName, checksumURL string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", checksumURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d downloading checksums", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var expectedSHA string
	for _, line := range lines {
		if strings.Contains(line, assetName) {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				expectedSHA = parts[0]
				break
			}
		}
	}

	if expectedSHA == "" {
		return fmt.Errorf("checksum entry for %s not found in %s", assetName, checksumURL)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actualSHA := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actualSHA, expectedSHA) {
		return fmt.Errorf("SHA-256 mismatch for %s: expected %s, got %s", assetName, expectedSHA, actualSHA)
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
		filePath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(filePath), filepath.Clean(destDir)) {
			return fmt.Errorf("Zip Slip path traversal attempt: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(filePath, 0755)
			continue
		}
		_ = os.MkdirAll(filepath.Dir(filePath), 0755)
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

func extractTarGzSecurely(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("Tar Slip path traversal attempt: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}
