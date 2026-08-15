package notification

import (
	"os/exec"
	"runtime"
	"strings"

	"android-mcp-go/internal/logging"
)

type Notifier interface {
	Notify(title, message string) error
}

type DefaultNotifier struct{}

func NewNotifier() Notifier {
	return &DefaultNotifier{}
}

func (n *DefaultNotifier) Notify(title, message string) error {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = notifyDarwin(title, message)
	case "linux":
		err = notifyLinux(title, message)
	default:
		logging.Infof("[Notification] %s: %s", title, message)
		return nil
	}

	if err != nil {
		logging.Debugf("Notification failed (non-fatal): %v", err)
		logging.Infof("[Notification] %s: %s", title, message)
	}
	return nil
}

func notifyDarwin(title, message string) error {
	tnPath, err := exec.LookPath("terminal-notifier")
	if err == nil {
		cmd := exec.Command(tnPath, "-title", title, "-message", message)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to osascript if terminal-notifier is not installed
	osascriptPath, err := exec.LookPath("osascript")
	if err == nil {
		script := strings.NewReplacer(`"`, `\"`).Replace(message)
		titleEsc := strings.NewReplacer(`"`, `\"`).Replace(title)
		cmd := exec.Command(osascriptPath, "-e", `display notification "`+script+`" with title "`+titleEsc+`"`)
		return cmd.Run()
	}

	return nil
}

func notifyLinux(title, message string) error {
	nsPath, err := exec.LookPath("notify-send")
	if err != nil {
		return err
	}
	cmd := exec.Command(nsPath, title, message)
	return cmd.Run()
}
