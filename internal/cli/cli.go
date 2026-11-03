// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package cli implements the go81 command line.
package cli

import (
	"fmt"
	"github.com/wilsonpilon/go81/internal/man"
	"io"
	"os"

	"github.com/wilsonpilon/go81/internal/version"
)

// Exit codes (MANUAL.md §13).
const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)

const usage = `go81 — emulador e ambiente de desenvolvimento ZX81

Uso:
  go81 <comando> [opções]

Comandos:
  gui        abre o ZX81 numa janela
  tui        abre o ZX81 no terminal
  run        roda o ZX81 sem janela (headless) e pode salvar a tela em PNG
  basic      ferramentas de BASIC (list: programa .p como texto)
  txt2p      converte BASIC em texto (formato do zxtext2p) em .p
  p2txt      converte .p em texto (formato do zxtext2p)
  wav2p      lê fitas gravadas em som (.wav) e grava cada programa num .p
  shell      shell de comandos do Go81 (interativo, script ou -c)
  link       porta de comunicação: outros programas controlam o ZX81
  config     arquivo de configuração: máquina padrão, janela, Link e perfis
  roms       lista as ROMs embutidas
  man        o manual: comandos, BASIC do ZX81 e do ZX80, ROM (go81 man man)
  version    mostra a versão
  help       mostra esta ajuda

Use "go81 <comando> -h" para as opções de cada comando.
`

// Main runs the command line and returns the exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return ExitOK
	}
	switch args[0] {
	case "-v", "--version", "version":
		fmt.Fprintln(stdout, "go81", version.Full())
		return ExitOK
	case "-h", "--help", "help":
		fmt.Fprint(stdout, usage)
		return ExitOK
	case "gui", "--gui":
		return cmdGUI(args[1:], stdout, stderr)
	case "tui", "--tui":
		return cmdTUI(args[1:], stdout, stderr)
	case "run":
		return cmdRun(args[1:], stdout, stderr)
	case "basic":
		return cmdBasic(args[1:], stdout, stderr)
	case "txt2p":
		return cmdTxt2p(args[1:], stdout, stderr)
	case "p2txt":
		return cmdP2txt(args[1:], stdout, stderr)
	case "wav2p":
		return cmdWav2p(args[1:], stdout, stderr)
	case "shell":
		return cmdShell(args[1:], stdout, stderr)
	case "link":
		return cmdLink(args[1:], stdout, stderr)
	case "config":
		return cmdConfig(args[1:], stdout, stderr)
	case "roms":
		return cmdROMs(stdout)
	case "man":
		text, err := man.Lookup(args[1:], 78)
		if err != nil {
			fmt.Fprintln(stderr, "go81 man:", err)
			return ExitError
		}
		fmt.Fprint(stdout, text)
		return ExitOK
	default:
		fmt.Fprintf(stderr, "go81: comando desconhecido: %s\n\n%s", args[0], usage)
		return ExitUsage
	}
}

// Run is Main with the process streams.
func Run() int { return Main(os.Args[1:], os.Stdout, os.Stderr) }
