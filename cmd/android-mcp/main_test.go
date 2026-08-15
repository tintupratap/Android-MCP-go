package main

import (
	"testing"
)

func TestResolvePreference(t *testing.T) {
	pref := resolvePreference("QV771A3JEE", "usb", "", "")
	if pref.Connection != "usb" || pref.Serial != "QV771A3JEE" || pref.Source != "--device" {
		t.Fatalf("unexpected preference: %+v", pref)
	}

	t.Setenv("ANDROID_MCP_DEVICE", "192.168.1.10:5555")
	t.Setenv("ANDROID_MCP_CONNECTION", "wifi")
	prefEnv := resolvePreference("", "", "", "")
	if prefEnv.Connection != "wifi" || prefEnv.Serial != "192.168.1.10:5555" || prefEnv.Source != "ANDROID_MCP_DEVICE" {
		t.Fatalf("unexpected env preference: %+v", prefEnv)
	}
}
