// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package shell

import (
	"bytes"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/termimg"
	"github.com/wilsonpilon/go81/internal/version"
)

// TestScreenshot writes docs/img/shell.png (a real shell session drawn as
// a terminal) when GO81_SCREENSHOTS names the output folder.
func TestScreenshot(t *testing.T) {
	dir := os.Getenv("GO81_SCREENSHOTS")
	if dir == "" {
		t.Skip("GO81_SCREENSHOTS não definida")
	}
	out, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile("../../examples/demo.bas")
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	os.WriteFile(filepath.Join(work, "demo.bas"), src, 0o644)
	old, _ := os.Getwd()
	os.Chdir(work)
	defer os.Chdir(old)

	sess, _ := session.New(session.DefaultSettings())
	lines := []string{"$ go81 shell",
		"Go81 " + version.Full() + " — shell do ZX81. Digite help para ver os comandos, quit para sair."}
	for _, cmd := range []string{"load demo.bas", "list", "peek 16396 4", "make demo.bas demo.p run=10", "ls"} {
		var b bytes.Buffer
		err := ExecLine(sess, cmd, &b)
		lines = append(lines, Prompt+cmd)
		lines = append(lines, strings.Split(strings.TrimRight(b.String(), "\n"), "\n")...)
		if err != nil {
			lines = append(lines, "erro: "+err.Error())
		}
	}
	lines = append(lines, Prompt)
	for i, l := range lines { // the listing has a long line of graphics
		if r := []rune(l); len(r) > 96 {
			lines[i] = string(r[:93]) + "..."
		}
	}
	cells := termimg.FromText(lines)
	green := color.RGBA{0x6A, 0xC9, 0x5C, 0xFF}
	white := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	for y, l := range lines { // prompts in green, commands in white
		n := 0
		switch {
		case strings.HasPrefix(l, Prompt):
			n = len(Prompt)
		case strings.HasPrefix(l, "$ "):
			n = 2
		default:
			continue
		}
		for x := range cells[y] {
			if x < n {
				cells[y][x].FG = green
			} else {
				cells[y][x].FG = white
			}
		}
	}
	if err := termimg.WritePNG(filepath.Join(out, "shell.png"), cells, "go81 shell"); err != nil {
		t.Fatal(err)
	}
}
