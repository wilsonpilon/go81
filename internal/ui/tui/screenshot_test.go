// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tui

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/termimg"
)

// TestScreenshot writes docs/img/tui.png when GO81_SCREENSHOTS names the
// output folder (tools/screenshots.ps1).
func TestScreenshot(t *testing.T) {
	dir := os.Getenv("GO81_SCREENSHOTS")
	if dir == "" {
		t.Skip("GO81_SCREENSHOTS não definida")
	}
	text, err := os.ReadFile("../../../examples/demo.bas")
	if err != nil {
		t.Fatal(err)
	}
	p, err := basic.TextToP(string(text), basic.POptions{AutoRun: 10})
	if err != nil {
		t.Fatal(err)
	}
	sess, _ := session.New(session.DefaultSettings())
	if err := sess.InsertProgram("demo.p", p, true); err != nil {
		t.Fatal(err)
	}
	sim := tcell.NewSimulationScreen("UTF-8")
	var cells [][]termimg.Cell
	go func() {
		time.Sleep(300 * time.Millisecond)
		sim.SetSize(112, 42)        // after Init, which sets 80x25
		time.Sleep(7 * time.Second) // the demo draws its screen
		for _, cmd := range []string{"peek 16396 2", "regs"} {
			for _, r := range cmd {
				sim.InjectKey(tcell.KeyRune, r, tcell.ModNone)
			}
			sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
			time.Sleep(500 * time.Millisecond)
		}
		for _, r := range "lis" { // left in the input line
			sim.InjectKey(tcell.KeyRune, r, tcell.ModNone)
		}
		time.Sleep(700 * time.Millisecond)
		cells = grab(sim)
		sim.InjectKey(tcell.KeyF10, 0, tcell.ModNone)
	}()
	if err := Run(Options{Session: sess, Screen: sim, FramePeriod: 5 * time.Millisecond, Console: true}); err != nil {
		t.Fatal(err)
	}
	if cells == nil {
		t.Fatal("no screen captured")
	}
	if err := termimg.WritePNG(filepath.Join(dir, "tui.png"), cells, "go81 tui"); err != nil {
		t.Fatal(err)
	}
}

// grab copies the simulated terminal into cells with real colours.
func grab(sim tcell.SimulationScreen) [][]termimg.Cell {
	sc, w, h := sim.GetContents()
	g := make([][]termimg.Cell, h)
	for y := range h {
		g[y] = make([]termimg.Cell, w)
		for x := range w {
			c := sc[y*w+x]
			fg, bg, attr := c.Style.Decompose()
			cell := termimg.Cell{R: ' ', FG: rgb(fg, termimg.DefaultFG), BG: rgb(bg, termimg.DefaultBG)}
			if attr&tcell.AttrReverse != 0 {
				cell.FG, cell.BG = cell.BG, cell.FG
			}
			if len(c.Runes) > 0 {
				cell.R = c.Runes[0]
			}
			g[y][x] = cell
		}
	}
	return g
}

func rgb(c tcell.Color, def color.RGBA) color.RGBA {
	r, g, b := c.RGB()
	if r < 0 {
		return def
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), 0xFF}
}
