// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package z80_test

import (
	"os"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/emu/z80"
)

// cpmBus is a flat 64K RAM with nothing on the I/O ports.
type cpmBus struct{ mem [65536]byte }

func (b *cpmBus) OpcodeFetch(a uint16) byte        { return b.mem[a] }
func (b *cpmBus) ReadOperand(a uint16) byte        { return b.mem[a] }
func (b *cpmBus) Read(a uint16) byte               { return b.mem[a] }
func (b *cpmBus) Write(a uint16, v byte)           { b.mem[a] = v }
func (b *cpmBus) In(port uint16, ts *int) byte     { return 0xff }
func (b *cpmBus) Out(port uint16, v byte, ts *int) {}

// runCPM runs a CP/M .COM program with a minimal BDOS (functions 2 and 9)
// and returns everything it printed.
func runCPM(t *testing.T, path string) string {
	t.Helper()
	prog, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b := &cpmBus{}
	copy(b.mem[0x100:], prog)
	b.mem[0x0005] = 0xc9 // BDOS entry: RET
	b.mem[0x0006] = 0x00 // top of memory at 0xF000 (read by ZEX for SP)
	b.mem[0x0007] = 0xf0
	cpu := z80.New(b)
	cpu.PC = 0x100
	cpu.SP = 0xf000

	var out strings.Builder
	for {
		switch cpu.PC {
		case 0x0000:
			return out.String()
		case 0x0005:
			switch cpu.C {
			case 2:
				out.WriteByte(cpu.E)
			case 9:
				for a := cpu.DE(); b.mem[a] != '$'; a++ {
					out.WriteByte(b.mem[a])
				}
			}
			if strings.HasSuffix(out.String(), "\n") && testing.Verbose() {
				line := out.String()
				if i := strings.LastIndex(line[:len(line)-1], "\n"); i >= 0 {
					line = line[i+1:]
				}
				t.Log(strings.TrimSpace(line))
			}
		}
		cpu.Step()
	}
}

func runZex(t *testing.T, name string) string {
	if os.Getenv("GO81_LONG") == "" {
		t.Skip("long test: set GO81_LONG=1 to run (takes minutes)")
	}
	out := runCPM(t, "testdata/"+name)
	if !strings.Contains(out, "Tests complete") {
		t.Fatalf("%s did not finish:\n%s", name, out)
	}
	return out
}

// countErrors returns the lines of the ZEX output that report an error.
func countErrors(out string) []string {
	var bad []string
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "ERROR") {
			bad = append(bad, strings.TrimSpace(l))
		}
	}
	return bad
}

// The EightyOne core is an old FUSE core: its DAA flags differ from a real
// Z80 and it has no MEMPTR, so it fails a few ZEX groups. These are the exact
// lines it prints (recorded with test/compat/z80ref/cpm.c); the port must
// print the same ones, with the same CRCs.
var (
	eightyOneZexdoc = []string{
		"<daa,cpl,scf,ccf>.............  ERROR **** crc expected:9b4ba675 found:cb604fa8",
	}
	eightyOneZexall = []string{
		"bit n,(<ix,iy>+1).............  ERROR **** crc expected:83534ee1 found:2b526613",
		"bit n,<b,c,d,e,h,l,(hl),a>....  ERROR **** crc expected:5e020e98 found:fe7d580c",
		"<daa,cpl,scf,ccf>.............  ERROR **** crc expected:6d2dd213 found:9f28af32",
	}
)

func checkZex(t *testing.T, name string, want []string) {
	got := countErrors(runZex(t, name))
	g, w := strings.Join(got, "\n"), strings.Join(want, "\n")
	if g != w {
		t.Fatalf("%s errors differ from EightyOne\nwant:\n%s\ngot:\n%s", name, w, g)
	}
}

// TestZexdoc runs ZEXDOC (documented flags only).
func TestZexdoc(t *testing.T) { checkZex(t, "zexdoc.com", eightyOneZexdoc) }

// TestZexall runs ZEXALL (all flags, including undocumented bits 3 and 5).
func TestZexall(t *testing.T) { checkZex(t, "zexall.com", eightyOneZexall) }
