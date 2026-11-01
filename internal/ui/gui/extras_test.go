// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package gui

import (
	"fmt"
	"testing"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

func TestIsProgram(t *testing.T) {
	for _, p := range []string{"jogo.p", "JOGO.P", "x.81", "a.P81", "lista.bas", "texto.TXT", "b.b81"} {
		if !isProgram(p) {
			t.Errorf("%s is a program", p)
		}
	}
	for _, p := range []string{"tela.png", "rom.rom", "semext", "x.p.zip"} {
		if isProgram(p) {
			t.Errorf("%s is not a program", p)
		}
	}
}

func TestAddRecent(t *testing.T) {
	var l []string
	for i := range maxRecent + 3 {
		l = addRecent(l, fmt.Sprintf("p%d.p", i))
	}
	if len(l) != maxRecent || l[0] != fmt.Sprintf("p%d.p", maxRecent+2) {
		t.Fatalf("list %v", l)
	}
	l = addRecent(l, l[3]) // opened again: goes to the top, not repeated
	if len(l) != maxRecent || l[0] != fmt.Sprintf("p%d.p", maxRecent-1) {
		t.Fatalf("list %v", l)
	}
	seen := map[string]bool{}
	for _, p := range l {
		if seen[p] {
			t.Fatalf("repeated %s in %v", p, l)
		}
		seen[p] = true
	}
}

func TestKeyboardLayout(t *testing.T) {
	n := 0
	for _, row := range zx81Keyboard {
		if len(row) != 10 {
			t.Fatalf("row with %d keys", len(row))
		}
		for _, k := range row {
			if _, err := zx81.ParseChord(k.Key); err != nil {
				t.Errorf("key %q: %v", k.Key, err)
			}
			n++
		}
	}
	if n != 40 {
		t.Fatalf("%d keys, the ZX81 has 40", n)
	}
}
