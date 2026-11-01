// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/basic"
)

// TestModelsBoot turns each emulated model on, types PRINT 6*7 and looks
// for the answer: the ROM, the memory map and the keyboard work. The TK85
// and the R-470 have other symbols on the keys (EightyOne maps the PC keys
// to the ZX81 positions, so "*" is another sign there): they print digits.
func TestModelsBoot(t *testing.T) {
	for _, m := range Models {
		if !m.Ready {
			continue
		}
		t.Run(m.ID, func(t *testing.T) {
			set, err := DefaultSettings().WithModel(m.ID)
			if err != nil {
				t.Fatal(err)
			}
			s, err := New(set)
			if err != nil {
				t.Fatal(err)
			}
			mustExec(t, s, "run 500") // the TK85 takes longer to start
			line, want := "P6*7", "42"
			switch m.ID {
			case "tk85", "r470":
				line, want = "P67", "67"
			case "zx80": // PRINT is the O key; * is typed where the ZX80 has it
				line = "O6*7"
			}
			mustExec(t, s, "typeln "+line)
			mustExec(t, s, "run 120")
			scr := mustExec(t, s, "screen")
			if !strings.Contains(scr, want) {
				t.Fatalf("%s: no 42 on the screen:\n%s", m.Name, scr)
			}
		})
	}
}

func TestModelSettings(t *testing.T) {
	set := DefaultSettings()
	if set.TotalRAM() != 16 || !strings.HasPrefix(set.Describe(), "ZX81 16K PAL") {
		t.Fatalf("default: %d %q", set.TotalRAM(), set.Describe())
	}
	ts, _ := set.WithModel("ts1500")
	if ts.ROM != "ts1500" || !ts.NTSC || ts.TotalRAM() != 16 || !strings.Contains(ts.Describe(), "M1NOT") {
		t.Fatalf("ts1500: %+v %q", ts, ts.Describe())
	}
	ts.RAMK = 16
	if ts.TotalRAM() != 32 || ts.ramTop() != 32<<10+16383 {
		t.Fatalf("ts1500 + 16K pack: %d", ts.TotalRAM())
	}
	ts.RAMK = 48
	if ts.ramTop() != 65535 {
		t.Fatalf("ts1500 + 48K: %d", ts.ramTop())
	}
	r, _ := set.WithModel("R470")
	r.LowRAM = true
	cfg, err := EmuConfig(r)
	if err != nil || cfg.RAM816K {
		t.Fatalf("R-470 has no RAM at 8-16K: %v %v", cfg.RAM816K, err)
	}
	t5, _ := set.WithModel("ts1500")
	if cfg, _ := EmuConfig(t5); cfg.M1Not != 49152 || !cfg.NoROMShadow {
		t.Fatalf("ts1500 config: %+v", cfg)
	}
	z, _ := set.WithModel("zx80")
	if cfg, _ := EmuConfig(z); !cfg.ZX80 || z.ROM != "zx80" || z.TotalRAM() != 16 {
		t.Fatalf("zx80: %+v", z)
	}
	if _, err := set.WithModel("ace"); err == nil {
		t.Fatal("Jupiter Ace is not emulated yet")
	}
	if _, err := set.WithModel("spectrum"); err == nil || !strings.Contains(err.Error(), "tk85") {
		t.Fatalf("unknown model: %v", err)
	}
	s := newBooted(t)
	if out := mustExec(t, s, "machine tk85 ram=16"); !strings.HasPrefix(out, "TK85 32K NTSC · ROM tk85") {
		t.Fatalf("machine tk85: %q", out)
	}
	if out := mustExec(t, s, "machine modelo=zx81 lowram=on m1not=on fpfix=on"); !strings.Contains(out, "RAM 8-16K, M1NOT, correção PF") {
		t.Fatalf("machine options: %q", out)
	}
}

// TestZX80Programs types a program on a ZX80, lists it, saves it as a .o,
// loads it back after a reset and runs it.
func TestZX80Programs(t *testing.T) {
	set, _ := DefaultSettings().WithModel("zx80")
	s, err := New(set)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, s, "run 200")
	mustExec(t, s, "type 10")
	mustExec(t, s, "key O") // PRINT
	mustExec(t, s, `typeln "OLA"`)
	mustExec(t, s, "run 100")
	if out := mustExec(t, s, "list"); out != "10 PRINT \"OLA\"\n" {
		t.Fatalf("list = %q", out)
	}
	o := filepath.Join(t.TempDir(), "ola.o")
	if out := mustExec(t, s, "save "+o); !strings.Contains(out, "ZX80") {
		t.Fatalf("save = %q", out)
	}
	if text, err := basic.ListZX80(mustRead(t, o)); err != nil || text != "10 PRINT \"OLA\"\n" {
		t.Fatalf("saved file lists as %q %v", text, err)
	}
	mustExec(t, s, "reset")
	mustExec(t, s, "run 200")
	mustExec(t, s, "load "+o)
	mustExec(t, s, "typeln R")
	mustExec(t, s, "run 100")
	if scr := mustExec(t, s, "screen"); !strings.Contains(scr, "OLA") || !strings.Contains(scr, "0/10") {
		t.Fatalf("run:\n%s", scr)
	}
	var buf strings.Builder
	if err := s.Exec("find PRINT", &buf); err == nil {
		t.Fatal("find should refuse the ZX80")
	}
	if err := s.Exec("load programa.p", &buf); err == nil || !strings.Contains(err.Error(), ".o") {
		t.Fatalf("a .p on the ZX80: %v", err)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
