// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"os"
	"os/exec"
	"runtime"
)

// openInEditor opens a text file: $EDITOR if set, else the system's
// program for it.
func openInEditor(path string) error {
	var cmd *exec.Cmd
	switch {
	case os.Getenv("EDITOR") != "":
		cmd = exec.Command(os.Getenv("EDITOR"), path)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		return cmd.Run()
	case runtime.GOOS == "windows":
		cmd = exec.Command("notepad", path)
	case runtime.GOOS == "darwin":
		cmd = exec.Command("open", "-t", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
