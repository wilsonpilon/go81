// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"testing"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

func TestKeyMapChordsAreValid(t *testing.T) {
	for pc, chord := range keyMap {
		if _, err := zx81.ParseChord(chord); err != nil {
			t.Errorf("key %q -> %q: %v", pc, chord, err)
		}
	}
}

func TestKeyCombinesHeldKeys(t *testing.T) {
	rom, _ := assets.ROM("zx81.edition3")
	m, err := zx81.New(zx81.DefaultConfig(rom))
	if err != nil {
		t.Fatal(err)
	}
	e := &emulator{m: m, pressed: map[string]bool{}}
	e.key("LeftShift", true)
	e.key("P", true)
	if m.Keyboard[0] != 1 || m.Keyboard[5] != 1 { // SHIFT row 0 bit 0, P row 5 bit 0
		t.Fatalf("Shift+P: keyboard %v", m.Keyboard)
	}
	e.key("BackSpace", true) // SHIFT+0
	if m.Keyboard[4] != 1 {
		t.Fatalf("Backspace: keyboard %v", m.Keyboard)
	}
	e.key("LeftShift", false)
	if m.Keyboard[0] != 1 { // still held through Backspace
		t.Fatalf("SHIFT released too early: %v", m.Keyboard)
	}
	e.releaseAll()
	if m.Keyboard != [8]byte{} {
		t.Fatalf("releaseAll left keys down: %v", m.Keyboard)
	}
	e.key("F12", true) // not a ZX81 key
	if m.Keyboard != [8]byte{} {
		t.Fatalf("unmapped key changed the keyboard: %v", m.Keyboard)
	}
}
