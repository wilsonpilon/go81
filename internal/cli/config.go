// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/wilsonpilon/go81/internal/session"
)

const configUsage = `Uso: go81 config <comando>

Arquivo de configuração: a máquina padrão, a janela, o Link e os perfis.

Comandos:
  path    mostra onde fica o arquivo
  init    cria o arquivo com um exemplo comentado (não sobrescreve)
  show    mostra a máquina que vale agora e os perfis
  edit    abre o arquivo no editor do sistema
`

func cmdConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, configUsage)
		return ExitOK
	}
	path := session.ConfigPath()
	switch args[0] {
	case "path":
		fmt.Fprintln(stdout, path)
	case "init":
		if _, err := os.Stat(path); err == nil {
			fmt.Fprintf(stderr, "go81 config: %s já existe\n", path)
			return ExitError
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fmt.Fprintln(stderr, "go81 config:", err)
			return ExitError
		}
		if err := os.WriteFile(path, []byte(session.ConfigTemplate), 0o644); err != nil {
			fmt.Fprintln(stderr, "go81 config:", err)
			return ExitError
		}
		fmt.Fprintln(stdout, "criado:", path)
	case "show":
		cfg, err := session.LoadConfig(path)
		if err != nil {
			fmt.Fprintln(stderr, "go81 config:", err)
			return ExitError
		}
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(stdout, "%s (não existe: valem os padrões; crie com \"go81 config init\")\n", path)
		} else {
			fmt.Fprintln(stdout, path)
		}
		set, err := cfg.MachineSettings("")
		if err != nil {
			fmt.Fprintln(stderr, "go81 config:", err)
			return ExitError
		}
		fmt.Fprintln(stdout, "máquina:", describe(set))
		for _, p := range cfg.ProfileNames() {
			if s, err := cfg.MachineSettings(p); err == nil {
				fmt.Fprintf(stdout, "perfil %-12s %s\n", p+":", describe(s))
			}
		}
	case "edit":
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			os.MkdirAll(filepath.Dir(path), 0o755)
			os.WriteFile(path, []byte(session.ConfigTemplate), 0o644)
		}
		if err := openInEditor(path); err != nil {
			fmt.Fprintln(stderr, "go81 config:", err)
			return ExitError
		}
	default:
		fmt.Fprintf(stderr, "go81 config: comando desconhecido: %s\n\n%s", args[0], configUsage)
		return ExitUsage
	}
	return ExitOK
}

func describe(s session.Settings) string { return s.Describe() }
