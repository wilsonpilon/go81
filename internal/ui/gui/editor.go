// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/basiced"
)

// The BASIC editor in the style of the ZX81 (Exibir → Editor de BASIC,
// F7): it takes the place of the ZX81 screen in the same window, drawn
// with the ZX81's characters. F7 again goes back to the emulator and, if
// the program changed, loads it into the ZX81.

type editorView struct {
	e      *emulator
	img    *canvas.Image
	screen fyne.CanvasObject // the ZX81 screen, hidden while editing
	ed     *basiced.Editor
	on     bool
	font   []byte
	kb     *basiced.Keyboard // the ROM keyboard tables (ZX81 keys mode)
	keys   bool              // the ZX81 keys mode (F8)
	shift  bool              // a SHIFT key of the PC is down
	last   int               // the current line when the editor was left
}

func newEditorView(e *emulator, screen fyne.CanvasObject) *editorView {
	v := &editorView{e: e, screen: screen, img: canvas.NewImageFromImage(nil)}
	v.img.FillMode = canvas.ImageFillContain
	v.img.ScaleMode = canvas.ImageScalePixels
	v.img.Hide()
	if rom, err := assets.ROM("zx81.edition3"); err == nil {
		v.font = rom[romFont : romFont+512]
		v.kb, _ = basiced.KeyboardFromROM(rom)
	}
	return v
}

// toggle enters or leaves the editor.
func (v *editorView) toggle() {
	if v.on {
		v.leave()
		return
	}
	lines, err := v.e.sess.ProgramLines()
	if err != nil {
		v.e.flash(err.Error())
		return
	}
	v.e.releaseAll() // no ZX81 key stays down
	v.ed = basiced.New(lines)
	v.ed.SetKeysMode(v.keys, v.kb)
	if v.last > 0 {
		for _, l := range lines {
			if l.Num == v.last {
				v.ed.Current = v.last
			}
		}
	}
	v.on = true
	v.screen.Hide()
	v.img.Show()
	v.draw()
	v.e.flash("Editor de BASIC · " + v.modeName() + " · F7 volta ao emulador · F8 troca o modo")
}

// prefEditorKeys remembers the mode of the editor.
const prefEditorKeys = "editor.teclas"

func (v *editorView) modeName() string {
	if v.keys {
		return "teclas do ZX81 (K, L, F, G)"
	}
	return "digitação direta"
}

// toggleKeys is F8: direct typing or the ZX81 keys.
func (v *editorView) toggleKeys() {
	if !v.on || v.kb == nil {
		return
	}
	v.keys = !v.keys
	v.ed.SetKeysMode(v.keys, v.kb)
	if v.e.prefs != nil {
		v.e.prefs.SetBool(prefEditorKeys, v.keys)
	}
	v.draw()
	v.e.flash("Editor de BASIC: " + v.modeName() + " (F8 troca)")
}

// zxKeys are the PC keys that are ZX81 keys in the ZX81 keys mode.
var zxKeys = map[fyne.KeyName]string{fyne.KeySpace: "SPACE", fyne.KeyPeriod: "."}

// keyDown is a PC key pressed while the editor is on: in the ZX81 keys
// mode, letters, digits, SPACE and "." are ZX81 keys (with SHIFT or not).
func (v *editorView) keyDown(name fyne.KeyName) {
	if !v.on || !v.keys {
		return
	}
	k := zxKeys[name]
	if s := string(name); len(s) == 1 && (s[0] >= 'A' && s[0] <= 'Z' || s[0] >= '0' && s[0] <= '9') {
		k = s
	}
	if k != "" {
		v.ed.ZXKey(k, v.shift)
		v.draw()
	}
}

// cursor is the cursor of the editor (for the keyboard map with CTRL).
func (v *editorView) cursor() byte { return v.ed.Cursor() }

func (v *editorView) leave() {
	v.on = false
	v.img.Hide()
	v.screen.Show()
	v.last = v.ed.Current
	if !v.ed.Changed {
		v.e.flash("Emulador (o programa não mudou)")
		return
	}
	lines := v.ed.Lines
	v.e.flash("Enviando o programa ao ZX81…")
	go func() {
		err := v.e.sess.SetProgram(lines)
		fyne.Do(func() {
			if err != nil {
				v.e.flash("O programa não foi para o ZX81: " + err.Error())
				return
			}
			v.e.flash("Programa do editor carregado no ZX81")
		})
	}()
}

// typedRune and typedKey take the keyboard while the editor is on.
func (v *editorView) typedRune(r rune) {
	if v.on {
		v.ed.TypeRune(r)
		v.draw()
	}
}

func (v *editorView) typedKey(k *fyne.KeyEvent) {
	if !v.on {
		return
	}
	name := string(k.Name)
	if name == "KP_Enter" {
		name = "Return"
	}
	if v.keys && name == "Return" && v.shift {
		v.ed.ZXKey("ENTER", true) // SHIFT+ENTER: FUNCTION
		v.draw()
		return
	}
	v.ed.Key(name)
	v.draw()
}

// handlesKey says whether a key of the window belongs to the editor (so it
// is not a ZX81 key): all but the F keys and the modifiers.
func (v *editorView) handlesKey(name fyne.KeyName) bool {
	if !v.on {
		return false
	}
	switch name {
	case fyne.KeyF1, fyne.KeyF2, fyne.KeyF3, fyne.KeyF4, fyne.KeyF7, fyne.KeyF8, fyne.KeyF11,
		desktop.KeyShiftLeft, desktop.KeyShiftRight, desktop.KeyControlLeft, desktop.KeyControlRight:
		return false
	}
	return true
}

var (
	edPaper = color.RGBA{255, 255, 255, 255}
	edInk   = color.RGBA{0, 0, 0, 255}
)

// draw paints the editor screen, the size of the ZX81 picture (with its
// border), black on white as on the ZX81.
func (v *editorView) draw() {
	if v.ed == nil || v.font == nil {
		return
	}
	b := v.e.picture().Bounds()
	img := image.NewRGBA(b)
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = 255, 255, 255, 255
	}
	x0 := b.Min.X + (b.Dx()-basiced.Cols*8)/2
	y0 := b.Min.Y + (b.Dy()-basiced.Rows*8)/2
	for r, row := range v.ed.Screen() {
		for c, code := range row {
			drawCodes(img, v.font, x0+c*8, y0+r*8, []byte{code}, edInk)
		}
	}
	v.img.Image = img
	v.img.Refresh()
}
