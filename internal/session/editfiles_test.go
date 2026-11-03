// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wilsonpilon/go81/internal/basic"
)

// The editor's files: the same lines go to .p, .p81 and text and back.
func TestProgramLinesFiles(t *testing.T) {
	dir := t.TempDir()
	lines, err := basic.ParseText("10 PRINT \"\\:: OI\"\n20 GOTO 10\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"jogo", "jogo.p81", "jogo.bas", "jogo.azx"} {
		written, err := WriteProgramLines(filepath.Join(dir, name), lines)
		if err != nil {
			t.Fatal(err)
		}
		back, err := ReadProgramLines(written)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if basic.List(back) != basic.List(lines) {
			t.Fatalf("%s: %q", name, basic.List(back))
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "jogo.p")); err != nil {
		t.Fatal("a name without extension gets .p")
	}
	os.WriteFile(filepath.Join(dir, "x.o"), []byte{0}, 0o644)
	if _, err := ReadProgramLines(filepath.Join(dir, "x.o")); err == nil {
		t.Fatal("a ZX80 program opened in the editor")
	}
}
