// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package z81

import (
	"reflect"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	s := &Snapshot{Model: "TS1500", PC: 0x0207, SP: 0x7FFC, AF: 0x1C54, HL_: 0x1234, IY: 0x4000, IR: 0x1EFF,
		IM: 1, IFF1: 1, IFF2: 1, Halted: true, NMI: true, Line: 5, RAMPack: "16K", LowRAM: true,
		HasLow: true, Start: 0x2000, Memory: []byte{1, 2, 2, 2, 0, 0, 0x76, 0x76}, ROM: "ts1500.rom",
		M1Not: true, HasM1Not: true, FPFix: true, NTSC: true, HasNTSC: true}
	text := s.Text()
	got, err := Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, s) {
		t.Fatalf("round trip:\n%+v\n%+v\n%s", got, s, text)
	}
}

// An older snapshot (VB81): HSYNC and ROW, a [CONFIG] section EightyOne
// does not know, no [MACHINE].
func TestOldNames(t *testing.T) {
	s, err := Parse(`[CPU]
PC 04CC   SP  AFC6
IR 1EFF
IM 00     IF1 01
HT 00     IF2 01

[ZX81]
NMI 01    HSYNC 01
ROW 007

[CONFIG]
GCARD=0
RAM=4

[MEMORY]
MEMRANGE 4000 FFFF
FF C0 *0004 00 31
`)
	if err != nil {
		t.Fatal(err)
	}
	if s.Model != "" || s.PC != 0x04CC || !s.NMI || !s.Sync || s.Line != 7 || s.Start != 0x4000 ||
		!reflect.DeepEqual(s.Memory, []byte{0xFF, 0xC0, 0, 0, 0, 0, 0x31}) || s.HasNTSC {
		t.Fatalf("%+v", s)
	}
	if _, err := Parse("[CPU]\nPC 0000\n"); err == nil {
		t.Fatal("snapshot without memory accepted")
	}
	if _, err := Parse("[MEMORY]\nMEMRANGE 4000 7FFF\nZZ\n[CPU]\n"); err == nil {
		t.Fatal("bad byte accepted")
	}
	if PackK("None") != 0 || PackK("16K") != 16 || PackK("x") != -1 || PackName(0) != "None" || PackName(48) != "48K" {
		t.Fatal("RAM pack names")
	}
}
