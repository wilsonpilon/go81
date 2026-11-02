// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"image"
	"image/color"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// The cursor map (CTRL held with the keyboard map on): what each key gives
// with the cursor the ZX81 shows now (K, L, F or G). The cursor is read from
// the system variables as OUT-CURS (079Dh) does; what each key gives comes
// from the keyboard tables of the ROM, as MODE-SORT (04DFh) uses them; the
// picture is drawn with the ZX81's own characters.

// romKeyOrder is the order of the keys in the ROM tables.
var romKeyOrder = []string{"Z", "X", "C", "V", "A", "S", "D", "F", "G", "Q", "W", "E", "R", "T",
	"1", "2", "3", "4", "5", "0", "9", "8", "7", "6", "P", "O", "I", "U", "Y", "ENTER",
	"L", "K", "J", "H", "SPACE", ".", "M", "N", "B"}

// mapRows is the keyboard as it looks.
var mapRows = [4][10]string{
	{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"},
	{"Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P"},
	{"A", "S", "D", "F", "G", "H", "J", "K", "L", "ENTER"},
	{"SHIFT", "Z", "X", "C", "V", "B", "N", "M", ".", "SPACE"},
}

// Keyboard tables of the ZX81 ROM.
const (
	romUnshift = 0x007E // K-UNSHIFT
	romShift   = 0x00A5 // K-SHIFT
	romFunct   = 0x00CC // K-FUNCT
	romGraph   = 0x00EF // the graphics on SHIFT in G mode (K-GRAPH)
	romFont    = 0x1E00 // the 64 characters
	noKey      = 0x78   // no character: the key only changes the mode
)

// cursorMode is the cursor the ZX81 shows: MODE (4006h) says F or G,
// otherwise bit 2 of FLAGS (4001h) says L or K.
func cursorMode(m *zx81.Machine) byte {
	if mode := m.Peek(0x4006); mode != 0 {
		if mode&4 != 0 {
			return 'G'
		}
		return 'F'
	}
	if m.Peek(0x4001)&4 != 0 {
		return 'L'
	}
	return 'K'
}

// chars81 are the ZX81 codes 0-63 as text (1-10 are graphics).
var chars81 = []rune(" \x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\"£$:?()><=+-*/;,.0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// text81 turns upper-case text into ZX81 codes.
func text81(s string) []byte {
	var out []byte
	for _, r := range s {
		for c, x := range chars81 {
			if x == r {
				out = append(out, byte(c))
				break
			}
		}
	}
	return out
}

var keyCodeNames = map[byte]string{0x70: "UP", 0x71: "DOWN", 0x72: "LEFT", 0x73: "RIGHT",
	0x74: "GRAPH", 0x75: "EDIT", 0x76: "NEWLINE", 0x77: "RUBOUT", 0x79: "FUNC"}

// codeLabel is how a key code is drawn: a character (inverse ones too) or
// the name of a keyword or of an editing key.
func codeLabel(c byte) []byte {
	switch {
	case c < 0x40 || c >= 0x80 && c < 0xC0:
		return []byte{c}
	case basic.Tokens[c] != "":
		return text81(basic.Tokens[c])
	}
	return text81(keyCodeNames[c])
}

// cursorKeys says, for each key, what it gives with the cursor alone and
// with SHIFT.
func cursorKeys(rom []byte, cursor byte) map[string][2][]byte {
	keys := map[string][2][]byte{}
	for i, k := range romKeyOrder {
		u, s := rom[romUnshift+i], rom[romShift+i]
		var main, shift []byte
		switch cursor {
		case 'K': // letters are commands; the rest as in L
			main, shift = codeLabel(u), codeLabel(s)
			if u >= 0x26 && u < 0x40 {
				main = codeLabel(u + 0xC0)
			}
		case 'F': // functions; SHIFT makes no difference
			if f := rom[romFunct+i]; f != noKey {
				main = codeLabel(f)
			}
		case 'G': // inverse characters; SHIFT gives the graphics
			main = codeLabel(u | 0x80)
			if u == 0x76 {
				main = codeLabel(u)
			}
			switch g := rom[romGraph+i]; {
			case s < 0x40:
				shift = codeLabel(s | 0x80)
			case g != noKey:
				shift = codeLabel(g)
			case k == "9": // GRAPHICS again leaves G
				shift = codeLabel(0x74)
			}
		default: // L
			main, shift = codeLabel(u), codeLabel(s)
		}
		keys[k] = [2][]byte{main, shift}
	}
	return keys
}

var cursorTitles = map[byte]string{
	'K': " COMANDOS (SHIFT: SINAIS E PALAVRAS)",
	'L': " LETRAS E NUMEROS (SHIFT: SINAIS E PALAVRAS)",
	'F': " FUNCOES (DEPOIS VOLTA AO CURSOR L)",
	'G': " INVERSOS (SHIFT: GRAFICOS)",
}

// The cursor map picture, in pixels of the ZX81 characters.
const (
	cellW, cellH = 60, 32
	cellGap      = 2
	titleH       = 12
	cursorMapW   = cellGap + 10*(cellW+cellGap)
	cursorMapH   = titleH + cellGap + 4*(cellH+cellGap)
)

var (
	inkBack  = color.RGBA{0, 0, 0, 255}
	inkKey   = color.RGBA{48, 48, 48, 255}
	inkName  = color.RGBA{150, 150, 150, 255}
	inkMain  = color.RGBA{255, 255, 255, 255}
	inkShift = color.RGBA{255, 110, 110, 255} // red, as on the ZX81 keys
)

// drawCodes draws ZX81 codes with the ROM characters; x is the left edge.
func drawCodes(img *image.RGBA, font []byte, x, y int, codes []byte, ink color.RGBA) {
	for n, c := range codes {
		glyph := font[int(c&0x3F)*8:]
		inv := c&0x80 != 0
		for row := range 8 {
			for col := range 8 {
				if (glyph[row]&(0x80>>col) != 0) != inv {
					img.SetRGBA(x+n*8+col, y+row, ink)
				}
			}
		}
	}
}

// centered draws codes centered in a cell.
func centered(img *image.RGBA, font []byte, x, y int, codes []byte, ink color.RGBA) {
	if len(codes) > cellW/8 {
		codes = codes[:cellW/8]
	}
	drawCodes(img, font, x+(cellW-8*len(codes))/2, y, codes, ink)
}

// cursorMapPicture draws the map for a cursor (K, L, F or G).
func cursorMapPicture(rom []byte, cursor byte) *image.RGBA {
	font := rom[romFont : romFont+512]
	img := image.NewRGBA(image.Rect(0, 0, cursorMapW, cursorMapH))
	fill := func(r image.Rectangle, c color.RGBA) {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				img.SetRGBA(x, y, c)
			}
		}
	}
	fill(img.Bounds(), inkBack)
	// The title: the cursor as the ZX81 shows it (inverse), and what it does.
	title := append([]byte{text81(string(cursor))[0] | 0x80}, text81(cursorTitles[cursor])...)
	drawCodes(img, font, cellGap+2, 2, title, inkMain)
	keys := cursorKeys(rom, cursor)
	for r, row := range mapRows {
		y := titleH + cellGap + r*(cellH+cellGap)
		for c, k := range row {
			x := cellGap + c*(cellW+cellGap)
			fill(image.Rect(x, y, x+cellW, y+cellH), inkKey)
			centered(img, font, x, y+2, text81(k), inkName)
			centered(img, font, x, y+12, keys[k][0], inkMain)
			centered(img, font, x, y+22, keys[k][1], inkShift)
		}
	}
	return img
}

var cursorMapCache = map[byte]image.Image{}

// cursorMapFor is the (cached) picture for a cursor, nil if there is none.
func cursorMapFor(cursor byte) image.Image {
	if img, ok := cursorMapCache[cursor]; ok {
		return img
	}
	if cursorTitles[cursor] == "" {
		return nil
	}
	rom, err := assets.ROM("zx81.edition3")
	if err != nil {
		return nil
	}
	img := cursorMapPicture(rom, cursor)
	cursorMapCache[cursor] = img
	return img
}
