// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"fmt"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/go81/internal/session"
)

// The hardware dialog follows EightyOne's (HW_.cpp): the machine buttons,
// the RAM pack, the ROM and the advanced settings, and the interfaces. The
// interfaces Go81 does not emulate yet are there too: choosing one says
// "A implementar".

// todoSelect is a list of EightyOne options where only the first one (none)
// works yet.
func todoSelect(w fyne.Window, label string, options ...string) *widget.Select {
	var sel *widget.Select
	sel = widget.NewSelect(options, func(v string) {
		if v != options[0] {
			showTodo(w, label+": "+v)
			sel.SetSelectedIndex(0)
		}
	})
	sel.SetSelectedIndex(0)
	return sel
}

// todoCheck is an EightyOne check box Go81 does not have yet.
func todoCheck(w fyne.Window, label string) *widget.Check {
	var c *widget.Check
	c = widget.NewCheck(label, func(on bool) {
		if on {
			showTodo(w, label)
			c.SetChecked(false)
		}
	})
	return c
}

func packLabel(k int) string {
	if k == 0 {
		return "Nenhum"
	}
	return fmt.Sprintf("%dK", k)
}

// hardwareDialog is Opções → Hardware…
func (e *emulator) hardwareDialog(w fyne.Window, setZoom func(float64), zoom func() float64) {
	set := e.sess.Settings()

	// Models: the ones Go81 does not emulate yet are marked.
	var names []string
	for _, m := range session.Models {
		n := m.Name
		if !m.Ready {
			n += " (a implementar)"
		}
		names = append(names, n)
	}
	modelAt := func(name string) (session.Model, bool) {
		for i, n := range names {
			if n == name {
				return session.Models[i], true
			}
		}
		return session.Model{}, false
	}

	romSel := widget.NewSelect(nil, nil)
	var packs []string
	for _, k := range session.RAMPacks {
		if k != 1 {
			packs = append(packs, packLabel(k))
		}
	}
	ramSel := widget.NewSelect(packs, nil)
	total := widget.NewLabel("")
	ntsc := widget.NewCheck("60 Hz (NTSC)", nil)
	lowRAM := widget.NewCheck("RAM em 8K-16K", nil)
	m1not := widget.NewCheck("Circuito M1NOT", nil)
	fpfix := widget.NewCheck("Correção do ponto flutuante (hardware)", nil)
	snd := widget.NewCheck("Som pelo sinal de vídeo (Video Sound)", nil)
	hires := widget.NewSelect([]string{"Nenhuma", "WRX", "G007", "Memotech", "Quicksilva"}, nil)
	hires.OnChanged = func(v string) {
		if v != "Nenhuma" && v != "WRX" {
			showTodo(w, "Alta resolução: "+v)
			hires.SetSelected("Nenhuma")
		}
	}

	// settings reads the form.
	cur := set
	settings := func() session.Settings {
		s := cur
		s.ROM = romSel.Selected
		s.RAMK = 0
		for _, k := range session.RAMPacks {
			if packLabel(k) == ramSel.Selected {
				s.RAMK = k
			}
		}
		s.NTSC, s.LowRAM, s.M1Not, s.FPFix, s.Sound = ntsc.Checked, lowRAM.Checked, m1not.Checked, fpfix.Checked, snd.Checked
		s.WRX = hires.Selected == "WRX"
		return s
	}
	updateTotal := func() { total.SetText(fmt.Sprintf("RAM total: %dK", settings().TotalRAM())) }
	ramSel.OnChanged = func(string) { updateTotal() }

	// fill puts the settings in the form.
	fill := func(s session.Settings) {
		cur = s
		m, _ := session.FindModel(s.Model)
		roms := append([]string(nil), m.ROMs...)
		if s.ROM != "" && !slices.Contains(roms, s.ROM) {
			roms = append(roms, s.ROM)
		}
		romSel.SetOptions(roms)
		romSel.SetSelected(s.ROM)
		ramSel.SetSelected(packLabel(s.RAMK))
		ntsc.SetChecked(s.NTSC)
		lowRAM.SetChecked(s.LowRAM && !m.NoLowRAM)
		if m.NoLowRAM {
			lowRAM.Disable()
		} else {
			lowRAM.Enable()
		}
		m1not.SetChecked(s.M1Not || m.M1NotFixed)
		if m.M1NotFixed {
			m1not.Disable()
		} else {
			m1not.Enable()
		}
		fpfix.SetChecked(s.FPFix)
		snd.SetChecked(s.Sound)
		if s.WRX {
			hires.SetSelected("WRX")
		} else {
			hires.SetSelected("Nenhuma")
		}
		updateTotal()
	}

	model := widget.NewRadioGroup(names, nil)
	m0, _ := session.FindModel(set.Model)
	model.SetSelected(names[slices.IndexFunc(session.Models, func(m session.Model) bool { return m.ID == m0.ID })])
	model.OnChanged = func(name string) {
		m, ok := modelAt(name)
		if !ok {
			return
		}
		if !m.Ready {
			showTodo(w, m.Name)
			prev, _ := session.FindModel(cur.Model)
			model.SetSelected(names[slices.IndexFunc(session.Models, func(x session.Model) bool { return x.ID == prev.ID })])
			return
		}
		if s, err := cur.WithModel(m.ID); err == nil {
			s.Sound, s.WRX = snd.Checked, hires.Selected == "WRX"
			fill(s)
		}
	}
	fill(set)

	browse := widget.NewButton("…", func() {
		d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			p := r.URI().Path()
			r.Close()
			opts := romSel.Options
			if !slices.Contains(opts, p) {
				romSel.SetOptions(append(opts, p))
			}
			romSel.SetSelected(p)
		}, w)
		d.Show()
	})

	machine := widget.NewForm(
		widget.NewFormItem("ROM", container.NewBorder(nil, nil, nil, browse, romSel)),
		widget.NewFormItem("RAM pack", container.NewHBox(ramSel, total)),
	)
	advanced := container.NewVBox(ntsc, lowRAM, m1not, fpfix,
		todoCheck(w, "Circuito de WAIT melhorado"), todoCheck(w, "Proteger a ROM contra escrita (desligar)"))
	interfaces := widget.NewForm(
		widget.NewFormItem("Alta resolução", hires),
		widget.NewFormItem("Som", container.NewVBox(snd, todoSelect(w, "Placa de som", "Nenhuma", "ZON X", "Quicksilva"))),
		widget.NewFormItem("Gerador de caracteres", todoSelect(w, "Gerador de caracteres", "Sinclair", "dk'tronics", "Quicksilva", "CHR$128", "Kayde")),
		widget.NewFormItem("Cor", todoSelect(w, "Cor", "Nenhuma", "Lambda", "Chroma")),
		widget.NewFormItem("Fala", todoSelect(w, "Fala", "Nenhuma", "The Parrot", "Sweet Talker", "Mageco", "S-Pack", "Talk-Back")),
		widget.NewFormItem("Joystick", todoSelect(w, "Joystick", "Nenhum", "Cursor", "Kempston", "Programável")),
		widget.NewFormItem("Cartucho de ROM", todoSelect(w, "Cartucho de ROM", "Nenhum", "Sinclair", "ZXC1", "ZXC2", "ZXC3", "ZXC4")),
		widget.NewFormItem("", container.NewHBox(todoCheck(w, "ZX Printer"), todoCheck(w, "ZXpand"), todoCheck(w, "Z80 Assembler"), todoCheck(w, "Memocalc"))),
	)
	drives := widget.NewForm(
		widget.NewFormItem("Disquete", todoSelect(w, "Unidade de disquete", "Nenhuma", "ZX1541", "Larken")),
		widget.NewFormItem("IDE", todoSelect(w, "Interface IDE", "Nenhuma", "MWCFIde")),
	)
	tabs := container.NewAppTabs(
		container.NewTabItem("Máquina", container.NewVBox(machine, widget.NewSeparator(), advanced)),
		container.NewTabItem("Interfaces", interfaces),
		container.NewTabItem("Unidades", drives),
	)
	content := container.NewBorder(nil, nil, container.NewVBox(widget.NewLabelWithStyle("Modelo", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), model), nil, tabs)

	d := dialog.NewCustomConfirm("Hardware", "Aplicar e reiniciar", "Cancelar", content, func(ok bool) {
		if !ok {
			return
		}
		if err := e.sess.Configure(settings()); err != nil {
			dialog.ShowError(err, w)
			return
		}
		setZoom(zoom())
	}, w)
	d.Show()
}
