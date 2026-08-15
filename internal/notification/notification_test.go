package notification

import (
	"testing"
)

type mockNotifier struct {
	lastTitle   string
	lastMessage string
	called      bool
}

func (m *mockNotifier) Notify(title, message string) error {
	m.lastTitle = title
	m.lastMessage = message
	m.called = true
	return nil
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
	if !mock.called || mock.lastTitle != "Title" || mock.lastMessage != "Message" {
		t.Fatalf("mock recorded incorrect data: %+v", mock)
	}
}
