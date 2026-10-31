// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package assets holds files embedded in the go81 executable, such as ROMs.
package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed roms/*.rom
var roms embed.FS

// ROM returns an embedded ROM by name, with or without the .rom extension
// (e.g. "zx81.edition3").
func ROM(name string) ([]byte, error) {
	if !strings.HasSuffix(name, ".rom") {
		name += ".rom"
	}
	b, err := roms.ReadFile("roms/" + name)
	if err != nil {
		return nil, fmt.Errorf("ROM %q not found (available: %s)", name, strings.Join(ROMNames(), ", "))
	}
	return b, nil
}

// ROMNames lists the embedded ROMs (without extension).
func ROMNames() []string {
	ents, _ := fs.ReadDir(roms, "roms")
	var names []string
	for _, e := range ents {
		names = append(names, strings.TrimSuffix(e.Name(), ".rom"))
	}
	sort.Strings(names)
	return names
}
