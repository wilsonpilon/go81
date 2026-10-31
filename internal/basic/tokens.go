// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package basic reads and writes ZX81 BASIC programs.
//
// The text format is the one of zxtext2p (Chris Cowley), so listings can
// be edited and turned back into programs:
//
//   - keywords are written in full (PRINT, GOTO, ...);
//   - block graphics are "\" + two characters ("\' ", "\::", ...);
//   - inverse video is "%" + the character ("%A");
//   - "\"" is the quote character inside strings;
//   - "#" is the pound sign.
//
// Go81 adds "\{XX}" (hexadecimal) for a byte that is not text, and for a
// keyword inside a REM or a string, so nothing is lost: machine code in a
// REM line, for example.
package basic

// Token codes.
const (
	TokRND    = 0x40
	TokINKEY  = 0x41
	TokPI     = 0x42
	TokQuote  = 0xC0 // the quote character inside a string ("")
	TokUnused = 0xC3
	TokREM    = 0xEA
	Newline   = 0x76
	NumMarker = 0x7E // followed by the 5-byte floating point number
	Quote     = 0x0B
)

// Tokens are the keywords, as the ZX81 prints them. Codes without a keyword
// are "".
var Tokens = map[byte]string{
	0x40: "RND", 0x41: "INKEY$", 0x42: "PI",
	0xC0: `""`, 0xC1: "AT", 0xC2: "TAB", 0xC4: "CODE", 0xC5: "VAL", 0xC6: "LEN", 0xC7: "SIN",
	0xC8: "COS", 0xC9: "TAN", 0xCA: "ASN", 0xCB: "ACS", 0xCC: "ATN", 0xCD: "LN", 0xCE: "EXP", 0xCF: "INT",
	0xD0: "SQR", 0xD1: "SGN", 0xD2: "ABS", 0xD3: "PEEK", 0xD4: "USR", 0xD5: "STR$", 0xD6: "CHR$", 0xD7: "NOT",
	0xD8: "**", 0xD9: "OR", 0xDA: "AND", 0xDB: "<=", 0xDC: ">=", 0xDD: "<>", 0xDE: "THEN", 0xDF: "TO",
	0xE0: "STEP", 0xE1: "LPRINT", 0xE2: "LLIST", 0xE3: "STOP", 0xE4: "SLOW", 0xE5: "FAST", 0xE6: "NEW", 0xE7: "SCROLL",
	0xE8: "CONT", 0xE9: "DIM", 0xEA: "REM", 0xEB: "FOR", 0xEC: "GOTO", 0xED: "GOSUB", 0xEE: "INPUT", 0xEF: "LOAD",
	0xF0: "LIST", 0xF1: "LET", 0xF2: "PAUSE", 0xF3: "NEXT", 0xF4: "POKE", 0xF5: "PRINT", 0xF6: "PLOT", 0xF7: "RUN",
	0xF8: "SAVE", 0xF9: "RAND", 0xFA: "IF", 0xFB: "CLS", 0xFC: "UNPLOT", 0xFD: "CLEAR", 0xFE: "RETURN", 0xFF: "COPY",
}

// charASCII is the text for character codes 0-63 (block graphics 1-10 are
// escapes, see graphics).
var charASCII = [64]string{
	" ", "", "", "", "", "", "", "", "", "", "", `"`, "#", "$", ":", "?",
	"(", ")", ">", "<", "=", "+", "-", "*", "/", ";", ",", ".", "0", "1", "2", "3",
	"4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J",
	"K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

// graphics are the zxtext2p escapes for codes 0-10 and 128-138.
var graphics = [22]string{
	"  ", "' ", " '", "''", ". ", ": ", ".'", ":'", "##", ",,", "~~",
	"::", ".:", ":.", "..", "':", " :", "'.", " .", "@@", ";;", "!!",
}
