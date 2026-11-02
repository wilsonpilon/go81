// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basiced

import (
	"strconv"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/basic"
)

func typeLine(e *Editor, s string) {
	for _, r := range s {
		e.TypeRune(r)
	}
	e.Key("Return")
}

func screen(e *Editor) string { return strings.Join(e.ScreenText(), "\n") }

func TestTypingLines(t *testing.T) {
	e := New(nil)
	typeLine(e, `20 goto 10`)
	typeLine(e, `10 print "oi"`)
	if got := basic.List(e.Lines); got != "10 PRINT \"OI\"\n20 GOTO 10\n" || !e.Changed || e.Current != 10 {
		t.Fatalf("program:\n%s current %d", got, e.Current)
	}
	rows := e.ScreenText()
	if strings.TrimRight(rows[0], " ") != `  10>PRINT "OI"` || strings.TrimRight(rows[1], " ") != "  20 GOTO 10" {
		t.Fatalf("listing:\n%s", screen(e))
	}
	// The cursor (inverse L) is on the bottom line.
	scr := e.Screen()
	if scr[Rows-1][0] != codeL|inverse {
		t.Fatalf("cursor: % x", scr[Rows-1][:4])
	}
	// A number alone deletes the line.
	typeLine(e, "20")
	if len(e.Lines) != 1 {
		t.Fatal("20 did not delete")
	}
	// A bad line stays in the input, with the error in inverse video.
	typeLine(e, `30 print "a@"`)
	if msg, isErr := e.Message(); !isErr || e.Input() != `30 PRINT "A@"` {
		t.Fatalf("error: %q %v input %q", msg, isErr, e.Input())
	}
	if scr := e.Screen(); scr[Rows-2][0]&inverse == 0 {
		t.Fatal("error not in inverse")
	}
	e.Key("Escape")
	if e.Input() != "" {
		t.Fatal("Esc did not clear")
	}
}

func TestEditingKeys(t *testing.T) {
	lines, _ := basic.ParseText("10 PRINT 1\n20 PRINT 2\n30 PRINT 3\n")
	e := New(lines)
	e.Key("Down")
	e.Key("Down")
	e.Key("Up")
	if e.Current != 20 {
		t.Fatalf("current %d", e.Current)
	}
	e.Key("Tab") // EDIT
	if e.Input() != "20 PRINT 2" {
		t.Fatalf("edit: %q", e.Input())
	}
	e.Key("BackSpace")
	e.TypeRune('9')
	e.Key("Home")
	e.Key("Delete")
	e.TypeRune('4')
	e.Key("End")
	e.Key("Left")
	e.Key("Right")
	e.Key("Return")
	if got := basic.List(e.Lines); got != "10 PRINT 1\n20 PRINT 2\n30 PRINT 3\n40 PRINT 9\n" {
		t.Fatalf("after edit:\n%s", got)
	}
}

func TestCommands(t *testing.T) {
	lines, _ := basic.ParseText("1 PRINT 1\n2 GOTO 1\n")
	e := New(lines)
	e.Key("Down")
	typeLine(e, "renum 100 5")
	if got := basic.List(e.Lines); got != "100 PRINT 1\n105 GOTO 100\n" || e.Current != 105 {
		t.Fatalf("renum:\n%s current %d", got, e.Current)
	}
	typeLine(e, "auto")
	if e.Input() != "115 " {
		t.Fatalf("auto input %q", e.Input())
	}
	typeLine(e, "print 2")
	typeLine(e, "stop")
	if e.Input() != "135 " {
		t.Fatalf("auto next %q", e.Input())
	}
	e.input, e.pos = nil, 0
	e.Key("Return") // empty: ends auto
	if e.auto != nil {
		t.Fatal("auto still on")
	}
	typeLine(e, "find print")
	if len(e.listed()) != 2 || !strings.HasPrefix(e.ScreenText()[0], " 100>PRINT 1") {
		t.Fatalf("find:\n%s", screen(e))
	}
	e.Key("Escape")
	typeLine(e, "delete 115,125")
	if len(e.Lines) != 2 {
		t.Fatalf("delete: %s", basic.List(e.Lines))
	}
	typeLine(e, "rem 1 10")
	if msg, _ := e.Message(); !strings.Contains(msg, "16514") {
		t.Fatalf("rem: %q", msg)
	}
	typeLine(e, "edit 105")
	if e.Input() != "105 GOTO 100" {
		t.Fatalf("edit 105: %q", e.Input())
	}
	e.Key("Escape")
	typeLine(e, "xyz")
	if msg, isErr := e.Message(); !isErr || !strings.Contains(msg, "HELP") {
		t.Fatalf("unknown: %q", msg)
	}
	e.Key("Escape")
	typeLine(e, "help")
	if msg, _ := e.Message(); !strings.Contains(msg, "RENUM") {
		t.Fatal("help")
	}
	typeLine(e, "new")
	if len(e.Lines) != 0 {
		t.Fatal("new")
	}
}

func TestScrolling(t *testing.T) {
	var text strings.Builder
	for i := 1; i <= 40; i++ {
		text.WriteString(itoa(i*10) + " PRINT " + itoa(i) + "\n")
	}
	lines, _ := basic.ParseText(text.String())
	e := New(lines)
	for range 30 {
		e.Key("Down")
	}
	rows := e.ScreenText()
	found := false
	for _, r := range rows {
		if strings.HasPrefix(r, " 310>") {
			found = true
		}
	}
	if !found {
		t.Fatalf("current line not in view:\n%s", screen(e))
	}
	e.Key("Prior")
	e.Key("Prior")
	e.Key("Prior")
	e.Key("Prior")
	if !strings.HasPrefix(e.ScreenText()[0], "  10>") {
		t.Fatalf("back to the top:\n%s", screen(e))
	}
}

func itoa(n int) string { return strconv.Itoa(n) }
