// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package termimg draws a grid of terminal characters as a PNG picture, for
// the screenshots of the documentation (terminal interface and shell). It
// uses the Go Mono font; block graphics (the ZX81 quadrants) are drawn as
// rectangles, so they are exact whatever the font has.
package termimg

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Cell is one character with its colours.
type Cell struct {
	R      rune
	FG, BG color.RGBA
}

// Default colours of a dark terminal.
var (
	DefaultFG = color.RGBA{0xD4, 0xD4, 0xD4, 0xFF}
	DefaultBG = color.RGBA{0x1E, 0x1E, 0x1E, 0xFF}
	titleBG   = color.RGBA{0x33, 0x33, 0x33, 0xFF}
)

// FromText turns lines of text into cells with the default colours.
func FromText(lines []string) [][]Cell {
	w := 0
	for _, l := range lines {
		w = max(w, len([]rune(l)))
	}
	g := make([][]Cell, len(lines))
	for y, l := range lines {
		g[y] = make([]Cell, w)
		rs := []rune(l)
		for x := range g[y] {
			c := Cell{R: ' ', FG: DefaultFG, BG: DefaultBG}
			if x < len(rs) {
				c.R = rs[x]
			}
			g[y][x] = c
		}
	}
	return g
}

// quadrants: bit 0 top-left, 1 top-right, 2 bottom-left, 3 bottom-right.
var quadrants = map[rune]int{
	'▘': 1, '▝': 2, '▀': 3, '▖': 4, '▌': 5, '▞': 6, '▛': 7,
	'▗': 8, '▚': 9, '▐': 10, '▜': 11, '▄': 12, '▙': 13, '▟': 14, '█': 15,
}

// Render draws the cells under a window title bar.
func Render(cells [][]Cell, title string) (*image.RGBA, error) {
	f, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: 15, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil, err
	}
	defer face.Close()
	adv, _ := face.GlyphAdvance('M')
	cw := adv.Ceil()
	m := face.Metrics()
	ch := (m.Ascent + m.Descent).Ceil()
	ascent := m.Ascent.Ceil()

	cols := 0
	for _, row := range cells {
		cols = max(cols, len(row))
	}
	const pad, bar = 12, 30
	img := image.NewRGBA(image.Rect(0, 0, cols*cw+2*pad, len(cells)*ch+2*pad+bar))
	draw.Draw(img, img.Bounds(), &image.Uniform{DefaultBG}, image.Point{}, draw.Src)
	// Title bar with the three window buttons.
	draw.Draw(img, image.Rect(0, 0, img.Bounds().Dx(), bar), &image.Uniform{titleBG}, image.Point{}, draw.Src)
	for i, c := range []color.RGBA{{0xFF, 0x5F, 0x56, 0xFF}, {0xFF, 0xBD, 0x2E, 0xFF}, {0x27, 0xC9, 0x3F, 0xFF}} {
		disc(img, 16+i*20, bar/2, 6, c)
	}
	d := &font.Drawer{Dst: img, Src: &image.Uniform{DefaultFG}, Face: face}
	tw := d.MeasureString(title).Ceil()
	d.Dot = fixed.P((img.Bounds().Dx()-tw)/2, bar/2+5)
	d.DrawString(title)

	for y, row := range cells {
		for x, c := range row {
			px, py := pad+x*cw, bar+pad+y*ch
			r := image.Rect(px, py, px+cw, py+ch)
			draw.Draw(img, r, &image.Uniform{c.BG}, image.Point{}, draw.Src)
			if q, ok := quadrants[c.R]; ok {
				hw, hh := cw/2, ch/2
				parts := []image.Rectangle{
					image.Rect(px, py, px+hw, py+hh), image.Rect(px+hw, py, px+cw, py+hh),
					image.Rect(px, py+hh, px+hw, py+ch), image.Rect(px+hw, py+hh, px+cw, py+ch),
				}
				for i, p := range parts {
					if q&(1<<i) != 0 {
						draw.Draw(img, p, &image.Uniform{c.FG}, image.Point{}, draw.Src)
					}
				}
				continue
			}
			if c.R == '▒' { // the ZX81 grey: a fine chequer
				for yy := py; yy < py+ch; yy++ {
					for xx := px; xx < px+cw; xx++ {
						if (xx+yy)%2 == 0 {
							img.SetRGBA(xx, yy, c.FG)
						}
					}
				}
				continue
			}
			if c.R == ' ' || c.R == 0 {
				continue
			}
			d.Src = &image.Uniform{c.FG}
			d.Dot = fixed.P(px, py+ascent)
			d.DrawString(string(c.R))
		}
	}
	return img, nil
}

func disc(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				img.SetRGBA(cx+x, cy+y, c)
			}
		}
	}
}

// WritePNG renders the cells and writes a PNG file.
func WritePNG(path string, cells [][]Cell, title string) error {
	img, err := Render(cells, title)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Encode renders the cells as PNG to w.
func Encode(w io.Writer, cells [][]Cell, title string) error {
	img, err := Render(cells, title)
	if err != nil {
		return err
	}
	return png.Encode(w, img)
}
