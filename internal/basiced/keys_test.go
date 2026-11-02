// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basiced

import (
	"testing"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/basic"
)

func keysEditor(t *testing.T, lines []basic.Line) *Editor {
	t.Helper()
	rom, err := assets.ROM("zx81.edition3")
	if err != nil {
		t.Fatal(err)
	}
	kb, err := KeyboardFromROM(rom)
	if err != nil {
		t.Fatal(err)
	}
	e := New(lines)
	e.SetKeysMode(true, kb)
	return e
}

// press types keys: "S+" before a key means with SHIFT.
func press(e *Editor, keys ...string) {
	for _, k := range keys {
		shift := len(k) > 2 && k[:2] == "S+"
		if shift {
			k = k[2:]
		}
		e.ZXKey(k, shift)
	}
}

func TestZX81Keys(t *testing.T) {
	e := keysEditor(t, nil)
	if e.Cursor() != 'K' {
		t.Fatalf("cursor at the start: %c", e.Cursor())
	}
	// 10 PRINT "OI";A — on the ZX81: 1 0 P SHIFT+P O I SHIFT+P SHIFT+X A
	press(e, "1", "0")
	if e.Cursor() != 'K' {
		t.Fatal("the line number keeps K")
	}
	press(e, "P")
	if e.Cursor() != 'L' {
		t.Fatal("after PRINT: L")
	}
	press(e, "S+P", "O", "I", "S+P", "S+X", "A")
	if e.Input() != `10 PRINT "OI";A` {
		t.Fatalf("input %q", e.Input())
	}
	press(e, "ENTER")
	if got := basic.List(e.Lines); got != "10 PRINT \"OI\";A\n" {
		t.Fatalf("program: %q (msg %q)", got, e.msg)
	}
	// 20 IF A THEN GOTO 10: K again after THEN; F for a function: LET B=PEEK 1
	press(e, "2", "0", "U", "A", "S+3")
	if e.Cursor() != 'K' {
		t.Fatalf("after THEN: %c", e.Cursor())
	}
	press(e, "G", "1", "0", "ENTER")
	press(e, "3", "0", "L", "B", "S+L", "S+ENTER")
	if e.Cursor() != 'F' {
		t.Fatal("FUNCTION: F")
	}
	press(e, "O", "1", "ENTER")
	// 40 PRINT "<inverse A><graphic>": G with SHIFT+9, inverse A, SHIFT+A grey graphic, SHIFT+9 leaves.
	press(e, "4", "0", "P", "S+P", "S+9")
	if e.Cursor() != 'G' {
		t.Fatal("GRAPHICS: G")
	}
	press(e, "A", "S+A", "S+9", "S+P", "ENTER")
	want := "10 PRINT \"OI\";A\n20 IF A THEN GOTO 10\n30 LET B=PEEK 1\n40 PRINT \"%A\\##\"\n"
	if got := basic.List(e.Lines); got != want {
		t.Fatalf("program:\n%s\nwant:\n%s(msg %q)", got, want, e.msg)
	}
	// RUBOUT (SHIFT+0) and EDIT (SHIFT+1) as on the ZX81.
	press(e, "5", "0", "Q", "S+0", "S+0")
	if e.Input() != "5" {
		t.Fatalf("rubout: %q", e.Input())
	}
	e.Key("Escape")
	e.Current = 30
	press(e, "S+1")
	if e.Input() != "30 LET B=PEEK 1" {
		t.Fatalf("edit: %q", e.Input())
	}
	// Back to direct typing with the line kept.
	e.SetKeysMode(false, nil)
	if e.Input() != "30 LET B=PEEK 1" || e.Cursor() != 'L' {
		t.Fatalf("direct: %q", e.Input())
	}
}

func TestSyntaxErrorS(t *testing.T) {
	// Keys mode: 70 LET =1 — the S goes on the "=", as on the ZX81.
	e := keysEditor(t, nil)
	press(e, "7", "0", "L", "S+L", "1", "ENTER")
	if len(e.Lines) != 0 || e.Cursor() != 'S' || e.cpos != 3 {
		t.Fatalf("lines %d cursor %c at %d", len(e.Lines), e.Cursor(), e.cpos)
	}
	if scr := e.Screen(); scr[Rows-1][4]&0x3F != 0x38 { // "70 LET " then the S
		t.Logf("bottom row %q", e.ScreenText()[Rows-1])
	}
	press(e, "A") // the S goes away with the next key: 70 LET A=1
	if e.Cursor() != 'L' || e.Input() != "70 LET A=1" {
		t.Fatalf("after fixing: %c %q", e.Cursor(), e.Input())
	}
	press(e, "ENTER")
	if len(e.Lines) != 1 {
		t.Fatalf("fixed line not accepted: %q", e.msg)
	}

	// Direct mode: the message says where.
	d := New(nil)
	typeLine(d, "10 print (")
	if msg, isErr := d.Message(); !isErr || len(d.Lines) != 0 || d.Cursor() != 'S' {
		t.Fatalf("direct: %q %v", msg, isErr)
	}
	d.TypeRune('1')
	if d.Cursor() != 'L' {
		t.Fatal("S stays after typing")
	}

	// RUN asks the window to run the program.
	typeLine(d, "")
	d.Key("Escape")
	typeLine(d, "run 100")
	if a := d.TakeAction(); a != "RUN 100" || d.TakeAction() != "" {
		t.Fatalf("action %q", a)
	}
}
