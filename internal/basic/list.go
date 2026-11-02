// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// Addresses of the program area.
const (
	ProgStart  = 0x407D // first program line
	SysVarBase = 0x4009 // first byte of a .p file (VERSN)
	sysDFile   = 0x400C // D_FILE: the display file follows the program
)

// Line is one program line.
type Line struct {
	Num  int
	Body []byte // without the final NEWLINE
}

// ProgramFromP returns the program area of a .p file.
func ProgramFromP(p []byte) ([]byte, error) {
	start := ProgStart - SysVarBase
	if len(p) < start {
		return nil, errors.New("arquivo .p curto demais")
	}
	dfile := int(p[sysDFile-SysVarBase]) | int(p[sysDFile-SysVarBase+1])<<8
	end := dfile - SysVarBase
	if end < start || end > len(p) {
		return nil, fmt.Errorf("D_FILE inválido no .p: %04X", dfile)
	}
	return p[start:end], nil
}

// ProgramFromMemory returns the program area of the running ZX81.
func ProgramFromMemory(m *zx81.Machine) []byte {
	dfile := int(m.Peek(sysDFile)) | int(m.Peek(sysDFile+1))<<8
	if dfile < ProgStart || dfile > 0xFFFF {
		return nil
	}
	b := make([]byte, dfile-ProgStart)
	for i := range b {
		b[i] = m.Peek(uint16(ProgStart + i))
	}
	return b
}

// Lines splits a program area into lines.
func Lines(prog []byte) ([]Line, error) {
	var lines []Line
	for i := 0; i < len(prog); {
		if prog[i] >= 0x40 { // the line number's high byte is < 64
			break
		}
		if i+4 > len(prog) {
			return lines, errors.New("linha incompleta no fim do programa")
		}
		num := int(prog[i])<<8 | int(prog[i+1])
		n := int(prog[i+2]) | int(prog[i+3])<<8
		if n < 1 || i+4+n > len(prog) {
			return lines, fmt.Errorf("linha %d: tamanho inválido (%d)", num, n)
		}
		body := prog[i+4 : i+4+n]
		if body[n-1] == Newline {
			body = body[:n-1]
		}
		lines = append(lines, Line{Num: num, Body: body})
		i += 4 + n
	}
	return lines, nil
}

// itemKind is what the lister found in a line.
type itemKind int

const (
	itemChar  itemKind = iota // a character code 0-63 or 128-191
	itemToken                 // a keyword
	itemSpace                 // a space the ROM adds around keywords
	itemRaw                   // a byte that is not text (the ROM prints "?")
)

type item struct {
	kind   itemKind
	code   byte
	hidden bool // not shown by the ROM (the 5 bytes of a number)
	inStr  bool // inside a string
	inREM  bool
}

// walk turns a line body into items, adding the spaces the ROM's listing
// adds around keywords (ROM routine TOKENS, 094B):
//   - a space before keywords from ** (D8) on that start with a letter
//     (TOKEN-ADD ends with ADD A,E4 on the first character), unless the
//     last thing printed was the space after a keyword;
//   - a space after keywords that end in a letter, digit or $.
func walk(body []byte, raw map[int]bool) []item {
	var out []item
	noLead := true // the ROM sets this after the line number
	inStr, inREM := false, false
	for i := 0; i < len(body); i++ {
		c := body[i]
		if raw[i] && c != NumMarker {
			// Written as "\{XX}" so the loader gives back exactly this byte.
			out = append(out, item{kind: itemRaw, code: c, inStr: inStr, inREM: inREM})
			noLead = false
			if c == Quote && !inREM {
				inStr = !inStr
			}
			if c == TokREM && !inStr {
				inREM = true
			}
			continue
		}
		switch {
		case c == NumMarker && !inREM:
			// The number is written in text before the marker. If the five
			// bytes are what the ROM makes of those digits, they are not
			// written; otherwise they are kept as raw bytes.
			var v [5]byte
			copy(v[:], body[i+1:min(i+6, len(body))])
			if raw[i] || i+5 >= len(body) || !matchesDigits(out, v) {
				for j := 0; j <= 5 && i+j < len(body); j++ {
					out = append(out, item{kind: itemRaw, code: body[i+j], hidden: true})
				}
			}
			i += 5
			continue
		case c == NumMarker: // in a REM: keep the bytes, the ROM hides them
			out = append(out, item{kind: itemRaw, code: c, hidden: true, inREM: true})
			for j := 1; j <= 5 && i+j < len(body); j++ {
				out = append(out, item{kind: itemRaw, code: body[i+j], hidden: true, inREM: true})
			}
			i += 5
			continue
		case c < 0x40 || c >= 0x80 && c < 0xC0:
			out = append(out, item{kind: itemChar, code: c, inStr: inStr, inREM: inREM})
			if c != 0 {
				noLead = false
			}
			if c == Quote && !inREM {
				inStr = !inStr
			}
		case Tokens[c] != "":
			kw := Tokens[c]
			if first := kw[0]; c >= 0xD8 && first >= 'A' && first <= 'Z' && !noLead {
				out = append(out, item{kind: itemSpace, inStr: inStr, inREM: inREM})
			}
			out = append(out, item{kind: itemToken, code: c, inStr: inStr, inREM: inREM})
			noLead = false
			last := kw[len(kw)-1]
			if c >= 0x80 && (last == '$' || last >= '0' && last <= '9' || last >= 'A' && last <= 'Z') {
				out = append(out, item{kind: itemSpace, inStr: inStr, inREM: inREM})
				noLead = true
			}
			if c == TokREM {
				inREM = true
			}
		default: // no keyword: the ROM prints "?"
			out = append(out, item{kind: itemRaw, code: c, inStr: inStr, inREM: inREM})
			noLead = false
		}
	}
	return out
}

// Screen returns a line as the ZX81 shows it after LIST (without the line
// number), using zx81.TextChars; inverse video is not marked.
func Screen(body []byte) string {
	var sb strings.Builder
	for _, it := range walk(body, nil) {
		switch {
		case it.hidden:
		case it.kind == itemSpace:
			sb.WriteByte(' ')
		case it.kind == itemToken:
			sb.WriteString(Tokens[it.code])
		case it.kind == itemRaw:
			sb.WriteByte('?')
		default:
			sb.WriteRune(zx81.TextChars[it.code&63])
		}
	}
	return sb.String()
}

// Text returns a line body in the text format (see the package comment).
func Text(body []byte) string {
	// Render, read the text back with the loader and, where it would give
	// other bytes (a variable called LN, a keyword glued to a letter, digits
	// without a hidden value...), write those bytes as "\{XX}" and try again.
	raw := map[int]bool{}
	for range len(body) + 8 {
		t := render(walk(body, raw))
		back, err := tokenize(t)
		if err != nil {
			break
		}
		k := 0
		for k < len(body) && k < len(back) && body[k] == back[k] {
			k++
		}
		if k == len(body) && len(back) == len(body) {
			return t
		}
		if k >= len(body) {
			k = len(body) - 1
		}
		// Escape the first different byte; if it already is, the cause is
		// earlier (digits that became a number): escape the byte before.
		for k > 0 && raw[k] {
			k--
		}
		raw[k] = true
	}
	for i := range body {
		raw[i] = true
	}
	return render(walk(body, raw))
}

// render writes the items in the text format.
func render(items []item) string {
	type piece struct {
		s         string
		romSpace  bool // a space the ROM adds when listing (not in the line)
		realSpace bool // a stored space written as " "
	}
	var ps []piece
	for _, it := range items {
		switch {
		case it.kind == itemSpace && (it.inREM || it.inStr):
			// The ROM shows spaces around keywords in a REM or a string,
			// but they are not in the line.
		case it.kind == itemSpace:
			ps = append(ps, piece{s: " ", romSpace: true})
		case it.kind == itemRaw, it.kind == itemToken && (it.inREM || it.inStr && it.code != TokQuote):
			// In a REM or a string, keywords are data: "\{XX}" keeps them
			// exact (the ROM adds spaces around them when listing).
			ps = append(ps, piece{s: fmt.Sprintf(`\{%02X}`, it.code)})
		case it.kind == itemToken && it.code == TokQuote:
			ps = append(ps, piece{s: `\"`})
		case it.kind == itemToken:
			ps = append(ps, piece{s: Tokens[it.code]})
		case it.code == 0 && (it.inStr || it.inREM):
			ps = append(ps, piece{s: " ", realSpace: true}) // kept as is
		default:
			ps = append(ps, piece{s: charText(it.code)})
		}
	}
	// Drop the ROM's spaces at the end; stored spaces at the end become
	// "\  " so editors that trim lines cannot lose them.
	for len(ps) > 0 && ps[len(ps)-1].romSpace {
		ps = ps[:len(ps)-1]
	}
	for i := len(ps) - 1; i >= 0 && ps[i].realSpace; i-- {
		ps[i].s = `\  `
	}
	var sb strings.Builder
	for _, p := range ps {
		sb.WriteString(p.s)
	}
	return sb.String()
}

// numberCodes are the characters a number can have: digits, ".", "E", "+", "-".
func numberCode(c byte) bool {
	return c >= 0x1C && c <= 0x25 || c == 0x1B || c == 0x2A || c == 0x15 || c == 0x16
}

var digitsCache sync.Map // string of codes -> [5]byte (or error)

// matchesDigits reports whether v is what the ROM makes of the number
// written just before (the last items).
func matchesDigits(items []item, v [5]byte) bool {
	var codes []byte
	for k := len(items) - 1; k >= 0; k-- {
		it := items[k]
		if it.kind != itemChar || it.inStr || it.inREM || !numberCode(it.code) {
			break
		}
		codes = append([]byte{it.code}, codes...)
	}
	// The number starts at a digit or "."; try the longest text first
	// ("2-1": the number is "1").
	for start := range codes {
		if c := codes[start]; !(c >= 0x1C && c <= 0x25 || c == 0x1B) {
			continue
		}
		key := string(codes[start:])
		var got [5]byte
		if cached, ok := digitsCache.Load(key); ok {
			r, ok := cached.([5]byte)
			if !ok {
				continue
			}
			got = r
		} else {
			b, n, err := ROMFloat(codes[start:])
			if err != nil || n != len(codes)-start {
				digitsCache.Store(key, false)
				continue
			}
			digitsCache.Store(key, b)
			got = b
		}
		return got == v
	}
	return false
}

// charText is the text of a character code 0-63 or 128-191. A space (0)
// outside strings is "\  ": plain spaces there only separate words.
func charText(c byte) string {
	switch {
	case c <= 10:
		return `\` + graphics[c]
	case c >= 128 && c <= 138:
		return `\` + graphics[c-128+11]
	case c >= 128:
		return "%" + charASCII[c&63]
	}
	return charASCII[c]
}

// List returns the whole program as text, one "number body" line each.
func List(lines []Line) string {
	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(strconv.Itoa(l.Num))
		if t := Text(l.Body); t != "" {
			sb.WriteByte(' ')
			sb.WriteString(t)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ListP lists a .p file.
func ListP(p []byte) (string, error) {
	prog, err := ProgramFromP(p)
	if err != nil {
		return "", err
	}
	lines, err := Lines(prog)
	return List(lines), err
}

// ScreenCodes returns a line body as the ZX81 shows it after LIST (without
// the line number), in ZX81 character codes: keywords spelled out, with the
// spaces the ROM adds; inverse and graphic characters kept.
func ScreenCodes(body []byte) []byte {
	var out []byte
	for _, it := range walk(body, nil) {
		switch {
		case it.hidden:
		case it.kind == itemSpace:
			out = append(out, 0)
		case it.kind == itemToken:
			out = append(out, TextCodes(Tokens[it.code])...)
		case it.code < 64 || it.code >= 128 && it.code < 192:
			out = append(out, it.code)
		default:
			out = append(out, 0x0F) // "?": a code the ROM would not print as a character
		}
	}
	return out
}

// TextCodes turns text into ZX81 character codes (letters in either case;
// a character the ZX81 does not have becomes "?").
func TextCodes(s string) []byte {
	out := make([]byte, 0, len(s))
	for _, r := range strings.ToUpper(s) {
		c, ok := codeOfRune(r)
		if !ok {
			c = 0x0F
		}
		out = append(out, c)
	}
	return out
}

// codeOfRune is the ZX81 code of a character of the text (£ also as #).
func codeOfRune(r rune) (byte, bool) {
	if r == '#' {
		r = '£'
	}
	if r == ' ' {
		return 0, true
	}
	for c, x := range zx81.TextChars {
		if c > 10 && x == r {
			return byte(c), true
		}
	}
	return 0, false
}

// CodesText turns a line as typed on the ZX81 (character codes and
// keyword tokens, without the hidden numbers) into the text format, so
// that ParseLine can read it: keywords spelled out with spaces around,
// graphics and inverse characters as escapes, "" inside a string as \".
func CodesText(codes []byte) string {
	var sb strings.Builder
	space := func() { // between words only
		s := sb.String()
		if s == "" {
			return
		}
		last := s[len(s)-1]
		if last >= 'A' && last <= 'Z' || last >= '0' && last <= '9' || last == '$' || last == '.' {
			sb.WriteByte(' ')
		}
	}
	inStr := false
	for _, c := range codes {
		switch {
		case c == Quote:
			inStr = !inStr
			sb.WriteString(`"`)
		case c == TokQuote:
			sb.WriteString(`\"`)
		case Tokens[c] != "" && !inStr:
			space()
			sb.WriteString(Tokens[c])
			sb.WriteByte(' ')
		case Tokens[c] != "":
			sb.WriteString(fmt.Sprintf(`\{%02X}`, c))
		case c == 0 && !inStr:
			space()
		case c < 64 || c >= 128 && c < 192:
			sb.WriteString(charText(c))
		}
	}
	return strings.TrimSpace(sb.String())
}

// TypedCodes is a line as it is typed on the ZX81: the number in digits,
// then the body without the hidden numbers (7Eh and five bytes) the ROM
// adds after each number.
func TypedCodes(l Line) []byte {
	out := TextCodes(strconv.Itoa(l.Num))
	inREM := false
	for i := 0; i < len(l.Body); i++ {
		c := l.Body[i]
		if i == 0 && c == TokREM {
			inREM = true
		}
		if c == NumMarker && !inREM && i+5 < len(l.Body) {
			i += 5
			continue
		}
		out = append(out, c)
	}
	return out
}
