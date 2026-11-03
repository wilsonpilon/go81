// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

// Program files for the BASIC editor (LOAD, SAVE and MERGE in the editor,
// and the window's menu): the lines of a .p, .81, .p81 or BASIC text, and
// back to a .p or to text, without the ZX81.

// ReadProgramLines reads the lines of a ZX81 program file: .p, .81, .p81 or
// BASIC text (.bas, .txt, .b81, .azx).
func ReadProgramLines(path string) ([]basic.Line, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if basic.IsZX80File(path) {
		return nil, fmt.Errorf("%s é um programa do ZX80: o editor é do ZX81", filepath.Base(path))
	}
	if IsBasicText(path) || looksLikeText(data) {
		return basic.ParseText(string(data))
	}
	if tape.IsTapeFile(path) { // the first program of the tape
		if data, err = FirstTapeProgram(data); err != nil {
			return nil, fmt.Errorf("%s: %v", filepath.Base(path), err)
		}
	}
	if strings.EqualFold(filepath.Ext(path), ".p81") {
		_, prog, ok := tape.SplitSaved(data)
		if !ok {
			return nil, fmt.Errorf("%s: não é um .p81", filepath.Base(path))
		}
		data = prog
	}
	prog, err := basic.ProgramFromP(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", filepath.Base(path), err)
	}
	return basic.Lines(prog)
}

// WriteProgramLines writes lines as a .p or, by the extension, as BASIC
// text (.bas, .txt, .b81, .azx). A name without an extension gets .p.
func WriteProgramLines(path string, lines []basic.Line) (string, error) {
	if filepath.Ext(path) == "" {
		path += ".p"
	}
	var data []byte
	if IsBasicText(path) {
		data = []byte(basic.List(lines))
	} else {
		p, err := basic.MakeP(lines, basic.POptions{})
		if err != nil {
			return "", err
		}
		data = p
		if strings.EqualFold(filepath.Ext(path), ".p81") { // the name of the program comes first
			base := filepath.Base(path)
			data = append(tape.ZX81Name(strings.TrimSuffix(base, filepath.Ext(base))), data...)
		}
	}
	return path, os.WriteFile(path, data, 0o644)
}
