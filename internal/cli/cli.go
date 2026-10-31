// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package cli implements the go81 command line.
package cli

import (
	"fmt"
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
  roms       lista as ROMs embutidas
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
	case "roms":
		return cmdROMs(stdout)
	default:
		fmt.Fprintf(stderr, "go81: comando desconhecido: %s\n\n%s", args[0], usage)
		return ExitUsage
	}
}

// Run is Main with the process streams.
func Run() int { return Main(os.Args[1:], os.Stdout, os.Stderr) }
