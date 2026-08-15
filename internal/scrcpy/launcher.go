package scrcpy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
)

type ActiveProfileInfo struct {
	Serial       string
	ProfileName  string
	VideoCodec   string
	Renderer     string
	Audio        string
	Bitrate      string
	Optimized    bool
	FallbackStep int
	Args         []string
}

func (m *Manager) EnsureRunning(ctx context.Context, adbClient *adb.Client, serial string, title string) error {
	m.mu.Lock()

	if m.fileLock == nil {
		logging.Warnf("EnsureRunning: another Android-MCP instance owns scrcpy.lock; skipping scrcpy launch.")
		m.mu.Unlock()
		return nil
	}

	// Double-check 1: Is active process alive for target serial?
	if m.activeProc != nil {
		if m.activeProc.Serial == serial && m.isProcAliveLocked(m.activeProc) {
			m.mu.Unlock()
			logging.Debugf("EnsureRunning: scrcpy process pid=%d gen=%d already active for serial %s", m.activeProc.PID, m.activeProc.Generation, serial)
			return nil
		}
		// Target serial changed or old process died -> stop old process
		logging.Infof("EnsureRunning: stopping existing process pid=%d (serial=%s, target=%s)", m.activeProc.PID, m.activeProc.Serial, serial)
		if m.activeProc.Cmd != nil && m.activeProc.Cmd.Process != nil {
			_ = m.activeProc.Cmd.Process.Signal(syscall.SIGTERM)
		}
		m.activeProc = nil
	}

	// Double-check 2: Singleflight gate if startup in progress
	for m.startInProg {
		logging.Debugf("EnsureRunning: launch in progress for serial %s, waiting on condition...", serial)
		m.startCond.Wait()
		if m.activeProc != nil && m.activeProc.Serial == serial && m.isProcAliveLocked(m.activeProc) {
			m.mu.Unlock()
			logging.Debugf("EnsureRunning: scrcpy process pid=%d gen=%d ready after waiting", m.activeProc.PID, m.activeProc.Generation)
			return nil
		}
	}

	// Mark startup in progress & increment generation
	m.startInProg = true
	m.startGen++
	currentGen := m.startGen
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.startInProg = false
		m.startCond.Broadcast()
		m.mu.Unlock()
	}()

	procInfo, err := m.launchInternal(ctx, adbClient, serial, title, currentGen)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.activeProc = procInfo
	m.mu.Unlock()

	// Monotonic generation process monitor
	go func(gen uint64, proc *exec.Cmd, s string) {
		_ = proc.Wait()
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.activeProc != nil && m.activeProc.Generation == gen {
			logging.Infof("scrcpy process pid=%d gen=%d exited for serial %s", proc.Process.Pid, gen, s)
			m.activeProc = nil
		} else {
			logging.Debugf("Ignored stale scrcpy exit callback for gen=%d (active gen=%d)", gen, m.startGen)
		}
	}(currentGen, procInfo.Cmd, serial)

	return nil
}

func (m *Manager) LaunchWithFallback(ctx context.Context, adbClient *adb.Client, serial string, title string) error {
	return m.EnsureRunning(ctx, adbClient, serial, title)
}

func (m *Manager) Launch(ctx context.Context, serial string, title string) error {
	return m.EnsureRunning(ctx, nil, serial, title)
}

func (m *Manager) launchInternal(ctx context.Context, adbClient *adb.Client, serial string, title string, gen uint64) (*ProcessInfo, error) {
	binPath := m.BinaryPath()
	if _, err := os.Stat(binPath); err != nil {
		logging.Warnf("Cannot launch scrcpy: binary not found at %s", binPath)
		return nil, fmt.Errorf("scrcpy binary not found at %s", binPath)
	}

	appCfg, errCfg := config.LoadConfig()
	var prefs *config.ScrcpyPreferences
	if errCfg == nil && appCfg != nil {
		prefs = &appCfg.Scrcpy
	}

	binCaps := DetectBinaryCapabilities(ctx, binPath)
	hostCaps := DetectHostCapabilities()
	devCaps := DetectDeviceCapabilities(ctx, adbClient, serial)

	sysCaps := SystemCapabilities{
		Binary: binCaps,
		Host:   hostCaps,
		Device: devCaps,
	}

	profile := ResolveOptimalProfile(prefs, sysCaps, serial, title)

	maxAttempts := 5
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		logging.Infof("Launching single-instance scrcpy for device %s (gen=%d, profile=%s, attempt=%d/%d): %s %s",
			serial, gen, profile.Name, attempt+1, maxAttempts, binPath, strings.Join(profile.Args, " "))

		cmd := exec.Command(binPath, profile.Args...)
		cmd.Stdout = nil
		cmd.Stderr = nil

		paths, errPaths := config.GetRuntimePaths()
		if errPaths == nil && paths != nil {
			cmd.Env = append(os.Environ(), "ADB="+paths.ADB)
		}

		if err := cmd.Start(); err != nil {
			lastErr = err
			logging.Warnf("scrcpy launch attempt %d failed: %v", attempt+1, err)
			degraded, ok := DegradeProfile(profile, sysCaps, err)
			if !ok {
				break
			}
			profile = degraded
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Check process liveness after brief stabilization pause
		time.Sleep(250 * time.Millisecond)
		if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
			lastErr = fmt.Errorf("scrcpy process exited immediately")
			degraded, ok := DegradeProfile(profile, sysCaps, lastErr)
			if !ok {
				break
			}
			profile = degraded
			continue
		}

		return &ProcessInfo{
			PID:         cmd.Process.Pid,
			Serial:      serial,
			Generation:  gen,
			StartTime:   time.Now(),
			ProfileName: profile.Name,
			VideoCodec:  profile.VideoCodec,
			Renderer:    profile.Renderer,
			Cmd:         cmd,
		}, nil
	}

	return nil, fmt.Errorf("scrcpy launch failed after %d fallback attempts: %v", maxAttempts, lastErr)
}
