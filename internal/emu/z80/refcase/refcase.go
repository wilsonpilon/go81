// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package refcase runs Z80 test cases in the same format as the C reference
// harness (test/compat/z80ref/harness.c), so the Go core can be compared
// with the original EightyOne core line by line.
package refcase

import (
	"fmt"
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/z80"
)

// Modes of a test case.
const (
	ModeStep = 0 // one instruction, Bytes placed at PC
	ModeINT  = 1 // maskable interrupt, data bus = Bytes[0], halted = Bytes[1]&1
	ModeNMI  = 2 // NMI, halted = Bytes[1]&1
	ModeRun  = 3 // run N instructions on random memory
)

// Case is one test case.
type Case struct {
	Seed  uint64
	Mode  int
	N     int
	Bytes [4]byte
}

// String returns the case as a harness input line.
func (c Case) String() string {
	return fmt.Sprintf("%d %d %d %d %d %d %d", c.Seed, c.Mode, c.N,
		c.Bytes[0], c.Bytes[1], c.Bytes[2], c.Bytes[3])
}

// Parse reads a harness input line.
func Parse(line string) (Case, error) {
	var c Case
	var b [4]int
	_, err := fmt.Sscanf(strings.TrimSpace(line), "%d %d %d %d %d %d %d",
		&c.Seed, &c.Mode, &c.N, &b[0], &b[1], &b[2], &b[3])
	for i := range b {
		c.Bytes[i] = byte(b[i])
	}
	return c, err
}

// Next is the splitmix64 generator shared with the C harness.
func Next(s *uint64) uint64 {
	*s += 0x9E3779B97F4A7C15
	z := *s
	z = (z ^ z>>30) * 0xBF58476D1CE4E5B9
	z = (z ^ z>>27) * 0x94D049BB133111EB
	return z ^ z>>31
}

const (
	fnvOffset = 0xCBF29CE484222325
	fnvPrime  = 0x100000001B3
)

// bus is the test bus; it logs every access into a hash.
type bus struct {
	mem       [65536]byte
	portState uint64
	logHash   uint64
	cpu       *z80.CPU
	verbose   *strings.Builder
}

func (b *bus) log(kind int, addr uint16, v byte, ts int) {
	raw := [8]byte{byte(kind), byte(addr), byte(addr >> 8), v,
		byte(ts), byte(ts >> 8), byte(ts >> 16), byte(ts >> 24)}
	for _, x := range raw {
		b.logHash ^= uint64(x)
		b.logHash *= fnvPrime
	}
	if b.verbose != nil {
		fmt.Fprintf(b.verbose, "  %c %04x %02x ts=%d\n", "FORWIO"[kind], addr, v, ts)
	}
}

func (b *bus) OpcodeFetch(a uint16) byte {
	b.log(0, a, b.mem[a], b.cpu.CurrentTStates())
	return b.mem[a]
}

func (b *bus) ReadOperand(a uint16) byte {
	b.log(1, a, b.mem[a], b.cpu.CurrentTStates())
	return b.mem[a]
}

func (b *bus) Read(a uint16) byte {
	b.log(2, a, b.mem[a], b.cpu.CurrentTStates())
	return b.mem[a]
}

func (b *bus) Write(a uint16, v byte) {
	b.log(3, a, v, b.cpu.CurrentTStates())
	b.mem[a] = v
}

func (b *bus) In(port uint16, ts *int) byte {
	s := b.portState ^ uint64(port)<<32 ^ uint64(uint32(*ts))
	v := byte(Next(&s))
	b.log(4, port, v, *ts)
	if port&0x0300 == 0x0300 {
		*ts++
	}
	return v
}

func (b *bus) Out(port uint16, v byte, ts *int) {
	b.log(5, port, v, *ts)
	if port&0x0300 == 0x0300 {
		*ts += 2
	}
}

func (b *bus) setup(seed uint64) {
	st := seed
	c := b.cpu
	for i := 0; i < 65536; i += 8 {
		v := Next(&st)
		for j := 0; j < 8; j++ {
			b.mem[i+j] = byte(v >> (8 * j))
		}
	}
	v := Next(&st)
	c.A, c.F, c.B, c.C = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	c.D, c.E, c.H, c.L = byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56)
	v = Next(&st)
	c.AF_, c.BC_, c.DE_, c.HL_ = uint16(v), uint16(v>>16), uint16(v>>32), uint16(v>>48)
	v = Next(&st)
	c.IX, c.IY, c.SP, c.PC = uint16(v), uint16(v>>16), uint16(v>>32), uint16(v>>48)
	v = Next(&st)
	c.I = byte(v)
	c.R = uint16(v>>8) & 0x7f
	c.R7 = byte(v>>16) & 0x80
	c.IFF1 = byte(v>>24) & 1
	c.IFF2 = byte(v>>25) & 1
	c.IM = byte(v>>26) % 3
	c.Halted = false
	b.portState = Next(&st)
	c.ResetDebugState()
	b.logHash = fnvOffset
}

func b2i(v bool) int {
	if v {
		return 1
	}
	return 0
}

func stateHash(c *z80.CPU) uint64 {
	v := [20]uint32{
		uint32(c.AF()), uint32(c.BC()), uint32(c.DE()), uint32(c.HL()),
		uint32(c.AF_), uint32(c.BC_), uint32(c.DE_), uint32(c.HL_),
		uint32(c.IX), uint32(c.IY), uint32(c.SP), uint32(c.PC),
		uint32(c.I), uint32(c.R), uint32(c.R7),
		uint32(c.IFF1) | uint32(c.IFF2)<<1 | uint32(c.IM)<<2 | uint32(b2i(c.Halted))<<4,
		uint32(c.TStates), uint32(c.StackChange), uint32(c.M1Cycles()), uint32(b2i(c.RetExecuted)),
	}
	h := uint64(fnvOffset)
	for _, x := range v {
		h ^= uint64(x)
		h *= fnvPrime
	}
	return h
}

// Runner runs cases on one reusable CPU.
type Runner struct {
	b   *bus
	cpu *z80.CPU
}

// NewRunner creates a runner.
func NewRunner() *Runner {
	b := &bus{}
	cpu := z80.New(b)
	b.cpu = cpu
	return &Runner{b: b, cpu: cpu}
}

// Run executes a case and returns the output line (without newline). When
// verbose is true the bus log is returned too.
func (r *Runner) Run(tc Case, verbose bool) (line string, log string) {
	b, c := r.b, r.cpu
	var sb strings.Builder
	b.verbose = nil
	if verbose {
		b.verbose = &sb
	}
	b.setup(tc.Seed)
	ts := 0
	switch tc.Mode {
	case ModeStep:
		for i := 0; i < 4; i++ {
			b.mem[c.PC+uint16(i)] = tc.Bytes[i]
		}
		ts = c.Step()
	case ModeINT:
		c.Halted = tc.Bytes[1]&1 != 0
		ts = c.Interrupt(tc.Bytes[0])
	case ModeNMI:
		c.Halted = tc.Bytes[1]&1 != 0
		ts = c.NMI()
	case ModeRun:
		acc := uint64(fnvOffset)
		for i := 0; i < tc.N; i++ {
			ts += c.Step()
			acc ^= stateHash(c)
			acc *= fnvPrime
		}
		b.logHash ^= acc
	}
	return format(c, tc.Mode, ts, b.logHash), sb.String()
}

func format(c *z80.CPU, mode, ts int, logHash uint64) string {
	var s strings.Builder
	fmt.Fprintf(&s, "%04x %04x %04x %04x %04x %04x %04x %04x %04x %04x %04x %04x "+
		"i=%02x r=%04x r7=%02x iff=%d%d im=%d h=%d ts=%d sc=%d ret=%d",
		c.AF(), c.BC(), c.DE(), c.HL(), c.AF_, c.BC_, c.DE_, c.HL_,
		c.IX, c.IY, c.SP, c.PC, c.I, c.R, c.R7, c.IFF1, c.IFF2, c.IM,
		b2i(c.Halted), ts, c.StackChange, b2i(c.RetExecuted))
	fmt.Fprintf(&s, " m1=%d", c.M1Cycles())
	if mode == ModeStep {
		s.WriteString(" mc=")
		for i, m := range c.MCycles() {
			if m == 0 {
				break
			}
			if i > 0 {
				s.WriteByte(',')
			}
			fmt.Fprintf(&s, "%d", m)
		}
		fmt.Fprintf(&s, " io=%d", c.IOMCycle())
	}
	fmt.Fprintf(&s, " log=%016x", logHash)
	return s.String()
}

// Cases returns the standard case set: every opcode (all prefixes) with
// perOp random states, interrupts, NMIs and runs of random code.
func Cases(seed uint64, perOp, runs, runLen int) []Case {
	st := seed
	var out []Case
	add := func(b0, b1, b2, b3 byte) {
		for k := 0; k < perOp; k++ {
			out = append(out, Case{Seed: Next(&st), Mode: ModeStep, Bytes: [4]byte{b0, b1, b2, b3}})
		}
	}
	rnd := func() byte { return byte(Next(&st)) }
	for op := 0; op < 256; op++ {
		add(byte(op), rnd(), rnd(), rnd())
		add(0xcb, byte(op), rnd(), rnd())
		add(0xed, byte(op), rnd(), rnd())
		add(0xdd, byte(op), rnd(), rnd())
		add(0xfd, byte(op), rnd(), rnd())
		add(0xdd, 0xcb, rnd(), byte(op))
		add(0xfd, 0xcb, rnd(), byte(op))
	}
	for k := 0; k < perOp*8; k++ {
		out = append(out, Case{Seed: Next(&st), Mode: ModeINT, Bytes: [4]byte{rnd(), rnd()}})
		out = append(out, Case{Seed: Next(&st), Mode: ModeNMI, Bytes: [4]byte{0, rnd()}})
	}
	for k := 0; k < runs; k++ {
		out = append(out, Case{Seed: Next(&st), Mode: ModeRun, N: runLen})
	}
	return out
}
