// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tui

import (
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

// Cell is one character of the ZX81 text screen.
type Cell = zx81.TextCell

// Display rows and columns.
const (
	Rows = zx81.TextRows
	Cols = zx81.TextCols
)

// TextScreen decodes the display file into 24 rows of 32 characters.
func TextScreen(m *zx81.Machine) [Rows][Cols]Cell { return m.TextScreen() }

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
