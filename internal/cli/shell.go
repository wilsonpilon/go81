// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/wilsonpilon/go81/internal/link"
	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/shell"
)

// stdin is the shell input (tests replace it).
var stdin io.Reader = os.Stdin

// stdinIsTerminal reports whether the shell talks to a person.
var stdinIsTerminal = func() bool {
	st, err := os.Stdin.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

// machineFlags adds the machine options shared by shell and link.
func machineFlags(fs *flag.FlagSet) func() (session.Settings, error) {
	romName := fs.String("rom", "zx81.edition3", "ROM embutida (veja \"go81 roms\") ou caminho de um arquivo .rom")
	ram := fs.Int("ram", 16, "RAM pack em K: 0 (só 1K interno), 2, 4, 8, 16, 32 ou 48")
	ntsc := fs.Bool("ntsc", false, "ZX81 americano / TS1000 (60 Hz)")
	wrx := fs.Bool("wrx", false, "alta resolução WRX")
	sound := fs.Bool("sound", false, "som pelo vídeo (Video Sound do EightyOne)")
	return func() (session.Settings, error) {
		switch *ram {
		case 0, 1, 2, 4, 8, 16, 32, 48:
		default:
			return session.Settings{}, fmt.Errorf("tamanho de RAM inválido: %d", *ram)
		}
		r := *ram
		if r == 1 {
			r = 0
		}
		return session.Settings{ROM: *romName, RAMK: r, NTSC: *ntsc, WRX: *wrx, Sound: *sound}, nil
	}
}

func cmdShell(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("shell", flag.ContinueOnError)
	fs.SetOutput(stderr)
	settings := machineFlags(fs)
	var (
		cmds     = fs.String("c", "", "executa estes comandos (separados por ;) e sai")
		linkOn   = fs.Bool("link", false, "liga também a porta de comunicação (link start)")
		port     = fs.Int("port", link.DefaultPort, "porta do protocolo Go81 (com -link)")
		zrcpPort = fs.Int("zrcp", link.DefaultZRCPPort, "porta ZRCP compatível com ZEsarUX (com -link; -1 desliga)")
	)
	fs.Usage = func() {
		fmt.Fprint(stderr, `Uso: go81 shell [opções] [script.g81]

Abre o shell do Go81: cada linha é um comando (help lista todos).
Com um script, executa o arquivo e para no primeiro erro.

Opções:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "go81 shell: só um script por vez")
		return ExitUsage
	}
	set, err := settings()
	if err != nil {
		fmt.Fprintln(stderr, "go81 shell:", err)
		return ExitUsage
	}
	sess, err := session.New(set)
	if err != nil {
		fmt.Fprintln(stderr, "go81 shell:", err)
		return ExitError
	}
	ports := link.Register(sess)
	defer ports.Stop()
	defer sess.StopClock()
	if *linkOn {
		if err := ports.Start(sess, *port, *zrcpPort); err != nil {
			fmt.Fprintln(stderr, "go81 shell:", err)
			return ExitError
		}
		sess.StartClock()
		fmt.Fprintln(stdout, ports.Status())
	}

	var runErr error
	switch {
	case *cmds != "":
		runErr = shell.ExecLine(sess, *cmds, stdout)
		if errors.Is(runErr, session.ErrQuit) {
			runErr = nil
		}
	case fs.NArg() == 1:
		runErr = sess.Exec(fmt.Sprintf("script %q", fs.Arg(0)), stdout)
		if errors.Is(runErr, session.ErrQuit) {
			runErr = nil
		}
	default:
		runErr = shell.Run(sess, stdin, stdout, stderr, shell.Options{Interactive: stdinIsTerminal()})
	}
	if runErr != nil {
		fmt.Fprintln(stderr, "erro:", runErr)
		return ExitError
	}
	return ExitOK
}

func cmdLink(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("link", flag.ContinueOnError)
	fs.SetOutput(stderr)
	settings := machineFlags(fs)
	var (
		port     = fs.Int("port", link.DefaultPort, "porta do protocolo Go81")
		zrcpPort = fs.Int("zrcp", link.DefaultZRCPPort, "porta ZRCP compatível com ZEsarUX (-1 desliga)")
	)
	fs.Usage = func() {
		fmt.Fprint(stderr, `Uso: go81 link [opções] [programa.p]

Roda o ZX81 sem janela, em tempo real, controlado pela porta de
comunicação (só em 127.0.0.1). Ctrl+C encerra.

  Protocolo Go81: uma linha por comando do shell; resposta = saída + "ok"
  ou "err <mensagem>". Teste com: telnet 127.0.0.1 8181
  ZRCP: subconjunto do protocolo do ZEsarUX (get-registers, read-memory...)

Opções:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "go81 link: só um programa por vez")
		return ExitUsage
	}
	set, err := settings()
	if err != nil {
		fmt.Fprintln(stderr, "go81 link:", err)
		return ExitUsage
	}
	sess, err := session.New(set)
	if err != nil {
		fmt.Fprintln(stderr, "go81 link:", err)
		return ExitError
	}
	ports := link.Register(sess)
	if fs.NArg() == 1 {
		if err := sess.Exec(fmt.Sprintf("load %q", fs.Arg(0)), stdout); err != nil {
			fmt.Fprintln(stderr, "go81 link:", err)
			return ExitError
		}
	}
	if err := ports.Start(sess, *port, *zrcpPort); err != nil {
		fmt.Fprintln(stderr, "go81 link:", err)
		return ExitError
	}
	sess.StartClock()
	fmt.Fprintln(stdout, ports.Status())
	fmt.Fprintln(stdout, "Ctrl+C encerra.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	sess.StopClock()
	ports.Stop()
	fmt.Fprintln(stdout, "link encerrado")
	return ExitOK
}
