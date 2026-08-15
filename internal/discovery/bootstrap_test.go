package discovery

import (
	"testing"
)

type mockNotifier struct {
	called bool
}

func (m *mockNotifier) Notify(title, message string) error {
	m.called = true
	return nil
}

func TestWirelessBootstrapperInit(t *testing.T) {
	n := &mockNotifier{}
	wb := NewWirelessBootstrapper(nil, n)
	if wb == nil {
		t.Fatalf("failed to create WirelessBootstrapper")
	}
}
