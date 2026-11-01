// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package zx81

import (
	"strings"
	"unicode"
)

// TextChars maps ZX81 character codes 0-63 to Unicode. Codes 1-10 are the
// block graphics, which are the Unicode quadrant characters (plus three grey
// patterns).
var TextChars = []rune(" ▘▝▀▖▌▞▛▒▄▀\"£$:?()><=+-*/;,.0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// ZX80TextChars is the ZX80 character set: the quote is code 1, the block
// graphics are 2-11 (9-11 the grey ones) and the signs are in another
// order (EightyOne's zx80BasicLister.cpp).
var ZX80TextChars = []rune(" \"▌▄▘▝▖▗▞▒▄▀£$:?()-+*/=><;,.0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// textChars is the character set of the machine.
func (m *Machine) textChars() ([]rune, int, int) {
	if m.cfg.ZX80 {
		return ZX80TextChars, 9, 11
	}
	return TextChars, 8, 10
}

// Text screen size.
const (
	TextRows = 24
	TextCols = 32
)

// TextCell is one character of the text screen.
type TextCell struct {
	R       rune
	Inverse bool
	Grey    bool // codes 8-10: chequered patterns
}

// TextScreen decodes the display file (D_FILE) into 24 rows of 32
// characters. It works for the normal text display (collapsed or full), not
// for hi-res pictures.
func (m *Machine) TextScreen() [TextRows][TextCols]TextCell {
	var s [TextRows][TextCols]TextCell
	for r := range s {
		for c := range s[r] {
			s[r][c] = TextCell{R: ' '}
		}
	}
	chars, grey0, grey1 := m.textChars()
	a := uint16(m.Peek(0x400C)) | uint16(m.Peek(0x400D))<<8
	if m.Peek(a) != 0x76 {
		return s
	}
	a++
	for r := 0; r < TextRows; r++ {
		for c := 0; ; c++ {
			b := m.Peek(a)
			a++
			if b == 0x76 || c > TextCols {
				break
			}
			if c == TextCols {
				continue // malformed line: skip until NEWLINE
			}
			code := int(b & 63)
			cell := TextCell{R: '?', Inverse: b&128 != 0}
			if b&64 == 0 {
				cell.R = chars[code]
				cell.Grey = code >= grey0 && code <= grey1
			}
			s[r][c] = cell
		}
	}
	return s
}

// ScreenText returns the text screen as 24 lines (inverse video is not
// marked; trailing spaces are kept so columns line up).
func (m *Machine) ScreenText() []string {
	s := m.TextScreen()
	lines := make([]string, TextRows)
	for r, row := range s {
		var sb strings.Builder
		for _, c := range row {
			sb.WriteRune(c.R)
		}
		lines[r] = sb.String()
	}
	return lines
}

// symbolKeys maps typed symbols to the ZX81 key that produces them.
var symbolKeys = map[rune]string{
	' ': "SPACE", '.': ".", '\n': "ENTER",
	'"': "SHIFT+P", '$': "SHIFT+U", '(': "SHIFT+I", ')': "SHIFT+O",
	'-': "SHIFT+J", '+': "SHIFT+K", '=': "SHIFT+L",
	':': "SHIFT+Z", ';': "SHIFT+X", '?': "SHIFT+C", '/': "SHIFT+V",
	'*': "SHIFT+B", '<': "SHIFT+N", '>': "SHIFT+M", ',': "SHIFT+.",
	'£': "SHIFT+SPACE",
}

// zx80SymbolKeys are the ZX80 positions of the symbols (the shifted half of
// the keyboard table in the ZX80 ROM, at 006Ch).
var zx80SymbolKeys = map[rune]string{
	' ': "SPACE", '.': ".", '\n': "ENTER",
	'"': "SHIFT+Y", '$': "SHIFT+U", '(': "SHIFT+I", ')': "SHIFT+O", '*': "SHIFT+P",
	'-': "SHIFT+J", '+': "SHIFT+K", '=': "SHIFT+L",
	':': "SHIFT+Z", ';': "SHIFT+X", '?': "SHIFT+C", '/': "SHIFT+V",
	'<': "SHIFT+N", '>': "SHIFT+M", ',': "SHIFT+.", '£': "SHIFT+SPACE",
}

// ChordFor returns the key chord that types r on this machine ("" if none):
// the ZX80 has its symbols on other keys.
func (m *Machine) ChordFor(r rune) string {
	if !m.cfg.ZX80 {
		return ChordForRune(r)
	}
	u := unicode.ToUpper(r)
	if u >= 'A' && u <= 'Z' || u >= '0' && u <= '9' {
		return string(u)
	}
	return zx80SymbolKeys[r]
}

// IsZX80 reports whether the machine is a ZX80.
func (m *Machine) IsZX80() bool { return m.cfg.ZX80 }

// ChordForRune returns the ZX81 key chord that types r ("" if none).
// Letters (either case) and digits are their own key; symbols use SHIFT.
func ChordForRune(r rune) string {
	u := unicode.ToUpper(r)
	if u >= 'A' && u <= 'Z' || u >= '0' && u <= '9' {
		return string(u)
	}
	return symbolKeys[r]
}
