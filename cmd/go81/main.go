// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Comando go81: emulador e ambiente de desenvolvimento para ZX81.
package main

import (
	"os"

	"github.com/wilsonpilon/go81/internal/cli"
)

func main() {
	os.Exit(cli.Run())
}
