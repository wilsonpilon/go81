// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tape

import (
	"bytes"
	"testing"
)

func TestZX81Name(t *testing.T) {
	cases := []struct {
		in   string
		want []byte
	}{
		{"OLA", []byte{0x34, 0x31, 0x26 | 128}},
		{"jogo_1", []byte{0x2F, 0x34, 0x2C, 0x34, 0x00, 0x1D | 128}},
		// '-' is code 1 in EightyOne's table (first occurrence), not 22.
		{"A-B", []byte{0x26, 0x01, 0x27 | 128}},
		// Conversion stops at the first character with no ZX81 code.
		{"AB!CD", []byte{0x26, 0x27 | 128}},
	}
	for _, c := range cases {
		if got := ZX81Name(c.in); !bytes.Equal(got, c.want) {
			t.Errorf("ZX81Name(%q) = % x, want % x", c.in, got, c.want)
		}
	}
}

func TestProgramBlock(t *testing.T) {
	data := []byte{0, 1, 2}
	b, err := ProgramBlock(`C:\jogos\OLA.p`, data)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{0x34, 0x31, 0x26 | 128, 0, 1, 2}; !bytes.Equal(b, want) {
		t.Fatalf(".p block = % x, want % x", b, want)
	}
	b, err = ProgramBlock("x/OLA.p81", data)
	if err != nil || !bytes.Equal(b, data) {
		t.Fatalf(".p81 block = % x, %v", b, err)
	}
	if _, err := ProgramBlock("OLA.tzx", data); err == nil {
		t.Fatal("expected error for .tzx")
	}
}

func TestFlashLoadDeliversBytes(t *testing.T) {
	d := NewDeck(3250000)
	d.Insert([]byte{10, 20, 30})
	d.WavStart()
	if !d.Player.Playing {
		t.Fatal("tape did not start")
	}
	var got []int
	// The player first waits a quarter of a second (clock/4 T-states).
	for i := 0; i < 1000000 && len(got) < 3; i++ {
		d.ClockTick(4)
		if d.Player.Byte != ByteEmpty {
			got = append(got, d.Player.Byte)
			d.Player.Byte = ByteEmpty // the ROM patch takes the byte
		}
	}
	if len(got) != 3 || got[0] != 10 || got[1] != 20 || got[2] != 30 {
		t.Fatalf("flash load delivered %v", got)
	}
}
