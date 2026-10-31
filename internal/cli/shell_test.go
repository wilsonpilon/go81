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
