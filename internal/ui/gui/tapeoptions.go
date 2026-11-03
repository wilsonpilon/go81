// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
	"github.com/wilsonpilon/go81/internal/session"
)

// Opções → Fita (EightyOne: Options → Tape): where LOAD reads from (the
// tape manager or the WAV loader) and where SAVE goes (the tape manager,
// the WAV loader or the audio output), as tape origem / tape destino.

// tapeOptions builds the Fita submenu. sync puts the check marks on what
// the deck uses now (a tape put in can change the source); it reports
// whether something changed.
func (e *emulator) tapeOptions(w fyne.Window, refresh func()) (item *fyne.MenuItem, sync func() bool) {
	inTZX := fyne.NewMenuItem("Gerenciador de fitas (TZX)", nil)
	inWav := fyne.NewMenuItem("Carregador WAV", nil)
	outTZX := fyne.NewMenuItem("Gerenciador de fitas (TZX)", nil)
	outWav := fyne.NewMenuItem("Carregador WAV (grava o som)", nil)
	outAudio := fyne.NewMenuItem("Saída de áudio", nil)
	sync = func() bool {
		var in bool
		var o tape.Output
		e.sess.Do(func(m *zx81.Machine) { in, o = m.Tape.In, m.Tape.Output() })
		marks := []struct {
			it *fyne.MenuItem
			on bool
		}{{inTZX, in}, {inWav, !in}, {outTZX, o == tape.OutTape}, {outWav, o == tape.OutWav}, {outAudio, o == tape.OutAudio}}
		changed := false
		for _, mk := range marks {
			if mk.it.Checked != mk.on {
				mk.it.Checked, changed = mk.on, true
			}
		}
		return changed
	}
	run := func(title, line string) func() {
		return func() {
			go func() {
				var out strings.Builder
				err := e.sess.Exec(line, &out)
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, w)
					} else {
						e.flash(strings.TrimSpace(out.String()))
					}
					sync()
					refresh()
				})
			}()
		}
	}
	inTZX.Action = run("Origem da fita", "tape origem tzx")
	inWav.Action = run("Origem da fita", "tape origem wav")
	outTZX.Action = run("Destino da fita", "tape destino tzx")
	outWav.Action = run("Destino da fita", "tape destino wav")
	outAudio.Action = run("Destino da fita", "tape destino som")
	source := fyne.NewMenuItem("Origem", nil)
	source.ChildMenu = fyne.NewMenu("", inTZX, inWav)
	dest := fyne.NewMenuItem("Destino", nil)
	dest.ChildMenu = fyne.NewMenu("", outTZX, outWav, outAudio)
	item = fyne.NewMenuItem("Fita", nil)
	item.ChildMenu = fyne.NewMenu("", source, dest)
	sync()
	return item, sync
}

// speedItem is Opções → Velocidade… (EightyOne: Options → Speed): the
// emulation speed, from 10% to 3200% of a real ZX81.
func (e *emulator) speedItem(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem("Velocidade…", func() {
		var opts []string
		for _, p := range session.Speeds {
			opts = append(opts, speedLabel(p))
		}
		cur := e.sess.Speed()
		sel := widget.NewSelect(opts, nil)
		sel.SetSelected(speedLabel(cur))
		if sel.Selected == "" {
			sel.PlaceHolder = speedLabel(cur)
		}
		other := widget.NewEntry()
		other.SetPlaceHolder(fmt.Sprintf("outra (%d a %d)", session.MinSpeed, session.MaxSpeed))
		note := widget.NewLabel("100% é a velocidade de um ZX81 de verdade. Fora de 100% não há som.\nO turbo (F6) continua sendo 10× por cima da velocidade escolhida.")
		content := container.NewVBox(
			widget.NewForm(widget.NewFormItem("Velocidade", sel), widget.NewFormItem("Outra, em %", other)),
			note)
		dialog.ShowCustomConfirm("Velocidade", "OK", "Cancelar", content, func(ok bool) {
			if !ok {
				return
			}
			p := cur
			if t := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(other.Text), "%")); t != "" {
				n, err := strconv.Atoi(t)
				if err != nil {
					dialog.ShowError(fmt.Errorf("a velocidade é um número em %%"), w)
					return
				}
				p = n
			} else if sel.Selected != "" {
				p, _ = strconv.Atoi(strings.Fields(sel.Selected)[0])
			}
			if err := e.sess.SetSpeed(p); err != nil {
				dialog.ShowError(err, w)
				return
			}
			e.flash(fmt.Sprintf("velocidade %d%%", p))
		}, w)
	})
}

func speedLabel(p int) string {
	s := fmt.Sprintf("%d %%", p)
	if p == 100 {
		s += " (normal)"
	} else if p > 100 {
		s += fmt.Sprintf(" (%g×)", float64(p)/100)
	}
	return s
}
