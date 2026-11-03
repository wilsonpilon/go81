// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package zx81ref

import (
	"testing"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

// The ZX81's own ROM loads the sound Go81 writes (tape.SaveWAV), in real
// time and without flash load, through the WAV loader of the deck: LOAD ""
// reads the bits from the EAR input as from a cassette. The program in
// memory is then the one saved.
func TestROMLoadsWAV(t *testing.T) {
	rom, err := assets.ROM("zx81.edition3")
	if err != nil {
		t.Fatal(err)
	}
	p, err := MakeP(rom, progOla...)
	if err != nil {
		t.Fatal(err)
	}
	player := tape.NewPlayer(3250000)
	player.AddGeneral(append(tape.ZX81Name("OLA"), p...))
	wav := tape.SaveWAV(player.Blocks)
	wp, err := tape.NewWavPlayer(wav)
	if err != nil {
		t.Fatal(err)
	}

	m, err := zx81.New(zx81.DefaultConfig(rom))
	if err != nil {
		t.Fatal(err)
	}
	m.Tape = tape.NewDeck(3250000)
	m.Tape.InsertWav(wp)
	keys, end := Type(150, 6, 6, 60, "J", "SHIFT+P", "SHIFT+P", "ENTER") // LOAD ""
	ev := 0
	// The program lines of the .p: from 407Dh to D_FILE (the screen changes
	// after LOAD: the ROM prints its report).
	dfile := int(p[0x400C-0x4009]) | int(p[0x400D-0x4009])<<8
	prog := p[0x407D-0x4009 : dfile-0x4009]
	loaded := func() bool {
		for i, b := range prog {
			if m.Peek(uint16(0x407D+i)) != b {
				return false
			}
		}
		return true
	}
	for f := 0; f < end+3000; f++ {
		for ev < len(keys) && keys[ev].Frame <= f {
			m.Keyboard[keys[ev].Row&7] = byte(keys[ev].Bits)
			ev++
		}
		m.RunFrame()
		if f > end && loaded() && !wp.Playing() {
			at, length := wp.Position()
			t.Logf("loaded by the ROM after %d frames, %.1f s of %.1f s of sound", f, at, length)
			return
		}
	}
	var got []byte
	for i := range prog {
		got = append(got, m.Peek(uint16(0x407D+i)))
	}
	at, length := wp.Position()
	t.Fatalf("the ROM did not load the WAV (%.1f of %.1f s played)\nwant % x\ngot  % x", at, length, prog, got)
}
