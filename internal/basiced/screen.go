// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basiced

import (
	"fmt"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// Character codes used on the screen.
const (
	codeSpace   = 0x00
	codeGreater = 0x12 // ">"
	inverse     = 0x80
)

// Screen draws the editor as 24 lines of 32 ZX81 character codes (bit 7 =
// inverse video), as the ZX81 shows its listing and its edit line.
func (e *Editor) Screen() [Rows][Cols]byte {
	var scr [Rows][Cols]byte
	input := e.inputRows()
	var msg [][]byte
	if e.msg != "" {
		codes := basic.TextCodes(e.msg)
		if e.msgErr {
			for i := range codes {
				codes[i] |= inverse
			}
		}
		msg = wrap(codes)
		if len(msg) > Rows/2 {
			msg = msg[:Rows/2]
		}
	}
	bottom := append(msg, input...)
	if len(bottom) > Rows {
		bottom = bottom[len(bottom)-Rows:]
	}
	avail := Rows - len(bottom)
	for i, row := range e.listing(avail) {
		copy(scr[i][:], row)
	}
	for i, row := range bottom {
		copy(scr[avail+i][:], row)
	}
	return scr
}

// inputRows is the line being typed with the cursor (an inverse L).
func (e *Editor) inputRows() [][]byte {
	var before, after []byte
	if e.keys { // keywords spelled out, as the ZX81 shows them
		before, after = basic.ScreenCodes(e.codes[:e.cpos]), basic.ScreenCodes(e.codes[e.cpos:])
	} else {
		before, after = basic.TextCodes(string(e.input[:e.pos])), basic.TextCodes(string(e.input[e.pos:]))
	}
	cursor := basic.TextCodes(string(rune(e.Cursor())))[0] | inverse
	codes := append(append(before, cursor), after...)
	return wrap(codes)
}

// lineRows is a program line as LIST shows it: the number in four
// columns, ">" (inverse) on the current line, and the body.
func (e *Editor) lineRows(l basic.Line) [][]byte {
	codes := basic.TextCodes(fmt.Sprintf("%4d", l.Num))
	if l.Num == e.Current {
		codes = append(codes, codeGreater|inverse)
	} else {
		codes = append(codes, codeSpace)
	}
	return wrap(append(codes, basic.ScreenCodes(l.Body)...))
}

// listing returns up to n rows of the listing, scrolled so that the
// current line is in view.
func (e *Editor) listing(n int) [][]byte {
	ls := e.listed()
	if len(ls) == 0 || n <= 0 {
		return nil
	}
	cur := e.index(e.Current)
	e.top = max(0, min(e.top, len(ls)-1))
	if cur < e.top {
		e.top = cur
	}
	for e.top < cur {
		rows := 0
		for _, l := range ls[e.top : cur+1] {
			rows += len(e.lineRows(l))
		}
		if rows <= n {
			break
		}
		e.top++
	}
	var out [][]byte
	for _, l := range ls[e.top:] {
		for _, r := range e.lineRows(l) {
			if len(out) == n {
				return out
			}
			out = append(out, r)
		}
	}
	return out
}

// wrap cuts codes in rows of Cols (at least one row).
func wrap(codes []byte) [][]byte {
	var rows [][]byte
	for len(codes) > Cols {
		rows = append(rows, codes[:Cols])
		codes = codes[Cols:]
	}
	row := make([]byte, Cols)
	copy(row, codes)
	for i := len(codes); i < Cols; i++ {
		row[i] = codeSpace
	}
	return append(rows, row)
}

// ScreenText is Screen as text (for tests and the terminal): inverse
// characters are not marked.
func (e *Editor) ScreenText() []string {
	scr := e.Screen()
	out := make([]string, Rows)
	for i, row := range scr {
		rs := make([]rune, Cols)
		for j, c := range row {
			rs[j] = textChar(c)
		}
		out[i] = string(rs)
	}
	return out
}

func textChar(c byte) rune { return zx81.TextChars[c&0x3F] }
