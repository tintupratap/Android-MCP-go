package mcp

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/device"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
	"github.com/tintupratap/Android-MCP-go/internal/service"
	"github.com/tintupratap/Android-MCP-go/internal/ui"
)

type ToolHandler func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error)

type Server struct {
	mu            sync.Mutex
	deviceManager device.DeviceManager
	adbClient     *adb.Client
	services      *service.Services
	tools         map[string]Tool
	handlers      map[string]ToolHandler
	reader        *bufio.Reader
	writer        io.Writer
}

func NewServer(dm device.DeviceManager, client *adb.Client, in io.Reader, out io.Writer) *Server {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if client == nil {
		client = adb.NewClient("")
	}
	s := &Server{
		deviceManager: dm,
		adbClient:     client,
		services:      service.NewServices(client),
		tools:         make(map[string]Tool),
		handlers:      make(map[string]ToolHandler),
		reader:        bufio.NewReader(in),
		writer:        out,
	}
	s.registerTools()
	return s
}

func (s *Server) requireDevice(ctx context.Context) (*device.Device, error) {
	dev, err := s.deviceManager.Resolve(ctx)
	if err != nil {
		devices, listErr := s.deviceManager.List(ctx)
		availMsg := ""
		if listErr == nil && len(devices) > 0 {
			var avail []string
			for _, d := range devices {
				if d.State == "device" {
					avail = append(avail, d.Serial)
				}
			}
			if len(avail) > 0 {
				availMsg = fmt.Sprintf(" Available devices: %s.", strings.Join(avail, ", "))
			}
		}
		return nil, fmt.Errorf("No device configured. Use --device flag, --wifi, --usb, or ANDROID_MCP_DEVICE.%s (%v)", availMsg, err)
	}
	return dev, nil
}

func textResult(text string) *CallToolResult {
	return &CallToolResult{
		Content: []ContentItem{
			{Type: "text", Text: text},
		},
	}
}

func errorResult(msg string) *CallToolResult {
	return &CallToolResult{
		Content: []ContentItem{
			{Type: "text", Text: msg},
		},
		IsError: true,
	}
}

func (s *Server) registerTools() {
	// 1. ListDevices
	listDevHandler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		devs, err := s.deviceManager.List(ctx)
		if err != nil || len(devs) == 0 {
			return textResult("No devices found. Ensure a device is connected and ADB is running."), nil
		}
		var lines []string
		for _, d := range devs {
			lines = append(lines, fmt.Sprintf("%s\t%s", d.Serial, d.State))
		}
		return textResult(strings.Join(lines, "\n")), nil
	}
	s.registerTool(Tool{
		Name:        "ListDevices",
		Description: "List available ADB devices",
		InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		Annotations: &ToolAnnotations{Title: "List Devices", ReadOnlyHint: true},
	}, listDevHandler)
	s.registerTool(Tool{
		Name:        "device_list",
		Description: "List available ADB devices (alias)",
		InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		Annotations: &ToolAnnotations{Title: "Device List", ReadOnlyHint: true},
	}, listDevHandler)

	// 2. ConnectDevice
	connectDevHandler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		serial, _ := args["serial"].(string)
		if serial == "" {
			return errorResult("Error: serial argument required"), nil
		}
		dev, err := s.deviceManager.Connect(ctx, serial)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to connect: %v", err)), nil
		}
		return textResult(fmt.Sprintf("Connected to %s", dev.Serial)), nil
	}
	s.registerTool(Tool{
		Name:        "ConnectDevice",
		Description: "Connect to an ADB device by serial number",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"serial": map[string]interface{}{"type": "string"},
			},
			"required": []string{"serial"},
		},
		Annotations: &ToolAnnotations{Title: "Connect Device"},
	}, connectDevHandler)

	// 3. Device
	s.registerTool(Tool{
		Name:        "Device",
		Description: "Manage ADB devices (list, connect, or disconnect)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type": "string",
					"enum": []string{"list", "connect", "disconnect"},
				},
				"serial": map[string]interface{}{"type": "string"},
			},
			"required": []string{"action"},
		},
		Annotations: &ToolAnnotations{Title: "Device"},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		action, _ := args["action"].(string)
		serial, _ := args["serial"].(string)

		switch action {
		case "list":
			return listDevHandler(ctx, args)
		case "connect":
			if serial == "" {
				dev, err := s.deviceManager.Resolve(ctx)
				if err != nil {
					return errorResult(err.Error()), nil
				}
				return textResult(fmt.Sprintf("Connected to %s", dev.Serial)), nil
			}
			return connectDevHandler(ctx, args)
		case "disconnect":
			_ = s.deviceManager.Disconnect(ctx)
			return textResult("Disconnected from device."), nil
		default:
			return errorResult(fmt.Sprintf("Unknown action: %s", action)), nil
		}
	})

	// 4. Click
	clickHandler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		x, _ := getIntArg(args, "x")
		y, _ := getIntArg(args, "y")
		if err := s.services.Input.Click(ctx, dev.Serial, x, y); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Clicked on (%d,%d)", x, y)), nil
	}
	s.registerTool(Tool{
		Name:        "Click",
		Description: "Click on a specific cordinate",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"x": map[string]interface{}{"type": "integer"},
				"y": map[string]interface{}{"type": "integer"},
			},
			"required": []string{"x", "y"},
		},
		Annotations: &ToolAnnotations{Title: "Click", DestructiveHint: true},
	}, clickHandler)
	s.registerTool(Tool{
		Name:        "ui_click",
		Description: "Click on a specific cordinate (alias)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"x": map[string]interface{}{"type": "integer"},
				"y": map[string]interface{}{"type": "integer"},
			},
			"required": []string{"x", "y"},
		},
		Annotations: &ToolAnnotations{Title: "UI Click", DestructiveHint: true},
	}, clickHandler)

	// 5. ClickBySelector
	s.registerTool(Tool{
		Name:        "ClickBySelector",
		Description: "Click on an element by selector (text, resourceId, className, description). More reliable than coordinate clicks.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text":        map[string]interface{}{"type": "string"},
				"resourceId":  map[string]interface{}{"type": "string"},
				"className":   map[string]interface{}{"type": "string"},
				"description": map[string]interface{}{"type": "string"},
				"index":       map[string]interface{}{"type": "integer", "default": 0},
				"timeout":     map[string]interface{}{"type": "number", "default": 5.0},
			},
		},
		Annotations: &ToolAnnotations{Title: "Click By Selector", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		text, _ := args["text"].(string)
		resourceID, _ := args["resourceId"].(string)
		className, _ := args["className"].(string)
		description, _ := args["description"].(string)
		idx, _ := getIntArg(args, "index")
		timeoutSec, _ := getFloatArg(args, "timeout")
		if timeoutSec <= 0 {
			timeoutSec = 5.0
		}

		if text == "" && resourceID == "" && className == "" && description == "" {
			return textResult("Error: at least one selector (text, resourceId, className, description) must be provided"), nil
		}

		sel := ui.Selector{
			Text:        text,
			ResourceID:  resourceID,
			ClassName:   className,
			Description: description,
			Index:       idx,
		}

		deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
		for {
			xmlData, err := s.services.UI.DumpHierarchy(ctx, dev.Serial)
			if err == nil {
				elem, err := s.services.UI.FindElement(xmlData, sel, "")
				if err == nil && elem != nil {
					if err := s.services.Input.Click(ctx, dev.Serial, elem.Coordinates.X, elem.Coordinates.Y); err != nil {
						return errorResult(err.Error()), nil
					}
					return textResult(fmt.Sprintf("Clicked element matching text=%q resId=%q coords=(%d,%d)", text, resourceID, elem.Coordinates.X, elem.Coordinates.Y)), nil
				}
			}

			if time.Now().After(deadline) {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		return textResult(fmt.Sprintf("Element not found with selectors text=%q resId=%q className=%q desc=%q within %.1fs", text, resourceID, className, description, timeoutSec)), nil
	})

	// 6. Snapshot
	snapshotHandler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		useVision, _ := args["use_vision"].(bool)
		useAnnotation := true
		if val, ok := args["use_annotation"].(bool); ok {
			useAnnotation = val
		}

		xmlData, err := s.services.UI.DumpHierarchy(ctx, dev.Serial)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to dump UI hierarchy: %v", err)), nil
		}

		treeState, err := s.services.UI.ParseTree(xmlData)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to parse tree state: %v", err)), nil
		}

		res := &CallToolResult{
			Content: []ContentItem{
				{Type: "text", Text: treeState.ToString()},
			},
		}

		if useVision {
			imgBytes, err := s.services.Screen.Screencap(ctx, dev.Serial)
			if err != nil {
				return errorResult(fmt.Sprintf("Failed to capture screenshot: %v", err)), nil
			}

			if useAnnotation {
				annotatedBytes, err := s.services.UI.Annotate(imgBytes, treeState.InteractiveElements, 1.0)
				if err == nil {
					imgBytes = annotatedBytes
				}
			}

			res.Content = append(res.Content, ContentItem{
				Type: "image",
				Data: base64.StdEncoding.EncodeToString(imgBytes),
				MIME: "image/png",
			})
		}

		return res, nil
	}
	s.registerTool(Tool{
		Name:        "Snapshot",
		Description: "Get the state of the device. Optionally includes visual screenshot when use_vision=True.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"use_vision":     map[string]interface{}{"type": "boolean", "default": false},
				"use_annotation": map[string]interface{}{"type": "boolean", "default": true},
			},
		},
		Annotations: &ToolAnnotations{Title: "Snapshot", ReadOnlyHint: true},
	}, snapshotHandler)
	s.registerTool(Tool{
		Name:        "ui_snapshot",
		Description: "Get the state of the device (alias)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"use_vision":     map[string]interface{}{"type": "boolean", "default": false},
				"use_annotation": map[string]interface{}{"type": "boolean", "default": true},
			},
		},
		Annotations: &ToolAnnotations{Title: "UI Snapshot", ReadOnlyHint: true},
	}, snapshotHandler)

	// 7. LongClick
	s.registerTool(Tool{
		Name:        "LongClick",
		Description: "Long click on a specific cordinate",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"x": map[string]interface{}{"type": "integer"},
				"y": map[string]interface{}{"type": "integer"},
			},
			"required": []string{"x", "y"},
		},
		Annotations: &ToolAnnotations{Title: "Long Click", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		x, _ := getIntArg(args, "x")
		y, _ := getIntArg(args, "y")
		if err := s.services.Input.LongClick(ctx, dev.Serial, x, y, 1000); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Long Clicked on (%d,%d)", x, y)), nil
	})

	// 8. Swipe
	s.registerTool(Tool{
		Name:        "Swipe",
		Description: "Swipe on a specific cordinate",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"x1": map[string]interface{}{"type": "integer"},
				"y1": map[string]interface{}{"type": "integer"},
				"x2": map[string]interface{}{"type": "integer"},
				"y2": map[string]interface{}{"type": "integer"},
			},
			"required": []string{"x1", "y1", "x2", "y2"},
		},
		Annotations: &ToolAnnotations{Title: "Swipe", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		x1, _ := getIntArg(args, "x1")
		y1, _ := getIntArg(args, "y1")
		x2, _ := getIntArg(args, "x2")
		y2, _ := getIntArg(args, "y2")

		if err := s.services.Input.Swipe(ctx, dev.Serial, x1, y1, x2, y2, 300); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Swiped from (%d,%d) to (%d,%d)", x1, y1, x2, y2)), nil
	})

	// 9. Type
	s.registerTool(Tool{
		Name:        "Type",
		Description: "Type on a specific cordinate",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text":  map[string]interface{}{"type": "string"},
				"x":     map[string]interface{}{"type": "integer"},
				"y":     map[string]interface{}{"type": "integer"},
				"clear": map[string]interface{}{"type": "boolean", "default": false},
			},
			"required": []string{"text", "x", "y"},
		},
		Annotations: &ToolAnnotations{Title: "Type", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		textStr, _ := args["text"].(string)
		x, _ := getIntArg(args, "x")
		y, _ := getIntArg(args, "y")
		clear, _ := args["clear"].(bool)

		if err := s.services.Input.Type(ctx, dev.Serial, x, y, textStr, clear); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Typed %q on (%d,%d)", textStr, x, y)), nil
	})

	// 10. Drag
	s.registerTool(Tool{
		Name:        "Drag",
		Description: "Drag from location and drop on another location",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"x1": map[string]interface{}{"type": "integer"},
				"y1": map[string]interface{}{"type": "integer"},
				"x2": map[string]interface{}{"type": "integer"},
				"y2": map[string]interface{}{"type": "integer"},
			},
			"required": []string{"x1", "y1", "x2", "y2"},
		},
		Annotations: &ToolAnnotations{Title: "Drag", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		x1, _ := getIntArg(args, "x1")
		y1, _ := getIntArg(args, "y1")
		x2, _ := getIntArg(args, "x2")
		y2, _ := getIntArg(args, "y2")

		if err := s.services.Input.Swipe(ctx, dev.Serial, x1, y1, x2, y2, 800); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Dragged from (%d,%d) and dropped on (%d,%d)", x1, y1, x2, y2)), nil
	})

	// 11. Press
	s.registerTool(Tool{
		Name:        "Press",
		Description: "Press on specific button on the device",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"button": map[string]interface{}{"type": "string"},
			},
			"required": []string{"button"},
		},
		Annotations: &ToolAnnotations{Title: "Press", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		btn, _ := args["button"].(string)
		if err := s.services.Input.Press(ctx, dev.Serial, btn); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Pressed the %q button", btn)), nil
	})

	// 12. Notification
	s.registerTool(Tool{
		Name:        "Notification",
		Description: "Access the notifications seen on the device",
		InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		Annotations: &ToolAnnotations{Title: "Notification", DestructiveHint: true, IdempotentHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		if err := s.adbClient.OpenNotification(ctx, dev.Serial); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult("Accessed notification bar"), nil
	})

	// 13. Wait
	s.registerTool(Tool{
		Name:        "Wait",
		Description: "Wait for a specific amount of time",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"duration": map[string]interface{}{"type": "integer"},
			},
			"required": []string{"duration"},
		},
		Annotations: &ToolAnnotations{Title: "Wait", DestructiveHint: true, IdempotentHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		durSec, _ := getIntArg(args, "duration")
		if durSec > 0 {
			time.Sleep(time.Duration(durSec) * time.Second)
		}
		return textResult(fmt.Sprintf("Waited for %d seconds", durSec)), nil
	})

	// 14. WaitForElement
	s.registerTool(Tool{
		Name:        "WaitForElement",
		Description: "Wait for an element to appear on screen.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text":        map[string]interface{}{"type": "string"},
				"resourceId":  map[string]interface{}{"type": "string"},
				"className":   map[string]interface{}{"type": "string"},
				"description": map[string]interface{}{"type": "string"},
				"timeout":     map[string]interface{}{"type": "number", "default": 10.0},
			},
		},
		Annotations: &ToolAnnotations{Title: "Wait For Element", ReadOnlyHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		text, _ := args["text"].(string)
		resourceID, _ := args["resourceId"].(string)
		className, _ := args["className"].(string)
		description, _ := args["description"].(string)
		timeoutSec, _ := getFloatArg(args, "timeout")
		if timeoutSec <= 0 {
			timeoutSec = 10.0
		}

		if text == "" && resourceID == "" && className == "" && description == "" {
			return textResult("Error: at least one selector (text, resourceId, className, description) must be provided"), nil
		}

		sel := ui.Selector{
			Text:        text,
			ResourceID:  resourceID,
			ClassName:   className,
			Description: description,
		}

		deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
		for {
			xmlData, err := s.services.UI.DumpHierarchy(ctx, dev.Serial)
			if err == nil {
				elem, err := s.services.UI.FindElement(xmlData, sel, "")
				if err == nil && elem != nil {
					return textResult(fmt.Sprintf("Element found: text=%q class=%s coords=(%d,%d) bounds=[%d,%d][%d,%d]", elem.Name, elem.ClassName, elem.Coordinates.X, elem.Coordinates.Y, elem.BoundingBox.X1, elem.BoundingBox.Y1, elem.BoundingBox.X2, elem.BoundingBox.Y2)), nil
				}
			}

			if time.Now().After(deadline) {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		return textResult(fmt.Sprintf("Element not found with selectors text=%q resId=%q className=%q desc=%q within %.1fs", text, resourceID, className, description, timeoutSec)), nil
	})

	// 15. list_apps
	s.registerTool(Tool{
		Name:        "list_apps",
		Description: "List installed application packages on device",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"third_party_only": map[string]interface{}{"type": "boolean", "default": false},
			},
		},
		Annotations: &ToolAnnotations{Title: "List Applications", ReadOnlyHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		thirdParty, _ := args["third_party_only"].(bool)
		pkgs, err := s.services.App.ListApps(ctx, dev.Serial, thirdParty)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(strings.Join(pkgs, "\n")), nil
	})

	// 16. launch_app
	s.registerTool(Tool{
		Name:        "launch_app",
		Description: "Launch an application package on device",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"package_name": map[string]interface{}{"type": "string"},
			},
			"required": []string{"package_name"},
		},
		Annotations: &ToolAnnotations{Title: "Launch Application", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		pkg, _ := args["package_name"].(string)
		if err := s.services.App.LaunchApp(ctx, dev.Serial, pkg); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Launched %s", pkg)), nil
	})

	// 17. stop_app
	s.registerTool(Tool{
		Name:        "stop_app",
		Description: "Force-stop an application package on device",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"package_name": map[string]interface{}{"type": "string"},
			},
			"required": []string{"package_name"},
		},
		Annotations: &ToolAnnotations{Title: "Stop Application", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		pkg, _ := args["package_name"].(string)
		if err := s.services.App.StopApp(ctx, dev.Serial, pkg); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Stopped %s", pkg)), nil
	})

	// 18. file_push
	s.registerTool(Tool{
		Name:        "file_push",
		Description: "Transfer a file from host to Android filesystem",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"local_path":  map[string]interface{}{"type": "string"},
				"remote_path": map[string]interface{}{"type": "string"},
			},
			"required": []string{"local_path", "remote_path"},
		},
		Annotations: &ToolAnnotations{Title: "Push File", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		loc, _ := args["local_path"].(string)
		rem, _ := args["remote_path"].(string)
		if err := s.services.File.Push(ctx, dev.Serial, loc, rem); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Pushed %s to %s", loc, rem)), nil
	})

	// 19. file_pull
	s.registerTool(Tool{
		Name:        "file_pull",
		Description: "Transfer a file from Android filesystem to host",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"remote_path": map[string]interface{}{"type": "string"},
				"local_path":  map[string]interface{}{"type": "string"},
			},
			"required": []string{"remote_path", "local_path"},
		},
		Annotations: &ToolAnnotations{Title: "Pull File", ReadOnlyHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		rem, _ := args["remote_path"].(string)
		loc, _ := args["local_path"].(string)
		if err := s.services.File.Pull(ctx, dev.Serial, rem, loc); err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(fmt.Sprintf("Pulled %s to %s", rem, loc)), nil
	})

	// 20. shell_exec
	s.registerTool(Tool{
		Name:        "shell_exec",
		Description: "Execute a shell command on the Android device",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command":         map[string]interface{}{"type": "string"},
				"timeout_seconds": map[string]interface{}{"type": "integer", "default": 15},
			},
			"required": []string{"command"},
		},
		Annotations: &ToolAnnotations{Title: "Execute Shell Command", DestructiveHint: true},
	}, func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		dev, err := s.requireDevice(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		cmdStr, _ := args["command"].(string)
		timeoutSec, _ := getIntArg(args, "timeout_seconds")

		res, err := s.services.Shell.Exec(ctx, dev.Serial, cmdStr, timeoutSec)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		jsonBytes, _ := json.MarshalIndent(res, "", "  ")
		return textResult(string(jsonBytes)), nil
	})
}

func (s *Server) registerTool(tool Tool, handler ToolHandler) {
	s.tools[tool.Name] = tool
	s.handlers[tool.Name] = handler
}

func getIntArg(args map[string]interface{}, key string) (int, bool) {
	val, ok := args[key]
	if !ok || val == nil {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		i, err := strconv.Atoi(v)
		return i, err == nil
	}
	return 0, false
}

func getFloatArg(args map[string]interface{}, key string) (float64, bool) {
	val, ok := args[key]
	if !ok || val == nil {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	}
	return 0, false
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	scanner := bufio.NewScanner(s.reader)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	logging.Infof("Android-MCP server listening on stdio...")

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		s.handleRequest(ctx, req)
	}

	return scanner.Err()
}

func (s *Server) handleRequest(ctx context.Context, req JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		var params InitializeParams
		if req.Params != nil {
			_ = json.Unmarshal(req.Params, &params)
		}
		result := InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			ServerInfo: ServerInfo{
				Name:    "Android-MCP",
				Version: "0.2.0",
			},
			Instructions: "Android MCP server provides tools to interact directly with the Android device, enabling to operate the mobile device like an actual USER.",
		}
		s.sendResponse(req.ID, result)

	case "notifications/initialized":
		// No response required for notification

	case "tools/list":
		var list []Tool
		for _, t := range s.tools {
			list = append(list, t)
		}
		s.sendResponse(req.ID, ToolsListResult{Tools: list})

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, -32602, "Invalid params", err.Error())
			return
		}

		handler, ok := s.handlers[params.Name]
		if !ok {
			s.sendError(req.ID, -32601, "Method not found", fmt.Sprintf("Unknown tool: %s", params.Name))
			return
		}

		res, err := handler(ctx, params.Arguments)
		if err != nil {
			s.sendError(req.ID, -32603, "Internal error", err.Error())
			return
		}

		s.sendResponse(req.ID, res)

	default:
		if req.ID != nil {
			s.sendError(req.ID, -32601, "Method not found", fmt.Sprintf("Unsupported method: %s", req.Method))
		}
	}
}

func (s *Server) sendResponse(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.writer.Write(append(data, '\n'))
}

func (s *Server) sendError(id interface{}, code int, message string, details string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    details,
		},
	}
	data, _ := json.Marshal(resp)
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.writer.Write(append(data, '\n'))
}
