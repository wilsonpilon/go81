// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne tzx/tzxload.cpp, tzx/tzxsave.cpp and
// tzx/TZXFILE.cpp — Copyright (C) 2003-2025 Michael D Wynne (GPL v2 or
// later).

package tape

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Tape files: .tzx (the TZX format, version 1.20) and .t81 (EightyOne's
// own: "EO81", then blocks with a 32-byte name and a 16-byte decimal
// length). The ZX81 programs are general data blocks (ID 19h); pauses,
// texts and the other TZX blocks are kept, so a tape saved again has them.
// The player plays general blocks and pauses and skips the rest, as
// EightyOne does for a ZX81.

// More TZX block IDs.
const (
	BlockROM      = 0x10
	BlockTurbo    = 0x11
	BlockTone     = 0x12
	BlockPulse    = 0x13
	BlockData     = 0x14
	BlockDRec     = 0x15
	BlockCSW      = 0x18
	BlockGStart   = 0x21
	BlockGEnd     = 0x22
	BlockJump     = 0x23
	BlockLStart   = 0x24
	BlockLEnd     = 0x25
	BlockCall     = 0x26
	BlockReturn   = 0x27
	BlockSelect   = 0x28
	BlockStop48K  = 0x2A
	BlockSetLevel = 0x2B
	BlockMessage  = 0x31
	BlockArchive  = 0x32
	BlockHWType   = 0x33
	BlockCustom   = 0x35
	BlockGlue     = 0x5A
)

const (
	tzxID = "ZXTape!\x1a"
	t81ID = "EO81"
)

// IsTapeFile reports whether a name is a tape of several blocks: .tzx,
// .t81 or a recorded tape (.wav).
func IsTapeFile(name string) bool {
	n := strings.ToLower(name)
	return strings.HasSuffix(n, ".tzx") || strings.HasSuffix(n, ".t81") || IsWavFile(n)
}

// IsWavFile reports whether a name is a sound file (.wav).
func IsWavFile(name string) bool { return strings.HasSuffix(strings.ToLower(name), ".wav") }

// LoadTape reads a .tzx, a .t81 or a .wav (ZX81 programs) by its first
// bytes.
func LoadTape(data []byte) ([]Block, error) {
	switch {
	case bytes.HasPrefix(data, []byte(tzxID)):
		return LoadTZX(data)
	case bytes.HasPrefix(data, []byte(t81ID)):
		return LoadT81(data)
	case bytes.HasPrefix(data, []byte("RIFF")):
		return LoadWAV(data, false)
	}
	return nil, errors.New("não é uma fita .tzx, .t81 nem .wav")
}

// tzxReader reads little-endian values; past the end it reads zeros and
// remembers the error.
type tzxReader struct {
	b   []byte
	i   int
	bad bool
}

func (r *tzxReader) bytes(n int) []byte {
	if n < 0 || r.i+n > len(r.b) {
		r.bad = true
		r.i = len(r.b)
		return make([]byte, max(n, 0))
	}
	out := r.b[r.i : r.i+n]
	r.i += n
	return out
}

func (r *tzxReader) u8() int  { return int(r.bytes(1)[0]) }
func (r *tzxReader) u16() int { return int(binary.LittleEndian.Uint16(r.bytes(2))) }
func (r *tzxReader) u24() int {
	b := r.bytes(3)
	return int(b[0]) | int(b[1])<<8 | int(b[2])<<16
}
func (r *tzxReader) u32() int { return int(binary.LittleEndian.Uint32(r.bytes(4))) }

// LoadTZX reads a TZX file.
func LoadTZX(data []byte) ([]Block, error) {
	if len(data) < 10 || !bytes.HasPrefix(data, []byte(tzxID)) {
		return nil, errors.New("não é uma fita .tzx")
	}
	r := &tzxReader{b: data, i: 10}
	var blocks []Block
	for r.i < len(r.b) {
		id := r.u8()
		if id == 0 || id == 0xF0 { // EightyOne stops there
			break
		}
		start := r.i
		b, err := r.block(id)
		if err != nil {
			return blocks, err
		}
		if r.bad {
			return blocks, fmt.Errorf("bloco %02Xh cortado no fim do arquivo", id)
		}
		if b.Raw == nil && b.ID != BlockGeneral && b.ID != BlockPause && b.ID != BlockText {
			b.Raw = append([]byte(nil), r.b[start:r.i]...)
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

// block reads the body of one block (after its ID).
func (r *tzxReader) block(id int) (Block, error) {
	b := Block{ID: byte(id)}
	switch id {
	case BlockGeneral:
		return r.general()
	case BlockPause:
		b.Pause = r.u16()
	case BlockText:
		b.Text = string(r.bytes(r.u8()))
	case BlockROM:
		b.Pause = r.u16()
		r.bytes(r.u16())
	case BlockTurbo:
		r.bytes(13)
		b.Pause = r.u16()
		r.bytes(r.u24())
	case BlockTone:
		r.bytes(4)
	case BlockPulse:
		r.bytes(2 * r.u8())
	case BlockData:
		r.bytes(5)
		b.Pause = r.u16()
		r.bytes(r.u24())
	case BlockDRec:
		r.bytes(2)
		b.Pause = r.u16()
		r.bytes(1)
		r.bytes(r.u24())
	case BlockCSW:
		r.bytes(r.u32())
	case BlockGStart:
		b.Text = string(r.bytes(r.u8()))
	case BlockGEnd, BlockLEnd, BlockReturn:
	case BlockJump, BlockLStart:
		r.bytes(2)
	case BlockCall:
		r.bytes(2 * r.u16())
	case BlockSelect, BlockArchive:
		r.bytes(r.u16())
	case BlockStop48K:
		r.bytes(r.u32())
	case BlockSetLevel:
		r.bytes(r.u32())
	case BlockMessage:
		r.bytes(1)
		b.Text = string(r.bytes(r.u8()))
	case BlockHWType:
		r.bytes(3 * r.u8())
	case BlockCustom:
		b.Text = strings.TrimRight(string(r.bytes(16)), " \x00")
		r.bytes(r.u32())
	case BlockGlue:
		r.bytes(9)
	default:
		if id < 0x10 || id > 0x5A {
			return b, fmt.Errorf("bloco desconhecido %02Xh", id)
		}
		// TZX rule for blocks of later versions: a 4-byte length first.
		r.bytes(r.u32())
	}
	return b, nil
}

// general reads a general data block (ID 19h), as LoadGeneralBlock: the
// pilot and sync table, the data symbols and the data.
func (r *tzxReader) general() (Block, error) {
	start := r.i
	if b, ok := r.oldGeneral(); ok {
		return b, nil
	}
	r.i, r.bad = start, false
	length := r.u32()
	end := r.i + length
	b := Block{ID: BlockGeneral}
	b.Pause = r.u16()
	b.TOTP = r.u32()
	b.NPP = r.u8()
	b.ASP = r.u8()
	if b.ASP == 0 && b.TOTP > 0 {
		b.ASP = 256
	}
	if b.TOTP == 0 {
		b.NPP, b.ASP = 0, 0
	}
	b.TOTD = r.u32()
	b.NPD = r.u8()
	b.ASD = r.u8()
	if b.ASD == 0 && b.TOTD > 0 {
		b.ASD = 256
	}
	if b.TOTD == 0 {
		b.NPD, b.ASD = 0, 0
	}
	// The file has NPP/NPD as "pulses", EightyOne stores them + 1 (the
	// flags word first).
	if b.TOTP > 0 {
		b.NPP++
		b.SymDefP = r.symbols(b.ASP, b.NPP)
		b.PRLE = make([]uint16, 2*b.TOTP)
		for i := range b.TOTP {
			b.PRLE[2*i] = uint16(r.u8())
			b.PRLE[2*i+1] = uint16(r.u16())
		}
	}
	if b.TOTD > 0 {
		b.NPD++
		b.SymDefD = r.symbols(b.ASD, b.NPD)
		bits := 0
		for i := 1; i < b.ASD; i <<= 1 {
			bits++
		}
		n := (bits*b.TOTD + 7) / 8
		b.Data = append([]byte(nil), r.bytes(n)...)
		b.DataLen = n
	}
	if r.i != end && end <= len(r.b) && !r.bad {
		r.i = end // trust the block length
	}
	return b, nil
}

// symbols reads a symbol table: per symbol a flags byte and n-1 pulses.
func (r *tzxReader) symbols(count, n int) []uint16 {
	t := make([]uint16, count*n)
	for i := range count {
		t[i*n] = uint16(r.u8())
		for j := 1; j < n; j++ {
			t[i*n+j] = uint16(r.u16())
		}
	}
	return t
}

// oldGeneral reads the general block of the first TZX 1.20 drafts, which
// early EightyOne versions wrote (LoadOldGeneralBlock); ok is false when
// the block is not in that form.
func (r *tzxReader) oldGeneral() (Block, bool) {
	bl := r.u32()
	flags := r.u8()
	pl := r.u16()
	pp := r.u16()
	ns := r.u8()
	if flags != 0 && flags != 1 || pl != 0 || pp != 0 || ns != 0 || r.bad {
		return Block{}, false
	}
	np := r.u8()
	as := r.u8()
	if as != 2 {
		return Block{}, false
	}
	r.bytes(2 * np * as)
	r.u8() // used bits
	pause := r.u16()
	n := bl - (11 + 2*(ns+np*as))
	data := r.bytes(n)
	if r.bad || n < 0 {
		return Block{}, false
	}
	b := Block{ID: BlockGeneral, Pause: pause, TOTD: n * 8, NPD: 19, ASD: 2, DataLen: n,
		Data: append([]byte(nil), data...)}
	b.SymDefD = append([]uint16(nil), zx81SymDefD...)
	return b, true
}

// SaveTZX writes the blocks as a TZX file (SaveFile).
func SaveTZX(blocks []Block) []byte {
	var w bytes.Buffer
	w.WriteString(tzxID)
	version := 13
	for _, b := range blocks {
		if b.ID == BlockGeneral || b.ID == BlockSetLevel {
			version = 20
		}
	}
	w.WriteByte(1)
	w.WriteByte(byte(version))
	u16 := func(v int) { w.WriteByte(byte(v)); w.WriteByte(byte(v >> 8)) }
	u32 := func(v int) { u16(v); u16(v >> 16) }
	for _, b := range blocks {
		switch {
		case b.ID == BlockGeneral:
			n := 14
			if b.TOTP > 0 {
				n += b.ASP*(1+(b.NPP-1)*2) + b.TOTP*3
			}
			if b.TOTD > 0 {
				n += b.ASD*(1+(b.NPD-1)*2) + b.DataLen
			}
			w.WriteByte(BlockGeneral)
			u32(n)
			u16(b.Pause)
			u32(b.TOTP)
			if b.TOTP == 0 {
				u16(0)
			} else {
				w.WriteByte(byte(b.NPP - 1))
				w.WriteByte(byte(b.ASP))
			}
			u32(b.TOTD)
			if b.TOTD == 0 {
				u16(0)
			} else {
				w.WriteByte(byte(b.NPD - 1))
				w.WriteByte(byte(b.ASD))
			}
			if b.TOTP > 0 {
				for i := range b.ASP {
					w.WriteByte(byte(u16At(b.SymDefP, i*b.NPP)))
					for j := 1; j < b.NPP; j++ {
						u16(int(u16At(b.SymDefP, i*b.NPP+j)))
					}
				}
				for i := range b.TOTP {
					w.WriteByte(byte(u16At(b.PRLE, 2*i)))
					u16(int(u16At(b.PRLE, 2*i+1)))
				}
			}
			if b.TOTD > 0 {
				for i := range b.ASD {
					w.WriteByte(byte(u16At(b.SymDefD, i*b.NPD)))
					for j := 1; j < b.NPD; j++ {
						u16(int(u16At(b.SymDefD, i*b.NPD+j)))
					}
				}
				data := make([]byte, b.DataLen)
				copy(data, b.Data)
				w.Write(data)
			}
		case b.ID == BlockPause:
			w.WriteByte(BlockPause)
			u16(b.Pause)
		case b.ID == BlockText:
			t := []byte(b.Text)
			if len(t) > 255 {
				t = t[:255]
			}
			w.WriteByte(BlockText)
			w.WriteByte(byte(len(t)))
			w.Write(t)
		case b.Raw != nil:
			w.WriteByte(b.ID)
			w.Write(b.Raw)
		}
	}
	return w.Bytes()
}

func u16At(t []uint16, i int) uint16 {
	if i >= 0 && i < len(t) {
		return t[i]
	}
	return 0
}

// LoadT81 reads a .t81 file (LoadT81File): a text block first, as a new
// tape of EightyOne, then a general block for each program and a pause
// block for each "<Silence>"; pauses after a program join it, and a
// program without a pause gets 5 s.
func LoadT81(data []byte) ([]Block, error) {
	if !bytes.HasPrefix(data, []byte(t81ID)) {
		return nil, errors.New("não é uma fita .t81")
	}
	p := NewPlayer(3250000)
	p.NewTape()
	for i := 4; i+48 <= len(data); {
		name := cString(data[i : i+32])
		n, _ := strconv.Atoi(cString(data[i+32 : i+48]))
		i += 48
		if len(name) > 32 || n < 1 || n > 65535 {
			break
		}
		if strings.HasPrefix(name, "<") && strings.HasSuffix(name, ">") {
			if name == "<Silence>" {
				p.Blocks = append(p.Blocks, Block{ID: BlockPause, Pause: n})
			} else {
				i += n // a block of another program: skipped
			}
			continue
		}
		if i+n > len(data) {
			n = len(data) - i
		}
		prog := data[i : i+n]
		i += n
		var block []byte
		if len(prog) > 0 && (prog[0] == 0x00 || prog[0] == 0xFF || prog[0] == 0x01) && name != "" {
			// The program has no ZX81 name in front: it is made from the name.
			block = append(ZX81Name(name), prog...)
		} else {
			block = append([]byte(nil), prog...)
		}
		for len(block) > 0 && block[len(block)-1] != 0x80 {
			block = block[:len(block)-1]
		}
		p.AddGeneral(block)
	}
	blocks := mergeBlocks(p.Blocks)
	for i := 1; i < len(blocks); i++ {
		if blocks[i].ID == BlockGeneral && blocks[i].Pause == 0 {
			blocks[i].Pause = 5000
		}
	}
	return blocks, nil
}

func cString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return string(b)
}

// mergeBlocks is MergeBlocks: a pause after a program (or a pause) joins
// it; a pause right after the information blocks at the start goes.
func mergeBlocks(blocks []Block) []Block {
	out := blocks[:0:0]
	for _, b := range blocks {
		if n := len(out); n > 0 && b.ID == BlockPause && (out[n-1].ID == BlockGeneral || out[n-1].ID == BlockPause) {
			out[n-1].Pause += b.Pause
			continue
		}
		out = append(out, b)
	}
	i := 0
	for i < len(out) && (out[i].ID == BlockText || out[i].ID == BlockMessage || out[i].ID == BlockArchive || out[i].ID == BlockHWType) {
		i++
	}
	if i < len(out) && out[i].ID == BlockPause {
		out = append(out[:i], out[i+1:]...)
	}
	return out
}

// SaveT81 writes the programs and pauses as a .t81 (SaveT81File): each
// program with its ASCII name, and a "<Silence>" after it when its pause is
// not the usual 5 s.
func SaveT81(blocks []Block) []byte {
	var w bytes.Buffer
	w.WriteString(t81ID)
	field := func(s string, n int) {
		f := make([]byte, n)
		copy(f, s)
		w.Write(f)
	}
	for _, b := range blocks {
		pause := b.ID == BlockPause
		if b.ID == BlockGeneral {
			name, _, _ := SplitSaved(b.Data)
			if len(name) > 32 {
				name = name[:32]
			}
			field(name, 32)
			field(strconv.Itoa(b.DataLen), 16)
			data := make([]byte, b.DataLen)
			copy(data, b.Data)
			w.Write(data)
			pause = b.Pause != 5000
		}
		if pause {
			field("<Silence>", 32)
			field(strconv.Itoa(b.Pause), 16)
		}
	}
	return w.Bytes()
}
