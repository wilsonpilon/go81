// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/go81/internal/man"
)

// showMan opens Ajuda → Manual embutido: the man of the shell in a window,
// with a field for what to look up (the same words as "man ...").
func showMan(a fyne.App, query string) {
	win := a.NewWindow("Go81 — manual (man)")
	text := widget.NewLabel("")
	text.TextStyle = fyne.TextStyle{Monospace: true}
	scroll := container.NewScroll(text)
	entry := widget.NewEntry()
	entry.SetPlaceHolder("PRINT, shell load, D_FILE, FP-CALC, -k fita, -l sysvar…")
	look := func(q string) {
		out, err := man.Lookup(strings.Fields(q), 80)
		if err != nil {
			out = err.Error()
		}
		text.SetText(strings.TrimRight(out, "\n"))
		scroll.ScrollToTop()
	}
	entry.OnSubmitted = look
	search := widget.NewButton("Mostrar", func() { look(entry.Text) })
	top := container.NewBorder(nil, nil, widget.NewLabel("man"), search, entry)
	win.SetContent(container.NewBorder(top, nil, nil, nil, scroll))
	win.Resize(fyne.NewSize(720, 560))
	entry.SetText(query)
	look(query)
	win.Show()
	win.Canvas().Focus(entry)
}
