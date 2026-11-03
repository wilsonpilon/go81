// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A program with graphics, inverse video, quotes, the pound sign and
// machine code in a REM goes to .p and back to the same text.
func TestTxt2pAndBack(t *testing.T) {
	dir := t.TempDir()
	text := `1 REM \{C9}\{00}
10 PRINT "\##\::\.'%A%B\"#"
20 PRINT AT 1,0;"\@@ FIM"
30 GOTO 10
`
	in := filepath.Join(dir, "jogo.txt")
	if err := os.WriteFile(in, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newBooted(t)
	var before []byte
	before = append(before, s.m.Memory[0x4000:0x8000]...)
	out := mustExec(t, s, `txt2p "`+in+`"`)
	if !strings.Contains(out, "jogo.p (4 linhas") {
		t.Fatalf("txt2p = %q", out)
	}
	out = mustExec(t, s, `p2txt "`+filepath.Join(dir, "jogo.p")+`" "`+filepath.Join(dir, "volta.txt")+`"`)
	if !strings.Contains(out, "aviso:") {
		t.Fatalf("p2txt = %q", out)
	}
	// The text goes back to the very same .p, byte for byte.
	mustExec(t, s, `txt2p "`+filepath.Join(dir, "volta.txt")+`"`)
	p1, _ := os.ReadFile(filepath.Join(dir, "jogo.p"))
	p2, _ := os.ReadFile(filepath.Join(dir, "volta.p"))
	if len(p1) == 0 || !bytes.Equal(p1, p2) {
		back, _ := os.ReadFile(filepath.Join(dir, "volta.txt"))
		t.Fatalf(".p → texto → .p differs; text:\n%s", back)
	}
	// Graphics and inverse video come out as \xx and %x.
	back, _ := os.ReadFile(filepath.Join(dir, "volta.txt"))
	if !strings.Contains(string(back), `10 PRINT "\##\::\.'%A%B\"#"`) {
		t.Fatalf("text:\n%s", back)
	}
	if !bytes.Equal(before, s.m.Memory[0x4000:0x8000]) {
		t.Fatal("txt2p/p2txt touched the ZX81")
	}
	// A bad text says where.
	bad := filepath.Join(dir, "ruim.txt")
	os.WriteFile(bad, []byte("10 PRINT \"@\"\n"), 0o644)
	if err := s.Exec(`txt2p "`+bad+`"`, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "'@'") {
		t.Fatalf("bad text: %v", err)
	}
}

// A numbered .AZX of Basic Dignified: small letters, CRLF, "" for the
// quote inside a string, % for inverse video. txt2p reads it; p2txt says
// it is already text.
func TestAZX(t *testing.T) {
	dir := t.TempDir()
	azx := filepath.Join(dir, "coelho.AZX")
	text := "10 rem Converted with Basic Dignified\r\n" +
		"60 print at 4,10;\"%O% %C\"\r\n" +
		"130 print at 16,4;\"NESSA \"\"CHAPA\"\", (\";x$;\")\"\r\n" +
		"140 let a$=\"\"\r\n" +
		"150 print \"\"\"\"\r\n"
	if err := os.WriteFile(azx, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newBooted(t)
	mustExec(t, s, `txt2p "`+azx+`"`)
	got, _, err := ProgramText(filepath.Join(dir, "coelho.p"))
	if err != nil {
		t.Fatal(err)
	}
	want := `10 REM CONVERTED WITH BASIC DIGNIFIED
60 PRINT AT 4,10;"%O\::%C"
130 PRINT AT 16,4;"NESSA \"CHAPA\", (";X$;")"
140 LET A$=""
150 PRINT "\""
`
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	err = s.Exec(`p2txt "`+azx+`"`, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "use txt2p") {
		t.Fatalf("p2txt on text: %v", err)
	}
	if !IsBasicText("jogo.azx") {
		t.Fatal(".azx is BASIC text")
	}
}
