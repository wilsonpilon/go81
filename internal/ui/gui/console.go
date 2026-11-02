// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/shell"
)

// maxConsoleLines limits the console text kept on screen.
const maxConsoleLines = 2000

// consoleWindow is the command console (F2): the shell's commands, with
// history and Tab completion, acting on the machine in the main window.
type consoleWindow struct {
	app    fyne.App
	core   *shell.Console
	onQuit func()

	win    fyne.Window
	out    *widget.Label
	scroll *container.Scroll
	input  *consoleEntry
	prompt *widget.Label
	lines  []string
}

func newConsoleWindow(a fyne.App, sess *session.Session, onQuit func()) *consoleWindow {
	return &consoleWindow{app: a, core: shell.NewConsole(sess), onQuit: onQuit}
}

// show opens the console (or brings it to the front).
func (c *consoleWindow) show() {
	if c.win != nil {
		c.win.RequestFocus()
		c.win.Canvas().Focus(c.input)
		return
	}
	c.win = c.app.NewWindow("Go81 — console de comandos")
	c.out = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	c.scroll = container.NewVScroll(c.out)
	c.input = newConsoleEntry()
	c.input.SetPlaceHolder("comando (help lista todos; Tab completa; ↑ ↓ histórico)")
	c.input.onKey = c.key
	c.input.OnSubmitted = c.submit
	c.prompt = widget.NewLabelWithStyle(c.core.Prompt(), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	c.win.SetContent(container.NewBorder(nil, container.NewBorder(nil, nil, c.prompt, nil, c.input), nil, nil, c.scroll))
	c.win.Resize(fyne.NewSize(720, 420))
	c.win.SetOnClosed(func() { c.win = nil })
	if len(c.lines) == 0 {
		c.print("Os mesmos comandos do go81 shell, no ZX81 da janela principal.")
	} else {
		c.refresh()
	}
	c.win.Show()
	c.win.Canvas().Focus(c.input)
}

// print adds text to the console (call from the Fyne thread).
func (c *consoleWindow) print(text string) {
	text = strings.TrimSuffix(text, "\n")
	c.lines = append(c.lines, strings.Split(text, "\n")...)
	if len(c.lines) > maxConsoleLines {
		c.lines = c.lines[len(c.lines)-maxConsoleLines:]
	}
	c.refresh()
}

func (c *consoleWindow) refresh() {
	if c.win == nil {
		return
	}
	c.out.SetText(strings.Join(c.lines, "\n"))
	c.scroll.ScrollToBottom()
}

func (c *consoleWindow) submit(line string) {
	line = strings.TrimSpace(line)
	if line == "" && !c.core.Auto() { // an empty line ends auto
		return
	}
	prompt := c.core.Prompt()
	ok := c.core.Run(line, func(r shell.Result) {
		fyne.Do(func() {
			if c.prompt != nil {
				c.prompt.SetText(c.core.Prompt())
			}
			if r.Output != "" {
				c.print(r.Output)
			}
			if r.Err != nil {
				c.print("erro: " + r.Err.Error())
			}
			if r.Quit && c.onQuit != nil {
				c.onQuit()
			}
		})
	})
	if !ok {
		c.print("(espere o comando anterior terminar)")
		return
	}
	c.print(prompt + line)
	c.input.SetText("")
}

// key handles ↑ ↓ and Tab; it reports whether the key was used.
func (c *consoleWindow) key(k *fyne.KeyEvent) bool {
	switch k.Name {
	case fyne.KeyUp:
		c.setInput(c.core.Prev(c.input.Text))
	case fyne.KeyDown:
		c.setInput(c.core.Next(c.input.Text))
	case fyne.KeyTab:
		text, choices := c.core.Complete(c.input.Text)
		c.setInput(text)
		if len(choices) > 0 {
			c.print(strings.Join(choices, "  "))
		}
	default:
		return false
	}
	return true
}

func (c *consoleWindow) setInput(text string) {
	c.input.SetText(text)
	c.input.CursorColumn = len([]rune(text))
	c.input.Refresh()
}

// consoleEntry is a one-line entry that lets the console handle ↑ ↓ Tab.
type consoleEntry struct {
	widget.Entry
	onKey func(*fyne.KeyEvent) bool
}

func newConsoleEntry() *consoleEntry {
	e := &consoleEntry{}
	e.ExtendBaseWidget(e)
	return e
}

// TypedKey gives ↑ ↓ Tab to the console before the normal entry handling.
func (e *consoleEntry) TypedKey(k *fyne.KeyEvent) {
	if e.onKey != nil && e.onKey(k) {
		return
	}
	e.Entry.TypedKey(k)
}

// AcceptsTab keeps Tab in the entry (for completion) instead of moving the
// focus.
func (e *consoleEntry) AcceptsTab() bool { return true }
