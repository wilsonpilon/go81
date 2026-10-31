// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package shell is the Go81 interactive shell: a small command line where
// each line is a session command (help lists them). The same commands run
// from script files, the consoles of the window and the terminal, and the
// Link port.
package shell

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"

	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/version"
)

// Prompt is shown before each line in interactive mode.
const Prompt = "go81> "

// MaxHistory is how many lines the history file keeps.
const MaxHistory = 1000

// Options control a shell run.
type Options struct {
	// Interactive shows the banner and prompt and keeps going after errors.
	// Otherwise (scripts, pipes) the first error stops the shell.
	Interactive bool
	// LineEditor uses the terminal directly: arrows, history and Tab
	// completion. Only for a real terminal on the process's stdin.
	LineEditor bool
	// HistoryFile keeps the history between runs ("" = DefaultHistoryFile).
	HistoryFile string
}

// DefaultHistoryFile is %APPDATA%\go81\shell_history (Windows) or
// ~/.config/go81/shell_history.
func DefaultHistoryFile() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "go81", "shell_history")
}

// History is the list of lines typed in this run (plus the loaded file).
type History struct {
	lines []string
}

// Add appends a line, skipping blanks and repeats of the last line.
func (h *History) Add(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" || (len(h.lines) > 0 && h.lines[len(h.lines)-1] == line) {
		return false
	}
	h.lines = append(h.lines, line)
	if len(h.lines) > MaxHistory {
		h.lines = h.lines[len(h.lines)-MaxHistory:]
	}
	return true
}

// Lines returns the history, oldest first.
func (h *History) Lines() []string { return h.lines }

// registerHistory adds the "history" command.
func registerHistory(sess *session.Session, h *History) {
	sess.Registry().Register(&session.Command{
		Name: "history", Aliases: []string{"historico"},
		Usage: "history [quantidade]", Help: "mostra os últimos comandos digitados (setas ↑ ↓ os trazem de volta)",
		Run: func(s *session.Session, a []string, out io.Writer) error {
			n := 20
			if len(a) == 1 {
				v, err := session.ParseNumber(a[0])
				if err != nil || v < 1 {
					return fmt.Errorf("uso: history [quantidade]")
				}
				n = v
			}
			l := h.Lines()
			start := max(len(l)-n, 0)
			for i := start; i < len(l); i++ {
				fmt.Fprintf(out, "%5d  %s\n", i+1, l[i])
			}
			return nil
		},
	})
}

// Run reads command lines until end of input or quit. It returns nil
// (quit is not an error) or, in script mode, the first error.
func Run(sess *session.Session, in io.Reader, out, errOut io.Writer, opt Options) error {
	if opt.Interactive {
		fmt.Fprintf(out, "Go81 %s — shell do ZX81. Digite help para ver os comandos, quit para sair.\n", version.Full())
	}
	hist := &History{}
	registerHistory(sess, hist)
	if opt.LineEditor && liner.TerminalSupported() {
		fmt.Fprintln(out, "Setas ↑ ↓ trazem comandos anteriores; Tab completa comandos e arquivos.")
		return runEditor(sess, out, errOut, opt, hist)
	}
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	n := 0
	for {
		if opt.Interactive {
			fmt.Fprint(out, Prompt)
		}
		if !sc.Scan() {
			if opt.Interactive {
				fmt.Fprintln(out)
			}
			return sc.Err()
		}
		n++
		hist.Add(sc.Text())
		if done, err := handle(sess, sc.Text(), out, errOut, opt.Interactive); done {
			if err != nil {
				return fmt.Errorf("linha %d: %w", n, err)
			}
			return nil
		}
	}
}

// handle runs a line. done is true when the shell must stop: after quit
// (err nil) or after an error in script mode.
func handle(sess *session.Session, line string, out, errOut io.Writer, interactive bool) (done bool, err error) {
	err = ExecLine(sess, line, out)
	switch {
	case errors.Is(err, session.ErrQuit):
		return true, nil
	case err == nil:
		return false, nil
	case interactive:
		fmt.Fprintln(errOut, "erro:", err)
		return false, nil
	default:
		return true, err
	}
}

// runEditor is the interactive loop with line editing.
func runEditor(sess *session.Session, out, errOut io.Writer, opt Options, hist *History) error {
	ln := liner.NewLiner()
	defer ln.Close()
	ln.SetCtrlCAborts(true)
	ln.SetTabCompletionStyle(liner.TabPrints)
	ln.SetCompleter(func(line string) []string { return session.Complete(sess.Registry(), line) })

	file := opt.HistoryFile
	if file == "" {
		file = DefaultHistoryFile()
	}
	if file != "" {
		if f, err := os.Open(file); err == nil {
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				if hist.Add(sc.Text()) {
					ln.AppendHistory(strings.TrimSpace(sc.Text()))
				}
			}
			f.Close()
		}
		defer saveHistory(file, hist)
	}

	for {
		line, err := ln.Prompt(Prompt)
		switch {
		case errors.Is(err, liner.ErrPromptAborted): // Ctrl+C: drop the line
			continue
		case errors.Is(err, io.EOF): // Ctrl+D / Ctrl+Z
			fmt.Fprintln(out)
			return nil
		case err != nil:
			return err
		}
		if hist.Add(line) {
			ln.AppendHistory(strings.TrimSpace(line))
		}
		if done, _ := handle(sess, line, out, errOut, true); done {
			return nil
		}
	}
}

func saveHistory(file string, hist *History) {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return
	}
	data := strings.Join(hist.Lines(), "\n")
	if data != "" {
		data += "\n"
	}
	_ = os.WriteFile(file, []byte(data), 0o644)
}

// ExecLine runs a line that may hold several commands separated by ';'.
func ExecLine(sess *session.Session, line string, out io.Writer) error {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}
	for _, c := range session.SplitCommands(line) {
		if err := sess.Exec(c, out); err != nil {
			return err
		}
	}
	return nil
}
