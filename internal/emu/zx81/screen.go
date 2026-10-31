// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Palette ported from EightyOne Artifacts_.cpp (TArtifacts::StandardPalette)
// — Copyright (C) 2003-2025 Michael D Wynne (GPL v2 or later).

package zx81

import (
	"image"
	"image/color"
)

// Palette holds the 16 display colours with EightyOne's default TV settings
// (brightness 230, contrast 85, colour 0, "vibrant colour" on).
var Palette = func() [16]color.RGBA {
	const brightTrack, contrastTrack, colourTrack = 230, 85, 0
	bright := 255 - brightTrack
	contrast := 255 - contrastTrack
	colourLevel := 255 - colourTrack
	bright -= contrast
	hiBright := bright + contrast/2 + 2*contrast
	contrast = bright + contrast + contrast

	clamp := func(v int) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	var p [16]color.RGBA
	for i := 0; i < 16; i++ {
		c := i
		level := contrast
		if c > 7 {
			level = hiBright
		}
		difference := (1000 * (level - bright)) / 16
		base := (difference * ((c & 7) + 9)) / 1000
		if c == 0 || c == 8 {
			base = bright
		}
		comp := func(bit int) int {
			if c&bit != 0 {
				return level
			}
			return bright
		}
		colb, colg, colr := comp(1), comp(4), comp(2)
		bw := bright + base
		r := (colr-bw)*colourLevel/255 + bw
		g := (colg-bw)*colourLevel/255 + bw
		b := (colb-bw)*colourLevel/255 + bw
		p[i] = color.RGBA{clamp(r), clamp(g), clamp(b), 255}
	}
	return p
}()

// Border sizes, as in EightyOne's View > Border menu.
const (
	BorderNone   = iota // 256x192 (just the display file area)
	BorderSmall         // 264x200
	BorderNormal        // 320x240
	BorderLarge         // 400x300
	BorderFull          // 414x313
)

// window returns the visible part of the raster for a border size
// (AccDraw_.cpp window constants, ZX81 shifted 4 pixels right).
func (m *Machine) window(border int) image.Rectangle {
	var l, t, w, h int
	switch border {
	case BorderNone:
		l, t, w, h = 74, 56, 256, 192
	case BorderSmall:
		l, t, w, h = 70, 52, 264, 200
	case BorderLarge:
		l, t, w, h = 7, 6, 400, 300
	case BorderFull:
		l, t, w, h = 0, 0, 414, 313
	default:
		l, t, w, h = 42, 32, 320, 240
	}
	if border <= BorderNormal {
		l += 4
		if m.cfg.NTSC {
			t -= 24
		}
	}
	return image.Rect(l, t, l+w, t+h)
}

// Screen returns the last complete frame as an image, cropped to the given
// border size.
func (m *Machine) Screen(border int) *image.RGBA {
	r := m.window(border)
	img := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	f := m.Display()
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			img.SetRGBA(x, y, Palette[f[(r.Min.Y+y)*FrameWidth+r.Min.X+x]>>4&15])
		}
	}
	return img
}
