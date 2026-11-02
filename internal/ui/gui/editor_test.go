// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/session"
)

func TestEditorView(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	sess, _ := session.New(session.DefaultSettings())
	if err := sess.Exec(`run 150`, io.Discard); err != nil {
		t.Fatal(err)
	}
	e := newEmulator(sess, 2)
	screen := canvas.NewImageFromImage(e.picture())
	v := newEditorView(e, screen)
	e.editor = v
	if v.handlesKey(fyne.KeyA) {
		t.Fatal("keys go to the ZX81 while the editor is off")
	}
	v.toggle()
	if !v.on || screen.Visible() || !v.img.Visible() || v.img.Image == nil {
		t.Fatal("the editor did not take the place of the screen")
	}
	if !v.handlesKey(fyne.KeyA) || v.handlesKey(fyne.KeyF7) {
		t.Fatal("editor keys")
	}
	for _, r := range `10 print "editor"` {
		v.typedRune(r)
	}
	v.typedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	v.ed.Paste("20 GOTO 10\n") // one line: into the input
	v.typedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	v.toggle() // back: the program goes to the ZX81
	if v.on || !screen.Visible() {
		t.Fatal("back to the emulator")
	}
	deadline := time.Now().Add(30 * time.Second)
	for {
		lines, _ := sess.ProgramLines()
		if basic.List(lines) == "10 PRINT \"EDITOR\"\n20 GOTO 10\n" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("program in the ZX81:\n%s", basic.List(lines))
		}
		time.Sleep(50 * time.Millisecond)
	}
	v.toggle() // the program comes back from the ZX81
	if len(v.ed.Lines) != 2 {
		t.Fatal("program not read again")
	}
}

// TestScreenshotEditor writes docs/img/gui-editor.png when GO81_SCREENSHOTS
// names the output folder.
func TestScreenshotEditor(t *testing.T) {
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
	e := newEmulator(sess, 2)
	screen := canvas.NewImageFromImage(e.picture())
	v := newEditorView(e, screen)
	e.editor = v
	v.toggle()
	v.ed.Paste("10 REM EDITOR DE BASIC DO GO81\n20 LET A=0\n30 FOR I=1 TO 10\n40 PRINT AT I,0;\"LINHA \";I\n50 LET A=A+I\n60 NEXT I\n70 PRINT \"SOMA: \";A\n80 IF A>50 THEN GOTO 100\n90 STOP\n100 PRINT \"FIM\"\n")
	v.ed.Key("Up")
	v.ed.Key("Up")
	for _, r := range "renum 100 5" {
		v.ed.TypeRune(r)
	}
	v.draw()
	r := e.picture().Bounds()
	v.img.SetMinSize(fyne.NewSize(float32(r.Dx()*2), float32(r.Dy()*2)))
	w := a.NewWindow("Go81")
	w.SetContent(container.NewStack(screen, v.img))
	w.Resize(w.Content().MinSize())
	if err := savePNG(filepath.Join(out, "gui-editor.png"), w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
	// The ZX81 keys mode with CTRL held: the keyboard map shows what each key
	// gives with the cursor of the line (F, after SHIFT+ENTER).
	v.ed.Key("Escape")
	v.ed.SetKeysMode(true, v.kb)
	for _, k := range []string{"1", "1", "0", "L", "B"} {
		v.ed.ZXKey(k, false)
	}
	v.ed.ZXKey("L", true)     // =
	v.ed.ZXKey("ENTER", true) // FUNCTION
	v.draw()
	km := newKbOverlay()
	e.keymap = km
	km.setModel("zx81", false)
	km.resize(float32(r.Dx() * 2))
	km.setMode(mapOpaque)
	km.setCtrl(true, e.cursorNow())
	w.SetContent(container.NewStack(screen, v.img, km.box))
	w.Resize(w.Content().MinSize())
	if err := savePNG(filepath.Join(out, "gui-editor-teclas.png"), w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}

func TestEditorZX81Keys(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	sess, _ := session.New(session.DefaultSettings())
	if err := sess.Exec(`run 150`, io.Discard); err != nil {
		t.Fatal(err)
	}
	e := newEmulator(sess, 2)
	e.keymap = newKbOverlay()
	v := newEditorView(e, canvas.NewImageFromImage(e.picture()))
	e.editor = v
	v.toggle()
	v.toggleKeys()
	if !v.keys || !v.ed.KeysMode() || e.cursorNow() != 'K' {
		t.Fatal("F8: ZX81 keys, cursor K for the keyboard map")
	}
	for _, k := range []fyne.KeyName{fyne.Key1, fyne.Key0, fyne.KeyP} { // 10 PRINT
		v.keyDown(k)
	}
	v.shift = true
	v.keyDown(fyne.KeyP) // "
	v.shift = false
	v.keyDown(fyne.KeyO)
	v.shift = true
	v.keyDown(fyne.KeyP)
	v.shift = false
	if v.ed.Input() != `10 PRINT "O"` || e.cursorNow() != 'L' {
		t.Fatalf("input %q cursor %c", v.ed.Input(), e.cursorNow())
	}
	v.shift = true
	v.typedKey(&fyne.KeyEvent{Name: fyne.KeyReturn}) // SHIFT+ENTER: FUNCTION
	v.shift = false
	if e.cursorNow() != 'F' {
		t.Fatal("SHIFT+ENTER should give F")
	}
	v.toggleKeys()
	if v.keys || v.ed.KeysMode() {
		t.Fatal("F8 back to direct typing")
	}
}
