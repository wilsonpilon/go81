// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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
func walk(body []byte) []item {
	var out []item
	noLead := true // the ROM sets this after the line number
	inStr, inREM := false, false
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == NumMarker && !inREM:
			// The number is written in text before the marker; skip it.
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
	for _, it := range walk(body) {
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
	var sb strings.Builder
	for _, it := range walk(body) {
		switch {
		case it.kind == itemSpace && (it.inREM || it.inStr):
			// The ROM shows spaces around keywords in a REM or a string,
			// but they are not in the line.
		case it.kind == itemSpace:
			sb.WriteByte(' ')
		case it.kind == itemRaw, it.kind == itemToken && (it.inREM || it.inStr && it.code != TokQuote):
			// In a REM or a string, keywords are data: "\{XX}" keeps them
			// exact (the ROM adds spaces around them when listing).
			fmt.Fprintf(&sb, `\{%02X}`, it.code)
		case it.kind == itemToken && it.code == TokQuote:
			sb.WriteString(`\"`)
		case it.kind == itemToken:
			sb.WriteString(Tokens[it.code])
		case it.code == 0 && (it.inStr || it.inREM):
			sb.WriteByte(' ') // a space in a string or REM is kept as is
		default:
			sb.WriteString(charText(it.code))
		}
	}
	return strings.TrimRight(sb.String(), " ")
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
