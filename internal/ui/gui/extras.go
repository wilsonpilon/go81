// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package gui

import (
	"path/filepath"
	"strings"
)

// programExts are the files the window opens (menu, recent files, drag and
// drop).
var programExts = []string{".p", ".81", ".p81", ".bas", ".txt", ".b81"}

// isProgram reports whether the file looks like a program the ZX81 loads.
func isProgram(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range programExts {
		if ext == e {
			return true
		}
	}
	return false
}

// maxRecent is how many files the "Abrir recente" menu keeps.
const maxRecent = 8

// addRecent puts path at the top of the list, without repeating it.
func addRecent(list []string, path string) []string {
	out := []string{path}
	for _, p := range list {
		if !samePath(p, path) && len(out) < maxRecent {
			out = append(out, p)
		}
	}
	return out
}

func samePath(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if filepath.Separator == '\\' { // Windows: names are not case sensitive
		return strings.EqualFold(a, b)
	}
	return a == b
}

// keyCapInfo is one key of the on-screen keyboard: the key (as zx81.ParseChord
// knows it), what SHIFT gives (printed in red on the ZX81) and the keyword of
// the K cursor.
type keyCapInfo struct {
	Key, Shifted, Keyword string
}

// zx81Keyboard is the ZX81 keyboard, row by row.
var zx81Keyboard = [][]keyCapInfo{
	{{"1", "EDIT", ""}, {"2", "AND", ""}, {"3", "THEN", ""}, {"4", "TO", ""}, {"5", "←", ""},
		{"6", "↓", ""}, {"7", "↑", ""}, {"8", "→", ""}, {"9", "GRAPHICS", ""}, {"0", "RUBOUT", ""}},
	{{"Q", `""`, "PLOT"}, {"W", "OR", "UNPLOT"}, {"E", "STEP", "REM"}, {"R", "<=", "RUN"}, {"T", "<>", "RAND"},
		{"Y", ">=", "RETURN"}, {"U", "$", "IF"}, {"I", "(", "INPUT"}, {"O", ")", "POKE"}, {"P", `"`, "PRINT"}},
	{{"A", "STOP", "NEW"}, {"S", "LPRINT", "SAVE"}, {"D", "SLOW", "DIM"}, {"F", "FAST", "FOR"}, {"G", "LLIST", "GOTO"},
		{"H", "**", "GOSUB"}, {"J", "-", "LOAD"}, {"K", "+", "LIST"}, {"L", "=", "LET"}, {"ENTER", "FUNCTION", ""}},
	{{"SHIFT", "", ""}, {"Z", ":", "COPY"}, {"X", ";", "CLEAR"}, {"C", "?", "CONT"}, {"V", "/", "CLS"},
		{"B", "*", "SCROLL"}, {"N", "<", "NEXT"}, {"M", ">", "PAUSE"}, {".", ",", ""}, {"SPACE", "£", "BREAK"}},
}
