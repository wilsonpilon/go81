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

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
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
	if km.box.Visible() || msg != "Mapa do teclado: só com SHIFT apertado (F4 troca, CTRL mostra o cursor K/L/F/G)" {
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
	// CTRL shows the cursor map instead of the picture, in every mode but off.
	km.setCtrl(true, 'F')
	if !km.box.Visible() || km.img.Visible() || !km.cursor.Visible() || km.cursor.Translucency != 0.75 {
		t.Fatal("CTRL should show the cursor map, faded")
	}
	km.setCursor('G')
	if km.cursor.Image != cursorMapFor('G') {
		t.Fatal("the cursor map follows the cursor")
	}
	km.setCtrl(false, 0)
	if !km.img.Visible() || km.cursor.Visible() {
		t.Fatal("CTRL up: the keyboard picture again")
	}
	km.setMode(mapOff)
	km.setCtrl(true, 'K')
	if km.box.Visible() {
		t.Fatal("CTRL does nothing with the map off")
	}
	km.setCtrl(false, 0)
	km.setMode(mapShift)
	km.setCtrl(true, 0) // ZX80: no cursor map
	if km.box.Visible() {
		t.Fatal("no cursor map on the ZX80")
	}
}

func label(b []byte) string {
	s := ""
	for _, c := range b {
		s += string(chars81[c&0x3F])
	}
	return s
}

func TestCursorKeys(t *testing.T) {
	rom, err := assets.ROM("zx81.edition3")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		cursor     byte
		key        string
		main, shft string
	}{
		{'K', "Q", "PLOT", `""`}, {'K', "P", "PRINT", `"`}, {'K', "1", "1", "EDIT"},
		{'K', "A", "NEW", "STOP"}, {'K', "ENTER", "NEWLINE", "FUNC"},
		{'L', "Q", "Q", `""`}, {'L', "H", "H", "**"}, {'L', "SPACE", " ", "£"},
		{'F', "O", "PEEK", ""}, {'F', "L", "USR", ""}, {'F', "B", "INKEY$", ""}, {'F', "1", "", ""},
		{'G', "9", "9", "GRAPH"}, {'G', "0", "0", "RUBOUT"},
	} {
		got := cursorKeys(rom, c.cursor)[c.key]
		if label(got[0]) != c.main || label(got[1]) != c.shft {
			t.Errorf("%c %s: %q %q, want %q %q", c.cursor, c.key, label(got[0]), label(got[1]), c.main, c.shft)
		}
	}
	g := cursorKeys(rom, 'G')
	if a := g["A"]; a[0][0] != 0xA6 || a[1][0] != 0x08 { // inverse A; the grey graphic
		t.Errorf("G A: %x %x", a[0], a[1])
	}
	if p := g["P"]; p[1][0] != 0x8B { // inverse quote
		t.Errorf("G SHIFT+P: %x", p[1])
	}
	for _, c := range []byte("KLFG") {
		if img := cursorMapFor(c); img == nil || img.Bounds().Dx() != cursorMapW {
			t.Fatalf("no picture for %c", c)
		}
	}
	if cursorMapFor('X') != nil {
		t.Fatal("picture for an unknown cursor")
	}
}

func TestCursorMode(t *testing.T) {
	sess, _ := session.New(session.DefaultSettings())
	cursor := func() byte {
		var c byte
		sess.Do(func(m *zx81.Machine) { c = cursorMode(m) })
		return c
	}
	if err := sess.Exec(`run 150`, io.Discard); err != nil {
		t.Fatal(err)
	}
	if c := cursor(); c != 'K' {
		t.Fatalf("after boot: %c", c)
	}
	if err := sess.Exec(`type 10P`, io.Discard); err != nil { // 10 PRINT, then the L cursor
		t.Fatal(err)
	}
	if err := sess.Exec(`run 20`, io.Discard); err != nil {
		t.Fatal(err)
	}
	if c := cursor(); c != 'L' {
		t.Fatalf("after 10 PRINT: %c", c)
	}
	sess.Do(func(m *zx81.Machine) { m.Poke(0x4006, 1) })
	if c := cursor(); c != 'F' {
		t.Fatalf("MODE 1: %c", c)
	}
	sess.Do(func(m *zx81.Machine) { m.Poke(0x4006, 4) })
	if c := cursor(); c != 'G' {
		t.Fatalf("MODE 4: %c", c)
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
	// CTRL held: the keys of the cursor the ZX81 shows (K, after the ENTER).
	km.setMode(mapOpaque)
	km.setCtrl(true, e.cursorNow())
	w.Resize(w.Content().MinSize())
	if err := savePNG(filepath.Join(out, "gui-mapa-cursor.png"), w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}
