// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne zx81/zx81.cpp — Copyright (C) 2003-2025 Michael D Wynne,
// Paul Farrow and contributors (GPL v2 or later).

// Package zx81 emulates the ZX81 family as EightyOne does: memory map, I/O
// ports, video generation driven by the Z80 (opcode fetches above M1NOT,
// NMI generator, HSYNC/VSYNC) and the TV raster.
//
// Version 1.3 covers the stock machine (ZX81 / TS1000) with RAM packs, the
// 8-16K RAM option and WRX hi-res. Other peripherals are "not fitted".
package zx81

import (
	"fmt"

	"github.com/wilsonpilon/go81/internal/emu/z80"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

// Hi-res modes (zx81config.h).
const (
	HiResNone       = 0
	HiResWRX        = 1
	HiResG007       = 2
	HiResMemotech   = 3
	HiResQuicksilva = 4
)

// Config describes the emulated hardware.
type Config struct {
	ROM              []byte // ROM image loaded at address 0
	RAMTop           int    // highest RAM address (16K pack: 32767)
	M1Not            int    // 32768, or 49152 with the M1NOT modification
	MaxIReg          int    // highest I value that points at the character set (0x1F)
	HiRes            int    // HiResNone or HiResWRX
	RAM816K          bool   // RAM fitted at 8-16K
	RAM816KWP        bool   // ... write protected
	ShadowROM        bool
	NoROMShadow      bool // no ROM copy at 32-48K: the idle data bus is read (TS1500)
	ProtectROM       bool
	NTSC             bool
	ImprovedWait     bool
	FloatingPointFix bool // ZX81 floating point hardware fix
	VSyncSound       bool // EightyOne "Video Sound": the beeper follows the video sync signal
	ExcludeHSyncs    bool // with VSyncSound: ignore the HSYNC pulses
	TStatesPerLine   int  // 207
	Scanlines        int  // 310 (PAL) or 262 (NTSC)
}

// DefaultConfig returns a stock 16K ZX81 with the given ROM.
func DefaultConfig(rom []byte) Config {
	return Config{
		ROM:            rom,
		RAMTop:         32767,
		M1Not:          32768,
		MaxIReg:        0x1F,
		ProtectROM:     true,
		TStatesPerLine: 207,
	}
}

// RAMTopForPack returns RAMTop for a RAM pack of k kilobytes (0 = none, the
// 1K internal RAM), as EightyOne's hardware dialog computes it.
func RAMTopForPack(k int) int {
	if k == 0 {
		k = 1
	}
	t := k<<10 + 16383
	if t > 65535 {
		t = 65535
	}
	return t
}

// Last I/O operation of an instruction (LASTINST*).
const (
	lastNone = iota
	lastInFE
	lastOutFE
	lastOutFD
	lastOutFF
)

const (
	colourBlack       = 0
	colourBrightWhite = 15
	idleDataBus       = 0xFF
)

// Machine is an emulated ZX81.
type Machine struct {
	cfg Config
	CPU *z80.CPU

	Memory   [65536]byte
	romTop   int
	Keyboard [8]byte // keyboard rows, bit set = key down (as EightyOne)

	// Video state (globals in zx81.cpp).
	nmiGeneratorEnabled  bool
	syncOutputWhite      bool
	lineCounter          int
	lineClockCounter     int
	lineClockCarry       int
	frameTStates         int
	tStatesCount         int
	configByte           int
	lastInstruction      int
	nmiWaitPositionStart int
	interruptPending     bool
	z80Halted            bool
	frameSynchronised    bool
	vsyncFound           bool
	scanlineCounter      int
	shiftRegister        int32
	shiftRegInv          int32
	ink, paper, border   int

	scanlinePixelLength  int
	backporchStart       int
	backporchEnd         int
	nmiWaitPositionEnd   int
	nmiWaitPositionAfter int

	carryOver [maxScanlineActivePixelLength]byte
	line      Scanline
	raster    Raster
	borrow    int

	allowSound bool
	snd        beeper

	// Tape is the tape deck (nil = no tape manager).
	Tape *tape.Deck
	// SaveByte receives each byte saved with SAVE (flash save).
	SaveByte func(byte)
	// SaveEnd is called when the ROM finishes a LOAD or SAVE.
	SaveEnd       func()
	autoLoadCount int

	// LineHook, if set, is called with every scanline before it is drawn.
	LineHook func(*Scanline)
	// StepHook, if set, is called before each instruction (debugger,
	// tracing in tests). Leave it nil when not needed.
	StepHook func()

	// ScanlinesPerFrame mirrors emulator.scanlinesPerFrame (0 = not in sync).
	ScanlinesPerFrame int
}

// New creates and initialises a machine.
func New(cfg Config) (*Machine, error) {
	if len(cfg.ROM) == 0 || len(cfg.ROM) > 65536 {
		return nil, fmt.Errorf("zx81: bad ROM size %d", len(cfg.ROM))
	}
	if cfg.TStatesPerLine == 0 {
		cfg.TStatesPerLine = 207
	}
	if cfg.Scanlines == 0 {
		cfg.Scanlines = 310
		if cfg.NTSC {
			cfg.Scanlines = 262
		}
	}
	m := &Machine{cfg: cfg}
	m.CPU = z80.New(m)
	m.initialise()
	return m, nil
}

// Config returns the machine configuration.
func (m *Machine) Config() Config { return m.cfg }

// TStatesPerFrame returns the length of a frame in T-states.
func (m *Machine) TStatesPerFrame() int { return m.cfg.TStatesPerLine * m.cfg.Scanlines }

// initialise is zx81_initialise for the stock machine.
func (m *Machine) initialise() {
	m.tStatesCount = 0
	for i := range m.Memory {
		m.Memory[i] = 7
	}
	n := copy(m.Memory[:], m.cfg.ROM)
	m.romTop = n - 1

	m.nmiGeneratorEnabled = false
	m.syncOutputWhite = false
	m.CPU.Reset()

	tpl := m.cfg.TStatesPerLine
	m.scanlinePixelLength = tpl * 2
	m.backporchStart = tpl
	m.backporchEnd = tpl - backporchDuration + 1
	m.nmiWaitPositionEnd = tpl
	m.nmiWaitPositionAfter = tpl - 1

	m.lineClockCounter = tpl
	m.lineCounter = 0
	m.lineClockCarry = 0
	m.nmiWaitPositionStart = 0
	m.scanlineCounter = 0
	m.vsyncFound = false

	m.ink = colourBlack
	m.paper = colourBrightWhite
	m.border = colourBrightWhite

	m.raster.init(tpl, m.cfg.NTSC)

	m.allowSound = false
	fps := 50
	if m.cfg.NTSC {
		fps = 60
	}
	m.snd.init(fps, m.TStatesPerFrame())
}

// Reset is a hard reset (power cycle), like choosing Reset in EightyOne.
func (m *Machine) Reset() { m.initialise() }

// Memory map ------------------------------------------------------------------

// Peek reads memory as the CPU sees it (zx81_ReadByte).
func (m *Machine) Peek(a uint16) byte {
	addr := int(a)
	switch {
	case m.cfg.RAM816K && addr >= 0x2000 && addr < 0x4000:
		return m.Memory[addr]
	case addr >= 0x4000 && addr <= m.cfg.RAMTop:
		return m.Memory[addr]
	case addr <= m.romTop:
		return m.Memory[addr]
	case addr&0x7FFF >= 0x4000:
		// RAM shadows at 16-32K and 48-64K.
		return m.Memory[0x4000|(addr&m.cfg.RAMTop&0x3FFF)]
	case m.cfg.RAM816K:
		return m.Memory[addr&0x7FFF]
	case m.cfg.NoROMShadow:
		return idleDataBus
	default:
		return m.Memory[addr&m.romTop]
	}
}

// Poke writes memory as the CPU does (zx81_WriteByte), honouring ROM protection.
func (m *Machine) Poke(a uint16, v byte) {
	addr := int(a)
	if addr <= m.romTop && m.cfg.ProtectROM {
		return
	}
	if addr > m.cfg.RAMTop {
		switch {
		case addr&0x7FFF >= 0x4000:
			addr = 0x4000 | (addr & m.cfg.RAMTop & 0x3FFF)
		case addr&0x7FFF >= 0x2000 && m.cfg.RAM816K && !m.cfg.RAM816KWP:
			addr &= 0x7FFF
		default:
			addr &= m.romTop
		}
	}
	if addr <= m.romTop && m.cfg.ProtectROM {
		return
	}
	if addr > 8191 && addr < 16384 {
		if m.cfg.ShadowROM && m.cfg.ProtectROM {
			return
		}
		if !m.cfg.RAM816K || m.cfg.RAM816KWP {
			return
		}
	}
	m.Memory[addr] = v
}

// z80.Bus implementation --------------------------------------------------------

// Read is a program memory read.
func (m *Machine) Read(a uint16) byte { return m.Peek(a) }

// Write is a program memory write.
func (m *Machine) Write(a uint16, v byte) { m.Poke(a, v) }

// ReadOperand is an instruction operand read.
func (m *Machine) ReadOperand(a uint16) byte { return m.Peek(a) }

// OpcodeFetch is zx81_opcode_fetch: below M1NOT it is a normal fetch; above
// it, opcodes with bit 6 clear are turned into NOPs and their value drives
// the video shift register.
func (m *Machine) OpcodeFetch(a uint16) byte {
	addr := int(a)
	if addr < m.cfg.M1Not {
		data := m.Peek(a)
		if m.cfg.FloatingPointFix && addr&0x4BFF == 0x0335 {
			data &= 0xBF
		}
		return data
	}

	fetchAddr := addr
	if fetchAddr >= 49152 {
		fetchAddr &= 32767
	}
	data := m.Peek(uint16(fetchAddr))
	opcode := data
	bit6 := opcode&64 != 0
	if !bit6 {
		opcode = 0
	}
	inv := data&128 != 0

	c := m.CPU
	i := int(c.I)
	region8K := i >= 0x20 && i < 0x40
	wrxAccess := m.cfg.HiRes == HiResWRX && !bit6
	notChr128 := i > m.cfg.MaxIReg
	chr128mode8kRAM := region8K && m.cfg.RAM816K

	update := false
	switch {
	case (notChr128 || chr128mode8kRAM) && wrxAccess:
		ra := uint16(i)<<8 | uint16(c.R7&128) | uint16((c.R-1)&127)
		data = m.Peek(ra)
		update = true
	case !bit6:
		data &= 63
		if i < 0x40 {
			ca := ((i&254)<<8 + int(data)<<3) | m.lineCounter
			data = m.Peek(uint16(ca))
		} else {
			data = 255
		}
		update = true
	}

	if update {
		m.border = colourBrightWhite
		m.shiftRegister |= int32(data)
		if inv {
			m.shiftRegInv |= 255
		}
		return 0
	}
	return opcode
}

// Out is zx81_writeport for the stock machine.
func (m *Machine) Out(port uint16, v byte, ts *int) {
	switch port & 255 {
	case 0x01:
		m.configByte = int(v)
	case 0xfd:
		m.lastInstruction = lastOutFD
	case 0x07, 0x1f, 0x3f, 0x73, 0x77, 0xbf, 0xc7, 0xcf, 0xd7, 0xdd, 0xfb:
		// Peripheral ports; nothing fitted.
	default:
		if port&1 == 0 {
			m.lastInstruction = lastOutFE
		}
	}
	if m.lastInstruction == lastNone {
		m.lastInstruction = lastOutFF
	}
	if m.allowSound {
		m.snd.set(true, m.frameTStates)
	}
}

// In is zx81_readport for the stock machine.
func (m *Machine) In(port uint16, ts *int) byte {
	if port&1 == 0 {
		if m.allowSound {
			m.snd.set(false, m.frameTStates)
		}
		data := byte(0)
		if m.cfg.NTSC {
			data |= 64
		}
		if m.Tape == nil || !m.Tape.EarState() {
			data |= 128
		}
		m.lastInstruction = lastInFE
		keyb := byte(port >> 8)
		for i := 0; i < 8; i++ {
			if keyb&(1<<i) == 0 {
				data |= m.Keyboard[i]
			}
		}
		return ^data
	}
	switch port & 255 {
	case 0x01:
		return m.configRead(m.configByte)
	case 0x07:
		return 0xfd
	case 0x17:
		return 0
	}
	return idleDataBus
}

// configRead emulates EightyOne's port 0x01, which returns raw bytes of its
// ZX81 configuration structure (an emulator detection feature).
func (m *Machine) configRead(n int) byte {
	b2 := func(v bool) byte {
		if v {
			return 1
		}
		return 0
	}
	le := func(v, k int) byte { return byte(v >> (8 * k)) }
	switch {
	case n == 0: // extfont
		return 0
	case n == 1:
		return b2(m.cfg.ShadowROM)
	case n == 2:
		return b2(m.cfg.RAM816K)
	case n == 3:
		return b2(m.cfg.RAM816KWP)
	case n == 4:
		return byte(m.cfg.HiRes)
	case n >= 8 && n < 12:
		return le(m.cfg.MaxIReg, n-8)
	case n >= 16 && n < 20:
		return le(m.cfg.RAMTop, n-16)
	case n >= 20 && n < 24:
		return le(m.romTop, n-20)
	case n >= 24 && n < 28:
		return le(m.cfg.M1Not, n-24)
	case n == 28:
		return b2(m.cfg.ImprovedWait)
	case n == 30: // chromaMode default
		return 0x0F
	case n == 32:
		return b2(m.cfg.FloatingPointFix)
	}
	return 0
}

// SetVSyncSound turns EightyOne's "Video Sound" on or off while running.
func (m *Machine) SetVSyncSound(on bool) { m.cfg.VSyncSound = on }
