// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
)

// Exibir → Reunir janelas / Trazer janelas para a frente (EightyOne: View →
// Gather Windows / Bring Windows to Front) act on the other windows of the
// emulator: console, on-screen keyboard, listings, keyboard help...

// otherWindows are the open windows besides the main one. A hidden window
// (the on-screen keyboard closed with its X) is skipped.
func otherWindows(a fyne.App, main fyne.Window, hidden func(fyne.Window) bool) []fyne.Window {
	var out []fyne.Window
	for _, w := range a.Driver().AllWindows() {
		if w == main || (hidden != nil && hidden(w)) {
			continue
		}
		out = append(out, w)
	}
	return out
}

// bringToFront raises the other windows over the main one; it returns how
// many there were.
func bringToFront(wins []fyne.Window) int {
	for _, w := range wins {
		w.Show()
		w.RequestFocus()
	}
	return len(wins)
}

// gatherWindows puts the other windows in the middle of the screen (Fyne
// cannot place a window anywhere else), in front: a window lost outside
// the screen comes back.
func gatherWindows(wins []fyne.Window) int {
	for _, w := range wins {
		w.CenterOnScreen()
	}
	return bringToFront(wins)
}

// windowsMessage is the status bar message after one of them.
func windowsMessage(what string, n int) string {
	if n == 0 {
		return "Não há outras janelas abertas"
	}
	return what + ": " + strconv.Itoa(n)
}
