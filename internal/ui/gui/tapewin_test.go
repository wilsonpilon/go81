// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/wilsonpilon/go81/internal/session"
)

// tapeWithPrograms is a session whose tape has three programs saved by the
// ZX81 (and the text block of a new tape).
func tapeWithPrograms(t *testing.T) *session.Session {
	t.Helper()
	dir := t.TempDir()
	sess, err := session.New(session.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	in := filepath.Join(dir, "demo.txt")
	os.WriteFile(in, []byte("10 PRINT \"OLA\"\n20 GOTO 10\n"), 0o644)
	for _, line := range []string{"run 150", `load "` + in + `"`, "tape new",
		`save "` + filepath.Join(dir, "a.p") + `" JOGO`,
		`save "` + filepath.Join(dir, "b.p") + `" NAVES`,
		`save "` + filepath.Join(dir, "c.p") + `" LABIRINTO`, "tape ir 2"} {
		if err := sess.Exec(line, io.Discard); err != nil {
			t.Fatalf("%s: %v", line, err)
		}
	}
	return sess
}

// The tape manager window lists the blocks, marks the current one and
// edits the tape with the tape commands.
func TestTapeManagerWindowScreenshot(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(theme.DefaultTheme())
	sess := tapeWithPrograms(t)
	e := newEmulator(sess, 2)
	v := e.readTapeView()
	if len(v.rows) != 4 || v.cur != 1 || !strings.Contains(v.rows[1], "programa NAVES") || !v.programs[1] || v.programs[3] {
		t.Fatalf("view: %+v", v)
	}
	if err := sess.Exec("tape mover 3 1", io.Discard); err != nil {
		t.Fatal(err)
	}
	if v := e.readTapeView(); !strings.Contains(v.rows[0], "LABIRINTO") {
		t.Fatalf("after move: %q", v.rows)
	}
	win := e.tapeManager(a, func() {})
	win.Show()
	if dir := os.Getenv("GO81_SCREENSHOTS"); dir != "" {
		out, _ := filepath.Abs(dir)
		if err := savePNG(filepath.Join(out, "gui-fitas.png"), win.Canvas().Capture()); err != nil {
			t.Fatal(err)
		}
	}
	win.Close()
}
