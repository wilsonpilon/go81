// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic_test

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
)

// floatNumbers are typed into the emulated ZX81; the ROM's own conversion
// is the reference.
var floatNumbers = []string{
	"0", "1", "2", "10", "255", "256", "1000", "9999", "32767", "65535", "65536", "123456789",
	"0.5", "0.1", "0.2", "0.3", "1.5", "3.14159", "2.718281828", "0.001", "99.99", "1E10",
	"1E-5", "1.5E3", "1E38", "7E-39", ".5", "16384", "4E9", "0.333333333",
}

// romTyped types "LET A=<number>" lines into an emulated ZX81 (ten per
// session, so the listing fits on the screen) and returns the five bytes
// the ROM stored for each number.
func romTyped(t *testing.T, numbers []string) map[string][5]byte {
	t.Helper()
	out := map[string][5]byte{}
	for start := 0; start < len(numbers); start += 10 {
		chunk := numbers[start:min(start+10, len(numbers))]
		s, err := session.New(session.DefaultSettings())
		if err != nil {
			t.Fatal(err)
		}
		for i, n := range chunk {
			// L = LET at the K cursor; E inside a number is the letter E.
			for _, c := range []string{fmt.Sprintf("typeln %dLA=%s", i+1, n), "run 100"} {
				if err := s.Exec(c, io.Discard); err != nil {
					t.Fatal(err)
				}
			}
		}
		var prog []byte
		s.Do(func(m *zx81.Machine) { prog = basic.ProgramFromMemory(m) })
		lines, err := basic.Lines(prog)
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range lines {
			k := strings.IndexByte(string(l.Body), basic.NumMarker)
			if k < 0 || k+6 > len(l.Body) {
				t.Fatalf("line %d: no number in % X", l.Num, l.Body)
			}
			var b [5]byte
			copy(b[:], l.Body[k+1:k+6])
			out[chunk[l.Num-1]] = b
		}
		for _, n := range chunk {
			if _, ok := out[n]; !ok {
				t.Errorf("the ROM did not accept %s", n)
			}
		}
	}
	return out
}

// TestROMFloatMatchesTyping: the loader's conversion (the ROM routine run
// directly) gives the same bytes as typing the number into the ZX81.
func TestROMFloatMatchesTyping(t *testing.T) {
	typed := romTyped(t, floatNumbers)
	for _, n := range floatNumbers {
		want, ok := typed[n]
		if !ok {
			continue
		}
		got, err := basic.ROMFloatText(n)
		if err != nil {
			t.Errorf("%s: %v", n, err)
			continue
		}
		if got != want {
			t.Errorf("%-12s typed % X  ROMFloat % X", n, want, got)
		}
		// Float (exact rounding) agrees on integers.
		if v, _ := strconv.ParseFloat(n, 64); v == float64(int64(v)) && v < 1e9 {
			if f, _ := basic.Float(v); f != want {
				t.Errorf("Float(%s) = % X, ROM % X", n, f, want)
			}
		}
	}
	if _, err := basic.ROMFloatText("1E40"); err == nil {
		t.Error("1E40 should be too big")
	}
	if b, err := basic.ROMFloatText("2"); err != nil || b != [5]byte{0x82, 0, 0, 0, 0} {
		t.Errorf("after an error the calculator must still work: % X %v", b, err)
	}
}

func TestFloatValue(t *testing.T) {
	for _, v := range []float64{1, -1, 0.5, 255, -65536, 3.25} {
		b, _ := basic.Float(v)
		if got := basic.FloatValue(b); got != v {
			t.Errorf("FloatValue(Float(%g)) = %g", v, got)
		}
	}
	if _, err := basic.Float(1e39); err == nil {
		t.Error("1E39 should not fit")
	}
}
