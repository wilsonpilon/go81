// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne zx81/rompatch.cpp and main_.cpp (DoAutoLoad) —
// Copyright (C) 2003-2025 Michael D Wynne (GPL v2 or later).

package zx81

import "github.com/wilsonpilon/go81/internal/media/tape"

// ROM patch handles: address<<8 | expected byte at that address.
const (
	patchSaveDelay = 0x02ffcd // ZX81SaveDelayStartSaving
	patchGetByte   = 0x03561f // ZX81LambdaGetByteStartLoading
	patchOutByte   = 0x031e5e // ZX81LambdaOutByteSaveByte
	patchStopTape  = 0x020721 // ZX80ZX81LambdaStopTape
)

// InsertTape puts a program in the tape deck (like opening a .p file in
// EightyOne). With autoLoad the machine is reset and LOAD "" is typed.
func (m *Machine) InsertTape(block []byte, autoLoad bool) {
	if m.Tape == nil {
		m.Tape = tape.NewDeck(3250000)
	}
	m.Tape.Insert(block)
	if autoLoad {
		m.autoLoadCount = 2
	}
}

// pop16 reads a word from the stack straight from memory (rompatch.cpp).
func (m *Machine) pop16() uint16 {
	c := m.CPU
	lo := m.Memory[c.SP]
	c.SP++
	hi := m.Memory[c.SP]
	c.SP++
	return uint16(hi)<<8 | uint16(lo)
}

// patchTest is PatchTest: called before each instruction.
func (m *Machine) patchTest() {
	if m.Tape == nil {
		return
	}
	c := m.CPU
	pc := c.PC
	handle := uint32(pc)<<8 | uint32(m.Memory[pc])
	d := m.Tape
	switch handle {
	case patchSaveDelay:
		if d.FlashSaveable() {
			c.SetDE(0x0001)
		}
	case patchGetByte:
		d.WavStart()
		if p := d.Player; p.Byte != tape.ByteEmpty {
			c.C = byte(p.Byte)
			c.PC = m.pop16()
			p.Byte = tape.ByteEmpty
			p.EventCounter = 0
		}
	case patchOutByte:
		if d.FlashSaveable() {
			if m.SaveByte != nil {
				m.SaveByte(m.Memory[c.HL()])
			}
			c.PC = m.pop16()
		}
	case patchStopTape:
		d.WavStop()
	}
}

// doAutoLoad is DoAutoLoad (ZX81 part): reset, then type LOAD "" ENTER.
// The counter advances 2 per frame (PAL) or 1 (NTSC).
func (m *Machine) doAutoLoad() {
	key := func(name string, down bool) {
		k := Keys[name]
		if down {
			m.Keyboard[k.Row] |= 1 << k.Bit
		} else {
			m.Keyboard[k.Row] &^= 1 << k.Bit
		}
	}
	const base = 340
	switch m.autoLoadCount {
	case 2:
		m.Keyboard = [8]byte{}
		m.CPU.Reset()
	case base + 0:
		key("J", true)
	case base + 10:
		key("J", false)
	case base + 20:
		key("SHIFT", true)
	case base + 30:
		key("P", true)
	case base + 40:
		key("P", false)
	case base + 50:
		key("P", true)
	case base + 60:
		key("P", false)
	case base + 70:
		key("SHIFT", false)
	case base + 80:
		key("ENTER", true)
	case base + 90:
		key("ENTER", false)
	}
	if m.cfg.NTSC {
		m.autoLoadCount++
	} else {
		m.autoLoadCount += 2
	}
	if m.autoLoadCount == 1000 {
		m.autoLoadCount = 0
	}
}

// AutoLoading reports whether the auto-load sequence is still running.
func (m *Machine) AutoLoading() bool { return m.autoLoadCount != 0 }
