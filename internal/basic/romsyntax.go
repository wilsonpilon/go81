// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"errors"
	"sync"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// The syntax of a line is checked by the ZX81 ROM itself, as when ENTER is
// pressed: the line goes into the edit line (E_LINE) of an emulated ZX81
// and LINE-SCAN (0CBAh) runs in syntax mode. On an error the ROM jumps to
// ERROR-1 (RST 08) with CH_ADD on the character where it stopped: that is
// where the ZX81 shows its S cursor.

const (
	romLineScan = 0x0CBA
	sysELine    = 0x4014
	sysXPtr     = 0x4018
	sysFlagX    = 0x402D
)

var romSyntax struct {
	sync.Mutex
	m   *zx81.Machine
	ram []byte // the RAM after power on, put back before each check
}

func syntaxMachine() (*zx81.Machine, error) {
	if romSyntax.m != nil {
		copy(romSyntax.m.Memory[0x4000:], romSyntax.ram)
		return romSyntax.m, nil
	}
	rom, err := assets.ROM("zx81.edition3")
	if err != nil {
		return nil, err
	}
	cfg := zx81.DefaultConfig(rom)
	cfg.RAMTop = zx81.RAMTopForPack(16)
	m, err := zx81.New(cfg)
	if err != nil {
		return nil, err
	}
	for range 200 {
		m.RunFrame()
	}
	romSyntax.m = m
	romSyntax.ram = append([]byte(nil), m.Memory[0x4000:0x8000]...)
	return m, nil
}

// ErrSyntax is a line the ZX81 would not accept.
var ErrSyntax = errors.New("erro de sintaxe")

// CheckSyntax checks a line as typed on the ZX81 (the number in digits and
// the body without the hidden numbers, see TypedCodes). It returns -1 if
// the ROM accepts it, or the index in codes where the ZX81 puts its S
// cursor.
func CheckSyntax(codes []byte) (int, error) {
	romSyntax.Lock()
	defer romSyntax.Unlock()
	m, err := syntaxMachine()
	if err != nil {
		return -1, err
	}
	eline := peek16(m, sysELine)
	for i, c := range codes {
		m.Poke(eline+uint16(i), c)
	}
	end := eline + uint16(len(codes))
	m.Poke(end, Newline)
	poke16(m, stkBot, end+1)
	poke16(m, stkEnd, end+1)
	poke16(m, sysXPtr, 0)
	m.Poke(sysFlagX, 0)
	c := m.CPU
	c.SP = 0x7F00
	c.SP -= 2
	poke16(m, c.SP, retAddr)
	c.PC, c.IFF1, c.IFF2, c.Halted = romLineScan, 0, 0, false
	for steps := 0; c.PC != retAddr; steps++ {
		if c.PC == romError {
			pos := int(peek16(m, chAdd)) - int(eline)
			return max(0, min(pos, len(codes))), nil
		}
		if steps > 5_000_000 {
			romSyntax.m = nil
			return -1, errors.New("a ROM não terminou de conferir a linha")
		}
		c.Step()
	}
	return -1, nil
}

// CheckLine is CheckSyntax for a line (see TypedCodes): nil, or ErrSyntax
// and the index in TypedCodes(l) where the error is.
func CheckLine(l Line) (int, error) {
	codes := TypedCodes(l)
	pos, err := CheckSyntax(codes)
	if err != nil {
		return -1, err
	}
	if pos >= 0 {
		return pos, ErrSyntax
	}
	return -1, nil
}
