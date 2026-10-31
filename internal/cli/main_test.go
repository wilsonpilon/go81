// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain keeps the tests away from the user's configuration file.
func TestMain(m *testing.M) {
	dir, _ := os.MkdirTemp("", "go81-cli-test")
	os.Setenv("GO81_CONFIG", filepath.Join(dir, "config.toml"))
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
