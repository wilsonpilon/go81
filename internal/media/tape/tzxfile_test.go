// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tape

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// prog is a saved ZX81 program: its name (last letter inverted), then
// bytes ending in 80h, as the ROM saves.
func prog(name string, body ...byte) []byte {
	return append(ZX81Name(name), append(body, 0x80)...)
}

// The TZX written for a tape of two programs and a pause has the layout of
// the format (header, general block with the ZX81 symbols) and reads back
// to the same blocks.
func TestTZXRoundTrip(t *testing.T) {
	p := NewPlayer(3250000)
	p.NewTape()
	a := p.AddGeneral(prog("UM", 0, 1, 2))
	p.Blocks[a].Pause = 5000
	p.Blocks = append(p.Blocks, Block{ID: BlockPause, Pause: 1000})
	p.AddGeneral(prog("DOIS", 3, 4))
	data := SaveTZX(p.Blocks)
	if !bytes.HasPrefix(data, []byte("ZXTape!\x1a\x01\x14")) {
		t.Fatalf("header % x", data[:10])
	}
	// The text block, then the general block: ID, length, pause.
	i := 10 + 2 + len("Created using the EightyOne emulator")
	if data[i] != BlockGeneral {
		t.Fatalf("block at %d: %02x", i, data[i])
	}
	n := int(binary.LittleEndian.Uint32(data[i+1:]))
	want := 14 + 2*(1+18*2) + len(prog("UM", 0, 1, 2))
	if n != want || binary.LittleEndian.Uint16(data[i+5:]) != 5000 {
		t.Fatalf("general block length %d (want %d)", n, want)
	}
	back, err := LoadTape(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(back) != 4 || back[0].Text != p.Blocks[0].Text || back[2].Pause != 1000 {
		t.Fatalf("blocks %+v", back)
	}
	for _, k := range []int{1, 3} {
		g, w := back[k], p.Blocks[k]
		if !bytes.Equal(g.Data, w.Data) || g.TOTD != w.TOTD || g.NPD != w.NPD || g.ASD != w.ASD || g.Pause != w.Pause {
			t.Fatalf("block %d: %+v want %+v", k, g, w)
		}
		for j := range w.SymDefD {
			if g.SymDefD[j] != w.SymDefD[j] {
				t.Fatalf("block %d: symbols %v", k, g.SymDefD)
			}
		}
	}
	if !bytes.Equal(SaveTZX(back), data) {
		t.Fatal("saved again, the TZX differs")
	}
}

// Blocks the player does not use (a Spectrum ROM block, an archive info,
// a glue block) are kept as they are when the tape is saved again.
func TestTZXKeepsOtherBlocks(t *testing.T) {
	var w bytes.Buffer
	w.WriteString("ZXTape!\x1a\x01\x14")
	w.Write([]byte{BlockROM, 0xE8, 0x03, 3, 0, 0x00, 0x11, 0x22})           // pause 1000, 3 bytes
	w.Write([]byte{BlockArchive, 5, 0, 1, 0x00, 2, 'O', 'I'})               // one string
	w.Write([]byte{BlockGlue, 'X', 'T', 'a', 'p', 'e', '!', 0x1a, 1, 0x14}) // glue
	w.Write([]byte{0x40, 2, 0, 0, 0, 0xAA, 0xBB})                           // a later block: DWORD length
	blocks, err := LoadTZX(w.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 4 || blocks[0].Pause != 1000 || blocks[3].ID != 0x40 {
		t.Fatalf("blocks %+v", blocks)
	}
	if !bytes.Equal(SaveTZX(blocks)[10:], w.Bytes()[10:]) {
		t.Fatalf("saved again:\n% x\n% x", SaveTZX(blocks), w.Bytes())
	}
	if _, err := LoadTZX(w.Bytes()[:len(w.Bytes())-1]); err == nil {
		t.Fatal("a cut block was accepted")
	}
}

// The .t81: "EO81", then a 32-byte name and a 16-byte length per block; a
// program without its ZX81 name gets one; pauses join the program before.
func TestT81(t *testing.T) {
	field := func(s string, n int) []byte { f := make([]byte, n); copy(f, s); return f }
	var w bytes.Buffer
	w.WriteString("EO81")
	p1 := prog("JOGO", 0, 1)
	w.Write(field("JOGO", 32))
	w.Write(field("8", 16))
	w.Write(p1[:8]) // the name is in the data
	body := []byte{0x00, 0x01, 0x02, 0x80}
	w.Write(field("SEM", 32))
	w.Write(field("4", 16))
	w.Write(body) // starts with 00: the name comes from the block
	w.Write(field("<Silence>", 32))
	w.Write(field("2000", 16))
	blocks, err := LoadT81(w.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 3 || blocks[0].ID != BlockText {
		t.Fatalf("blocks %+v", blocks)
	}
	if name, _, _ := SplitSaved(blocks[1].Data); name != "JOGO" || blocks[1].Pause != 5000 {
		t.Fatalf("first: %q pause %d", name, blocks[1].Pause)
	}
	if name, rest, _ := SplitSaved(blocks[2].Data); name != "SEM" || !bytes.Equal(rest, body) || blocks[2].Pause != 2000 {
		t.Fatalf("second: %q % x pause %d", name, rest, blocks[2].Pause)
	}
	// Saved again: the pause of 2 s is a <Silence> after the program.
	again := SaveT81(blocks)
	back, err := LoadT81(again)
	if err != nil || len(back) != 3 || back[2].Pause != 2000 || !bytes.Equal(back[1].Data, blocks[1].Data) {
		t.Fatalf("again: %v %+v", err, back)
	}
	if !bytes.Contains(again, field("<Silence>", 32)) {
		t.Fatal("no <Silence> for the 2 s pause")
	}
}
