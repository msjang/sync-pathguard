// Package trayicon renders the tray/menu-bar icon as PNG bytes. The symbol is a
// horizontal ruler (placeholder for the FontAwesome ruler-horizontal asset,
// ADR-0005 / OBS-01) tinted by state color. Five states: gray (idle), blue
// (scanning), green (ok), yellow (warn), red (over).
//
// It is drawn as an outlined rectangle with a few graduation ticks so it still
// reads as a ruler after the OS scales it down to menu-bar size.
package trayicon

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

type State int

const (
	Idle State = iota
	Scanning
	OK
	Warn
	Over
)

var colors = map[State]color.RGBA{
	Idle:     {0x9a, 0xa0, 0xa6, 0xff}, // gray
	Scanning: {0x1a, 0x73, 0xe8, 0xff}, // blue
	OK:       {0x34, 0xa8, 0x53, 0xff}, // green
	Warn:     {0xf9, 0xab, 0x00, 0xff}, // amber
	Over:     {0xea, 0x43, 0x35, 0xff}, // red
}

// defaultSize suits a Retina menu bar (≈22pt @2x).
const defaultSize = 44

// PNG returns the icon for a state at the default size.
func PNG(s State) []byte { return PNGSize(s, defaultSize) }

// PNGSize returns the icon for a state rendered at size×size pixels.
func PNGSize(s State, size int) []byte {
	c, ok := colors[s]
	if !ok {
		c = colors[Idle]
	}
	img := render(c, size)
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func render(c color.RGBA, size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	w, h := float64(size), float64(size)

	stroke := iround(w * 0.09)
	if stroke < 1 {
		stroke = 1
	}
	// Ruler body: a wide, short outlined rectangle, vertically centered.
	x0, x1 := iround(0.08*w), iround(0.92*w)
	y0, y1 := iround(0.34*h), iround(0.66*h)

	fill(img, x0, y0, x1, y1, c)                                 // solid block…
	fill(img, x0+stroke, y0+stroke, x1-stroke, y1-stroke, clear) // …hollowed to an outline

	// Full-height graduation ticks — stay legible after the OS scales the icon
	// down to menu-bar size (~22px).
	for _, f := range []float64{0.25, 0.375, 0.5, 0.625, 0.75} {
		tx := iround(f * w)
		fill(img, tx, y0, tx+stroke, y1, c)
	}
	return img
}

var clear = color.RGBA{}

func fill(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func iround(f float64) int { return int(math.Round(f)) }
