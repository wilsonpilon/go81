// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/ui/tui"
)

func cmdTUI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		romName = fs.String("rom", "zx81.edition3", "ROM embutida (veja \"go81 roms\") ou caminho de um arquivo .rom")
		ram     = fs.Int("ram", 16, "RAM pack em K: 0 (só 1K interno), 2, 4, 8, 16, 32 ou 48")
		ntsc    = fs.Bool("ntsc", false, "ZX81 americano / TS1000 (60 Hz)")
		wrx     = fs.Bool("wrx", false, "alta resolução WRX")
		exit    = fs.Int("exit-after", 0, "automação: sai depois de N quadros e imprime a tela de texto")
	)
	fs.Usage = func() {
		fmt.Fprint(stderr, "Uso: go81 tui [opções] [programa.p]\n\nAbre o ZX81 no terminal. F1 mostra a ajuda, F10 sai.\n\nOpções:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "go81 tui: só um programa por vez")
		return ExitUsage
	}
	switch *ram {
	case 0, 1, 2, 4, 8, 16, 32, 48:
	default:
		fmt.Fprintf(stderr, "go81 tui: tamanho de RAM inválido: %d\n", *ram)
		return ExitUsage
	}
	rom, err := loadROM(*romName)
	if err != nil {
		fmt.Fprintln(stderr, "go81 tui:", err)
		return ExitError
	}
	cfg := zx81.DefaultConfig(rom)
	if *ram > 1 {
		cfg.RAMTop = zx81.RAMTopForPack(*ram)
	} else {
		cfg.RAMTop = zx81.RAMTopForPack(0)
	}
	cfg.NTSC = *ntsc
	if *wrx {
		cfg.HiRes = zx81.HiResWRX
	}
	desc := fmt.Sprintf("ZX81 %dK", max(*ram, 1))
	if *ntsc {
		desc += " NTSC"
	} else {
		desc += " PAL"
	}
	opts := tui.Options{Config: cfg, Description: desc, ExitAfter: *exit}
	if *exit > 0 {
		opts.OnExit = func(s tui.Screen) { fmt.Fprint(stdout, s.String()) }
	}
	if fs.NArg() == 1 {
		data, err := os.ReadFile(fs.Arg(0))
		if err != nil {
			fmt.Fprintln(stderr, "go81 tui:", err)
			return ExitError
		}
		opts.TapeName, opts.TapeData = filepath.Base(fs.Arg(0)), data
	}
	if err := tui.Run(opts); err != nil {
		fmt.Fprintln(stderr, "go81 tui:", err)
		return ExitError
	}
	return ExitOK
}
