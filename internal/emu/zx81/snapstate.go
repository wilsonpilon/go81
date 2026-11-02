// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package zx81

// VideoState is the part of the video circuit that EightyOne keeps in a
// snapshot ([ZX81] of a .z81): the NMI generator, the sync output and the
// line counter.
type VideoState struct {
	NMI, Sync bool
	Line      int
}

// VideoState returns the video state for a snapshot.
func (m *Machine) VideoState() VideoState {
	return VideoState{NMI: m.nmiGeneratorEnabled, Sync: m.syncOutputWhite, Line: m.lineCounter}
}

// SetVideoState restores it (after the CPU registers, as halted comes from
// the CPU).
func (m *Machine) SetVideoState(v VideoState) {
	m.nmiGeneratorEnabled = v.NMI
	m.syncOutputWhite = v.Sync
	m.lineCounter = v.Line & 7
	m.z80Halted = m.CPU.Halted
}

// ROMTop is the last address of the ROM (1FFFh for 8K).
func (m *Machine) ROMTop() int { return m.romTop }
