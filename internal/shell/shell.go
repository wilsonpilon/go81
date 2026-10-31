// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package shell is the Go81 interactive shell: a small command line where
// each line is a session command (help lists them). The same commands run
// from script files and through the Link port.
package shell

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/version"
)

// Prompt is shown before each line in interactive mode.
const Prompt = "go81> "

// Options control a shell run.
type Options struct {
	// Interactive shows the banner and prompt and keeps going after errors.
	// Otherwise (scripts, pipes) the first error stops the shell.
	Interactive bool
}

// Run reads command lines from in until end of input or quit. It returns
// nil, session.ErrQuit is not an error, or the first error in script mode.
func Run(sess *session.Session, in io.Reader, out, errOut io.Writer, opt Options) error {
	if opt.Interactive {
		fmt.Fprintf(out, "Go81 %s — shell do ZX81. Digite help para ver os comandos, quit para sair.\n", version.Full())
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
		err := ExecLine(sess, sc.Text(), out)
		switch {
		case errors.Is(err, session.ErrQuit):
			return nil
		case err == nil:
		case opt.Interactive:
			fmt.Fprintln(errOut, "erro:", err)
		default:
			return fmt.Errorf("linha %d: %w", n, err)
		}
	}
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
