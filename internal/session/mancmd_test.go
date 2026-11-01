// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/man"
)

// TestEveryCommandHasAManPage: a new command needs a page in internal/man.
func TestEveryCommandHasAManPage(t *testing.T) {
	s, err := New(DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range s.reg.Commands() {
		if len(man.Find("shell", c.Name)) == 0 {
			t.Errorf("o comando %s não tem página no man (internal/man/shell.go)", c.Name)
		}
	}
	if out := mustExec(t, s, "man shell load"); !strings.HasPrefix(out, "load(shell)") {
		t.Fatalf("man shell load:\n%s", out)
	}
	if out := mustExec(t, s, "man D_FILE"); !strings.Contains(out, "16396") {
		t.Fatalf("man D_FILE:\n%s", out)
	}
}
