// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tape

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"
)

// twoPrograms is a tape with two programs and a pause between them; the
// programs end the way a ZX81 .p does, so the reader knows where they stop
// (E_LINE in the system variables).
func twoPrograms() []Block {
	p := NewPlayer(3250000)
	p.NewTape()
	for _, name := range []string{"UM", "DOIS"} {
		i := p.AddGeneral(append(ZX81Name(name), fakeP(40+len(name))...))
		p.Blocks[i].Pause = 2000
	}
	return p.Blocks
}

// fakeP is a small .p: the system variables from VERSN (4009h) with E_LINE
// pointing n bytes after them, then n bytes.
func fakeP(n int) []byte {
	p := make([]byte, 0x407D-0x4009+n)
	eline := 0x4009 + len(p) // the ROM saves up to E_LINE
	p[0x4014-0x4009] = byte(eline)
	p[0x4015-0x4009] = byte(eline >> 8)
	for i := 0x407D - 0x4009; i < len(p); i++ {
		p[i] = byte(i * 7)
	}
	return p
}

func programs(blocks []Block) [][]byte {
	var out [][]byte
	for _, b := range blocks {
		if b.ID == BlockGeneral {
			out = append(out, b.Data)
		}
	}
	return out
}

func TestWAVRoundTrip(t *testing.T) {
	tape := twoPrograms()
	wav := SaveWAV(tape)
	if string(wav[:4]) != "RIFF" || string(wav[8:16]) != "WAVEfmt " || binary.LittleEndian.Uint32(wav[24:]) != 44100 {
		t.Fatalf("header % x", wav[:44])
	}
	back, err := LoadWAV(wav, false)
	if err != nil {
		t.Fatal(err)
	}
	want, got := programs(tape), programs(back)
	if len(got) != 2 || !bytes.Equal(got[0], want[0]) || !bytes.Equal(got[1], want[1]) {
		t.Fatalf("programs: %d, first % x", len(got), got)
	}
	if name, _, _ := SplitSaved(got[1]); name != "DOIS" {
		t.Fatalf("name %q", name)
	}
}

// A WAV as a sound card records it: 16 bits, stereo, 22050 Hz, lower
// volume, some noise. The ROM's bit detection still reads it.
func TestWAVFromRecording(t *testing.T) {
	src := SaveWAV(twoPrograms())
	w, err := parseWAV(src)
	if err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewSource(1))
	var pcm bytes.Buffer
	for i := 0; i < w.n; i += 2 { // 44100 -> 22050
		v := (int(w.data[i]) - 128) * 200 // 8 bits unsigned -> 16 bits signed, ~80% volume
		v += rng.Intn(1500) - 750
		for range 2 {
			binary.Write(&pcm, binary.LittleEndian, int16(v))
		}
	}
	var h bytes.Buffer
	u32 := func(v int) { binary.Write(&h, binary.LittleEndian, uint32(v)) }
	u16 := func(v int) { binary.Write(&h, binary.LittleEndian, uint16(v)) }
	h.WriteString("RIFF")
	u32(36 + 12 + pcm.Len())
	h.WriteString("WAVE")
	h.WriteString("LIST") // a chunk the reader must skip
	u32(4)
	h.WriteString("INFO")
	h.WriteString("fmt ")
	u32(16)
	u16(1)
	u16(2)
	u32(22050)
	u32(22050 * 4)
	u16(4)
	u16(16)
	h.WriteString("data")
	u32(pcm.Len())
	h.Write(pcm.Bytes())
	back, err := LoadWAV(h.Bytes(), false)
	if err != nil {
		t.Fatal(err)
	}
	want, got := programs(twoPrograms()), programs(back)
	if len(got) != 2 || !bytes.Equal(got[0], want[0]) || !bytes.Equal(got[1], want[1]) {
		t.Fatalf("programs from the recording: %d", len(got))
	}
}

func TestWAVErrors(t *testing.T) {
	if _, err := LoadWAV([]byte("RIFF....WAVE"), false); err == nil {
		t.Fatal("a WAV without data was read")
	}
	silence := SaveWAV([]Block{{ID: BlockPause, Pause: 500}})
	if _, err := LoadWAV(silence, false); err == nil {
		t.Fatal("silence gave a program")
	}
}
