// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

// Arquivo → Carregar snapshot / Salvar snapshot: the whole machine in a
// .z81 file (EightyOne's format).

func (e *emulator) snapshotLoadItem(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem("Carregar snapshot…", func() {
		d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			r.Close()
			if r.URI().Scheme() != "file" {
				dialog.ShowError(fmt.Errorf("só dá para ler um arquivo do disco"), w)
				return
			}
			e.openFile(w, r.URI().Path())
		}, w)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".z81", ".Z81"}))
		d.Show()
	})
}

func (e *emulator) snapshotSaveItem(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem("Salvar snapshot…", func() {
		d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			wc.Close()
			if wc.URI().Scheme() != "file" {
				dialog.ShowError(fmt.Errorf("só dá para gravar num arquivo do disco"), w)
				return
			}
			path := wc.URI().Path()
			if err := e.sess.SaveSnapshot(path); err != nil {
				dialog.ShowError(err, w)
				return
			}
			e.flash("Snapshot gravado: " + filepath.Base(path))
		}, w)
		d.SetFileName("go81.z81")
		d.SetFilter(storage.NewExtensionFileFilter([]string{".z81"}))
		d.Show()
	})
}
