// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne wavtape/wavload_.cpp (ClockTick while recording,
// StartRec) and wavtape/WavCInterface.cpp (WavStartRec) — Copyright (C)
// 2003-2025 Michael D Wynne (GPL v2 or later).

package tape

// The WAV loader as the output of the deck (EightyOne: Options → Tape →
// Destination → WAV loader): SAVE is not flash-saved; the ROM saves in
// real time and the MIC signal is recorded as sound, sample by sample.

// WavRecorder is the sound recorded from the MIC output.
type WavRecorder struct {
	w         wavWriter
	recording bool
}

// NewWavRecorder creates an empty recording (44100 Hz, 8 bits, mono).
func NewWavRecorder() *WavRecorder {
	return &WavRecorder{w: wavWriter{tsSample: wavClock / wavRate, level: 128}}
}

// Recording reports whether a SAVE is being recorded now.
func (r *WavRecorder) Recording() bool { return r.recording }

// Seconds is the length of the recording.
func (r *WavRecorder) Seconds() float64 { return float64(len(r.w.out)) / wavRate }

// WAV is the recording as a .wav file.
func (r *WavRecorder) WAV() []byte { return wavFile(r.w.out) }

// Clear empties the recording.
func (r *WavRecorder) Clear() { *r = *NewWavRecorder() }

// WavStartRec is called by the ROM SAVE patch: with the WAV loader as the
// output, it starts recording (TWavLoad::StartRec).
func (d *Deck) WavStartRec() {
	if !d.Out && d.Rec != nil {
		d.Rec.recording = true
	}
}

// ClockTickMic is WavClockTick: the tape advances, and the MIC level goes
// to the recording while a SAVE is recorded.
func (d *Deck) ClockTickMic(tstates int, mic bool) {
	d.ClockTick(tstates)
	if d.Rec != nil && d.Rec.recording {
		d.Rec.w.tick(tstates, mic)
	}
}

// Output is where SAVE goes (EightyOne: Options → Tape → Destination).
type Output int

const (
	OutTape  Output = iota // the tape manager: flash save, a block per program
	OutWav                 // the WAV loader: real-time SAVE, recorded as sound
	OutAudio               // the audio output: real-time SAVE, only heard
)

// SetOutput chooses where SAVE goes. It stays chosen when a tape is put in.
func (d *Deck) SetOutput(o Output) {
	d.outChosen = true
	d.Out = o == OutTape
	switch {
	case o != OutWav:
		if d.Rec != nil {
			d.Rec.recording = false
		}
		d.Rec = nil
	case d.Rec == nil:
		d.Rec = NewWavRecorder()
	}
}

// Output tells where SAVE goes.
func (d *Deck) Output() Output {
	switch {
	case d.Out:
		return OutTape
	case d.Rec != nil:
		return OutWav
	}
	return OutAudio
}

// keepOutput is what putting a tape in does to the output: the tape
// manager, unless the user chose one.
func (d *Deck) keepOutput() {
	if !d.outChosen {
		d.Out = true
	}
}
