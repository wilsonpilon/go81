// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package z80_test

import (
	"os"
	"testing"

	"github.com/wilsonpilon/go81/internal/emu/z80"
)

// BenchmarkStep measures instructions per second running ZEXDOC code.
// The ZX81 runs at 3.25 MHz, roughly 0.5 million instructions per second.
func BenchmarkStep(b *testing.B) {
	prog, err := os.ReadFile("testdata/zexdoc.com")
	if err != nil {
		b.Fatal(err)
	}
	bus := &cpmBus{}
	copy(bus.mem[0x100:], prog)
	bus.mem[5] = 0xc9
	cpu := z80.New(bus)
	cpu.PC, cpu.SP = 0x100, 0xf000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if cpu.PC == 0 {
			cpu.PC = 0x100
		}
		cpu.Step()
	}
}
