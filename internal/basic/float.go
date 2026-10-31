// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"errors"
	"math"
)

// ErrNumberTooBig is a number the ZX81 cannot store (about 1.7E38 or more).
var ErrNumberTooBig = errors.New("número grande demais para o ZX81")

// Float encodes a number in the ZX81's 5-byte floating point format:
// exponent+128, then a 32-bit mantissa (0.5 <= m < 1) whose top bit is
// replaced by the sign. Zero is five zero bytes. Numbers too small to store
// become zero, as in the ROM.
func Float(v float64) ([5]byte, error) {
	var b [5]byte
	if v == 0 || math.IsNaN(v) {
		return b, nil
	}
	if math.IsInf(v, 0) {
		return b, ErrNumberTooBig
	}
	neg := v < 0
	m, e := math.Frexp(math.Abs(v)) // v = m * 2^e, 0.5 <= m < 1
	mant := uint64(math.Round(m * (1 << 32)))
	if mant == 1<<32 { // rounding carried into a new bit
		mant >>= 1
		e++
	}
	exp := e + 128
	if exp > 255 {
		return b, ErrNumberTooBig
	}
	if exp < 1 {
		return b, nil
	}
	mant &= 0x7FFFFFFF
	if neg {
		mant |= 0x80000000
	}
	b[0] = byte(exp)
	b[1], b[2], b[3], b[4] = byte(mant>>24), byte(mant>>16), byte(mant>>8), byte(mant)
	return b, nil
}

// FloatValue decodes the 5-byte format.
func FloatValue(b [5]byte) float64 {
	if b[0] == 0 {
		return 0
	}
	mant := uint32(b[1])<<24 | uint32(b[2])<<16 | uint32(b[3])<<8 | uint32(b[4])
	neg := mant&0x80000000 != 0
	mant |= 0x80000000
	v := math.Ldexp(float64(mant)/(1<<32), int(b[0])-128)
	if neg {
		v = -v
	}
	return v
}
