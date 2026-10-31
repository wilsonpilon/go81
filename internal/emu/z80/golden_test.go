// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package z80_test

import (
	"bufio"
	"compress/gzip"
	"os"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/emu/z80/refcase"
)

// TestEightyOneGolden replays results recorded from the original EightyOne
// Z80 core (see test/compat/z80diff) and checks the Go core matches them:
// registers, flags, T-states, M-cycles and every bus access.
func TestEightyOneGolden(t *testing.T) {
	f, err := os.Open("testdata/eightyone.txt.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(zr)
	r := refcase.NewRunner()
	n, bad := 0, 0
	for sc.Scan() {
		in, want, ok := strings.Cut(sc.Text(), "\t")
		if !ok {
			t.Fatalf("bad golden line: %q", sc.Text())
		}
		tc, err := refcase.Parse(in)
		if err != nil {
			t.Fatal(err)
		}
		n++
		if got, _ := r.Run(tc, false); got != want {
			bad++
			if bad <= 5 {
				t.Errorf("case %s\nwant %s\ngot  %s", in, want, got)
			}
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if n < 7000 {
		t.Fatalf("only %d golden cases read", n)
	}
	if bad > 0 {
		t.Fatalf("%d of %d cases differ from EightyOne", bad, n)
	}
}
