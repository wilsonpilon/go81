// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !windows

package link

import (
	"errors"

	"github.com/wilsonpilon/go81/internal/session"
)

// PipesSupported reports whether named pipes work on this system.
const PipesSupported = false

// ListenPipe fails: named pipes are a Windows feature. Use ListenUnix.
func ListenPipe(sess *session.Session, name string, opt ...Options) (*Server, error) {
	return nil, errors.New("pipe nomeado só existe no Windows; use um socket (link socket <arquivo>)")
}
