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

	"github.com/wilsonpilon/go81/internal/link"
	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/ui/gui"
)

func cmdGUI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gui", flag.ContinueOnError)
	fs.SetOutput(stderr)
	mo := machineFlags(fs)
	var (
		zoom     = fs.Float64("zoom", 2, "tamanho da janela: 1 a 8 (também com fração: 2.5)")
		border   = fs.String("border", "normal", "borda: none, small, normal, large ou full")
		noAudio  = fs.Bool("no-audio", false, "não abre a placa de som")
		console  = fs.Bool("console", false, "abre também o console de comandos (F2)")
		linkOn   = fs.Bool("link", false, "liga a porta de comunicação: outros programas controlam este ZX81")
		port     = fs.Int("port", link.DefaultPort, "porta do protocolo Go81 (com -link)")
		zrcpPort = fs.Int("zrcp", link.DefaultZRCPPort, "porta ZRCP (com -link; -1 desliga)")
		exit     = fs.Int("exit-after", 0, "automação: fecha a janela depois de N quadros")
		capture  = fs.String("capture", "", "automação: com --exit-after, salva a imagem da janela neste PNG")
	)
	fs.Usage = func() {
		fmt.Fprint(stderr, `Uso: go81 gui [opções] [programa.p]

Abre o ZX81 numa janela. F2 abre o console de comandos.

Opções:
`)
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
	set, cfg, err := mo.load()
	if err != nil {
		fmt.Fprintln(stderr, "go81 gui:", err)
		return ExitUsage
	}
	*port, *zrcpPort = linkPorts(mo, cfg, port, zrcpPort)
	if !mo.given("zoom") && cfg.Window.Zoom != 0 {
		*zoom = cfg.Window.Zoom
	}
	if !mo.given("border") && cfg.Window.Border != "" {
		*border = cfg.Window.Border
	}
	b, ok := borders[strings.ToLower(*border)]
	if !ok {
		fmt.Fprintf(stderr, "go81 gui: borda inválida: %s\n", *border)
		return ExitUsage
	}
	if !session.ValidZoom(*zoom) {
		fmt.Fprintf(stderr, "go81 gui: zoom inválido: %g (1 a 8)\n", *zoom)
		return ExitUsage
	}
	sess, err := session.New(set)
	if err != nil {
		fmt.Fprintln(stderr, "go81 gui:", err)
		return ExitError
	}
	if fs.NArg() == 1 {
		data, err := os.ReadFile(fs.Arg(0))
		if err == nil {
			err = sess.InsertProgram(filepath.Base(fs.Arg(0)), data, true)
		}
		if err != nil {
			fmt.Fprintln(stderr, "go81 gui:", err)
			return ExitError
		}
	}
	ports := link.Register(sess)
	ports.Token = mo.linkToken(cfg)
	defer ports.Stop()
	if *linkOn {
		if err := ports.Start(sess, *port, *zrcpPort); err != nil {
			fmt.Fprintln(stderr, "go81 gui:", err)
			return ExitError
		}
		fmt.Fprintln(stdout, ports.Status())
	}
	opts := gui.Options{Session: sess, ConfigPath: cfg.Path, Zoom: *zoom, Border: b, NoSound: *noAudio,
		Console: *console, ExitAfter: *exit, CaptureFile: *capture}
	if err := gui.Run(opts); err != nil {
		fmt.Fprintln(stderr, "go81 gui:", err)
		return ExitError
	}
	return ExitOK
}
