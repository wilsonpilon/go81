// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basiced

import (
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/internal/basic"
)

// Files, the size of the program and scrolling the listing.

// Files reads and writes program files for LOAD, SAVE and MERGE (the
// window gives one; without it those commands say they are not available).
type Files interface {
	Read(name string) ([]basic.Line, error)
	Write(name string, lines []basic.Line) (string, error) // returns the name written
}

// fileName is the argument of LOAD/SAVE/MERGE: the rest of the line,
// without quotes ("" or nothing = ask the window for one).
func fileName(text, cmd string) string {
	return strings.Trim(strings.TrimSpace(text[len(cmd):]), `"`)
}

// fileCommand runs LOAD, SAVE or MERGE; ok is false when the line stays
// in the input (an error).
func (e *Editor) fileCommand(name, file string) bool {
	if file == "" { // the window asks for the file
		e.action = name
		e.msg = ""
		return true
	}
	if e.Files == nil {
		e.fail("%s NAO DISPONIVEL AQUI", name)
		return false
	}
	switch name {
	case "SAVE":
		written, err := e.Files.Write(file, e.Lines)
		if err != nil {
			e.fail("%s", strings.ToUpper(err.Error()))
			return false
		}
		e.Changed = false
		e.say("GRAVADO: %s (%d LINHAS, %d BYTES)", strings.ToUpper(filepath.Base(written)), len(e.Lines), e.Size())
	case "LOAD", "MERGE":
		lines, err := e.Files.Read(file)
		if err != nil {
			e.fail("%s", strings.ToUpper(err.Error()))
			return false
		}
		if name == "LOAD" {
			e.SetLines(lines)
			e.say("CARREGADO: %s (%d LINHAS, %d BYTES)", strings.ToUpper(filepath.Base(file)), len(lines), e.Size())
		} else {
			e.Merge(lines)
			e.say("JUNTADAS %d LINHAS DE %s", len(lines), strings.ToUpper(filepath.Base(file)))
		}
	}
	return true
}

// SetLines replaces the program (LOAD).
func (e *Editor) SetLines(lines []basic.Line) {
	e.Lines = append([]basic.Line(nil), lines...)
	e.Current, e.top, e.find, e.scroll = 0, 0, "", false
	if len(lines) > 0 {
		e.Current = lines[0].Num
	}
	e.Changed = true
}

// Merge puts the lines into the program; a line with the same number is
// replaced, as MERGE does on the ZX Spectrum.
func (e *Editor) Merge(lines []basic.Line) {
	for _, l := range lines {
		e.Lines = basic.PutLine(e.Lines, l)
	}
	if len(lines) > 0 {
		e.Current = lines[0].Num
		e.Changed = true
	}
	e.scroll = false
}

// Size is the number of bytes of the program in the ZX81's memory: each
// line takes its body plus five (number, length and the NEWLINE at the
// end).
func (e *Editor) Size() int {
	n := 0
	for _, l := range e.Lines {
		n += len(l.Body) + 5
	}
	return n
}

// Free is about how many bytes the ZX81 has left with RAMK of RAM: from
// the end of the program to RAMTOP, less the screen (full from 4K on,
// collapsed below that), the variables, the edit line and a little stack.
// It is -1 when the RAM is not known.
func (e *Editor) Free() int {
	if e.RAMK <= 0 {
		return -1
	}
	ramtop := 0x4000 + min(e.RAMK, 16)*1024
	dfile := 25 // collapsed: just the 25 NEWLINEs
	if e.RAMK >= 4 {
		dfile = 25 + 24*32
	}
	used := basic.ProgStart + e.Size() + dfile + 1 + 3 + 6
	return max(0, ramtop-used)
}

// ShowMem shows what MEM shows: the size of the program and the free memory.
func (e *Editor) ShowMem() {
	if free := e.Free(); free >= 0 {
		e.say("PROGRAMA: %d LINHAS, %d BYTES. LIVRE: CERCA DE %d BYTES", len(e.Lines), e.Size(), free)
		return
	}
	e.say("PROGRAMA: %d LINHAS, %d BYTES", len(e.Lines), e.Size())
}

// Scroll moves the listing by d lines without changing the current line
// (Ctrl+Up/Down, Ctrl+PgUp/PgDn). Choosing a line brings it back in view.
func (e *Editor) Scroll(d int) {
	n := len(e.listed())
	if n == 0 {
		return
	}
	e.top = max(0, min(n-1, e.top+d))
	e.scroll = true
}
