// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build nogui

package gui

import (
	"errors"

	"github.com/wilsonpilon/go81/internal/session"
)

// Available reports whether this build has the graphical interface.
const Available = false

// Options configure the window.
type Options struct {
	Session     *session.Session
	ConfigPath  string
	Zoom        int
	Border      int
	NoSound     bool
	Console     bool
	ExitAfter   int
	CaptureFile string
}

// Run fails: this executable was built with -tags nogui.
func Run(opts Options) error {
	return errors.New("este executável foi compilado sem interface gráfica (-tags nogui)")
}
