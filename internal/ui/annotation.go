package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func getRandomColor() color.RGBA {
	return color.RGBA{
		R: uint8(rng.Intn(200) + 55),
		G: uint8(rng.Intn(200) + 55),
		B: uint8(rng.Intn(200) + 55),
		A: 255,
	}
}

// AnnotateScreenshot renders colored bounding box rectangles and index badges on top of PNG image bytes
func AnnotateScreenshot(pngBytes []byte, nodes []ElementNode, scale float64) ([]byte, error) {
	if scale <= 0 {
		scale = 1.0
	}

	srcImg, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot png: %w", err)
	}

	bounds := srcImg.Bounds()
	dstImg := image.NewRGBA(bounds)
	draw.Draw(dstImg, bounds, srcImg, bounds.Min, draw.Src)

	for i, node := range nodes {
		c := getRandomColor()
		box := node.BoundingBox

		x1 := int(float64(box.X1) * scale)
		y1 := int(float64(box.Y1) * scale)
		x2 := int(float64(box.X2) * scale)
		y2 := int(float64(box.Y2) * scale)

		// Clamp box coordinates to image dimensions
		if x1 < bounds.Min.X {
			x1 = bounds.Min.X
		}
		if y1 < bounds.Min.Y {
			y1 = bounds.Min.Y
		}
		if x2 > bounds.Max.X {
			x2 = bounds.Max.X
		}
		if y2 > bounds.Max.Y {
			y2 = bounds.Max.Y
		}

		// Draw rectangle outline (thickness = 2)
		drawRectOutline(dstImg, x1, y1, x2, y2, 2, c)

		// Draw index badge label
		labelStr := strconv.Itoa(i)
		drawBadgeLabel(dstImg, labelStr, x1, y1, c)
	}

	var outBuf bytes.Buffer
	if err := png.Encode(&outBuf, dstImg); err != nil {
		return nil, fmt.Errorf("failed to encode annotated png: %w", err)
	}

	return outBuf.Bytes(), nil
}

func drawRectOutline(dst *image.RGBA, x1, y1, x2, y2, thickness int, c color.RGBA) {
	for t := 0; t < thickness; t++ {
		// Top and bottom horizontal lines
		for x := x1; x < x2; x++ {
			if y1+t < dst.Bounds().Max.Y {
				dst.Set(x, y1+t, c)
			}
			if y2-1-t >= dst.Bounds().Min.Y {
				dst.Set(x, y2-1-t, c)
			}
		}
		// Left and right vertical lines
		for y := y1; y < y2; y++ {
			if x1+t < dst.Bounds().Max.X {
				dst.Set(x1+t, y, c)
			}
			if x2-1-t >= dst.Bounds().Min.X {
				dst.Set(x2-1-t, y, c)
			}
		}
	}
}

func drawBadgeLabel(dst *image.RGBA, label string, x, y int, bg color.RGBA) {
	face := basicfont.Face7x13
	advance := font.MeasureString(face, label)
	textWidth := advance.Ceil()
	textHeight := 13

	badgeW := textWidth + 6
	badgeH := textHeight + 4

	bx1 := x
	by1 := y - badgeH
	if by1 < 0 {
		by1 = y
	}
	bx2 := bx1 + badgeW
	by2 := by1 + badgeH

	// Draw filled background rectangle for badge
	for px := bx1; px < bx2 && px < dst.Bounds().Max.X; px++ {
		for py := by1; py < by2 && py < dst.Bounds().Max.Y; py++ {
			if px >= 0 && py >= 0 {
				dst.Set(px, py, bg)
			}
		}
	}

	// Draw white text inside badge
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.White),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(bx1 + 3),
			Y: fixed.I(by1 + 12),
		},
	}
	d.DrawString(label)
}
