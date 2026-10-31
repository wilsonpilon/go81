// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/wilsonpilon/go81/internal/link"
	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/ui/tui"
)

func cmdTUI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	fs.SetOutput(stderr)
	mo := machineFlags(fs)
	var (
		exit     = fs.Int("exit-after", 0, "automação: sai depois de N quadros e imprime a tela de texto")
		console  = fs.Bool("console", false, "abre com o console de comandos (F2) visível")
		linkOn   = fs.Bool("link", false, "liga a porta de comunicação: outros programas controlam este ZX81")
		port     = fs.Int("port", link.DefaultPort, "porta do protocolo Go81 (com -link)")
		zrcpPort = fs.Int("zrcp", link.DefaultZRCPPort, "porta ZRCP (com -link; -1 desliga)")
	)
	fs.Usage = func() {
		fmt.Fprint(stderr, "Uso: go81 tui [opções] [programa.p]\n\nAbre o ZX81 no terminal. F1 ajuda, F2 console de comandos, F10 sai.\n\nOpções:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "go81 tui: só um programa por vez")
		return ExitUsage
	}
	set, cfg, err := mo.load()
	if err != nil {
		fmt.Fprintln(stderr, "go81 tui:", err)
		return ExitUsage
	}
	*port, *zrcpPort = linkPorts(mo, cfg, port, zrcpPort)
	sess, err := session.New(set)
	if err != nil {
		fmt.Fprintln(stderr, "go81 tui:", err)
		return ExitError
	}
	if fs.NArg() == 1 {
		data, err := os.ReadFile(fs.Arg(0))
		if err == nil {
			err = sess.InsertProgram(filepath.Base(fs.Arg(0)), data, true)
		}
		if err != nil {
			fmt.Fprintln(stderr, "go81 tui:", err)
			return ExitError
		}
	}
	ports := link.Register(sess)
	defer ports.Stop()
	if *linkOn {
		if err := ports.Start(sess, *port, *zrcpPort); err != nil {
			fmt.Fprintln(stderr, "go81 tui:", err)
			return ExitError
		}
	}
	opts := tui.Options{Session: sess, ExitAfter: *exit, Console: *console}
	if *exit > 0 {
		opts.OnExit = func(s tui.Screen) { fmt.Fprint(stdout, s.String()) }
	}
	if err := tui.Run(opts); err != nil {
		fmt.Fprintln(stderr, "go81 tui:", err)
		return ExitError
	}
	return ExitOK
}
