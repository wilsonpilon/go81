// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"errors"
	"fmt"
	"sync"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// The ZX81 converts the digits of a number with its own arithmetic
// (routine DEC-TO-FP, 14D9H), which truncates: 0.1 is 7D 4C CC CC CC, not
// ...CD, and 0.5 becomes 7F 7F FF FF FF. To store exactly the bytes the ZX81
// stores when a line is typed, the loader runs that ROM routine on an
// emulated Z80.

const (
	romDecToFP = 0x14D9
	romError   = 0x0008 // RST 08: error report
	chAdd      = 0x4016
	stkBot     = 0x401A
	stkEnd     = 0x401C
	sysMEM     = 0x401F
	memBot     = 0x405D
	scratch    = 0x6000 // where the digits are placed (16K RAM, free)
	retAddr    = 0xFFFF // return address that ends the call
)

var romCalc struct {
	sync.Mutex
	m     *zx81.Machine
	cache map[string]romResult
}

type romResult struct {
	v    [5]byte
	read int
	err  error
}

func calcMachine() (*zx81.Machine, error) {
	if romCalc.m != nil {
		return romCalc.m, nil
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
	for range 200 { // power on: the ROM sets up the system variables
		m.RunFrame()
	}
	romCalc.m = m
	return m, nil
}

// ROMFloat converts a number written in ZX81 characters (digits, ".", "E",
// "+", "-") the way the ZX81 ROM does. It returns the 5 bytes and how many
// characters the ROM read.
func ROMFloat(codes []byte) ([5]byte, int, error) {
	romCalc.Lock()
	defer romCalc.Unlock()
	key := string(codes)
	if r, ok := romCalc.cache[key]; ok {
		return r.v, r.read, r.err
	}
	if romCalc.cache == nil {
		romCalc.cache = map[string]romResult{}
	}
	v, read, err := romFloat(codes)
	romCalc.cache[key] = romResult{v, read, err}
	return v, read, err
}

func romFloat(codes []byte) ([5]byte, int, error) {
	var out [5]byte
	m, err := calcMachine()
	if err != nil {
		return out, 0, err
	}
	for i, c := range codes {
		m.Poke(uint16(scratch+i), c)
	}
	m.Poke(uint16(scratch+len(codes)), Newline)
	poke16(m, chAdd, scratch)
	poke16(m, stkEnd, peek16(m, stkBot))
	poke16(m, sysMEM, memBot) // calculator memory, as when a line is checked
	c := m.CPU
	c.SP = 0x7F00
	c.SP -= 2
	poke16(m, c.SP, retAddr)
	// As after GET-CHAR (RST 18H): A = the current character, HL = CH_ADD.
	c.A = m.Peek(scratch)
	c.SetHL(scratch)
	c.PC, c.IFF1, c.IFF2, c.Halted = romDecToFP, 0, 0, false
	for steps := 0; c.PC != retAddr; steps++ {
		if c.PC == romError || steps > 2_000_000 {
			romCalc.m = nil // the machine state is lost: make a new one next time
			return out, 0, ErrNumberTooBig
		}
		c.Step()
	}
	end := peek16(m, stkEnd)
	if end-5 < peek16(m, stkBot) {
		return out, 0, errors.New("a ROM não deixou o número na pilha")
	}
	for i := range out {
		out[i] = m.Peek(end - 5 + uint16(i))
	}
	read := int(peek16(m, chAdd)) - scratch
	return out, read, nil
}

// ROMFloatText is ROMFloat for a number in ASCII ("3.14", "1E-5").
func ROMFloatText(s string) ([5]byte, error) {
	codes := make([]byte, 0, len(s))
	for _, r := range s {
		c, ok := asciiCode(r)
		if !ok {
			return [5]byte{}, fmt.Errorf("número inválido: %s", s)
		}
		codes = append(codes, c)
	}
	b, n, err := ROMFloat(codes)
	if err == nil && n != len(codes) {
		err = fmt.Errorf("número inválido: %s", s)
	}
	return b, err
}

func peek16(m *zx81.Machine, a uint16) uint16 {
	return uint16(m.Peek(a)) | uint16(m.Peek(a+1))<<8
}

func poke16(m *zx81.Machine, a, v uint16) {
	m.Poke(a, byte(v))
	m.Poke(a+1, byte(v>>8))
}

// asciiCode maps a character (as in the text format) to a ZX81 code 0-63.
// Lower case letters are upper case.
func asciiCode(r rune) (byte, bool) {
	if r >= 'a' && r <= 'z' {
		r -= 'a' - 'A'
	}
	for i, s := range charASCII {
		if len(s) == 1 && rune(s[0]) == r {
			return byte(i), true
		}
	}
	return 0, false
}
