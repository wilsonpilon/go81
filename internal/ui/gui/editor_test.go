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
}
