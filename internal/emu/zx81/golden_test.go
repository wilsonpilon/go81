// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package zx81_test

import (
	"bufio"
	"compress/gzip"
	"os"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81/zx81ref"
)

// TestEightyOneGolden replays frame-by-frame results recorded from
// EightyOne's own zx81.cpp (see test/compat/zx81diff): CPU state, T-states,
// every scanline, the TV picture and memory, for each standard scenario.
func TestEightyOneGolden(t *testing.T) {
	rom, err := assets.ROM("zx81.edition3")
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("testdata/eightyone.txt.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{}
	sc := bufio.NewScanner(zr)
	for sc.Scan() {
		name, line, _ := strings.Cut(sc.Text(), "\t")
		want[name] = append(want[name], line)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	for _, s := range zx81ref.Scenarios() {
		t.Run(s.Name, func(t *testing.T) {
			w := want[s.Name]
			if len(w) == 0 {
				t.Fatalf("no golden data for %s", s.Name)
			}
			got, err := zx81ref.Run(rom, s)
			if err != nil {
				t.Fatal(err)
			}
			for i := range w {
				if i >= len(got) || got[i] != w[i] {
					g := "<missing>"
					if i < len(got) {
						g = got[i]
					}
					t.Fatalf("frame %d differs from EightyOne\nwant %s\ngot  %s", i, w[i], g)
				}
			}
		})
	}
}
