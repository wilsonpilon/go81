// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
)

// Ferramentas → Gerenciador de fitas… (EightyOne: Tools → Tape Manager):
// the blocks of the tape, the current one marked, the Play/Stop/Rewind
// buttons and the block editing (go to, delete, move, rename). Every
// button is a tape command, so the shell sees the same tape.

// tapeView is what the window shows, read from the deck.
type tapeView struct {
	rows      []string
	programs  []bool // the block is a ZX81 program (can be renamed)
	cur       int
	playing   bool
	flash     bool
	autoStart bool
	status    string
}

func (e *emulator) readTapeView() tapeView {
	var v tapeView
	e.sess.Do(func(m *zx81.Machine) {
		d := m.Tape
		p := d.Player
		v.cur, v.playing, v.flash, v.autoStart = p.CurBlock, p.Playing, p.FlashLoad, p.AutoStart
		for i, b := range p.Blocks {
			pause := ""
			if b.Pause > 0 {
				pause = fmt.Sprintf("   (pausa %.1f s)", float64(b.Pause)/1000)
			}
			v.rows = append(v.rows, fmt.Sprintf("%d   %s%s", i+1, session.BlockText(b), pause))
			v.programs = append(v.programs, strings.HasPrefix(session.BlockText(b), "programa "))
		}
		state := "parada"
		if p.Playing {
			state = "tocando"
		}
		v.status = fmt.Sprintf("Fita %s · %d bloco(s)", state, len(p.Blocks))
		if d.Wav != nil && !d.In {
			at, length := d.Wav.Position()
			v.status = fmt.Sprintf("Origem: som (WAV) %.1f de %.1f s · %s", at, length, v.status)
		}
	})
	return v
}

func (v tapeView) equal(o tapeView) bool {
	if v.cur != o.cur || v.playing != o.playing || v.status != o.status || v.flash != o.flash || v.autoStart != o.autoStart || len(v.rows) != len(o.rows) {
		return false
	}
	for i := range v.rows {
		if v.rows[i] != o.rows[i] {
			return false
		}
	}
	return true
}

// tapeManagerItem opens the tape manager window (one at a time).
func (e *emulator) tapeManagerItem(a fyne.App, w fyne.Window) *fyne.MenuItem {
	var win fyne.Window
	return fyne.NewMenuItem("Gerenciador de fitas…", func() {
		if win != nil {
			win.RequestFocus()
			return
		}
		win = e.tapeManager(a, func() { win = nil })
		win.Show()
	})
}

// tapeManager builds the window.
func (e *emulator) tapeManager(a fyne.App, closed func()) fyne.Window {
	win := a.NewWindow("Gerenciador de fitas — Go81")
	view := e.readTapeView()
	selected := -1
	list := widget.NewList(
		func() int { return len(view.rows) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.MediaPlayIcon()), widget.NewLabel("000   programa NOME DO PROGRAMA, 00000 bytes   (pausa 0.0 s)"))
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			if id == view.cur { // the current block: the next LOAD starts here
				icon.SetResource(theme.MediaPlayIcon())
			} else {
				icon.SetResource(nil)
			}
			box.Objects[1].(*widget.Label).SetText(view.rows[id])
		})
	list.OnSelected = func(id widget.ListItemID) { selected = id }
	list.OnUnselected = func(widget.ListItemID) { selected = -1 }
	status := widget.NewLabel(view.status)
	flash := widget.NewCheck("Carga rápida", nil)
	auto := widget.NewCheck("Partida automática", nil)
	var update func()
	update = func() {
		nv := e.readTapeView()
		if nv.equal(view) {
			return
		}
		view = nv
		status.SetText(view.status)
		flash.SetChecked(view.flash)
		auto.SetChecked(view.autoStart)
		if selected >= len(view.rows) {
			list.UnselectAll()
		}
		list.Refresh()
	}
	run := func(line string) {
		var out strings.Builder
		if err := e.sess.Exec(line, &out); err != nil {
			dialog.ShowError(err, win)
		}
		update()
	}
	onSel := func(f func(n int)) func() {
		return func() {
			if selected < 0 || selected >= len(view.rows) {
				dialog.ShowInformation("Gerenciador de fitas", "Escolha um bloco na lista.", win)
				return
			}
			f(selected + 1)
		}
	}
	flash.SetChecked(view.flash)
	auto.SetChecked(view.autoStart)
	flash.OnChanged = func(on bool) {
		if on != view.flash {
			run(map[bool]string{true: "tape rapida sim", false: "tape rapida nao"}[on])
		}
	}
	auto.OnChanged = func(on bool) {
		if on != view.autoStart {
			run(map[bool]string{true: "tape automatica sim", false: "tape automatica nao"}[on])
		}
	}
	btn := func(label string, icon fyne.Resource, f func()) *widget.Button {
		return widget.NewButtonWithIcon(label, icon, f)
	}
	transport := container.NewHBox(
		btn("", theme.MediaPlayIcon(), func() { run("tape play") }),
		btn("", theme.MediaStopIcon(), func() { run("tape stop") }),
		btn("", theme.MediaSkipPreviousIcon(), func() { run("tape rewind") }),
		widget.NewSeparator(),
		flash, auto)
	edit := container.NewHBox(
		btn("Ir para", theme.NavigateNextIcon(), onSel(func(n int) { run(fmt.Sprintf("tape ir %d", n)) })),
		btn("Subir", theme.MoveUpIcon(), onSel(func(n int) {
			if n > 1 {
				run(fmt.Sprintf("tape mover %d %d", n, n-1))
				list.Select(n - 2)
			}
		})),
		btn("Descer", theme.MoveDownIcon(), onSel(func(n int) {
			if n < len(view.rows) {
				run(fmt.Sprintf("tape mover %d %d", n, n+1))
				list.Select(n)
			}
		})),
		btn("Renomear…", theme.DocumentCreateIcon(), onSel(func(n int) {
			if !view.programs[n-1] {
				dialog.ShowInformation("Renomear", "Só um programa do ZX81 tem nome.", win)
				return
			}
			entry := widget.NewEntry()
			dialog.ShowForm("Renomear o bloco", "OK", "Cancelar", []*widget.FormItem{widget.NewFormItem("Nome novo", entry)}, func(ok bool) {
				if ok && strings.TrimSpace(entry.Text) != "" {
					run(fmt.Sprintf("tape renomear %d %q", n, strings.ToUpper(strings.TrimSpace(entry.Text))))
				}
			}, win)
		})),
		btn("Apagar", theme.DeleteIcon(), onSel(func(n int) {
			dialog.ShowConfirm("Apagar o bloco", fmt.Sprintf("Tirar o bloco %d da fita?", n), func(ok bool) {
				if ok {
					run(fmt.Sprintf("tape apagar %d", n))
					list.UnselectAll()
				}
			}, win)
		})))
	files := container.NewHBox(
		btn("Abrir…", theme.FolderOpenIcon(), func() { e.tapeOpenItem(win).Action() }),
		btn("Salvar…", theme.DocumentSaveIcon(), func() { e.tapeSaveItem(win).Action() }),
		btn("Nova", theme.ContentClearIcon(), func() { run("tape new") }))
	top := container.NewVBox(container.NewHBox(transport, widget.NewSeparator(), files), edit)
	win.SetContent(container.NewBorder(top, status, nil, nil, list))
	win.Resize(fyne.NewSize(max(680, top.MinSize().Width+8), 420))
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				fyne.Do(update)
			}
		}
	}()
	win.SetOnClosed(func() {
		close(stop)
		closed()
	})
	return win
}
