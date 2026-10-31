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
		romName    = fs.String("rom", "zx81.edition3", "ROM embutida (veja \"go81 roms\") ou caminho de um arquivo .rom")
		frames     = fs.Int("frames", 250, "quadros a emular (50 por segundo em PAL)")
		ram        = fs.Int("ram", 16, "RAM pack em K: 0 (só 1K interno), 2, 4, 8, 16, 32 ou 48")
		ntsc       = fs.Bool("ntsc", false, "ZX81 americano / TS1000 (60 Hz)")
		wrx        = fs.Bool("wrx", false, "alta resolução WRX")
		keys       = fs.String("keys", "", "teclas a digitar, separadas por vírgula; combinações com +, ex.: \"P,SHIFT+P,H,I,SHIFT+P,ENTER\"")
		keysStart  = fs.Int("keys-start", 150, "quadro em que a digitação começa (o ZX81 leva cerca de 2 s para ligar)")
		keyHold    = fs.Int("key-hold", 6, "quadros com cada tecla pressionada")
		keyGap     = fs.Int("key-gap", 6, "quadros entre teclas")
		enterGap   = fs.Int("enter-gap", 60, "quadros extras depois de ENTER (a ROM redesenha a tela)")
		screenshot = fs.String("screenshot", "", "salva a última tela neste arquivo (.png, .bmp, .pbm ou .txt)")
		border     = fs.String("border", "normal", "borda da imagem: none, small, normal, large ou full")
		tapeFile   = fs.String("tape", "", "programa a carregar da fita (.p, .81 ou .p81); também pode ser o último argumento")
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
	switch *ram {
	case 0, 1, 2, 4, 8, 16, 32, 48:
	default:
		fmt.Fprintf(stderr, "go81 run: tamanho de RAM inválido: %d\n", *ram)
		return ExitUsage
	}

	rom, err := loadROM(*romName)
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
		block, err := tape.ProgramBlock(*tapeFile, data)
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

func loadROM(name string) ([]byte, error) {
	if st, err := os.Stat(name); err == nil && !st.IsDir() {
		return os.ReadFile(name)
	}
	return assets.ROM(name)
}

func cmdROMs(stdout io.Writer) int {
	for _, n := range assets.ROMNames() {
		fmt.Fprintln(stdout, n)
	}
	return ExitOK
}
