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
	"strings"

	"github.com/wilsonpilon/go81/internal/basic"
)

const basicUsage = `Uso: go81 basic <comando> [opções]

Ferramentas de BASIC do ZX81.

Comandos:
  list <programa.p> [-o arquivo.bas]   lista o programa como texto (formato do zxtext2p)
  make <texto.bas> [-o programa.p]     transforma o texto num programa .p (-run N, -1k)
`

func cmdBasic(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, basicUsage)
		return ExitOK
	}
	switch args[0] {
	case "list":
		return cmdBasicList(args[1:], stdout, stderr)
	case "make":
		return cmdBasicMake(args[1:], stdout, stderr)
	}
	fmt.Fprintf(stderr, "go81 basic: comando desconhecido: %s\n\n%s", args[0], basicUsage)
	return ExitUsage
}

func cmdBasicList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("basic list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outFile := fs.String("o", "", "grava a listagem neste arquivo (padrão: a tela)")
	fs.Usage = func() {
		fmt.Fprint(stderr, "Uso: go81 basic list [-o arquivo.txt] <programa.p|.o|.80>\n\nLista um programa .p (ZX81) ou .o/.80 (ZX80) como texto.\n\nOpções:\n")
		fs.PrintDefaults()
	}
	// Accept the options before or after the file name.
	files, err := parseInterspersed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
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
	listFn := basic.ListP
	if basic.IsZX80File(files[0]) {
		listFn = basic.ListZX80
	}
	text, err := listFn(data)
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

// parseInterspersed parses flags that may come before or after the file
// names and returns the names.
func parseInterspersed(fs *flag.FlagSet, args []string) ([]string, error) {
	var files []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		if fs.NArg() == 0 {
			break
		}
		files = append(files, fs.Arg(0))
		args = fs.Args()[1:]
	}
	return files, nil
}

func cmdBasicMake(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("basic make", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outFile := fs.String("o", "", "programa .p a gravar (padrão: o nome do texto com .p)")
	autorun := fs.Int("run", 0, "o programa roda sozinho a partir desta linha ao ser carregado")
	oneK := fs.Bool("1k", false, "tela compacta, para um ZX81 de 1K (padrão: tela cheia, 4K ou mais)")
	fs.Usage = func() {
		fmt.Fprint(stderr, `Uso: go81 basic make [opções] <texto.bas>

Transforma um programa BASIC escrito em texto (o formato do "go81 basic list",
compatível com o zxtext2p) num programa .p. Os números recebem o valor oculto
calculado pela própria ROM do ZX81, como ao digitar.

Opções:
`)
		fs.PrintDefaults()
	}
	files, err := parseInterspersed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if len(files) != 1 {
		fs.Usage()
		return ExitUsage
	}
	text, err := os.ReadFile(files[0])
	if err != nil {
		fmt.Fprintln(stderr, "go81 basic make:", err)
		return ExitError
	}
	p, err := basic.TextToP(string(text), basic.POptions{AutoRun: *autorun, Collapsed: *oneK})
	if err != nil {
		fmt.Fprintf(stderr, "go81 basic make: %s: %v\n", files[0], err)
		return ExitError
	}
	out := *outFile
	if out == "" {
		out = strings.TrimSuffix(files[0], filepath.Ext(files[0])) + ".p"
	}
	if err := os.WriteFile(out, p, 0o644); err != nil {
		fmt.Fprintln(stderr, "go81 basic make:", err)
		return ExitError
	}
	fmt.Fprintf(stdout, "%s gravado (%d bytes)\n", out, len(p))
	return ExitOK
}
