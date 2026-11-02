// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

func TestSnapshot(t *testing.T) {
	s := newBooted(t)
	loadText(t, s, "5 LET A=0\n10 LET A=A+1\n20 PRINT AT 0,0;A\n30 GOTO 10\n")
	mustExec(t, s, "typeln R") // RUN
	mustExec(t, s, "run 100")
	file := filepath.Join(t.TempDir(), "conta.z81")
	mustExec(t, s, "snapshot save "+file)
	var mem []byte
	var pc uint16
	s.Do(func(m *zx81.Machine) { mem = append(mem, m.Memory[0x4000:0x8000]...); pc = m.CPU.PC })
	text, _ := os.ReadFile(file)
	if !strings.HasPrefix(string(text), "[MACHINE]\nMODEL ZX81\n") || !strings.Contains(string(text), "RAM_PACK 16K") {
		t.Fatalf("snapshot:\n%.300s", text)
	}

	// Another machine, then the snapshot brings everything back.
	mustExec(t, s, "machine ts1000")
	out := mustExec(t, s, "load "+file) // load takes .z81 too
	if !strings.Contains(out, "snapshot carregado") || s.Settings().RAMK != 16 || s.Settings().NTSC {
		t.Fatalf("load .z81 = %q %+v", out, s.Settings())
	}
	same := true
	s.Do(func(m *zx81.Machine) {
		same = string(m.Memory[0x4000:0x8000]) == string(mem) && m.CPU.PC == pc
	})
	if !same {
		t.Fatal("memory or PC not restored")
	}
	// The program goes on counting.
	before := screenNumber(t, s)
	mustExec(t, s, "run 100")
	if after := screenNumber(t, s); after <= before {
		t.Fatalf("the program did not go on: %d → %d", before, after)
	}
	if err := s.Exec("snapshot load "+filepath.Join(t.TempDir(), "nada.z81"), &strings.Builder{}); err == nil {
		t.Fatal("missing file accepted")
	}
}

// screenNumber is the number on the first line of the screen.
func screenNumber(t *testing.T, s *Session) int {
	t.Helper()
	var line string
	s.Do(func(m *zx81.Machine) { line = strings.TrimSpace(m.ScreenText()[0]) })
	n, err := ParseNumber(line)
	if err != nil {
		t.Fatalf("screen: %q", line)
	}
	return n
}
