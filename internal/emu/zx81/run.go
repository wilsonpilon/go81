// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package zx81

// RunFrame runs one emulated frame (one 50/60 Hz tick), as EightyOne's main
// loop does: it executes scanlines until the frame's T-states are used, and
// carries the difference over to the next frame.
func (m *Machine) RunFrame() {
	m.frameTStates = 0
	j := m.TStatesPerFrame() + m.borrow
	for j > 0 {
		j -= m.doScanline()
		if m.LineHook != nil {
			m.LineHook(&m.line)
		}
		m.raster.draw(&m.line)
	}
	m.borrow = j
}

// FrameTStates returns the T-states run in the last frame.
func (m *Machine) FrameTStates() int { return m.frameTStates }

// Display returns the last complete TV frame (FrameWidth x FrameHeight,
// one byte per pixel holding colour<<4, colour 0 = black, 15 = bright white).
func (m *Machine) Display() *[FrameWidth * FrameHeight]byte { return &m.raster.Frame }

// RasterY returns the current raster line (used by the comparison tests).
func (m *Machine) RasterY() int { return m.raster.y }

// FramesCompleted returns how many TV frames the raster has completed.
func (m *Machine) FramesCompleted() int { return m.raster.Frames }
