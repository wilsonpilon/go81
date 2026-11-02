// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/basic"
)

func TestGameConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GO81_CONFIG", filepath.Join(dir, "config.toml"))
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(ConfigTemplate+`
[jogos.dois]
perfil = "wrx"
ram = 32
`), 0o644); err != nil {
		t.Fatal(err)
	}
	write := func(name, text string) string {
		p, err := basic.TextToP(text, basic.POptions{})
		if err != nil {
			t.Fatal(err)
		}
		f := filepath.Join(dir, name)
		if err := os.WriteFile(f, p, 0o644); err != nil {
			t.Fatal(err)
		}
		return f
	}
	um := write("um.p", "10 PRINT 1\n")
	dois := write("DOIS.P", "10 PRINT 2\n")
	tres := write("tres.p", "10 PRINT 3\n")
	if err := os.WriteFile(filepath.Join(dir, "um.toml"), []byte("ram = 4\nntsc = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := newBooted(t)
	before := s.Settings()
	out := mustExec(t, s, `load "`+um+`"`)
	if set := s.Settings(); set.RAMK != 4 || !set.NTSC || !strings.Contains(out, "máquina (um.toml)") {
		t.Fatalf("um.toml: %+v %q", set, out)
	}
	if out := mustExec(t, s, "list"); out != "10 PRINT 1\n" {
		t.Fatalf("program after the machine changed: %q", out)
	}
	// The second game starts from the machine before the first one.
	out = mustExec(t, s, `load "`+dois+`"`)
	if set := s.Settings(); set.RAMK != 32 || !set.WRX || set.NTSC || !strings.Contains(out, `[jogos."dois"]`) {
		t.Fatalf("[jogos.dois]: %+v %q", set, out)
	}
	if out := mustExec(t, s, "jogo"); !strings.Contains(out, `configuração: [jogos."dois"]`) {
		t.Fatalf("jogo = %q", out)
	}
	// No configuration: back to the machine before.
	out = mustExec(t, s, `load "`+tres+`"`)
	if s.Settings() != before || !strings.Contains(out, "a máquina de antes") {
		t.Fatalf("back: %+v %q", s.Settings(), out)
	}
	if out := mustExec(t, s, "list"); out != "10 PRINT 3\n" {
		t.Fatalf("program 3: %q", out)
	}
	// jogo salvar writes tres.toml; nojogo ignores it.
	mustExec(t, s, "machine ram=8")
	mustExec(t, s, "jogo salvar")
	data, err := os.ReadFile(filepath.Join(dir, "tres.toml"))
	if err != nil || !strings.Contains(string(data), "ram  = 8") {
		t.Fatalf("tres.toml: %q %v", data, err)
	}
	mustExec(t, s, "machine ram=16")
	mustExec(t, s, `load "`+tres+`" nojogo`)
	if s.Settings().RAMK != 16 {
		t.Fatal("nojogo changed the machine")
	}
	mustExec(t, s, `load "`+tres+`"`)
	if s.Settings().RAMK != 8 {
		t.Fatalf("tres.toml not used: %+v", s.Settings())
	}
	mustExec(t, s, "jogo apagar")
	if _, err := os.Stat(filepath.Join(dir, "tres.toml")); err == nil {
		t.Fatal("jogo apagar")
	}
	// A bad file is an error, not ignored.
	if err := os.WriteFile(filepath.Join(dir, "um.toml"), []byte("rampa = 4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Exec(`load "`+um+`"`, &strings.Builder{}); err == nil || !strings.Contains(err.Error(), "rampa") {
		t.Fatalf("bad um.toml: %v", err)
	}
}
