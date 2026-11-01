// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/z80"
)

// Variables, expressions and blocks for scripts (.g81) and the shell.
//
//	set nome valor            calc nome expressão      print expressão
//	capture nome comando      assert expressão         fail mensagem
//	waitfor quadros expressão
//	if expressão ... [else ...] end
//	while expressão ... end   repeat N ... end         break
//
// ${nome} is replaced by the variable in any command line.

var varName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

var varRef = regexp.MustCompile(`\$\{([^}]*)\}`)

// expandVars replaces ${name}. Caller holds s.mu.
func (s *Session) expandVars(line string) (string, error) {
	var err error
	out := varRef.ReplaceAllStringFunc(line, func(m string) string {
		name := m[2 : len(m)-1]
		v, ok := s.vars[name]
		if !ok && err == nil {
			err = fmt.Errorf("variável não definida: %s", name)
		}
		return v
	})
	return out, err
}

// SetVar sets a variable (front-ends, tests).
func (s *Session) SetVar(name, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars[name] = value
}

// regValue reads a register by name.
func regValue(c *z80.CPU, r string) (int, bool) {
	switch r {
	case "A":
		return int(c.A), true
	case "F":
		return int(c.F), true
	case "B":
		return int(c.B), true
	case "C":
		return int(c.C), true
	case "D":
		return int(c.D), true
	case "E":
		return int(c.E), true
	case "H":
		return int(c.H), true
	case "L":
		return int(c.L), true
	case "AF":
		return int(c.AF()), true
	case "BC":
		return int(c.BC()), true
	case "DE":
		return int(c.DE()), true
	case "HL":
		return int(c.HL()), true
	case "IX":
		return int(c.IX), true
	case "IY":
		return int(c.IY), true
	case "SP":
		return int(c.SP), true
	case "PC":
		return int(c.PC), true
	case "I":
		return int(c.I), true
	case "R":
		return int(c.RegR()), true
	case "AF'":
		return int(c.AF_), true
	case "BC'":
		return int(c.BC_), true
	case "DE'":
		return int(c.DE_), true
	case "HL'":
		return int(c.HL_), true
	}
	return 0, false
}

// errBreak ends the innermost while or repeat.
var errBreak = errors.New("break fora de um while ou repeat")

func registerScriptCommands(r *Registry) {
	for _, c := range []*Command{
		{Name: "set", Raw: true, Usage: "set [nome valor]", Help: "guarda um valor numa variável (${nome} nos comandos); sem nada, lista as variáveis", Run: cmdSet},
		{Name: "unset", Usage: "unset nome", Help: "apaga uma variável", Run: func(s *Session, a []string, out io.Writer) error {
			if len(a) != 1 {
				return usageErr("unset nome")
			}
			delete(s.vars, a[0])
			return nil
		}},
		{Name: "calc", Raw: true, Usage: "calc nome expressão", Help: "calcula uma expressão e guarda o resultado (ex.: calc i i+1)", Run: cmdCalc},
		{Name: "print", Raw: true, Usage: "print expressão", Help: `mostra o valor de uma expressão (ex.: print peek(16396)+256*peek(16397))`, Run: func(s *Session, a []string, out io.Writer) error {
			v, err := s.Eval(a[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(out, v)
			return nil
		}},
		{Name: "capture", Raw: true, Usage: "capture nome comando", Help: "executa um comando e guarda a saída numa variável", Run: cmdCapture},
		{Name: "assert", Raw: true, Usage: "assert expressão", Help: "para o script com erro se a expressão for falsa", Run: func(s *Session, a []string, out io.Writer) error {
			v, err := s.Eval(a[0])
			if err != nil {
				return err
			}
			if !v.True() {
				return fmt.Errorf("assert falhou: %s", a[0])
			}
			return nil
		}},
		{Name: "fail", Raw: true, Usage: "fail mensagem", Help: "para o script com esta mensagem de erro", Run: func(s *Session, a []string, out io.Writer) error {
			if a[0] == "" {
				return errors.New("fail")
			}
			return errors.New(a[0])
		}},
		{Name: "waitfor", Raw: true, Usage: "waitfor quadros expressão", Help: `roda o ZX81 até a expressão ser verdadeira (ex.: waitfor 500 screen("0/0"))`, Run: cmdWaitFor},
		{Name: "if", Raw: true, Usage: "if expressão … [else …] end", Help: "bloco condicional (em scripts .g81)", Run: blockOnly},
		{Name: "while", Raw: true, Usage: "while expressão … end", Help: "repete enquanto a expressão for verdadeira (em scripts)", Run: blockOnly},
		{Name: "repeat", Raw: true, Usage: "repeat N … end", Help: "repete N vezes (em scripts)", Run: blockOnly},
		{Name: "break", Usage: "break", Help: "sai do while ou repeat", Run: func(s *Session, a []string, out io.Writer) error { return errBreak }},
		{Name: "else", Usage: "else", Help: "parte do if", Run: blockOnly},
		{Name: "end", Usage: "end", Help: "fecha if, while ou repeat", Run: blockOnly},
	} {
		r.Register(c)
	}
}

func blockOnly(s *Session, a []string, out io.Writer) error {
	return errors.New("if, while e repeat só funcionam dentro de scripts (.g81)")
}

func cmdSet(s *Session, a []string, out io.Writer) error {
	name, value, _ := strings.Cut(strings.TrimSpace(a[0]), " ")
	if name == "" {
		names := make([]string, 0, len(s.vars))
		for k := range s.vars {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			fmt.Fprintf(out, "%s = %s\n", k, s.vars[k])
		}
		return nil
	}
	if !varName.MatchString(name) {
		return fmt.Errorf("nome de variável inválido: %s", name)
	}
	s.vars[name] = strings.TrimSpace(value)
	return nil
}

func cmdCalc(s *Session, a []string, out io.Writer) error {
	name, expr, ok := strings.Cut(strings.TrimSpace(a[0]), " ")
	if !ok || !varName.MatchString(name) {
		return usageErr("calc nome expressão")
	}
	v, err := s.Eval(expr)
	if err != nil {
		return err
	}
	s.vars[name] = v.String()
	return nil
}

func cmdCapture(s *Session, a []string, out io.Writer) error {
	name, cmd, ok := strings.Cut(strings.TrimSpace(a[0]), " ")
	if !ok || !varName.MatchString(name) {
		return usageErr("capture nome comando")
	}
	var buf bytes.Buffer
	if err := s.execLocked(cmd, &buf); err != nil {
		return err
	}
	s.vars[name] = strings.TrimRight(buf.String(), "\n")
	return nil
}

func cmdWaitFor(s *Session, a []string, out io.Writer) error {
	n, expr, ok := strings.Cut(strings.TrimSpace(a[0]), " ")
	max, err := ParseNumber(n)
	if !ok || err != nil || max < 1 {
		return usageErr("waitfor quadros expressão")
	}
	var evalErr error
	done := s.waitUntil(func() bool {
		v, err := s.Eval(expr)
		if err != nil {
			evalErr = err
			return true
		}
		return v.True()
	}, max)
	if evalErr != nil {
		return evalErr
	}
	if !done {
		return fmt.Errorf("waitfor: %s não ficou verdadeira em %d quadros", expr, max)
	}
	return nil
}

// A script is a tree of blocks.
type scriptNode struct {
	line   int
	kind   string // "", "if", "while", "repeat"
	text   string // the command, or the condition / count
	body   []scriptNode
	orElse []scriptNode
}

func firstWord(l string) (string, string) {
	w, rest, _ := strings.Cut(strings.TrimSpace(l), " ")
	return strings.ToLower(w), strings.TrimSpace(rest)
}

// parseScript builds the blocks of a script.
func parseScript(lines []string, name string) ([]scriptNode, error) {
	type frame struct {
		node   scriptNode
		inElse bool
	}
	root := []scriptNode{}
	var stack []frame
	add := func(n scriptNode) {
		if len(stack) == 0 {
			root = append(root, n)
			return
		}
		f := &stack[len(stack)-1]
		if f.inElse {
			f.node.orElse = append(f.node.orElse, n)
		} else {
			f.node.body = append(f.node.body, n)
		}
	}
	for i, raw := range lines {
		l := strings.TrimSpace(raw)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		w, rest := firstWord(l)
		switch w {
		case "if", "while", "repeat":
			if rest == "" {
				return nil, fmt.Errorf("%s:%d: %s sem condição", name, i+1, w)
			}
			stack = append(stack, frame{node: scriptNode{line: i + 1, kind: w, text: rest}})
		case "else":
			if len(stack) == 0 || stack[len(stack)-1].node.kind != "if" || stack[len(stack)-1].inElse {
				return nil, fmt.Errorf("%s:%d: else sem if", name, i+1)
			}
			stack[len(stack)-1].inElse = true
		case "end":
			if len(stack) == 0 {
				return nil, fmt.Errorf("%s:%d: end sem if, while ou repeat", name, i+1)
			}
			n := stack[len(stack)-1].node
			stack = stack[:len(stack)-1]
			add(n)
		default:
			add(scriptNode{line: i + 1, text: l})
		}
	}
	if len(stack) > 0 {
		n := stack[len(stack)-1].node
		return nil, fmt.Errorf("%s:%d: %s sem end", name, n.line, n.kind)
	}
	return root, nil
}

// maxLoop guards against endless loops in scripts.
const maxLoop = 1_000_000

// runScript runs the blocks. Caller holds s.mu.
func (s *Session) runScript(nodes []scriptNode, name string, out io.Writer) error {
	at := func(n scriptNode, err error) error {
		if errors.Is(err, ErrQuit) || errors.Is(err, errBreak) {
			return err
		}
		var le *lineError
		if errors.As(err, &le) {
			return err
		}
		return &lineError{fmt.Sprintf("%s:%d", name, n.line), err}
	}
	cond := func(n scriptNode) (bool, error) {
		v, err := s.Eval(n.text)
		if err != nil {
			return false, at(n, err)
		}
		return v.True(), nil
	}
	for _, n := range nodes {
		switch n.kind {
		case "":
			if err := s.execLocked(n.text, out); err != nil {
				return at(n, err)
			}
		case "if":
			ok, err := cond(n)
			if err != nil {
				return err
			}
			body := n.orElse
			if ok {
				body = n.body
			}
			if err := s.runScript(body, name, out); err != nil {
				return err
			}
		case "while", "repeat":
			count := -1
			if n.kind == "repeat" {
				v, err := s.Eval(n.text)
				if err != nil || v.IsStr {
					return at(n, firstErr(err, fmt.Errorf("repeat precisa de um número")))
				}
				count = int(v.N)
			}
			for i := 0; ; i++ {
				if i >= maxLoop {
					return at(n, fmt.Errorf("laço com mais de %d voltas", maxLoop))
				}
				if count >= 0 {
					if i >= count {
						break
					}
				} else if ok, err := cond(n); err != nil {
					return err
				} else if !ok {
					break
				}
				err := s.runScript(n.body, name, out)
				if errors.Is(err, errBreak) {
					break
				}
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// lineError is an error with the script position in front.
type lineError struct {
	where string
	err   error
}

func (e *lineError) Error() string { return e.where + ": " + e.err.Error() }
func (e *lineError) Unwrap() error { return e.err }

func cmdScript(s *Session, a []string, out io.Writer) error {
	if len(a) != 1 {
		return usageErr("script <arquivo.g81>")
	}
	f, err := os.Open(a[0])
	if err != nil {
		return err
	}
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	f.Close()
	if err := sc.Err(); err != nil {
		return err
	}
	nodes, err := parseScript(lines, a[0])
	if err != nil {
		return err
	}
	err = s.runScript(nodes, a[0], out)
	if errors.Is(err, errBreak) {
		return fmt.Errorf("%s: %v", a[0], errBreak)
	}
	return err
}
