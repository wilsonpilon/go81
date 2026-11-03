// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
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
	// A syntax error found by the ROM: 115 LET =1, the S on the "=".
	km.setCtrl(false, 0)
	km.setMode(mapOff)
	v.ed.Key("Escape")
	for _, k := range []string{"1", "1", "5", "L"} {
		v.ed.ZXKey(k, false)
	}
	v.ed.ZXKey("L", true) // =
	v.ed.ZXKey("1", false)
	v.ed.ZXKey("ENTER", false)
	v.draw()
	if v.ed.Cursor() != 'S' {
		t.Fatal("no S cursor")
	}
	if err := savePNG(filepath.Join(out, "gui-editor-erro.png"), w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
	// Back to direct typing. MEM: the size of the program and the free memory; the listing scrolled
	// by Ctrl+Down, the > line still 80.
	v.ed.Key("Escape")
	v.ed.SetKeysMode(false, v.kb)
	w.SetContent(container.NewStack(screen, v.img))
	for _, r := range "mem" {
		v.ed.TypeRune(r)
	}
	v.ed.Key("Return")
	v.ctrl = true
	v.scrollKey(fyne.KeyDown)
	v.scrollKey(fyne.KeyDown)
	v.ctrl = false
	for _, r := range `save "jogo.p"` {
		v.ed.TypeRune(r)
	}
	v.draw()
	if err := savePNG(filepath.Join(out, "gui-editor-mem.png"), w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
	v.ed.Key("Escape")
}

// The editor opens, joins and saves files, and Ctrl+arrows scroll.
func TestEditorFilesAndScroll(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	sess, _ := session.New(session.DefaultSettings())
	if err := sess.Exec(`run 150`, io.Discard); err != nil {
		t.Fatal(err)
	}
	e := newEmulator(sess, 2)
	v := newEditorView(e, canvas.NewImageFromImage(e.picture()))
	e.editor = v
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "um.bas"), []byte("10 PRINT 1\n20 GOTO 10\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "dois.txt"), []byte("20 PRINT 2\n30 STOP\n"), 0o644)
	v.openFile(filepath.Join(dir, "um.bas"), false) // enters the editor
	if !v.on || len(v.ed.Lines) != 2 || v.ed.RAMK != 16 {
		t.Fatalf("open: on %v lines %d", v.on, len(v.ed.Lines))
	}
	v.openFile(filepath.Join(dir, "dois.txt"), true)
	if got := basic.List(v.ed.Lines); got != "10 PRINT 1\n20 PRINT 2\n30 STOP\n" {
		t.Fatalf("merge:\n%s", got)
	}
	// SAVE typed in capitals finds the folder; MERGE of a name in another
	// case finds the file.
	v.saveFile(filepath.Join(dir, "tres.p"))
	lines, err := session.ReadProgramLines(filepath.Join(dir, "tres.p"))
	if err != nil || len(lines) != 3 {
		t.Fatalf("save: %v %d", err, len(lines))
	}
	if got := anyCase(filepath.Join(dir, "TRES.P")); !strings.EqualFold(got, filepath.Join(dir, "tres.p")) {
		t.Fatalf("anyCase = %q", got)
	} else if _, err := os.Stat(got); err != nil {
		t.Fatalf("anyCase: %v", err)
	}
	cur := v.ed.Current
	v.ctrl = true
	if !v.scrollKey(fyne.KeyDown) {
		t.Fatal("Ctrl+Down not taken")
	}
	v.typedKey(&fyne.KeyEvent{Name: fyne.KeyDown}) // the same key, typed: ignored
	v.ctrl = false
	if v.ed.Current != cur {
		t.Fatal("scrolling changed the current line")
	}
	if v.scrollKey(fyne.KeyDown) {
		t.Fatal("Down without Ctrl scrolled")
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

func TestEditorRun(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	sess, _ := session.New(session.DefaultSettings())
	if err := sess.Exec(`run 150`, io.Discard); err != nil {
		t.Fatal(err)
	}
	e := newEmulator(sess, 2)
	v := newEditorView(e, canvas.NewImageFromImage(e.picture()))
	e.editor = v
	v.toggle()
	v.ed.Paste("10 POKE 30000,42\n20 STOP\n")
	for _, r := range "run" {
		v.typedRune(r)
	}
	v.typedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	if v.on {
		t.Fatal("RUN should go back to the emulator")
	}
	deadline := time.Now().Add(60 * time.Second)
	for {
		var got byte
		sess.Do(func(m *zx81.Machine) { got = m.Peek(30000) })
		if got == 42 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the program did not run: PEEK 30000 = %d", got)
		}
		sess.Exec("run 10", io.Discard)
		time.Sleep(20 * time.Millisecond)
	}
}
