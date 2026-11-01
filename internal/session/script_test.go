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

func TestExpressions(t *testing.T) {
	s, _ := New(DefaultSettings())
	s.SetVar("x", "10")
	s.SetVar("nome", "ZX81")
	cases := map[string]string{
		"1+2*3":            "7",
		"(1+2)*3":          "9",
		"$10 + 0x10 + 'A'": "97",
		"x*2 - 1":          "19",
		"-x + 3":           "-7",
		"17 % 5":           "2",
		"x > 5 && x < 20":  "1",
		"!(x == 10)":       "0",
		`nome == "ZX81"`:   "1",
		`"GO" + "81"`:      "GO81",
		`len("OLA")`:       "3",
		"peek(0)":          "211",
		`reg("IY")`:        "0",
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for in, want := range cases {
		v, err := s.Eval(in)
		if err != nil || v.String() != want {
			t.Errorf("%s = %v (%v), want %s", in, v, err, want)
		}
	}
	for _, bad := range []string{"1/0", "y+1", `"a" - 1`, "1 +", "foo(1)", "(1", `"aberto`} {
		if _, err := s.Eval(bad); err == nil {
			t.Errorf("%s should fail", bad)
		}
	}
}

func runScriptText(t *testing.T, s *Session, text string) (string, error) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "t.g81")
	os.WriteFile(p, []byte(text), 0o644)
	var out bytes.Buffer
	err := s.Exec("script "+p, &out)
	return out.String(), err
}

func TestScriptBlocks(t *testing.T) {
	s, _ := New(DefaultSettings())
	out, err := runScriptText(t, s, `
# laço com variáveis
set soma 0
calc i 1
while i <= 10
  calc soma soma + i
  calc i i + 1
end
echo soma=${soma}
repeat 3
  if soma > 50
    echo grande
  else
    echo pequeno
  end
end
set n 0
while 1
  calc n n + 1
  if n == 4
    break
  end
end
print n * 100
capture v version
assert len(v) > 5
`)
	if err != nil {
		t.Fatal(err)
	}
	want := "soma=55\ngrande\ngrande\ngrande\n400\n"
	if out != want {
		t.Fatalf("output:\n%s\nwant:\n%s", out, want)
	}
}

func TestScriptWithTheZX81(t *testing.T) {
	s, _ := New(DefaultSettings())
	out, err := runScriptText(t, s, `
waitfor 300 screen("")
typeln 10P"OLA"
typeln R
waitfor 200 screen("0/10")
assert screen("OLA")
poke 30000 42
assert peek(30000) == 42
set addr 30000
print peek(${addr}) + 1
`)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if strings.TrimSpace(out) != "43" {
		t.Fatalf("output %q", out)
	}
}

func TestScriptErrors(t *testing.T) {
	s, _ := New(DefaultSettings())
	cases := map[string]string{
		"if 1\necho x\n":            "if sem end",
		"end\n":                     "end sem",
		"echo ${nada}\n":            "variável não definida: nada",
		"set a 1\nassert a == 2\n":  "t.g81:2: assert falhou",
		"fail parou aqui\n":         "parou aqui",
		"break\n":                   "break fora",
		"waitfor 5 frames() < 0\n":  "não ficou verdadeira",
		"repeat \"x\"\necho\nend\n": "repeat precisa de um número",
	}
	for text, want := range cases {
		_, err := runScriptText(t, s, text)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: error %v, want %q", text, err, want)
		}
	}
	// Blocks typed directly (not in a script) are refused.
	if err := s.Exec("if 1", &bytes.Buffer{}); err == nil {
		t.Error("if outside a script should fail")
	}
}

func TestEvents(t *testing.T) {
	s, _ := New(DefaultSettings())
	var got []string
	cancel := s.Subscribe(func(e Event) { got = append(got, e.Kind+" "+e.Text) })
	mustExec(t, s, "run 150")
	mustExec(t, s, `typeln 10P"OI"`)
	mustExec(t, s, "typeln 20PX") // X is not defined: report 2/20
	mustExec(t, s, "typeln R")
	mustExec(t, s, "run 50")
	dir := t.TempDir()
	mustExec(t, s, "save "+filepath.Join(dir, "ev.p")+" EV")
	all := strings.Join(got, "\n")
	for _, want := range []string{"report 2/20", "save EV"} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q in:\n%s", want, all)
		}
	}
	cancel()
	n := len(got)
	mustExec(t, s, "typeln R")
	mustExec(t, s, "run 50")
	if len(got) != n {
		t.Errorf("events after cancel: %v", got[n:])
	}
	// load: a new session loading the program.
	s2, _ := New(DefaultSettings())
	var loads int
	s2.Subscribe(func(e Event) {
		if e.Kind == "load" {
			loads++
		}
	})
	mustExec(t, s2, "load "+filepath.Join(dir, "ev.p"))
	if loads != 1 {
		t.Errorf("load events: %d", loads)
	}
}
