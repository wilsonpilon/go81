// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package basiced is the BASIC editor in the style of the ZX81 (SPEC
// §6.1.1): the listing on top, with the current line marked by ">", and
// the line being typed at the bottom, on a 32×24 screen of ZX81
// characters. A line with a number goes into the program (a number alone
// deletes the line); a line without a number is an editor command (RENUM,
// AUTO, DELETE, FIND, LIST, EDIT, REM, NEW, HELP).
//
// This is the "direct typing" mode: the line is typed as text, as in a
// .bas file, and turned into keywords on ENTER.
package basiced

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/wilsonpilon/go81/internal/basic"
)

// Screen size, in characters.
const (
	Cols = 32
	Rows = 24
)

// Editor is the state of the editor.
type Editor struct {
	Lines   []basic.Line // the program, in order of number
	Current int          // the line with the > cursor (0 = none)
	Changed bool         // the program was changed

	input  []rune
	pos    int // cursor position in input
	msg    string
	msgErr bool
	top    int // index of the first line listed
	auto   *autoState
	find   string // FIND: only the lines with this text are listed
}

type autoState struct{ next, step int }

// New starts the editor on a program.
func New(lines []basic.Line) *Editor {
	e := &Editor{Lines: append([]basic.Line(nil), lines...)}
	if len(lines) > 0 {
		e.Current = lines[0].Num
	}
	e.say("EDITOR DE BASIC. HELP MOSTRA OS COMANDOS")
	return e
}

// Input is the line being typed.
func (e *Editor) Input() string { return string(e.input) }

// Message is the last message (and whether it is an error).
func (e *Editor) Message() (string, bool) { return e.msg, e.msgErr }

func (e *Editor) say(f string, a ...any)  { e.msg, e.msgErr = fmt.Sprintf(f, a...), false }
func (e *Editor) fail(f string, a ...any) { e.msg, e.msgErr = fmt.Sprintf(f, a...), true }

func (e *Editor) setInput(s string) {
	e.input = []rune(s)
	e.pos = len(e.input)
}

// TypeRune types a character at the cursor (letters become capitals: the
// ZX81 has no small letters).
func (e *Editor) TypeRune(r rune) {
	if r < ' ' || r == 0x7F {
		return
	}
	r = unicode.ToUpper(r)
	e.input = append(e.input[:e.pos], append([]rune{r}, e.input[e.pos:]...)...)
	e.pos++
}

// Key handles an editing key, by its Fyne name: Return, BackSpace, Delete,
// Left, Right, Home, End, Up, Down, Prior/Next (pages), Tab (EDIT: the
// current line comes down to be changed), Escape (clears the line).
func (e *Editor) Key(name string) {
	switch name {
	case "Return", "Enter":
		e.enter()
	case "BackSpace":
		if e.pos > 0 {
			e.input = append(e.input[:e.pos-1], e.input[e.pos:]...)
			e.pos--
		}
	case "Delete":
		if e.pos < len(e.input) {
			e.input = append(e.input[:e.pos], e.input[e.pos+1:]...)
		}
	case "Left":
		if e.pos > 0 {
			e.pos--
		}
	case "Right":
		if e.pos < len(e.input) {
			e.pos++
		}
	case "Home":
		e.pos = 0
	case "End":
		e.pos = len(e.input)
	case "Up":
		e.moveCurrent(-1)
	case "Down":
		e.moveCurrent(1)
	case "Prior":
		e.moveCurrent(-10)
	case "Next":
		e.moveCurrent(10)
	case "Tab":
		e.edit(e.Current)
	case "Escape":
		e.setInput("")
		if e.auto != nil {
			e.auto = nil
			e.say("AUTO TERMINADO")
		}
		e.find = ""
	}
}

// listed are the lines shown (all, or those FIND chose).
func (e *Editor) listed() []basic.Line {
	if e.find == "" {
		return e.Lines
	}
	return basic.FindLines(e.Lines, e.find)
}

func (e *Editor) index(num int) int {
	for i, l := range e.listed() {
		if l.Num >= num {
			return i
		}
	}
	return len(e.listed()) - 1
}

func (e *Editor) moveCurrent(d int) {
	ls := e.listed()
	if len(ls) == 0 {
		return
	}
	i := max(0, min(len(ls)-1, e.index(e.Current)+d))
	e.Current = ls[i].Num
}

// edit brings a line down to the input, to be changed.
func (e *Editor) edit(num int) {
	for _, l := range e.Lines {
		if l.Num == num {
			e.setInput(strings.TrimSpace(strconv.Itoa(l.Num) + " " + basic.Text(l.Body)))
			return
		}
	}
	e.fail("NAO HA A LINHA %d", num)
}

func (e *Editor) enter() {
	text := strings.TrimSpace(string(e.input))
	if text == "" {
		if e.auto != nil {
			e.auto = nil
			e.say("AUTO TERMINADO")
		}
		return
	}
	if text[0] >= '0' && text[0] <= '9' {
		e.programLine(text)
		return
	}
	e.command(text)
}

// programLine puts (or deletes) a numbered line.
func (e *Editor) programLine(text string) {
	l, err := basic.ParseLine(text)
	if err != nil {
		e.fail("%s", strings.ToUpper(err.Error()))
		return
	}
	if l.Num < 1 || l.Num > basic.MaxLine {
		e.fail("LINHA FORA DE 1-%d", basic.MaxLine)
		return
	}
	if len(l.Body) == 0 {
		var n int
		e.Lines, n = basic.DeleteLines(e.Lines, l.Num, l.Num)
		if n == 0 {
			e.fail("NAO HA A LINHA %d", l.Num)
		} else {
			e.Changed = true
			e.say("LINHA %d APAGADA", l.Num)
		}
		e.setInput("")
		return
	}
	e.Lines = basic.PutLine(e.Lines, l)
	e.Current, e.Changed = l.Num, true
	e.msg = ""
	e.setInput("")
	if e.auto != nil {
		e.auto.next = l.Num + e.auto.step
		if e.auto.next > basic.MaxLine {
			e.auto = nil
			e.say("AUTO TERMINADO: ACABARAM OS NUMEROS")
			return
		}
		e.setInput(strconv.Itoa(e.auto.next) + " ")
	}
}

// Help is the text of HELP.
const Help = "LINHA COM NUMERO: ENTRA NO PROGRAMA (SO O NUMERO APAGA). " +
	"COMANDOS: RENUM [INICIO [PASSO]], AUTO [INICIO [PASSO]], DELETE A[,B], FIND TEXTO, " +
	"LIST [N], EDIT [N], REM N TAMANHO, NEW. " +
	"TECLAS: CIMA/BAIXO ESCOLHEM A LINHA, TAB TRAZ A LINHA PARA EDITAR, ESC LIMPA."

func numbers(args []string, def ...int) ([]int, error) {
	out := append([]int(nil), def...)
	for i, a := range args {
		n, err := strconv.Atoi(a)
		if err != nil {
			return nil, fmt.Errorf("NUMERO INVALIDO: %s", a)
		}
		if i < len(out) {
			out[i] = n
		} else {
			out = append(out, n)
		}
	}
	return out, nil
}

// command runs an editor command (a line without a number).
func (e *Editor) command(text string) {
	f := strings.Fields(strings.NewReplacer(",", " ", "-", " ").Replace(text))
	name, args := strings.ToUpper(f[0]), f[1:]
	ok := true
	switch name {
	case "HELP", "AJUDA", "?":
		e.say("%s", Help)
	case "RENUM":
		n, err := numbers(args, 10, 10)
		if err != nil || len(n) > 2 {
			e.fail("USO: RENUM [INICIO [PASSO]]")
			return
		}
		res, err := basic.Renumber(e.Lines, n[0], n[1], 0, basic.MaxLine)
		if err != nil {
			e.fail("%s", strings.ToUpper(err.Error()))
			return
		}
		i := lineIndex(e.Lines, e.Current) // the order stays: same index
		e.Lines, e.Changed = res.Lines, e.Changed || res.Moved > 0
		if i >= 0 {
			e.Current = res.Lines[i].Num
		}
		e.say("%d LINHAS RENUMERADAS, %d REFERENCIAS", res.Moved, res.Refs)
		if len(res.Warnings) > 0 {
			e.fail("AVISO: %s", strings.ToUpper(res.Warnings[0]))
		}
	case "AUTO":
		start := 10
		if len(e.Lines) > 0 {
			start = e.Lines[len(e.Lines)-1].Num + 10
		}
		n, err := numbers(args, start, 10)
		if err != nil || len(n) > 2 || n[0] < 1 || n[0] > basic.MaxLine || n[1] < 1 {
			e.fail("USO: AUTO [INICIO [PASSO]]")
			return
		}
		e.auto = &autoState{next: n[0], step: n[1]}
		e.say("AUTO: ENTER NUMA LINHA VAZIA OU ESC TERMINA")
		e.setInput(strconv.Itoa(n[0]) + " ")
		return
	case "DELETE", "DEL":
		n, err := numbers(args)
		if err != nil || len(n) < 1 || len(n) > 2 {
			e.fail("USO: DELETE A[,B]")
			return
		}
		to := n[0]
		if len(n) == 2 {
			to = n[1]
		}
		var k int
		e.Lines, k = basic.DeleteLines(e.Lines, n[0], to)
		e.Changed = e.Changed || k > 0
		e.say("%d LINHAS APAGADAS", k)
	case "FIND", "PROCURAR":
		if len(args) == 0 {
			e.find = ""
			e.say("LISTANDO TODAS AS LINHAS")
			break
		}
		e.find = strings.TrimSpace(text[len(f[0]):])
		ls := e.listed()
		if len(ls) == 0 {
			e.find = ""
			e.fail("NADA ENCONTRADO")
			break
		}
		e.Current = ls[0].Num
		e.say("%d LINHAS COM %s (ESC MOSTRA TODAS)", len(ls), strings.ToUpper(e.find))
	case "LIST":
		e.find = ""
		n, err := numbers(args)
		if err != nil || len(n) > 1 {
			e.fail("USO: LIST [N]")
			return
		}
		if len(n) == 1 && len(e.Lines) > 0 {
			e.Current = e.Lines[e.index(n[0])].Num
			e.top = e.index(n[0])
		}
		e.msg = ""
	case "EDIT":
		n, err := numbers(args, e.Current)
		if err != nil || len(n) > 1 {
			e.fail("USO: EDIT [N]")
			return
		}
		e.edit(n[0])
		return
	case "REM":
		n, err := numbers(args, 0, 0, 0)
		if err != nil || len(args) < 2 || len(n) > 3 || n[2] < 0 || n[2] > 255 {
			e.fail("USO: REM LINHA TAMANHO [BYTE]")
			return
		}
		l, err := basic.RemLine(n[0], n[1], byte(n[2]))
		if err != nil {
			e.fail("%s", strings.ToUpper(err.Error()))
			return
		}
		e.Lines = basic.PutLine(e.Lines, l)
		e.Current, e.Changed = l.Num, true
		code := basic.LineAddress(e.Lines, l.Num) + 5
		e.say("REM COM %d BYTES; O CODIGO COMECA EM %d", n[1], code)
	case "NEW":
		e.Lines, e.Current, e.Changed, e.top = nil, 0, true, 0
		e.say("PROGRAMA APAGADO")
	default:
		ok = false
		e.fail("COMANDO DESCONHECIDO. A LINHA PRECISA DE NUMERO (HELP)")
	}
	if ok {
		e.setInput("")
	}
}

func lineIndex(lines []basic.Line, num int) int {
	for i, l := range lines {
		if l.Num == num {
			return i
		}
	}
	return -1
}

// Paste takes text from the clipboard: one line goes into the input at
// the cursor; several lines are entered one by one, as if typed with
// ENTER (a program in text). It stops at the first line with an error,
// which stays in the input.
func (e *Editor) Paste(text string) {
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.TrimRight(text, "\n")
	if !strings.Contains(text, "\n") {
		for _, r := range text {
			e.TypeRune(r)
		}
		return
	}
	n := 0
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		e.setInput(strings.TrimSpace(line))
		e.enter()
		if e.msgErr {
			return
		}
		n++
	}
	e.say("%d LINHAS COLADAS", n)
}
