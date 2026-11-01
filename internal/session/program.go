// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/internal/basic"
)

// TextExts are the extensions of BASIC programs written as text; they are
// turned into a .p when loaded (see internal/basic).
var TextExts = []string{".bas", ".txt", ".b81"}

// IsBasicText reports whether a file name is a BASIC text program.
func IsBasicText(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, e := range TextExts {
		if ext == e {
			return true
		}
	}
	return false
}

// checkProgramFor refuses a program made for the other machine: a ZX80
// (with its ROM) loads .o/.80 files; the ZX81 family loads the others.
func checkProgramFor(m interface{ ZX80ROM() bool }, name string) error {
	zx80File := basic.IsZX80File(name)
	switch {
	case m.ZX80ROM() && !zx80File:
		return fmt.Errorf("%s: o ZX80 só carrega programas .o ou .80", filepath.Base(name))
	case !m.ZX80ROM() && zx80File:
		return fmt.Errorf("%s é um programa do ZX80: troque o modelo (machine zx80)", filepath.Base(name))
	}
	return nil
}

// ProgramBytes prepares a program file for the tape: a BASIC text becomes
// a .p (the name keeps the file name, with .p). Other files are returned as
// they are.
func ProgramBytes(name string, data []byte) (string, []byte, error) {
	if !IsBasicText(name) {
		return name, data, nil
	}
	p, err := basic.TextToP(string(data), basic.POptions{})
	if err != nil {
		return "", nil, err
	}
	return strings.TrimSuffix(name, filepath.Ext(name)) + ".p", p, nil
}
