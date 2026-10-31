// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
)

// newTestSession returns a stock 16K ZX81 that has run n frames.
func newTestSession(t *testing.T, n int) *session.Session {
	t.Helper()
	sess, err := session.New(session.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	sess.Tick(n)
	return sess
}

func textOf(sess *session.Session) [Rows][Cols]Cell {
	var s [Rows][Cols]Cell
	sess.Do(func(m *zx81.Machine) { s = m.TextScreen() })
	return s
}

func TestTextScreenShowsCursorAfterBoot(t *testing.T) {
	sess := newTestSession(t, 160)
	s := textOf(sess)
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
func typeText(t *testing.T, sess *session.Session, s string) {
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
		if err := sess.PressChord(c); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTypingRunsProgram(t *testing.T) {
	sess := newTestSession(t, 150)
	// "p" in K mode is PRINT, "r" is RUN.
	typeText(t, sess, "10p\"ola mundo\"\nr\n")
	sess.Tick(600)
	text := (&Screen{Text: textOf(sess)}).String()
	lines := strings.Split(text, "\n")
	if !strings.HasPrefix(lines[0], "OLA MUNDO") {
		t.Fatalf("first line %q, want OLA MUNDO\n%s", lines[0], text)
	}
	if !strings.HasPrefix(lines[23], "0/10") {
		t.Fatalf("report line %q, want 0/10\n%s", lines[23], text)
	}
}

func TestRunOnSimulatedTerminal(t *testing.T) {
	sim := tcell.NewSimulationScreen("UTF-8")
	sim.SetSize(80, 30)
	var final Screen
	err := Run(Options{
		Session:     newTestSession(t, 0),
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
	sess := newTestSession(t, 160)
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(64, 48)
	sess.Do(func(m *zx81.Machine) { drawPixels(sim, m, 0, 0, 64, 48) })
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

func TestConsoleRunsCommands(t *testing.T) {
	sess := newTestSession(t, 0)
	sim := tcell.NewSimulationScreen("UTF-8")
	sim.SetSize(100, 45)
	go func() {
		time.Sleep(300 * time.Millisecond)
		for _, r := range "poke 30000 9; typeln 10p\"oi\"" {
			sim.InjectKey(tcell.KeyRune, r, tcell.ModNone)
		}
		sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	}()
	var final Screen
	err := Run(Options{Session: sess, Screen: sim, FramePeriod: time.Millisecond,
		ExitAfter: 1500, Console: true, OnExit: func(s Screen) { final = s }})
	if err != nil {
		t.Fatal(err)
	}
	var v byte
	sess.Do(func(m *zx81.Machine) { v = m.Peek(30000) })
	if v != 9 {
		t.Fatalf("poke from the console did not run: %d", v)
	}
	if !strings.Contains(final.String(), "10>PRINT \"OI\"") {
		t.Fatalf("typeln from the console did not type:\n%s", final.String())
	}
}
