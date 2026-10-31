// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne tzx/tzxload.cpp (LoadPFile) and Utils.cpp
// (ConvertASCIIZX81, RemoveExt) — Copyright (C) 2003-2025 Michael D Wynne
// (GPL v2 or later).

package tape

import (
	"fmt"
	"path/filepath"
	"strings"
)

// zxCharTable is EightyOne's ZXCharTable: the index of a character is its
// ZX81 code. Note that '-' maps to code 1 (its first occurrence), as in
// EightyOne.
var zxCharTable = " ----------'£$:?()><=+-*/;,.0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// ZX81Name converts a file name to ZX81 characters with the last one
// inverted (bit 7 set), as EightyOne's ConvertASCIIZX81 does. Conversion
// stops at the first character that has no ZX81 equivalent.
func ZX81Name(name string) []byte {
	table := []rune(zxCharTable)
	var out []byte
	for _, r := range strings.ToUpper(name) {
		if r == '_' {
			r = ' '
		}
		idx := -1
		for i, t := range table {
			if t == r {
				idx = i
				break
			}
		}
		if idx < 0 {
			break
		}
		out = append(out, byte(idx))
	}
	if len(out) > 0 {
		out[len(out)-1] |= 128
	}
	return out
}

// removeExt mirrors EightyOne's RemoveExt: only known tape/snapshot
// extensions are removed.
func removeExt(name string) string {
	if len(name) < 3 {
		return name
	}
	pos := strings.LastIndexByte(name, '.')
	if pos <= 0 {
		return name
	}
	switch name[pos:] {
	case ".p", ".P", ".T81", ".t81", ".P81", ".p81", ".A83", ".a83", ".o", ".O", ".81", ".80",
		".TZX", ".tzx", ".TAP", ".tap", ".WAV", ".wav", ".sna", ".z80", ".Z80", ".ini", ".INI",
		".mdr", ".MDR", ".mdv", ".MDV", ".dsk", ".DSK", ".img", ".IMG", ".opd", ".OPD", ".opu", ".OPU",
		".mgt", ".MGT", ".trd", ".TRD", ".hdf", ".HDF", ".b81", ".B81", ".b82", ".B82", ".txt", ".TXT",
		".bas", ".BAS", ".web", ".WEB", ".pdf", ".PDF", ".ref", ".REF":
		return name[:pos]
	}
	return name
}

// ProgramBlock builds the tape block for a ZX81 program file: for .p and
// .81 files the ZX81 name (taken from the file name) is put in front of the
// data; .p81 files already contain it.
func ProgramBlock(path string, data []byte) ([]byte, error) {
	if len(data) > 65536 {
		data = data[:65536]
	}
	ext := strings.ToUpper(filepath.Ext(path))
	switch ext {
	case ".P", ".81":
		name := ZX81Name(removeExt(filepath.Base(path)))
		if len(name) == 0 {
			return nil, fmt.Errorf("tape: %s: file name has no ZX81 characters", path)
		}
		return append(name, data...), nil
	case ".P81":
		return append([]byte(nil), data...), nil
	default:
		return nil, fmt.Errorf("tape: %s: unsupported file type %s (use .p, .81 or .p81)", path, ext)
	}
}
