// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne z80_ops.c — Copyright (c) 1999-2003 Philip Kendall,
// Witold Filipczyk; EightyOne modifications (c) Michael D Wynne
// (GPL v2 or later).

package z80

// reg8 returns register number r (B,C,D,E,H,L,-,A); r must not be 6.
func (c *CPU) reg8(r byte) byte {
	switch r & 7 {
	case 0:
		return c.B
	case 1:
		return c.C
	case 2:
		return c.D
	case 3:
		return c.E
	case 4:
		return c.H
	case 5:
		return c.L
	default:
		return c.A
	}
}

// setReg8 sets register number r (B,C,D,E,H,L,-,A); r must not be 6.
func (c *CPU) setReg8(r byte, v byte) {
	switch r & 7 {
	case 0:
		c.B = v
	case 1:
		c.C = v
	case 2:
		c.D = v
	case 3:
		c.E = v
	case 4:
		c.H = v
	case 5:
		c.L = v
	default:
		c.A = v
	}
}

// alu applies ALU operation op (ADD,ADC,SUB,SBC,AND,XOR,OR,CP) to v.
func (c *CPU) alu(op byte, v byte) {
	switch op & 7 {
	case 0:
		c.add(v)
	case 1:
		c.adc(v)
	case 2:
		c.sub(v)
	case 3:
		c.sbc(v)
	case 4:
		c.and(v)
	case 5:
		c.xor(v)
	case 6:
		c.or(v)
	default:
		c.cp(v)
	}
}

// cond evaluates condition cc (NZ,Z,NC,C,PO,PE,P,M).
func (c *CPU) cond(cc byte) bool {
	switch cc & 7 {
	case 0:
		return c.F&FlagZ == 0
	case 1:
		return c.F&FlagZ != 0
	case 2:
		return c.F&FlagC == 0
	case 3:
		return c.F&FlagC != 0
	case 4:
		return c.F&FlagP == 0
	case 5:
		return c.F&FlagP != 0
	case 6:
		return c.F&FlagS == 0
	default:
		return c.F&FlagS != 0
	}
}

// Step executes one instruction (or one prefix that falls back) and returns
// the T-states used.
func (c *CPU) Step() int {
	c.mCycleIndex = -1
	c.ioMCycle = -1
	c.ts = 0
	c.RetExecuted = false

	c.insM(4)
	c.ts += 4
	c.R++
	c.RZXCounter--
	c.m1Cycles = 1
	op := c.fetch()

	switch {
	case op >= 0x40 && op <= 0x7f:
		c.opLD(op)
	case op >= 0x80 && op <= 0xbf:
		if op&7 == 6 {
			c.insM(3)
			c.ts += 3
			c.alu(op>>3, c.rd(c.HL()))
		} else {
			c.alu(op>>3, c.reg8(op))
		}
	default:
		c.opMain(op)
	}

	c.R &= 127
	c.TStates = c.ts
	return c.ts
}

// opLD handles 0x40-0x7f: LD r,r', LD r,(HL), LD (HL),r and HALT.
func (c *CPU) opLD(op byte) {
	dst, src := (op>>3)&7, op&7
	switch {
	case op == 0x76: // HALT
		c.Halted = true
		c.PC--
	case src == 6: // LD r,(HL)
		c.insM(3)
		c.ts += 3
		c.setReg8(dst, c.rd(c.HL()))
	case dst == 6: // LD (HL),r
		c.insM(3)
		c.ts += 3
		c.wr(c.HL(), c.reg8(src))
	default:
		c.setReg8(dst, c.reg8(src))
	}
}

// imm reads an immediate byte operand (InsertMCycle 3, 3 T-states).
func (c *CPU) imm() byte {
	c.insM(3)
	c.ts += 3
	return c.fetch()
}

// imm16 reads an immediate word operand.
func (c *CPU) imm16() uint16 {
	lo := c.imm()
	hi := c.imm()
	return uint16(hi)<<8 | uint16(lo)
}

// jrCond is JR cc / JR: one operand cycle, optional jump, then PC++.
func (c *CPU) jrCond(take bool) {
	c.insM(3)
	c.ts += 3
	if take {
		c.jr()
	}
	c.PC++
}

func (c *CPU) jpCond(take bool) {
	c.insM(3)
	c.insM(3)
	c.ts += 6
	if take {
		c.jp()
	} else {
		c.PC += 2
	}
}

func (c *CPU) callCond(take bool) {
	c.insM(3)
	c.insM(3)
	c.ts += 6
	if take {
		c.call()
	} else {
		c.PC += 2
	}
}

func (c *CPU) retCond(take bool) {
	c.ts++
	c.addM(1)
	if take {
		c.ret()
	}
}

func (c *CPU) rst(addr uint16) {
	c.ts++
	c.addM(1)
	c.push(c.PC)
	c.PC = addr
}

func (c *CPU) incDec16Timing() {
	c.ts += 2
	c.addM(2)
}

func (c *CPU) opMain(op byte) {
	switch op {
	case 0x00: // NOP
	case 0x01: // LD BC,nn
		c.C = c.imm()
		c.B = c.imm()
	case 0x02: // LD (BC),A
		c.insM(3)
		c.ts += 3
		c.wr(c.BC(), c.A)
	case 0x03: // INC BC
		c.incDec16Timing()
		c.SetBC(c.BC() + 1)
	case 0x04:
		c.B = c.inc(c.B)
	case 0x05:
		c.B = c.dec(c.B)
	case 0x06:
		c.B = c.imm()
	case 0x07: // RLCA
		c.A = c.A<<1 | c.A>>7
		c.F = c.F&(FlagP|FlagZ|FlagS) | c.A&(FlagC|Flag3|Flag5)
	case 0x08: // EX AF,AF'
		t := c.AF()
		c.SetAF(c.AF_)
		c.AF_ = t
	case 0x09:
		c.SetHL(c.add16(c.HL(), c.BC()))
	case 0x0a: // LD A,(BC)
		c.insM(3)
		c.ts += 3
		c.A = c.rd(c.BC())
	case 0x0b:
		c.incDec16Timing()
		c.SetBC(c.BC() - 1)
	case 0x0c:
		c.C = c.inc(c.C)
	case 0x0d:
		c.C = c.dec(c.C)
	case 0x0e:
		c.C = c.imm()
	case 0x0f: // RRCA
		c.F = c.F&(FlagP|FlagZ|FlagS) | c.A&FlagC
		c.A = c.A>>1 | c.A<<7
		c.F |= c.A & (Flag3 | Flag5)
	case 0x10: // DJNZ
		c.ts++
		c.addM(1)
		c.insM(3)
		c.ts += 3
		c.B--
		if c.B != 0 {
			c.jr()
		}
		c.PC++
	case 0x11:
		c.E = c.imm()
		c.D = c.imm()
	case 0x12:
		c.insM(3)
		c.ts += 3
		c.wr(c.DE(), c.A)
	case 0x13:
		c.incDec16Timing()
		c.SetDE(c.DE() + 1)
	case 0x14:
		c.D = c.inc(c.D)
	case 0x15:
		c.D = c.dec(c.D)
	case 0x16:
		c.D = c.imm()
	case 0x17: // RLA
		t := c.A
		c.A = c.A<<1 | c.F&FlagC
		c.F = c.F&(FlagP|FlagZ|FlagS) | c.A&(Flag3|Flag5) | t>>7
	case 0x18:
		c.jrCond(true)
	case 0x19:
		c.SetHL(c.add16(c.HL(), c.DE()))
	case 0x1a:
		c.insM(3)
		c.ts += 3
		c.A = c.rd(c.DE())
	case 0x1b:
		c.incDec16Timing()
		c.SetDE(c.DE() - 1)
	case 0x1c:
		c.E = c.inc(c.E)
	case 0x1d:
		c.E = c.dec(c.E)
	case 0x1e:
		c.E = c.imm()
	case 0x1f: // RRA
		t := c.A
		c.A = c.A>>1 | c.F<<7
		c.F = c.F&(FlagP|FlagZ|FlagS) | c.A&(Flag3|Flag5) | t&FlagC
	case 0x20:
		c.jrCond(c.F&FlagZ == 0)
	case 0x21:
		c.L = c.imm()
		c.H = c.imm()
	case 0x22:
		c.ld16nnrr(c.L, c.H)
	case 0x23:
		c.incDec16Timing()
		c.SetHL(c.HL() + 1)
	case 0x24:
		c.H = c.inc(c.H)
	case 0x25:
		c.H = c.dec(c.H)
	case 0x26:
		c.H = c.imm()
	case 0x27: // DAA
		add := byte(0)
		carry := c.F & FlagC
		if c.F&FlagH != 0 || c.A&0x0f > 9 {
			add = 6
		}
		if carry != 0 || c.A > 0x9f {
			add |= 0x60
		}
		if c.A > 0x99 {
			carry = 1
		}
		if c.F&FlagN != 0 {
			c.sub(add)
		} else {
			if c.A > 0x90 && c.A&0x0f > 9 {
				add |= 0x60
			}
			c.add(add)
		}
		c.F = c.F&^(FlagC|FlagP) | carry | parity[c.A]
	case 0x28:
		c.jrCond(c.F&FlagZ != 0)
	case 0x29:
		c.SetHL(c.add16(c.HL(), c.HL()))
	case 0x2a:
		c.SetHL(c.ld16rrnn())
	case 0x2b:
		c.incDec16Timing()
		c.SetHL(c.HL() - 1)
	case 0x2c:
		c.L = c.inc(c.L)
	case 0x2d:
		c.L = c.dec(c.L)
	case 0x2e:
		c.L = c.imm()
	case 0x2f: // CPL
		c.A ^= 0xff
		c.F = c.F&(FlagC|FlagP|FlagZ|FlagS) | c.A&(Flag3|Flag5) | FlagN | FlagH
	case 0x30:
		c.jrCond(c.F&FlagC == 0)
	case 0x31: // LD SP,nn
		lo := c.imm()
		hi := c.imm()
		c.SP = uint16(hi)<<8 | uint16(lo)
		c.StackChange = 0
	case 0x32: // LD (nn),A
		a := c.imm16()
		c.insM(3)
		c.ts += 3
		c.wr(a, c.A)
	case 0x33:
		c.incDec16Timing()
		c.SP++
		c.StackChange--
	case 0x34: // INC (HL)
		c.insM(4)
		c.ts += 4
		v := c.inc(c.rd(c.HL()))
		c.insM(3)
		c.ts += 3
		c.wr(c.HL(), v)
	case 0x35: // DEC (HL)
		c.insM(4)
		c.ts += 4
		v := c.dec(c.rd(c.HL()))
		c.insM(3)
		c.ts += 3
		c.wr(c.HL(), v)
	case 0x36: // LD (HL),n
		c.insM(3)
		c.insM(3)
		c.ts += 6
		c.wr(c.HL(), c.fetch())
	case 0x37: // SCF
		c.F &^= FlagN | FlagH
		c.F |= c.A&(Flag3|Flag5) | FlagC
	case 0x38:
		c.jrCond(c.F&FlagC != 0)
	case 0x39:
		c.SetHL(c.add16(c.HL(), c.SP))
	case 0x3a: // LD A,(nn)
		a := c.imm16()
		c.insM(3)
		c.ts += 3
		c.A = c.rd(a)
	case 0x3b:
		c.incDec16Timing()
		c.SP--
		c.StackChange++
	case 0x3c:
		c.A = c.inc(c.A)
	case 0x3d:
		c.A = c.dec(c.A)
	case 0x3e:
		c.A = c.imm()
	case 0x3f: // CCF
		f := c.F&(FlagP|FlagZ|FlagS) | c.A&(Flag3|Flag5)
		if c.F&FlagC != 0 {
			f |= FlagH
		} else {
			f |= FlagC
		}
		c.F = f

	case 0xc0, 0xc8, 0xd0, 0xd8, 0xe0, 0xe8, 0xf0, 0xf8: // RET cc
		c.retCond(c.cond(op >> 3))
	case 0xc1:
		c.SetBC(c.pop())
	case 0xd1:
		c.SetDE(c.pop())
	case 0xe1:
		c.SetHL(c.pop())
	case 0xf1:
		c.SetAF(c.pop())
	case 0xc2, 0xca, 0xd2, 0xda, 0xe2, 0xea, 0xf2, 0xfa: // JP cc,nn
		c.jpCond(c.cond(op >> 3))
	case 0xc3:
		c.jpCond(true)
	case 0xc4, 0xcc, 0xd4, 0xdc, 0xe4, 0xec, 0xf4, 0xfc: // CALL cc,nn
		c.callCond(c.cond(op >> 3))
	case 0xcd:
		c.callCond(true)
	case 0xc5, 0xd5, 0xe5, 0xf5: // PUSH
		c.ts++
		c.addM(1)
		switch op {
		case 0xc5:
			c.push(c.BC())
		case 0xd5:
			c.push(c.DE())
		case 0xe5:
			c.push(c.HL())
		default:
			c.push(c.AF())
		}
	case 0xc6, 0xce, 0xd6, 0xde, 0xe6, 0xee, 0xf6, 0xfe: // ALU A,n
		c.alu(op>>3, c.imm())
	case 0xc7, 0xcf, 0xd7, 0xdf, 0xe7, 0xef, 0xf7, 0xff: // RST
		c.rst(uint16(op & 0x38))
	case 0xc9:
		c.ret()
	case 0xcb:
		c.insM(4)
		c.ts += 4
		c.m1Cycles++
		op2 := c.fetch()
		c.R++
		c.RZXCounter--
		c.opCB(op2)
	case 0xd3: // OUT (n),A
		c.insM(3)
		c.ts += 4
		port := uint16(c.fetch()) + uint16(c.A)<<8
		c.outp(port, c.A)
	case 0xd9: // EXX
		t := c.BC()
		c.SetBC(c.BC_)
		c.BC_ = t
		t = c.DE()
		c.SetDE(c.DE_)
		c.DE_ = t
		t = c.HL()
		c.SetHL(c.HL_)
		c.HL_ = t
	case 0xdb: // IN A,(n)
		c.insM(3)
		c.ts += 4
		port := uint16(c.fetch()) + uint16(c.A)<<8
		c.insM(4)
		c.ioMCycle = c.mCycleIndex
		c.ts += 3
		c.A = c.in(port)
	case 0xdd:
		c.insM(4)
		c.ts += 4
		c.m1Cycles++
		op2 := c.fetch()
		c.R++
		c.RZXCounter--
		c.opDDFD(op2, &c.IX)
	case 0xe3: // EX (SP),HL
		lo := c.rd(c.SP)
		hi := c.rd(c.SP + 1)
		c.insM(3)
		c.insM(4)
		c.insM(3)
		c.insM(5)
		c.ts += 15
		c.wr(c.SP, c.L)
		c.wr(c.SP+1, c.H)
		c.L, c.H = lo, hi
	case 0xe9: // JP (HL)
		c.PC = c.HL()
	case 0xeb: // EX DE,HL
		t := c.DE()
		c.SetDE(c.HL())
		c.SetHL(t)
	case 0xed:
		c.insM(4)
		c.ts += 4
		c.m1Cycles++
		op2 := c.fetch()
		c.R++
		c.RZXCounter--
		c.opED(op2)
	case 0xf3: // DI
		c.IFF1, c.IFF2 = 0, 0
	case 0xf9: // LD SP,HL
		c.incDec16Timing()
		c.SP = c.HL()
		c.StackChange = 0
	case 0xfb: // EI
		c.IFF1, c.IFF2 = 1, 1
	case 0xfd:
		// EightyOne does not count this M1 cycle for FD (it does for DD).
		c.insM(4)
		c.ts += 4
		op2 := c.fetch()
		c.R++
		c.RZXCounter--
		c.opDDFD(op2, &c.IY)
	}
}

// opCB handles the CB prefix.
func (c *CPU) opCB(op byte) {
	r := op & 7
	n := uint(op>>3) & 7
	switch op >> 6 {
	case 0: // rotates and shifts
		if r == 6 {
			v := c.rd(c.HL())
			c.insM(4)
			c.insM(3)
			c.ts += 7
			v = c.rot(op>>3, v)
			c.wr(c.HL(), v)
		} else {
			c.setReg8(r, c.rot(op>>3, c.reg8(r)))
		}
	case 1: // BIT
		var v byte
		if r == 6 {
			v = c.rd(c.HL())
			c.insM(4)
			c.ts += 4
		} else {
			v = c.reg8(r)
		}
		if n == 7 {
			c.bit7(v)
		} else {
			c.bit(n, v)
		}
	case 2: // RES
		if r == 6 {
			c.insM(4)
			c.insM(3)
			c.ts += 7
			c.wr(c.HL(), c.rd(c.HL())&^(1<<n))
		} else {
			c.setReg8(r, c.reg8(r)&^(1<<n))
		}
	default: // SET
		if r == 6 {
			c.insM(4)
			c.insM(3)
			c.ts += 7
			c.wr(c.HL(), c.rd(c.HL())|1<<n)
		} else {
			c.setReg8(r, c.reg8(r)|1<<n)
		}
	}
}
