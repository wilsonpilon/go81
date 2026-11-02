// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import "testing"

func TestCheckSyntax(t *testing.T) {
	for _, c := range []struct {
		line string
		pos  int // -1 = accepted; else where the S goes (in TypedCodes)
	}{
		{`10 PRINT "OI"`, -1},
		{`20 GOTO 10`, -1},
		{`30 IF A=1 THEN PRINT A;B,C`, -1},
		{`40 LET A$="X"+CHR$ 65`, -1},
		{`50 FOR I=1 TO 10 STEP 2`, -1},
		{`60 PRINT (`, -2},
		{`70 LET =1`, -2},
		{`80 GOTO`, -2},
		{`90 PRINT 1+`, -2},
		{`100 NEXT 5`, -2},
		{`110 REM QUALQUER COISA (`, -1},
	} {
		l, err := ParseLine(c.line)
		if err != nil {
			t.Fatalf("%s: %v", c.line, err)
		}
		pos, err := CheckLine(l)
		switch {
		case c.pos == -1 && err != nil:
			t.Errorf("%s: rejected at %d", c.line, pos)
		case c.pos == -2 && err == nil:
			t.Errorf("%s: accepted", c.line)
		case err != nil && (pos < 0 || pos > len(TypedCodes(l))):
			t.Errorf("%s: position %d", c.line, pos)
		}
	}
	// The S of "70 LET =1" is on the "=" (index 3: 7, 0, LET, =).
	l, _ := ParseLine("70 LET =1")
	if pos, _ := CheckLine(l); pos != 3 {
		t.Errorf("70 LET =1: S at %d, want 3", pos)
	}
}
