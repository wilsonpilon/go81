// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadText(t *testing.T, s *Session, text string) {
	t.Helper()
	f := filepath.Join(t.TempDir(), "prog.bas")
	if err := os.WriteFile(f, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	mustExec(t, s, `load "`+f+`"`)
}

func TestProgramTools(t *testing.T) {
	s := newBooted(t)
	loadText(t, s, "10 PRINT \"OLA\"\n20 GOTO 10\n30 PRINT 3\n40 GOTO 30\n")

	if out := mustExec(t, s, "find goto"); out != "20 GOTO 10\n40 GOTO 30\n" {
		t.Fatalf("find = %q", out)
	}
	if out := mustExec(t, s, "find NADA"); !strings.Contains(out, "nada") {
		t.Fatalf("find nothing = %q", out)
	}
	if out := mustExec(t, s, "delete 20,30"); !strings.Contains(out, "2 linha") {
		t.Fatalf("delete = %q", out)
	}
	if out := mustExec(t, s, "list"); out != "10 PRINT \"OLA\"\n40 GOTO 30\n" {
		t.Fatalf("list after delete = %q", out)
	}
	if out := mustExec(t, s, "delete 40"); !strings.Contains(out, "1 linha") {
		t.Fatalf("delete one = %q", out)
	}
	out := mustExec(t, s, "rem 1 4 $3E")
	if !strings.Contains(out, "16514 ($4082)") {
		t.Fatalf("rem = %q", out)
	}
	if out := mustExec(t, s, "peek 16514 4"); !strings.Contains(strings.ToUpper(out), "3E 3E 3E 3E") {
		t.Fatalf("peek = %q", out)
	}
	if out := mustExec(t, s, "list"); !strings.HasPrefix(out, "1 REM ") || !strings.HasSuffix(out, "10 PRINT \"OLA\"\n") {
		t.Fatalf("list after rem = %q", out)
	}

	var buf bytes.Buffer
	for _, bad := range []string{"delete", "delete 50,10", "delete x", "rem 1", "rem 1 10 118", "rem 1 10 300", "find"} {
		if err := s.Exec(bad, &buf); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
}
