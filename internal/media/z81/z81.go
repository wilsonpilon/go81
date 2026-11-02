// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package z81 reads and writes EightyOne's ZX81 snapshots (.z81): a text
// file with sections ([MACHINE], [CPU], [ZX81], [MEMORY], [ADVANCED]...) of
// names and hexadecimal values, as zx81/snap.cpp writes them. Memory is a
// list of bytes in hex where "*NNNN XX" is NNNN (hex) copies of XX.
// Unknown sections and names are skipped, as EightyOne does; the older
// names HSYNC and ROW (VB81 snapshots) are read too.
package z81

import (
	"fmt"
	"strconv"
	"strings"
)

// Snapshot is the state in a .z81 file.
type Snapshot struct {
	Model string // ZX81, TS1000, TS1500, TK85, R470, ZX80, LAMBDA... ("" = not given)

	PC, SP, HL, DE, BC, AF         uint16
	HL_, DE_, BC_, AF_, IX, IY, IR uint16
	IM, IFF1, IFF2                 byte
	Halted                         bool

	NMI, Sync bool // NMI generator on, sync output white
	Line      int  // line counter (0-7)

	RAMPack  string // as the hardware dialog shows it: None, 2K, 4K, 8K, 16K, 32K, 48K ("" = not given)
	LowRAM   bool   // RAM at 8-16K
	Start    int    // address of Memory[0]
	Memory   []byte
	ROM      string // ROM file name ("" = not given)
	M1Not    bool
	FPFix    bool
	NTSC     bool
	HasNTSC  bool // the file says NTSC or not
	HasM1Not bool
	HasLow   bool // the file says 8K_RAM_ENABLED
}

// Parse reads a .z81 file.
func Parse(text string) (*Snapshot, error) {
	s := &Snapshot{Start: 0x4000}
	tok := strings.Fields(text)
	section := ""
	i := 0
	next := func() string {
		if i >= len(tok) {
			return ""
		}
		i++
		return tok[i-1]
	}
	hexv := func(name string) (int, error) {
		t := next()
		v, err := strconv.ParseUint(t, 16, 32)
		if err != nil {
			return 0, fmt.Errorf("z81: [%s] %s: valor inválido %q", section, name, t)
		}
		return int(v), nil
	}
	cpu := map[string]*uint16{"PC": &s.PC, "SP": &s.SP, "HL": &s.HL, "DE": &s.DE, "BC": &s.BC, "AF": &s.AF,
		"HL_": &s.HL_, "DE_": &s.DE_, "BC_": &s.BC_, "AF_": &s.AF_, "IX": &s.IX, "IY": &s.IY, "IR": &s.IR}
	sawCPU, sawMem := false, false
	for i < len(tok) {
		t := next()
		if strings.HasPrefix(t, "[") {
			section = strings.ToUpper(strings.Trim(t, "[]"))
			continue
		}
		switch section {
		case "MACHINE":
			if t == "MODEL" {
				s.Model = strings.ToUpper(next())
			}
		case "CPU":
			sawCPU = true
			if p, ok := cpu[t]; ok {
				v, err := hexv(t)
				if err != nil {
					return nil, err
				}
				*p = uint16(v)
				continue
			}
			switch t {
			case "IM", "IF1", "IF2", "HT":
				v, err := hexv(t)
				if err != nil {
					return nil, err
				}
				switch t {
				case "IM":
					s.IM = byte(v)
				case "IF1":
					s.IFF1 = byte(v)
				case "IF2":
					s.IFF2 = byte(v)
				default:
					s.Halted = v != 0
				}
			}
		case "ZX81":
			switch t {
			case "NMI", "SYNC", "HSYNC", "LINE", "ROW":
				v, err := hexv(t)
				if err != nil {
					return nil, err
				}
				switch t {
				case "NMI":
					s.NMI = v != 0
				case "SYNC", "HSYNC":
					s.Sync = v != 0
				default:
					s.Line = v & 7
				}
			}
		case "MEMORY":
			switch {
			case t == "RAM_PACK":
				s.RAMPack = next()
			case t == "MEMRANGE":
				v, err := hexv(t)
				if err != nil {
					return nil, err
				}
				next() // the end: the bytes say it
				if sawMem && len(s.Memory) > 0 {
					return nil, fmt.Errorf("z81: mais de um MEMRANGE")
				}
				s.Start = v
			case t == "8K_RAM_ENABLED":
				v, err := hexv(t)
				if err != nil {
					return nil, err
				}
				s.LowRAM, s.HasLow = v != 0, true
			case t == "ROM_PROTECTED":
				next()
			case strings.HasPrefix(t, "*"):
				n, err := strconv.ParseUint(t[1:], 16, 32)
				if err != nil {
					return nil, fmt.Errorf("z81: repetição inválida %q", t)
				}
				b, err := hexv("*")
				if err != nil {
					return nil, err
				}
				for range n {
					s.Memory = append(s.Memory, byte(b))
				}
				sawMem = true
			default:
				b, err := strconv.ParseUint(t, 16, 8)
				if err != nil {
					return nil, fmt.Errorf("z81: byte inválido %q", t)
				}
				s.Memory = append(s.Memory, byte(b))
				sawMem = true
			}
		case "ADVANCED":
			switch t {
			case "ROM":
				s.ROM = next()
			case "M1NOT", "FLOATING_POINT_FIX", "FRAME_RATE_60HZ":
				v, err := hexv(t)
				if err != nil {
					return nil, err
				}
				switch t {
				case "M1NOT":
					s.M1Not, s.HasM1Not = v != 0, true
				case "FLOATING_POINT_FIX":
					s.FPFix = v != 0
				default:
					s.NTSC, s.HasNTSC = v != 0, true
				}
			}
		}
	}
	if !sawCPU || !sawMem {
		return nil, fmt.Errorf("z81: não é um snapshot do ZX81 (faltam [CPU] ou [MEMORY])")
	}
	if s.Start+len(s.Memory) > 0x10000 { // some files (VB81) have bytes past FFFFh: dropped
		s.Memory = s.Memory[:0x10000-s.Start]
	}
	return s, nil
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Text writes the snapshot as EightyOne does (save_snap_zx81).
func (s *Snapshot) Text() string {
	var b strings.Builder
	model := s.Model
	if model == "" {
		model = "ZX81"
	}
	fmt.Fprintf(&b, "[MACHINE]\nMODEL %s\n", model)
	fmt.Fprintf(&b, "\n[CPU]\n")
	fmt.Fprintf(&b, "PC %04X    SP  %04X\n", s.PC, s.SP)
	fmt.Fprintf(&b, "HL %04X    HL_ %04X\n", s.HL, s.HL_)
	fmt.Fprintf(&b, "DE %04X    DE_ %04X\n", s.DE, s.DE_)
	fmt.Fprintf(&b, "BC %04X    BC_ %04X\n", s.BC, s.BC_)
	fmt.Fprintf(&b, "AF %04X    AF_ %04X\n", s.AF, s.AF_)
	fmt.Fprintf(&b, "IX %04X    IY  %04X\n", s.IX, s.IY)
	fmt.Fprintf(&b, "IR %04X\n", s.IR)
	fmt.Fprintf(&b, "IM %02X      IF1 %02X\n", s.IM, s.IFF1)
	fmt.Fprintf(&b, "HT %02X      IF2 %02X\n", b2i(s.Halted), s.IFF2)
	fmt.Fprintf(&b, "\n[ZX81]\n")
	fmt.Fprintf(&b, "NMI %02X     SYNC %02X\n", b2i(s.NMI), b2i(s.Sync))
	fmt.Fprintf(&b, "LINE %03X\n", s.Line)
	fmt.Fprintf(&b, "\n[MEMORY]\n")
	pack := s.RAMPack
	if pack == "" {
		pack = "16K"
	}
	fmt.Fprintf(&b, "RAM_PACK %s\n", pack)
	fmt.Fprintf(&b, "8K_RAM_ENABLED %02X\n", b2i(s.LowRAM))
	fmt.Fprintf(&b, "ROM_PROTECTED %02X\n", 0)
	fmt.Fprintf(&b, "MEMRANGE %04X %04X\n", s.Start, s.Start+len(s.Memory)-1)
	for a := 0; a < len(s.Memory); {
		c := s.Memory[a]
		n := 1
		for a+n < len(s.Memory) && s.Memory[a+n] == c {
			n++
		}
		if n > 1 {
			fmt.Fprintf(&b, "*%04X %02X ", n, c)
		} else {
			fmt.Fprintf(&b, "%02X ", c)
		}
		a += n
	}
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "\n[ADVANCED]\n")
	if s.ROM != "" {
		fmt.Fprintf(&b, "ROM %s\n", s.ROM)
	}
	fmt.Fprintf(&b, "PROTECT_ROM %02X\n", 1)
	fmt.Fprintf(&b, "M1NOT %02X\n", b2i(s.M1Not))
	fmt.Fprintf(&b, "FLOATING_POINT_FIX %02X\n", b2i(s.FPFix))
	fmt.Fprintf(&b, "FRAME_RATE_60HZ %02X\n", b2i(s.NTSC))
	return b.String()
}

// PackK is the RAM pack in K (None = 0), or -1 if the name is unknown.
func PackK(name string) int {
	if strings.EqualFold(name, "None") {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSuffix(strings.ToUpper(name), "K"))
	if err != nil {
		return -1
	}
	return n
}

// PackName is the name of a RAM pack of k K.
func PackName(k int) string {
	if k == 0 {
		return "None"
	}
	return strconv.Itoa(k) + "K"
}
