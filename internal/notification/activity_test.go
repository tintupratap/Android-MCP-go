package notification

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSanitizeDescription(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Clicked Login button", "Clicked Login button"},
		{"Typed password mySecret123", "Executed action containing sensitive parameters (redacted)"},
		{"shell_exec token=12345", "Executed action containing sensitive parameters (redacted)"},
		{strings.Repeat("a", 100), strings.Repeat("a", 77) + "..."},
	}

	for _, tt := range tests {
		got := SanitizeDescription(tt.input)
		if got != tt.expected {
			t.Errorf("SanitizeDescription(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGenerateActionID(t *testing.T) {
	id1 := GenerateActionID()
	id2 := GenerateActionID()
	if len(id1) == 0 || len(id2) == 0 {
		t.Fatalf("expected non-empty action IDs")
	}
}

func TestActivityNotifierDebugMode(t *testing.T) {
	mock := &mockNotifier{}
	an := NewActivityNotifier(mock, LevelDebug)

	ctx := context.Background()
	act := Activity{
		ActionID:    "abc123",
		Tool:        "click",
		Description: "Clicked Login button",
		DurationMS:  15,
		Timestamp:   time.Now(),
	}

	an.NotifyActivity(ctx, act)
	time.Sleep(350 * time.Millisecond)

	if !mock.WasCalled() {
		t.Fatalf("expected activity notification to trigger in debug mode")
	}
}

func TestActivityNotifierSilentMode(t *testing.T) {
	mock := &mockNotifier{}
	an := NewActivityNotifier(mock, LevelSilent)

	ctx := context.Background()
	act := Activity{
		ActionID:    "abc123",
		Tool:        "click",
		Description: "Clicked Login button",
		DurationMS:  15,
		Timestamp:   time.Now(),
	}

	an.NotifyActivity(ctx, act)
	time.Sleep(350 * time.Millisecond)

	if mock.WasCalled() {
		t.Fatalf("expected no activity notification in silent mode")
	}
}
