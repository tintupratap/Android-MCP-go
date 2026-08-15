package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(LevelInfo, buf)

	logger.Debugf("this should be hidden")
	if buf.Len() > 0 {
		t.Fatalf("expected debug log to be hidden, got %q", buf.String())
	}

	logger.Infof("hello %s", "world")
	if !strings.Contains(buf.String(), "INFO  hello world") {
		t.Fatalf("expected info log message, got %q", buf.String())
	}

	buf.Reset()
	logger.Warnf("warning test")
	if !strings.Contains(buf.String(), "WARN  warning test") {
		t.Fatalf("expected warn log message, got %q", buf.String())
	}

	buf.Reset()
	logger.Errorf("error test")
	if !strings.Contains(buf.String(), "ERROR error test") {
		t.Fatalf("expected error log message, got %q", buf.String())
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"invalid", LevelInfo},
	}

	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.expected {
			t.Errorf("ParseLevel(%q) = %v; want %v", tt.input, got, tt.expected)
		}
	}
}
