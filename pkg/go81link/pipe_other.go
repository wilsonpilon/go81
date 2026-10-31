// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !windows

package go81link

import (
	"errors"
	"net"
)

func dialPipe(name string) (net.Conn, error) {
	return nil, errors.New("go81link: pipe nomeado só existe no Windows; use unix:<arquivo>")
}
