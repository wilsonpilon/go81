// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"fmt"
	"sort"
	"strings"
)

// Program utilities: they work on a list of lines (see Lines) and return a
// new list; the lines stay in order of number.

// MaxLine is the highest line number the ZX81 accepts.
const MaxLine = 9999

// DeleteLines removes the lines from..to (both included) and returns the
// remaining ones and how many were removed.
func DeleteLines(lines []Line, from, to int) ([]Line, int) {
	var out []Line
	for _, l := range lines {
		if l.Num < from || l.Num > to {
			out = append(out, l)
		}
	}
	return out, len(lines) - len(out)
}

// FindLines returns the lines whose listing contains the text (upper and
// lower case are the same, as the ZX81 has only capitals).
func FindLines(lines []Line, text string) []Line {
	want := strings.ToUpper(text)
	var out []Line
	for _, l := range lines {
		if strings.Contains(strings.ToUpper(Text(l.Body)), want) {
			out = append(out, l)
		}
	}
	return out
}

// PutLine adds a line, or replaces the line with the same number.
func PutLine(lines []Line, l Line) []Line {
	i := sort.Search(len(lines), func(i int) bool { return lines[i].Num >= l.Num })
	if i < len(lines) && lines[i].Num == l.Num {
		out := append([]Line(nil), lines...)
		out[i] = l
		return out
	}
	out := make([]Line, 0, len(lines)+1)
	out = append(out, lines[:i]...)
	out = append(out, l)
	return append(out, lines[i:]...)
}

// RemLine is a line "num REM" followed by size bytes of fill, the usual
// place for machine code on the ZX81. The fill cannot be NEWLINE (118),
// which would end the line for the ROM's LIST.
func RemLine(num, size int, fill byte) (Line, error) {
	if num < 1 || num > MaxLine {
		return Line{}, fmt.Errorf("número de linha fora de 1-%d: %d", MaxLine, num)
	}
	if size < 1 || size > 16000 {
		return Line{}, fmt.Errorf("tamanho fora de 1-16000: %d", size)
	}
	if fill == Newline {
		return Line{}, fmt.Errorf("o byte %d (NEWLINE) não pode encher uma linha REM", Newline)
	}
	body := make([]byte, size+1)
	body[0] = TokREM
	for i := 1; i <= size; i++ {
		body[i] = fill
	}
	return Line{Num: num, Body: body}, nil
}

// LineAddress is the address of the line num in memory once the program is
// loaded (its first byte: the high byte of the number), or -1.
func LineAddress(lines []Line, num int) int {
	a := ProgStart
	for _, l := range lines {
		if l.Num == num {
			return a
		}
		a += 4 + len(l.Body) + 1
	}
	return -1
}
