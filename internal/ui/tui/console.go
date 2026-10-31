// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/shell"
)

// console is the command panel (F2): the shell's commands inside the
// terminal interface.
type console struct {
	box     *tview.Flex
	out     *tview.TextView
	input   *tview.InputField
	core    *shell.Console
	onQuit  func()
	onClose func()
}

func newConsole(app *tview.Application, sess *session.Session) *console {
	c := &console{core: shell.NewConsole(sess)}
	c.out = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWrap(true)
	c.out.SetChangedFunc(func() { c.out.ScrollToEnd() })
	c.input = tview.NewInputField().SetLabel(shell.Prompt).SetFieldBackgroundColor(tcell.ColorDefault)
	c.box = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(c.out, 0, 1, false).
		AddItem(c.input, 1, 0, true)
	c.box.SetBorder(true).SetTitle(" Console — help lista os comandos · Tab completa · ↑ ↓ histórico · Esc fecha ")
	fmt.Fprintln(c.out, "[gray]Os mesmos comandos do go81 shell. O ZX81 continua rodando.[-]")

	c.input.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyUp:
			c.input.SetText(c.core.Prev(c.input.GetText()))
			return nil
		case tcell.KeyDown:
			c.input.SetText(c.core.Next(c.input.GetText()))
			return nil
		case tcell.KeyTab:
			text, choices := c.core.Complete(c.input.GetText())
			c.input.SetText(text)
			if len(choices) > 0 {
				fmt.Fprintln(c.out, "[gray]"+tview.Escape(strings.Join(choices, "  "))+"[-]")
			}
			return nil
		case tcell.KeyEscape:
			if c.onClose != nil {
				c.onClose()
			}
			return nil
		case tcell.KeyEnter:
			c.submit(app)
			return nil
		}
		return ev
	})
	return c
}

func (c *console) submit(app *tview.Application) {
	line := strings.TrimSpace(c.input.GetText())
	if line == "" {
		return
	}
	ok := c.core.Run(line, func(r shell.Result) {
		app.QueueUpdateDraw(func() {
			if r.Output != "" {
				fmt.Fprint(c.out, tview.Escape(r.Output))
			}
			if r.Err != nil {
				fmt.Fprintln(c.out, "[red]erro: "+tview.Escape(r.Err.Error())+"[-]")
			}
			if r.Quit && c.onQuit != nil {
				c.onQuit()
			}
		})
	})
	if !ok {
		fmt.Fprintln(c.out, "[yellow]espere o comando anterior terminar[-]")
		return
	}
	fmt.Fprintln(c.out, "[::b]"+tview.Escape(shell.Prompt+line)+"[::-]")
	c.input.SetText("")
}
