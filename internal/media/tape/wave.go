// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tape

// A tape sound to look at (the waveform of EightyOne's Wave Loader
// window): samples from 0 to 255, 128 being silence.

// Wave is the sound of a tape, sample by sample.
type Wave interface {
	Len() int         // number of samples
	Rate() int        // samples per second
	Sample(i int) int // 0 to 255; 128 = silence
	Position() int    // the sample playing now (-1 = none)
}

type playerWave struct{ w *wavTape }

func (p playerWave) Len() int         { return p.w.n }
func (p playerWave) Rate() int        { return p.w.rate }
func (p playerWave) Sample(i int) int { return p.w.sample(i) }
func (p playerWave) Position() int {
	if !p.w.playing && p.w.pos == 0 {
		return -1
	}
	return p.w.pos
}

// Wave is the WAV being played into the ZX81.
func (p *WavPlayer) Wave() Wave { return playerWave{p.w} }

// Threshold is the level the loader reads as a pulse (128 − sample ≥ it).
func (p *WavPlayer) Threshold() int { return p.w.Threshold }

// SampleWave is a wave held as 8-bit samples at 44100 Hz.
type SampleWave struct {
	S   []byte
	Pos int // -1 = no position
}

func (s *SampleWave) Len() int  { return len(s.S) }
func (s *SampleWave) Rate() int { return wavRate }
func (s *SampleWave) Sample(i int) int {
	if i < 0 || i >= len(s.S) {
		return 128
	}
	return int(s.S[i])
}
func (s *SampleWave) Position() int { return s.Pos }

// Wave is a copy of the sound recorded from SAVE so far; the position is
// its end while it records.
func (r *WavRecorder) Wave() Wave {
	pos := -1
	if r.recording {
		pos = len(r.w.out)
	}
	return &SampleWave{S: append([]byte(nil), r.w.out...), Pos: pos}
}

// BlocksWave is a tape of blocks as it sounds (what tape save x.wav writes).
func BlocksWave(blocks []Block) Wave {
	return &SampleWave{S: encodeBlocks(blocks), Pos: -1}
}
