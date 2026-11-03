// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
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
