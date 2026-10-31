// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
)

func TestConsoleWindow(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	sess, err := session.New(session.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	c := newConsoleWindow(a, sess, nil)
	c.show()

	// Tab completes, ↑ brings back the last command.
	c.input.SetText("pe")
	c.key(&fyne.KeyEvent{Name: fyne.KeyTab})
	if c.input.Text != "peek " {
		t.Fatalf("Tab: %q", c.input.Text)
	}
	c.submit("poke 30000 5")
	deadline := time.Now().Add(5 * time.Second)
	for !c.core.Busy() && time.Now().Before(deadline) && !strings.Contains(strings.Join(c.lines, "\n"), "poke 30000 5") {
		time.Sleep(10 * time.Millisecond)
	}
	for c.core.Busy() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	var v byte
	sess.Do(func(m *zx81.Machine) { v = m.Peek(30000) })
	if v != 5 {
		t.Fatalf("poke did not run: %d", v)
	}
	c.input.SetText("")
	c.key(&fyne.KeyEvent{Name: fyne.KeyUp})
	if c.input.Text != "poke 30000 5" {
		t.Fatalf("history: %q", c.input.Text)
	}
	if !strings.Contains(c.out.Text, "go81> poke 30000 5") {
		t.Fatalf("console text: %q", c.out.Text)
	}
}
