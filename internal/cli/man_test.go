// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/man"
)

// TestEveryCLICommandHasAManPage: each go81 subcommand has a page in the
// cli section, and go81 man works.
func TestEveryCLICommandHasAManPage(t *testing.T) {
	for _, c := range []string{"gui", "tui", "run", "basic", "shell", "link", "config", "roms", "version", "help", "man"} {
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
