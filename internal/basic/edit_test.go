// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import "testing"

func mustParse(t *testing.T, text string) []Line {
	t.Helper()
	lines, err := ParseText(text)
	if err != nil {
		t.Fatal(err)
	}
	return lines
}

func TestDeleteLines(t *testing.T) {
	lines := mustParse(t, "10 PRINT 1\n20 PRINT 2\n30 PRINT 3\n40 GOTO 10\n")
	out, n := DeleteLines(lines, 20, 30)
	if n != 2 || List(out) != "10 PRINT 1\n40 GOTO 10\n" {
		t.Fatalf("%d %q", n, List(out))
	}
	if _, n := DeleteLines(lines, 50, 60); n != 0 {
		t.Fatalf("deleted %d", n)
	}
}

func TestFindLines(t *testing.T) {
	lines := mustParse(t, "10 PRINT \"OLA\"\n20 GOTO 10\n30 LET A=1\n")
	if got := List(FindLines(lines, "goto")); got != "20 GOTO 10\n" {
		t.Fatalf("%q", got)
	}
	if got := FindLines(lines, "xyz"); len(got) != 0 {
		t.Fatalf("%v", got)
	}
}

func TestPutLineAndRem(t *testing.T) {
	lines := mustParse(t, "10 PRINT 1\n30 PRINT 3\n")
	rem, err := RemLine(1, 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	lines = PutLine(lines, rem)
	l2 := mustParse(t, "20 CLS\n")[0]
	lines = PutLine(lines, l2)
	lines = PutLine(lines, mustParse(t, "30 PRINT 33\n")[0])
	if got := List(lines); got != `1 REM \  \  \  `+"\n10 PRINT 1\n20 CLS\n30 PRINT 33\n" {
		t.Fatalf("%q", got)
	}
	if a := LineAddress(lines, 1); a != ProgStart {
		t.Fatalf("line 1 at %04X", a)
	}
	// 1 REM + 3 bytes: 4 + 4 + 1 = 9 bytes
	if a := LineAddress(lines, 10); a != ProgStart+9 {
		t.Fatalf("line 10 at %04X", a)
	}
	if LineAddress(lines, 99) != -1 {
		t.Fatal("missing line")
	}
	if _, err := RemLine(1, 10, Newline); err == nil {
		t.Fatal("NEWLINE fill accepted")
	}
	if _, err := RemLine(10000, 10, 0); err == nil {
		t.Fatal("line 10000 accepted")
	}
}
