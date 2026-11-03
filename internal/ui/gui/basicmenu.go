// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"bytes"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// askText asks for a few fields in a form.
func askText(w fyne.Window, title string, labels, defaults []string, done func([]string)) {
	entries := make([]*widget.Entry, len(labels))
	var items []*widget.FormItem
	for i, l := range labels {
		entries[i] = widget.NewEntry()
		entries[i].SetText(defaults[i])
		items = append(items, widget.NewFormItem(l, entries[i]))
	}
	dialog.ShowForm(title, "OK", "Cancelar", items, func(ok bool) {
		if !ok {
			return
		}
		vals := make([]string, len(entries))
		for i, en := range entries {
			vals[i] = strings.TrimSpace(en.Text)
		}
		done(vals)
	}, w)
}

// showCommandText runs a command and shows its output in a window with a
// monospace font and a Copiar button (Listagem BASIC, Procurar linhas).
func (e *emulator) showCommandText(a fyne.App, w fyne.Window, title, line string) {
	go func() {
		var out bytes.Buffer
		err := e.sess.Exec(line, &out)
		fyne.Do(func() {
			if err != nil && out.Len() == 0 {
				dialog.ShowError(err, w)
				return
			}
			text := widget.NewLabel(strings.TrimRight(out.String(), "\n"))
			text.TextStyle = fyne.TextStyle{Monospace: true}
			copyBtn := widget.NewButton("Copiar", func() { a.Clipboard().SetContent(out.String()) })
			win := a.NewWindow(title)
			win.SetContent(container.NewBorder(nil, container.NewHBox(copyBtn), nil, nil, container.NewScroll(text)))
			win.Resize(fyne.NewSize(640, 480))
			win.Show()
		})
	}()
}

// basicMenu is Ferramentas → Programa BASIC: the program utilities of the
// shell (list, find, delete, renum, rem). auto is typed in the console (F2).
func (e *emulator) basicMenu(a fyne.App, w fyne.Window) *fyne.MenuItem {
	list := fyne.NewMenuItem("Listagem…", func() { e.showCommandText(a, w, "Listagem BASIC", "list") })
	find := fyne.NewMenuItem("Procurar linhas…", func() {
		askText(w, "Procurar linhas", []string{"Texto"}, []string{""}, func(v []string) {
			if v[0] == "" {
				return
			}
			e.showCommandText(a, w, "Linhas com "+v[0], "find "+v[0])
		})
	})
	del := fyne.NewMenuItem("Apagar linhas…", func() {
		askText(w, "Apagar linhas", []string{"Da linha", "Até a linha"}, []string{"", ""}, func(v []string) {
			if v[0] == "" {
				return
			}
			to := v[1]
			if to == "" {
				to = v[0]
			}
			e.runCommand(w, "Apagar linhas", fmt.Sprintf("delete %s,%s", v[0], to))
		})
	})
	rem := fyne.NewMenuItem("Criar linha REM…", func() {
		askText(w, "Criar linha REM para código de máquina", []string{"Linha", "Tamanho (bytes)", "Byte de enchimento"},
			[]string{"1", "100", "0"}, func(v []string) {
				e.runCommand(w, "Linha REM", fmt.Sprintf("rem %s %s %s", v[0], v[1], v[2]))
			})
	})
	renum := fyne.NewMenuItem("Renumerar…", func() {
		askText(w, "Renumerar (corrige GOTO, GOSUB, RUN, LIST e LLIST)", []string{"Primeira linha", "Passo", "Só da linha (vazio: todas)", "Até a linha"},
			[]string{"10", "10", "", ""}, func(v []string) {
				cmd := fmt.Sprintf("renum %s %s", v[0], v[1])
				if v[2] != "" {
					to := v[3]
					if to == "" {
						to = "9999"
					}
					cmd += fmt.Sprintf(" %s,%s", v[2], to)
				}
				e.runCommand(w, "Renumerar", cmd)
			})
	})
	m := fyne.NewMenuItem("Programa BASIC", nil)
	m.ChildMenu = fyne.NewMenu("", append([]*fyne.MenuItem{list, find, del, renum, rem, fyne.NewMenuItemSeparator()}, convertItems(w)...)...)
	return m
}

// recordMenu is Ferramentas → Gravar o som em WAV: the record command of
// the shell; the same item stops the recording.
func (e *emulator) recordMenu(w fyne.Window) *fyne.MenuItem {
	const startLabel, stopLabel = "Gravar o som em WAV…", "Parar a gravação do som"
	var item *fyne.MenuItem
	setLabel := func(l string) {
		item.Label = l
		if m := w.MainMenu(); m != nil {
			m.Refresh()
		}
	}
	item = fyne.NewMenuItem(startLabel, func() {
		if rec, _ := e.sess.Recording(); rec != "" {
			msg, err := e.sess.StopRecording()
			setLabel(startLabel)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			e.flash("Gravado " + msg)
			return
		}
		d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			path := wc.URI().Path()
			wc.Close()
			if err := e.sess.StartRecording(path); err != nil {
				dialog.ShowError(err, w)
				return
			}
			setLabel(stopLabel)
			if !e.sess.Settings().Sound {
				dialog.ShowInformation("Gravar o som", "O som pelo vídeo está desligado (Opções → Som pelo vídeo):\no arquivo terá silêncio até ele ser ligado.", w)
			}
			e.flash("Gravando o som em " + path)
		}, w)
		d.SetFileName("go81-som.wav")
		d.Show()
	})
	return item
}
