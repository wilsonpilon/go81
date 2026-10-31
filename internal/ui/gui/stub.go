// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build nogui

package gui

import (
	"errors"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// Available reports whether this build has the graphical interface.
const Available = false

// Options configure the window.
type Options struct {
	Config      zx81.Config
	TapeName    string
	TapeData    []byte
	Zoom        int
	Border      int
	Description string
	ExitAfter   int
	CaptureFile string
}

// Run fails: this executable was built with -tags nogui.
func Run(opts Options) error {
	return errors.New("este executável foi compilado sem interface gráfica (-tags nogui)")
}
