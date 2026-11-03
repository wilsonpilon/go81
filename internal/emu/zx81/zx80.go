// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne zx81/zx81.cpp (zx80_do_scanline, zx80_DrawPixel) —
// Copyright (C) 2003-2025 Michael D Wynne, Paul Farrow and contributors
// (GPL v2 or later).

package zx81

// The ZX80 makes its picture with the same CPU tricks as the ZX81, but it
// has no NMI generator: the ROM itself counts the lines. Three flip-flops
// (EightyOne's videoFlipFlop1Q-3Q) follow the M1 cycles and the IN/OUT of
// the sync: when flip-flop 3 is low, the TV sees a sync pulse.

const (
	zx80IdleDataBus = 0x40 // idleDataBus80

	zx80HSyncDuration            = 20
	zx80HSyncAcceptanceDuration  = (3 * zx80HSyncDuration) / 2
	zx80HSyncAcceptancePixels    = zx80HSyncAcceptanceDuration * 2
	zx80MaxScanlineOverhangPixel = zx80HSyncDuration * 2 * 2
)

// zx80Video is the ZX80 video state.
type zx80Video struct {
	ff1, ff2, ff3 bool // videoFlipFlop1Q, 2Q, 3Q
	ff3Clear      bool // videoFlipFlop3Clear
}

func (v *zx80Video) reset() {
	v.ff1, v.ff2, v.ff3, v.ff3Clear = true, false, false, false
}

// drawPixelZX80 is zx80_DrawPixel with all "colourise" options off: the
// paper turns black during the sync pulses.
func (m *Machine) drawPixelZX80(colour byte, inOp bool) {
	if colour == byte(m.paper<<4) {
		vsyncPeriod := inOp || !m.zx80.ff3Clear
		if vsyncPeriod || !m.zx80.ff3 {
			colour = 0
		}
	}
	l := &m.line
	l.Pixels[l.Len] = colour
	l.Len++
}

// doScanlineZX80 is zx80_do_scanline.
func (m *Machine) doScanlineZX80() int {
	l := &m.line
	c := m.CPU
	v := &m.zx80
	tstotal := 0
	scanlinesPerFrame := 0
	acceptancePos := m.scanlinePixelLength - zx80HSyncAcceptancePixels
	threshold := m.scanlinePixelLength + zx80MaxScanlineOverhangPixel

	l.Len = 0
	if l.SyncType != SyncNone {
		l.SyncType = SyncNone
		l.SyncLen = 0
	}
	if m.lineClockCarry > 0 {
		n := m.lineClockCarry * 2
		copy(l.Pixels[:n], m.carryOver[:n])
		l.Len = n
		m.lineClockCarry = 0
	}

	for {
		m.lastInstruction = lastNone
		m.patchTest()
		if m.StepHook != nil {
			m.StepHook()
		}
		ts := c.Step()

		prevFF3 := v.ff3
		for i := 0; i < c.M1Cycles(); i++ {
			if v.ff3Clear {
				v.ff3 = v.ff2
			}
			v.ff2 = !v.ff1
		}
		if !v.ff3 {
			v.ff1 = false
			if prevFF3 {
				m.lineCounter = (m.lineCounter + 1) & 7
			}
		}

		m.z80Halted = c.Halted
		intResponse := 0
		if m.interruptPending {
			intResponse = c.Interrupt(m.idleBus)
			if intResponse > 0 {
				if v.ff3Clear {
					v.ff3 = v.ff2
				}
				v.ff2 = !v.ff1
				v.ff1 = true
				ts += intResponse
			}
			m.interruptPending = false
		}

		pixels := ts << 1
		inFE := m.lastInstruction == lastInFE
		for i := 0; i < pixels; i++ {
			bit := (m.shiftRegister ^ m.shiftRegInv) & 0x8000
			col := m.paper
			if bit != 0 {
				col = m.ink
			}
			inOp := inFE && i >= pixels-portActiveDurationPixels
			m.drawPixelZX80(byte(col<<4), inOp)
			m.shiftRegister <<= 1
			m.shiftRegInv <<= 1
		}

		if c.R&0x40 == 0 {
			m.interruptPending = true
		}

		m.frameTStates += ts
		m.tStatesCount += ts
		if m.Tape != nil {
			m.Tape.ClockTickMic(ts, !v.ff3) // WavClockTick(ts, !videoFlipFlop3Q)
		}

		switch m.lastInstruction {
		case lastOutFD, lastOutFE, lastOutFF: // VSYNC end
			v.ff1 = false
			v.ff2 = true
			v.ff3Clear = true
			if !v.ff3 {
				l.SyncLen += ts
				if l.SyncLen > zx80HSyncAcceptanceDuration {
					v.ff3 = true
				}
			}
		case lastInFE: // VSYNC start
			if v.ff3 {
				l.SyncLen = portActiveDuration
			} else {
				l.SyncLen += ts
			}
			v.ff1 = true
			v.ff3Clear = false
			v.ff3 = false
			m.lineCounter = 0
		default:
			if !v.ff3 {
				l.SyncLen += ts
			}
		}

		if m.allowSound && prevFF3 != v.ff3 {
			m.snd.set(v.ff3, m.frameTStates)
		}

		if v.ff3 && l.SyncLen > 0 {
			if l.SyncLen <= zx80HSyncAcceptanceDuration {
				l.SyncType = SyncH
			} else {
				scanlinesPerFrame = m.raster.y + 1
				l.SyncType = SyncV
			}

			if l.SyncType == SyncV {
				overhang := l.Len - m.scanlinePixelLength
				switch {
				case overhang < 0:
					if l.Len >= acceptancePos {
						for k := l.Len; k < m.scanlinePixelLength; k++ {
							l.Pixels[k] = blankColour
						}
						m.lineClockCarry = 0
						l.Len = m.scanlinePixelLength
					} else {
						copy(m.carryOver[:l.Len], l.Pixels[:l.Len])
						for k := 0; k < m.scanlinePixelLength; k++ {
							l.Pixels[k] = blankColour
						}
						m.lineClockCarry = l.Len / 2
						l.Len = m.scanlinePixelLength
						scanlinesPerFrame--
					}
				case overhang > 0:
					copy(m.carryOver[:overhang], l.Pixels[m.scanlinePixelLength:m.scanlinePixelLength+overhang])
					m.lineClockCarry = overhang / 2
					l.Len = m.scanlinePixelLength
				}
			} else if l.Len >= acceptancePos {
				actual := l.Len - pixels
				copy(m.carryOver[:pixels], l.Pixels[actual:actual+pixels])
				for k := actual; k < m.scanlinePixelLength; k++ {
					l.Pixels[k] = blankColour
				}
				m.lineClockCarry = ts
				l.Len = m.scanlinePixelLength
			}
		}

		tstotal += ts
		if !(l.Len < threshold && l.SyncType == SyncNone) {
			break
		}
	}

	if l.SyncType == SyncV {
		ry := m.raster.y
		if m.vsyncFound {
			m.frameSynchronised = ry >= m.raster.vsyncTolMin && ry <= vsyncToleranceMaxQS &&
				m.scanlineCounter >= m.raster.vsyncTolMin && m.scanlineCounter <= vsyncToleranceMaxQS
			m.vsyncFound = m.frameSynchronised
			if m.frameSynchronised {
				m.ScanlinesPerFrame = scanlinesPerFrame
			} else {
				m.ScanlinesPerFrame = 0
			}
		} else {
			m.vsyncFound = m.scanlineCounter >= m.raster.vsyncTolMin && m.scanlineCounter <= vsyncToleranceMaxQS
		}
		m.scanlineCounter = 0
		m.allowSound = m.cfg.VSyncSound
		if !m.vsyncFound {
			l.SyncType = SyncNone
			l.SyncLen = 0
		}
	} else {
		if m.scanlineCounter < vsyncToleranceMaxQS && l.SyncType == SyncH {
			m.scanlineCounter++
		}
		if (l.SyncType == SyncNone && v.ff3) || m.scanlineCounter == vsyncToleranceMaxQS {
			m.frameSynchronised = false
			m.vsyncFound = false
			m.allowSound = m.cfg.VSyncSound
			m.ScanlinesPerFrame = 0
		}
		if l.SyncType == SyncNone {
			if overhang := l.Len - m.scanlinePixelLength; overhang > 0 {
				copy(m.carryOver[:overhang], l.Pixels[m.scanlinePixelLength:m.scanlinePixelLength+overhang])
				m.lineClockCarry = overhang / 2
				l.Len = m.scanlinePixelLength
			}
		}
	}
	return tstotal
}
