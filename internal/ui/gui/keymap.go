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

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// The keyboard map (Exibir → Mapa do teclado, F4): the picture of the
// model's keyboard (EightyOne's KBPics: keywords, symbols and functions on
// each key) drawn over the bottom of the ZX81 screen, faded or opaque,
// always or only while SHIFT is held. Ajuda → Teclado (F1) shows the same
// picture in a window of its own. With the map on, CTRL held shows instead
// what each key gives with the cursor the ZX81 shows now (cursormap.go).

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
	img    *canvas.Image
	cursor *canvas.Image   // the cursor map, while CTRL is held
	box    *fyne.Container // what goes over the screen
	mode   int
	fade   float64
	shift  bool // a SHIFT key of the PC is down
	ctrl   bool // a CTRL key of the PC is down
	mode81 byte // the cursor of the ZX81 (K, L, F, G; 0 = none, as on the ZX80)
	model  string
	ntsc   bool
}

func newKbOverlay() *kbOverlay {
	km := &kbOverlay{img: canvas.NewImageFromImage(nil), fade: 0.5}
	km.img.FillMode = canvas.ImageFillContain
	km.img.ScaleMode = canvas.ImageScaleSmooth
	km.cursor = canvas.NewImageFromImage(nil)
	km.cursor.FillMode = canvas.ImageFillContain
	km.cursor.ScaleMode = canvas.ImageScalePixels
	km.box = container.NewVBox(layout.NewSpacer(), km.img, km.cursor)
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
	km.cursor.SetMinSize(fyne.NewSize(width, width*cursorMapH/cursorMapW))
}

// showsCursor: CTRL is down, the map is on and the machine has a cursor map.
func (km *kbOverlay) showsCursor() bool {
	return km.ctrl && km.mode != mapOff && km.cursor.Image != nil
}

func (km *kbOverlay) visible() bool {
	if km.showsCursor() {
		return true
	}
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
	km.cursor.Translucency = km.img.Translucency
	if km.showsCursor() {
		km.img.Hide()
		km.cursor.Show()
	} else {
		km.cursor.Hide()
		km.img.Show()
	}
	if km.visible() {
		km.box.Show()
	} else {
		km.box.Hide()
	}
	km.box.Refresh() // lay out again: the picture or the cursor map
}

// setMode chooses a mode and returns a message for the status bar.
func (km *kbOverlay) setMode(m int) string {
	km.mode = (m%mapModes + mapModes) % mapModes
	km.apply()
	msg := "Mapa do teclado: " + mapModeNames[km.mode] + " (F4 troca"
	if km.mode != mapOff {
		msg += ", CTRL mostra o cursor K/L/F/G"
	}
	return msg + ")"
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

// setCtrl follows the CTRL keys of the PC; cursor is the cursor of the ZX81
// (0 if the machine has no cursor map).
func (km *kbOverlay) setCtrl(down bool, cursor byte) {
	km.ctrl = down
	km.setCursor(cursor)
	km.apply()
}

// setCursor changes the cursor map while CTRL is held (every frame).
func (km *kbOverlay) setCursor(cursor byte) {
	if cursor == km.mode81 {
		return
	}
	km.mode81 = cursor
	km.cursor.Image = cursorMapFor(cursor)
	km.apply()
}

// cursorNow reads the cursor of the ZX81 (0 on the ZX80: no F and G).
func (e *emulator) cursorNow() byte {
	if e.sess.Settings().Model == "zx80" {
		return 0
	}
	var c byte
	e.sess.Do(func(m *zx81.Machine) { c = cursorMode(m) })
	return c
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
	hint := fyne.NewMenuItem("Segure CTRL: as teclas do cursor atual (K, L, F, G)", nil)
	hint.Disabled = true
	items := append(append(modes, fyne.NewMenuItemSeparator()), fades...)
	items = append(items, fyne.NewMenuItemSeparator(), hint)
	item := fyne.NewMenuItem("Mapa do teclado", nil)
	item.ChildMenu = fyne.NewMenu("", items...)
	mark()
	return item
}
