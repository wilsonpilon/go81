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
	"github.com/wilsonpilon/go81/internal/media/tape"
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
	Sound  bool // EightyOne "Video Sound"
	ZX80   bool // a ZX80 with its ROM (zx80_do_scanline)
	Keys   []KeyEvent

	// Tape, if set, is inserted before the first frame (file name decides
	// the ZX81 program name). AutoLoad types LOAD "" as EightyOne does.
	TapeName string
	TapeData []byte
	AutoLoad bool
}

// Args returns the harness command-line options; tapePath is where the
// caller wrote TapeData.
func (s Scenario) Args(tapePath string) []string {
	b := func(v bool) int {
		if v {
			return 1
		}
		return 0
	}
	args := []string{fmt.Sprint(s.Frames), fmt.Sprintf("wrx=%d", b(s.WRX)),
		fmt.Sprintf("ram=%d", s.RAMK), fmt.Sprintf("ntsc=%d", b(s.NTSC))}
	if s.Sound {
		args = append(args, "sound=1")
	}
	if s.ZX80 {
		args = append(args, "zx80=1")
	}
	if s.TapeData != nil {
		args = append(args, "tape="+tapePath, fmt.Sprintf("autoload=%d", b(s.AutoLoad)))
	}
	return args
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

// MakeP types a BASIC program on a fresh 16K ZX81, then SAVE "T", and
// returns what the ROM saved (flash save, as EightyOne records it) without
// the name: a .p file.
func MakeP(rom []byte, chords ...string) ([]byte, error) {
	m, err := zx81.New(zx81.DefaultConfig(rom))
	if err != nil {
		return nil, err
	}
	m.Tape = tape.NewDeck(3250000)
	m.Tape.Out = true
	var saved []byte
	m.SaveByte = func(b byte) { saved = append(saved, b) }

	all := append(append([]string{}, chords...), "S", "SHIFT+P", "T", "SHIFT+P", "ENTER")
	keys, end := Type(150, 6, 6, 60, all...)
	ev := 0
	for f := 0; f < end+100; f++ {
		for ev < len(keys) && keys[ev].Frame <= f {
			m.Keyboard[keys[ev].Row&7] = byte(keys[ev].Bits)
			ev++
		}
		m.RunFrame()
	}
	for i, b := range saved {
		if b&128 != 0 { // last character of the name
			return saved[i+1:], nil
		}
	}
	return nil, fmt.Errorf("zx81ref: SAVE produced no data (%d bytes)", len(saved))
}

// Test programs (typed, so they are our own).
var (
	progOla = []string{"1", "0", "P", "SHIFT+P", "O", "L", "A", "SHIFT+P", "ENTER",
		"2", "0", "G", "1", "0", "ENTER"}
	progPlot = []string{"1", "0", "F", "I", "SHIFT+L", "0", "SHIFT+4", "4", "0", "ENTER",
		"2", "0", "Q", "I", "SHIFT+.", "I", "ENTER",
		"3", "0", "N", "I", "ENTER"}
)

// TapeScenarios returns scenarios that load programs from tape with
// EightyOne's flash load and auto-load, then RUN them.
func TapeScenarios(rom []byte) ([]Scenario, error) {
	ola, err := MakeP(rom, progOla...)
	if err != nil {
		return nil, err
	}
	plot, err := MakeP(rom, progPlot...)
	if err != nil {
		return nil, err
	}
	run, _ := Type(450, 6, 6, 60, "R", "ENTER")
	// Tapes of several blocks, written by Go81 and read by EightyOne's own
	// LoadFile: OLA, a pause, then PLOT; auto-load takes the first.
	deck := tape.NewPlayer(3250000)
	deck.NewTape()
	deck.Blocks[deck.AddGeneral(append(tape.ZX81Name("OLA"), ola...))].Pause = 5000
	deck.Blocks = append(deck.Blocks, tape.Block{ID: tape.BlockPause, Pause: 1000})
	deck.Blocks[deck.AddGeneral(append(tape.ZX81Name("PLOT"), plot...))].Pause = 5000
	tzx := tape.SaveTZX(deck.Blocks)
	t81 := tape.SaveT81(deck.Blocks)
	// LOAD "PLOT" (J, SHIFT+P, P L O T, SHIFT+P): the ROM skips OLA on the tape.
	loadPlot, end := Type(150, 6, 6, 60, "J", "SHIFT+P", "P", "L", "O", "T", "SHIFT+P", "ENTER")
	runPlot, _ := Type(end+900, 6, 6, 60, "R", "ENTER")
	return []Scenario{
		{Name: "load-tzx-16k", Frames: 700, RAMK: 16, TapeName: "fita.tzx", TapeData: tzx, AutoLoad: true, Keys: run},
		{Name: "load-t81-16k", Frames: 700, RAMK: 16, TapeName: "fita.t81", TapeData: t81, AutoLoad: true, Keys: run},
		{Name: "load-tzx-by-name", Frames: end + 1300, RAMK: 16, TapeName: "fita.tzx", TapeData: tzx, Keys: append(loadPlot, runPlot...)},
		{Name: "load-ola-16k", Frames: 700, RAMK: 16, TapeName: "OLA.p", TapeData: ola, AutoLoad: true, Keys: run},
		{Name: "load-ola-ntsc", Frames: 900, RAMK: 16, NTSC: true, TapeName: "OLA.p", TapeData: ola, AutoLoad: true, Keys: run},
		{Name: "load-plot-16k", Frames: 800, RAMK: 16, TapeName: "Plot_Test.p", TapeData: plot, AutoLoad: true, Keys: run},
		{Name: "tape-no-load", Frames: 300, RAMK: 16, TapeName: "OLA.p", TapeData: ola},
	}, nil
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
		{Name: "sound-boot", Frames: 300, RAMK: 16, Sound: true},
		{Name: "sound-hello", Frames: end2 + 300, RAMK: 16, Sound: true, Keys: keys},
		{Name: "sound-ntsc", Frames: end2 + 300, RAMK: 16, NTSC: true, Sound: true, Keys: keys},
	}
}

// MakeO types a program on a ZX80 and saves it (SAVE is the E key on the
// ZX80, with no name); it returns the bytes saved: a .o file.
func MakeO(rom []byte, chords ...string) ([]byte, error) {
	cfg := zx81.DefaultConfig(rom)
	cfg.ZX80 = true
	m, err := zx81.New(cfg)
	if err != nil {
		return nil, err
	}
	m.Tape = tape.NewDeck(3250000)
	m.Tape.Out = true
	var saved []byte
	m.SaveByte = func(b byte) { saved = append(saved, b) }
	keys, end := Type(150, 6, 6, 60, append(append([]string{}, chords...), "E", "ENTER")...)
	ev := 0
	for f := 0; f < end+200; f++ {
		for ev < len(keys) && keys[ev].Frame <= f {
			m.Keyboard[keys[ev].Row&7] = byte(keys[ev].Bits)
			ev++
		}
		m.RunFrame()
	}
	if len(saved) < 0x28 {
		return nil, fmt.Errorf("zx81ref: the ZX80 saved only %d bytes", len(saved))
	}
	return saved, nil
}

// ZX80TapeScenarios load a program saved by the ZX80 back with the
// auto-load (LOAD is W, then ENTER) and run it.
func ZX80TapeScenarios(rom []byte) ([]Scenario, error) {
	o, err := MakeO(rom, "1", "0", "O", "SHIFT+Y", "H", "I", "SHIFT+Y", "ENTER")
	if err != nil {
		return nil, err
	}
	run, _ := Type(500, 6, 6, 60, "R", "ENTER")
	return []Scenario{
		{Name: "zx80-load", Frames: 800, RAMK: 16, ZX80: true, TapeName: "HI.o", TapeData: o, AutoLoad: true, Keys: run},
	}, nil
}

// ZX80Scenarios are the ZX80 runs (they need the ZX80 ROM). The keys type
// a short program line and run it; on the ZX80 the picture goes away while
// a key is held and while a program runs.
func ZX80Scenarios() []Scenario {
	keys, end := Type(150, 6, 6, 60,
		"1", "0", "O", "SHIFT+Y", "H", "I", "SHIFT+Y", "ENTER",
		"R", "ENTER")
	return []Scenario{
		{Name: "zx80-boot", Frames: 300, RAMK: 16, ZX80: true},
		{Name: "zx80-1k", Frames: 300, RAMK: 0, ZX80: true},
		{Name: "zx80-ntsc", Frames: 300, RAMK: 16, NTSC: true, ZX80: true},
		{Name: "zx80-keys", Frames: end + 200, RAMK: 16, ZX80: true, Keys: keys},
		{Name: "zx80-sound", Frames: end + 200, RAMK: 16, ZX80: true, Sound: true, Keys: keys},
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

func audioHash(a []int16) uint64 {
	b := make([]byte, 2*len(a))
	for i, v := range a {
		binary.LittleEndian.PutUint16(b[2*i:], uint16(v))
	}
	return fnv(fnvOffset, b)
}

// Run runs a scenario on the Go machine and returns the output lines.
func Run(rom []byte, s Scenario) ([]string, error) {
	cfg := zx81.DefaultConfig(rom)
	cfg.RAMTop = zx81.RAMTopForPack(s.RAMK)
	if s.WRX {
		cfg.HiRes = zx81.HiResWRX
	}
	cfg.NTSC = s.NTSC
	cfg.VSyncSound = s.Sound
	cfg.ZX80 = s.ZX80
	m, err := zx81.New(cfg)
	if err != nil {
		return nil, err
	}
	if s.TapeData != nil && tape.IsTapeFile(s.TapeName) {
		blocks, err := tape.LoadTape(s.TapeData)
		if err != nil {
			return nil, err
		}
		m.InsertTapeBlocks(blocks, s.AutoLoad)
	} else if s.TapeData != nil {
		block, err := tape.ProgramBlock(s.TapeName, s.TapeData)
		if err != nil {
			return nil, err
		}
		m.InsertTape(block, s.AutoLoad)
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
		playing, cur := 0, 0
		if m.Tape != nil {
			cur = m.Tape.Player.CurBlock
			if m.Tape.Player.Playing {
				playing = 1
			}
		}
		out = append(out, fmt.Sprintf("f=%d pc=%04x af=%04x bc=%04x de=%04x hl=%04x ix=%04x iy=%04x sp=%04x i=%02x r=%02x ts=%d "+
			"line=%016x disp=%016x mem=%016x ry=%d tape=%d/%d snd=%016x",
			f, c.PC, c.AF(), c.BC(), c.DE(), c.HL(), c.IX, c.IY, c.SP, c.I, c.RegR(), m.FrameTStates(),
			lineHash, dispHash, fnv(fnvOffset, m.Memory[:]), m.RasterY(), playing, cur, audioHash(m.Audio())))
	}
	return out, nil
}
