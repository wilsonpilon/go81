// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newBooted(t *testing.T) *Session {
	t.Helper()
	s, err := New(DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, s, "run 150")
	return s
}

func mustExec(t *testing.T, s *Session, line string) string {
	t.Helper()
	var out bytes.Buffer
	if err := s.Exec(line, &out); err != nil {
		t.Fatalf("%s: %v\n%s", line, err, out.String())
	}
	return out.String()
}

func TestParseNumber(t *testing.T) {
	cases := map[string]int{
		"16384": 16384, "$4000": 0x4000, "0x4000": 0x4000, "4000h": 0x4000,
		"FFh": 255, "1010%": 10, "'A'": 'A', "-5": -5,
	}
	for in, want := range cases {
		got, err := ParseNumber(in)
		if err != nil || got != want {
			t.Errorf("ParseNumber(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "xyz", "12G"} {
		if _, err := ParseNumber(bad); err == nil {
			t.Errorf("ParseNumber(%q) should fail", bad)
		}
	}
}

func TestSplit(t *testing.T) {
	a, err := SplitArgs(`save "MEU JOGO.p" "X \"Y"`)
	if err != nil || len(a) != 3 || a[1] != "MEU JOGO.p" || a[2] != `X "Y` {
		t.Fatalf("SplitArgs = %q, %v", a, err)
	}
	if _, err := SplitArgs(`"aberto`); err == nil {
		t.Error("unclosed quote should fail")
	}
	c := SplitCommands(`run 10; echo "a;b"; regs`)
	if len(c) != 3 || strings.TrimSpace(c[1]) != `echo "a;b"` {
		t.Fatalf("SplitCommands = %q", c)
	}
}

func TestTypeAndScreen(t *testing.T) {
	s := newBooted(t)
	mustExec(t, s, "typeln P2*21") // P = PRINT at the K cursor
	scr := mustExec(t, s, "screen")
	if !strings.Contains(scr, "|42 ") {
		t.Fatalf("expected 42 on screen:\n%s", scr)
	}
	if !strings.Contains(scr, "0/0") {
		t.Fatalf("expected report 0/0:\n%s", scr)
	}
}

func TestPeekPokeRegs(t *testing.T) {
	s := newBooted(t)
	mustExec(t, s, "poke $7000 1 2 255")
	out := mustExec(t, s, "peek 7000h 3")
	if !strings.HasPrefix(out, "7000: 01 02 FF") {
		t.Fatalf("peek = %q", out)
	}
	out = mustExec(t, s, "peek 0")
	if !strings.HasPrefix(out, "0000: D3 (211)") {
		t.Fatalf("peek 0 = %q", out)
	}
	mustExec(t, s, "poke 0 0") // ROM is protected
	if out = mustExec(t, s, "peek 0"); !strings.HasPrefix(out, "0000: D3") {
		t.Fatalf("ROM changed: %q", out)
	}
	if out = mustExec(t, s, "regs"); !strings.Contains(out, "PC=") || !strings.Contains(out, "IY=4000") {
		t.Fatalf("regs = %q", out)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "teste.p")
	s := newBooted(t)
	mustExec(t, s, `typeln 10P"GO81"`)
	out := mustExec(t, s, "save "+p)
	if !strings.Contains(out, `SAVE "TESTE"`) {
		t.Fatalf("save output: %q", out)
	}
	data, err := os.ReadFile(p)
	if err != nil || len(data) < 116 {
		t.Fatalf("saved file: %d bytes, %v", len(data), err)
	}

	s2 := newBooted(t)
	mustExec(t, s2, "load "+p)
	mustExec(t, s2, "typeln R") // RUN
	scr := mustExec(t, s2, "screen")
	if !strings.Contains(scr, "|GO81") {
		t.Fatalf("program did not run:\n%s", scr)
	}
}

func TestScriptAndErrors(t *testing.T) {
	dir := t.TempDir()
	sc := filepath.Join(dir, "a.g81")
	os.WriteFile(sc, []byte("# comentário\nrun 150\npoke 30000 7\n\nnaoexiste\n"), 0o644)
	s, _ := New(DefaultSettings())
	var out bytes.Buffer
	err := s.Exec("script "+sc, &out)
	if err == nil || !strings.Contains(err.Error(), "a.g81:5") {
		t.Fatalf("script error = %v", err)
	}
	if v := mustExec(t, s, "peek 30000"); !strings.HasPrefix(v, "7530: 07") {
		t.Fatalf("script did not run line 3: %q", v)
	}
	if err := s.Exec("quit", &out); !errors.Is(err, ErrQuit) {
		t.Fatalf("quit = %v", err)
	}
	if err := s.Exec("poke 1", &out); err == nil {
		t.Fatal("poke without value should fail")
	}
}

func TestMachine(t *testing.T) {
	s, _ := New(DefaultSettings())
	if out := mustExec(t, s, "machine ram=2 ntsc=on"); !strings.Contains(out, "2K NTSC") {
		t.Fatalf("machine = %q", out)
	}
	mustExec(t, s, "run 150")
	// RAMTOP with 2K: 16384 + 2048
	if out := mustExec(t, s, "peek $4004 2"); !strings.HasPrefix(out, "4004: 00 48") {
		t.Fatalf("RAMTOP = %q", out)
	}
	if err := s.Exec("machine ram=3", &bytes.Buffer{}); err == nil {
		t.Fatal("ram=3 should fail")
	}
}
