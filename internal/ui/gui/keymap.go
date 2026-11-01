// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
)

// The keyboard map (Exibir → Mapa do teclado, F4): the picture of the
// model's keyboard (EightyOne's KBPics: keywords, symbols and functions on
// each key) drawn over the bottom of the ZX81 screen, faded or opaque,
// always or only while SHIFT is held. Ajuda → Teclado (F1) shows the same
// picture in a window of its own.

// Keyboard map modes, in the order F4 goes through them.
const (
	mapOff    = iota // not shown
	mapFaded         // always, see-through
	mapOpaque        // always, solid
	mapShift         // only while SHIFT is held (see-through)
	mapModes
)

var mapModeNames = []string{"desligado", "esmaecido", "opaco", "só com SHIFT apertado"}

// Faded levels (Translucency of the picture: 0 = solid, 1 = invisible).
var mapFadeLevels = []float64{0.25, 0.5, 0.75}

const (
	prefMapMode = "mapa.modo"
	prefMapFade = "mapa.transparencia"
)

type kbOverlay struct {
	img   *canvas.Image
	box   *fyne.Container // what goes over the screen
	mode  int
	fade  float64
	shift bool // a SHIFT key of the PC is down
	model string
	ntsc  bool
}

func newKbOverlay() *kbOverlay {
	km := &kbOverlay{img: canvas.NewImageFromImage(nil), fade: 0.5}
	km.img.FillMode = canvas.ImageFillContain
	km.img.ScaleMode = canvas.ImageScaleSmooth
	km.box = container.NewVBox(layout.NewSpacer(), km.img)
	km.apply()
	return km
}

// setModel loads the picture of the model's keyboard.
func (km *kbOverlay) setModel(model string, ntsc bool) {
	if km.img.Image != nil && model == km.model && ntsc == km.ntsc {
		return
	}
	if img, err := keyboardPicture(model, ntsc); err == nil {
		km.img.Image = img
		km.model, km.ntsc = model, ntsc
		km.img.Refresh()
	}
}

// resize fits the picture to the width of the screen.
func (km *kbOverlay) resize(width float32) {
	km.img.SetMinSize(fyne.NewSize(width, width*223/505))
}

func (km *kbOverlay) visible() bool {
	switch km.mode {
	case mapFaded, mapOpaque:
		return true
	case mapShift:
		return km.shift
	}
	return false
}

func (km *kbOverlay) apply() {
	if km.mode == mapOpaque {
		km.img.Translucency = 0
	} else {
		km.img.Translucency = km.fade
	}
	if km.visible() {
		km.box.Show()
	} else {
		km.box.Hide()
	}
	km.img.Refresh()
}

// setMode chooses a mode and returns a message for the status bar.
func (km *kbOverlay) setMode(m int) string {
	km.mode = (m%mapModes + mapModes) % mapModes
	km.apply()
	return "Mapa do teclado: " + mapModeNames[km.mode] + " (F4 troca)"
}

// next is F4.
func (km *kbOverlay) next() string { return km.setMode(km.mode + 1) }

func (km *kbOverlay) setFade(f float64) {
	km.fade = f
	km.apply()
}

// setShift follows the SHIFT keys of the PC.
func (km *kbOverlay) setShift(down bool) {
	if down == km.shift {
		return
	}
	km.shift = down
	if km.mode == mapShift {
		km.apply()
	}
}

func fadeLabel(f float64) string { return fmt.Sprintf("Transparência %.0f%%", f*100) }

// setKeyMapMode changes the mode, remembers it and says so.
func (e *emulator) setKeyMapMode(m int) {
	msg := e.keymap.setMode(m)
	if e.prefs != nil {
		e.prefs.SetInt(prefMapMode, e.keymap.mode)
	}
	if e.onKeyMapChanged != nil {
		e.onKeyMapChanged()
	}
	e.flash(msg)
}

// keyMapMenu is Exibir → Mapa do teclado.
func (e *emulator) keyMapMenu(refresh func()) *fyne.MenuItem {
	km := e.keymap
	var modes, fades []*fyne.MenuItem
	mark := func() {
		for i, it := range modes {
			it.Checked = i == km.mode
		}
		for i, it := range fades {
			it.Checked = mapFadeLevels[i] == km.fade
		}
		refresh()
	}
	for m := range mapModes {
		label := "Desligado"
		switch m {
		case mapFaded:
			label = "Esmaecido sobre a tela"
		case mapOpaque:
			label = "Opaco sobre a tela"
		case mapShift:
			label = "Só enquanto SHIFT estiver apertado"
		}
		modes = append(modes, fyne.NewMenuItem(label, func() { e.setKeyMapMode(m) }))
	}
	for _, f := range mapFadeLevels {
		fades = append(fades, fyne.NewMenuItem(fadeLabel(f), func() {
			km.setFade(f)
			if e.prefs != nil {
				e.prefs.SetFloat(prefMapFade, f)
			}
			mark()
		}))
	}
	modes[mapFaded].Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF4}
	e.onKeyMapChanged = mark
	item := fyne.NewMenuItem("Mapa do teclado", nil)
	item.ChildMenu = fyne.NewMenu("", append(append(modes, fyne.NewMenuItemSeparator()), fades...)...)
	mark()
	return item
}
