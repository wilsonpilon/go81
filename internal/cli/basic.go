// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/wilsonpilon/go81/internal/basic"
)

const basicUsage = `Uso: go81 basic <comando> [opções]

Ferramentas de BASIC do ZX81.

Comandos:
  list <programa.p> [-o arquivo.txt]   lista o programa como texto (formato do zxtext2p)
`

func cmdBasic(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, basicUsage)
		return ExitOK
	}
	switch args[0] {
	case "list":
		return cmdBasicList(args[1:], stdout, stderr)
	}
	fmt.Fprintf(stderr, "go81 basic: comando desconhecido: %s\n\n%s", args[0], basicUsage)
	return ExitUsage
}

func cmdBasicList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("basic list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outFile := fs.String("o", "", "grava a listagem neste arquivo (padrão: a tela)")
	fs.Usage = func() {
		fmt.Fprint(stderr, "Uso: go81 basic list [-o arquivo.txt] <programa.p>\n\nLista um programa .p como texto.\n\nOpções:\n")
		fs.PrintDefaults()
	}
	// Accept the options before or after the file name.
	var files []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return ExitOK
			}
			return ExitUsage
		}
		if fs.NArg() == 0 {
			break
		}
		files = append(files, fs.Arg(0))
		args = fs.Args()[1:]
	}
	if len(files) != 1 {
		fs.Usage()
		return ExitUsage
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		fmt.Fprintln(stderr, "go81 basic list:", err)
		return ExitError
	}
	text, err := basic.ListP(data)
	if err != nil {
		fmt.Fprintln(stderr, "go81 basic list:", err)
		if text == "" {
			return ExitError
		}
	}
	if *outFile == "" {
		fmt.Fprint(stdout, text)
		return ExitOK
	}
	if err := os.WriteFile(*outFile, []byte(text), 0o644); err != nil {
		fmt.Fprintln(stderr, "go81 basic list:", err)
		return ExitError
	}
	return ExitOK
}
