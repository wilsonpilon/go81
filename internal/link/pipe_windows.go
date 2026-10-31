// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build windows

package link

import (
	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"

	"github.com/wilsonpilon/go81/internal/session"
)

// PipesSupported reports whether named pipes work on this system.
const PipesSupported = true

// ListenPipe starts the Go81 text protocol on a Windows named pipe
// (name "go81" = \\.\pipe\go81). Only the current user (and SYSTEM) may
// connect.
func ListenPipe(sess *session.Session, name string) (*Server, error) {
	path := PipePath(name)
	sddl, err := userOnlySDDL()
	if err != nil {
		return nil, err
	}
	ln, err := winio.ListenPipe(path, &winio.PipeConfig{SecurityDescriptor: sddl})
	if err != nil {
		return nil, err
	}
	return serve(sess, ln, textHandler, "Go81 no pipe "+path), nil
}

// userOnlySDDL grants full access to the current user and to SYSTEM only.
func userOnlySDDL() (string, error) {
	tok := windows.GetCurrentProcessToken()
	u, err := tok.GetTokenUser()
	if err != nil {
		return "", err
	}
	return "D:P(A;;GA;;;SY)(A;;GA;;;" + u.User.Sid.String() + ")", nil
}
