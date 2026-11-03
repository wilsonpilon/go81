// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basiced

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/basic"
)

// memFiles keeps "files" in memory.
type memFiles map[string][]basic.Line

func (m memFiles) Read(name string) ([]basic.Line, error) {
	l, ok := m[name]
	if !ok {
		return nil, fmt.Errorf("não há %s", name)
	}
	return l, nil
}

func (m memFiles) Write(name string, lines []basic.Line) (string, error) {
	m[name] = append([]basic.Line(nil), lines...)
	return name, nil
}

func TestLoadSaveMerge(t *testing.T) {
	files := memFiles{}
	e := New(nil)
	typeLine(e, `save "a.p"`)
	if msg, isErr := e.Message(); !isErr || !strings.Contains(msg, "NAO DISPONIVEL") {
		t.Fatalf("without files: %q", msg)
	}
	e.Key("Escape")
	e.Files = files
	typeLine(e, `10 print "um"`)
	typeLine(e, `20 goto 10`)
	typeLine(e, `save "a.p"`)
	if len(files["A.P"]) != 2 || e.Changed {
		t.Fatalf("save: %v %v %q", files, e.Changed, e.msg)
	}
	typeLine(e, "new")
	typeLine(e, `20 print "dois"`)
	typeLine(e, `30 stop`)
	typeLine(e, `merge "a.p"`)
	if got := basic.List(e.Lines); got != "10 PRINT \"UM\"\n20 GOTO 10\n30 STOP\n" {
		t.Fatalf("merge:\n%s", got)
	}
	typeLine(e, "new")
	typeLine(e, `load "a.p"`)
	if got := basic.List(e.Lines); got != "10 PRINT \"UM\"\n20 GOTO 10\n" || e.Current != 10 {
		t.Fatalf("load:\n%s", got)
	}
	typeLine(e, `load "nada.p"`)
	if _, isErr := e.Message(); !isErr || e.Input() != `LOAD "NADA.P"` {
		t.Fatalf("missing file: %q", e.Input())
	}
	// Without a name, the window is asked for the file.
	e.Key("Escape")
	typeLine(e, "load")
	if a := e.TakeAction(); a != "LOAD" {
		t.Fatalf("action %q", a)
	}
}

func TestMemAndScroll(t *testing.T) {
	e := New(nil)
	for i := 1; i <= 40; i++ {
		typeLine(e, fmt.Sprintf("%d rem linha", i*10))
	}
	// REM LINHA: REM + 5 characters + NEWLINE = 7 bytes, plus 4 (number, length).
	if e.Size() != 40*11 {
		t.Fatalf("size %d", e.Size())
	}
	e.RAMK = 16
	typeLine(e, "mem")
	if msg, _ := e.Message(); !strings.Contains(msg, "440 BYTES") || !strings.Contains(msg, "LIVRE") {
		t.Fatalf("mem: %q", msg)
	}
	if free := e.Free(); free != 0x8000-basic.ProgStart-440-793-10 {
		t.Fatalf("free %d", free)
	}
	e.RAMK = 1
	if free := e.Free(); free != 0x4400-basic.ProgStart-440-25-10 {
		t.Fatalf("free 1K %d", free)
	}
	// The current line is the last one typed (400), at the bottom. Scrolling
	// shows the top of the program without changing it.
	e.msg = ""
	e.Screen()
	e.Scroll(-100)
	rows := e.ScreenText()
	if !strings.HasPrefix(rows[0], "  10 REM") || e.Current != 400 {
		t.Fatalf("scrolled:\n%s current %d", screen(e), e.Current)
	}
	e.Scroll(1)
	if rows := e.ScreenText(); !strings.HasPrefix(rows[0], "  20 REM") {
		t.Fatalf("scroll 1:\n%s", screen(e))
	}
	// Choosing a line brings it back in view.
	e.Key("Up")
	if rows := e.ScreenText(); !strings.Contains(strings.Join(rows, "\n"), " 390>REM") {
		t.Fatalf("after Up:\n%s", screen(e))
	}
}
