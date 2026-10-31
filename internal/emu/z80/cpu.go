// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from the EightyOne Z80 core (z80.c, z80_macros.h):
//   Copyright (c) 1999-2003 Philip Kendall, Witold Filipczyk
//   EightyOne modifications (c) 2003-2025 Michael D Wynne
//   Licensed GPL v2 or later.

// Package z80 is a Z80 CPU core that reproduces the behaviour of the core
// used by the EightyOne emulator (itself derived from FUSE): same results,
// same T-state counts, same M-cycle bookkeeping and the same order of bus
// accesses.
//
// The only deliberate difference: addresses are always wrapped to 16 bits.
// EightyOne can pass values such as IX+d above 0xFFFF to the machine, which
// is undefined behaviour in its C code.
package z80

// Flag bits of the F register.
const (
	FlagC byte = 0x01
	FlagN byte = 0x02
	FlagP byte = 0x04
	FlagV      = FlagP
	Flag3 byte = 0x08
	FlagH byte = 0x10
	Flag5 byte = 0x20
	FlagZ byte = 0x40
	FlagS byte = 0x80
)

// Bus is what the CPU talks to. It mirrors the EightyOne machine hooks.
//
// ts points to the T-states counted so far in the current instruction; port
// handlers may read it and add wait states, as in EightyOne.
type Bus interface {
	// OpcodeFetch is used for M1 fetches and most immediate operands.
	OpcodeFetch(addr uint16) byte
	// ReadOperand is used for JP, JR and CALL operands.
	ReadOperand(addr uint16) byte
	// Read is a normal memory read by the program.
	Read(addr uint16) byte
	// Write is a normal memory write by the program.
	Write(addr uint16, v byte)
	// In reads an I/O port.
	In(port uint16, ts *int) byte
	// Out writes an I/O port.
	Out(port uint16, v byte, ts *int)
}

const maxMCycles = 11

// CPU holds the Z80 state.
type CPU struct {
	A, F, B, C, D, E, H, L byte
	AF_, BC_, DE_, HL_     uint16
	IX, IY, SP, PC         uint16
	I                      byte
	R                      uint16 // low 7 bits of R (wider, as in EightyOne)
	R7                     byte   // bit 7 of R
	IFF1, IFF2, IM         byte
	Halted                 bool

	// TStates is the length of the last instruction or interrupt.
	TStates int

	// Debugger bookkeeping, as in EightyOne.
	StackChange int
	RetExecuted bool
	NMIOccurred bool
	RZXCounter  uint16

	mCycles     [maxMCycles + 1]int // +1: room for the 0 terminator
	mCycleIndex int
	ioMCycle    int
	m1Cycles    int

	ts  int
	bus Bus
}

// New creates a CPU attached to bus, already reset.
func New(bus Bus) *CPU {
	c := &CPU{bus: bus}
	c.Reset()
	return c
}

// SetBus changes the bus.
func (c *CPU) SetBus(bus Bus) { c.bus = bus }

// Reset puts the CPU in its power-on state (all registers zero).
func (c *CPU) Reset() {
	c.A, c.F, c.B, c.C, c.D, c.E, c.H, c.L = 0, 0, 0, 0, 0, 0, 0, 0
	c.AF_, c.BC_, c.DE_, c.HL_ = 0, 0, 0, 0
	c.IX, c.IY = 0, 0
	c.I, c.R, c.R7 = 0, 0, 0
	c.SP, c.PC = 0, 0
	c.IFF1, c.IFF2, c.IM = 0, 0, 0
	c.Halted = false
	c.StackChange = 0
	c.NMIOccurred = false
}

// Register pair helpers.

func (c *CPU) AF() uint16 { return uint16(c.A)<<8 | uint16(c.F) }
func (c *CPU) BC() uint16 { return uint16(c.B)<<8 | uint16(c.C) }
func (c *CPU) DE() uint16 { return uint16(c.D)<<8 | uint16(c.E) }
func (c *CPU) HL() uint16 { return uint16(c.H)<<8 | uint16(c.L) }

func (c *CPU) SetAF(v uint16) { c.A, c.F = byte(v>>8), byte(v) }
func (c *CPU) SetBC(v uint16) { c.B, c.C = byte(v>>8), byte(v) }
func (c *CPU) SetDE(v uint16) { c.D, c.E = byte(v>>8), byte(v) }
func (c *CPU) SetHL(v uint16) { c.H, c.L = byte(v>>8), byte(v) }

// RegR returns the full R register as the program sees it.
func (c *CPU) RegR() byte { return byte(c.R&0x7f) | c.R7&0x80 }

// MCycles returns the M-cycle lengths of the last instruction, terminated by 0.
func (c *CPU) MCycles() []int {
	n := c.mCycleIndex + 1
	if n > maxMCycles {
		n = maxMCycles
	}
	return c.mCycles[:n]
}

// CurrentTStates returns the T-states counted so far in the instruction
// being executed. Bus handlers can use it for timing.
func (c *CPU) CurrentTStates() int { return c.ts }

// ResetDebugState clears the per-instruction and debugger bookkeeping.
func (c *CPU) ResetDebugState() {
	c.StackChange = 0
	c.RetExecuted = false
	c.m1Cycles = 0
	c.mCycles = [maxMCycles + 1]int{}
	c.mCycleIndex = -1
	c.ioMCycle = -1
	c.ts = 0
}

// IOMCycle returns the index of the I/O M-cycle of the last instruction, or -1.
func (c *CPU) IOMCycle() int { return c.ioMCycle }

// M1Cycles returns the number of M1 cycles of the last instruction.
func (c *CPU) M1Cycles() int { return c.m1Cycles }

// Flag lookup tables.
var (
	halfcarryAdd = [8]byte{0, FlagH, FlagH, FlagH, 0, 0, 0, FlagH}
	halfcarrySub = [8]byte{0, 0, FlagH, 0, FlagH, 0, FlagH, FlagH}
	overflowAdd  = [8]byte{0, 0, 0, FlagV, FlagV, 0, 0, 0}
	overflowSub  = [8]byte{0, FlagV, 0, 0, 0, 0, FlagV, 0}

	sz53   [256]byte
	parity [256]byte
	sz53p  [256]byte
)

func init() {
	for i := 0; i < 256; i++ {
		sz53[i] = byte(i) & (Flag3 | Flag5 | FlagS)
		p := byte(0)
		for j := i; j != 0; j >>= 1 {
			p ^= byte(j & 1)
		}
		if p == 0 {
			parity[i] = FlagP
		}
		sz53p[i] = sz53[i] | parity[i]
	}
	sz53[0] |= FlagZ
	sz53p[0] |= FlagZ
}

// M-cycle bookkeeping (InsertMCycle / AddToMCycle).

func (c *CPU) insM(n int) {
	if c.mCycleIndex < maxMCycles-1 {
		c.mCycleIndex++
		c.mCycles[c.mCycleIndex] = n
		c.mCycles[c.mCycleIndex+1] = 0
	}
}

func (c *CPU) addM(n int) {
	if c.mCycleIndex < maxMCycles-1 {
		c.mCycles[c.mCycleIndex] += n
	}
}

// Bus access helpers.

// fetch reads the byte at PC through OpcodeFetch and advances PC.
func (c *CPU) fetch() byte {
	v := c.bus.OpcodeFetch(c.PC)
	c.PC++
	return v
}

// operand reads the byte at PC through ReadOperand and advances PC.
func (c *CPU) operand() byte {
	v := c.bus.ReadOperand(c.PC)
	c.PC++
	return v
}

func (c *CPU) rd(a uint16) byte     { return c.bus.Read(a) }
func (c *CPU) wr(a uint16, v byte)  { c.bus.Write(a, v) }
func (c *CPU) in(p uint16) byte     { return c.bus.In(p, &c.ts) }
func (c *CPU) out(p uint16, v byte) { c.bus.Out(p, v, &c.ts) }

// Interrupt handles a maskable interrupt with the given value on the data
// bus. It returns the T-states used, or 0 if interrupts are disabled.
func (c *CPU) Interrupt(data byte) int {
	if c.IFF1 == 0 {
		return 0
	}
	if c.Halted {
		c.PC++
		c.Halted = false
	}
	c.IFF1, c.IFF2 = 0, 0
	c.SP--
	c.wr(c.SP, byte(c.PC>>8))
	c.SP--
	c.wr(c.SP, byte(c.PC))
	c.R++
	switch c.IM {
	case 0, 1:
		// IM 0 assumes RST 38h on the data bus.
		c.PC = 0x0038
		c.StackChange += 2
		return 13
	case 2:
		v := uint16(c.I)<<8 + uint16(data)
		lo := c.rd(v)
		v++
		hi := c.rd(v)
		c.PC = uint16(hi)<<8 | uint16(lo)
		c.StackChange += 2
		return 19
	default:
		return 12
	}
}

// NMI handles a non-maskable interrupt and returns the T-states used.
func (c *CPU) NMI() int {
	c.StackChange += 2
	c.IFF1 = 0
	if c.Halted {
		c.Halted = false
		c.PC++
	}
	c.SP--
	c.wr(c.SP, byte(c.PC>>8))
	c.SP--
	c.wr(c.SP, byte(c.PC))
	c.m1Cycles++
	c.R++
	c.PC = 0x0066
	c.NMIOccurred = false
	return 11
}
