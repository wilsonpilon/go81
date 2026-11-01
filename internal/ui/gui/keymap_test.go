// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/wilsonpilon/go81/internal/session"
)

func TestKeyboardOverlay(t *testing.T) {
	km := newKbOverlay()
	km.setModel("zx81", false)
	if km.img.Image == nil || km.box.Visible() {
		t.Fatal("starts off, with the picture loaded")
	}
	km.next() // faded
	if !km.box.Visible() || km.img.Translucency != 0.5 {
		t.Fatalf("faded: visible %v, translucency %v", km.box.Visible(), km.img.Translucency)
	}
	km.next() // opaque
	if !km.box.Visible() || km.img.Translucency != 0 {
		t.Fatal("opaque")
	}
	msg := km.next() // only with SHIFT
	if km.box.Visible() || msg != "Mapa do teclado: só com SHIFT apertado (F4 troca)" {
		t.Fatalf("shift mode: %v %q", km.box.Visible(), msg)
	}
	km.setShift(true)
	if !km.box.Visible() {
		t.Fatal("SHIFT down should show it")
	}
	km.setShift(false)
	if km.box.Visible() {
		t.Fatal("SHIFT up should hide it")
	}
	km.setFade(0.75)
	km.next() // off again
	if km.mode != mapOff || km.box.Visible() {
		t.Fatal("back to off")
	}
	km.setMode(mapFaded)
	if km.img.Translucency != 0.75 {
		t.Fatal("fade level kept")
	}
	km.resize(505)
	if s := km.img.MinSize(); s.Width != 505 || s.Height != 223 {
		t.Fatalf("size %v", s)
	}
	km.setModel("tk85", true)
	if km.model != "tk85" {
		t.Fatal("model picture")
	}
}

// TestScreenshotKeyMap writes docs/img/gui-mapa.png (the keyboard map faded
// over the ZX81 screen) when GO81_SCREENSHOTS names the output folder.
func TestScreenshotKeyMap(t *testing.T) {
	dir := os.Getenv("GO81_SCREENSHOTS")
	if dir == "" {
		t.Skip("GO81_SCREENSHOTS não definida")
	}
	out, _ := filepath.Abs(dir)
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(theme.DefaultTheme())
	sess, _ := session.New(session.DefaultSettings())
	if err := sess.Exec(`run 150`, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := sess.Exec(`typeln 10P"MAPA DO TECLADO"`, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := sess.Exec(`run 60`, io.Discard); err != nil {
		t.Fatal(err)
	}
	e := newEmulator(sess, 2)
	screen := canvas.NewImageFromImage(e.picture())
	screen.FillMode = canvas.ImageFillContain
	screen.ScaleMode = canvas.ImageScalePixels
	r := e.picture().Bounds()
	screen.SetMinSize(fyne.NewSize(float32(r.Dx()*2), float32(r.Dy()*2)))
	km := newKbOverlay()
	km.setModel("zx81", false)
	km.resize(float32(r.Dx() * 2))
	km.setMode(mapFaded)
	w := a.NewWindow("Go81")
	w.SetContent(container.NewStack(screen, km.box))
	w.Resize(w.Content().MinSize())
	if err := savePNG(filepath.Join(out, "gui-mapa.png"), w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}
