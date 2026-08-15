package notification

import (
	"sync"
	"testing"
)

type mockNotifier struct {
	mu          sync.Mutex
	lastTitle   string
	lastMessage string
	called      bool
}

func (m *mockNotifier) Notify(title, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastTitle = title
	m.lastMessage = message
	m.called = true
	return nil
}

func (m *mockNotifier) WasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

func TestNotifier(t *testing.T) {
	n := NewNotifier()
	// Notify should never panic or return fatal error
	if err := n.Notify("Android-MCP Test", "Test message content"); err != nil {
		t.Fatalf("expected non-fatal notification behavior, got: %v", err)
	}

	mock := &mockNotifier{}
	if err := mock.Notify("Title", "Message"); err != nil {
		t.Fatalf("mock failed: %v", err)
	}
	if !mock.WasCalled() {
		t.Fatalf("expected mock to be called")
	}
}
