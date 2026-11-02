// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"fmt"
	"sort"
	"strconv"
)

// Keywords followed by a line number.
const (
	TokGOTO  = 0xEC
	TokGOSUB = 0xED
	TokRUN   = 0xF7
	TokLIST  = 0xF0
	TokLLIST = 0xE2
)

var lineRefTokens = map[byte]bool{TokGOTO: true, TokGOSUB: true, TokRUN: true, TokLIST: true, TokLLIST: true}

// RenumResult says what Renumber did.
type RenumResult struct {
	Lines    []Line
	Moved    int      // lines with a new number
	Refs     int      // GOTO/GOSUB/RUN/LIST/LLIST numbers changed
	Warnings []string // references that could not be changed
}

// Renumber gives the lines from..to (both included) the numbers start,
// start+step... and fixes the line numbers written after GOTO, GOSUB, RUN,
// LIST and LLIST in the whole program. A computed number (GOTO A*10) cannot
// be fixed: it is reported in Warnings. The lines must keep their order.
func Renumber(lines []Line, start, step, from, to int) (RenumResult, error) {
	if start < 1 || start > MaxLine {
		return RenumResult{}, fmt.Errorf("primeira linha fora de 1-%d: %d", MaxLine, start)
	}
	if step < 1 || step > MaxLine {
		return RenumResult{}, fmt.Errorf("passo fora de 1-%d: %d", MaxLine, step)
	}
	newNum := map[int]int{}
	var before, after = 0, MaxLine + 1 // the neighbours of the block
	n := start
	for _, l := range lines {
		switch {
		case l.Num < from:
			before = l.Num
		case l.Num > to:
			if after > MaxLine {
				after = l.Num
			}
		default:
			newNum[l.Num] = n
			n += step
		}
	}
	if len(newNum) == 0 {
		return RenumResult{Lines: lines}, nil
	}
	last := n - step
	if last > MaxLine {
		return RenumResult{}, fmt.Errorf("a última linha ficaria com o número %d (máximo %d): use um passo menor", last, MaxLine)
	}
	if start <= before || last >= after {
		return RenumResult{}, fmt.Errorf("as linhas renumeradas (%d a %d) passariam por cima de outras (%d e %d): escolha outro início ou passo", start, last, before, after)
	}
	// target is where a GOTO n goes: line n, or the first one after it.
	target := func(v int) (int, bool) {
		i := sort.Search(len(lines), func(i int) bool { return lines[i].Num >= v })
		if i == len(lines) {
			return 0, false
		}
		nn, ok := newNum[lines[i].Num]
		return nn, ok
	}
	res := RenumResult{}
	for _, l := range lines {
		num := l.Num
		if nn, ok := newNum[num]; ok {
			if nn != num {
				res.Moved++
			}
			num = nn
		}
		body, refs, warns := fixLineRefs(l.Body, target)
		res.Refs += refs
		for _, w := range warns {
			res.Warnings = append(res.Warnings, fmt.Sprintf("linha %d (antes %d): %s", num, l.Num, w))
		}
		res.Lines = append(res.Lines, Line{Num: num, Body: body})
	}
	return res, nil
}

// fixLineRefs rewrites the constant line numbers after GOTO and the others.
func fixLineRefs(body []byte, target func(int) (int, bool)) ([]byte, int, []string) {
	if len(body) > 0 && body[0] == TokREM {
		return body, 0, nil
	}
	out := make([]byte, 0, len(body))
	refs := 0
	var warns []string
	inString := false
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == Quote:
			inString = !inString
		case inString:
		case c == NumMarker && i+5 < len(body):
			out = append(out, body[i:i+6]...)
			i += 5
			continue
		case lineRefTokens[c]:
			out = append(out, c)
			j := i + 1
			for j < len(body) && body[j] >= 0x1C && body[j] <= 0x25 {
				j++
			}
			name := Tokens[c]
			switch {
			case j == i+1 && j == len(body): // RUN, LIST alone
			case j > i+1 && j+6 == len(body) && body[j] == NumMarker:
				v, _ := strconv.Atoi(digitText(body[i+1 : j]))
				if nn, ok := target(v); ok && nn != v {
					f, err := Float(float64(nn))
					if err == nil {
						out = append(out, digitCodes(nn)...)
						out = append(out, NumMarker)
						out = append(out, f[:]...)
						refs++
						i = len(body) - 1
						continue
					}
				}
			default:
				warns = append(warns, fmt.Sprintf("%s com número calculado não foi mudado", name))
			}
			continue
		}
		out = append(out, c)
	}
	return out, refs, warns
}

func digitText(codes []byte) string {
	s := make([]byte, len(codes))
	for i, c := range codes {
		s[i] = '0' + c - 0x1C
	}
	return string(s)
}

func digitCodes(n int) []byte {
	s := strconv.Itoa(n)
	out := make([]byte, len(s))
	for i := range s {
		out[i] = s[i] - '0' + 0x1C
	}
	return out
}
