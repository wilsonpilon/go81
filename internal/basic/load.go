// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"bufio"
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// The loader: text (the format of the package comment) to a program.
//
// Outside strings and REM lines, spaces only separate words (a stored space
// is "\  "), keywords are recognised in any case, and every number gets the
// hidden 5-byte value the ZX81 adds, computed by the ROM's own routine. A
// number followed by "\{7E}" and five more "\{XX}" keeps those bytes
// instead (the lister writes this when a program stores a value that is not
// the one of its digits).

// LoadError is an error in a line of the text.
type LoadError struct {
	Line int // line of the text (1 = first)
	Msg  string
}

func (e *LoadError) Error() string { return fmt.Sprintf("linha %d: %s", e.Line, e.Msg) }

// keyword is a word the loader turns into a token.
type keyword struct {
	text  string // upper case
	code  byte
	alpha bool // starts with a letter: must not touch other letters
}

var keywords = func() []keyword {
	var ks []keyword
	add := func(text string, code byte) {
		ks = append(ks, keyword{text, code, text[0] >= 'A' && text[0] <= 'Z'})
	}
	for c, t := range Tokens {
		if c == TokQuote || t == "" {
			continue
		}
		add(t, c)
	}
	// Alternative spellings (as in zxtext2p).
	add("GO TO", 0xEC)
	add("GO SUB", 0xED)
	add("RANDOMIZE", 0xF9)
	add("CONTINUE", 0xE8)
	// Longest first, so GOSUB wins over GO..., <= over <, etc.
	sort.Slice(ks, func(i, j int) bool {
		if len(ks[i].text) != len(ks[j].text) {
			return len(ks[i].text) > len(ks[j].text)
		}
		return ks[i].text < ks[j].text
	})
	return ks
}()

// graphicCode maps a block graphic escape ("\" + 2 characters) to its code.
var graphicCode = func() map[string]byte {
	m := map[string]byte{}
	for i, g := range graphics {
		c := byte(i)
		if i > 10 {
			c = byte(128 + i - 11)
		}
		m[g] = c
	}
	return m
}()

// ParseText reads a whole program text. Blank lines and lines starting with
// "#" are ignored; a line ending in "\" continues on the next one.
func ParseText(text string) ([]Line, error) {
	var lines []Line
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	n, last := 0, -1
	for sc.Scan() {
		n++
		start := n
		l := strings.TrimRight(sc.Text(), "\r")
		for strings.HasSuffix(l, `\`) && !strings.HasSuffix(l, `\\`) && sc.Scan() {
			n++
			l = l[:len(l)-1] + strings.TrimRight(sc.Text(), "\r")
		}
		if strings.TrimSpace(l) == "" || strings.HasPrefix(l, "#") {
			continue
		}
		line, err := ParseLine(l)
		if err != nil {
			return lines, &LoadError{Line: start, Msg: err.Error()}
		}
		if line.Num <= last {
			return lines, &LoadError{Line: start, Msg: fmt.Sprintf("a linha %d não vem depois da %d", line.Num, last)}
		}
		last = line.Num
		lines = append(lines, line)
	}
	return lines, sc.Err()
}

// ParseLine turns "10 PRINT ..." into a program line.
func ParseLine(s string) (Line, error) {
	t := strings.TrimLeft(s, " \t")
	i := 0
	for i < len(t) && t[i] >= '0' && t[i] <= '9' {
		i++
	}
	if i == 0 {
		return Line{}, fmt.Errorf("falta o número da linha")
	}
	num, _ := strconv.Atoi(t[:i])
	// The ZX81 accepts 1-9999 when typing, but programs may hold line 0
	// (and up to 16383) by POKE; the ROM lists and runs them.
	if num > 16383 {
		return Line{}, fmt.Errorf("número de linha fora de 0..16383: %d", num)
	}
	rest := t[i:] // "10PRINT" is fine too
	body, err := tokenize(strings.TrimLeft(rest, " \t"))
	if err != nil {
		return Line{}, err
	}
	return Line{Num: num, Body: body}, nil
}

// lexer walks a line body.
type lexer struct {
	s   []rune
	i   int
	out []byte
}

func (lx *lexer) peek(k int) rune {
	if lx.i+k < len(lx.s) {
		return lx.s[lx.i+k]
	}
	return 0
}

func isLetter(r rune) bool { return r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' }
func isDigit(r rune) bool  { return r >= '0' && r <= '9' }

// escape reads an escape at lx.i ("\..", "\{XX}", "\"", "%c") and returns
// its code.
func (lx *lexer) escape() (byte, error) {
	r := lx.s[lx.i]
	if r == '%' {
		if lx.i+1 >= len(lx.s) {
			return 0, fmt.Errorf("%% sem caractere no fim da linha")
		}
		c, ok := asciiCode(lx.s[lx.i+1])
		if !ok || c > 63 {
			return 0, fmt.Errorf("%%%c: não há esse caractere em vídeo inverso", lx.s[lx.i+1])
		}
		lx.i += 2
		return c | 0x80, nil
	}
	// Backslash.
	if lx.peek(1) == '"' {
		lx.i += 2
		return TokQuote, nil
	}
	if lx.peek(1) == '{' {
		end := lx.i + 2
		for end < len(lx.s) && lx.s[end] != '}' {
			end++
		}
		if end >= len(lx.s) {
			return 0, fmt.Errorf(`\{ sem }`)
		}
		v, err := strconv.ParseUint(string(lx.s[lx.i+2:end]), 16, 8)
		if err != nil {
			return 0, fmt.Errorf(`\{%s}: use dois dígitos hexadecimais`, string(lx.s[lx.i+2:end]))
		}
		lx.i = end + 1
		return byte(v), nil
	}
	if lx.i+2 < len(lx.s) {
		if c, ok := graphicCode[string(lx.s[lx.i+1:lx.i+3])]; ok {
			lx.i += 3
			return c, nil
		}
	}
	return 0, fmt.Errorf(`escape desconhecido: \%s`, string(lx.s[lx.i+1:min(lx.i+3, len(lx.s))]))
}

// char adds one source character (not a keyword or number) at lx.i.
func (lx *lexer) char() error {
	r := lx.s[lx.i]
	if r == '\\' || r == '%' {
		c, err := lx.escape()
		if err != nil {
			return err
		}
		lx.out = append(lx.out, c)
		return nil
	}
	c, ok := asciiCode(r)
	if !ok {
		return fmt.Errorf("o caractere %q não existe no ZX81", r)
	}
	lx.out = append(lx.out, c)
	lx.i++
	return nil
}

// keyword tries to read a keyword at lx.i.
func (lx *lexer) keyword(prevLetter bool) (byte, bool) {
	for _, k := range keywords {
		n := len(k.text)
		if lx.i+n > len(lx.s) {
			continue
		}
		if !strings.EqualFold(string(lx.s[lx.i:lx.i+n]), k.text) {
			continue
		}
		if k.alpha && (prevLetter || isLetter(lx.peek(n)) && !strings.HasSuffix(k.text, "$")) {
			continue
		}
		lx.i += n
		return k.code, true
	}
	return 0, false
}

// number reads a numeric literal at lx.i, adds its characters and the
// hidden value. It reports false if there is no number here.
func (lx *lexer) number() (bool, error) {
	j := lx.i
	for j < len(lx.s) && isDigit(lx.s[j]) {
		j++
	}
	if j < len(lx.s) && lx.s[j] == '.' {
		j++
		for j < len(lx.s) && isDigit(lx.s[j]) {
			j++
		}
	}
	digits := j - lx.i
	if digits == 0 || digits == 1 && lx.s[lx.i] == '.' {
		return false, nil
	}
	// Exponent: E, optional sign, digits.
	if j < len(lx.s) && (lx.s[j] == 'E' || lx.s[j] == 'e') {
		k := j + 1
		if k < len(lx.s) && (lx.s[k] == '+' || lx.s[k] == '-') {
			k++
		}
		if k < len(lx.s) && isDigit(lx.s[k]) {
			for k < len(lx.s) && isDigit(lx.s[k]) {
				k++
			}
			j = k
		}
	}
	text := strings.ToUpper(string(lx.s[lx.i:j]))
	codes := make([]byte, 0, len(text))
	for _, r := range text {
		c, _ := asciiCode(r)
		codes = append(codes, c)
	}
	lx.out = append(lx.out, codes...)
	lx.i = j
	// Explicit value written by the lister: keep it.
	if strings.HasPrefix(string(lx.s[lx.i:]), `\{7E}`) || strings.HasPrefix(string(lx.s[lx.i:]), `\{7e}`) {
		return true, nil
	}
	v, n, err := ROMFloat(codes)
	if err != nil {
		return true, fmt.Errorf("%s: %v", text, err)
	}
	if n != len(codes) {
		return true, fmt.Errorf("número inválido: %s", text)
	}
	lx.out = append(lx.out, NumMarker)
	lx.out = append(lx.out, v[:]...)
	return true, nil
}

// tokenize turns a line body into ZX81 codes.
func tokenize(body string) ([]byte, error) {
	lx := &lexer{s: []rune(body)}
	inStr, inREM := false, false
	prevLetter := false // the last thing added was a letter (part of a name)
	for lx.i < len(lx.s) {
		r := lx.s[lx.i]
		switch {
		case inREM || inStr:
			if inStr && r == '"' {
				inStr = false
			}
			if r == '\\' || r == '%' {
				if err := lx.char(); err != nil {
					return nil, err
				}
				continue
			}
			c, ok := asciiCode(r)
			if !ok {
				return nil, fmt.Errorf("o caractere %q não existe no ZX81", r)
			}
			lx.out = append(lx.out, c)
			lx.i++
			continue
		case r == ' ' || r == '\t':
			lx.i++
			prevLetter = false // a space ends a name: "A THEN", "X TO X"
			continue
		case r == '"':
			inStr = true
			lx.out = append(lx.out, Quote)
			lx.i++
			prevLetter = false
			continue
		}
		if code, ok := lx.keyword(prevLetter); ok {
			lx.out = append(lx.out, code)
			prevLetter = false
			if code == TokREM {
				inREM = true
				if lx.peek(0) == ' ' { // the space after REM is not stored
					lx.i++
				}
			}
			continue
		}
		if !prevLetter && (isDigit(r) || r == '.' && isDigit(lx.peek(1))) {
			ok, err := lx.number()
			if err != nil {
				return nil, err
			}
			if ok {
				continue
			}
		}
		if err := lx.char(); err != nil {
			return nil, err
		}
		prevLetter = isLetter(r) || prevLetter && (isDigit(r) || r == '$')
	}
	return lx.out, nil
}

// Empty programs as the ZX81 ROM (edition 3) saves them: the system
// variables, the display file and the end of the variables. On 16K the
// display file is full (24 lines of 32 spaces); on 1K it is collapsed (24
// NEWLINEs). See TestEmptyTemplates.
//
//go:embed templates/vazio16k.p
var empty16K []byte

//go:embed templates/vazio1k.p
var empty1K []byte

// EmptyP returns a copy of an empty program (collapsed: the 1K one).
func EmptyP(collapsed bool) []byte {
	if collapsed {
		return append([]byte(nil), empty1K...)
	}
	return append([]byte(nil), empty16K...)
}

// POptions control MakeP.
type POptions struct {
	// AutoRun > 0 makes the program run from that line as soon as it is
	// loaded (the line must exist).
	AutoRun int
	// Collapsed uses the display file of a 1K ZX81 (24 empty lines): the
	// .p is smaller and runs on any memory size. On 4K or more the first
	// CLS expands it one space at a time (about two seconds, as on a real
	// ZX81); the default full display avoids that wait.
	Collapsed bool
}

// System variables moved when the program grows (they point past it).
var movedSysVars = []int{0x400C, 0x400E, 0x4010, 0x4012, 0x4014, 0x4016, 0x4018, 0x401A, 0x401C}

const (
	sysEPPC   = 0x400A
	sysNXTLIN = 0x4029
)

// MakeP builds a .p file with the lines.
func MakeP(lines []Line, opt POptions) ([]byte, error) {
	autorun := opt.AutoRun
	emptyP := empty16K
	if opt.Collapsed {
		emptyP = empty1K
	}
	var prog []byte
	runAddr := -1
	for _, l := range lines {
		if l.Num == autorun {
			runAddr = ProgStart + len(prog)
		}
		n := len(l.Body) + 1
		prog = append(prog, byte(l.Num>>8), byte(l.Num), byte(n), byte(n>>8))
		prog = append(prog, l.Body...)
		prog = append(prog, Newline)
	}
	if autorun > 0 && runAddr < 0 {
		return nil, fmt.Errorf("autorun: a linha %d não existe", autorun)
	}
	if ProgStart+len(prog)+len(emptyP) > 0xFFFF {
		return nil, fmt.Errorf("programa grande demais")
	}
	start := ProgStart - SysVarBase
	p := make([]byte, 0, len(emptyP)+len(prog))
	p = append(p, emptyP[:start]...)
	p = append(p, prog...)
	p = append(p, emptyP[start:]...)
	set := func(addr, v int) {
		o := addr - SysVarBase
		p[o], p[o+1] = byte(v), byte(v>>8)
	}
	get := func(addr int) int {
		o := addr - SysVarBase
		return int(p[o]) | int(p[o+1])<<8
	}
	for _, sv := range movedSysVars {
		if v := get(sv); v >= ProgStart {
			set(sv, v+len(prog))
		}
	}
	if len(lines) > 0 {
		set(sysEPPC, lines[0].Num)
	}
	// NXTLIN: where the ROM goes on after LOAD. The display file (the end of
	// the program) means stop; a line address means run from there.
	if runAddr >= 0 {
		set(sysNXTLIN, runAddr)
	} else {
		set(sysNXTLIN, get(sysDFile))
	}
	return p, nil
}

// TextToP turns a program text into a .p file.
func TextToP(text string, opt POptions) ([]byte, error) {
	lines, err := ParseText(text)
	if err != nil {
		return nil, err
	}
	return MakeP(lines, opt)
}
