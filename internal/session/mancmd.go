// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"io"

	"github.com/wilsonpilon/go81/internal/man"
)

func registerMan(r *Registry) {
	r.Register(&Command{Name: "man", Usage: "man [seção] nome | man -k palavra | man -l [seção]",
		Help: "o manual: comandos, BASIC do ZX81 e do ZX80, variáveis de sistema, ROM, erros", Run: cmdMan})
}

func cmdMan(s *Session, a []string, out io.Writer) error {
	text, err := man.Lookup(a, 78)
	if err != nil {
		return err
	}
	fmt.Fprint(out, text)
	return nil
}
