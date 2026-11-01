// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"strings"
)

// Model is a machine of the ZX81 family as EightyOne's hardware dialog
// sets it up (HW_.cpp: the machine buttons, RefreshDefaultRamSettings,
// LoadRomBox).
type Model struct {
	ID    string // zx81, ts1000...
	Short string // in the status line: ZX81, TS1500...
	Name  string // as shown
	// ROMs are the ROMs EightyOne offers for the model; the first one is
	// the default.
	ROMs []string
	// BaseRAM is the RAM inside the machine in K. A RAM pack replaces it,
	// or is added to it when PackAdds is set (TS1500, TK85, R-470).
	BaseRAM  int
	PackAdds bool
	// DefaultPack is the RAM pack chosen with the model (0 = none).
	DefaultPack int
	NTSC        bool
	M1NotFixed  bool // the M1NOT modification is always there (TS1500)
	NoROMShadow bool // no copy of the ROM at 32-48K (TS1500)
	NoLowRAM    bool // no RAM at 8-16K (R-470)
	ZX80        bool // the ZX80 video circuit
	// Ready is false for the models Go81 does not emulate yet.
	Ready bool
}

// Models are the machines of the hardware dialog, in EightyOne's order.
var Models = []Model{
	{ID: "zx81", Short: "ZX81", Name: "Sinclair ZX81", ROMs: []string{"zx81.edition3", "zx81.edition2", "zx81.edition1", "tree-forth", "zx81-forth", "zx.aszmic.e04", "zx.aszmic.e07"},
		BaseRAM: 1, DefaultPack: 16, Ready: true},
	{ID: "ts1000", Short: "TS1000", Name: "Timex Sinclair 1000", ROMs: []string{"zx81.edition3", "tree-forth"},
		BaseRAM: 2, DefaultPack: 0, NTSC: true, Ready: true},
	{ID: "ts1500", Short: "TS1500", Name: "Timex Sinclair 1500", ROMs: []string{"ts1500", "tree-forth"},
		BaseRAM: 16, PackAdds: true, NTSC: true, M1NotFixed: true, NoROMShadow: true, Ready: true},
	{ID: "tk85", Short: "TK85", Name: "Microdigital TK85", ROMs: []string{"tk85"},
		BaseRAM: 16, PackAdds: true, NTSC: true, Ready: true},
	{ID: "r470", Short: "R-470", Name: "Ringo R-470", ROMs: []string{"ringo470"},
		BaseRAM: 16, PackAdds: true, NTSC: true, NoLowRAM: true, Ready: true},
	{ID: "zx80", Short: "ZX80", Name: "Sinclair ZX80", ROMs: []string{"zx80", "zx81.edition1"}, BaseRAM: 1, DefaultPack: 16, ZX80: true, Ready: true},
	{ID: "lambda", Short: "Lambda", Name: "Lambda 8300", ROMs: []string{"lambda8300", "lambda8300colour", "pc8300timex"}, BaseRAM: 2, DefaultPack: 16, NTSC: true},
	{ID: "zx97", Short: "ZX97", Name: "ZX97 Lite", ROMs: []string{"zx97"}, BaseRAM: 1, DefaultPack: 0},
	{ID: "ace", Short: "Ace", Name: "Jupiter Ace", ROMs: []string{"jupiterace"}, BaseRAM: 3, PackAdds: true, DefaultPack: 16},
}

// RAMPacks are the RAM pack sizes (K) of the hardware dialog; 0 = none.
var RAMPacks = []int{0, 1, 2, 4, 8, 16, 32, 48}

// FindModel returns the model with that id (case does not matter; "" is
// the ZX81).
func FindModel(id string) (Model, error) {
	if id == "" {
		id = "zx81"
	}
	for _, m := range Models {
		if strings.EqualFold(m.ID, id) {
			return m, nil
		}
	}
	var ids []string
	for _, m := range Models {
		if m.Ready {
			ids = append(ids, m.ID)
		}
	}
	return Model{}, fmt.Errorf("modelo desconhecido: %s (modelos: %s)", id, strings.Join(ids, ", "))
}

// model returns the model of the settings (the ZX81 if unknown).
func (set Settings) model() Model {
	m, err := FindModel(set.Model)
	if err != nil {
		m, _ = FindModel("")
	}
	return m
}

// WithModel switches the settings to a model with its defaults (ROM, RAM
// pack, NTSC), as clicking its button in EightyOne's hardware dialog.
func (set Settings) WithModel(id string) (Settings, error) {
	m, err := FindModel(id)
	if err != nil {
		return set, err
	}
	if !m.Ready {
		return set, fmt.Errorf("o %s ainda não é emulado pelo Go81", m.Name)
	}
	set.Model = m.ID
	set.ROM = m.ROMs[0]
	set.RAMK = m.DefaultPack
	set.NTSC = m.NTSC
	if m.NoLowRAM {
		set.LowRAM = false
	}
	return set, nil
}

// TotalRAM is the RAM of the machine in K: the internal RAM, replaced by
// the RAM pack or added to it.
func (set Settings) TotalRAM() int {
	m := set.model()
	switch {
	case set.RAMK <= 0:
		return m.BaseRAM
	case m.PackAdds:
		return m.BaseRAM + set.RAMK
	}
	return set.RAMK
}

// ramTop is the highest RAM address (ConfigureRamTop in HW_.cpp).
func (set Settings) ramTop() int {
	t := set.TotalRAM()<<10 + 16383
	if t > 65535 {
		t = 65535
	}
	return t
}

// CheckPack validates a RAM pack size.
func CheckPack(k int) (int, error) {
	for _, p := range RAMPacks {
		if k == p {
			if k == 1 { // 1K pack = no pack on the ZX81
				return 0, nil
			}
			return k, nil
		}
	}
	return 0, fmt.Errorf("RAM inválida: %d (0, 1, 2, 4, 8, 16, 32 ou 48)", k)
}

// Describe is the one-line summary of the machine of the settings.
func (set Settings) Describe() string {
	tv := "PAL"
	if set.NTSC {
		tv = "NTSC"
	}
	d := fmt.Sprintf("%s %dK %s · ROM %s · WRX %s · som %s", set.model().Short, set.TotalRAM(), tv, set.ROM, onOff(set.WRX), onOff(set.Sound))
	var extra []string
	if set.LowRAM && !set.model().NoLowRAM {
		extra = append(extra, "RAM 8-16K")
	}
	if set.M1Not || set.model().M1NotFixed {
		extra = append(extra, "M1NOT")
	}
	if set.FPFix {
		extra = append(extra, "correção PF")
	}
	if len(extra) > 0 {
		d += " · " + strings.Join(extra, ", ")
	}
	return d
}
