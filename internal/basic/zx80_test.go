// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic

import (
	"os"
	"strings"
	"testing"
)

// zx80Code turns plain text into ZX80 character codes (letters, digits,
// space and a few signs only).
func zx80Code(t *testing.T, s string) []byte {
	t.Helper()
	var b []byte
	for _, r := range s {
		found := false
		for c, txt := range zx80Chars {
			if txt == string(r) {
				b = append(b, byte(c))
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("no ZX80 code for %q", r)
		}
	}
	return b
}

// makeO builds a .o file with the lines (body bytes as given).
func makeO(lines []Line) []byte {
	o := make([]byte, ZX80ProgStart-ZX80SysVars)
	for _, l := range lines {
		o = append(o, byte(l.Num>>8), byte(l.Num))
		o = append(o, l.Body...)
		o = append(o, Newline)
	}
	vars := ZX80SysVars + len(o)
	o[zx80VARS-ZX80SysVars], o[zx80VARS-ZX80SysVars+1] = byte(vars), byte(vars>>8)
	return append(o, 0x80) // the end of the variables
}

func TestListZX80(t *testing.T) {
	cat := func(parts ...[]byte) []byte {
		var b []byte
		for _, p := range parts {
			b = append(b, p...)
		}
		return b
	}
	o := makeO([]Line{
		{10, cat([]byte{0xF4, 0x01}, zx80Code(t, "OLA"), []byte{0xD4})},              // PRINT "OLA"
		{20, cat([]byte{0xFA, 0x26, 0xE3, 0x1D, 0xD5, 0xEC}, zx80Code(t, "10"))},     // IF A=1 THEN GO TO 10
		{30, cat([]byte{0xF0, 0x27, 0xE3, 0xDB, 0x26, 0xE0, 0x28})},                  // LET B=NOT A AND C
		{40, cat([]byte{0xFE}, zx80Code(t, "OI  "), []byte{0xF4, 0x82, 0x80, 0x8C})}, // REM with a token, inverse chars
		{50, []byte{0xE7}}, // RETURN
	})
	got, err := ListZX80(o)
	if err != nil {
		t.Fatal(err)
	}
	want := "10 PRINT \"OLA\"\n" +
		"20 IF A=1 THEN GO TO 10\n" +
		"30 LET B=NOT A AND C\n" +
		"40 REM OI  " + `\{F4}\ :% %#` + "\n" +
		"50 RETURN\n"
	if got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if !IsZX80File("JOGO.O") || !IsZX80File("x.80") || IsZX80File("x.p") {
		t.Fatal("IsZX80File")
	}
	if _, err := ListZX80([]byte{1, 2}); err == nil {
		t.Fatal("short file accepted")
	}
}

// TestZX80TokensROM checks the keyword table against the one in the ZX80
// ROM (local reference, not in Git: skipped without it).
func TestZX80TokensROM(t *testing.T) {
	rom, err := os.ReadFile("../../81emul/exe/ROMs/zx80.rom")
	if err != nil {
		t.Skip("zx80.rom não encontrada (81emul/ fica fora do Git)")
	}
	// The table starts with the closing quote ($D4); each keyword ends
	// with bit 7 set; codes without a keyword are a single "?" ($0F).
	text := func(c byte) string {
		c &= 0x3F
		if c == 0 {
			return " "
		}
		return zx80Chars[c]
	}
	i := 0xBC
	for code := 0xD4; code <= 0xFF; code++ {
		var sb strings.Builder
		for {
			c := rom[i]
			i++
			sb.WriteString(text(c))
			if c&0x80 != 0 {
				break
			}
		}
		word := sb.String()
		if word == "?" {
			word = ""
		}
		if ZX80Tokens[byte(code)] != word {
			t.Errorf("$%02X: tabela %q, ROM %q", code, ZX80Tokens[byte(code)], word)
		}
	}
}
