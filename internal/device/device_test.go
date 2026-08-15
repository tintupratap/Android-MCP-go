package device

import (
	"testing"
)

func TestDevicePreference(t *testing.T) {
	pref := DevicePreference{
		Connection: "wifi",
		Serial:     "192.168.1.5",
		Source:     "--wifi",
	}

	if pref.Connection != "wifi" || pref.Serial != "192.168.1.5" {
		t.Fatalf("unexpected preference struct: %+v", pref)
	}
}

func TestDeviceStates(t *testing.T) {
	dm := NewDeviceManager(nil, nil, DevicePreference{})
	if dm.State() != StateNoDevice {
		t.Fatalf("expected initial state NoDevice, got %s", dm.State())
	}
}
