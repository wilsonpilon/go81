// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne z80_ed.c — Copyright (c) 1999-2003 Philip Kendall
// (GPL v2 or later).

package z80

// ldBlockFlags sets flags for LDI/LDD/LDIR/LDDR.
func (c *CPU) ldBlockFlags(v byte) {
	v += c.A
	f := c.F&(FlagC|FlagZ|FlagS) | v&Flag3
	if c.BC() != 0 {
		f |= FlagV
	}
	if v&0x02 != 0 {
		f |= Flag5
	}
	c.F = f
}

// cpBlock is the common part of CPI/CPD/CPIR/CPDR. dir is +1 or -1.
func (c *CPU) cpBlock(dir uint16) {
	value := c.rd(c.HL())
	t := c.A - value
	lookup := (c.A&0x08)>>3 | (value&0x08)>>2 | (t&0x08)>>1
	c.insM(3)
	c.insM(5)
	c.ts += 8
	c.SetHL(c.HL() + dir)
	c.SetBC(c.BC() - 1)
	f := c.F&FlagC | halfcarrySub[lookup] | t&FlagS
	if c.BC() != 0 {
		f |= FlagV | FlagN
	} else {
		f |= FlagN
	}
	if t == 0 {
		f |= FlagZ
	}
	c.F = f
	if c.F&FlagH != 0 {
		t--
	}
	c.F |= t & Flag3
	if t&0x02 != 0 {
		c.F |= Flag5
	}
}

// ldBlock is the common part of LDI/LDD (dir +1/-1). It returns nothing; the
// order of BC-- and the write differs between LDI and LDIR, so callers pass
// decFirst.
func (c *CPU) ldBlock(dir uint16, decFirst bool) {
	v := c.rd(c.HL())
	c.insM(3)
	c.insM(5)
	c.ts += 8
	if decFirst {
		c.SetBC(c.BC() - 1)
		c.wr(c.DE(), v)
		c.SetDE(c.DE() + dir)
		c.SetHL(c.HL() + dir)
	} else {
		c.wr(c.DE(), v)
		c.SetHL(c.HL() + dir)
		c.SetDE(c.DE() + dir)
		c.SetBC(c.BC() - 1)
	}
	c.ldBlockFlags(v)
}

// inBlock is the common part of INI/IND/INIR/INDR.
func (c *CPU) inBlock(dir uint16) {
	v := c.in(c.BC())
	c.addM(1)
	c.insM(4)
	c.ioMCycle = c.mCycleIndex
	c.insM(3)
	c.ts += 2
	c.ts += 3
	c.ts += 3
	c.wr(c.HL(), v)
	c.B--
	c.SetHL(c.HL() + dir)
	f := sz53[c.B]
	if v&0x80 != 0 {
		f |= FlagN
	}
	c.F = f
}

// outBlockFlags sets flags for OUTI/OUTD/OTIR/OTDR.
func (c *CPU) outBlockFlags(v byte) {
	f := sz53[c.B]
	if v&0x80 != 0 {
		f |= FlagN
	}
	c.F = f
}

func (c *CPU) opED(op byte) {
	switch op {
	case 0x40, 0x48, 0x50, 0x58, 0x60, 0x68, 0x78: // IN r,(C)
		c.ts++
		c.setReg8(op>>3, c.inp(c.BC()))
	case 0x70: // IN F,(C) (value discarded)
		c.ts++
		c.inp(c.BC())
	case 0x41, 0x49, 0x51, 0x59, 0x61, 0x69, 0x79: // OUT (C),r
		c.ts++
		c.outp(c.BC(), c.reg8(op>>3))
	case 0x71: // OUT (C),0
		c.ts++
		c.outp(c.BC(), 0)
	case 0x42, 0x52, 0x62, 0x72: // SBC HL,rr
		c.ts += 7
		c.sbc16(c.rr16(op))
	case 0x4a, 0x5a, 0x6a, 0x7a: // ADC HL,rr
		c.ts += 7
		c.adc16(c.rr16(op))
	case 0x43: // LD (nn),BC
		c.ld16nnrr(c.C, c.B)
	case 0x53:
		c.ld16nnrr(c.E, c.D)
	case 0x63:
		c.ld16nnrr(c.L, c.H)
	case 0x73:
		c.ld16nnrr(byte(c.SP), byte(c.SP>>8))
	case 0x4b: // LD BC,(nn)
		c.SetBC(c.ld16rrnn())
	case 0x5b:
		c.SetDE(c.ld16rrnn())
	case 0x6b:
		c.SetHL(c.ld16rrnn())
	case 0x7b:
		c.SP = c.ld16rrnn()
		c.StackChange = 0
	case 0x44, 0x4c, 0x54, 0x5c, 0x64, 0x6c, 0x74, 0x7c: // NEG
		t := c.A
		c.A = 0
		c.sub(t)
	case 0x45, 0x4d, 0x55, 0x5d, 0x65, 0x6d, 0x75, 0x7d: // RETN / RETI
		c.IFF1 = c.IFF2
		c.ret()
	case 0x46, 0x4e, 0x66, 0x6e: // IM 0
		c.IM = 0
	case 0x56, 0x76: // IM 1
		c.IM = 1
	case 0x5e, 0x7e: // IM 2
		c.IM = 2
	case 0x47: // LD I,A
		c.ts++
		c.addM(1)
		c.I = c.A
	case 0x4f: // LD R,A
		c.ts++
		c.addM(1)
		c.R = uint16(c.A)
		c.R7 = c.A
	case 0x57: // LD A,I
		c.ts++
		c.addM(1)
		c.A = c.I
		c.F = c.F&FlagC | sz53[c.A] | c.iff2Flag()
	case 0x5f: // LD A,R
		c.ts++
		c.addM(1)
		c.A = byte(c.R&0x7f) | c.R7&0x80
		c.F = c.F&FlagC | sz53[c.A] | c.iff2Flag()
	case 0x67: // RRD
		v := c.rd(c.HL())
		c.insM(3)
		c.addM(4)
		c.insM(3)
		c.ts += 10
		c.wr(c.HL(), c.A<<4|v>>4)
		c.A = c.A&0xf0 | v&0x0f
		c.F = c.F&FlagC | sz53p[c.A]
	case 0x6f: // RLD
		v := c.rd(c.HL())
		c.insM(3)
		c.addM(4)
		c.insM(3)
		c.ts += 10
		c.wr(c.HL(), v<<4|c.A&0x0f)
		c.A = c.A&0xf0 | v>>4
		c.F = c.F&FlagC | sz53p[c.A]

	case 0xa0: // LDI
		c.ldBlock(1, true)
	case 0xa8: // LDD
		c.ldBlock(0xffff, true)
	case 0xb0: // LDIR
		c.ldBlock(1, false)
		c.repeat(c.BC() != 0)
	case 0xb8: // LDDR
		c.ldBlock(0xffff, false)
		c.repeat(c.BC() != 0)
	case 0xa1: // CPI
		c.cpBlock(1)
	case 0xa9: // CPD
		c.cpBlock(0xffff)
	case 0xb1: // CPIR
		c.cpBlock(1)
		c.repeat(c.F&(FlagV|FlagZ) == FlagV)
	case 0xb9: // CPDR
		c.cpBlock(0xffff)
		c.repeat(c.F&(FlagV|FlagZ) == FlagV)
	case 0xa2: // INI
		c.inBlock(1)
	case 0xaa: // IND
		c.inBlock(0xffff)
	case 0xb2: // INIR
		c.inBlock(1)
		c.repeat(c.B != 0)
	case 0xba: // INDR
		c.inBlock(0xffff)
		c.repeat(c.B != 0)
	case 0xa3, 0xab: // OUTI, OUTD
		v := c.rd(c.HL())
		c.addM(1)
		c.insM(3)
		c.insM(4)
		c.ioMCycle = c.mCycleIndex
		c.ts += 8
		c.B--
		c.out(c.BC(), v)
		if op == 0xa3 {
			c.SetHL(c.HL() + 1)
		} else {
			c.SetHL(c.HL() - 1)
		}
		c.outBlockFlags(v)
	case 0xb3, 0xbb: // OTIR, OTDR
		v := c.rd(c.HL())
		c.addM(1)
		c.insM(3)
		c.ts += 5
		c.B--
		c.out(c.BC(), v)
		if op == 0xb3 {
			c.SetHL(c.HL() + 1)
		} else {
			c.SetHL(c.HL() - 1)
		}
		c.outBlockFlags(v)
		c.insM(4)
		c.ioMCycle = c.mCycleIndex
		if c.B != 0 {
			c.addM(5)
			c.ts += 8
			c.PC -= 2
		} else {
			c.ts += 3
		}
	default:
		// All other ED opcodes (including the 0xfb .slt trap) act as NOPs.
	}
}

// repeat finishes a repeating block instruction.
func (c *CPU) repeat(again bool) {
	if again {
		c.addM(5)
		c.ts += 5
		c.PC -= 2
	}
}

func (c *CPU) iff2Flag() byte {
	if c.IFF2 != 0 {
		return FlagV
	}
	return 0
}

// rr16 returns BC, DE, HL or SP selected by bits 4-5 of op.
func (c *CPU) rr16(op byte) uint16 {
	switch (op >> 4) & 3 {
	case 0:
		return c.BC()
	case 1:
		return c.DE()
	case 2:
		return c.HL()
	default:
		return c.SP
	}
}
