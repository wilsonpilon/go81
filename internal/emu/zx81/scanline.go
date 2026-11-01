// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne zx81/zx81.cpp (zx81_do_scanline, zx81_DrawPixel) —
// Copyright (C) 2003-2025 Michael D Wynne, Paul Farrow and contributors
// (GPL v2 or later).

package zx81

// Sync types of a scanline.
const (
	SyncNone = 0
	SyncH    = 1
	SyncV    = 2
)

// Scanline is what the machine produces for the raster (SCANLINE).
// Pixels hold colour<<4 (two pixels per T-state).
type Scanline struct {
	SyncLen  int
	SyncType int
	Len      int
	Pixels   [4000]byte
}

// Positions count down from tperscanline to 1 along a line (as in EightyOne).
const (
	hsyncDuration        = 16
	backporchDuration    = 16
	hsyncPositionStart   = hsyncDuration
	hsyncPositionEnd     = hsyncPositionStart - hsyncDuration + 1
	hsyncPositionAfter   = hsyncPositionEnd - 1
	hsyncPositionTol     = (7 * hsyncDuration) / 4
	hsyncAcceptanceStart = hsyncPositionTol + hsyncPositionEnd
	hsyncTolPixels       = hsyncPositionTol * 2

	intAckDuration           = 3
	intAckOffsetStart        = 2
	intAckOffsetEnd          = intAckOffsetStart + intAckDuration - 1
	intAckOffsetToHSync      = 16
	intAckPositionStart      = hsyncPositionStart + intAckOffsetToHSync
	intResponsePositionStart = intAckPositionStart + intAckOffsetEnd

	nmiDetectionOffset        = 2
	nmiDetectionPositionStart = hsyncPositionStart - nmiDetectionOffset

	portActiveDuration       = 3
	portActiveDurationPixels = portActiveDuration * 2

	vsyncToleranceMaxQS = 354

	maxScanlineActivePixelLength = 500
	blankColour                  = 0
)

// drawPixel is zx81_DrawPixel with all EightyOne "colourise" options off.
func (m *Machine) drawPixel(position int, colour byte, inOp, outOp bool) {
	if position < hsyncPositionEnd {
		position += m.cfg.TStatesPerLine
	}
	if colour == byte(m.paper<<4) {
		if position == m.nmiWaitPositionAfter {
			m.nmiWaitPositionStart = 0
		}
		vsyncPeriod := inOp || (!m.syncOutputWhite && !outOp)
		switch {
		case position <= hsyncPositionStart && position >= hsyncPositionEnd:
			colour = 0
			if m.allowSound && !m.cfg.ExcludeHSyncs && position == hsyncPositionStart && !vsyncPeriod && m.syncOutputWhite {
				m.snd.set(false, m.frameTStates)
			}
		case position <= m.backporchStart && position >= m.backporchEnd:
			colour = 0
			if m.allowSound && !m.cfg.ExcludeHSyncs && position == m.backporchEnd && !vsyncPeriod && m.syncOutputWhite {
				m.snd.set(true, m.frameTStates)
			}
		case vsyncPeriod:
			colour = 0
		}
	}
	l := &m.line
	l.Pixels[l.Len] = colour
	l.Len++
}

func (m *Machine) drawClockCycle(position int, colour byte) {
	m.drawPixel(position, colour, false, false)
	m.drawPixel(position, colour, false, false)
}

// doScanline is zx81_do_scanline: runs instructions until a scanline is
// complete and returns the T-states used.
func (m *Machine) doScanline() int {
	l := &m.line
	c := m.CPU
	tpl := m.cfg.TStatesPerLine
	tstotal := 0
	scanlinesPerFrame := 0

	hsyncThreshold := m.scanlinePixelLength - hsyncTolPixels
	activeLen := m.scanlinePixelLength

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

		afterInstr := m.lineClockCounter - ts
		overlapsHSync := afterInstr < hsyncPositionStart
		startOfHSync := m.lineClockCounter >= hsyncPositionStart && overlapsHSync
		if m.syncOutputWhite && startOfHSync {
			m.lineCounter = (m.lineCounter + 1) & 7
		}

		m.z80Halted = c.Halted
		nmiDetectable := m.lineClockCounter >= hsyncPositionStart && afterInstr <= nmiDetectionPositionStart
		nmiOnInstruction := m.lastInstruction == lastOutFE
		nmiTurnedOn := !m.nmiGeneratorEnabled && nmiOnInstruction
		checkWait := !m.z80Halted && nmiDetectable && (m.nmiGeneratorEnabled || nmiTurnedOn) && !m.cfg.ImprovedWait

		if checkWait {
			pos := m.lineClockCounter
			nmiOff := m.lastInstruction == lastOutFD
			mc := c.MCycles()
			io := c.IOMCycle()
			for k := 0; k < len(mc) && mc[k] != 0; k++ {
				t2 := pos - 1
				isIO := k == io
				if isIO {
					t2--
				}
				if (t2 <= hsyncPositionStart && !nmiOff) || (isIO && nmiTurnedOn) {
					ts += t2 - hsyncPositionAfter
					m.nmiWaitPositionStart = t2 - 1
					break
				}
				pos -= mc[k]
			}
		}

		intResponse := 0
		if m.interruptPending {
			intResponse = c.Interrupt(m.idleBus)
			if intResponse > 0 {
				if afterInstr != intResponsePositionStart {
					adj := intResponsePositionStart - afterInstr
					m.lineClockCounter += adj
					afterInstr += adj
					adjPixels := adj * 2
					if adjPixels >= -hsyncTolPixels {
						activeLen += adjPixels
						if activeLen > maxScanlineActivePixelLength {
							activeLen = maxScanlineActivePixelLength
						}
					}
				}
				ts += intResponse
			}
			m.interruptPending = false
		}

		pixels := ts << 1
		inFE := m.lastInstruction == lastInFE && !m.nmiGeneratorEnabled
		outFF := (m.lastInstruction == lastOutFF || m.lastInstruction == lastOutFE ||
			m.lastInstruction == lastOutFD) && !m.nmiGeneratorEnabled

		for i := 0; i < pixels; i++ {
			bit := (m.shiftRegister ^ m.shiftRegInv) & 0x8000
			col := m.paper
			if bit != 0 {
				col = m.ink
			}
			inOp := inFE && i >= pixels-portActiveDurationPixels
			outOp := outFF && i >= pixels-portActiveDurationPixels
			m.drawPixel(m.lineClockCounter-i/2, byte(col<<4), inOp, outOp)
			m.shiftRegister <<= 1
			m.shiftRegInv <<= 1
		}

		if c.R&0x40 == 0 {
			m.interruptPending = true
		}

		m.frameTStates += ts
		m.tStatesCount += ts

		if m.Tape != nil {
			m.Tape.ClockTick(ts) // WavClockTick
		}

		prevSyncWhite := m.syncOutputWhite
		switch m.lastInstruction {
		case lastOutFD: // NMI generator off
			m.nmiGeneratorEnabled = false
			m.syncOutputWhite = true
		case lastOutFE: // NMI generator on
			m.nmiGeneratorEnabled = true
			m.syncOutputWhite = true
		case lastOutFF: // VSYNC end
			m.syncOutputWhite = true
		case lastInFE: // VSYNC start
			if !m.nmiGeneratorEnabled {
				if m.syncOutputWhite {
					l.SyncLen = portActiveDuration
				} else {
					l.SyncLen += ts
				}
				m.syncOutputWhite = false
				m.lineCounter = 0
			}
		default:
			if !m.syncOutputWhite {
				l.SyncLen += ts
			}
		}

		if m.syncOutputWhite && !prevSyncWhite {
			portStart := afterInstr + portActiveDuration
			if portStart <= hsyncPositionStart {
				remaining := portStart - hsyncPositionAfter
				switch {
				case remaining >= portActiveDuration:
					l.SyncLen += ts
				case remaining > 0:
					effActive := portActiveDuration - remaining
					l.SyncLen += ts - effActive
				default:
					l.SyncLen += ts - portActiveDuration
				}
				if portStart >= hsyncPositionEnd && l.SyncLen <= hsyncDuration*2 {
					l.SyncLen = 0
				}
			} else {
				l.SyncLen += ts - portActiveDuration
			}

			switch {
			case l.SyncLen > hsyncDuration:
				// VSYNC ended before the HSYNC period: move it to the next frame.
				if m.frameSynchronised && portStart > hsyncPositionStart {
					// EightyOne copies from scanline+len (not from the start).
					n := l.Len
					copy(l.Pixels[activeLen:activeLen+n], l.Pixels[n:2*n])
					for k := 0; k < m.scanlinePixelLength; k++ {
						l.Pixels[k] = blankColour
					}
					m.lineClockCounter = ts - l.Len/2
					l.Len += activeLen
				}
			case l.SyncLen > 0 && portStart+l.SyncLen <= hsyncAcceptanceStart:
				if portStart > hsyncPositionStart {
					l.SyncType = SyncH
				}
			default:
				l.SyncLen = 0
			}
		}

		m.lineClockCounter -= ts

		otherOverlaps := !nmiOnInstruction && startOfHSync
		nmiOnOverlaps := nmiOnInstruction && afterInstr+portActiveDuration >= hsyncPositionEnd && startOfHSync

		if m.nmiGeneratorEnabled && (otherOverlaps || nmiOnOverlaps) {
			nmiDuration := c.NMI()
			waitDuration := 0
			t2 := m.lineClockCounter - 1
			if t2 >= hsyncPositionEnd {
				if !m.cfg.ImprovedWait || m.z80Halted {
					waitDuration = t2 - hsyncPositionAfter
					m.nmiWaitPositionStart = t2 - 1
				}
			}
			actual := nmiDuration + waitDuration
			overhang := hsyncPositionStart - m.lineClockCounter
			remaining := hsyncDuration - overhang

			if m.lineClockCounter >= hsyncPositionEnd {
				m.lineClockCarry += actual - remaining
				if m.lineClockCarry < 0 {
					m.lineClockCarry = 0
				}
			} else {
				m.lineClockCarry += -m.lineClockCounter + actual
			}

			ts += actual
			m.frameTStates += actual
			m.tStatesCount += actual

			col := m.ink
			if m.syncOutputWhite {
				col = m.paper
			}
			for ; actual > 0; actual-- {
				m.drawClockCycle(m.lineClockCounter, byte(col<<4))
				m.lineClockCounter--
			}
		} else if m.lineClockCounter < hsyncPositionEnd {
			m.lineClockCarry = -m.lineClockCounter
		}

		if m.lineClockCounter < hsyncPositionEnd {
			if m.syncOutputWhite {
				if l.SyncLen <= hsyncDuration {
					if !m.frameSynchronised || l.Len >= hsyncThreshold {
						if m.frameSynchronised && l.SyncLen > 0 {
							beyond := -m.lineClockCounter - portActiveDuration
							if beyond > 0 {
								activeLen += beyond * 2
								m.lineClockCarry -= beyond
							}
						}
						l.SyncLen = hsyncDuration
						l.SyncType = SyncH
					}
				} else {
					scanlinesPerFrame = m.raster.y + 1
					l.SyncType = SyncV
				}
			}

			overhang := l.Len - activeLen
			if overhang > 0 {
				copy(m.carryOver[:overhang], l.Pixels[activeLen:activeLen+overhang])
				if l.SyncType == SyncV && overhang/2 > ts {
					scanlinesPerFrame--
				}
				if m.frameSynchronised && activeLen < m.scanlinePixelLength {
					for k := activeLen; k < m.scanlinePixelLength; k++ {
						l.Pixels[k] = blankColour
					}
				}
				l.Len -= overhang
				activeLen = l.Len
			}
			m.lineClockCounter += tpl
		}

		tstotal += ts

		if !(l.Len < activeLen && l.SyncType == SyncNone) {
			break
		}
	}

	if l.SyncType == SyncV {
		ry := m.raster.y
		if m.vsyncFound {
			m.frameSynchronised = ry >= m.raster.vsyncTolMin && ry <= vsyncToleranceMaxQS
			m.vsyncFound = m.frameSynchronised
			if m.frameSynchronised {
				m.ScanlinesPerFrame = scanlinesPerFrame
			} else {
				m.ScanlinesPerFrame = 0
			}
		} else {
			m.vsyncFound = ry >= m.raster.vsyncTolMin && ry <= vsyncToleranceMaxQS
		}
		m.scanlineCounter = 0
		m.allowSound = m.cfg.VSyncSound
	} else {
		if m.scanlineCounter < vsyncToleranceMaxQS {
			m.scanlineCounter++
		} else {
			m.frameSynchronised = false
			m.vsyncFound = false
			m.allowSound = m.cfg.VSyncSound
			m.ScanlinesPerFrame = 0
		}
	}

	l.Len = m.scanlinePixelLength
	return tstotal
}
