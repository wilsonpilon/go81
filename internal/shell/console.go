// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package shell

import (
	"bytes"
	"errors"
	"strings"
	"sync"

	"github.com/wilsonpilon/go81/internal/session"
)

// Console is the command console of the window and of the terminal: the
// shell's commands, history and Tab completion, with commands running in
// the background (they wait for the frames the interface runs).
type Console struct {
	sess *session.Session
	hist *History

	mu   sync.Mutex
	busy bool
	pos  int    // history position while browsing with ↑ ↓
	draf string // line being typed before browsing
}

// NewConsole creates a console for the session and adds the "history"
// command to it.
func NewConsole(sess *session.Session) *Console {
	c := &Console{sess: sess, hist: &History{}}
	registerHistory(sess, c.hist)
	c.pos = -1
	return c
}

// Busy reports whether a command is running.
func (c *Console) Busy() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.busy
}

// Result is what a command printed.
type Result struct {
	Line   string // the command
	Output string
	Err    error // nil, or the error (never ErrQuit)
	Quit   bool  // the command was quit
}

// Run starts a command line in the background and calls done (from another
// goroutine) when it ends. It returns false if a command is still running.
func (c *Console) Run(line string, done func(Result)) bool {
	c.mu.Lock()
	if c.busy {
		c.mu.Unlock()
		return false
	}
	c.busy = true
	c.hist.Add(line)
	c.pos, c.draf = -1, ""
	c.mu.Unlock()
	go func() {
		var out bytes.Buffer
		err := ExecLine(c.sess, line, &out)
		r := Result{Line: line, Output: out.String(), Err: err}
		if errors.Is(err, session.ErrQuit) {
			r.Err, r.Quit = nil, true
		}
		c.mu.Lock()
		c.busy = false
		c.mu.Unlock()
		done(r)
	}()
	return true
}

// Prompt is the prompt to show: "go81> ", or the line number while auto is on.
func (c *Console) Prompt() string { return c.sess.Prompt() }

// Auto reports whether auto is on (then an empty line counts: it ends auto).
func (c *Console) Auto() bool { return c.sess.Auto() }

// Prev returns the previous history line (↑); current is the text in the
// input, kept to come back to with Next.
func (c *Console) Prev(current string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.hist.Lines()
	if len(l) == 0 {
		return current
	}
	if c.pos == -1 {
		c.pos, c.draf = len(l), current
	}
	if c.pos > 0 {
		c.pos--
	}
	return l[c.pos]
}

// Next returns the next history line (↓), or the line being typed.
func (c *Console) Next(current string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.hist.Lines()
	if c.pos == -1 {
		return current
	}
	c.pos++
	if c.pos >= len(l) {
		c.pos = -1
		return c.draf
	}
	return l[c.pos]
}

// Complete completes the text (Tab). It returns the new text and, when
// there is more than one choice and the text cannot grow, the choices to
// show (only the part after the last space, to keep it short).
func (c *Console) Complete(text string) (string, []string) {
	cands := session.Complete(c.sess.Registry(), text)
	switch len(cands) {
	case 0:
		return text, nil
	case 1:
		return cands[0], nil
	}
	if p := session.CommonPrefix(cands); len(p) > len(text) {
		return p, nil
	}
	// Show only the word being completed: drop the text up to the last
	// space or ';' outside quotes.
	cut, quoted := 0, false
	for i, r := range text {
		switch {
		case r == '"':
			quoted = !quoted
			if quoted {
				cut = i
			}
		case !quoted && (r == ' ' || r == ';'):
			cut = i + 1
		}
	}
	show := make([]string, len(cands))
	for i, s := range cands {
		show[i] = strings.TrimRight(s[min(cut, len(s)):], " ")
	}
	return text, show
}
