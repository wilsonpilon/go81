// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/version"
)

// Arquivo → Abrir fita, Nova fita and Estado da fita (the tape command),
// Exibir → Mostrar tela de abertura and Opções → Configuração → Apagar
// configurações.

// tapeOpenItem puts a program on the tape without loading it: the ZX81
// loads it with LOAD "" (EightyOne: File → Open Tape).
func (e *emulator) tapeOpenItem(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem("Abrir fita…", func() {
		d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			r.Close()
			if r.URI().Scheme() != "file" {
				dialog.ShowError(fmt.Errorf("só dá para abrir um arquivo do disco"), w)
				return
			}
			e.runCommand(w, "Fita", `tape insert "`+r.URI().Path()+`"`)
		}, w)
		var exts []string
		for _, x := range tapeExts {
			exts = append(exts, x, strings.ToUpper(x))
		}
		d.SetFilter(storage.NewExtensionFileFilter(exts))
		d.Show()
	})
}

// tapeExts are the files that go on the tape (programs, not snapshots).
var tapeExts = []string{".tzx", ".t81", ".wav", ".p", ".81", ".p81", ".o", ".80", ".bas", ".txt", ".b81", ".azx"}

// tapeSaveItem writes the tape: every block as .tzx or .t81, or the first
// program as .p or .p81 (EightyOne: File → Save Tape).
func (e *emulator) tapeSaveItem(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem("Salvar fita…", func() {
		d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			wc.Close()
			if wc.URI().Scheme() != "file" {
				dialog.ShowError(fmt.Errorf("só dá para gravar num arquivo do disco"), w)
				return
			}
			e.runCommand(w, "Fita", `tape save "`+wc.URI().Path()+`"`)
		}, w)
		d.SetFileName("fita.tzx")
		d.SetFilter(storage.NewExtensionFileFilter([]string{".tzx", ".t81", ".wav", ".p", ".p81", ".TZX", ".T81", ".WAV", ".P", ".P81"}))
		d.Show()
	})
}

// wavLoaderItem plays a recorded tape (.wav) as sound into the ZX81, in
// real time (EightyOne: Tools → Wave Loader): the ROM reads the bits.
func (e *emulator) wavLoaderItem(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem("Carregador WAV (tempo real)…", func() {
		d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			r.Close()
			if r.URI().Scheme() != "file" {
				dialog.ShowError(fmt.Errorf("só dá para abrir um arquivo do disco"), w)
				return
			}
			e.runCommand(w, "Carregador WAV", `tape insert "`+r.URI().Path()+`" som`)
		}, w)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".wav", ".WAV"}))
		d.Show()
	})
}

func (e *emulator) tapeNewItem(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem("Nova fita", func() { e.runCommand(w, "Fita", "tape new") })
}

func (e *emulator) tapeStatusItem(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem("Estado da fita", func() { e.runCommand(w, "Fita", "tape") })
}

// splashItem shows the opening screen: the keyboard of the model and the
// version (EightyOne: View → Show splash screen).
func (e *emulator) splashItem(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem("Mostrar tela de abertura", func() {
		set := e.sess.Settings()
		pic, err := keyboardPicture(set.Model, set.NTSC)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		img := canvas.NewImageFromImage(pic)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(480, 190))
		title := widget.NewLabelWithStyle("Go81 "+version.Full(), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		text := widget.NewLabelWithStyle("Emulador e ambiente de desenvolvimento para o ZX81 e clones,\n"+
			"compatível com o EightyOne. Licença GPL v3 ou posterior.\n\nMáquina: "+e.sess.Description(),
			fyne.TextAlignCenter, fyne.TextStyle{})
		dialog.ShowCustom("Go81", "Fechar", container.NewVBox(title, img, text), w)
	})
}

// deleteConfigItem removes the configuration file and goes back to the
// standard machine (EightyOne: Options → Configuration → Delete settings).
func (e *emulator) deleteConfigItem(w fyne.Window, after func()) *fyne.MenuItem {
	return fyne.NewMenuItem("Apagar configurações…", func() {
		path := e.configPath
		dialog.ShowConfirm("Apagar configurações",
			"Apagar o arquivo de configuração?\n\n"+path+"\n\nA máquina padrão, a janela, os perfis e a máquina de cada jogo\n"+
				"gravados nele se perdem; o Go81 volta ao ZX81 padrão.", func(ok bool) {
				if !ok {
					return
				}
				if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
					dialog.ShowError(err, w)
					return
				}
				if err := e.sess.Configure(session.DefaultSettings()); err != nil {
					dialog.ShowError(err, w)
					return
				}
				after()
				dialog.ShowInformation("Apagar configurações", "Configurações apagadas.", w)
			}, w)
	})
}
