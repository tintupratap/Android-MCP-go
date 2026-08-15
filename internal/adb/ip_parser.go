package adb

import (
	"fmt"
	"net"
	"strings"
)

// ParseIPFromAddrShow parses IPv4 address from `ip -4 addr show` output
func ParseIPFromAddrShow(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "inet ") && (strings.Contains(line, "wlan") || strings.Contains(line, "ap") || strings.Contains(line, "eth")) {
			fields := strings.Fields(line)
			for i, f := range fields {
				if f == "inet" && i+1 < len(fields) {
					ipCIDR := fields[i+1]
					parts := strings.Split(ipCIDR, "/")
					if len(parts) > 0 && isValidPrivateIPv4(parts[0]) {
						return parts[0]
					}
				}
			}
		}
	}
	return ""
}

// ParseIPFromRoute parses IPv4 address from `ip route` output
func ParseIPFromRoute(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if (strings.Contains(line, "dev wlan") || strings.Contains(line, "dev ap") || strings.Contains(line, "dev eth")) && strings.Contains(line, "src") {
			fields := strings.Fields(line)
			for i, f := range fields {
				if f == "src" && i+1 < len(fields) {
					candidate := fields[i+1]
					if isValidPrivateIPv4(candidate) {
						return candidate
					}
				}
			}
		}
	}
	return ""
}

// ParseIPFromFallbackAddrShow parses any valid non-loopback, non-cellular IPv4 address from `ip -4 addr show`
func ParseIPFromFallbackAddrShow(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "inet ") && !strings.Contains(line, " lo") && !strings.Contains(line, " rmnet") && !strings.Contains(line, " 127.0.0.1") {
			fields := strings.Fields(line)
			for i, f := range fields {
				if f == "inet" && i+1 < len(fields) {
					ipCIDR := fields[i+1]
					parts := strings.Split(ipCIDR, "/")
					if len(parts) > 0 {
						candidate := parts[0]
						if isValidPrivateIPv4(candidate) {
							return candidate
						}
					}
				}
			}
		}
	}
	return ""
}

func isValidPrivateIPv4(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	ip = ip.To4()
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsMulticast() {
		return false
	}
	// 127.0.0.0/8
	if ip[0] == 127 {
		return false
	}
	return true
}

func FormatWiFiSerial(host string, port int) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if strings.Contains(host, ":") {
		return host
	}
	if port <= 0 {
		port = 5555
	}
	return fmt.Sprintf("%s:%d", host, port)
}
