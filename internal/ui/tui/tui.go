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
	"github.com/wilsonpilon/go81/internal/version"
)

// Options configure the terminal interface.
type Options struct {
	Config      zx81.Config
	Description string // e.g. "ZX81 16K PAL"
	TapeName    string
	TapeData    []byte

	// Screen replaces the real terminal (tests use a simulation screen).
	Screen tcell.Screen
	// FramePeriod overrides the frame time (default 1/50 s or 1/60 s).
	FramePeriod time.Duration
	// ExitAfter closes the interface after this many frames (automation).
	ExitAfter int
	// OnExit, if set, receives the final text screen.
	OnExit func(Screen)
}

// quadrants: index bit 0 = top-left, 1 = top-right, 2 = bottom-left, 3 = bottom-right.
var quadrants = []rune(" ▘▝▀▖▌▞▛▗▚▐▜▄▙▟█")

// Run starts the terminal interface and blocks until the user quits.
func Run(opts Options) error {
	md, err := newModel(opts.Config)
	if err != nil {
		return err
	}
	if opts.TapeData != nil {
		if err := md.insert(opts.TapeName, opts.TapeData); err != nil {
			return err
		}
	}
	var mu sync.Mutex

	app := tview.NewApplication()
	if opts.Screen != nil {
		app.SetScreen(opts.Screen)
	}

	status := tview.NewTextView().SetDynamicColors(true)
	view := tview.NewBox()
	view.SetDrawFunc(func(scr tcell.Screen, x, y, w, h int) (int, int, int, int) {
		mu.Lock()
		defer mu.Unlock()
		if md.pixels {
			drawPixels(scr, md.m, x, y, w, h)
		} else {
			drawText(scr, TextScreen(md.m), x, y, w, h)
		}
		return x, y, w, h
	})

	pages := tview.NewPages()
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(view, 0, 1, true).
		AddItem(status, 1, 0, false)
	pages.AddPage("main", layout, true, true)
	help := tview.NewModal().SetText(helpText).AddButtons([]string{"OK"}).
		SetDoneFunc(func(int, string) { pages.HidePage("help") })
	pages.AddPage("help", help, true, false)

	app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if name, _ := pages.GetFrontPage(); name == "help" {
			return ev
		}
		mu.Lock()
		defer mu.Unlock()
		switch ev.Key() {
		case tcell.KeyF1:
			pages.ShowPage("help")
			return nil
		case tcell.KeyF5:
			md.paused = !md.paused
			return nil
		case tcell.KeyF6:
			md.turbo = !md.turbo
			return nil
		case tcell.KeyF8:
			md.m.Reset()
			return nil
		case tcell.KeyF9:
			md.pixels = !md.pixels
			return nil
		case tcell.KeyF10, tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		if chord := chordForKey(ev); chord != "" {
			_ = md.typeChord(chord)
		}
		return nil
	})

	period := opts.FramePeriod
	if period == 0 {
		period = time.Second / 50
		if opts.Config.NTSC {
			period = time.Second / 60
		}
	}
	desc := opts.Description
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
			mu.Lock()
			if !md.paused {
				n := 1
				if md.turbo {
					n = 10
				}
				for i := 0; i < n; i++ {
					md.frame()
				}
				total += n
			}
			state := ""
			switch {
			case md.paused:
				state = " [yellow]PAUSADO[-]"
			case md.turbo:
				state = " [yellow]TURBO[-]"
			}
			if md.m.Tape != nil && md.m.Tape.Player.Playing {
				state += " [green]fita[-]"
			}
			mode := "texto"
			if md.pixels {
				mode = "pixels"
			}
			line := fmt.Sprintf(" Go81 %s · %s · tela: %s%s · [::b]F1[::-] ajuda  [::b]F10[::-] sair",
				version.Number(), desc, mode, state)
			done := opts.ExitAfter > 0 && total >= opts.ExitAfter
			mu.Unlock()
			app.QueueUpdateDraw(func() { status.SetText(line) })
			if done {
				app.Stop()
				return
			}
		}
	}()

	err = app.SetRoot(pages, true).Run()
	close(stop)
	if opts.OnExit != nil {
		mu.Lock()
		opts.OnExit(Screen{Text: TextScreen(md.m)})
		mu.Unlock()
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

// drawPixels draws the TV picture (display area only, 256x192) at half
// resolution with quadrant characters: 64x48 cells.
func drawPixels(scr tcell.Screen, m *zx81.Machine, x, y, w, h int) {
	img := m.Screen(zx81.BorderNone)
	b := img.Bounds()
	cw, ch := b.Dx()/4, b.Dy()/4
	ox := x + max((w-cw)/2, 0)
	oy := y + max((h-ch)/2, 0)
	dark := func(px, py int) bool {
		// A 2x2 block of the picture is "ink" if most of it is dark.
		n := 0
		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 2; dx++ {
				c := img.RGBAAt(px+dx, py+dy)
				if int(c.R)+int(c.G)+int(c.B) < 3*128 {
					n++
				}
			}
		}
		return n >= 2
	}
	for cy := 0; cy < ch && cy < h; cy++ {
		for cx := 0; cx < cw && cx < w; cx++ {
			px, py := cx*4, cy*4
			q := 0
			if dark(px, py) {
				q |= 1
			}
			if dark(px+2, py) {
				q |= 2
			}
			if dark(px, py+2) {
				q |= 4
			}
			if dark(px+2, py+2) {
				q |= 8
			}
			scr.SetContent(ox+cx, oy+cy, quadrants[q], nil, paperStyle)
		}
	}
}
