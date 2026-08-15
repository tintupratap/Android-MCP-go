package scrcpy

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
	"github.com/tintupratap/Android-MCP-go/internal/notification"
)

type LiveViewState string

const (
	StateStopped          LiveViewState = "STOPPED"
	StateInitializing     LiveViewState = "INITIALIZING"
	StateWaitingForDevice LiveViewState = "WAITING_FOR_DEVICE"
	StateDeviceFound      LiveViewState = "DEVICE_FOUND"
	StateConnecting       LiveViewState = "CONNECTING"
	StateDeviceConnected  LiveViewState = "DEVICE_CONNECTED"
	StatePreparingScrcpy  LiveViewState = "PREPARING_SCRCPY"
	StateStartingScrcpy   LiveViewState = "STARTING_SCRCPY"
	StateRunning          LiveViewState = "RUNNING"
	StateClosedByUser     LiveViewState = "CLOSED_BY_USER"
	StateCrashed          LiveViewState = "CRASHED"
	StateRestarting       LiveViewState = "RESTARTING"
	StateError            LiveViewState = "ERROR"
	StateDisabled         LiveViewState = "DISABLED"
)

type DeviceResolverFunc func(ctx context.Context) (serial string, model string, err error)

type LiveViewManager struct {
	mu           sync.RWMutex
	relaunchMu   sync.Mutex
	state        LiveViewState
	lastError    string
	activeSerial string
	activeModel  string
	restartCount int
	maxRestarts  int
	enabled      bool
	autoStart    bool
	scrcpyMgr    *Manager
	resolver     DeviceResolverFunc
	adbClient    *adb.Client
	notifier     notification.Notifier
	cancel       context.CancelFunc
}

func NewLiveViewManager(scrcpyMgr *Manager, resolver DeviceResolverFunc, adbClient *adb.Client, notifier notification.Notifier, noScrcpyFlag bool) *LiveViewManager {
	if notifier == nil {
		notifier = notification.NewNotifier()
	}

	enabled := true
	autoStart := true

	if noScrcpyFlag || os.Getenv("ANDROID_MCP_SCRCPY") == "false" {
		enabled = false
		autoStart = false
	} else {
		appCfg, err := config.LoadConfig()
		if err == nil && appCfg != nil {
			enabled = appCfg.Scrcpy.Enabled
			autoStart = appCfg.Scrcpy.AutoStart
		}
	}

	initialState := StateStopped
	if !enabled || !autoStart {
		initialState = StateDisabled
	}

	return &LiveViewManager{
		state:       initialState,
		maxRestarts: 3,
		enabled:     enabled,
		autoStart:   autoStart,
		scrcpyMgr:   scrcpyMgr,
		resolver:    resolver,
		adbClient:   adbClient,
		notifier:    notifier,
	}
}

func (l *LiveViewManager) SetState(s LiveViewState, errStr string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = s
	l.lastError = errStr
	logging.Infof("LiveView state transition -> %s (error=%s)", s, errStr)
}

func (l *LiveViewManager) GetStateInfo() (LiveViewState, string, string, string, int) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.state, l.lastError, l.activeSerial, l.activeModel, l.restartCount
}

func (l *LiveViewManager) StartBackground(parentCtx context.Context) {
	if !l.enabled || !l.autoStart {
		l.SetState(StateDisabled, "disabled by configuration or CLI flag")
		return
	}

	ctx, cancel := context.WithCancel(parentCtx)
	l.cancel = cancel

	l.SetState(StateInitializing, "")

	go l.runLoop(ctx)
}

func (l *LiveViewManager) Stop() {
	if l.cancel != nil {
		l.cancel()
	}
	if l.scrcpyMgr != nil && l.activeSerial != "" {
		l.scrcpyMgr.Stop(l.activeSerial)
	}
	l.SetState(StateStopped, "")
}

func (l *LiveViewManager) runLoop(ctx context.Context) {
	// Run first step immediately upon server startup
	l.step(ctx)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			l.SetState(StateStopped, "context canceled")
			return
		case <-ticker.C:
			l.step(ctx)
		}
	}
}

func (l *LiveViewManager) step(ctx context.Context) {
	l.mu.RLock()
	st := l.state
	activeSer := l.activeSerial
	l.mu.RUnlock()

	switch st {
	case StateInitializing, StateWaitingForDevice, StateError:
		l.SetState(StateWaitingForDevice, "")
		if l.resolver == nil {
			return
		}

		ser, mod, err := l.resolver(ctx)
		if err != nil || ser == "" {
			logging.Debugf("LiveView background polling: waiting for device...")
			return
		}

		l.mu.Lock()
		l.activeSerial = ser
		l.activeModel = mod
		l.mu.Unlock()

		l.SetState(StateDeviceFound, "")
		l.SetState(StateDeviceConnected, "")
		activeSer = ser
		fallthrough

	case StateDeviceConnected, StateRestarting:
		if l.scrcpyMgr == nil {
			return
		}

		l.SetState(StatePreparingScrcpy, "")

		// Ensure scrcpy binary is installed
		if !l.scrcpyMgr.IsInstalled() {
			l.SetState(StatePreparingScrcpy, "downloading scrcpy release...")
			if _, err := l.scrcpyMgr.EnsureInstalled(ctx); err != nil {
				l.SetState(StateError, fmt.Sprintf("scrcpy installation failed: %v", err))
				return
			}
		}

		l.SetState(StateStartingScrcpy, "")
		title := fmt.Sprintf("Android-MCP — %s (%s)", l.activeModel, activeSer)

		err := l.scrcpyMgr.LaunchWithFallback(ctx, l.adbClient, activeSer, title)
		if err != nil {
			l.mu.Lock()
			l.restartCount++
			rc := l.restartCount
			l.mu.Unlock()

			if rc >= l.maxRestarts {
				l.SetState(StateError, fmt.Sprintf("scrcpy launch failed after %d restarts: %v", rc, err))
			} else {
				l.SetState(StateRestarting, fmt.Sprintf("retry %d/%d: %v", rc, l.maxRestarts, err))
			}
			return
		}

		l.mu.Lock()
		l.restartCount = 0
		l.mu.Unlock()

		l.SetState(StateRunning, "")
		_ = l.notifier.Notify("Android-MCP", fmt.Sprintf("%s live view started.", l.activeModel))

	case StateRunning:
		if l.scrcpyMgr == nil || activeSer == "" {
			return
		}

		// Check if active mirror process is still alive
		if !l.scrcpyMgr.IsRunning(activeSer) {
			logging.Warnf("LiveView detected scrcpy exit for device %s", activeSer)
			if l.resolver != nil {
				ser, _, err := l.resolver(ctx)
				if err == nil && ser == activeSer {
					// User closed the scrcpy window manually -> wait for next MCP tool call
					l.SetState(StateClosedByUser, "scrcpy window closed by user; will auto-relaunch on next tool call")
					return
				}
			}
			// Device disconnected -> return to WAITING_FOR_DEVICE
			l.mu.Lock()
			l.activeSerial = ""
			l.activeModel = ""
			l.mu.Unlock()
			l.SetState(StateWaitingForDevice, "device disconnected")
		}

	case StateClosedByUser:
		// Do NOT automatically restart in background loop when closed by user!
		// Wait for EnsureLiveView() to be invoked on next MCP tool call.
		return
	}
}

func (l *LiveViewManager) EnsureLiveView(ctx context.Context) error {
	l.relaunchMu.Lock()
	defer l.relaunchMu.Unlock()

	l.mu.RLock()
	activeSer := l.activeSerial
	activeMod := l.activeModel
	enabled := l.enabled
	l.mu.RUnlock()

	if !enabled {
		return nil
	}

	appCfg, errCfg := config.LoadConfig()
	autoRelaunch := true
	requireLiveView := true
	if errCfg == nil && appCfg != nil {
		autoRelaunch = appCfg.Scrcpy.AutoRelaunchOnToolCall
		requireLiveView = appCfg.Scrcpy.RequireLiveView
	}

	if l.scrcpyMgr != nil && activeSer != "" && l.scrcpyMgr.IsRunning(activeSer) {
		return nil
	}

	if !autoRelaunch {
		if requireLiveView {
			return fmt.Errorf("scrcpy live view is closed and auto_relaunch_on_tool_call is false")
		}
		return nil
	}

	// Resolve device if activeSerial is empty
	if activeSer == "" && l.resolver != nil {
		ser, mod, err := l.resolver(ctx)
		if err != nil || ser == "" {
			if requireLiveView {
				return fmt.Errorf("cannot restore live view: no connected Android device found (%v)", err)
			}
			return nil
		}
		l.mu.Lock()
		l.activeSerial = ser
		l.activeModel = mod
		activeSer = ser
		activeMod = mod
		l.mu.Unlock()
	}

	logging.Infof("EnsureLiveView: Restoring scrcpy live display mirror for device %s...", activeSer)
	l.SetState(StateStartingScrcpy, "relaunching for MCP tool call")

	title := fmt.Sprintf("Android-MCP — %s (%s)", activeMod, activeSer)
	err := l.scrcpyMgr.LaunchWithFallback(ctx, l.adbClient, activeSer, title)
	if err != nil {
		l.SetState(StateError, fmt.Sprintf("failed to relaunch scrcpy: %v", err))
		if requireLiveView {
			return fmt.Errorf("Live view could not be restored; action was not executed: %w", err)
		}
		return nil
	}

	l.SetState(StateRunning, "")
	_ = l.notifier.Notify("Android-MCP", fmt.Sprintf("%s live view restored.", activeMod))
	return nil
}
