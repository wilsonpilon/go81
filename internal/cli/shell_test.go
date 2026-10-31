// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellDashC(t *testing.T) {
	var out, errb bytes.Buffer
	code := Main([]string{"shell", "-c", "run 150; typeln P2*21; screen"}, &out, &errb)
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "|42 ") {
		t.Fatalf("output:\n%s", out.String())
	}
}

func TestShellScriptStopsOnError(t *testing.T) {
	sc := filepath.Join(t.TempDir(), "s.g81")
	os.WriteFile(sc, []byte("echo antes\nnaoexiste\necho depois\n"), 0o644)
	var out, errb bytes.Buffer
	if code := Main([]string{"shell", sc}, &out, &errb); code != ExitError {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out.String(), "antes") || strings.Contains(out.String(), "depois") {
		t.Fatalf("output %q", out.String())
	}
	if !strings.Contains(errb.String(), "s.g81:2") {
		t.Fatalf("stderr %q", errb.String())
	}
}

func TestShellStdin(t *testing.T) {
	old, oldTerm := stdin, stdinIsTerminal
	defer func() { stdin, stdinIsTerminal = old, oldTerm }()
	stdin = strings.NewReader("version\npeek 0\nquit\necho nunca\n")
	stdinIsTerminal = func() bool { return false }
	var out, errb bytes.Buffer
	if code := Main([]string{"shell", "-ram", "4"}, &out, &errb); code != ExitOK {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	s := out.String()
	if !strings.Contains(s, "go81 ") || !strings.Contains(s, "0000: D3") || strings.Contains(s, "nunca") {
		t.Fatalf("output %q", s)
	}
}

func TestLinkStdio(t *testing.T) {
	old := stdin
	defer func() { stdin = old }()
	stdin = strings.NewReader("poke 30000 7\npeek 30000\n{\"cmd\":\"echo ok\"}\nquit\npeek 0\n")
	var out, errb bytes.Buffer
	if code := Main([]string{"link", "-stdio"}, &out, &errb); code != ExitOK {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	want := []string{"go81 link ", "ok", "7530: 07 (7)", "ok", `{"ok":true,"output":"ok\n"}`, "ok tchau"}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != len(want) {
		t.Fatalf("output:\n%s", out.String())
	}
	for i, w := range want {
		if !strings.HasPrefix(lines[i], w) {
			t.Fatalf("line %d = %q, want %q\n%s", i, lines[i], w, out.String())
		}
	}
	if errb.Len() != 0 {
		t.Fatalf("stderr: %q", errb.String())
	}
}

func TestBasicMakeAndList(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "ola.bas")
	text := "10 FOR I=1 TO 3\n20 PRINT \"OLA \";I*0.5\n30 NEXT I\n"
	os.WriteFile(src, []byte(text), 0o644)
	var out, errb bytes.Buffer
	if code := Main([]string{"basic", "make", src}, &out, &errb); code != ExitOK {
		t.Fatalf("make: exit %d: %s", code, errb.String())
	}
	p := filepath.Join(dir, "ola.p")
	out.Reset()
	if code := Main([]string{"basic", "list", p}, &out, &errb); code != ExitOK {
		t.Fatalf("list: exit %d: %s", code, errb.String())
	}
	if out.String() != text {
		t.Fatalf("list after make:\n%s", out.String())
	}
	// The shell loads the text directly and runs it.
	out.Reset()
	if code := Main([]string{"shell", "-c", "load " + src + "; typeln R; run 200; screen"}, &out, &errb); code != ExitOK {
		t.Fatalf("shell: exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "|OLA 0.5") || !strings.Contains(out.String(), "|OLA 1.5") {
		t.Fatalf("program did not run:\n%s", out.String())
	}
	// Errors name the line.
	os.WriteFile(src, []byte("10 CLS\n20 PRINT !\n"), 0o644)
	errb.Reset()
	if code := Main([]string{"basic", "make", src}, &out, &errb); code != ExitError || !strings.Contains(errb.String(), "linha 2") {
		t.Fatalf("bad text: exit %d: %s", code, errb.String())
	}
}
