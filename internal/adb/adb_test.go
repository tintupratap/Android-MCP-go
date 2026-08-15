package adb

import (
	"testing"
)

func TestParseADBDevices(t *testing.T) {
	sampleOutput := `List of devices attached
QV771A3JEE	device product:SOG09 model:SOG09 device:SOG09 transport_id:1
192.168.1.3:5555	device product:SOG09 model:SOG09 device:SOG09 transport_id:2
emulator-5554	device product:sdk_gphone64_arm64 model:sdk_gphone64_arm64 device:emulator_arm64 transport_id:3
`

	devices := ParseADBDevices(sampleOutput)
	if len(devices) != 3 {
		t.Fatalf("expected 3 devices, got %d", len(devices))
	}

	usb := devices[0]
	if !usb.IsUSB() || usb.Serial != "QV771A3JEE" || usb.Model != "SOG09" {
		t.Fatalf("unexpected USB device parse: %+v", usb)
	}

	wifi := devices[1]
	if !wifi.IsWiFi() || wifi.Serial != "192.168.1.3:5555" {
		t.Fatalf("unexpected WiFi device parse: %+v", wifi)
	}

	emu := devices[2]
	if !emu.IsEmulator() || emu.Serial != "emulator-5554" {
		t.Fatalf("unexpected Emulator device parse: %+v", emu)
	}
}

func TestParseIPFromAddrShow(t *testing.T) {
	output := `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
28: wlan0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 3000
    link/ether 02:00:00:00:00:00 brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.45/24 brd 192.168.1.255 scope global wlan0
       valid_lft forever preferred_lft forever
`

	ip := ParseIPFromAddrShow(output)
	if ip != "192.168.1.45" {
		t.Fatalf("expected 192.168.1.45, got %q", ip)
	}
}

func TestParseIPFromRoute(t *testing.T) {
	output := `192.168.1.0/24 dev wlan0 proto kernel scope link src 192.168.1.88`
	ip := ParseIPFromRoute(output)
	if ip != "192.168.1.88" {
		t.Fatalf("expected 192.168.1.88, got %q", ip)
	}
}

func TestFormatWiFiSerial(t *testing.T) {
	if got := FormatWiFiSerial("192.168.1.5", 5555); got != "192.168.1.5:5555" {
		t.Errorf("expected 192.168.1.5:5555, got %q", got)
	}
	if got := FormatWiFiSerial("192.168.1.5:5555", 5555); got != "192.168.1.5:5555" {
		t.Errorf("expected 192.168.1.5:5555, got %q", got)
	}
}
