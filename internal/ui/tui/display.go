// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// Display modes (F9 cycles through them).
const (
	modeWide    = iota // text, each ZX81 character two cells wide: about 4:3
	modePixels         // the TV picture scaled to the space, 4:3
	modeCompact        // text, one cell per character (narrow terminals)
	modeCount
)

var modeNames = [modeCount]string{"texto 4:3", "pixels 4:3", "texto compacto"}

// wideCols is the width of the wide text screen with its border.
const wideCols = 2*Cols + 4

// effectiveMode is the mode actually drawn: wide text needs wideCols columns.
func effectiveMode(mode, w int) int {
	if mode == modeWide && w < wideCols {
		return modeCompact
	}
	return mode
}

// draw draws the screen in a mode.
func draw(scr tcell.Screen, m *zx81.Machine, mode, x, y, w, h int) {
	switch effectiveMode(mode, w) {
	case modePixels:
		drawPixelsFit(scr, m, x, y, w, h)
	case modeCompact:
		drawText(scr, m.TextScreen(), x, y, w, h)
	default:
		drawWide(scr, m.TextScreen(), x, y, w, h)
	}
}

// Wide characters: letters, digits and symbols use the Unicode full-width
// forms (the terminal draws them two cells wide); block graphics become two
// cells of half blocks, so the picture keeps its exact shape.

// fullWidth returns the full-width form of an ASCII character (or of £).
func fullWidth(r rune) rune {
	switch {
	case r == '£':
		return '￡'
	case r > ' ' && r <= '~':
		return r + 0xFEE0
	}
	return 0
}

// halves are the half-block characters: bit 0 = top, bit 1 = bottom.
var halves = []rune(" ▀▄█")

// quadBits maps the ZX81 quadrant characters to their bits (0 top-left,
// 1 top-right, 2 bottom-left, 3 bottom-right).
var quadBits = func() map[rune]int {
	m := map[rune]int{}
	for i, r := range quadrants {
		m[r] = i
	}
	return m
}()

// wideCell returns the two cells for a ZX81 character.
func wideCell(c Cell) (left, right rune, wide bool) {
	if c.Grey {
		return c.R, c.R, false // ▒ ▄ ▀ in grey: the same on both cells
	}
	if q, ok := quadBits[c.R]; ok {
		l := q&1 | q>>2&1<<1    // top-left, bottom-left
		r := q>>1&1 | q>>3&1<<1 // top-right, bottom-right
		return halves[l], halves[r], false
	}
	if f := fullWidth(c.R); f != 0 {
		return f, 0, true
	}
	return ' ', ' ', false
}

// drawWide draws the text screen two cells per character, with a border,
// centred: 68x26 cells, close to the 4:3 of a TV.
func drawWide(scr tcell.Screen, s [Rows][Cols]Cell, x, y, w, h int) {
	const bw, bh = wideCols, Rows + 2
	ox := x + max((w-bw)/2, 0)
	oy := y + max((h-bh)/2, 0)
	for j := 0; j < bh && j < h; j++ {
		for i := 0; i < bw && i < w; i++ {
			scr.SetContent(ox+i, oy+j, ' ', nil, paperStyle)
		}
	}
	for r := 0; r < Rows; r++ {
		py := oy + 1 + r
		if py >= y+h {
			break
		}
		for c := 0; c < Cols; c++ {
			cell := s[r][c]
			st := paperStyle
			if cell.Inverse {
				st = inkStyle
			}
			if cell.Grey {
				if cell.Inverse {
					st = greyInk
				} else {
					st = greyPaper
				}
			}
			px := ox + 2 + 2*c
			if px+1 >= x+w {
				break
			}
			left, right, wide := wideCell(cell)
			if wide {
				scr.SetContent(px, py, left, nil, st)
				continue
			}
			scr.SetContent(px, py, left, nil, st)
			scr.SetContent(px+1, py, right, nil, st)
		}
	}
}

// drawPixelsFit draws the TV picture (display area, 256x192) as large as the
// space allows at 4:3, with quadrant characters (2x2 dots per cell). Cells
// are about twice as tall as wide, so 4:3 is 8:3 in cells.
func drawPixelsFit(scr tcell.Screen, m *zx81.Machine, x, y, w, h int) {
	img := m.Screen(zx81.BorderNone)
	b := img.Bounds()
	maxW, maxH := w-4, h-2 // room for a border
	cw := min(maxW, maxH*8/3)
	ch := cw * 3 / 8
	// The whole resolution (2x2 dots per cell) fits: use it exactly. The
	// cells are a little wider than half their height, so this is still
	// close to 4:3.
	if maxW >= b.Dx()/2 && maxH >= b.Dy()/2 {
		cw, ch = b.Dx()/2, b.Dy()/2
	}
	shrink := 2*cw < b.Dx() // fewer dots than pixels
	if cw < 2 || ch < 1 {
		return
	}
	ox := x + (w-cw-4)/2
	oy := y + (h-ch-2)/2
	for j := 0; j < ch+2; j++ {
		for i := 0; i < cw+4; i++ {
			scr.SetContent(ox+i, oy+j, ' ', nil, paperStyle)
		}
	}
	isInk := func(xx, yy int) bool {
		c := img.RGBAAt(b.Min.X+xx, b.Min.Y+yy)
		return int(c.R)+int(c.G)+int(c.B) < 3*128
	}
	// blockInk: is the background of the 8x8 character cell ink (inverse)?
	inv := map[[2]int]bool{}
	blockInk := func(bx, by int) bool {
		k := [2]int{bx, by}
		if v, ok := inv[k]; ok {
			return v
		}
		n := 0
		for yy := by * 8; yy < by*8+8 && yy < b.Dy(); yy++ {
			for xx := bx * 8; xx < bx*8+8 && xx < b.Dx(); xx++ {
				if isInk(xx, yy) {
					n++
				}
			}
		}
		inv[k] = n > 32
		return n > 32
	}
	// dark reports whether a dot is ink.
	dots := func(n, size int) func(int) (int, int) {
		return func(i int) (int, int) { return i * size / n, max((i+1)*size/n, i*size/n+1) }
	}
	dx, dy := dots(2*cw, b.Dx()), dots(2*ch, b.Dy())
	dark := func(i, j int) bool {
		x0, x1 := dx(i)
		y0, y1 := dy(j)
		ink, all := 0, 0
		for yy := y0; yy < y1; yy++ {
			for xx := x0; xx < x1; xx++ {
				if isInk(xx, yy) {
					ink++
				}
				all++
			}
		}
		if shrink {
			// Keep thin strokes: the dot shows the foreground if any pixel
			// under it differs from the background of its ZX81 character
			// (8x8 pixels), which works for normal and inverse text.
			bg := blockInk((x0+x1)/2/8, (y0+y1)/2/8)
			if bg {
				return ink == all
			}
			return ink > 0
		}
		return 2*ink >= all
	}
	for j := 0; j < ch; j++ {
		for i := 0; i < cw; i++ {
			q := 0
			if dark(2*i, 2*j) {
				q |= 1
			}
			if dark(2*i+1, 2*j) {
				q |= 2
			}
			if dark(2*i, 2*j+1) {
				q |= 4
			}
			if dark(2*i+1, 2*j+1) {
				q |= 8
			}
			scr.SetContent(ox+2+i, oy+1+j, quadrants[q], nil, paperStyle)
		}
	}
}
