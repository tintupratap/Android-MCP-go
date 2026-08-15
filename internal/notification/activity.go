package notification

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/logging"
)

type NotificationLevel int

const (
	LevelSilent NotificationLevel = iota
	LevelNormal
	LevelDebug
)

type Activity struct {
	ActionID    string    `json:"action_id"`
	Tool        string    `json:"tool"`
	Device      string    `json:"device,omitempty"`
	Target      string    `json:"target,omitempty"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Success     bool      `json:"success"`
	DurationMS  int64     `json:"duration_ms"`
}

type ActivityNotifier interface {
	NotifyActivity(ctx context.Context, activity Activity)
	SetLevel(level NotificationLevel)
	Level() NotificationLevel
}

type DefaultActivityNotifier struct {
	mu           sync.Mutex
	notifier     Notifier
	level        NotificationLevel
	rateInterval time.Duration
	lastNotify   time.Time
	queue        chan Activity
	closeOnce    sync.Once
}

func NewActivityNotifier(notifier Notifier, level NotificationLevel) *DefaultActivityNotifier {
	if notifier == nil {
		notifier = NewNotifier()
	}
	an := &DefaultActivityNotifier{
		notifier:     notifier,
		level:        level,
		rateInterval: 250 * time.Millisecond,
		queue:        make(chan Activity, 100),
	}
	go an.worker()
	return an
}

func GenerateActionID() string {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano()%0xFFFFFF)
	}
	return hex.EncodeToString(b)
}

func (an *DefaultActivityNotifier) SetLevel(level NotificationLevel) {
	an.mu.Lock()
	defer an.mu.Unlock()
	an.level = level
}

func (an *DefaultActivityNotifier) Level() NotificationLevel {
	an.mu.Lock()
	defer an.mu.Unlock()
	return an.level
}

func (an *DefaultActivityNotifier) NotifyActivity(ctx context.Context, activity Activity) {
	an.mu.Lock()
	lvl := an.level
	an.mu.Unlock()

	if lvl < LevelDebug {
		return
	}

	// Sanitize sensitive target or tool content
	activity.Description = SanitizeDescription(activity.Description)

	logging.Debugf("[ACTION %s] tool=%s device=%s duration=%dms desc=%s",
		activity.ActionID, activity.Tool, activity.Device, activity.DurationMS, activity.Description)

	select {
	case an.queue <- activity:
	default:
		// Channel full, drop non-critical debug notification
	}
}

func (an *DefaultActivityNotifier) worker() {
	for act := range an.queue {
		an.mu.Lock()
		elapsed := time.Since(an.lastNotify)
		if elapsed < an.rateInterval {
			time.Sleep(an.rateInterval - elapsed)
		}
		an.lastNotify = time.Now()
		an.mu.Unlock()

		title := "Android-MCP (AI Action)"
		msg := fmt.Sprintf("AI: %s", act.Description)
		_ = an.notifier.Notify(title, msg)
	}
}

func SanitizeDescription(desc string) string {
	low := strings.ToLower(desc)
	if strings.Contains(low, "password") || strings.Contains(low, "token") || strings.Contains(low, "secret") || strings.Contains(low, "key=") {
		return "Executed action containing sensitive parameters (redacted)"
	}
	if len(desc) > 80 {
		return desc[:77] + "..."
	}
	return desc
}
