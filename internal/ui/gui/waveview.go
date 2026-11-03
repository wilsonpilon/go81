// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

// Ferramentas → Fita em som… (the waveform of EightyOne's Wave Loader):
// the sound of the tape drawn as a wave, with where it is playing. It
// shows the WAV played into the ZX81, the SAVE recorded as sound
// (Opções → Fita → Destino → Carregador WAV) or the tape as it sounds.

const (
	waveSourceAuto = iota
	waveSourceIn
	waveSourceRec
	waveSourceTape
)

var waveSourceNames = []string{"Automático", "Som na entrada (carregador WAV)", "SAVE gravado como som", "A fita como som"}

// Colours of the waveform (the Wave Loader's: green on black).
var (
	waveBack   = color.RGBA{0x10, 0x14, 0x10, 0xFF}
	waveMid    = color.RGBA{0x30, 0x48, 0x30, 0xFF}
	waveLine   = color.RGBA{0x40, 0xE0, 0x60, 0xFF}
	waveCursor = color.RGBA{0xFF, 0x50, 0x40, 0xFF}
	waveThresh = color.RGBA{0x80, 0x80, 0x30, 0xFF}
)

// renderWave draws samples from..from+count of a wave in w×h pixels: each
// column is the lowest and highest sample under it. pos is the playing
// sample (-1 = none) and thresh the loader's threshold (0 = none).
func renderWave(wv tape.Wave, from, count, w, h, pos, thresh int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 0xFF
	}
	fill := func(x, y0, y1 int, c color.RGBA) {
		for y := max(0, y0); y <= min(h-1, y1); y++ {
			img.SetRGBA(x, y, c)
		}
	}
	yOf := func(s int) int { return (255 - s) * (h - 1) / 255 }
	for x := range w {
		fill(x, 0, h-1, waveBack)
		img.SetRGBA(x, yOf(128), waveMid)
		if thresh > 0 {
			img.SetRGBA(x, yOf(128-thresh), waveThresh)
		}
	}
	if count <= 0 || wv == nil {
		return img
	}
	for x := range w {
		a := from + x*count/w
		b := from + (x+1)*count/w
		if b <= a {
			b = a + 1
		}
		if a >= wv.Len() {
			break
		}
		lo, hi := 255, 0
		for i := a; i < min(b, wv.Len()); i++ {
			s := wv.Sample(i)
			lo, hi = min(lo, s), max(hi, s)
		}
		fill(x, yOf(hi), yOf(lo), waveLine)
	}
	if pos >= from && pos < from+count {
		fill((pos-from)*w/count, 0, h-1, waveCursor)
	}
	return img
}

// waveItem opens the waveform window (one at a time).
func (e *emulator) waveItem(a fyne.App) *fyne.MenuItem {
	var win fyne.Window
	return fyne.NewMenuItem("Fita em som (forma de onda)…", func() {
		if win != nil {
			win.RequestFocus()
			return
		}
		win = e.waveWindow(a, func() { win = nil })
		win.Show()
	})
}

// pickWave is the wave a source shows now, and its name.
func (e *emulator) pickWave(source int) (tape.Wave, string, int) {
	var wv tape.Wave
	name, thresh := "", 0
	e.sess.Do(func(m *zx81.Machine) {
		d := m.Tape
		src := source
		if src == waveSourceAuto {
			switch {
			case d.Wav != nil && !d.In:
				src = waveSourceIn
			case d.Rec != nil && (d.Rec.Recording() || d.Rec.Seconds() > 0):
				src = waveSourceRec
			default:
				src = waveSourceTape
			}
		}
		switch src {
		case waveSourceIn:
			if d.Wav != nil {
				wv, thresh = d.Wav.Wave(), d.Wav.Threshold()
			}
		case waveSourceRec:
			if d.Rec != nil {
				wv = d.Rec.Wave()
			}
		case waveSourceTape:
			wv = tape.BlocksWave(d.Player.Blocks)
		}
		name = waveSourceNames[src]
	})
	return wv, name, thresh
}

// The zoom choices: how much sound one window width shows.
var waveZooms = []struct {
	label string
	secs  float64 // 0 = all
}{{"5 ms", 0.005}, {"20 ms", 0.02}, {"100 ms", 0.1}, {"0,5 s", 0.5}, {"2 s", 2}, {"10 s", 10}, {"Tudo", 0}}

func (e *emulator) waveWindow(a fyne.App, closed func()) fyne.Window {
	win := a.NewWindow("Fita em som — Go81")
	const w, h = 760, 200
	img := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, w, h)))
	img.FillMode = canvas.ImageFillStretch
	img.ScaleMode = canvas.ImageScalePixels
	img.SetMinSize(fyne.NewSize(w, h))
	source := waveSourceAuto
	zoom := 2 // 100 ms
	follow := true
	var wv tape.Wave
	var name string
	var thresh int
	info := widget.NewLabel("")
	scroll := widget.NewSlider(0, 1)
	scroll.Step = 0
	var draw func()
	reload := func() {
		wv, name, thresh = e.pickWave(source)
		draw()
	}
	window := func() (count int) {
		if wv == nil || wv.Len() == 0 {
			return 0
		}
		if z := waveZooms[zoom].secs; z > 0 {
			return max(1, min(wv.Len(), int(z*float64(wv.Rate()))))
		}
		return wv.Len()
	}
	draw = func() {
		count := window()
		pos := -1
		total := 0
		if wv != nil {
			pos, total = wv.Position(), wv.Len()
		}
		from := 0
		if count > 0 && total > count {
			if follow && pos >= 0 {
				from = max(0, min(total-count, pos-count/2))
				scroll.Value = float64(from) / float64(total-count)
				scroll.Refresh()
			} else {
				from = int(scroll.Value * float64(total-count))
			}
		}
		img.Image = renderWave(wv, from, count, w, h, pos, thresh)
		img.Refresh()
		rate := 44100
		if wv != nil {
			rate = wv.Rate()
		}
		at := ""
		if pos >= 0 {
			at = fmt.Sprintf(" · tocando em %.2f s", float64(pos)/float64(rate))
		}
		info.SetText(fmt.Sprintf("%s · %.2f s de som · na tela: %.3f a %.3f s%s", name,
			float64(total)/float64(rate), float64(from)/float64(rate), float64(from+count)/float64(rate), at))
	}
	scroll.OnChanged = func(float64) {
		if !follow {
			draw()
		}
	}
	src := widget.NewSelect(waveSourceNames, func(s string) {
		for i, n := range waveSourceNames {
			if n == s {
				source = i
			}
		}
		reload()
	})
	src.SetSelectedIndex(0)
	var zl []string
	for _, z := range waveZooms {
		zl = append(zl, z.label)
	}
	zsel := widget.NewSelect(zl, func(s string) {
		for i, z := range waveZooms {
			if z.label == s {
				zoom = i
			}
		}
		draw()
	})
	zsel.SetSelectedIndex(zoom)
	fol := widget.NewCheck("Acompanhar", func(on bool) { follow = on; draw() })
	fol.SetChecked(follow)
	top := container.NewHBox(widget.NewLabel("Som:"), src, widget.NewLabel("Na tela:"), zsel, fol,
		widget.NewButton("Atualizar", reload))
	win.SetContent(container.NewBorder(top, container.NewVBox(scroll, info), nil, nil, img))
	reload()
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(150 * time.Millisecond)
		defer t.Stop()
		n := 0
		for {
			select {
			case <-stop:
				return
			case <-t.C:
			}
			n++
			fyne.Do(func() {
				// The tape as sound is rebuilt only every few seconds (it is
				// long to make); the WAV and the recording follow at once.
				if wv == nil || wv.Position() >= 0 || n%20 == 0 {
					reload()
				}
			})
		}
	}()
	win.SetOnClosed(func() {
		close(stop)
		closed()
	})
	return win
}
