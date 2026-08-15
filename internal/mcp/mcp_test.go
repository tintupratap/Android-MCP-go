package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"android-mcp-go/internal/device"
)

func TestMCPServerProtocol(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	dm := device.NewDeviceManager(nil, nil, device.DevicePreference{})
	inBuf := &bytes.Buffer{}
	outBuf := &bytes.Buffer{}

	srv := NewServer(dm, nil, inBuf, outBuf)

	// Send initialize request
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}` + "\n"
	inBuf.WriteString(initReq)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Direct call to handleRequest
	var req JSONRPCRequest
	_ = json.Unmarshal([]byte(initReq), &req)
	srv.handleRequest(ctx, req)

	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("expected response for initialize")
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[0]), &resp); err != nil {
		t.Fatalf("failed to parse json-rpc response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error in initialize response: %+v", resp.Error)
	}

	// Test tools/list request
	outBuf.Reset()
	srv.handleRequest(ctx, JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	})

	var toolsResp JSONRPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &toolsResp); err != nil {
		t.Fatalf("failed to unmarshal tools/list response: %v", err)
	}

	data, _ := json.Marshal(toolsResp.Result)
	var listRes ToolsListResult
	_ = json.Unmarshal(data, &listRes)

	if len(listRes.Tools) < 14 {
		t.Fatalf("expected at least 14 MCP tools, got %d", len(listRes.Tools))
	}
}

func TestLazyDeviceResolutionNoCrash(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	// Point PATH to a directory without adb to guarantee no device found
	t.Setenv("PATH", filepath.Join(tempHome, "bin"))
	t.Setenv("ADB", filepath.Join(tempHome, "bin", "adb"))

	dm := device.NewDeviceManager(nil, nil, device.DevicePreference{})
	srv := NewServer(dm, nil, nil, nil)

	// Server startup / tool resolution should not panic when no device is connected
	_, err := srv.requireDevice(context.Background())
	if err == nil {
		t.Fatalf("expected error when no device connected, got nil")
	}

	if !strings.Contains(err.Error(), "No device configured") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
