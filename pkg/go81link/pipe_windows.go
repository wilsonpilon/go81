// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build windows

package go81link

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func dialPipe(name string) (net.Conn, error) {
	timeout := 5 * time.Second
	return winio.DialPipe(pipePath(name), &timeout)
}
