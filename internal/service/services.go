package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/ui"
)

type Services struct {
	Device *DeviceService
	Input  *InputService
	Screen *ScreenService
	UI     *UIService
	App    *AppService
	File   *FileService
	Shell  *ShellService
}

func NewServices(client *adb.Client) *Services {
	if client == nil {
		client = adb.NewClient("")
	}
	return &Services{
		Device: &DeviceService{adbClient: client},
		Input:  &InputService{adbClient: client},
		Screen: &ScreenService{adbClient: client},
		UI:     &UIService{adbClient: client},
		App:    &AppService{adbClient: client},
		File:   &FileService{adbClient: client},
		Shell:  &ShellService{adbClient: client},
	}
}

// -----------------------------------------------------------------------------
// InputService
// -----------------------------------------------------------------------------
type InputService struct {
	adbClient *adb.Client
}

func (s *InputService) Click(ctx context.Context, serial string, x, y int) error {
	return s.adbClient.Tap(ctx, serial, x, y)
}

func (s *InputService) LongClick(ctx context.Context, serial string, x, y int, durationMs int) error {
	if durationMs <= 0 {
		durationMs = 1000
	}
	return s.adbClient.Swipe(ctx, serial, x, y, x, y, durationMs)
}

func (s *InputService) Swipe(ctx context.Context, serial string, x1, y1, x2, y2, durationMs int) error {
	return s.adbClient.Swipe(ctx, serial, x1, y1, x2, y2, durationMs)
}

func (s *InputService) Drag(ctx context.Context, serial string, x1, y1, x2, y2, durationMs int) error {
	return s.adbClient.Drag(ctx, serial, x1, y1, x2, y2, durationMs)
}

func (s *InputService) Pinch(ctx context.Context, serial string, x1, y1, x2, y2, x3, y3, x4, y4, durationMs int) error {
	return s.adbClient.Pinch(ctx, serial, x1, y1, x2, y2, x3, y3, x4, y4, durationMs)
}

func (s *InputService) Type(ctx context.Context, serial string, x, y int, text string, clear bool) error {
	if x > 0 && y > 0 {
		_ = s.adbClient.Tap(ctx, serial, x, y)
		time.Sleep(200 * time.Millisecond)
	}
	if clear {
		_ = s.adbClient.KeyEvent(ctx, serial, "KEYCODE_MOVE_END")
		for i := 0; i < 50; i++ {
			_ = s.adbClient.KeyEvent(ctx, serial, "KEYCODE_DEL")
		}
	}
	return s.adbClient.SendKeys(ctx, serial, text)
}

func (s *InputService) Press(ctx context.Context, serial string, keycode string) error {
	return s.adbClient.KeyEvent(ctx, serial, keycode)
}

// -----------------------------------------------------------------------------
// ScreenService
// -----------------------------------------------------------------------------
type ScreenService struct {
	adbClient *adb.Client
}

func (s *ScreenService) Screencap(ctx context.Context, serial string) ([]byte, error) {
	return s.adbClient.Screencap(ctx, serial)
}

// -----------------------------------------------------------------------------
// UIService
// -----------------------------------------------------------------------------
type UIService struct {
	adbClient *adb.Client
}

func (s *UIService) DumpHierarchy(ctx context.Context, serial string) (string, error) {
	return s.adbClient.DumpHierarchy(ctx, serial)
}

func (s *UIService) ParseTree(xmlData string) (*ui.TreeState, error) {
	return ui.ParseTreeState(xmlData)
}

func (s *UIService) FindElement(xmlData string, sel ui.Selector, pkg string) (*ui.ElementNode, error) {
	return ui.FindElementBySelector(xmlData, sel, pkg)
}

func (s *UIService) Annotate(imgBytes []byte, nodes []ui.ElementNode, scale float64) ([]byte, error) {
	return ui.AnnotateScreenshot(imgBytes, nodes, scale)
}

// -----------------------------------------------------------------------------
// AppService
// -----------------------------------------------------------------------------
type AppService struct {
	adbClient *adb.Client
}

func (s *AppService) ListApps(ctx context.Context, serial string, thirdPartyOnly bool) ([]string, error) {
	args := []string{"pm", "list", "packages"}
	if thirdPartyOnly {
		args = append(args, "-3")
	}
	out, err := s.adbClient.ExecuteShell(ctx, serial, args...)
	if err != nil {
		return nil, err
	}
	var pkgs []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package:") {
			pkgs = append(pkgs, strings.TrimPrefix(line, "package:"))
		}
	}
	return pkgs, nil
}

func (s *AppService) LaunchApp(ctx context.Context, serial string, packageName string) error {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return fmt.Errorf("package_name is required")
	}
	_, err := s.adbClient.ExecuteShell(ctx, serial, "monkey", "-p", packageName, "-c", "android.intent.category.LAUNCHER", "1")
	return err
}

func (s *AppService) StopApp(ctx context.Context, serial string, packageName string) error {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return fmt.Errorf("package_name is required")
	}
	_, err := s.adbClient.ExecuteShell(ctx, serial, "am", "force-stop", packageName)
	return err
}

// -----------------------------------------------------------------------------
// FileService
// -----------------------------------------------------------------------------
type FileService struct {
	adbClient *adb.Client
}

func (s *FileService) Push(ctx context.Context, serial, localPath, remotePath string) error {
	return s.adbClient.PushFile(ctx, serial, localPath, remotePath)
}

func (s *FileService) Pull(ctx context.Context, serial, remotePath, localPath string) error {
	return s.adbClient.PullFile(ctx, serial, remotePath, localPath)
}

// -----------------------------------------------------------------------------
// ShellService
// -----------------------------------------------------------------------------
type ShellService struct {
	adbClient *adb.Client
}

type ShellResult struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMS int64  `json:"duration_ms"`
}

func (s *ShellService) Exec(ctx context.Context, serial string, command string, timeoutSec int) (*ShellResult, error) {
	if timeoutSec <= 0 {
		timeoutSec = 15
	}
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	start := time.Now()
	out, err := s.adbClient.ExecuteShell(execCtx, serial, command)
	duration := time.Since(start).Milliseconds()

	exitCode := 0
	errMsg := ""
	if err != nil {
		exitCode = 1
		errMsg = err.Error()
	}

	return &ShellResult{
		Stdout:     out,
		Stderr:     errMsg,
		ExitCode:   exitCode,
		DurationMS: duration,
	}, nil
}

// -----------------------------------------------------------------------------
// DeviceService
// -----------------------------------------------------------------------------
type DeviceService struct {
	adbClient *adb.Client
}

type DeviceInfo struct {
	Serial             string `json:"serial"`
	Model              string `json:"model"`
	Manufacturer       string `json:"manufacturer"`
	AndroidVersion     string `json:"android_version"`
	SDKVersion         string `json:"sdk_version"`
	ABI                string `json:"abi"`
	ScreenDensity      string `json:"screen_density"`
	WiFiIP             string `json:"wifi_ip"`
}

func (s *DeviceService) GetInfo(ctx context.Context, serial string) (*DeviceInfo, error) {
	model, _ := s.adbClient.GetProp(ctx, serial, "ro.product.model")
	mfr, _ := s.adbClient.GetProp(ctx, serial, "ro.product.manufacturer")
	ver, _ := s.adbClient.GetProp(ctx, serial, "ro.build.version.release")
	sdk, _ := s.adbClient.GetProp(ctx, serial, "ro.build.version.sdk")
	abi, _ := s.adbClient.GetProp(ctx, serial, "ro.product.cpu.abi")
	density, _ := s.adbClient.GetProp(ctx, serial, "ro.sf.lcd_density")
	ip, _ := s.adbClient.GetDeviceIP(ctx, serial)

	return &DeviceInfo{
		Serial:         serial,
		Model:          model,
		Manufacturer:   mfr,
		AndroidVersion: ver,
		SDKVersion:     sdk,
		ABI:            abi,
		ScreenDensity:  density,
		WiFiIP:         ip,
	}, nil
}
