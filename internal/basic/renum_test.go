// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"strings"
	"testing"
)

func TestRenumber(t *testing.T) {
	lines, err := ParseText(`5 REM GOTO 7
7 PRINT "GOTO 5"
9 IF A=1 THEN GOTO 7
12 GOSUB 15
13 GOTO A*2
15 RETURN
20 LIST 6
22 RUN
`)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Renumber(lines, 100, 10, 0, MaxLine)
	if err != nil {
		t.Fatal(err)
	}
	got := List(res.Lines)
	want := `100 REM GOTO 7
110 PRINT "GOTO 5"
120 IF A=1 THEN GOTO 110
130 GOSUB 150
140 GOTO A*2
150 RETURN
160 LIST 110
170 RUN
`
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if res.Moved != 8 || res.Refs != 3 || len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "linha 140 (antes 13): GOTO") {
		t.Fatalf("moved %d refs %d warnings %q", res.Moved, res.Refs, res.Warnings)
	}
	// The number after GOTO is a real number again: the program round-trips.
	p, err := MakeP(res.Lines, POptions{})
	if err != nil {
		t.Fatal(err)
	}
	if l, _ := ListP(p); l != want {
		t.Fatalf("after MakeP:\n%s", l)
	}
	again, _ := Lines(mustProg(t, p))
	if b := again[2].Body; b[len(b)-6] != NumMarker || FloatValue([5]byte(b[len(b)-5:])) != 110 {
		t.Fatalf("number after GOTO not rewritten: % x", b)
	}
}

func mustProg(t *testing.T, p []byte) []byte {
	t.Helper()
	b, err := ProgramFromP(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestRenumberBlock(t *testing.T) {
	lines, _ := ParseText("10 GOTO 30\n20 PRINT 1\n30 PRINT 2\n40 GOTO 20\n100 STOP\n")
	res, err := Renumber(lines, 50, 5, 20, 40)
	if err != nil {
		t.Fatal(err)
	}
	if got := List(res.Lines); got != "10 GOTO 55\n50 PRINT 1\n55 PRINT 2\n60 GOTO 50\n100 STOP\n" {
		t.Fatalf("block:\n%s", got)
	}
	if _, err := Renumber(lines, 5, 10, 20, 40); err == nil {
		t.Fatal("block over line 10 accepted")
	}
	if _, err := Renumber(lines, 9000, 500, 0, MaxLine); err == nil {
		t.Fatal("number over 9999 accepted")
	}
	// GOTO to a line that does not exist goes to the next one.
	lines, _ = ParseText("10 GOTO 15\n20 STOP\n")
	res, _ = Renumber(lines, 1, 1, 0, MaxLine)
	if got := List(res.Lines); got != "1 GOTO 2\n2 STOP\n" {
		t.Fatalf("missing target:\n%s", got)
	}
}
