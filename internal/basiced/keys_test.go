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
