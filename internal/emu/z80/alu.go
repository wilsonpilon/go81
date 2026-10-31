// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne z80_macros.h — Copyright (c) 1999-2001 Philip Kendall
// (GPL v2 or later).

package z80

func (c *CPU) and(v byte) {
	c.A &= v
	c.F = FlagH | sz53p[c.A]
}

func (c *CPU) or(v byte) {
	c.A |= v
	c.F = sz53p[c.A]
}

func (c *CPU) xor(v byte) {
	c.A ^= v
	c.F = sz53p[c.A]
}

func (c *CPU) add(v byte) {
	t := uint16(c.A) + uint16(v)
	lookup := (c.A&0x88)>>3 | (v&0x88)>>2 | byte((t&0x88)>>1)
	c.A = byte(t)
	c.F = carry8(t) | halfcarryAdd[lookup&7] | overflowAdd[lookup>>4] | sz53[c.A]
}

func (c *CPU) adc(v byte) {
	t := uint16(c.A) + uint16(v) + uint16(c.F&FlagC)
	lookup := (c.A&0x88)>>3 | (v&0x88)>>2 | byte((t&0x88)>>1)
	c.A = byte(t)
	c.F = carry8(t) | halfcarryAdd[lookup&7] | overflowAdd[lookup>>4] | sz53[c.A]
}

func (c *CPU) sub(v byte) {
	t := uint16(c.A) - uint16(v)
	lookup := (c.A&0x88)>>3 | (v&0x88)>>2 | byte((t&0x88)>>1)
	c.A = byte(t)
	c.F = carry8(t) | FlagN | halfcarrySub[lookup&7] | overflowSub[lookup>>4] | sz53[c.A]
}

func (c *CPU) sbc(v byte) {
	t := uint16(c.A) - uint16(v) - uint16(c.F&FlagC)
	lookup := (c.A&0x88)>>3 | (v&0x88)>>2 | byte((t&0x88)>>1)
	c.A = byte(t)
	c.F = carry8(t) | FlagN | halfcarrySub[lookup&7] | overflowSub[lookup>>4] | sz53[c.A]
}

func (c *CPU) cp(v byte) {
	t := uint16(c.A) - uint16(v)
	lookup := (c.A&0x88)>>3 | (v&0x88)>>2 | byte((t&0x88)>>1)
	f := byte(0)
	if t&0x100 != 0 {
		f = FlagC
	} else if t == 0 {
		f = FlagZ
	}
	c.F = f | FlagN | halfcarrySub[lookup&7] | overflowSub[lookup>>4] | v&(Flag3|Flag5) | byte(t)&FlagS
}

func carry8(t uint16) byte {
	if t&0x100 != 0 {
		return FlagC
	}
	return 0
}

func (c *CPU) inc(v byte) byte {
	v++
	f := c.F & FlagC
	if v == 0x80 {
		f |= FlagV
	}
	if v&0x0f == 0 {
		f |= FlagH
	}
	c.F = f | sz53[v]
	return v
}

func (c *CPU) dec(v byte) byte {
	f := c.F&FlagC | FlagN
	if v&0x0f == 0 {
		f |= FlagH
	}
	v--
	if v == 0x7f {
		f |= FlagV
	}
	c.F = f | sz53[v]
	return v
}

// add16 is ADD16: adds b to a, 7 extra T-states.
func (c *CPU) add16(a, b uint16) uint16 {
	t := uint32(a) + uint32(b)
	lookup := byte((a&0x0800)>>11 | (b&0x0800)>>10 | uint16((t&0x0800)>>9))
	c.ts += 7
	c.addM(4)
	c.addM(3)
	f := c.F & (FlagV | FlagZ | FlagS)
	if t&0x10000 != 0 {
		f |= FlagC
	}
	c.F = f | byte(t>>8)&(Flag3|Flag5) | halfcarryAdd[lookup]
	return uint16(t)
}

// adc16 is ADC16 on HL (T-states are added by the caller).
func (c *CPU) adc16(v uint16) {
	hl := c.HL()
	t := uint32(hl) + uint32(v) + uint32(c.F&FlagC)
	lookup := byte((hl&0x8800)>>11 | (v&0x8800)>>10 | uint16((t&0x8800)>>9))
	c.SetHL(uint16(t))
	f := overflowAdd[lookup>>4] | c.H&(Flag3|Flag5|FlagS) | halfcarryAdd[lookup&7]
	if t&0x10000 != 0 {
		f |= FlagC
	}
	if c.HL() == 0 {
		f |= FlagZ
	}
	c.F = f
	c.addM(4)
	c.addM(3)
}

// sbc16 is SBC16 on HL (T-states are added by the caller).
func (c *CPU) sbc16(v uint16) {
	hl := c.HL()
	t := uint32(hl) - uint32(v) - uint32(c.F&FlagC)
	lookup := byte((hl&0x8800)>>11 | (v&0x8800)>>10 | uint16((t&0x8800)>>9))
	c.SetHL(uint16(t))
	f := FlagN | overflowSub[lookup>>4] | c.H&(Flag3|Flag5|FlagS) | halfcarrySub[lookup&7]
	if t&0x10000 != 0 {
		f |= FlagC
	}
	if c.HL() == 0 {
		f |= FlagZ
	}
	c.F = f
	c.addM(4)
	c.addM(3)
}

func (c *CPU) bit(n uint, v byte) {
	f := c.F&FlagC | v&(Flag3|Flag5)
	if v&(1<<n) != 0 {
		f |= FlagH
	} else {
		f |= FlagP | FlagH | FlagZ
	}
	c.F = f
}

func (c *CPU) bit7(v byte) {
	f := c.F&FlagC | v&(Flag3|Flag5)
	if v&0x80 != 0 {
		f |= FlagH | FlagS
	} else {
		f |= FlagP | FlagH | FlagZ
	}
	c.F = f
}

// Rotates and shifts used by the CB and DDFDCB tables.

func (c *CPU) rlc(v byte) byte {
	v = v<<1 | v>>7
	c.F = v&FlagC | sz53p[v]
	return v
}

func (c *CPU) rrc(v byte) byte {
	f := v & FlagC
	v = v>>1 | v<<7
	c.F = f | sz53p[v]
	return v
}

func (c *CPU) rl(v byte) byte {
	t := v
	v = v<<1 | c.F&FlagC
	c.F = t>>7 | sz53p[v]
	return v
}

func (c *CPU) rr(v byte) byte {
	t := v
	v = v>>1 | c.F<<7
	c.F = t&FlagC | sz53p[v]
	return v
}

func (c *CPU) sla(v byte) byte {
	f := v >> 7
	v <<= 1
	c.F = f | sz53p[v]
	return v
}

func (c *CPU) sra(v byte) byte {
	f := v & FlagC
	v = v&0x80 | v>>1
	c.F = f | sz53p[v]
	return v
}

func (c *CPU) sll(v byte) byte {
	f := v >> 7
	v = v<<1 | 0x01
	c.F = f | sz53p[v]
	return v
}

func (c *CPU) srl(v byte) byte {
	f := v & FlagC
	v >>= 1
	c.F = f | sz53p[v]
	return v
}

// rot applies CB rotate/shift number op (0..7) to v.
func (c *CPU) rot(op byte, v byte) byte {
	switch op & 7 {
	case 0:
		return c.rlc(v)
	case 1:
		return c.rrc(v)
	case 2:
		return c.rl(v)
	case 3:
		return c.rr(v)
	case 4:
		return c.sla(v)
	case 5:
		return c.sra(v)
	case 6:
		return c.sll(v)
	default:
		return c.srl(v)
	}
}

// Control-flow and stack macros.

// push is PUSH16.
func (c *CPU) push(v uint16) {
	c.insM(3)
	c.SP--
	c.ts += 3
	c.wr(c.SP, byte(v>>8))
	c.insM(3)
	c.SP--
	c.ts += 3
	c.wr(c.SP, byte(v))
	c.StackChange += 2
}

// pop is POP16.
func (c *CPU) pop() uint16 {
	c.insM(3)
	c.ts += 3
	lo := c.rd(c.SP)
	c.SP++
	c.insM(3)
	c.ts += 3
	hi := c.rd(c.SP)
	c.SP++
	c.StackChange -= 2
	if c.StackChange < 0 {
		c.StackChange = 0
	}
	return uint16(hi)<<8 | uint16(lo)
}

// ret is RET().
func (c *CPU) ret() {
	c.insM(3)
	c.ts += 3
	lo := c.rd(c.SP)
	c.SP++
	c.insM(3)
	c.ts += 3
	hi := c.rd(c.SP)
	c.SP++
	c.PC = uint16(hi)<<8 | uint16(lo)
	c.StackChange -= 2
	c.RetExecuted = true
}

// jp is JP(): operands read through ReadOperand without moving PC first.
func (c *CPU) jp() {
	lo := c.bus.ReadOperand(c.PC)
	hi := c.bus.ReadOperand(c.PC + 1)
	c.PC = uint16(hi)<<8 | uint16(lo)
}

// jr is JR(); the caller then does PC++.
func (c *CPU) jr() {
	c.addM(5)
	c.ts += 5
	c.PC += uint16(int8(c.bus.ReadOperand(c.PC)))
}

// call is CALL().
func (c *CPU) call() {
	lo := c.bus.ReadOperand(c.PC)
	c.PC++
	c.addM(1)
	c.ts++
	hi := c.bus.ReadOperand(c.PC)
	c.PC++
	c.push(c.PC)
	c.PC = uint16(hi)<<8 | uint16(lo)
}

// ld16nnrr is LD16_NNRR: store lo,hi at the address that follows the opcode.
func (c *CPU) ld16nnrr(lo, hi byte) {
	c.insM(3)
	c.ts += 3
	a := uint16(c.operand())
	c.insM(3)
	c.ts += 3
	a |= uint16(c.operand()) << 8
	c.insM(3)
	c.ts += 3
	c.wr(a, lo)
	a++
	c.insM(3)
	c.ts += 3
	c.wr(a, hi)
}

// ld16rrnn is LD16_RRNN: load a word from the address that follows the opcode.
func (c *CPU) ld16rrnn() uint16 {
	c.insM(3)
	c.ts += 3
	a := uint16(c.operand())
	c.insM(3)
	c.ts += 3
	a |= uint16(c.operand()) << 8
	c.insM(3)
	c.ts += 3
	lo := c.rd(a)
	a++
	c.insM(3)
	c.ts += 3
	hi := c.rd(a)
	return uint16(hi)<<8 | uint16(lo)
}

// inp is the IN(reg,port) macro; returns the value read.
func (c *CPU) inp(port uint16) byte {
	c.insM(4)
	c.ioMCycle = c.mCycleIndex
	c.ts += 3
	v := c.in(port)
	c.F = c.F&FlagC | sz53p[v]
	return v
}

// outp is the OUT(port,reg) macro.
func (c *CPU) outp(port uint16, v byte) {
	c.insM(4)
	c.ioMCycle = c.mCycleIndex
	c.ts += 3
	c.out(port, v)
}
