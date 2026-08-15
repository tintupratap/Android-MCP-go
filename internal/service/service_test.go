package service

import (
	"testing"
)

func TestNewServices(t *testing.T) {
	s := NewServices(nil)
	if s == nil || s.Input == nil || s.App == nil || s.File == nil || s.Shell == nil {
		t.Fatalf("failed to initialize services struct")
	}
}
