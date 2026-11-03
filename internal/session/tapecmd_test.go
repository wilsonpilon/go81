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

// tape new empties the tape, tape insert puts a program on it without
// loading it, and LOAD "" in the ZX81 takes it from there.
func TestTapeCommand(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "fita.txt")
	if err := os.WriteFile(in, []byte("10 PRINT \"FITA\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newBooted(t)
	mustExec(t, s, "tape new")
	if out := mustExec(t, s, "tape"); !strings.Contains(out, "1 bloco(s)") {
		t.Fatalf("tape = %q", out)
	}
	mustExec(t, s, `txt2p "`+in+`"`)
	if out := mustExec(t, s, `tape insert "`+filepath.Join(dir, "fita.p")+`"`); !strings.Contains(out, `LOAD ""`) {
		t.Fatalf("tape insert = %q", out)
	}
	if out := mustExec(t, s, "tape"); !strings.Contains(out, "programa FITA") {
		t.Fatalf("tape = %q", out)
	}
	mustExec(t, s, `typeln J""`) // LOAD "" (J is LOAD in K mode)
	mustExec(t, s, "run 400")
	if out := mustExec(t, s, "list"); !strings.Contains(out, `10 PRINT "FITA"`) {
		t.Fatalf("after LOAD: %q", out)
	}
	if err := s.Exec("tape voar", nil); err == nil {
		t.Fatal("tape voar accepted")
	}
}

// What the ZX81 SAVEs goes on the tape; tape save writes it as .tzx and
// .t81; a tape of two programs loads the second by its name; load and
// p2txt read .tzx and .t81.
func TestTapeFiles(t *testing.T) {
	dir := t.TempDir()
	s := newBooted(t)
	in := filepath.Join(dir, "dois.txt")
	os.WriteFile(in, []byte("10 PRINT \"DOIS\"\n"), 0o644)
	mustExec(t, s, `load "`+in+`"`)
	// A new tape, two SAVEs.
	mustExec(t, s, "tape new")
	mustExec(t, s, `save "`+filepath.Join(dir, "a.p")+`" DOIS`)
	mustExec(t, s, `save "`+filepath.Join(dir, "b.p")+`" OUTRO`)
	if out := mustExec(t, s, "tape"); !strings.Contains(out, "programa DOIS") || !strings.Contains(out, "programa OUTRO") || !strings.Contains(out, "3 bloco(s)") {
		t.Fatalf("tape after SAVE:\n%s", out)
	}
	tzx, t81 := filepath.Join(dir, "fita.tzx"), filepath.Join(dir, "fita.t81")
	mustExec(t, s, `tape save "`+tzx+`"`)
	mustExec(t, s, `tape save "`+t81+`"`)
	mustExec(t, s, `tape save "`+filepath.Join(dir, "primeiro.p")+`"`)
	first, _ := os.ReadFile(filepath.Join(dir, "primeiro.p"))
	saved, _ := os.ReadFile(filepath.Join(dir, "a.p"))
	if len(first) == 0 || !bytes.Equal(first, saved) {
		t.Fatal("tape save .p: not the first program")
	}
	// Both files list the first program, and the editor reads it.
	for _, f := range []string{tzx, t81} {
		text, _, err := ProgramText(f)
		if err != nil || text != "10 PRINT \"DOIS\"\n" {
			t.Fatalf("%s: %q %v", f, text, err)
		}
		if lines, err := ReadProgramLines(f); err != nil || len(lines) != 1 {
			t.Fatalf("editor %s: %v", f, err)
		}
	}
	// A new ZX81, the tape in: LOAD "OUTRO" skips DOIS.
	mustExec(t, s, "reset")
	mustExec(t, s, "run 150")
	if out := mustExec(t, s, `tape insert "`+tzx+`"`); !strings.Contains(out, "3 blocos") {
		t.Fatalf("tape insert: %q", out)
	}
	mustExec(t, s, `typeln J"OUTRO"`)
	mustExec(t, s, "run 900")
	if out := mustExec(t, s, "list"); !strings.Contains(out, `10 PRINT "DOIS"`) {
		t.Fatalf("after LOAD \"OUTRO\": %q", out)
	}
	if out := mustExec(t, s, "tape"); !strings.Contains(out, "> 3") && !strings.Contains(out, "fita parada") {
		t.Fatalf("tape after LOAD:\n%s", out)
	}
	// load of a .t81 loads its first program by itself.
	mustExec(t, s, "reset")
	mustExec(t, s, `load "`+t81+`"`)
	if out := mustExec(t, s, "list"); !strings.Contains(out, `10 PRINT "DOIS"`) {
		t.Fatalf("load .t81: %q", out)
	}
}
