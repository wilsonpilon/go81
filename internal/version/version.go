// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package version guarda o número de versão do Go81 e seus codinomes.
package version

import "fmt"

// Versão atual. Ver SPEC.md §10 para as regras.
const (
	Major = 1
	Minor = 7
	Patch = 11

	MajorName = "Sinclair"   // X: marco da história do ZX81
	MinorName = "Krazy Kong" // Y: jogo do ZX81
	Subtitle  = "Som em WAV" // Z: trabalho do momento
)

// Number devolve "X.Y.Z".
func Number() string {
	return fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
}

// Full devolve a versão com codinomes.
func Full() string {
	return fmt.Sprintf("%s — %s · %s · %q", Number(), MajorName, MinorName, Subtitle)
}
