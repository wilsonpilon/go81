// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne wavtape/wavload_.cpp and wavtape/WavCInterface.cpp
// — Copyright (C) 2003-2025 Michael D Wynne (GPL v2 or later).

package tape

// The WAV loader as the input of the deck (EightyOne: Options → Tape →
// Source → WAV loader): the ZX81 hears the sound itself, in real time, and
// its ROM reads the bits as from a cassette (no flash load).

// WavPlayer is a WAV tape played as sound.
type WavPlayer struct{ w *wavTape }

// NewWavPlayer reads a WAV file to play it.
func NewWavPlayer(data []byte) (*WavPlayer, error) {
	w, err := parseWAV(data)
	if err != nil {
		return nil, err
	}
	return &WavPlayer{w: w}, nil
}

// Playing reports whether the sound is playing.
func (p *WavPlayer) Playing() bool { return p.w.playing }

// Start is TWavLoad::Start (auto start/stop on): play from where it is.
func (p *WavPlayer) Start() {
	if !p.w.playing && p.w.pos < p.w.n {
		p.w.playing = true
	}
}

// Stop is TWavLoad::Stop.
func (p *WavPlayer) Stop() { p.w.playing = false }

// Rewind goes back to the start.
func (p *WavPlayer) Rewind() { p.w.playing, p.w.pos, p.w.cur = false, 0, 0 }

// Position is where the sound is, in seconds, and its length.
func (p *WavPlayer) Position() (at, length float64) {
	return float64(p.w.pos) / float64(p.w.rate), float64(p.w.n) / float64(p.w.rate)
}

// InsertWav makes the WAV the input of the deck (TZXin off); SAVE still
// goes to the tape manager.
func (d *Deck) InsertWav(p *WavPlayer) {
	d.StopTape()
	d.Wav = p
	d.In = false
	d.keepOutput()
}
