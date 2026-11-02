// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/z81"
)

// Snapshots (.z81, EightyOne's format): the whole state of the machine —
// model and options, CPU registers, video state and memory — saved between
// two frames and loaded back.

func registerSnapshot(r *Registry) {
	r.Register(&Command{Name: "snapshot", Aliases: []string{"snap"}, Usage: "snapshot save|load arquivo.z81",
		Help: "grava ou carrega o estado inteiro da máquina (.z81 do EightyOne)", Run: cmdSnapshot})
}

func cmdSnapshot(s *Session, a []string, out io.Writer) error {
	if len(a) != 2 {
		return usageErr("snapshot save|load arquivo.z81")
	}
	switch strings.ToLower(a[0]) {
	case "save", "salvar", "gravar":
		if err := os.WriteFile(a[1], []byte(s.snapshot().Text()), 0o644); err != nil {
			return err
		}
		fmt.Fprintln(out, "snapshot gravado em", a[1])
	case "load", "carregar":
		if err := s.loadSnapshot(a[1]); err != nil {
			return err
		}
		fmt.Fprintf(out, "snapshot carregado: %s\n%s\n", filepath.Base(a[1]), s.description())
	default:
		return usageErr("snapshot save|load arquivo.z81")
	}
	return nil
}

// IsSnapshot reports whether a file is a snapshot (by its extension).
func IsSnapshot(name string) bool { return strings.EqualFold(filepath.Ext(name), ".z81") }

// snapModels are the model names EightyOne writes, and ours.
var snapModels = map[string]string{"ZX81": "zx81", "TS1000": "ts1000", "TS1500": "ts1500", "TK85": "tk85",
	"R470": "r470", "ZX80": "zx80", "LAMBDA": "lambda", "ZX97LE": "zx97", "ACE": "ace"}

// snapshot takes the state of the machine.
func (s *Session) snapshot() *z81.Snapshot {
	m, c := s.m, s.m.CPU
	sn := &z81.Snapshot{Model: "ZX81", RAMPack: z81.PackName(s.set.RAMK), LowRAM: s.set.LowRAM,
		M1Not: s.set.M1Not, FPFix: s.set.FPFix, NTSC: s.set.NTSC}
	for name, id := range snapModels {
		if id == s.set.model().ID {
			sn.Model = name
		}
	}
	sn.ROM = filepath.Base(s.set.ROM)
	if !strings.Contains(sn.ROM, ".") {
		sn.ROM += ".rom"
	}
	sn.PC, sn.SP = c.PC, c.SP
	sn.AF = uint16(c.A)<<8 | uint16(c.F)
	sn.BC = uint16(c.B)<<8 | uint16(c.C)
	sn.DE = uint16(c.D)<<8 | uint16(c.E)
	sn.HL = uint16(c.H)<<8 | uint16(c.L)
	sn.AF_, sn.BC_, sn.DE_, sn.HL_ = c.AF_, c.BC_, c.DE_, c.HL_
	sn.IX, sn.IY = c.IX, c.IY
	sn.IR = uint16(c.I)<<8 | uint16(c.R7&0x80) | (c.R & 0x7F)
	sn.IM, sn.IFF1, sn.IFF2, sn.Halted = c.IM, c.IFF1, c.IFF2, c.Halted
	v := m.VideoState()
	sn.NMI, sn.Sync, sn.Line = v.NMI, v.Sync, v.Line
	sn.Start = m.ROMTop() + 1
	sn.Memory = append([]byte(nil), m.Memory[sn.Start:m.Config().RAMTop+1]...)
	return sn
}

// loadSnapshot sets the machine as the snapshot says and restores its
// state.
func (s *Session) loadSnapshot(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	sn, err := z81.Parse(string(data))
	if err != nil {
		return err
	}
	set := s.set
	if sn.Model != "" {
		id, ok := snapModels[sn.Model]
		if !ok {
			return fmt.Errorf("modelo desconhecido no snapshot: %s", sn.Model)
		}
		if set, err = set.WithModel(id); err != nil {
			return err
		}
	}
	end := sn.Start + len(sn.Memory) - 1
	switch k := z81.PackK(sn.RAMPack); {
	case sn.RAMPack == "":
		// No RAM pack named (older snapshots): the smallest one that holds the memory.
		for _, p := range RAMPacks {
			set.RAMK = p
			if set.ramTop() >= end {
				break
			}
		}
	case k < 0:
		return fmt.Errorf("RAM_PACK desconhecido no snapshot: %s", sn.RAMPack)
	default:
		if set.RAMK, err = CheckPack(k); err != nil {
			return err
		}
	}
	set.LowRAM = sn.LowRAM
	if !sn.HasLow && sn.Start < 0x4000 { // older files: memory from 2000h means RAM there
		set.LowRAM = true
	}
	set.FPFix = sn.FPFix
	if sn.HasM1Not {
		set.M1Not = sn.M1Not
	}
	if sn.HasNTSC {
		set.NTSC = sn.NTSC
	}
	if sn.ROM != "" {
		name := strings.ToLower(strings.TrimSuffix(sn.ROM, filepath.Ext(sn.ROM)))
		if _, err := assets.ROM(name); err == nil {
			set.ROM = name
		}
	}
	if err := s.configure(set); err != nil {
		return err
	}
	m, c := s.m, s.m.CPU
	if sn.Start <= m.ROMTop() {
		return errors.New("o snapshot escreveria por cima da ROM")
	}
	copy(m.Memory[sn.Start:], sn.Memory)
	c.PC, c.SP = sn.PC, sn.SP
	c.A, c.F = byte(sn.AF>>8), byte(sn.AF)
	c.B, c.C = byte(sn.BC>>8), byte(sn.BC)
	c.D, c.E = byte(sn.DE>>8), byte(sn.DE)
	c.H, c.L = byte(sn.HL>>8), byte(sn.HL)
	c.AF_, c.BC_, c.DE_, c.HL_ = sn.AF_, sn.BC_, sn.DE_, sn.HL_
	c.IX, c.IY = sn.IX, sn.IY
	c.I, c.R, c.R7 = byte(sn.IR>>8), sn.IR&0xFF, byte(sn.IR&0x80)
	c.IM, c.IFF1, c.IFF2, c.Halted = sn.IM, sn.IFF1, sn.IFF2, sn.Halted
	m.SetVideoState(zx81.VideoState{NMI: sn.NMI, Sync: sn.Sync, Line: sn.Line})
	s.lastProgram = ""
	return nil
}

// LoadSnapshot is loadSnapshot for the windows.
func (s *Session) LoadSnapshot(file string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadSnapshot(file)
}

// SaveSnapshot writes the state of the machine to a .z81 file.
func (s *Session) SaveSnapshot(file string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.WriteFile(file, []byte(s.snapshot().Text()), 0o644)
}
