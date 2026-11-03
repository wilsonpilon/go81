// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/man"
)

// TestEveryCLICommandHasAManPage: each go81 subcommand has a page in the
// cli section, and go81 man works.
func TestEveryCLICommandHasAManPage(t *testing.T) {
	for _, c := range []string{"gui", "tui", "run", "basic", "txt2p", "p2txt", "wav2p", "shell", "link", "config", "roms", "version", "help", "man"} {
		if len(man.Find("cli", "go81 "+c)) == 0 {
			t.Errorf("go81 %s não tem página no man (internal/man/ref.go)", c)
		}
	}
	var out, errb bytes.Buffer
	if code := Main([]string{"man", "basic", "PRINT"}, &out, &errb); code != ExitOK || !strings.HasPrefix(out.String(), "PRINT(basic)") {
		t.Fatalf("go81 man basic PRINT: %d %q %q", code, out.String(), errb.String())
	}
	out.Reset()
	if code := Main([]string{"man", "nada-disso-xyz"}, &out, &errb); code == ExitOK {
		t.Fatal("unknown page should fail")
	}
}

func TestTxt2pP2txtCLI(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"a", "b"} {
		if err := os.WriteFile(filepath.Join(dir, n+".txt"), []byte(`10 PRINT "\:: `+strings.ToUpper(n)+"\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out, errb bytes.Buffer
	if code := Main([]string{"txt2p", "-run", "10", filepath.Join(dir, "*.txt")}, &out, &errb); code != ExitOK {
		t.Fatalf("txt2p: %d %s", code, errb.String())
	}
	if strings.Count(out.String(), " → ") != 2 {
		t.Fatalf("txt2p output: %q", out.String())
	}
	out.Reset()
	if code := Main([]string{"p2txt", "-o", "-", filepath.Join(dir, "b.p")}, &out, &errb); code != ExitOK || out.String() != "10 PRINT \"\\:: B\"\n" {
		t.Fatalf("p2txt -o -: %d %q %s", code, out.String(), errb.String())
	}
	out.Reset()
	if code := Main([]string{"p2txt", filepath.Join(dir, "*.p")}, &out, &errb); code != ExitOK {
		t.Fatalf("p2txt: %d %s", code, errb.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "a.txt")); err != nil {
		t.Fatal(err)
	}
	if code := Main([]string{"txt2p", "-o", "x.p", filepath.Join(dir, "*.txt")}, &out, &errb); code != ExitUsage {
		t.Fatal("-o with many files accepted")
	}
}
