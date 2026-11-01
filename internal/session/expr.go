// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Expressions of the shell (calc, print, if, while, assert, waitfor) and of
// ZRCP's evaluate. Values are integers or strings.
//
//	numbers   16384  $4000  0x4000  'A'
//	strings   "texto"
//	names     a variable (set/calc), e.g.  i + 1
//	functions peek(a) dpeek(a) reg("HL") pc() sp() frames() screen("texto")
//	operators || && == != < <= > >= + - * / % ! and unary -

// Value is the result of an expression.
type Value struct {
	N     int64
	S     string
	IsStr bool
}

func (v Value) String() string {
	if v.IsStr {
		return v.S
	}
	return strconv.FormatInt(v.N, 10)
}

// True reports whether the value counts as true (not 0, not "").
func (v Value) True() bool {
	if v.IsStr {
		return v.S != ""
	}
	return v.N != 0
}

type tokKind int

const (
	tkEOF tokKind = iota
	tkNum
	tkStr
	tkName
	tkOp
)

type token struct {
	kind tokKind
	text string
	num  int64
}

func lexExpr(s string) ([]token, error) {
	var toks []token
	r := []rune(s)
	for i := 0; i < len(r); {
		c := r[i]
		switch {
		case unicode.IsSpace(c):
			i++
		case c == '"':
			var sb strings.Builder
			i++
			for i < len(r) && r[i] != '"' {
				if r[i] == '\\' && i+1 < len(r) && r[i+1] == '"' {
					i++
				}
				sb.WriteRune(r[i])
				i++
			}
			if i >= len(r) {
				return nil, fmt.Errorf("aspas não fechadas")
			}
			i++
			toks = append(toks, token{kind: tkStr, text: sb.String()})
		case c == '\'' && i+2 < len(r) && r[i+2] == '\'':
			toks = append(toks, token{kind: tkNum, num: int64(r[i+1])})
			i += 3
		case c >= '0' && c <= '9' || c == '$' && i+1 < len(r) && isHexRune(r[i+1]):
			j := i + 1
			for j < len(r) && (unicode.IsLetter(r[j]) || unicode.IsDigit(r[j])) {
				j++
			}
			n, err := ParseNumber(string(r[i:j]))
			if err != nil {
				return nil, err
			}
			toks = append(toks, token{kind: tkNum, num: int64(n), text: string(r[i:j])})
			i = j
		case unicode.IsLetter(c) || c == '_':
			j := i + 1
			for j < len(r) && (unicode.IsLetter(r[j]) || unicode.IsDigit(r[j]) || r[j] == '_') {
				j++
			}
			toks = append(toks, token{kind: tkName, text: string(r[i:j])})
			i = j
		default:
			two := ""
			if i+1 < len(r) {
				two = string(r[i : i+2])
			}
			switch two {
			case "||", "&&", "==", "!=", "<=", ">=":
				toks = append(toks, token{kind: tkOp, text: two})
				i += 2
				continue
			}
			if strings.ContainsRune("+-*/%()<>!,", c) {
				toks = append(toks, token{kind: tkOp, text: string(c)})
				i++
				continue
			}
			return nil, fmt.Errorf("caractere inesperado na expressão: %q", c)
		}
	}
	return append(toks, token{kind: tkEOF}), nil
}

func isHexRune(r rune) bool {
	return r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
}

// exprParser evaluates while parsing (precedence climbing).
type exprParser struct {
	toks []token
	pos  int
	s    *Session // caller holds s.mu
}

func (p *exprParser) peek() token { return p.toks[p.pos] }
func (p *exprParser) next() token { t := p.toks[p.pos]; p.pos++; return t }

func (p *exprParser) accept(op string) bool {
	if t := p.peek(); t.kind == tkOp && t.text == op {
		p.pos++
		return true
	}
	return false
}

func (p *exprParser) expect(op string) error {
	if !p.accept(op) {
		return fmt.Errorf("esperava %q na expressão", op)
	}
	return nil
}

var precedence = map[string]int{
	"||": 1, "&&": 2, "==": 3, "!=": 3, "<": 4, "<=": 4, ">": 4, ">=": 4,
	"+": 5, "-": 5, "*": 6, "/": 6, "%": 6,
}

func (p *exprParser) binary(min int) (Value, error) {
	left, err := p.unary()
	if err != nil {
		return left, err
	}
	for {
		t := p.peek()
		prec, ok := precedence[t.text]
		if t.kind != tkOp || !ok || prec < min {
			return left, nil
		}
		p.pos++
		right, err := p.binary(prec + 1)
		if err != nil {
			return left, err
		}
		if left, err = apply(t.text, left, right); err != nil {
			return left, err
		}
	}
}

func boolVal(b bool) Value {
	if b {
		return Value{N: 1}
	}
	return Value{}
}

func apply(op string, a, b Value) (Value, error) {
	switch op {
	case "&&":
		return boolVal(a.True() && b.True()), nil
	case "||":
		return boolVal(a.True() || b.True()), nil
	}
	if a.IsStr || b.IsStr {
		if op == "+" {
			return Value{S: a.String() + b.String(), IsStr: true}, nil
		}
		if !(a.IsStr && b.IsStr) {
			return Value{}, fmt.Errorf("não dá para usar %s entre número e texto (%s, %s)", op, a, b)
		}
		switch op {
		case "==":
			return boolVal(a.S == b.S), nil
		case "!=":
			return boolVal(a.S != b.S), nil
		case "<":
			return boolVal(a.S < b.S), nil
		case "<=":
			return boolVal(a.S <= b.S), nil
		case ">":
			return boolVal(a.S > b.S), nil
		case ">=":
			return boolVal(a.S >= b.S), nil
		}
		return Value{}, fmt.Errorf("operador %s não vale para textos", op)
	}
	x, y := a.N, b.N
	switch op {
	case "+":
		return Value{N: x + y}, nil
	case "-":
		return Value{N: x - y}, nil
	case "*":
		return Value{N: x * y}, nil
	case "/", "%":
		if y == 0 {
			return Value{}, fmt.Errorf("divisão por zero")
		}
		if op == "/" {
			return Value{N: x / y}, nil
		}
		return Value{N: x % y}, nil
	case "==":
		return boolVal(x == y), nil
	case "!=":
		return boolVal(x != y), nil
	case "<":
		return boolVal(x < y), nil
	case "<=":
		return boolVal(x <= y), nil
	case ">":
		return boolVal(x > y), nil
	case ">=":
		return boolVal(x >= y), nil
	}
	return Value{}, fmt.Errorf("operador desconhecido: %s", op)
}

func (p *exprParser) unary() (Value, error) {
	if p.accept("-") {
		v, err := p.unary()
		if err != nil || v.IsStr {
			return v, firstErr(err, fmt.Errorf("- antes de um texto"))
		}
		return Value{N: -v.N}, nil
	}
	if p.accept("!") {
		v, err := p.unary()
		return boolVal(!v.True()), err
	}
	return p.primary()
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func (p *exprParser) primary() (Value, error) {
	t := p.next()
	switch t.kind {
	case tkNum:
		return Value{N: t.num}, nil
	case tkStr:
		return Value{S: t.text, IsStr: true}, nil
	case tkOp:
		if t.text == "(" {
			v, err := p.binary(1)
			if err != nil {
				return v, err
			}
			return v, p.expect(")")
		}
	case tkName:
		if p.accept("(") {
			var args []Value
			if !p.accept(")") {
				for {
					v, err := p.binary(1)
					if err != nil {
						return v, err
					}
					args = append(args, v)
					if p.accept(")") {
						break
					}
					if err := p.expect(","); err != nil {
						return v, err
					}
				}
			}
			return p.s.callFunc(t.text, args)
		}
		v, ok := p.s.vars[t.text]
		if !ok {
			return Value{}, fmt.Errorf("variável não definida: %s", t.text)
		}
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return Value{N: n}, nil
		}
		return Value{S: v, IsStr: true}, nil
	case tkEOF:
		return Value{}, fmt.Errorf("a expressão terminou antes da hora")
	}
	return Value{}, fmt.Errorf("inesperado na expressão: %q", t.text)
}

// EvalExpr evaluates an expression (front-ends, ZRCP evaluate).
func (s *Session) EvalExpr(expr string) (Value, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Eval(expr)
}

// Eval evaluates an expression. Caller holds s.mu.
func (s *Session) Eval(expr string) (Value, error) {
	toks, err := lexExpr(expr)
	if err != nil {
		return Value{}, err
	}
	p := &exprParser{toks: toks, s: s}
	v, err := p.binary(1)
	if err != nil {
		return v, err
	}
	if t := p.peek(); t.kind != tkEOF {
		return v, fmt.Errorf("sobrou na expressão: %q", t.text)
	}
	return v, nil
}

// callFunc runs a function of the expressions. Caller holds s.mu.
func (s *Session) callFunc(name string, args []Value) (Value, error) {
	num := func(i int) (int64, error) {
		if i >= len(args) || args[i].IsStr {
			return 0, fmt.Errorf("%s: falta um número", name)
		}
		return args[i].N, nil
	}
	str := func(i int) (string, error) {
		if i >= len(args) || !args[i].IsStr {
			return "", fmt.Errorf("%s: falta um texto entre aspas", name)
		}
		return args[i].S, nil
	}
	c := s.m.CPU
	switch strings.ToLower(name) {
	case "peek":
		a, err := num(0)
		return Value{N: int64(s.m.Peek(uint16(a)))}, err
	case "dpeek":
		a, err := num(0)
		return Value{N: int64(s.m.Peek(uint16(a))) | int64(s.m.Peek(uint16(a+1)))<<8}, err
	case "pc":
		return Value{N: int64(c.PC)}, nil
	case "sp":
		return Value{N: int64(c.SP)}, nil
	case "frames":
		return Value{N: s.frames}, nil
	case "reg":
		r, err := str(0)
		if err != nil {
			return Value{}, err
		}
		v, ok := regValue(c, strings.ToUpper(r))
		if !ok {
			return Value{}, fmt.Errorf("registrador desconhecido: %s", r)
		}
		return Value{N: int64(v)}, nil
	case "screen":
		t, err := str(0)
		if err != nil {
			return Value{}, err
		}
		return boolVal(strings.Contains(strings.Join(s.m.ScreenText(), "\n"), strings.ToUpper(t))), nil
	case "len":
		t, err := str(0)
		return Value{N: int64(len([]rune(t)))}, err
	}
	return Value{}, fmt.Errorf("função desconhecida: %s (peek, dpeek, reg, pc, sp, frames, screen, len)", name)
}
