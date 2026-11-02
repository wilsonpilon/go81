// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/wilsonpilon/go81/internal/basic"
)

// BASIC program utilities (delete, find, rem): they read the program in the
// ZX81, change it as text lines and load it back by the tape. Like a LOAD,
// that clears the variables.

func registerProgramTools(r *Registry) {
	for _, c := range []*Command{
		{Name: "delete", Aliases: []string{"del"}, Usage: "delete <linha> | <de>,<até>", Help: "apaga uma linha ou um bloco de linhas do programa (as variáveis são apagadas, como num LOAD)", Run: cmdDelete},
		{Name: "find", Aliases: []string{"procurar"}, Raw: true, Usage: "find <texto>", Help: "mostra as linhas do programa que contêm o texto (palavras-chave também: find GOTO)", Run: cmdFind},
		{Name: "renum", Aliases: []string{"renumerar"}, Usage: "renum [início [passo [de,até]]]", Help: "renumera o programa (padrão 10 em 10) e corrige os GOTO, GOSUB, RUN, LIST e LLIST com número fixo", Run: cmdRenum},
		{Name: "auto", Usage: "auto [início [passo]]", Help: "numera sozinho as linhas de BASIC digitadas a seguir; uma linha vazia ou um ponto termina", Run: cmdAuto},
		{Name: "rem", Usage: "rem <linha> <tamanho> [byte]", Help: "cria (ou troca) uma linha REM com N bytes para código de máquina e mostra o endereço deles", Run: cmdRem},
	} {
		r.Register(c)
	}
}

// programLines returns the lines of the program in the ZX81.
func (s *Session) programLines() ([]basic.Line, error) {
	if err := s.requireZX81("delete, find, rem, renum e auto"); err != nil {
		return nil, err
	}
	return basic.Lines(basic.ProgramFromMemory(s.m))
}

// replaceProgram loads a new version of the program.
func (s *Session) replaceProgram(lines []basic.Line) error {
	p, err := basic.MakeP(lines, basic.POptions{Collapsed: s.set.RAMK < 4})
	if err != nil {
		return err
	}
	return s.loadProgram("programa.p", p)
}

// parseRange reads "N", "A,B", "A-B" or "A B" (args already split).
func parseRange(a []string) (int, int, error) {
	text := strings.Join(a, ",")
	text = strings.ReplaceAll(text, "-", ",")
	parts := strings.Split(text, ",")
	var nums []int
	for _, p := range parts {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}
		n, err := ParseNumber(p)
		if err != nil || n < 0 || n > basic.MaxLine {
			return 0, 0, fmt.Errorf("número de linha inválido: %s", p)
		}
		nums = append(nums, n)
	}
	switch len(nums) {
	case 1:
		return nums[0], nums[0], nil
	case 2:
		if nums[0] > nums[1] {
			return 0, 0, fmt.Errorf("a primeira linha (%d) vem depois da última (%d)", nums[0], nums[1])
		}
		return nums[0], nums[1], nil
	}
	return 0, 0, errors.New("diga uma linha ou um bloco: 10 ou 10,50")
}

func cmdDelete(s *Session, a []string, out io.Writer) error {
	from, to, err := parseRange(a)
	if err != nil {
		return fmt.Errorf("%v\nuso: delete <linha> | <de>,<até>", err)
	}
	lines, err := s.programLines()
	if err != nil {
		return err
	}
	rest, n := basic.DeleteLines(lines, from, to)
	if n == 0 {
		fmt.Fprintln(out, "nenhuma linha nesse intervalo")
		return nil
	}
	if err := s.replaceProgram(rest); err != nil {
		return err
	}
	fmt.Fprintf(out, "%d linha(s) apagada(s)\n", n)
	return nil
}

func cmdFind(s *Session, a []string, out io.Writer) error {
	text := strings.TrimSpace(strings.Join(a, " "))
	if text == "" {
		return usageErr("find <texto>")
	}
	lines, err := s.programLines()
	if err != nil {
		return err
	}
	found := basic.FindLines(lines, text)
	if len(found) == 0 {
		fmt.Fprintln(out, "nada encontrado")
		return nil
	}
	fmt.Fprint(out, basic.List(found))
	return nil
}

func cmdRem(s *Session, a []string, out io.Writer) error {
	if len(a) < 2 || len(a) > 3 {
		return usageErr("rem <linha> <tamanho> [byte]")
	}
	var v [3]int
	for i, arg := range a {
		n, err := ParseNumber(arg)
		if err != nil {
			return err
		}
		v[i] = n
	}
	if v[2] < 0 || v[2] > 255 {
		return fmt.Errorf("byte fora de 0-255: %d", v[2])
	}
	l, err := basic.RemLine(v[0], v[1], byte(v[2]))
	if err != nil {
		return err
	}
	lines, err := s.programLines()
	if err != nil {
		return err
	}
	lines = basic.PutLine(lines, l)
	if err := s.replaceProgram(lines); err != nil {
		return err
	}
	code := basic.LineAddress(lines, v[0]) + 5 // number, length and REM
	fmt.Fprintf(out, "linha %d REM com %d bytes; o código começa em %d ($%04X)\n", v[0], v[1], code, code)
	return nil
}

func cmdRenum(s *Session, a []string, out io.Writer) error {
	if len(a) > 4 {
		return usageErr("renum [início [passo [de,até]]]")
	}
	start, step, from, to := 10, 10, 0, basic.MaxLine
	for i, arg := range a {
		if i == 2 {
			var err error
			if from, to, err = parseRange(a[2:]); err != nil {
				return fmt.Errorf("%v\nuso: renum [início [passo [de,até]]]", err)
			}
			break
		}
		n, err := ParseNumber(arg)
		if err != nil {
			return err
		}
		if i == 0 {
			start = n
		} else {
			step = n
		}
	}
	lines, err := s.programLines()
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		fmt.Fprintln(out, "não há programa")
		return nil
	}
	res, err := basic.Renumber(lines, start, step, from, to)
	if err != nil {
		return err
	}
	if res.Moved == 0 {
		fmt.Fprintln(out, "nada mudou")
		return nil
	}
	if err := s.replaceProgram(res.Lines); err != nil {
		return err
	}
	fmt.Fprintf(out, "%d linha(s) renumerada(s), %d referência(s) corrigida(s)\n", res.Moved, res.Refs)
	for _, w := range res.Warnings {
		fmt.Fprintln(out, "aviso:", w)
	}
	return nil
}

// autoState is the auto command at work: the lines typed so far.
type autoState struct {
	next, step int
	lines      []basic.Line
}

// Prompt is the shell prompt: "go81> ", or the next line number while
// auto is on.
func (s *Session) Prompt() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.auto != nil {
		return fmt.Sprintf("auto %d> ", s.auto.next)
	}
	return "go81> "
}

// Auto reports whether auto is on: the lines typed are BASIC, not commands
// (an empty line ends it).
func (s *Session) Auto() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.auto != nil
}

func cmdAuto(s *Session, a []string, out io.Writer) error {
	if len(a) > 2 {
		return usageErr("auto [início [passo]]")
	}
	lines, err := s.programLines()
	if err != nil {
		return err
	}
	st := &autoState{next: 10, step: 10}
	if len(lines) > 0 {
		st.next = lines[len(lines)-1].Num + 10
	}
	for i, arg := range a {
		n, err := ParseNumber(arg)
		if err != nil {
			return err
		}
		if n < 1 || n > basic.MaxLine {
			return fmt.Errorf("fora de 1-%d: %d", basic.MaxLine, n)
		}
		if i == 0 {
			st.next = n
		} else {
			st.step = n
		}
	}
	if len(a) == 1 && len(lines) > 0 {
		st.step = 10
	}
	if st.next > basic.MaxLine {
		return fmt.Errorf("não cabem mais linhas depois da %d", lines[len(lines)-1].Num)
	}
	s.auto = st
	fmt.Fprintf(out, "digite as linhas de BASIC (sem número); uma linha vazia ou um ponto termina. Um número no começo troca a numeração.\n")
	return nil
}

// autoLine takes a line typed while auto is on.
func (s *Session) autoLine(line string, out io.Writer) error {
	st := s.auto
	text := strings.TrimSpace(line)
	if text == "" || text == "." {
		s.auto = nil
		if len(st.lines) == 0 {
			fmt.Fprintln(out, "auto terminado (nenhuma linha)")
			return nil
		}
		prog, err := s.programLines()
		if err != nil {
			return err
		}
		for _, l := range st.lines {
			prog = basic.PutLine(prog, l)
		}
		if err := s.replaceProgram(prog); err != nil {
			return err
		}
		fmt.Fprintf(out, "auto terminado: %d linha(s) no programa\n", len(st.lines))
		return nil
	}
	if text[0] < '0' || text[0] > '9' {
		text = fmt.Sprintf("%d %s", st.next, text)
	}
	l, err := basic.ParseLine(text)
	if err != nil {
		return err
	}
	if len(l.Body) == 0 {
		return fmt.Errorf("a linha %d está vazia", l.Num)
	}
	st.lines = append(st.lines, l)
	st.next = l.Num + st.step
	if st.next > basic.MaxLine {
		return s.autoLine("", out) // no more numbers: finish
	}
	return nil
}

// ProgramLines returns the program in the ZX81 (for the BASIC editor).
func (s *Session) ProgramLines() ([]basic.Line, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.programLines()
}

// SetProgram loads a new version of the program by the tape (the BASIC
// editor hands its program back this way; variables are cleared, as in a
// LOAD).
func (s *Session) SetProgram(lines []basic.Line) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireZX81("o editor de BASIC"); err != nil {
		return err
	}
	return s.replaceProgram(lines)
}
