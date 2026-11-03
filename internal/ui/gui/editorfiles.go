// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

// Exibir → Editor de BASIC: open a program in the editor, join one to it
// (MERGE) and save the program of the editor, without going through the
// ZX81. LOAD, SAVE and MERGE typed in the editor without a name come here.

// editorExts are the files the editor opens.
var editorExts = []string{".p", ".81", ".p81", ".tzx", ".t81", ".wav", ".bas", ".txt", ".b81", ".azx"}

func (v *editorView) fileItems() []*fyne.MenuItem {
	return []*fyne.MenuItem{
		fyne.NewMenuItem("Abrir no editor (LOAD)…", func() { v.askOpen(false) }),
		fyne.NewMenuItem("Juntar ao editor (MERGE)…", func() { v.askOpen(true) }),
		fyne.NewMenuItem("Gravar o programa do editor (SAVE)…", v.askSave),
		fyne.NewMenuItem("Memória usada (MEM)", func() {
			if !v.on {
				v.toggle()
			}
			if v.on {
				v.ed.ShowMem()
				v.draw()
			}
		}),
	}
}

func (v *editorView) askOpen(merge bool) {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			return
		}
		r.Close()
		if r.URI().Scheme() != "file" {
			dialog.ShowError(fmt.Errorf("só dá para abrir um arquivo do disco"), v.w)
			return
		}
		v.openFile(r.URI().Path(), merge)
	}, v.w)
	var exts []string
	for _, x := range editorExts {
		exts = append(exts, x, strings.ToUpper(x))
	}
	d.SetFilter(storage.NewExtensionFileFilter(exts))
	d.Show()
}

func (v *editorView) askSave() {
	if v.ed == nil || !v.on {
		v.e.flash("Abra o editor (F7) para gravar o programa dele")
		return
	}
	d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		wc.Close()
		if wc.URI().Scheme() != "file" {
			dialog.ShowError(fmt.Errorf("só dá para gravar num arquivo do disco"), v.w)
			return
		}
		v.saveFile(wc.URI().Path())
		v.draw()
	}, v.w)
	d.SetFileName("programa.p")
	d.Show()
}
