// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
	"github.com/wilsonpilon/go81/internal/session"
)

var borders = map[string]int{
	"none": zx81.BorderNone, "small": zx81.BorderSmall, "normal": zx81.BorderNormal,
	"large": zx81.BorderLarge, "full": zx81.BorderFull,
}

func cmdRun(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		model      = fs.String("model", "zx81", "modelo: zx81, ts1000, ts1500, tk85 ou r470 (traz a ROM, a RAM e o NTSC dele)")
		romName    = fs.String("rom", "zx81.edition3", "ROM embutida (veja \"go81 roms\") ou caminho de um arquivo .rom")
		lowRAM     = fs.Bool("lowram", false, "RAM em 8K-16K")
		m1not      = fs.Bool("m1not", false, "modificação M1NOT (código em 32K-48K)")
		fpfix      = fs.Bool("fpfix", false, "correção do ponto flutuante (ROM edição 1)")
		frames     = fs.Int("frames", 250, "quadros a emular (50 por segundo em PAL)")
		ram        = fs.Int("ram", 16, "RAM pack em K: 0 (nenhum), 2, 4, 8, 16, 32 ou 48 (soma aos 16K do TS1500, TK85 e R-470)")
		ntsc       = fs.Bool("ntsc", false, "ZX81 americano / TS1000 (60 Hz)")
		wrx        = fs.Bool("wrx", false, "alta resolução WRX")
		keys       = fs.String("keys", "", "teclas a digitar, separadas por vírgula; combinações com +, ex.: \"P,SHIFT+P,H,I,SHIFT+P,ENTER\"")
		keysStart  = fs.Int("keys-start", 150, "quadro em que a digitação começa (o ZX81 leva cerca de 2 s para ligar)")
		keyHold    = fs.Int("key-hold", 6, "quadros com cada tecla pressionada")
		keyGap     = fs.Int("key-gap", 6, "quadros entre teclas")
		enterGap   = fs.Int("enter-gap", 60, "quadros extras depois de ENTER (a ROM redesenha a tela)")
		screenshot = fs.String("screenshot", "", "salva a última tela neste arquivo (.png, .bmp, .pbm ou .txt)")
		border     = fs.String("border", "normal", "borda da imagem: none, small, normal, large ou full")
		tapeFile   = fs.String("tape", "", "programa a carregar da fita (.p, .81, .p81 ou BASIC em texto .bas/.txt); também pode ser o último argumento")
		saveFile   = fs.String("save", "", "grava num arquivo .p o que o programa salvar com SAVE")
		autoLoad   = fs.Bool("autoload", true, "com um programa: reinicia e digita LOAD \"\" como o EightyOne")
	)
	fs.Usage = func() {
		fmt.Fprint(stderr, "Uso: go81 run [opções] [programa.p]\n\nRoda o ZX81 sem janela.\n\nOpções:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "go81 run: só um programa por vez")
		return ExitUsage
	}
	if fs.NArg() == 1 {
		*tapeFile = fs.Arg(0)
	}
	b, ok := borders[strings.ToLower(*border)]
	if !ok {
		fmt.Fprintf(stderr, "go81 run: borda inválida: %s\n", *border)
		return ExitUsage
	}
	// The model first (it brings its ROM, RAM and NTSC); the options typed
	// win over it.
	given := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { given[f.Name] = true })
	set := session.DefaultSettings()
	if given["model"] {
		var err error
		if set, err = set.WithModel(*model); err != nil {
			fmt.Fprintln(stderr, "go81 run:", err)
			return ExitUsage
		}
	}
	if given["rom"] || !given["model"] {
		set.ROM = *romName
	}
	if given["ram"] || !given["model"] {
		var err error
		if set.RAMK, err = session.CheckPack(*ram); err != nil {
			fmt.Fprintln(stderr, "go81 run:", err)
			return ExitUsage
		}
	}
	if given["ntsc"] || !given["model"] {
		set.NTSC = *ntsc
	}
	set.WRX, set.LowRAM, set.M1Not, set.FPFix = *wrx, *lowRAM, *m1not, *fpfix
	cfg, err := session.EmuConfig(set)
	if err != nil {
		fmt.Fprintln(stderr, "go81 run:", err)
		return ExitError
	}

	type press struct {
		frame int
		rows  [8]byte
	}
	var presses []press
	if *keys != "" {
		f := *keysStart
		for _, chord := range strings.Split(*keys, ",") {
			rows, err := zx81.ParseChord(chord)
			if err != nil {
				fmt.Fprintln(stderr, "go81 run:", err)
				return ExitUsage
			}
			presses = append(presses, press{f, rows}, press{f + *keyHold, [8]byte{}})
			f += *keyHold + *keyGap
			if strings.EqualFold(strings.TrimSpace(chord), "ENTER") {
				f += *enterGap
			}
		}
		if f > *frames {
			fmt.Fprintf(stderr, "go81 run: aviso: a digitação termina no quadro %d, depois de --frames %d\n", f, *frames)
		}
	}

	m, err := zx81.New(cfg)
	if err != nil {
		fmt.Fprintln(stderr, "go81 run:", err)
		return ExitError
	}

	if *tapeFile != "" {
		data, err := os.ReadFile(*tapeFile)
		if err != nil {
			fmt.Fprintln(stderr, "go81 run:", err)
			return ExitError
		}
		name, data, err := session.ProgramBytes(*tapeFile, data)
		if err != nil {
			fmt.Fprintln(stderr, "go81 run:", err)
			return ExitError
		}
		block, err := tape.ProgramBlock(name, data)
		if err != nil {
			fmt.Fprintln(stderr, "go81 run:", err)
			return ExitError
		}
		m.InsertTape(block, *autoLoad)
	}

	var saved []byte
	if *saveFile != "" {
		if m.Tape == nil {
			m.Tape = tape.NewDeck(3250000)
		}
		m.Tape.Out = true
		m.SaveByte = func(b byte) { saved = append(saved, b) }
	}

	next := 0
	for f := 0; f < *frames; f++ {
		for next < len(presses) && presses[next].frame <= f {
			m.SetKeys(presses[next].rows)
			next++
		}
		m.RunFrame()
	}

	c := m.CPU
	fmt.Fprintf(stdout, "go81 run: %d quadros, PC=%04X, quadros de TV completos: %d, linhas por quadro: %d\n",
		*frames, c.PC, m.FramesCompleted(), m.ScanlinesPerFrame)

	if *saveFile != "" {
		_, prog, ok := tape.SplitSaved(saved)
		if !ok {
			fmt.Fprintln(stderr, "go81 run: o programa não executou SAVE; nada gravado")
			return ExitError
		}
		if err := os.WriteFile(*saveFile, prog, 0o644); err != nil {
			fmt.Fprintln(stderr, "go81 run:", err)
			return ExitError
		}
		fmt.Fprintf(stdout, "go81 run: SAVE gravado em %s (%d bytes)\n", *saveFile, len(prog))
	}

	if *screenshot != "" {
		if err := session.SaveScreen(*screenshot, m, b); err != nil {
			fmt.Fprintln(stderr, "go81 run:", err)
			return ExitError
		}
		fmt.Fprintln(stdout, "go81 run: tela salva em", *screenshot)
	}
	return ExitOK
}

func cmdROMs(stdout io.Writer) int {
	for _, n := range assets.ROMNames() {
		fmt.Fprintln(stdout, n)
	}
	return ExitOK
}
