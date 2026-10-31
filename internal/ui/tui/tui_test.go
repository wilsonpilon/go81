// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

func newTestModel(t *testing.T) *model {
	t.Helper()
	rom, err := assets.ROM("zx81.edition3")
	if err != nil {
		t.Fatal(err)
	}
	md, err := newModel(zx81.DefaultConfig(rom))
	if err != nil {
		t.Fatal(err)
	}
	return md
}

func TestTextScreenShowsCursorAfterBoot(t *testing.T) {
	md := newTestModel(t)
	for i := 0; i < 160; i++ {
		md.frame()
	}
	s := TextScreen(md.m)
	if c := s[23][0]; c.R != 'K' || !c.Inverse {
		t.Fatalf("bottom left = %q inverse=%v, want inverse K", c.R, c.Inverse)
	}
}

func TestChordForKey(t *testing.T) {
	cases := []struct {
		ev   *tcell.EventKey
		want string
	}{
		{tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone), "A"},
		{tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModShift), "Q"},
		{tcell.NewEventKey(tcell.KeyRune, '7', tcell.ModNone), "7"},
		{tcell.NewEventKey(tcell.KeyRune, '"', tcell.ModShift), "SHIFT+P"},
		{tcell.NewEventKey(tcell.KeyRune, '=', tcell.ModNone), "SHIFT+L"},
		{tcell.NewEventKey(tcell.KeyRune, ',', tcell.ModNone), "SHIFT+."},
		{tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModAlt), "SHIFT+F"},
		{tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), "ENTER"},
		{tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone), "SHIFT+0"},
		{tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), "SHIFT+7"},
		{tcell.NewEventKey(tcell.KeyRune, '!', tcell.ModNone), ""},
		{tcell.NewEventKey(tcell.KeyF3, 0, tcell.ModNone), ""},
	}
	for _, c := range cases {
		if got := chordForKey(c.ev); got != c.want {
			t.Errorf("%v: got %q, want %q", c.ev.Name(), got, c.want)
		}
	}
}

// typeText types a string through the same path as the terminal.
func typeText(t *testing.T, md *model, s string) {
	t.Helper()
	for _, r := range s {
		var ev *tcell.EventKey
		if r == '\n' {
			ev = tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		} else {
			ev = tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
		}
		c := chordForKey(ev)
		if c == "" {
			t.Fatalf("no ZX81 key for %q", r)
		}
		if err := md.typeChord(c); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTypingRunsProgram(t *testing.T) {
	md := newTestModel(t)
	for i := 0; i < 150; i++ {
		md.frame()
	}
	// "p" in K mode is PRINT, "r" is RUN.
	typeText(t, md, "10p\"ola mundo\"\nr\n")
	for i := 0; i < 600; i++ {
		md.frame()
	}
	text := (&Screen{Text: TextScreen(md.m)}).String()
	lines := strings.Split(text, "\n")
	if !strings.HasPrefix(lines[0], "OLA MUNDO") {
		t.Fatalf("first line %q, want OLA MUNDO\n%s", lines[0], text)
	}
	if !strings.HasPrefix(lines[23], "0/10") {
		t.Fatalf("report line %q, want 0/10\n%s", lines[23], text)
	}
}

func TestRunOnSimulatedTerminal(t *testing.T) {
	rom, _ := assets.ROM("zx81.edition3")
	sim := tcell.NewSimulationScreen("UTF-8")
	sim.SetSize(80, 30)
	var final Screen
	err := Run(Options{
		Config:      zx81.DefaultConfig(rom),
		Description: "teste",
		Screen:      sim,
		FramePeriod: time.Millisecond,
		ExitAfter:   200,
		OnExit:      func(s Screen) { final = s },
	})
	if err != nil {
		t.Fatal(err)
	}
	if c := final.Text[23][0]; c.R != 'K' || !c.Inverse {
		t.Fatalf("final screen bottom left = %q", c.R)
	}
}

func TestPixelModeDrawsCursor(t *testing.T) {
	md := newTestModel(t)
	for i := 0; i < 160; i++ {
		md.frame()
	}
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(64, 48)
	drawPixels(sim, md.m, 0, 0, 64, 48)
	// The inverse K cursor (display row 23, column 0) covers cells
	// x 0-1, y 46-47 at half resolution.
	ink := 0
	for y := 46; y < 48; y++ {
		for x := 0; x < 2; x++ {
			r, _, _, _ := sim.GetContent(x, y)
			if r != ' ' {
				ink++
			}
		}
	}
	if ink == 0 {
		t.Fatal("cursor not drawn in pixel mode")
	}
}
