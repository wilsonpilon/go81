// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne sound/sound.cpp (CSound::Beeper, CSound::Frame) —
// Copyright (C) 2003-2025 Michael D Wynne; beeper code originally from
// Fuse by Russell Marks (GPL v2 or later).

package zx81

// Audio format produced by the machine (EightyOne's defaults).
const (
	SampleRate    = 44100
	AudioChannels = 2

	amplBeeper    = 31
	beeperFadeout = ((1 << 24) / 150) / amplBeeper
)

// beeper is the part of EightyOne's CSound used by the ZX81.
type beeper struct {
	frameSize  int
	tPerFrame  int
	volume     int // VolumeLevel[3]: 31 - slider position (16) = 15
	buf        []int16
	out        []int16
	oldVal     int
	oldValOrig int
	oldPos     int
	fillPos    int
	lastSubpos int
	tick       int
	tickIncr   int
}

func (b *beeper) init(fps, tPerFrame int) {
	b.frameSize = SampleRate / fps
	b.tPerFrame = tPerFrame
	b.volume = 15
	b.buf = make([]int16, b.frameSize*AudioChannels)
	b.out = make([]int16, b.frameSize*AudioChannels)
	b.oldVal, b.oldValOrig = 0, 0
	b.oldPos = -1
	b.fillPos = 0
	b.lastSubpos = 0
	b.tick = 0
	b.tickIncr = (1 << 24) / SampleRate
}

// adjust is BEEPER_OLDVAL_ADJUST: a slow fade towards zero.
func (b *beeper) adjust() {
	b.tick += b.tickIncr
	if b.tick >= beeperFadeout {
		b.tick -= beeperFadeout
		if b.oldVal > 0 {
			b.oldVal--
		} else if b.oldVal < 0 {
			b.oldVal++
		}
	}
}

func (b *beeper) put(f, v int) {
	s := int16(v)
	b.buf[f*AudioChannels] = s
	b.buf[f*AudioChannels+1] = s
}

// set is CSound::Beeper.
func (b *beeper) set(on bool, frameTStates int) {
	val := -amplBeeper
	if on {
		val = amplBeeper
	}
	if val == b.oldValOrig {
		return
	}
	newpos := (frameTStates * b.frameSize) / b.tPerFrame
	subpos := (frameTStates*b.frameSize*amplBeeper)/b.tPerFrame - amplBeeper*newpos

	if newpos == b.oldPos {
		if on {
			b.lastSubpos += amplBeeper - subpos
		} else {
			b.lastSubpos -= amplBeeper - subpos
		}
	} else if on {
		b.lastSubpos = amplBeeper - subpos
	} else {
		b.lastSubpos = subpos
	}
	subval := 0 - amplBeeper + b.lastSubpos

	if newpos >= 0 {
		for f := b.fillPos; f < newpos && f < b.frameSize; f++ {
			b.adjust()
			b.put(f, (b.oldVal*256*b.volume)/amplBeeper)
		}
		if newpos < b.frameSize {
			if (b.oldVal < 0 && subval < b.oldVal) || (b.oldVal >= 0 && subval > b.oldVal) {
				subval = b.oldVal
			}
			b.put(newpos, (subval*256*b.volume)/amplBeeper)
		}
	}
	b.oldPos = newpos
	b.fillPos = newpos + 1
	b.oldVal, b.oldValOrig = val, val
}

// frame is CSound::Frame: completes the frame's samples and publishes them.
func (b *beeper) frame() {
	for f := b.fillPos; f < b.frameSize; f++ {
		b.adjust()
		b.put(f, (b.oldVal*256*b.volume)/amplBeeper)
	}
	b.oldPos = -1
	b.fillPos = 0
	copy(b.out, b.buf)
}
