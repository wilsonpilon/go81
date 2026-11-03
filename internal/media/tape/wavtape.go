// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne wavtape/wavload_.cpp and wavtape/WavFile.cpp —
// Copyright (C) 2003-2025 Michael D Wynne (GPL v2 or later).

package tape

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Tapes as sound (.wav): EightyOne's Wave Loader. Reading runs the bit
// detection of the ZX81 ROM's LOAD on the sound, with the same T-state
// counts, one program at a time ("Convert next block"), so a WAV recorded
// from a real cassette becomes tape blocks. Writing makes the pulses the
// ZX81 SAVE makes (EncodeBlock), 44100 Hz, 8 bits, mono.

// wavClock is the ZX81's clock, used to step through the samples.
const wavClock = 3250000

// WavThreshold is EightyOne's default threshold: how far from the middle
// the sound must go to count as a pulse.
const WavThreshold = 50

// wavTape is TWavFile plus the playing part of TWavLoad.
type wavTape struct {
	data       []byte // the sample bytes
	rate       int
	channels   int
	bytesFrame int // bytes per sample, all channels
	signed     bool
	n          int // number of samples

	Threshold int
	tsSample  int
	pos       int
	cur       int
	playing   bool
}

// parseWAV reads a RIFF WAV (8 or 16 bits, mono or stereo), as
// TWavFile::LoadFile.
func parseWAV(data []byte) (*wavTape, error) {
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, errors.New("não é um arquivo WAV")
	}
	w := &wavTape{Threshold: WavThreshold}
	var haveFmt bool
	for i := 12; i+8 <= len(data); {
		id := string(data[i : i+4])
		n := int(binary.LittleEndian.Uint32(data[i+4:]))
		body := data[i+8:]
		if n > len(body) {
			n = len(body)
		}
		switch id {
		case "fmt ":
			if n < 16 {
				return nil, errors.New("WAV: bloco fmt curto")
			}
			if f := binary.LittleEndian.Uint16(body); f != 1 {
				return nil, fmt.Errorf("WAV: só PCM (formato %d)", f)
			}
			w.channels = int(binary.LittleEndian.Uint16(body[2:]))
			w.rate = int(binary.LittleEndian.Uint32(body[4:]))
			w.bytesFrame = int(binary.LittleEndian.Uint16(body[12:]))
			bits := int(binary.LittleEndian.Uint16(body[14:]))
			w.signed = bits > 8
			haveFmt = true
		case "data":
			if !haveFmt {
				return nil, errors.New("WAV: dados antes do formato")
			}
			w.data = body[:n]
		}
		i += 8 + n + n&1
		if w.data != nil {
			break
		}
	}
	if !haveFmt || w.data == nil || w.rate <= 0 || w.channels <= 0 || w.bytesFrame <= 0 {
		return nil, errors.New("WAV sem formato ou sem dados")
	}
	if s := w.bytesFrame / w.channels; s != 1 && s != 2 {
		return nil, fmt.Errorf("WAV: amostras de %d bits não são lidas (use 8 ou 16)", 8*s)
	}
	w.n = len(w.data) / w.bytesFrame
	w.tsSample = wavClock / w.rate
	if w.tsSample < 1 {
		return nil, errors.New("WAV: taxa de amostragem alta demais")
	}
	return w, nil
}

// sample is TWavFile::Sample (first channel), inverted, as EightyOne.
func (w *wavTape) sample(i int) int {
	if i >= w.n {
		return 128
	}
	pos := i * w.bytesFrame
	var d byte
	switch w.bytesFrame / w.channels {
	case 1:
		d = w.data[pos]
	case 2:
		d = byte(binary.LittleEndian.Uint16(w.data[pos:]) / 256)
	default:
		d = 128
	}
	if w.signed {
		d = byte(128 + int(int8(d)))
	}
	d = 255 - d
	v := int(d) - 128
	v = max(-127, min(127, v))
	return v + 128
}

// ear is TWavLoad::GetEarState.
func (w *wavTape) ear() bool {
	if !w.playing {
		return false
	}
	i := 128 - w.sample(w.pos+200)
	return w.Threshold >= 0 && i >= w.Threshold || w.Threshold < 0 && i < w.Threshold
}

// tick is TWavLoad::ClockTick while playing: at most one sample a call.
func (w *wavTape) tick(ts int) {
	if !w.playing {
		return
	}
	w.cur += ts
	if w.cur < w.tsSample {
		return
	}
	w.cur -= w.tsSample
	w.pos++
	if w.pos > w.n {
		w.playing = false
	}
}

// getByte is TWavLoad::GetByte: the ROM's LOAD loop (GET-BYTE, 0337h) on
// the sound; -1 when no byte came.
func (w *wavTape) getByte() int {
	data := 1
	reset := true
	b := 256
	w.tick(14)
	for b--; b != 0; b-- {
		w.tick(68)
		if !w.playing {
			return -1
		}
		if !w.ear() {
			continue
		}
		e := 0x94
		w.tick(23)
		var zf bool
		for {
			if reset {
				b = 0x1a
				reset = false
				w.tick(7)
			}
			if !w.playing {
				return -1
			}
			e--
			w.tick(31)
			cf := w.ear()
			zf = e&128 == 0
			if cf {
				reset = true
				w.tick(12)
			} else {
				b--
				w.tick(20)
			}
			if !cf && b == 0 {
				break
			}
		}
		w.tick(37)
		if zf && e >= 0x56 {
			b = 256
		} else {
			w.tick(8)
			v := 0
			if e < 32 {
				v = 1
			}
			data = data<<1 | v
			if data&256 != 0 {
				return data & 255
			}
		}
	}
	return -1
}

// nextBlock is ConvertNextBlock1Click: the silence before a program and
// the program (nil if none). zx80 reads a ZX80 program (no name).
func (w *wavTape) nextBlock(zx80 bool) (silence int, prog []byte) {
	start := w.pos
	c := -1
	for c == -1 && w.playing {
		c = w.getByte()
	}
	silence = (w.pos - start) * w.tsSample
	var buf []byte
	dataAt := -1 // index in buf of the first byte after the name
	count, eLine := 0, -1
	for c != -1 && count != eLine {
		buf = append(buf, byte(c))
		if dataAt < 0 {
			if zx80 || c&0x80 != 0 {
				dataAt = len(buf)
			}
		} else {
			count++
		}
		d := func(k int) int {
			if dataAt+k < len(buf) {
				return int(buf[dataAt+k])
			}
			return 0
		}
		if !zx80 && count == 12 {
			eLine = d(11) + d(12)*256 - 16392
		}
		if zx80 && count == 11 {
			eLine = d(10) + d(11)*256 - 16383
		}
		c = w.getByte()
	}
	if len(buf) > 32 {
		prog = buf
	}
	return silence, prog
}

// LoadWAV reads every program of a WAV tape: a pause for each silence and
// a general block for each program, merged as EightyOne does. zx80 reads
// ZX80 programs.
func LoadWAV(data []byte, zx80 bool) ([]Block, error) {
	w, err := parseWAV(data)
	if err != nil {
		return nil, err
	}
	p := NewPlayer(wavClock)
	w.playing = true
	for w.playing {
		silence, prog := w.nextBlock(zx80)
		if ms := silence / 3250; ms > 0 {
			p.Blocks = append(p.Blocks, Block{ID: BlockPause, Pause: min(ms, 65535)})
		}
		if prog != nil {
			p.AddGeneral(prog)
		}
	}
	blocks := mergeBlocks(p.Blocks)
	found := false
	for _, b := range blocks {
		if b.ID == BlockGeneral {
			found = true
		}
	}
	if !found {
		return nil, errors.New("nenhum programa encontrado no WAV (volume baixo? gravação de outra máquina?)")
	}
	return blocks, nil
}

// wavWriter is the recording part of TWavLoad (EncodeBlock).
type wavWriter struct {
	out              []byte
	tsSample, cur    int
	count, level     int
	last             bool
	countSinceChange bool
}

const wavRate = 44100

func (w *wavWriter) tick(ts int, mic bool) {
	w.cur += ts
	if w.cur < w.tsSample {
		return
	}
	w.cur -= w.tsSample
	if mic == w.last {
		w.count++
	} else {
		w.count = 0
	}
	const pulseWidth, mid = 6, 128
	if w.count < 128 {
		w.level = mid + 96
		if mic {
			w.level = mid - 96
		}
		if w.count > pulseWidth {
			w.level = mid
		}
	} else {
		w.level = (w.level-mid)*9/10 + mid
	}
	w.out = append(w.out, byte(w.level))
	w.last = mic
}

const tsClick = 4

func (w *wavWriter) bit(one bool) {
	n := 4
	if one {
		n = 8
	}
	for range n {
		for range 160 {
			w.tick(tsClick, true)
		}
		for range 160 {
			w.tick(tsClick, false)
		}
	}
	for range 1280 {
		w.tick(tsClick, true)
	}
}

func (w *wavWriter) silence(ts int) {
	for range ts / tsClick {
		w.tick(tsClick, false)
	}
}

func (w *wavWriter) block(data []byte) {
	w.silence(128 * w.tsSample * tsClick)
	for _, c := range data {
		for range 8 {
			w.bit(c&128 != 0)
			c <<= 1
		}
	}
	w.silence(128 * w.tsSample * tsClick)
}

// SaveWAV writes the programs of a tape as sound a ZX81 loads (44100 Hz,
// 8 bits, mono): the pulses of each program, and the pauses between them.
func SaveWAV(blocks []Block) []byte { return wavFile(encodeBlocks(blocks)) }

// encodeBlocks is the sound of the blocks, as 8-bit samples at 44100 Hz.
func encodeBlocks(blocks []Block) []byte {
	w := &wavWriter{tsSample: wavClock / wavRate, level: 128}
	for _, b := range blocks {
		switch b.ID {
		case BlockGeneral:
			w.block(b.Data[:min(b.DataLen, len(b.Data))])
			w.silence(b.Pause * 3250)
		case BlockPause:
			w.silence(b.Pause * 3250)
		}
	}
	return w.out
}

// wavFile makes a .wav of 8-bit mono samples at 44100 Hz.
func wavFile(samples []byte) []byte {
	var h bytes.Buffer
	u32 := func(v int) { binary.Write(&h, binary.LittleEndian, uint32(v)) }
	u16 := func(v int) { binary.Write(&h, binary.LittleEndian, uint16(v)) }
	h.WriteString("RIFF")
	u32(36 + len(samples))
	h.WriteString("WAVEfmt ")
	u32(16)
	u16(1)
	u16(1)
	u32(wavRate)
	u32(wavRate)
	u16(1)
	u16(8)
	h.WriteString("data")
	u32(len(samples))
	return append(h.Bytes(), samples...)
}
