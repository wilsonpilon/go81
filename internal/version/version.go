// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package version guarda o número de versão do Go81 e seus codinomes.
package version

import "fmt"

// Versão atual. Ver SPEC.md §10 para as regras.
const (
	Major = 1
	Minor = 9
	Patch = 5

	MajorName = "Sinclair"    // X: marco da história do ZX81
	MinorName = "1K ZX Chess" // Y: jogo do ZX81
	Subtitle  = "Snapshots"   // Z: trabalho do momento
)

// Number devolve "X.Y.Z".
func Number() string {
	return fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
}

// Full devolve a versão com codinomes.
func Full() string {
	return fmt.Sprintf("%s — %s · %s · %q", Number(), MajorName, MinorName, Subtitle)
}
