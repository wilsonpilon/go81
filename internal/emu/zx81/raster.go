// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne AccDraw_.cpp (AccurateDraw, CompleteFrame) —
// Copyright (C) 2003-2025 Michael D Wynne (GPL v2 or later).

package zx81

// Raster geometry and sync tolerances (GDI mode, no TV effects).
const (
	FrameWidth  = 520 // TVW
	FrameHeight = 380 // TVH

	hsyncMinLen = 10
	vsyncMinLen = 350
)

// Raster turns scanlines into frames the way a TV (and EightyOne) does.
type Raster struct {
	// Sync tolerances as set by EightyOne's TV settings (Artifacts dialog,
	// default sliders): VSYNC 283-323 lines (NTSC: 60 fewer), HSYNC 390.
	vsyncTolMin, vsyncTolMax, hsyncTol int

	x, y    int
	row     int
	scanLen int
	work    [FrameWidth * FrameHeight]byte
	// Frame is the last complete frame: one byte per pixel, colour<<4.
	Frame [FrameWidth * FrameHeight]byte
	// Frames counts completed frames.
	Frames int
}

func (r *Raster) init(tstatesPerLine int, ntsc bool) {
	r.vsyncTolMin = 283
	r.vsyncTolMax = r.vsyncTolMin + 40
	r.hsyncTol = 390
	if ntsc {
		r.vsyncTolMin -= 60
		r.vsyncTolMax -= 60
	}
	r.x, r.y, r.row = 0, 0, 0
	r.scanLen = tstatesPerLine * 2
}

func (r *Raster) completeFrame() {
	x, y := r.x, r.y
	for ; y < FrameHeight; y++ {
		for ; x < FrameWidth; x++ {
			r.work[y*FrameWidth+x] = 0
		}
		x = 0
	}
	r.Frame = r.work
	r.Frames++
}

// draw is AccurateDraw for one scanline.
func (r *Raster) draw(l *Scanline) {
	for i := 0; i < l.Len; i++ {
		if r.row < FrameHeight && r.x < FrameWidth {
			r.work[r.row*FrameWidth+r.x] = l.Pixels[i]
		}
		r.x++
		if r.x >= r.scanLen {
			r.x = 0
			r.row++
			r.y++
			if r.y >= FrameHeight {
				i = l.Len + 1
				l.SyncType = SyncV
				if l.SyncLen < hsyncMinLen {
					l.SyncLen = hsyncMinLen
				}
			}
		}
	}
	if l.SyncLen < hsyncMinLen {
		l.SyncType = SyncNone
	}
	if l.SyncType != SyncNone {
		if r.x > r.hsyncTol {
			r.x = 0
			r.y++
			r.row++
		}
		if r.y >= FrameHeight || r.y >= r.vsyncTolMax ||
			(l.SyncLen > vsyncMinLen && r.y > r.vsyncTolMin) {
			r.completeFrame()
			r.x, r.y, r.row = 0, 0, 0
		}
	}
}
