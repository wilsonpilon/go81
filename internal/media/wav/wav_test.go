// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package wav

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "som.wav")
	w, err := Create(path, 44100, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Write([]int16{1, -1, 300, -300}); err != nil {
		t.Fatal(err)
	}
	if err := w.Sync(); err != nil { // valid even before Close
		t.Fatal(err)
	}
	check := func(samples int) []byte {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		le := binary.LittleEndian
		if string(b[0:4]) != "RIFF" || string(b[8:16]) != "WAVEfmt " || string(b[36:40]) != "data" {
			t.Fatalf("header %q", b[:44])
		}
		if le.Uint16(b[22:]) != 2 || le.Uint32(b[24:]) != 44100 || le.Uint32(b[28:]) != 44100*4 || le.Uint16(b[34:]) != 16 {
			t.Fatal("fmt chunk")
		}
		if int(le.Uint32(b[40:])) != samples*2 || int(le.Uint32(b[4:])) != 36+samples*2 || len(b) != 44+samples*2 {
			t.Fatalf("sizes: data %d riff %d file %d", le.Uint32(b[40:]), le.Uint32(b[4:]), len(b))
		}
		return b
	}
	b := check(4)
	if int16(binary.LittleEndian.Uint16(b[46:])) != -1 {
		t.Fatal("sample")
	}
	if err := w.Write(make([]int16, 44100*2)); err != nil {
		t.Fatal(err)
	}
	if s := w.Seconds(); s < 1 || s > 1.001 {
		t.Fatalf("seconds %v", s)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	check(4 + 44100*2)
	if _, err := Create(path, 0, 2); err == nil {
		t.Fatal("rate 0 accepted")
	}
}
