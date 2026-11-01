// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/wilsonpilon/go81/internal/session"
)

// TestScreenshotConsole writes docs/img/gui-console.png (the command
// console of the window) when GO81_SCREENSHOTS names the output folder.
func TestScreenshotConsole(t *testing.T) {
	dir := os.Getenv("GO81_SCREENSHOTS")
	if dir == "" {
		t.Skip("GO81_SCREENSHOTS não definida")
	}
	out, _ := filepath.Abs(dir)
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(theme.DefaultTheme())
	sess, _ := session.New(session.DefaultSettings())
	c := newConsoleWindow(a, sess, nil)
	c.show()
	c.win.Resize(fyne.NewSize(760, 470))
	old, _ := os.Getwd()
	os.Chdir("../../../examples") // so the command shows just "demo.bas"
	defer os.Chdir(old)
	for _, cmd := range []string{"load demo.bas", "list", "regs", "peek $4000 16"} {
		c.submit(cmd)
		for i := 0; i < 500 && c.core.Busy(); i++ {
			time.Sleep(20 * time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond)
	}
	c.input.SetText("scr")
	c.key(&fyne.KeyEvent{Name: fyne.KeyTab}) // shows the choices
	time.Sleep(200 * time.Millisecond)
	img := c.win.Canvas().Capture()
	f, err := os.Create(filepath.Join(out, "gui-console.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

// TestScreenshotKeyboard writes docs/img/gui-teclado.png (the on-screen
// keyboard, with SHIFT locked) when GO81_SCREENSHOTS names the output folder.
func TestScreenshotKeyboard(t *testing.T) {
	dir := os.Getenv("GO81_SCREENSHOTS")
	if dir == "" {
		t.Skip("GO81_SCREENSHOTS não definida")
	}
	out, _ := filepath.Abs(dir)
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(theme.DefaultTheme())
	sess, _ := session.New(session.DefaultSettings())
	e := newEmulator(sess, 0)
	kb := newKeyboardWindow(a, e, nil)
	kb.show()
	kb.w.Resize(kb.w.Content().MinSize())
	e.screenKey("SHIFT", true)
	time.Sleep(100 * time.Millisecond)
	if err := savePNG(filepath.Join(out, "gui-teclado.png"), kb.w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}
