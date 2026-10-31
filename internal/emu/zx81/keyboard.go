// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package zx81

import (
	"fmt"
	"strings"
)

// Key is a position in the ZX81 keyboard matrix.
type Key struct{ Row, Bit int }

// Keys maps key names to their matrix position. Row n is read when address
// line A(8+n) is low.
var Keys = map[string]Key{
	"SHIFT": {0, 0}, "Z": {0, 1}, "X": {0, 2}, "C": {0, 3}, "V": {0, 4},
	"A": {1, 0}, "S": {1, 1}, "D": {1, 2}, "F": {1, 3}, "G": {1, 4},
	"Q": {2, 0}, "W": {2, 1}, "E": {2, 2}, "R": {2, 3}, "T": {2, 4},
	"1": {3, 0}, "2": {3, 1}, "3": {3, 2}, "4": {3, 3}, "5": {3, 4},
	"0": {4, 0}, "9": {4, 1}, "8": {4, 2}, "7": {4, 3}, "6": {4, 4},
	"P": {5, 0}, "O": {5, 1}, "I": {5, 2}, "U": {5, 3}, "Y": {5, 4},
	"ENTER": {6, 0}, "L": {6, 1}, "K": {6, 2}, "J": {6, 3}, "H": {6, 4},
	"SPACE": {7, 0}, ".": {7, 1}, "M": {7, 2}, "N": {7, 3}, "B": {7, 4},
}

// ParseChord parses keys joined by "+" (e.g. "SHIFT+P") into row bits.
func ParseChord(chord string) ([8]byte, error) {
	var rows [8]byte
	for _, k := range strings.Split(strings.ToUpper(strings.TrimSpace(chord)), "+") {
		p, ok := Keys[k]
		if !ok {
			return rows, fmt.Errorf("unknown ZX81 key %q", k)
		}
		rows[p.Row] |= 1 << p.Bit
	}
	return rows, nil
}

// SetKeys sets the whole keyboard state (all keys up, then rows pressed).
func (m *Machine) SetKeys(rows [8]byte) { m.Keyboard = rows }
