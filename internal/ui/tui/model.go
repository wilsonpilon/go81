// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tui

import (
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

// zxChars maps ZX81 character codes 0-63 to Unicode. Codes 1-10 are the
// block graphics, which are exactly the Unicode quadrant characters (plus
// three grey patterns).
var zxChars = []rune(" ▘▝▀▖▌▞▛▒▄▀\"£$:?()><=+-*/;,.0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Cell is one character of the ZX81 text screen.
type Cell struct {
	R       rune
	Inverse bool
	Grey    bool // codes 8-10: chequered patterns
}

// Display rows and columns.
const (
	Rows = 24
	Cols = 32
)

// TextScreen decodes the display file (D_FILE) into 24 rows of 32
// characters. It works for the normal text display (collapsed or full),
// not for hi-res pictures.
func TextScreen(m *zx81.Machine) [Rows][Cols]Cell {
	var s [Rows][Cols]Cell
	for r := range s {
		for c := range s[r] {
			s[r][c] = Cell{R: ' '}
		}
	}
	dfile := uint16(m.Peek(0x400C)) | uint16(m.Peek(0x400D))<<8
	a := dfile
	if m.Peek(a) != 0x76 {
		return s
	}
	a++
	for r := 0; r < Rows; r++ {
		for c := 0; ; c++ {
			b := m.Peek(a)
			a++
			if b == 0x76 || c > Cols {
				break
			}
			if c == Cols {
				continue // malformed line: skip until NEWLINE
			}
			code := int(b & 63)
			valid := b&64 == 0
			cell := Cell{R: '?', Inverse: b&128 != 0}
			if valid {
				cell.R = zxChars[code]
				cell.Grey = code >= 8 && code <= 10
			}
			s[r][c] = cell
		}
	}
	return s
}

// String renders a text screen as plain text (inverse characters are not
// marked). Useful for tests and logs.
func (s *Screen) String() string {
	var sb strings.Builder
	for _, row := range s.Text {
		for _, c := range row {
			sb.WriteRune(c.R)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Screen is what the terminal shows.
type Screen struct {
	Text [Rows][Cols]Cell
}

// Key timing: terminals only report key presses, so each key is held for a
// few frames and released; ENTER waits longer while the ROM redraws.
const (
	keyHold  = 5
	keyGap   = 5
	enterGap = 60
)

// model is the emulator state shared by the terminal UI and the tests.
type model struct {
	m      *zx81.Machine
	cfg    zx81.Config
	queue  [][8]byte
	enters []bool
	hold   int
	gap    int
	paused bool
	turbo  bool
	pixels bool
}

func newModel(cfg zx81.Config) (*model, error) {
	m, err := zx81.New(cfg)
	if err != nil {
		return nil, err
	}
	return &model{m: m, cfg: cfg}, nil
}

func (md *model) insert(name string, data []byte) error {
	block, err := tape.ProgramBlock(name, data)
	if err != nil {
		return err
	}
	md.m.InsertTape(block, true)
	return nil
}

// typeChord queues a ZX81 key chord (e.g. "SHIFT+P").
func (md *model) typeChord(chord string) error {
	rows, err := zx81.ParseChord(chord)
	if err != nil {
		return err
	}
	md.queue = append(md.queue, rows)
	md.enters = append(md.enters, chord == "ENTER")
	return nil
}

// frame runs one emulated frame, pressing and releasing queued keys.
func (md *model) frame() {
	switch {
	case md.hold > 0:
		md.hold--
		if md.hold == 0 {
			md.m.SetKeys([8]byte{})
		}
	case md.gap > 0:
		md.gap--
	case len(md.queue) > 0:
		md.m.SetKeys(md.queue[0])
		md.hold = keyHold
		md.gap = keyGap
		if md.enters[0] {
			md.gap += enterGap
		}
		md.queue, md.enters = md.queue[1:], md.enters[1:]
	}
	md.m.RunFrame()
}
