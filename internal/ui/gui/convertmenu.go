// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/session"
)

// Ferramentas → Programa BASIC → Converter: BASIC text (the zxtext2p
// format) to .p and back, without touching the ZX81 (txt2p and p2txt).

// convertFile asks for the file to convert, then where to write it.
func convertFile(w fyne.Window, title string, inExts []string, outExt string, convert func(in, out string) (string, error)) {
	open := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			return
		}
		r.Close()
		if r.URI().Scheme() != "file" {
			dialog.ShowError(fmt.Errorf("só dá para converter um arquivo do disco"), w)
			return
		}
		in := r.URI().Path()
		save := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			wc.Close()
			msg, err := convert(in, wc.URI().Path())
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			dialog.ShowInformation(title, msg, w)
		}, w)
		save.SetFileName(strings.TrimSuffix(filepath.Base(in), filepath.Ext(in)) + outExt)
		if dir, err := storage.ListerForURI(storage.NewFileURI(filepath.Dir(in))); err == nil {
			save.SetLocation(dir)
		}
		save.Show()
	}, w)
	var exts []string
	for _, x := range inExts {
		exts = append(exts, x, strings.ToUpper(x))
	}
	open.SetFilter(storage.NewExtensionFileFilter(exts))
	open.Show()
}

func convertItems(w fyne.Window) []*fyne.MenuItem {
	toP := fyne.NewMenuItem("Converter texto em .p (txt2p)…", func() {
		convertFile(w, "Texto → .p", []string{".txt", ".bas", ".b81", ".azx"}, ".p", func(in, out string) (string, error) {
			return session.TextToPFile(in, out, basic.POptions{})
		})
	})
	toText := fyne.NewMenuItem("Converter .p em texto (p2txt)…", func() {
		convertFile(w, ".p → texto", []string{".p", ".81", ".p81", ".o", ".80"}, ".txt", session.PToTextFile)
	})
	return []*fyne.MenuItem{toP, toText}
}
