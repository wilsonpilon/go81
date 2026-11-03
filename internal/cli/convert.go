// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/session"
)

// go81 txt2p and go81 p2txt: BASIC text (the zxtext2p format) to .p and
// back, for one file or many (the patterns are expanded here too, as the
// Windows shell does not do it).

// expandFiles expands the patterns (*.p, jogos/??.txt); a name without a
// match stays (and gives its own error).
func expandFiles(args []string) []string {
	var out []string
	for _, a := range args {
		if strings.ContainsAny(a, "*?[") {
			if m, err := filepath.Glob(a); err == nil && len(m) > 0 {
				out = append(out, m...)
				continue
			}
		}
		out = append(out, a)
	}
	return out
}

func cmdTxt2p(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("txt2p", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outFile := fs.String("o", "", "o .p a gravar (só com um texto; padrão: o nome do texto com .p)")
	autorun := fs.Int("run", 0, "o programa roda sozinho a partir desta linha ao ser carregado")
	oneK := fs.Bool("1k", false, "tela compacta, para um ZX81 de 1K (padrão: tela cheia)")
	fs.Usage = func() {
		io.WriteString(stderr, `Uso: go81 txt2p [opções] <texto.txt>...

Converte programas BASIC em texto (o formato do zxtext2p: \.' gráficos,
%A inverso, \" aspas, # libra) em programas .p, sem abrir o emulador. Aceita
vários arquivos e curingas: go81 txt2p *.txt

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
	files = expandFiles(files)
	if len(files) == 0 || *outFile != "" && len(files) > 1 {
		fs.Usage()
		return ExitUsage
	}
	code := ExitOK
	for _, f := range files {
		msg, err := session.TextToPFile(f, *outFile, basic.POptions{AutoRun: *autorun, Collapsed: *oneK})
		if err != nil {
			fmt.Fprintln(stderr, "go81 txt2p:", err)
			code = ExitError
			continue
		}
		fmt.Fprintln(stdout, msg)
	}
	return code
}

func cmdP2txt(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("p2txt", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outFile := fs.String("o", "", `o texto a gravar (só com um programa; "-" = a tela; padrão: o nome do programa com .txt)`)
	readable := fs.Bool("legivel", false, "para ler: palavras-chave dentro de REM e de textos por extenso (pode não voltar ao mesmo .p)")
	fs.Usage = func() {
		io.WriteString(stderr, `Uso: go81 p2txt [opções] <programa.p>...

Converte programas .p, .81, .p81 (ZX81) e .o, .80 (ZX80) em texto no formato do
zxtext2p, com os números de linha, os gráficos como \xx e o vídeo inverso como
%x, de modo que o go81 txt2p (ou o zxtext2p) os transforme de volta. Bytes sem
forma de texto (código de máquina num REM) viram \{XX}: o Go81 lê de volta, o
zxtext2p não (há um aviso). Com -legivel, as palavras-chave guardadas dentro
de REM e de textos saem por extenso (** em vez de \{D8}): fica mais fácil de
ler e de usar em outras ferramentas, mas pode não voltar ao mesmo .p. Aceita vários arquivos e curingas: go81 p2txt *.p

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
	files = expandFiles(files)
	if len(files) == 0 || *outFile != "" && len(files) > 1 {
		fs.Usage()
		return ExitUsage
	}
	code := ExitOK
	for _, f := range files {
		if *outFile == "-" {
			programText := session.ProgramText
			if *readable {
				programText = session.ProgramTextReadable
			}
			text, _, err := programText(f)
			fmt.Fprint(stdout, text)
			if err != nil {
				fmt.Fprintln(stderr, "go81 p2txt:", err)
				code = ExitError
			}
			continue
		}
		toText := session.PToTextFile
		if *readable {
			toText = session.PToTextFileReadable
		}
		msg, err := toText(f, *outFile)
		if err != nil {
			fmt.Fprintln(stderr, "go81 p2txt:", err)
			code = ExitError
			continue
		}
		fmt.Fprintln(stdout, msg)
	}
	return code
}
