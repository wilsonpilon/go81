// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
)

// TestTokensMatchROM reads the keyword table of the ROM (0111H): one entry
// per code, the last character with bit 7 set. Entry 0 is "?", then codes
// C0-FF, then RND, INKEY$ and PI.
func TestTokensMatchROM(t *testing.T) {
	rom, err := assets.ROM("zx81.edition3")
	if err != nil {
		t.Fatal(err)
	}
	var entries []string
	var sb strings.Builder
	for a := 0x0111; len(entries) < 1+64+3; a++ {
		b := rom[a]
		sb.WriteRune(zx81.TextChars[b&63])
		if b&0x80 != 0 {
			entries = append(entries, sb.String())
			sb.Reset()
		}
	}
	codes := []byte{}
	for c := 0xC0; c <= 0xFF; c++ {
		codes = append(codes, byte(c))
	}
	codes = append(codes, basic.TokRND, basic.TokINKEY, basic.TokPI)
	for i, c := range codes {
		want := entries[i+1]
		if c == basic.TokUnused {
			want = "" // "?" in the ROM: not a keyword
		}
		if basic.Tokens[c] != want {
			t.Errorf("token %02X = %q, ROM has %q", c, basic.Tokens[c], entries[i+1])
		}
	}
}

// line builds a program line.
func line(num int, body ...byte) []byte {
	n := len(body) + 1
	return append(append([]byte{byte(num >> 8), byte(num), byte(n), byte(n >> 8)}, body...), basic.Newline)
}

// zx converts plain text (letters, digits, simple symbols) to ZX81 codes.
func zx(s string) []byte {
	var b []byte
	for _, r := range s {
		for c, ch := range zx81.TextChars[:64] {
			if ch == r {
				b = append(b, byte(c))
				break
			}
		}
	}
	return b
}

// testLines covers every keyword and the special cases of the lister.
func testLines() [][]byte {
	var bodies [][]byte
	// Each keyword between letters, and after a space and a keyword.
	for c := 0xC0; c <= 0xFF; c++ {
		if basic.Tokens[byte(c)] == "" {
			continue
		}
		bodies = append(bodies, append(append(zx("A"), byte(c)), zx("B")...))
		bodies = append(bodies, []byte{byte(c), 0xF5, 0x26, byte(c), 0})
	}
	bodies = append(bodies,
		[]byte{0xF1, 0x26, 0x14, basic.TokRND, 0x17, 0x40, 0x41, 0x42},                                 // LET A=RND*RNDINKEY$PI
		append(append([]byte{0xF5, basic.Quote}, zx("OI")...), basic.TokQuote, 0x26, basic.Quote),      // PRINT "OI\"A"
		[]byte{0xF5, 1, 2, 3, 8, 9, 10, 0x80, 0x81, 0x88, 0x8A, 0xA6, 0x8B, 0x8C},                      // graphics, inverse
		append(append(zx("1.5"), basic.NumMarker, 0x81, 0x40, 0, 0, 0), zx("+2")...),                   // number
		[]byte{basic.TokREM, 0x26, 0xF5, 0x43, 0x00, basic.NumMarker, 1, 2, 3, 4, 5, 0x27, 0xC9, 0xFF}, // REM with code
		[]byte{0xF5, 0x43, 0x55, 0xC3, 0x26},                                                           // codes with no keyword
	)
	var lines [][]byte
	for i, b := range bodies {
		lines = append(lines, line(i+1, b...))
	}
	return lines
}

// romListing loads the lines into the emulated ZX81, runs LIST and returns
// the screen rows (the ROM's own listing).
func romListing(t *testing.T, empty []byte, lines [][]byte) []string {
	t.Helper()
	var prog []byte
	for _, l := range lines {
		prog = append(prog, l...)
	}
	// Insert the program into an empty .p and move the pointers that point
	// past the program area.
	start := basic.ProgStart - basic.SysVarBase
	p := append(append(append([]byte{}, empty[:start]...), prog...), empty[start:]...)
	for _, sv := range []int{0x400C, 0x400E, 0x4010, 0x4012, 0x4014, 0x4016, 0x4018, 0x401A, 0x401C, 0x4029} {
		o := sv - basic.SysVarBase
		v := int(p[o]) | int(p[o+1])<<8
		if v >= basic.ProgStart {
			v += len(prog)
			p[o], p[o+1] = byte(v), byte(v>>8)
		}
	}
	file := filepath.Join(t.TempDir(), "t.p")
	if err := os.WriteFile(file, p, 0o644); err != nil {
		t.Fatal(err)
	}
	s, _ := session.New(session.DefaultSettings())
	for _, c := range []string{"load " + file, "typeln K"} { // K = LIST
		if err := s.Exec(c, io.Discard); err != nil {
			t.Fatal(err)
		}
	}
	var rows []string
	s.Do(func(m *zx81.Machine) { rows = m.ScreenText() })
	return rows
}

func emptyP(t *testing.T) []byte {
	t.Helper()
	s, _ := session.New(session.DefaultSettings())
	file := filepath.Join(t.TempDir(), "vazio.p")
	if err := s.Exec("save "+file, io.Discard); err != nil {
		t.Fatal(err)
	}
	p, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestListMatchesROM compares the lister with the ZX81's own LIST.
func TestListMatchesROM(t *testing.T) {
	empty := emptyP(t)
	all := testLines()
	for start := 0; start < len(all); start += 20 { // one screen at a time
		chunk := all[start:min(start+20, len(all))]
		rows := romListing(t, empty, chunk)
		lines, err := basic.Lines(append([]byte{}, flatten(chunk)...))
		if err != nil {
			t.Fatal(err)
		}
		row := 0
		for _, l := range lines {
			// The ROM writes the number in 4 columns, then '>' or ' ', then
			// the body, wrapping at 32 columns.
			want := basic.Screen(l.Body)
			got := rows[row][5:]
			rest := []rune(want)
			if len(rest) > 27 {
				got += rows[row+1]
				row++
			}
			row++
			if strings.TrimRight(got, " ") != strings.TrimRight(want, " ") {
				t.Errorf("line %d (% X):\n ROM   %q\n lister %q", l.Num, l.Body, strings.TrimRight(got, " "), want)
			}
		}
	}
}

func flatten(ls [][]byte) []byte {
	var b []byte
	for _, l := range ls {
		b = append(b, l...)
	}
	return b
}

func TestText(t *testing.T) {
	cases := []struct {
		body []byte
		want string
	}{
		{[]byte{0xF5, basic.Quote, 0x2D, 0x2E, basic.TokQuote, basic.Quote}, `PRINT "HI\""`},
		{[]byte{0xF1, 0x26, 0x14, 0x1D, basic.NumMarker, 0x81, 0, 0, 0, 0}, "LET A=1"},
		{[]byte{0xF5, 1, 0x80, 0x8A, 0xA6, 12}, `PRINT \' \::\!!%A#`},
		{[]byte{basic.TokREM, 0x26, 0xF5, 0x43, basic.NumMarker, 1, 2, 3, 4, 5}, `REM A\{F5}\{43}\{7E}\{01}\{02}\{03}\{04}\{05}`},
		{[]byte{0xFA, 0x26, 0xDB, 0x1D, 0xDE, 0xEC, 0x1D, 0x1C}, "IF A<=1 THEN GOTO 10"},
		{[]byte{0xF1, 0x26, 0x14, 0x1E, 0xD8, 0x1F}, "LET A=2**3"},
		{[]byte{0xF5, basic.Quote, 0x26, 0xDB, 0x27, 0xD9, 0x00, basic.Quote}, `PRINT "A\{DB}B\{D9} "`},
	}
	for _, c := range cases {
		if got := basic.Text(c.body); got != c.want {
			t.Errorf("basic.Text(% X) = %q, want %q", c.body, got, c.want)
		}
	}
}

func TestListP(t *testing.T) {
	empty := emptyP(t)
	start := basic.ProgStart - basic.SysVarBase
	prog := append(line(10, 0xF5, basic.Quote, 0x2D, 0x2E, basic.Quote), line(20, 0xEC, 0x1D, 0x1C, basic.NumMarker, 0x84, 0x20, 0, 0, 0)...)
	p := append(append(append([]byte{}, empty[:start]...), prog...), empty[start:]...)
	d := int(p[3]) | int(p[4])<<8
	d += len(prog)
	p[3], p[4] = byte(d), byte(d>>8)
	got, err := basic.ListP(p)
	if err != nil {
		t.Fatal(err)
	}
	if got != "10 PRINT \"HI\"\n20 GOTO 10\n" {
		t.Fatalf("basic.ListP = %q", got)
	}
	if l, _ := basic.ListP(empty); l != "" {
		t.Fatalf("empty program listed %q", l)
	}
}
