// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne z80_ddfd.c and z80_ddfdcb.c — Copyright (c)
// 1999-2003 Philip Kendall (GPL v2 or later).

package z80

// idxReg8 is like reg8 but H and L are replaced by the index register halves.
func (c *CPU) idxReg8(r byte, x *uint16) byte {
	switch r & 7 {
	case 4:
		return byte(*x >> 8)
	case 5:
		return byte(*x)
	default:
		return c.reg8(r)
	}
}

func setHi(x *uint16, v byte) { *x = *x&0x00ff | uint16(v)<<8 }
func setLo(x *uint16, v byte) { *x = *x&0xff00 | uint16(v) }

// idxAddr reads the displacement and returns IX/IY+d (wrapped to 16 bits).
func (c *CPU) idxAddr(x *uint16) uint16 {
	return *x + uint16(int8(c.fetch()))
}

// idxPrefix is the common timing of "op (IX+d)" instructions.
func (c *CPU) idxPrefix() {
	c.ts += 11
	c.insM(3)
	c.addM(5)
	c.insM(3)
}

// opDDFD handles the DD and FD prefixes; x points to IX or IY.
func (c *CPU) opDDFD(op byte, x *uint16) {
	switch op {
	case 0x09:
		*x = c.add16(*x, c.BC())
	case 0x19:
		*x = c.add16(*x, c.DE())
	case 0x29:
		*x = c.add16(*x, *x)
	case 0x39:
		*x = c.add16(*x, c.SP)
	case 0x21: // LD IX,nn
		setLo(x, c.imm())
		setHi(x, c.imm())
	case 0x22: // LD (nn),IX
		c.ld16nnrr(byte(*x), byte(*x>>8))
	case 0x2a: // LD IX,(nn)
		*x = c.ld16rrnn()
	case 0x23:
		c.incDec16Timing()
		*x++
	case 0x2b:
		c.incDec16Timing()
		*x--
	case 0x24:
		setHi(x, c.inc(byte(*x>>8)))
	case 0x25:
		setHi(x, c.dec(byte(*x>>8)))
	case 0x26:
		setHi(x, c.imm())
	case 0x2c:
		setLo(x, c.inc(byte(*x)))
	case 0x2d:
		setLo(x, c.dec(byte(*x)))
	case 0x2e:
		setLo(x, c.imm())
	case 0x34, 0x35: // INC/DEC (IX+d)
		c.ts += 15
		c.insM(3)
		c.addM(5)
		c.insM(4)
		c.insM(3)
		a := c.idxAddr(x)
		v := c.rd(a)
		if op == 0x34 {
			v = c.inc(v)
		} else {
			v = c.dec(v)
		}
		c.wr(a, v)
	case 0x36: // LD (IX+d),n
		c.ts += 11
		c.insM(3)
		c.addM(5)
		c.insM(3)
		a := c.idxAddr(x)
		c.wr(a, c.fetch())

	// LD r,IXh / LD r,IXl / LD IXh,r ...
	case 0x44, 0x45, 0x4c, 0x4d, 0x54, 0x55, 0x5c, 0x5d, 0x7c, 0x7d:
		c.setReg8(op>>3, c.idxReg8(op, x))
	case 0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x67:
		setHi(x, c.idxReg8(op, x))
	case 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6f:
		setLo(x, c.idxReg8(op, x))

	case 0x46, 0x4e, 0x56, 0x5e, 0x66, 0x6e, 0x7e: // LD r,(IX+d)
		c.idxPrefix()
		c.setReg8(op>>3, c.rd(c.idxAddr(x)))
	case 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x77: // LD (IX+d),r
		c.idxPrefix()
		a := c.idxAddr(x)
		c.wr(a, c.reg8(op))

	case 0x84, 0x85, 0x8c, 0x8d, 0x94, 0x95, 0x9c, 0x9d,
		0xa4, 0xa5, 0xac, 0xad, 0xb4, 0xb5, 0xbc, 0xbd: // ALU A,IXh/IXl
		c.alu(op>>3, c.idxReg8(op, x))
	case 0x96: // SUB (IX+d): EightyOne omits the first InsertMCycle here.
		c.ts += 11
		c.addM(5)
		c.insM(3)
		c.sub(c.rd(c.idxAddr(x)))
	case 0x86, 0x8e, 0x9e, 0xa6, 0xae, 0xb6, 0xbe: // ALU A,(IX+d)
		c.idxPrefix()
		c.alu(op>>3, c.rd(c.idxAddr(x)))

	case 0xcb:
		c.insM(3)
		c.ts += 3
		a := c.idxAddr(x)
		c.addM(5)
		c.ts += 4
		op3 := c.fetch()
		c.opDDFDCB(op3, a)

	case 0xe1: // POP IX
		*x = c.pop()
	case 0xe3: // EX (SP),IX
		lo := c.rd(c.SP)
		hi := c.rd(c.SP + 1)
		c.insM(3)
		c.insM(4)
		c.ts += 7
		c.wr(c.SP, byte(*x))
		c.wr(c.SP+1, byte(*x>>8))
		c.insM(3)
		c.insM(5)
		c.ts += 8
		*x = uint16(hi)<<8 | uint16(lo)
	case 0xe5: // PUSH IX
		c.ts++
		c.addM(1)
		c.push(*x)
	case 0xe9: // JP (IX)
		c.PC = *x
	case 0xf9: // LD SP,IX
		c.incDec16Timing()
		c.SP = *x
		c.StackChange = 0
	default:
		// Not an index instruction: back up and run it as a plain opcode
		// on the next Step.
		c.PC--
		c.R--
	}
}

// opDDFDCB handles DD CB d op / FD CB d op with the address already computed.
func (c *CPU) opDDFDCB(op byte, a uint16) {
	r := op & 7
	n := uint(op>>3) & 7
	switch op >> 6 {
	case 0: // rotate/shift, result also copied to r (undocumented)
		c.ts += 8
		c.insM(4)
		c.insM(3)
		v := c.rot(op>>3, c.rd(a))
		if r != 6 {
			c.setReg8(r, v)
		}
		c.wr(a, v)
	case 1: // BIT n,(IX+d)
		c.ts += 5
		c.insM(4)
		v := c.rd(a)
		if n == 7 {
			c.bit7(v)
		} else {
			c.bit(n, v)
		}
	case 2: // RES
		c.ts += 8
		c.insM(4)
		c.insM(3)
		v := c.rd(a) &^ (1 << n)
		if r != 6 {
			c.setReg8(r, v)
		}
		c.wr(a, v)
	default: // SET
		c.ts += 8
		c.insM(4)
		c.insM(3)
		v := c.rd(a) | 1<<n
		if r != 6 {
			c.setReg8(r, v)
		}
		c.wr(a, v)
	}
}
