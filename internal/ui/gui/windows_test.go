// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestOtherWindows(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	main := a.NewWindow("main")
	con := a.NewWindow("console")
	kb := a.NewWindow("teclado")
	for _, w := range []fyne.Window{main, con, kb} {
		w.Show()
	}
	has := func(list []fyne.Window, w fyne.Window) bool {
		for _, x := range list {
			if x == w {
				return true
			}
		}
		return false
	}
	kbHidden := true
	hidden := func(w fyne.Window) bool { return w == kb && kbHidden }
	got := otherWindows(a, main, hidden)
	if has(got, main) || has(got, kb) || !has(got, con) {
		t.Fatalf("hidden keyboard: main %v kb %v con %v", has(got, main), has(got, kb), has(got, con))
	}
	kbHidden = false
	got = otherWindows(a, main, hidden)
	if !has(got, kb) || gatherWindows(got) != len(got) {
		t.Fatal("shown keyboard not gathered")
	}
	con.Close()
	if has(otherWindows(a, main, hidden), con) {
		t.Fatal("closed console still listed")
	}
	if m := windowsMessage("Janelas reunidas", 0); m != "Não há outras janelas abertas" {
		t.Fatal(m)
	}
	if m := windowsMessage("Janelas reunidas", 2); m != "Janelas reunidas: 2" {
		t.Fatal(m)
	}
}
