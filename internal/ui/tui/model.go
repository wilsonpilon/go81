// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tui

import (
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
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
