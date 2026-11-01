// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// minHold is the shortest press of an on-screen key: a quick click must
// still last a few frames for the ZX81 to read it.
const minHold = 100 * time.Millisecond

var (
	capColor     = color.NRGBA{0x30, 0x30, 0x30, 0xff}
	capDownColor = color.NRGBA{0x70, 0x70, 0x70, 0xff}
	capText      = color.NRGBA{0xf0, 0xf0, 0xf0, 0xff}
	capShifted   = color.NRGBA{0xff, 0x60, 0x50, 0xff} // red, like the ZX81
	capKeyword   = color.NRGBA{0xc0, 0xc0, 0xc0, 0xff}
)

// keyCap is one key of the on-screen keyboard; it holds the ZX81 key while
// the mouse button is down.
type keyCap struct {
	widget.BaseWidget
	info keyCapInfo
	e    *emulator
	bg   *canvas.Rectangle
	down time.Time
}

func newKeyCap(e *emulator, info keyCapInfo) *keyCap {
	k := &keyCap{info: info, e: e}
	k.ExtendBaseWidget(k)
	return k
}

func (k *keyCap) CreateRenderer() fyne.WidgetRenderer {
	k.bg = canvas.NewRectangle(capColor)
	k.bg.CornerRadius = 4
	k.bg.SetMinSize(fyne.NewSize(62, 58))
	small := func(s string, c color.Color) *canvas.Text {
		t := canvas.NewText(s, c)
		t.TextSize = 10
		t.Alignment = fyne.TextAlignCenter
		return t
	}
	label := k.info.Key
	main := canvas.NewText(label, capText)
	main.TextStyle = fyne.TextStyle{Bold: true}
	main.TextSize = 18
	if len(label) > 1 {
		main.TextSize = 12
	}
	main.Alignment = fyne.TextAlignCenter
	body := container.NewVBox(small(k.info.Shifted, capShifted), main, small(k.info.Keyword, capKeyword))
	return widget.NewSimpleRenderer(container.NewStack(k.bg, container.NewCenter(body)))
}

// Refresh shows whether the key is held.
func (k *keyCap) Refresh() {
	if k.bg != nil {
		k.bg.FillColor = capColor
		if k.e.screenKeyDown(k.info.Key) {
			k.bg.FillColor = capDownColor
		}
		k.bg.Refresh()
	}
	k.BaseWidget.Refresh()
}

func (k *keyCap) MouseDown(*desktop.MouseEvent) {
	k.down = time.Now()
	k.e.screenKey(k.info.Key, true)
}

func (k *keyCap) MouseUp(*desktop.MouseEvent) {
	if k.info.Key == "SHIFT" { // SHIFT locks on the click itself
		return
	}
	release := func() { k.e.screenKey(k.info.Key, false) }
	if wait := minHold - time.Since(k.down); wait > 0 {
		time.AfterFunc(wait, release)
		return
	}
	release()
}

// keyboardWindow is the ZX81 keyboard drawn on the screen (Exibir menu, F3).
type keyboardWindow struct {
	w       fyne.Window
	caps    []*keyCap
	visible bool
}

func newKeyboardWindow(a fyne.App, e *emulator, onKey func(*fyne.KeyEvent, bool)) *keyboardWindow {
	kw := &keyboardWindow{w: a.NewWindow("Teclado do ZX81")}
	rows := container.NewVBox()
	for _, row := range zx81Keyboard {
		line := container.NewHBox()
		for _, info := range row {
			c := newKeyCap(e, info)
			kw.caps = append(kw.caps, c)
			line.Add(c)
		}
		rows.Add(line)
	}
	hint := widget.NewLabel("Clique nas teclas. SHIFT fica preso até a próxima tecla (clique de novo para soltar).")
	kw.w.SetContent(container.NewBorder(nil, hint, nil, nil, container.NewPadded(rows)))
	kw.w.SetFixedSize(true)
	kw.w.SetCloseIntercept(func() {
		kw.visible = false
		kw.w.Hide()
	})
	// The PC keys also work while this window has the focus.
	if dc, ok := kw.w.Canvas().(desktop.Canvas); ok {
		dc.SetOnKeyDown(func(k *fyne.KeyEvent) { onKey(k, true) })
		dc.SetOnKeyUp(func(k *fyne.KeyEvent) { onKey(k, false) })
	}
	e.onScreenKeys = func() { fyne.Do(kw.refresh) }
	return kw
}

func (kw *keyboardWindow) refresh() {
	for _, c := range kw.caps {
		c.Refresh()
	}
}

func (kw *keyboardWindow) show() {
	kw.visible = true
	kw.w.Show()
	kw.w.RequestFocus()
}
