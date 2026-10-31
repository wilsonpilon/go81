// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package zx81_test

import (
	"testing"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// BenchmarkFrame measures one emulated frame (1/50 s of a real ZX81).
func BenchmarkFrame(b *testing.B) {
	rom, _ := assets.ROM("zx81.edition3")
	m, err := zx81.New(zx81.DefaultConfig(rom))
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		m.RunFrame()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RunFrame()
	}
}
