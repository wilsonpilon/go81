// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basiced

import (
	"fmt"

	"github.com/wilsonpilon/go81/internal/basic"
)

// The "ZX81 keys" mode: the line is typed as on the ZX81 itself. With the
// K cursor (start of the line, after the number, after THEN) a letter key
// gives a whole command (P = PRINT); with L it gives the letter; SHIFT
// gives the sign or word on the key; SHIFT+ENTER (FUNCTION) gives the F
// cursor, where a key gives a function; SHIFT+9 (GRAPHICS) gives the G
// cursor, for inverse characters and, with SHIFT, the block graphics.
// What each key gives is read from the keyboard tables of the ROM.

// KeyOrder is the order of the keys in the ROM tables.
var KeyOrder = []string{"Z", "X", "C", "V", "A", "S", "D", "F", "G", "Q", "W", "E", "R", "T",
	"1", "2", "3", "4", "5", "0", "9", "8", "7", "6", "P", "O", "I", "U", "Y", "ENTER",
	"L", "K", "J", "H", "SPACE", ".", "M", "N", "B"}

// Keyboard is what each key gives, from the ROM tables.
type Keyboard struct {
	Unshift, Shift, Funct, Graph [39]byte
}

// Tables of the ZX81 ROM (edition 3; the TS1000 and the clones keep them).
const (
	romUnshift = 0x007E // K-UNSHIFT
	romShift   = 0x00A5 // K-SHIFT
	romFunct   = 0x00CC // K-FUNCT
	romGraph   = 0x00EF // graphics on SHIFT with the G cursor (K-GRAPH)
	noKey      = 0x78   // the key gives nothing (only leaves F)
)

// KeyboardFromROM reads the tables from a ZX81 ROM.
func KeyboardFromROM(rom []byte) (*Keyboard, error) {
	if len(rom) < romGraph+39 {
		return nil, fmt.Errorf("ROM pequena demais: %d bytes", len(rom))
	}
	k := &Keyboard{}
	copy(k.Unshift[:], rom[romUnshift:])
	copy(k.Shift[:], rom[romShift:])
	copy(k.Funct[:], rom[romFunct:])
	copy(k.Graph[:], rom[romGraph:])
	return k, nil
}

func keyIndex(name string) int {
	for i, k := range KeyOrder {
		if k == name {
			return i
		}
	}
	return -1
}

// Editing codes in the tables.
const (
	codeUp       = 0x70
	codeDown     = 0x71
	codeLeft     = 0x72
	codeRight    = 0x73
	codeGraphics = 0x74
	codeEdit     = 0x75
	codeNewline  = 0x76
	codeRubout   = 0x77
	codeFunction = 0x79
	tokTHEN      = 0xDE
)

// Cursor is the cursor of the line being typed: K, L, F or G (in the
// direct mode, L).
func (e *Editor) Cursor() byte {
	if e.serr {
		return 'S'
	}
	if !e.keys {
		return 'L'
	}
	switch {
	case e.fmode:
		return 'F'
	case e.gmode:
		return 'G'
	}
	// K: nothing but the line number before the cursor, or THEN just before.
	before := e.codes[:e.cpos]
	i := 0
	for i < len(before) && (before[i] >= 0x1C && before[i] <= 0x25 || before[i] == 0) {
		i++
	}
	rest := before[i:]
	for len(rest) > 0 && rest[len(rest)-1] == 0 {
		rest = rest[:len(rest)-1]
	}
	if len(rest) == 0 || rest[len(rest)-1] == tokTHEN {
		return 'K'
	}
	return 'L'
}

// KeysMode reports whether the ZX81 keys mode is on.
func (e *Editor) KeysMode() bool { return e.keys }

// SetKeysMode turns the ZX81 keys mode on (kb are the ROM tables) or off.
// The line being typed is kept.
func (e *Editor) SetKeysMode(on bool, kb *Keyboard) {
	if on == e.keys || on && kb == nil {
		return
	}
	text := e.lineText()
	e.keys, e.kb, e.fmode, e.gmode = on, kb, false, false
	if on {
		e.codes = basic.TextCodes(text)
		if l, err := basic.ParseLine(text); err == nil && text != "" && text[0] >= '0' && text[0] <= '9' {
			e.codes = basic.TypedCodes(l)
		}
		e.cpos = len(e.codes)
		e.say("TECLAS DO ZX81: K COMANDO, L LETRA, SHIFT+ENTER FUNCAO, SHIFT+9 GRAFICOS")
	} else {
		e.setInput(text)
		e.say("DIGITACAO DIRETA")
	}
}

// lineText is the line being typed, as text.
func (e *Editor) lineText() string {
	if e.keys {
		return basic.CodesText(e.codes)
	}
	return string(e.input)
}

// ZXKey is a key of the ZX81 keyboard (a name in KeyOrder), with SHIFT or
// not, in the ZX81 keys mode.
func (e *Editor) ZXKey(name string, shift bool) {
	i := keyIndex(name)
	if !e.keys || i < 0 {
		return
	}
	e.serr = false
	kb := e.kb
	var c byte
	switch e.Cursor() {
	case 'F': // a function; SHIFT makes no difference
		e.fmode = false
		if c = kb.Funct[i]; c == noKey {
			return
		}
	case 'G':
		switch u, s := kb.Unshift[i], kb.Shift[i]; {
		case !shift && u == codeNewline:
			c = u
		case !shift:
			c = u | 0x80 // inverse
		case s < 0x40:
			c = s | 0x80
		case s == codeGraphics || s == codeRubout || s >= codeUp && s <= codeRight:
			c = s // GRAPHICS leaves G; RUBOUT and the arrows work
		case kb.Graph[i] != noKey:
			c = kb.Graph[i]
		default:
			return
		}
	case 'K':
		if shift {
			c = kb.Shift[i]
		} else if c = kb.Unshift[i]; c >= 0x26 && c < 0x40 {
			c += 0xC0 // a letter: the command
		}
	default: // L
		if shift {
			c = kb.Shift[i]
		} else {
			c = kb.Unshift[i]
		}
	}
	switch c {
	case codeUp:
		e.Key("Up")
	case codeDown:
		e.Key("Down")
	case codeLeft:
		e.Key("Left")
	case codeRight:
		e.Key("Right")
	case codeGraphics:
		e.gmode = !e.gmode
	case codeEdit:
		e.Key("Tab")
	case codeNewline:
		e.gmode = false
		e.Key("Return")
	case codeRubout:
		e.Key("BackSpace")
	case codeFunction:
		e.fmode = true
	default:
		e.codes = append(e.codes[:e.cpos], append([]byte{c}, e.codes[e.cpos:]...)...)
		e.cpos++
	}
}
