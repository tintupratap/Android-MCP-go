package bench

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/tintupratap/Android-MCP-go/internal/adb"
	"github.com/tintupratap/Android-MCP-go/internal/ui"
)

const sampleXML = `<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy rotation="0">
  <node index="0" text="" resource-id="" class="android.widget.FrameLayout" package="com.example.app" content-desc="" checkable="false" checked="false" clickable="false" enabled="true" focusable="false" focused="false" scrollable="false" long-clickable="false" password="false" selected="false" bounds="[0,0][1080,2400]">
    <node index="0" text="Login" resource-id="com.example.app:id/btn_login" class="android.widget.Button" package="com.example.app" content-desc="" checkable="false" checked="false" clickable="true" enabled="true" focusable="true" focused="false" scrollable="false" long-clickable="false" password="false" selected="false" bounds="[100,500][980,650]" />
    <node index="1" text="Password" resource-id="com.example.app:id/input_pass" class="android.widget.EditText" package="com.example.app" content-desc="" checkable="false" checked="false" clickable="true" enabled="true" focusable="true" focused="false" scrollable="false" long-clickable="false" password="true" selected="false" bounds="[100,700][980,850]" />
  </node>
</hierarchy>`

func BenchmarkParseTreeState(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ui.ParseTreeState(sampleXML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFindElementBySelector(b *testing.B) {
	b.ReportAllocs()
	sel := ui.Selector{ResourceID: "btn_login"}
	for i := 0; i < b.N; i++ {
		_, err := ui.FindElementBySelector(sampleXML, sel, "com.example.app")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormatTable(b *testing.B) {
	b.ReportAllocs()
	headers := []string{"Label", "Name", "ResourceId", "Class", "Coordinates"}
	rows := [][]string{
		{"0", "Login", "btn_login", "android.widget.Button", "(540,575)"},
		{"1", "Password", "input_pass", "android.widget.EditText", "(540,775)"},
	}
	for i := 0; i < b.N; i++ {
		_ = ui.FormatTable(headers, rows)
	}
}

func BenchmarkParseADBDevices(b *testing.B) {
	sampleOutput := `List of devices attached
QV771A3JEE	device product:SOG09 model:SOG09 device:SOG09 transport_id:1
192.168.1.3:5555	device product:SOG09 model:SOG09 device:SOG09 transport_id:2
emulator-5554	device product:sdk_gphone64_arm64 model:sdk_gphone64_arm64 device:emulator_arm64 transport_id:3
`
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = adb.ParseADBDevices(sampleOutput)
	}
}

func BenchmarkAnnotateScreenshot(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 500, 1000))
	for x := 0; x < 500; x++ {
		for y := 0; y < 1000; y++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	pngBytes := buf.Bytes()

	nodes := []ui.ElementNode{
		{
			Name:        "Login",
			ClassName:   "android.widget.Button",
			Coordinates: ui.CenterCord{X: 250, Y: 300},
			BoundingBox: ui.BoundingBox{X1: 50, Y1: 250, X2: 450, Y2: 350},
			ResourceID:  "btn_login",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ui.AnnotateScreenshot(pngBytes, nodes, 1.0)
		if err != nil {
			b.Fatal(err)
		}
	}
}
