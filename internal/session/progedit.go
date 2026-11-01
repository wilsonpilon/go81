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
		{Name: "rem", Usage: "rem <linha> <tamanho> [byte]", Help: "cria (ou troca) uma linha REM com N bytes para código de máquina e mostra o endereço deles", Run: cmdRem},
	} {
		r.Register(c)
	}
}

// programLines returns the lines of the program in the ZX81.
func (s *Session) programLines() ([]basic.Line, error) {
	if err := s.requireZX81("delete, find e rem"); err != nil {
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
