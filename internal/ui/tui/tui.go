// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package tui is the terminal interface of the emulator (tview/tcell).
package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/version"
)

// Options configure the terminal interface.
type Options struct {
	// Session is the machine. The terminal runs its frames in real time,
	// so its commands (console, Link) wait for the terminal.
	Session *session.Session

	// Screen replaces the real terminal (tests use a simulation screen).
	Screen tcell.Screen
	// FramePeriod overrides the frame time (default 1/50 s or 1/60 s).
	FramePeriod time.Duration
	// ExitAfter closes the interface after this many frames (automation).
	ExitAfter int
	// OnExit, if set, receives the final text screen.
	OnExit func(Screen)
	// Console opens with the command console visible.
	Console bool
}

// quadrants: index bit 0 = top-left, 1 = top-right, 2 = bottom-left, 3 = bottom-right.
var quadrants = []rune(" ▘▝▀▖▌▞▛▗▚▐▜▄▙▟█")

// ui holds the terminal state that is not the machine.
type ui struct {
	mu    sync.Mutex
	turbo bool
	mode  int // modeWide, modePixels, modeCompact

	message string    // shown in the status line for a few seconds
	msgEnd  time.Time // when it goes away
}

// Run starts the terminal interface and blocks until the user quits.
func Run(opts Options) error {
	sess := opts.Session
	sess.Drive()
	st := &ui{}

	app := tview.NewApplication()
	if opts.Screen != nil {
		app.SetScreen(opts.Screen)
	}

	status := tview.NewTextView().SetDynamicColors(true)
	view := &zxView{Box: tview.NewBox()}
	app.EnablePaste(true)
	view.SetDrawFunc(func(scr tcell.Screen, x, y, w, h int) (int, int, int, int) {
		st.mu.Lock()
		mode := st.mode
		st.mu.Unlock()
		sess.Do(func(m *zx81.Machine) { draw(scr, m, mode, x, y, w, h) })
		return x, y, w, h
	})

	// Text pasted on the ZX81 screen: a BASIC program is loaded, other
	// text is typed (the console's input field takes its own pastes).
	view.onPaste = func(text string) {
		msg, err := sess.Paste(text)
		if err != nil {
			msg = "erro: " + err.Error()
		}
		st.mu.Lock()
		st.message, st.msgEnd = msg, time.Now().Add(4*time.Second)
		st.mu.Unlock()
	}
	con := newConsole(app, sess)
	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	consoleOpen := false
	setConsole := func(open bool) {
		consoleOpen = open
		layout.Clear()
		layout.AddItem(view, 0, 1, !open)
		if open {
			layout.AddItem(con.box, 12, 0, true)
			app.SetFocus(con.input)
		} else {
			app.SetFocus(view)
		}
		layout.AddItem(status, 1, 0, false)
	}
	setConsole(opts.Console)

	pages := tview.NewPages()
	pages.AddPage("main", layout, true, true)
	help := tview.NewModal().SetText(helpText).AddButtons([]string{"OK"}).
		SetDoneFunc(func(int, string) {
			pages.HidePage("help")
			setConsole(consoleOpen)
		})
	pages.AddPage("help", help, true, false)
	con.onQuit = app.Stop
	con.onClose = func() { setConsole(false) }

	app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if name, _ := pages.GetFrontPage(); name == "help" {
			return ev
		}
		switch ev.Key() {
		case tcell.KeyF1:
			pages.ShowPage("help")
			return nil
		case tcell.KeyF2:
			setConsole(!consoleOpen)
			return nil
		case tcell.KeyF5:
			sess.SetPaused(!sess.Paused())
			return nil
		case tcell.KeyF6:
			st.mu.Lock()
			st.turbo = !st.turbo
			st.mu.Unlock()
			return nil
		case tcell.KeyF8:
			sess.Reset()
			return nil
		case tcell.KeyF9:
			st.mu.Lock()
			st.mode = (st.mode + 1) % modeCount
			st.mu.Unlock()
			return nil
		case tcell.KeyF10, tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		if consoleOpen {
			return ev // the console's input field gets the key
		}
		if chord := chordForKey(ev); chord != "" {
			_ = sess.PressChord(chord)
		}
		return nil
	})

	period := opts.FramePeriod
	if period == 0 {
		period = time.Second / 50
		if sess.Settings().NTSC {
			period = time.Second / 60
		}
	}
	stop := make(chan struct{})
	go func() {
		tick := time.NewTicker(period)
		defer tick.Stop()
		total := 0
		for {
			select {
			case <-stop:
				return
			case <-tick.C:
			}
			st.mu.Lock()
			n := 1
			if st.turbo {
				n = 10
			}
			turbo, mode := st.turbo, st.mode
			st.mu.Unlock()
			if sess.Tick(n) {
				total += n
			}
			state := ""
			switch {
			case sess.Paused():
				state = " [yellow]PAUSADO[-]"
			case turbo:
				state = " [yellow]TURBO[-]"
			}
			sess.Do(func(m *zx81.Machine) {
				if m.Tape != nil && m.Tape.Player.Playing {
					state += " [green]fita[-]"
				}
			})
			_, _, vw, _ := view.GetInnerRect()
			modeName := modeNames[effectiveMode(mode, vw)]
			st.mu.Lock()
			flash := st.message != "" && time.Now().Before(st.msgEnd)
			msg := st.message
			st.mu.Unlock()
			if flash {
				state += " · [green]" + tview.Escape(msg) + "[-]"
			}
			line := fmt.Sprintf(" Go81 %s · %s · tela: %s%s · [::b]F1[::-] ajuda  [::b]F2[::-] console  [::b]F10[::-] sair",
				version.Number(), sess.Description(), modeName, state)
			done := opts.ExitAfter > 0 && total >= opts.ExitAfter
			app.QueueUpdateDraw(func() { status.SetText(line) })
			if done {
				app.Stop()
				return
			}
		}
	}()

	err := app.SetRoot(pages, true).Run()
	close(stop)
	if opts.OnExit != nil {
		var s Screen
		sess.Do(func(m *zx81.Machine) { s.Text = m.TextScreen() })
		opts.OnExit(s)
	}
	return err
}

var (
	paperStyle = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	inkStyle   = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	greyPaper  = paperStyle.Foreground(tcell.ColorGray)
	greyInk    = inkStyle.Foreground(tcell.ColorGray)
)

// drawText draws the 32x24 text screen with a white border, centred.
func drawText(scr tcell.Screen, s [Rows][Cols]Cell, x, y, w, h int) {
	const bw, bh = Cols + 4, Rows + 2
	ox := x + max((w-bw)/2, 0)
	oy := y + max((h-bh)/2, 0)
	for j := 0; j < bh && j < h; j++ {
		for i := 0; i < bw && i < w; i++ {
			scr.SetContent(ox+i, oy+j, ' ', nil, paperStyle)
		}
	}
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			cell := s[r][c]
			st := paperStyle
			if cell.Inverse {
				st = inkStyle
			}
			if cell.Grey {
				if cell.Inverse {
					st = greyInk
				} else {
					st = greyPaper
				}
			}
			px, py := ox+2+c, oy+1+r
			if px < x+w && py < y+h {
				scr.SetContent(px, py, cell.R, nil, st)
			}
		}
	}
}

// zxView is the ZX81 screen; it takes pasted text.
type zxView struct {
	*tview.Box
	onPaste func(string)
}

// PasteHandler gives pasted text to onPaste.
func (v *zxView) PasteHandler() func(string, func(tview.Primitive)) {
	return v.WrapPasteHandler(func(text string, _ func(tview.Primitive)) {
		if v.onPaste != nil {
			v.onPaste(text)
		}
	})
}
