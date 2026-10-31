// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// chordForKey translates a terminal key event into a ZX81 chord ("" if the
// key has no ZX81 meaning).
func chordForKey(ev *tcell.EventKey) string {
	switch ev.Key() {
	case tcell.KeyEnter:
		return "ENTER"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "SHIFT+0" // RUBOUT
	case tcell.KeyLeft:
		return "SHIFT+5"
	case tcell.KeyDown:
		return "SHIFT+6"
	case tcell.KeyUp:
		return "SHIFT+7"
	case tcell.KeyRight:
		return "SHIFT+8"
	case tcell.KeyEscape:
		return "SHIFT+1" // EDIT
	case tcell.KeyRune:
	default:
		return ""
	}
	r := ev.Rune()
	up := strings.ToUpper(string(r))
	isLetterOrDigit := len(up) == 1 && (up[0] >= 'A' && up[0] <= 'Z' || up[0] >= '0' && up[0] <= '9')
	if ev.Modifiers()&tcell.ModAlt != 0 && isLetterOrDigit {
		return "SHIFT+" + up // Alt+key = SHIFT+key (FAST, STOP, GRAPHICS, ...)
	}
	if r == '\n' {
		return ""
	}
	return zx81.ChordForRune(r)
}

// helpText is shown by F1.
const helpText = `Go81 — ZX81 no terminal

Digite normalmente: letras, números e símbolos viram as teclas do ZX81
(" = SHIFT+P, + = SHIFT+K, = = SHIFT+L, ( ) $ : ; ? / * < > , - também).

  Enter       ENTER
  Backspace   RUBOUT (apaga)
  Setas       move o cursor
  Esc         EDIT (traz a linha para edição)
  Alt+tecla   SHIFT+tecla: Alt+F = FAST, Alt+D = SLOW,
              Alt+A = STOP, Alt+9 = GRAPHICS, Alt+4 = TO ...

No cursor K cada tecla é um comando: P = PRINT, R = RUN, L = LET,
G = GOTO, S = SAVE, J = LOAD, F = FOR, N = NEXT, Q = PLOT.

  F1   esta ajuda          F5   pausa / continua
  F6   turbo (10×)         F8   reinicia (desliga e liga)
  F9   tela de texto / pixels
  F10  sair (ou Ctrl+C)`
