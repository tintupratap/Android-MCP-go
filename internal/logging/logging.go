package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	mu     sync.Mutex
	level  Level
	output io.Writer
}

var globalLogger = New(LevelInfo, os.Stderr)

func New(level Level, output io.Writer) *Logger {
	if output == nil {
		output = os.Stderr
	}
	return &Logger{
		level:  level,
		output: output,
	}
}

func SetLevel(level Level) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.level = level
}

func SetOutput(w io.Writer) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.output = w
}

func ParseLevel(lvlStr string) Level {
	switch strings.ToLower(strings.TrimSpace(lvlStr)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *Logger) logf(level Level, prefix, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("%-5s %s\n", prefix, msg)
	_, _ = l.output.Write([]byte(entry))
}

func Debugf(format string, args ...interface{}) {
	globalLogger.logf(LevelDebug, "DEBUG", format, args...)
}

func Infof(format string, args ...interface{}) {
	globalLogger.logf(LevelInfo, "INFO", format, args...)
}

func Warnf(format string, args ...interface{}) {
	globalLogger.logf(LevelWarn, "WARN", format, args...)
}

func Errorf(format string, args ...interface{}) {
	globalLogger.logf(LevelError, "ERROR", format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf(LevelDebug, "DEBUG", format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(LevelInfo, "INFO", format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf(LevelWarn, "WARN", format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(LevelError, "ERROR", format, args...)
}

// Standard log redirection helper
func RedirectStdLog(l *Logger) {
	log.SetOutput(wAdapter{l: l})
}

type wAdapter struct {
	l *Logger
}

func (w wAdapter) Write(p []byte) (n int, err error) {
	w.l.Infof("%s", strings.TrimSpace(string(p)))
	return len(p), nil
}
