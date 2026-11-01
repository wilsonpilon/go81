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

	// The ZX80 ROM (InitPatches, MACHINEZX80 with the ZX80 ROM).
	patch80SaveDelay = 0x01ba3e // ZX80SaveDelayStartSaving
	patch80GetByte   = 0x02223e // ZX80GetByteStartLoading
	patch80OutByte   = 0x01cb11 // ZX80OutByteSaveByte
	patch80StopTape  = 0x0203c3 // ZX80ZX81LambdaStopTape
)

// zx80ROM reports whether the machine is a ZX80 with its own 4K ROM (with a
// ZX81 ROM, a ZX80 uses the ZX81 patches, as in EightyOne).
func (m *Machine) zx80ROM() bool { return m.cfg.ZX80 && m.romTop == 0x0FFF }

// ZX80ROM reports whether the machine runs the ZX80's own BASIC (a ZX80
// with its 4K ROM): programs are .o files, saved with no name.
func (m *Machine) ZX80ROM() bool { return m.zx80ROM() }

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
	if m.zx80ROM() {
		m.patchTestZX80(handle)
		return
	}
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
		if m.SaveEnd != nil {
			m.SaveEnd()
		}
	}
}

// patchTestZX80 is PatchTest with the ZX80 patches.
func (m *Machine) patchTestZX80(handle uint32) {
	c := m.CPU
	d := m.Tape
	switch handle {
	case patch80SaveDelay:
		if d.FlashSaveable() {
			c.SetDE(0x0001)
		}
	case patch80GetByte:
		d.WavStart()
		if p := d.Player; p.Byte != tape.ByteEmpty {
			m.Memory[c.HL()] = byte(p.Byte)
			c.PC = 0x0248
			p.Byte = tape.ByteEmpty
			p.EventCounter = 0
		}
	case patch80OutByte:
		if d.FlashSaveable() {
			if m.SaveByte != nil {
				m.SaveByte(m.Memory[c.HL()])
			}
			c.PC = 0x01f3
		}
	case patch80StopTape:
		d.WavStop()
		if m.SaveEnd != nil {
			m.SaveEnd()
		}
	}
}

// doAutoLoad is DoAutoLoad (ZX81 part): reset, then type LOAD "" ENTER.
// On a ZX80 with its ROM: LOAD (W) ENTER; the ZX81 keys that follow are
// pressed too, as EightyOne does.
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
	zx80 := m.zx80ROM()
	first, second := "J", "SHIFT"
	if zx80 {
		first, second = "W", "ENTER"
	}
	switch m.autoLoadCount {
	case 2:
		m.Keyboard = [8]byte{}
		m.CPU.Reset()
	case base + 0:
		key(first, true)
	case base + 10:
		key(first, false)
	case base + 20:
		key(second, true)
	case base + 30:
		if zx80 {
			key("ENTER", false)
		} else {
			key("P", true)
		}
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
