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
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/ui/gui"
)

func cmdGUI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gui", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		romName = fs.String("rom", "zx81.edition3", "ROM embutida (veja \"go81 roms\") ou caminho de um arquivo .rom")
		ram     = fs.Int("ram", 16, "RAM pack em K: 0 (só 1K interno), 2, 4, 8, 16, 32 ou 48")
		ntsc    = fs.Bool("ntsc", false, "ZX81 americano / TS1000 (60 Hz)")
		wrx     = fs.Bool("wrx", false, "alta resolução WRX")
		zoom    = fs.Int("zoom", 2, "tamanho da janela: 1 a 4")
		border  = fs.String("border", "normal", "borda: none, small, normal, large ou full")
		exit    = fs.Int("exit-after", 0, "automação: fecha a janela depois de N quadros")
		capture = fs.String("capture", "", "automação: com --exit-after, salva a imagem da janela neste PNG")
	)
	fs.Usage = func() {
		fmt.Fprint(stderr, "Uso: go81 gui [opções] [programa.p]\n\nAbre o ZX81 numa janela.\n\nOpções:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if !gui.Available {
		fmt.Fprintln(stderr, "go81 gui: este executável foi compilado sem interface gráfica (-tags nogui)")
		return ExitError
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "go81 gui: só um programa por vez")
		return ExitUsage
	}
	b, ok := borders[strings.ToLower(*border)]
	if !ok {
		fmt.Fprintf(stderr, "go81 gui: borda inválida: %s\n", *border)
		return ExitUsage
	}
	if *zoom < 1 || *zoom > 4 {
		fmt.Fprintf(stderr, "go81 gui: zoom inválido: %d\n", *zoom)
		return ExitUsage
	}
	switch *ram {
	case 0, 1, 2, 4, 8, 16, 32, 48:
	default:
		fmt.Fprintf(stderr, "go81 gui: tamanho de RAM inválido: %d\n", *ram)
		return ExitUsage
	}
	rom, err := loadROM(*romName)
	if err != nil {
		fmt.Fprintln(stderr, "go81 gui:", err)
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
	opts := gui.Options{Config: cfg, Zoom: *zoom, Border: b, Description: desc, ExitAfter: *exit, CaptureFile: *capture}
	if fs.NArg() == 1 {
		data, err := os.ReadFile(fs.Arg(0))
		if err != nil {
			fmt.Fprintln(stderr, "go81 gui:", err)
			return ExitError
		}
		opts.TapeName, opts.TapeData = filepath.Base(fs.Arg(0)), data
	}
	if err := gui.Run(opts); err != nil {
		fmt.Fprintln(stderr, "go81 gui:", err)
		return ExitError
	}
	return ExitOK
}
