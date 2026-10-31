// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package gui

// keyMap maps PC key names (as Fyne reports them) to ZX81 key chords.
// Keys sit where they are on the ZX81 (Shift is the ZX81 SHIFT); a few PC
// keys are shortcuts for ZX81 combinations.
var keyMap = map[string]string{
	"Return": "ENTER", "Enter": "ENTER", "KP_Enter": "ENTER",
	"Space":      "SPACE",
	"LeftShift":  "SHIFT",
	"RightShift": "SHIFT",
	".":          ".",

	// Shortcuts
	"BackSpace": "SHIFT+0", // RUBOUT
	"Left":      "SHIFT+5",
	"Down":      "SHIFT+6",
	"Up":        "SHIFT+7",
	"Right":     "SHIFT+8",
	",":         "SHIFT+.",
	"Escape":    "SHIFT+1", // EDIT
}

func init() {
	for c := 'A'; c <= 'Z'; c++ {
		keyMap[string(c)] = string(c)
	}
	for c := '0'; c <= '9'; c++ {
		keyMap[string(c)] = string(c)
	}
}

// KeyHelp describes the keyboard for the Help menu.
const KeyHelp = `As teclas do PC ficam na mesma posição das teclas do ZX81.
Shift do PC = SHIFT do ZX81. Ex.: Shift+P = "  ·  Shift+L = =

Atalhos:
  Backspace  → RUBOUT (SHIFT+0)
  Setas      → cursor (SHIFT+5/6/7/8)
  Vírgula    → , (SHIFT+.)
  Esc        → EDIT (SHIFT+1)

Palavras-chave: no cursor K, cada tecla é um comando
(P = PRINT, R = RUN, G = GOTO, L = LET, S = SAVE, J = LOAD).`
