package scrcpy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/logging"
	"github.com/tintupratap/Android-MCP-go/internal/notification"
)

type ProcessInfo struct {
	PID          int
	Serial       string
	Generation   uint64
	StartTime    time.Time
	ProfileName  string
	VideoCodec   string
	Renderer     string
	Cmd          *exec.Cmd
}

type ManagedAsset struct {
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
	activeProc   *ProcessInfo
	startGen     uint64
	startInProg  bool
	startCond    *sync.Cond
	restartCount map[string]int
	fileLock     *os.File
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

	lockFile, ok := acquireProcessLock(baseDir)
	if !ok {
		logging.Warnf("Another Android-MCP instance owns scrcpy.lock; managed scrcpy launch disabled for this instance.")
	}

	m := &Manager{
		baseDir:      baseDir,
		notifier:     notifier,
		restartCount: make(map[string]int),
		fileLock:     lockFile,
	}
	m.startCond = sync.NewCond(&m.mu)
	return m, nil
}

func acquireProcessLock(baseDir string) (*os.File, bool) {
	_ = os.MkdirAll(baseDir, 0755)
	lockPath := filepath.Join(baseDir, "scrcpy.lock")
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, false
	}
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		_ = file.Close()
		return nil, false
	}
	return file, true
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
	if custom := os.Getenv("ANDROID_MCP_SCRCPY_PATH"); custom != "" {
		return custom
	}

	scrcpyDir := m.Path()
	binName := "scrcpy"
	if os.Getenv("OS") == "Windows_NT" {
		binName = "scrcpy.exe"
	}

	directPath := filepath.Join(scrcpyDir, binName)
	if _, err := os.Stat(directPath); err == nil {
		return directPath
	}

	var found string
	_ = filepath.Walk(scrcpyDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() == binName {
			found = path
			return filepath.SkipAll
		}
		return nil
	})

	if found != "" {
		return found
	}

	return directPath
}

func (m *Manager) isProcAliveLocked(proc *ProcessInfo) bool {
	if proc == nil || proc.Cmd == nil || proc.Cmd.Process == nil {
		return false
	}
	err := proc.Cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

func (m *Manager) IsRunning(serial string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeProc == nil {
		return false
	}
	if serial != "" && m.activeProc.Serial != serial {
		return false
	}
	return m.isProcAliveLocked(m.activeProc)
}

func (m *Manager) GetActiveProcessInfo() (*ProcessInfo, int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeProc != nil && m.isProcAliveLocked(m.activeProc) {
		return m.activeProc, 1
	}
	return nil, 0
}

func (m *Manager) Stop(serial string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeProc != nil {
		if serial == "" || m.activeProc.Serial == serial {
			logging.Infof("Stopping managed scrcpy process pid=%d gen=%d for serial %s", m.activeProc.PID, m.activeProc.Generation, m.activeProc.Serial)
			if m.activeProc.Cmd != nil && m.activeProc.Cmd.Process != nil {
				_ = m.activeProc.Cmd.Process.Signal(syscall.SIGTERM)
			}
			m.activeProc = nil
		}
	}
}

func (m *Manager) StopAll() {
	m.Stop("")
}

func (m *Manager) IsInstalled() bool {
	binPath := m.BinaryPath()
	info, err := os.Stat(binPath)
	return err == nil && !info.IsDir()
}

func (m *Manager) EnsureInstalled(ctx context.Context) (string, error) {
	return m.Ensure(ctx)
}

func (m *Manager) Ensure(ctx context.Context) (string, error) {
	binPath := m.BinaryPath()
	if m.IsInstalled() {
		return binPath, nil
	}

	logging.Infof("Downloading official scrcpy release...")
	rel, err := FetchLatestRelease(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest scrcpy release: %w", err)
	}

	asset, err := ResolveAssetForPlatform(rel, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to resolve platform asset: %w", err)
	}

	if err := os.MkdirAll(m.DownloadsDir(), 0755); err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %w", err)
	}

	archivePath := filepath.Join(m.DownloadsDir(), asset.Name)
	if err := DownloadFile(ctx, asset.BrowserDownloadURL, archivePath); err != nil {
		return "", fmt.Errorf("failed to download scrcpy asset: %w", err)
	}

	stagingDir := m.StagingDir()
	_ = os.RemoveAll(stagingDir)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create staging directory: %w", err)
	}

	if strings.HasSuffix(archivePath, ".zip") {
		if err := ExtractZip(archivePath, stagingDir); err != nil {
			return "", fmt.Errorf("failed to extract zip: %w", err)
		}
	} else if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		if err := ExtractTarGz(archivePath, stagingDir); err != nil {
			return "", fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	}

	_ = os.RemoveAll(m.Path())
	if err := os.Rename(stagingDir, m.Path()); err != nil {
		return "", fmt.Errorf("failed to replace scrcpy directory: %w", err)
	}

	return m.BinaryPath(), nil
}

func (m *Manager) GetVersion(ctx context.Context) (string, error) {
	binPath := m.BinaryPath()
	cmd := exec.CommandContext(ctx, binPath, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run scrcpy --version: %w", err)
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "unknown", nil
}
