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
// shell (list, find, delete, rem).
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
	m := fyne.NewMenuItem("Programa BASIC", nil)
	m.ChildMenu = fyne.NewMenu("", list, find, del, rem)
	return m
}
