// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

// Methods for front-ends (window, terminal) that show the machine and run
// its frames themselves, while commands come from a console or the Link.

// Drive tells the session that the front-end runs the frames in real time
// (calling Tick). Commands then wait for the front-end instead of running
// frames themselves, and "clock" cannot be changed.
func (s *Session) Drive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.live && !s.driven {
		close(s.stopLiv) // the front-end replaces the clock
	}
	s.live, s.driven = true, true
}

// Tick runs n frames (turbo uses n > 1) unless paused. It reports whether
// the frames ran.
func (s *Session) Tick(n int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.paused {
		return false
	}
	for range n {
		s.frame()
	}
	return true
}

// Paused reports whether the machine is paused.
func (s *Session) Paused() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.paused
}

// SetPaused pauses or resumes the machine.
func (s *Session) SetPaused(p bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = p
}

// PressChord queues a key chord (e.g. "SHIFT+P") without waiting.
func (s *Session) PressChord(chord string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queueChord(chord, KeyHold)
}

// QueueText types text on the ZX81 without waiting (a new line is ENTER;
// carriage returns and characters with no ZX81 key are skipped). It
// returns how many characters were skipped.
func (s *Session) QueueText(text string) (skipped int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range text {
		if r == '\r' {
			continue
		}
		c := zx81.ChordForRune(r)
		if c == "" || s.queueChord(c, KeyHold) != nil {
			skipped++
		}
	}
	return skipped
}

// SetKeys sets the keyboard matrix directly (a window that knows which keys
// are held down).
func (s *Session) SetKeys(rows [8]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m.SetKeys(rows)
}

// InsertProgram puts a program on the tape without waiting; with auto the
// ZX81 is reset and types LOAD "" by itself.
func (s *Session) InsertProgram(name string, data []byte, auto bool) error {
	name, data, err := ProgramBytes(name, data) // a .bas becomes a .p
	if err != nil {
		return err
	}
	block, err := tape.ProgramBlock(name, data)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue, s.hold, s.gap = nil, 0, 0
	s.m.InsertTape(block, auto)
	return nil
}

// Configure replaces the machine (hard reset).
func (s *Session) Configure(set Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.configure(set)
}

// Reset turns the ZX81 off and on.
func (s *Session) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.powerCycle()
}

// SetSound turns the "Video Sound" on or off without a reset.
func (s *Session) SetSound(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Sound = on
	s.m.SetVSyncSound(on)
}

// Description is a one-line summary of the machine.
func (s *Session) Description() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.description()
}

// Frames returns how many frames ran since the machine was created.
func (s *Session) Frames() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.frames
}
