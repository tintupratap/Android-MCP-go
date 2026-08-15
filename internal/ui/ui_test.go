package ui

import (
	"image"
	"image/color"
	"image/png"
	"bytes"
	"strings"
	"testing"
)

const sampleXML = `<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy rotation="0">
  <node index="0" text="" resource-id="" class="android.widget.FrameLayout" package="com.example.app" content-desc="" checkable="false" checked="false" clickable="false" enabled="true" focusable="false" focused="false" scrollable="false" long-clickable="false" password="false" selected="false" bounds="[0,0][1080,2400]">
    <node index="0" text="Login" resource-id="com.example.app:id/btn_login" class="android.widget.Button" package="com.example.app" content-desc="" checkable="false" checked="false" clickable="true" enabled="true" focusable="true" focused="false" scrollable="false" long-clickable="false" password="false" selected="false" bounds="[100,500][980,650]" />
    <node index="1" text="Password" resource-id="com.example.app:id/input_pass" class="android.widget.EditText" package="com.example.app" content-desc="" checkable="false" checked="false" clickable="true" enabled="true" focusable="true" focused="false" scrollable="false" long-clickable="false" password="true" selected="false" bounds="[100,700][980,850]" />
  </node>
</hierarchy>`

func TestParseTreeState(t *testing.T) {
	ts, err := ParseTreeState(sampleXML)
	if err != nil {
		t.Fatalf("failed to parse tree state: %v", err)
	}

	if len(ts.InteractiveElements) != 2 {
		t.Fatalf("expected 2 interactive elements, got %d", len(ts.InteractiveElements))
	}

	btn := ts.InteractiveElements[0]
	if btn.Name != "Login" || btn.ResourceID != "btn_login" || btn.ClassName != "android.widget.Button" {
		t.Fatalf("unexpected button element: %+v", btn)
	}
	if btn.Coordinates.X != 540 || btn.Coordinates.Y != 575 {
		t.Fatalf("unexpected button coordinates: %+v", btn.Coordinates)
	}

	pass := ts.InteractiveElements[1]
	if pass.Name != "Password" || pass.ResourceID != "input_pass" {
		t.Fatalf("unexpected pass element: %+v", pass)
	}

	tableStr := ts.ToString()
	if !strings.Contains(tableStr, "Label") || !strings.Contains(tableStr, "btn_login") || !strings.Contains(tableStr, "(540,575)") {
		t.Fatalf("unexpected table representation: %s", tableStr)
	}
}

func TestFindElementBySelector(t *testing.T) {
	sel := Selector{ResourceID: "btn_login"}
	elem, err := FindElementBySelector(sampleXML, sel, "com.example.app")
	if err != nil || elem == nil {
		t.Fatalf("failed to find element by short resourceId: err=%v, elem=%+v", err, elem)
	}
	if elem.Name != "Login" {
		t.Fatalf("expected element name Login, got %q", elem.Name)
	}

	selText := Selector{Text: "Password"}
	elemText, err := FindElementBySelector(sampleXML, selText, "")
	if err != nil || elemText == nil {
		t.Fatalf("failed to find element by text: err=%v, elem=%+v", err, elemText)
	}
	if elemText.ResourceID != "input_pass" {
		t.Fatalf("expected input_pass, got %q", elemText.ResourceID)
	}
}

func TestAnnotateScreenshot(t *testing.T) {
	// Create dummy 200x200 PNG image
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for x := 0; x < 200; x++ {
		for y := 0; y < 200; y++ {
			img.Set(x, y, color.RGBA{100, 100, 100, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode dummy png: %v", err)
	}

	nodes := []ElementNode{
		{
			Name:        "TestNode",
			ClassName:   "android.widget.Button",
			Coordinates: CenterCord{X: 50, Y: 50},
			BoundingBox: BoundingBox{X1: 10, Y1: 10, X2: 90, Y2: 90},
			ResourceID:  "test_btn",
		},
	}

	annotated, err := AnnotateScreenshot(buf.Bytes(), nodes, 1.0)
	if err != nil {
		t.Fatalf("failed to annotate screenshot: %v", err)
	}
	if len(annotated) == 0 {
		t.Fatalf("annotated screenshot returned empty bytes")
	}

	// Verify annotated bytes decode back into valid PNG image
	resImg, err := png.Decode(bytes.NewReader(annotated))
	if err != nil {
		t.Fatalf("annotated bytes failed png decode: %v", err)
	}
	if resImg.Bounds().Dx() != 200 || resImg.Bounds().Dy() != 200 {
		t.Fatalf("annotated screenshot bounds changed: %+v", resImg.Bounds())
	}
}
