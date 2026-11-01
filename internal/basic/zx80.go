// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// ZX80 BASIC (4K ROM). The ZX80 has its own character set and keyword
// codes ($D4-$FF), keeps numbers only as digits (integer BASIC, no hidden
// floating point value) and ends each line with NEWLINE ($76) without a
// length. Tables from EightyOne (zx80BasicLister.cpp), checked against the
// token table of the ZX80 ROM.

// ZX80Tokens are the ZX80 keywords, as listed. The ROM has no keyword for
// $F1, $F2, $F5 and $FF (the ZXpand uses some of them).
var ZX80Tokens = map[byte]string{
	0xD4: `"`, 0xD5: "THEN", 0xD6: "TO", 0xD7: ";", 0xD8: ",", 0xD9: ")", 0xDA: "(", 0xDB: "NOT",
	0xDC: "-", 0xDD: "+", 0xDE: "*", 0xDF: "/", 0xE0: "AND", 0xE1: "OR", 0xE2: "**", 0xE3: "=",
	0xE4: ">", 0xE5: "<", 0xE6: "LIST", 0xE7: "RETURN", 0xE8: "CLS", 0xE9: "DIM", 0xEA: "SAVE", 0xEB: "FOR",
	0xEC: "GO TO", 0xED: "POKE", 0xEE: "INPUT", 0xEF: "RANDOMISE", 0xF0: "LET", 0xF3: "NEXT", 0xF4: "PRINT",
	0xF6: "NEW", 0xF7: "RUN", 0xF8: "STOP", 0xF9: "CONTINUE", 0xFA: "IF", 0xFB: "GO SUB", 0xFC: "LOAD",
	0xFD: "CLEAR", 0xFE: "REM",
}

// zx80Spaced tells where the ZX80 listing puts spaces around a keyword:
// 'b' both sides, 'a' after only; the others (signs) have none.
var zx80Spaced = map[byte]byte{
	0xD5: 'b', 0xD6: 'b', 0xDB: 'a', 0xE0: 'b', 0xE1: 'b',
}

// zx80Chars is the text of the ZX80 character codes 0-63 (the pound sign
// is "#", as in the ZX81 text format).
var zx80Chars = [64]string{
	" ", `"`, `\: `, `\..`, `\' `, `\ '`, `\. `, `\ .`, `\.'`, `\##`, `\,,`, `\~~`, "#", "$", ":", "?",
	"(", ")", "-", "+", "*", "/", "=", ">", "<", ";", ",", ".", "0", "1", "2", "3",
	"4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J",
	"K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

// zx80Inverse are the inverse block graphics, codes $82-$8B.
var zx80Inverse = [10]string{`\ :`, `\''`, `\.:`, `\:.`, `\':`, `\:'`, `\'.`, `\@@`, `\;;`, `\!!`}

// ZX80 memory layout.
const (
	ZX80SysVars   = 0x4000 // first byte of a .o file
	ZX80ProgStart = 0x4028
	zx80VARS      = 0x4008 // the variables follow the program
)

// IsZX80File reports whether a file name is a ZX80 program (.o or .80).
func IsZX80File(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".o", ".80":
		return true
	}
	return false
}

// ZX80Lines splits the program of a ZX80 .o file into lines.
func ZX80Lines(o []byte) ([]Line, error) {
	start := ZX80ProgStart - ZX80SysVars
	if len(o) < start {
		return nil, errors.New("arquivo .o curto demais")
	}
	vars := int(o[zx80VARS-ZX80SysVars]) | int(o[zx80VARS-ZX80SysVars+1])<<8
	end := vars - ZX80SysVars
	if end < start || end > len(o) {
		return nil, fmt.Errorf("VARS inválido no .o: %04X", vars)
	}
	prog := o[start:end]
	var lines []Line
	for i := 0; i < len(prog); {
		if prog[i] >= 0x40 { // end of the program
			break
		}
		if i+2 >= len(prog) {
			return lines, errors.New("linha incompleta no fim do programa")
		}
		num := int(prog[i])<<8 | int(prog[i+1])
		j := i + 2
		for j < len(prog) && prog[j] != Newline {
			j++
		}
		if j == len(prog) {
			return lines, fmt.Errorf("linha %d sem NEWLINE no fim", num)
		}
		lines = append(lines, Line{Num: num, Body: prog[i+2 : j]})
		i = j + 1
	}
	return lines, nil
}

// zx80CharText is the text of one code that is not a keyword.
func zx80CharText(c byte) string {
	switch {
	case c < 64:
		return zx80Chars[c]
	case c == 0x80:
		return "% "
	case c == 0x81:
		return `%"`
	case c >= 0x82 && c <= 0x8B:
		return zx80Inverse[c-0x82]
	case c >= 0x8C && c < 0xC0:
		return "%" + zx80Chars[c&63]
	}
	return fmt.Sprintf(`\{%02X}`, c)
}

// ZX80Text returns a ZX80 line body as text, with the spaces the ZX80 puts
// around keywords.
func ZX80Text(body []byte) string {
	var sb strings.Builder
	space := true    // a space was just written (or the line starts)
	kwSpace := false // ...and it came after a keyword
	inREM := false
	put := func(s string) {
		if s == "" {
			return
		}
		sb.WriteString(s)
		space = strings.HasSuffix(s, " ")
		kwSpace = false
	}
	for _, c := range body {
		kw, ok := ZX80Tokens[c]
		if !ok || inREM {
			put(zx80CharText(c))
			continue
		}
		switch {
		case c >= 0xE6 || zx80Spaced[c] == 'b': // spaces around it
			if !space {
				put(" ")
			}
			put(kw + " ")
			kwSpace = true
			inREM = c == 0xFE
		case zx80Spaced[c] == 'a':
			put(kw + " ")
			kwSpace = true
		default:
			put(kw)
		}
	}
	t := sb.String()
	if kwSpace {
		t = t[:len(t)-1]
	}
	return t
}

// ListZX80 lists a ZX80 .o file, one "number body" line each.
func ListZX80(o []byte) (string, error) {
	lines, err := ZX80Lines(o)
	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(strconv.Itoa(l.Num))
		if t := ZX80Text(l.Body); t != "" {
			sb.WriteByte(' ')
			sb.WriteString(t)
		}
		sb.WriteByte('\n')
	}
	return sb.String(), err
}
