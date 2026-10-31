// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"strings"
)

// LooksLikeBasic reports whether a text is a BASIC program: every line
// that is not blank or a "#" comment starts with a line number followed by
// a space or a letter ("10 PRINT", "20GOTO 10").
func LooksLikeBasic(text string) bool {
	n := 0
	for _, l := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		i := 0
		for i < len(t) && t[i] >= '0' && t[i] <= '9' {
			i++
		}
		if i == 0 || i == len(t) {
			return false
		}
		if c := t[i]; c != ' ' && c != '\t' && !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z') {
			return false
		}
		n++
	}
	return n > 0
}

// Paste puts a pasted text into the ZX81: a BASIC program (see
// LooksLikeBasic) is turned into a .p and loaded; any other text is typed.
// It returns a short message for the user.
func (s *Session) Paste(text string) (string, error) {
	if LooksLikeBasic(text) {
		if err := s.InsertProgram("colado.bas", []byte(text), true); err != nil {
			return "", fmt.Errorf("programa colado: %v", err)
		}
		return "programa BASIC colado: carregando", nil
	}
	return s.PasteKeys(text), nil
}

// PasteKeys types a pasted text as keys and returns a short message.
func (s *Session) PasteKeys(text string) string {
	n := len([]rune(strings.ReplaceAll(text, "\r", "")))
	skipped := s.QueueText(text)
	if skipped > 0 {
		return fmt.Sprintf("texto colado: %d caracteres digitados, %d sem tecla no ZX81 ignorados", n-skipped, skipped)
	}
	return fmt.Sprintf("texto colado: %d caracteres digitados", n)
}
