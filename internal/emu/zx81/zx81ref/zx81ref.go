// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package zx81ref runs ZX81 scenarios and prints one line per frame in the
// same format as the C reference harness (test/compat/zx81ref/harness.cpp),
// so the Go machine can be compared with EightyOne's zx81.cpp.
package zx81ref

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// KeyEvent sets keyboard row Row to Bits from frame Frame on.
type KeyEvent struct{ Frame, Row, Bits int }

// Scenario is one comparison run.
type Scenario struct {
	Name   string
	Frames int
	WRX    bool
	RAMK   int // RAM pack in K (0 = 1K internal only)
	NTSC   bool
	Keys   []KeyEvent
}

// Args returns the harness command-line options.
func (s Scenario) Args() []string {
	b := func(v bool) int {
		if v {
			return 1
		}
		return 0
	}
	return []string{fmt.Sprint(s.Frames), fmt.Sprintf("wrx=%d", b(s.WRX)),
		fmt.Sprintf("ram=%d", s.RAMK), fmt.Sprintf("ntsc=%d", b(s.NTSC))}
}

// KeyInput returns the harness stdin.
func (s Scenario) KeyInput() string {
	var sb strings.Builder
	for _, e := range s.Keys {
		fmt.Fprintf(&sb, "%d %d %d\n", e.Frame, e.Row, e.Bits)
	}
	return sb.String()
}

// Type builds key events that press each chord in turn, starting at frame
// start. A chord is keys joined by "+", e.g. "SHIFT+P". Each chord is held
// for hold frames and followed by gap frames with no key down; ENTER gets
// enterGap extra frames so the ROM can redraw the screen.
func Type(start, hold, gap, enterGap int, chords ...string) ([]KeyEvent, int) {
	var ev []KeyEvent
	f := start
	for _, ch := range chords {
		rows, err := zx81.ParseChord(ch)
		if err != nil {
			panic("zx81ref: " + err.Error())
		}
		for r, bits := range rows {
			if bits != 0 {
				ev = append(ev, KeyEvent{f, r, int(bits)})
			}
		}
		f += hold
		for r, bits := range rows {
			if bits != 0 {
				ev = append(ev, KeyEvent{f, r, 0})
			}
		}
		f += gap
		if ch == "ENTER" {
			f += enterGap
		}
	}
	return ev, f
}

// Scenarios returns the standard comparison set.
func Scenarios() []Scenario {
	// 10 PRINT "HELLO"  /  20 GOTO 10  /  RUN  (fills the screen, stops with 5/10)
	prog, end := Type(150, 6, 6, 60,
		"1", "0", "P", "SHIFT+P", "H", "E", "L", "L", "O", "SHIFT+P", "ENTER",
		"2", "0", "G", "1", "0", "ENTER",
		"R", "ENTER")
	// FAST, then SLOW: exercises the NMI generator being switched off and on.
	fast, end2 := Type(end+100, 6, 6, 60, "SHIFT+F", "ENTER")
	slow, _ := Type(end2, 6, 6, 60, "SHIFT+D", "ENTER")
	keys := append(append(prog, fast...), slow...)

	return []Scenario{
		{Name: "boot-16k", Frames: 300, RAMK: 16},
		{Name: "boot-1k", Frames: 300, RAMK: 0},
		{Name: "boot-ntsc", Frames: 300, RAMK: 16, NTSC: true},
		{Name: "boot-wrx", Frames: 200, RAMK: 16, WRX: true},
		{Name: "hello-16k", Frames: end2 + 300, RAMK: 16, Keys: keys},
		{Name: "hello-1k", Frames: end2 + 300, RAMK: 0, Keys: keys},
		{Name: "hello-ntsc", Frames: end2 + 300, RAMK: 16, NTSC: true, Keys: keys},
	}
}

const (
	fnvOffset = 0xCBF29CE484222325
	fnvPrime  = 0x100000001B3
)

func fnv(h uint64, p []byte) uint64 {
	for _, b := range p {
		h ^= uint64(b)
		h *= fnvPrime
	}
	return h
}

// Run runs a scenario on the Go machine and returns the output lines.
func Run(rom []byte, s Scenario) ([]string, error) {
	cfg := zx81.DefaultConfig(rom)
	cfg.RAMTop = zx81.RAMTopForPack(s.RAMK)
	if s.WRX {
		cfg.HiRes = zx81.HiResWRX
	}
	cfg.NTSC = s.NTSC
	m, err := zx81.New(cfg)
	if err != nil {
		return nil, err
	}

	var lineHash uint64
	m.LineHook = func(l *zx81.Scanline) {
		lineHash = fnv(lineHash, l.Pixels[:l.Len])
		var meta [12]byte
		binary.LittleEndian.PutUint32(meta[0:], uint32(l.Len))
		binary.LittleEndian.PutUint32(meta[4:], uint32(l.SyncLen))
		binary.LittleEndian.PutUint32(meta[8:], uint32(l.SyncType))
		lineHash = fnv(lineHash, meta[:])
	}

	var out []string
	var dispHash uint64
	frames := 0
	ev := 0
	for f := 0; f < s.Frames; f++ {
		for ev < len(s.Keys) && s.Keys[ev].Frame <= f {
			m.Keyboard[s.Keys[ev].Row&7] = byte(s.Keys[ev].Bits)
			ev++
		}
		lineHash = fnvOffset
		m.RunFrame()
		if n := m.FramesCompleted(); n != frames {
			frames = n
			d := m.Display()
			dispHash = fnv(fnvOffset, d[:])
		}
		c := m.CPU
		out = append(out, fmt.Sprintf("f=%d pc=%04x af=%04x bc=%04x de=%04x hl=%04x ix=%04x iy=%04x sp=%04x i=%02x r=%02x ts=%d "+
			"line=%016x disp=%016x mem=%016x ry=%d",
			f, c.PC, c.AF(), c.BC(), c.DE(), c.HL(), c.IX, c.IY, c.SP, c.I, c.RegR(), m.FrameTStates(),
			lineHash, dispHash, fnv(fnvOffset, m.Memory[:]), m.RasterY()))
	}
	return out, nil
}
