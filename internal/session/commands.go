// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
	"github.com/wilsonpilon/go81/internal/version"
)

var borderNames = map[string]int{
	"none": zx81.BorderNone, "small": zx81.BorderSmall, "normal": zx81.BorderNormal,
	"large": zx81.BorderLarge, "full": zx81.BorderFull,
}

func registerBuiltins(r *Registry) {
	for _, c := range []*Command{
		{Name: "help", Aliases: []string{"ajuda", "?"}, Usage: "help [comando]", Help: "lista os comandos ou explica um deles", Run: cmdHelp},
		{Name: "version", Usage: "version", Help: "mostra a versão do Go81", Run: func(s *Session, a []string, out io.Writer) error {
			fmt.Fprintln(out, "go81", version.Full())
			return nil
		}},
		{Name: "machine", Usage: "machine [zx81|ts1000] [ram=K] [ntsc=on|off] [wrx=on|off] [sound=on|off] [rom=nome]", Help: "mostra a máquina ou troca a configuração (desliga e liga)", Run: cmdMachine},
		{Name: "reset", Usage: "reset", Help: "desliga e liga o ZX81 (a memória é apagada)", Run: func(s *Session, a []string, out io.Writer) error {
			s.powerCycle()
			return nil
		}},
		{Name: "status", Usage: "status", Help: "mostra quadro, PC, relógio, fita e teclas pendentes", Run: cmdStatus},
		{Name: "run", Aliases: []string{"wait"}, Usage: "run [quadros]", Help: "deixa o ZX81 rodar N quadros (padrão 50 = 1 s em PAL)", Run: cmdRun},
		{Name: "pause", Usage: "pause", Help: "pausa o relógio de tempo real", Run: func(s *Session, a []string, out io.Writer) error {
			s.paused = true
			return nil
		}},
		{Name: "continue", Aliases: []string{"cont"}, Usage: "continue", Help: "retoma o relógio de tempo real", Run: func(s *Session, a []string, out io.Writer) error {
			s.paused = false
			return nil
		}},
		{Name: "clock", Aliases: []string{"relogio"}, Usage: "clock [on|off]", Help: "liga ou desliga o relógio de tempo real (o ZX81 roda sozinho)", Run: cmdClock},
		{Name: "type", Raw: true, Usage: "type <texto>", Help: `digita um texto no ZX81 (\n = ENTER); espera terminar`, Run: func(s *Session, a []string, out io.Writer) error {
			return typeText(s, unescape(a[0]))
		}},
		{Name: "typeln", Raw: true, Usage: "typeln <texto>", Help: "como type, e tecla ENTER no fim", Run: func(s *Session, a []string, out io.Writer) error {
			return typeText(s, unescape(a[0])+"\n")
		}},
		{Name: "key", Usage: "key <tecla> [tecla…]", Help: "aperta teclas do ZX81 em sequência, ex.: key SHIFT+F ENTER", Run: cmdKey},
		{Name: "peek", Usage: "peek <endereço> [quantidade]", Help: "lê a memória (como o Z80 vê)", Run: cmdPeek},
		{Name: "poke", Usage: "poke <endereço> <valor> [valor…]", Help: "escreve na memória (a ROM continua protegida)", Run: cmdPoke},
		{Name: "regs", Aliases: []string{"registers"}, Usage: "regs", Help: "mostra os registradores do Z80", Run: cmdRegs},
		{Name: "screen", Usage: "screen", Help: "mostra a tela do ZX81 como texto (32×24)", Run: cmdScreen},
		{Name: "screenshot", Usage: "screenshot <arquivo.png|.bmp|.pbm|.txt> [none|small|normal|large|full]", Help: "salva a tela em PNG, BMP, PBM ou texto (pela extensão)", Run: cmdScreenshot},
		{Name: "load", Usage: "load <programa.p|.bas> [noauto]", Help: `carrega um programa (.p, .81, .p81 ou BASIC em texto .bas/.txt) pela fita`, Run: cmdLoad},
		{Name: "save", Usage: "save <programa.p> [NOME]", Help: `executa SAVE "NOME" no ZX81 e grava o programa no arquivo`, Run: cmdSave},
		{Name: "list", Aliases: []string{"listar"}, Usage: "list [programa.p]", Help: "lista o programa BASIC (do ZX81 ou de um arquivo .p) como texto", Run: cmdList},
		{Name: "make", Usage: "make <texto.bas> [programa.p] [run=N] [1k]", Help: "transforma um programa BASIC em texto num .p", Run: cmdMake},
		{Name: "script", Aliases: []string{"source"}, Usage: "script <arquivo.g81>", Help: "executa os comandos de um arquivo (# = comentário)", Run: cmdScript},
		{Name: "echo", Raw: true, Usage: "echo <texto>", Help: "escreve um texto", Run: func(s *Session, a []string, out io.Writer) error {
			fmt.Fprintln(out, a[0])
			return nil
		}},
		{Name: "sleep", Usage: "sleep <milissegundos>", Help: "espera um tempo (útil com o relógio de tempo real)", Run: cmdSleep},
		{Name: "pwd", Usage: "pwd", Help: "mostra a pasta atual", Run: func(s *Session, a []string, out io.Writer) error {
			d, err := os.Getwd()
			fmt.Fprintln(out, d)
			return err
		}},
		{Name: "cd", Usage: "cd <pasta>", Help: "muda a pasta atual", Run: func(s *Session, a []string, out io.Writer) error {
			if len(a) != 1 {
				return usageErr("cd <pasta>")
			}
			return os.Chdir(a[0])
		}},
		{Name: "ls", Aliases: []string{"dir"}, Usage: "ls [pasta]", Help: "lista os arquivos", Run: cmdLs},
		{Name: "roms", Usage: "roms", Help: "lista as ROMs embutidas", Run: func(s *Session, a []string, out io.Writer) error {
			for _, n := range assets.ROMNames() {
				fmt.Fprintln(out, n)
			}
			return nil
		}},
		{Name: "quit", Aliases: []string{"exit", "sair"}, Usage: "quit", Help: "sai", Run: func(s *Session, a []string, out io.Writer) error {
			return ErrQuit
		}},
	} {
		r.Register(c)
	}
}

func usageErr(u string) error { return fmt.Errorf("uso: %s", u) }

func cmdHelp(s *Session, a []string, out io.Writer) error {
	if len(a) == 1 {
		c := s.reg.Lookup(a[0])
		if c == nil {
			return fmt.Errorf("comando desconhecido: %s", a[0])
		}
		fmt.Fprintf(out, "%s\n  %s\n", c.Usage, c.Help)
		if len(c.Aliases) > 0 {
			fmt.Fprintf(out, "  também: %s\n", strings.Join(c.Aliases, ", "))
		}
		return nil
	}
	fmt.Fprintln(out, "Comandos (help <comando> para detalhes; números: 16384, $4000, 4000h, 0x4000, 1010%):")
	for _, c := range s.reg.Commands() {
		fmt.Fprintf(out, "  %-12s %s\n", c.Name, c.Help)
	}
	return nil
}

func (s *Session) description() string {
	set := s.set
	ram := set.RAMK
	if ram == 0 {
		ram = 1
	}
	tv := "PAL"
	if set.NTSC {
		tv = "NTSC"
	}
	return fmt.Sprintf("ZX81 %dK %s · ROM %s · WRX %s · som %s", ram, tv, set.ROM, onOff(set.WRX), onOff(set.Sound))
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func cmdMachine(s *Session, a []string, out io.Writer) error {
	if len(a) == 0 {
		fmt.Fprintln(out, s.description())
		return nil
	}
	set := s.set
	for _, arg := range a {
		k, v, hasV := strings.Cut(arg, "=")
		var err error
		switch strings.ToLower(k) {
		case "zx81":
			set.NTSC = false
		case "ts1000":
			set.NTSC, set.RAMK = true, 2
		case "ram":
			var n int
			if n, err = ParseNumber(v); err == nil {
				switch n {
				case 0, 1:
					set.RAMK = 0
				case 2, 4, 8, 16, 32, 48:
					set.RAMK = n
				default:
					err = fmt.Errorf("RAM inválida: %d (0, 2, 4, 8, 16, 32 ou 48)", n)
				}
			}
		case "ntsc":
			set.NTSC, err = parseOnOff(v)
		case "wrx":
			set.WRX, err = parseOnOff(v)
		case "sound", "som":
			set.Sound, err = parseOnOff(v)
		case "rom":
			if !hasV || v == "" {
				err = fmt.Errorf("rom=<nome>")
			}
			set.ROM = v
		default:
			err = fmt.Errorf("opção desconhecida: %s", arg)
		}
		if err != nil {
			return err
		}
	}
	if err := s.configure(set); err != nil {
		return err
	}
	fmt.Fprintln(out, s.description())
	return nil
}

func cmdStatus(s *Session, a []string, out io.Writer) error {
	clock := "parado (o ZX81 só roda com run, type, load…)"
	if s.live {
		clock = "tempo real"
		if s.paused {
			clock = "tempo real, PAUSADO"
		}
	}
	fita := "parada"
	if s.m.Tape != nil && s.m.Tape.Player.Playing {
		fita = "tocando"
	}
	fmt.Fprintf(out, "%s\nquadro %d · PC=%04X · relógio: %s · fita: %s · teclas pendentes: %d\n",
		s.description(), s.frames, s.m.CPU.PC, clock, fita, len(s.queue))
	return nil
}

// cmdClock runs with s.mu held; StartClock/StopClock lock it themselves.
func cmdClock(s *Session, a []string, out io.Writer) error {
	if len(a) == 1 {
		on, err := parseOnOff(a[0])
		if err != nil {
			return err
		}
		if s.driven {
			return fmt.Errorf("aqui quem dá o ritmo é a janela/terminal; use pause e continue")
		}
		if on {
			s.Unlocked(s.StartClock)
		} else {
			s.Unlocked(s.StopClock)
		}
	} else if len(a) > 1 {
		return usageErr("clock [on|off]")
	}
	if s.driven {
		fmt.Fprintln(out, "relógio: tempo real (da janela/terminal)")
	} else if s.live {
		fmt.Fprintln(out, "relógio: tempo real")
	} else {
		fmt.Fprintln(out, "relógio: parado (o ZX81 só roda com run, type, load…)")
	}
	return nil
}

func cmdRun(s *Session, a []string, out io.Writer) error {
	n := 50
	if len(a) > 0 {
		v, err := ParseNumber(a[0])
		if err != nil || v < 0 {
			return usageErr("run [quadros]")
		}
		n = v
	}
	target := s.frames + int64(n)
	s.waitUntil(func() bool { return s.frames >= target }, n+1)
	return nil
}

func unescape(t string) string {
	return strings.NewReplacer(`\n`, "\n", `\\`, `\`).Replace(t)
}

func typeText(s *Session, text string) error {
	s.waitBoot()
	if err := s.queueText(text, KeyHold); err != nil {
		return err
	}
	s.waitUntil(func() bool { return !s.keysPending() }, 100000)
	return nil
}

func cmdKey(s *Session, a []string, out io.Writer) error {
	if len(a) == 0 {
		return usageErr("key <tecla> [tecla…]")
	}
	s.waitBoot()
	for _, k := range a {
		if err := s.queueChord(k, KeyHold); err != nil {
			return err
		}
	}
	s.waitUntil(func() bool { return !s.keysPending() }, 100000)
	return nil
}

func cmdPeek(s *Session, a []string, out io.Writer) error {
	if len(a) < 1 || len(a) > 2 {
		return usageErr("peek <endereço> [quantidade]")
	}
	addr, err := parseAddr(a[0])
	if err != nil {
		return err
	}
	n := 1
	if len(a) == 2 {
		if n, err = ParseNumber(a[1]); err != nil || n < 1 || n > 65536 {
			return fmt.Errorf("quantidade inválida: %s", a[1])
		}
	}
	if n == 1 {
		v := s.m.Peek(addr)
		fmt.Fprintf(out, "%04X: %02X (%d)\n", addr, v, v)
		return nil
	}
	for i := 0; i < n; i += 16 {
		var hex, txt strings.Builder
		for j := 0; j < 16 && i+j < n; j++ {
			b := s.m.Peek(addr + uint16(i+j))
			fmt.Fprintf(&hex, " %02X", b)
			if b&64 == 0 {
				txt.WriteRune(zx81.TextChars[b&63])
			} else {
				txt.WriteByte('.')
			}
		}
		fmt.Fprintf(out, "%04X:%-48s  %s\n", addr+uint16(i), hex.String(), txt.String())
	}
	return nil
}

func cmdPoke(s *Session, a []string, out io.Writer) error {
	if len(a) < 2 {
		return usageErr("poke <endereço> <valor> [valor…]")
	}
	addr, err := parseAddr(a[0])
	if err != nil {
		return err
	}
	for i, t := range a[1:] {
		v, err := ParseNumber(t)
		if err != nil || v < 0 || v > 255 {
			return fmt.Errorf("valor inválido (0..255): %s", t)
		}
		s.m.Poke(addr+uint16(i), byte(v))
	}
	return nil
}

// FlagString writes F as letters (SZ5H3PNC, '-' when clear), as ZEsarUX does.
func FlagString(f byte) string {
	const names = "SZ5H3PNC"
	b := []byte("--------")
	for i := 0; i < 8; i++ {
		if f&(0x80>>i) != 0 {
			b[i] = names[i]
		}
	}
	return string(b)
}

func cmdRegs(s *Session, a []string, out io.Writer) error {
	c := s.m.CPU
	fmt.Fprintf(out, "PC=%04X SP=%04X AF=%04X BC=%04X DE=%04X HL=%04X IX=%04X IY=%04X\n",
		c.PC, c.SP, c.AF(), c.BC(), c.DE(), c.HL(), c.IX, c.IY)
	halted := ""
	if c.Halted {
		halted = " HALT"
	}
	fmt.Fprintf(out, "AF'=%04X BC'=%04X DE'=%04X HL'=%04X I=%02X R=%02X IM%d IFF1=%d IFF2=%d F=%s%s\n",
		c.AF_, c.BC_, c.DE_, c.HL_, c.I, c.RegR(), c.IM, c.IFF1, c.IFF2, FlagString(c.F), halted)
	return nil
}

func cmdScreen(s *Session, a []string, out io.Writer) error {
	line := "+" + strings.Repeat("-", zx81.TextCols) + "+"
	fmt.Fprintln(out, line)
	for _, l := range s.m.ScreenText() {
		fmt.Fprintf(out, "|%s|\n", l)
	}
	fmt.Fprintln(out, line)
	return nil
}

func cmdScreenshot(s *Session, a []string, out io.Writer) error {
	if len(a) < 1 || len(a) > 2 {
		return usageErr("screenshot <arquivo.png|.bmp|.pbm|.txt> [borda]")
	}
	b := zx81.BorderNormal
	if len(a) == 2 {
		v, ok := borderNames[strings.ToLower(a[1])]
		if !ok {
			return fmt.Errorf("borda inválida: %s", a[1])
		}
		b = v
	}
	if err := SaveScreen(a[0], s.m, b); err != nil {
		return err
	}
	fmt.Fprintln(out, "tela salva em", a[0])
	return nil
}

func cmdLoad(s *Session, a []string, out io.Writer) error {
	if len(a) < 1 || len(a) > 2 {
		return usageErr("load <programa.p|.bas> [noauto]")
	}
	auto := !(len(a) == 2 && strings.EqualFold(a[1], "noauto"))
	data, err := os.ReadFile(a[0])
	if err != nil {
		return err
	}
	name, data, err := ProgramBytes(a[0], data) // a .bas becomes a .p
	if err != nil {
		return err
	}
	block, err := tape.ProgramBlock(name, data)
	if err != nil {
		return err
	}
	s.queue, s.hold, s.gap = nil, 0, 0
	s.m.InsertTape(block, auto)
	if !auto {
		fmt.Fprintln(out, `fita inserida; digite LOAD "" no ZX81`)
		return nil
	}
	p := s.m.Tape.Player
	started := false
	done := s.waitUntil(func() bool {
		if p.Playing {
			started = true
		}
		return started && !p.Playing && !s.m.AutoLoading()
	}, 3000)
	if !done {
		return fmt.Errorf("o programa não terminou de carregar")
	}
	end := s.frames + 60 // the ROM redraws the screen after LOAD
	s.waitUntil(func() bool { return s.frames >= end }, 61)
	fmt.Fprintln(out, "programa carregado:", filepath.Base(a[0]))
	return nil
}

func cmdSave(s *Session, a []string, out io.Writer) error {
	if len(a) < 1 || len(a) > 2 {
		return usageErr("save <programa.p> [NOME]")
	}
	name := strings.TrimSuffix(filepath.Base(a[0]), filepath.Ext(a[0]))
	if len(a) == 2 {
		name = a[1]
	}
	name = strings.ToUpper(name)
	for _, r := range name {
		if zx81.ChordForRune(r) == "" || r == '"' {
			return fmt.Errorf("o nome %q tem caracteres que o ZX81 não aceita num SAVE", name)
		}
	}
	s.lastSave = nil
	s.saveCmd = true
	defer func() { s.saveCmd = false }()
	s.waitBoot()
	// S is SAVE at the K cursor; the name goes between quotes.
	if err := s.queueChord("S", KeyHold); err != nil {
		return err
	}
	if err := s.queueText(`"`+name+`"`+"\n", KeyHold); err != nil {
		return err
	}
	if !s.waitUntil(func() bool { return s.lastSave != nil }, 5000) {
		return errors.New(`o SAVE não aconteceu (o ZX81 estava esperando um comando, com o cursor K?)`)
	}
	if err := os.WriteFile(a[0], s.lastSave.prog, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "SAVE \"%s\" gravado em %s (%d bytes)\n", s.lastSave.name, a[0], len(s.lastSave.prog))
	return nil
}

func cmdList(s *Session, a []string, out io.Writer) error {
	var prog []byte
	switch len(a) {
	case 0:
		prog = basic.ProgramFromMemory(s.m)
	case 1:
		data, err := os.ReadFile(a[0])
		if err != nil {
			return err
		}
		if prog, err = basic.ProgramFromP(data); err != nil {
			return err
		}
	default:
		return usageErr("list [programa.p]")
	}
	lines, err := basic.Lines(prog)
	if len(lines) == 0 && err == nil {
		fmt.Fprintln(out, "(nenhum programa)")
		return nil
	}
	fmt.Fprint(out, basic.List(lines))
	return err
}

func cmdMake(s *Session, a []string, out io.Writer) error {
	if len(a) < 1 {
		return usageErr("make <texto.bas> [programa.p] [run=N] [1k]")
	}
	var opt basic.POptions
	dest := ""
	for _, arg := range a[1:] {
		switch k, v, ok := strings.Cut(arg, "="); {
		case ok && strings.EqualFold(k, "run"):
			n, err := ParseNumber(v)
			if err != nil || n < 0 {
				return fmt.Errorf("run=<linha>")
			}
			opt.AutoRun = n
		case strings.EqualFold(arg, "1k"):
			opt.Collapsed = true
		case dest == "":
			dest = arg
		default:
			return usageErr("make <texto.bas> [programa.p] [run=N] [1k]")
		}
	}
	if dest == "" {
		dest = strings.TrimSuffix(a[0], filepath.Ext(a[0])) + ".p"
	}
	text, err := os.ReadFile(a[0])
	if err != nil {
		return err
	}
	p, err := basic.TextToP(string(text), opt)
	if err != nil {
		return fmt.Errorf("%s: %v", a[0], err)
	}
	if err := os.WriteFile(dest, p, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s gravado (%d bytes)\n", dest, len(p))
	return nil
}

func cmdScript(s *Session, a []string, out io.Writer) error {
	if len(a) != 1 {
		return usageErr("script <arquivo.g81>")
	}
	f, err := os.Open(a[0])
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := s.execLocked(line, out); err != nil {
			if errors.Is(err, ErrQuit) {
				return err
			}
			return fmt.Errorf("%s:%d: %v", a[0], n, err)
		}
	}
	return sc.Err()
}

func cmdSleep(s *Session, a []string, out io.Writer) error {
	if len(a) != 1 {
		return usageErr("sleep <milissegundos>")
	}
	ms, err := ParseNumber(a[0])
	if err != nil || ms < 0 {
		return usageErr("sleep <milissegundos>")
	}
	s.mu.Unlock()
	time.Sleep(time.Duration(ms) * time.Millisecond)
	s.mu.Lock()
	return nil
}

func cmdLs(s *Session, a []string, out io.Writer) error {
	dir := "."
	if len(a) > 0 {
		dir = a[0]
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range ents {
		if e.IsDir() {
			fmt.Fprintf(out, "%10s  %s/\n", "", e.Name())
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		fmt.Fprintf(out, "%10d  %s\n", info.Size(), e.Name())
	}
	return nil
}
