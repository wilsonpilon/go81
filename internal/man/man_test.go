// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package man

import (
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/basic"
)

func page(t *testing.T, args ...string) string {
	t.Helper()
	out, err := Lookup(args, 78)
	if err != nil {
		t.Fatalf("man %v: %v", args, err)
	}
	return out
}

func TestKeysFromROM(t *testing.T) {
	for args, want := range map[string]string{
		"basic PRINT":   "tecla P com o cursor K",
		"basic PEEK":    "FUNCTION (SHIFT+ENTER, cursor F) e depois O",
		"basic USR":     "FUNCTION (SHIFT+ENTER, cursor F) e depois L",
		"basic STOP":    "SHIFT+A",
		"basic INKEY$":  "FUNCTION (SHIFT+ENTER, cursor F) e depois B",
		"basic THEN":    "SHIFT+3",
		"basic GOTO":    "tecla G com o cursor K",
		"basic80 PRINT": "tecla O com o cursor K",
		"basic80 LOAD":  "tecla W com o cursor K",
		"basic80 SAVE":  "tecla E com o cursor K",
	} {
		f := strings.Fields(args)
		r := Find(f[0], f[1])
		if len(r) == 0 || !strings.Contains(r[0].Page, "TECLA: "+want) {
			t.Errorf("man %s: no %q", args, want)
		}
	}
}

func TestEveryKeywordHasAPage(t *testing.T) {
	skip := map[string]bool{`""`: true, "**": true, "<=": true, ">=": true, "<>": true}
	for _, kw := range basic.Tokens {
		if kw == "" || skip[kw] {
			continue
		}
		if len(Find("basic", kw)) == 0 {
			t.Errorf("no page for the ZX81 keyword %s", kw)
		}
	}
	for _, kw := range basic.ZX80Tokens {
		if len(kw) < 2 || strings.ContainsAny(kw, "*=<>") {
			continue
		}
		if len(Find("basic80", kw)) == 0 {
			t.Errorf("no page for the ZX80 keyword %s", kw)
		}
	}
	for _, name := range []string{"**", "<=", "<>"} {
		if len(Find("ref", name)) == 0 {
			t.Errorf("no page for the operator %s", name)
		}
	}
}

func TestLookup(t *testing.T) {
	// The BASIC page comes first; the others are listed at the end.
	out := page(t, "print")
	if !strings.HasPrefix(out, "PRINT(basic)") || !strings.Contains(out, "man shell print") {
		t.Fatalf("man print:\n%s", out)
	}
	if out := page(t, "shell", "print"); !strings.HasPrefix(out, "print(shell)") {
		t.Fatalf("man shell print:\n%s", out)
	}
	if out := page(t, "D_FILE"); !strings.Contains(out, "400Ch (16396)") || !strings.Contains(out, "arquivo de tela") {
		t.Fatalf("man D_FILE:\n%s", out)
	}
	if out := page(t, "FP-CALC"); !strings.Contains(out, "0028h") || !strings.Contains(out, "RST 28") {
		t.Fatalf("man FP-CALC:\n%s", out)
	}
	if out := page(t, "rom", "SLOW/FAST"); !strings.Contains(out, "0207h") {
		t.Fatalf("man rom SLOW/FAST:\n%s", out)
	}
	if out := page(t, "erro", "4"); !strings.Contains(out, "sem memória") {
		t.Fatalf("man erro 4:\n%s", out)
	}
	if out := page(t, "GO SUB"); !strings.HasPrefix(out, "GOSUB(basic)") {
		t.Fatalf("alias GO SUB:\n%s", out)
	}
	if out := page(t, "sysvar"); !strings.Contains(out, "FRAMES") || strings.Contains(out, "FP-CALC") {
		t.Fatalf("man sysvar should list the section:\n%s", out)
	}
	if out := page(t, "-k", "fita"); !strings.Contains(out, "basic LOAD") {
		t.Fatalf("man -k fita:\n%s", out)
	}
	if out := page(t, "-l", "erro"); !strings.Contains(out, "BREAK") {
		t.Fatalf("man -l erro:\n%s", out)
	}
	if out := page(t); !strings.Contains(out, "SEÇÕES") {
		t.Fatalf("man:\n%s", out)
	}
	if out := page(t, "memoria"); !strings.Contains(out, "407D") {
		t.Fatalf("man memoria:\n%s", out)
	}
	// Not a page: the pages that mention it.
	if out := page(t, "som"); !strings.Contains(out, "não há página") || !strings.Contains(out, "record") {
		t.Fatalf("man som (no such page):\n%s", out)
	}
	if _, err := Lookup([]string{"xyzzy123"}, 78); err == nil {
		t.Fatal("unknown page accepted")
	}
	if _, err := Lookup([]string{"-l", "nada"}, 78); err == nil {
		t.Fatal("unknown section accepted")
	}
}

func TestROMLabels(t *testing.T) {
	n := 0
	for _, tp := range Topics() {
		if tp.Section == "rom" {
			n++
		}
	}
	if n < 600 {
		t.Fatalf("only %d ROM labels", n)
	}
	if len(Find("sysvar", "")) != 0 {
		t.Fatal("empty name found")
	}
	sv := 0
	for _, tp := range Topics() {
		if tp.Section == "sysvar" {
			sv++
		}
	}
	if sv != len(sysvars) {
		t.Fatalf("%d system variables, %d described", sv, len(sysvars))
	}
}

func TestWrap(t *testing.T) {
	got := Wrap("uma frase comprida que precisa ser quebrada em mais linhas\n  exemplo que não quebra mesmo sendo bem comprido de verdade", 20)
	for _, l := range got[:len(got)-1] {
		if len([]rune(l)) > 20 {
			t.Fatalf("line too long: %q", l)
		}
	}
	if !strings.HasPrefix(got[len(got)-1], "  exemplo") {
		t.Fatalf("indented line wrapped: %q", got)
	}
}
