// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
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
		{Name: "machine", Usage: "machine [perfil=nome] [zx81|ts1000|ts1500|tk85|r470|zx80] [ram=K] [rom=nome] [ntsc=on|off] [wrx=on|off] [sound=on|off] [lowram=on|off] [m1not=on|off] [fpfix=on|off]", Help: "mostra a máquina ou troca o modelo e a configuração (desliga e liga)", Run: cmdMachine},
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
		{Name: "load", Usage: "load <programa.p|.bas> [noauto] [nojogo]", Help: `carrega um programa (.p, .81, .p81 ou BASIC em texto .bas/.txt) pela fita; troca a máquina se o jogo tiver configuração (jogo.toml ou [jogos] no config.toml)`, Run: cmdLoad},
		{Name: "save", Usage: "save <programa.p> [NOME]", Help: `executa SAVE "NOME" no ZX81 e grava o programa no arquivo`, Run: cmdSave},
		{Name: "list", Aliases: []string{"listar"}, Usage: "list [programa.p|.o]", Help: "lista o programa BASIC (do ZX81, de um .p, ou de um .o/.80 do ZX80) como texto", Run: cmdList},
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

func (s *Session) description() string { return s.set.Describe() }

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
		case "perfil", "profile":
			cfg, cerr := LoadConfig(ConfigPath())
			if cerr != nil {
				return cerr
			}
			set, err = cfg.MachineSettings(v)
		case "zx81":
			if set.model().ID == "zx81" {
				set.NTSC = false // as before: "machine zx81" = the British ZX81
			} else {
				set, err = set.WithModel("zx81")
			}
		case "ts1000":
			set, err = set.WithModel("ts1000")
			set.RAMK = 2
		case "modelo", "model":
			set, err = set.WithModel(v)
		case "ram":
			var n int
			if n, err = ParseNumber(v); err == nil {
				set.RAMK, err = CheckPack(n)
			}
		case "lowram", "ram816":
			set.LowRAM, err = parseOnOff(v)
		case "m1not":
			set.M1Not, err = parseOnOff(v)
		case "fpfix":
			set.FPFix, err = parseOnOff(v)
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
			if m, ferr := FindModel(k); ferr == nil && !hasV {
				set, err = set.WithModel(m.ID)
				break
			}
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
	if len(a) < 1 || len(a) > 3 {
		return usageErr("load <programa.p|.bas> [noauto] [nojogo]")
	}
	auto, game := true, true
	for _, opt := range a[1:] {
		switch strings.ToLower(opt) {
		case "noauto":
			auto = false
		case "nojogo", "nogame":
			game = false
		default:
			return usageErr("load <programa.p|.bas> [noauto] [nojogo]")
		}
	}
	if IsSnapshot(a[0]) { // a .z81 is the whole machine, not a program
		return cmdSnapshot(s, []string{"load", a[0]}, out)
	}
	if game {
		if err := s.gameMachine(a[0], out); err != nil {
			return err
		}
	} else {
		s.lastProgram = a[0]
	}
	if !tape.IsTapeFile(a[0]) {
		if err := checkProgramFor(s.m, a[0]); err != nil {
			return err
		}
	}
	data, err := os.ReadFile(a[0])
	if err != nil {
		return err
	}
	if !auto {
		n, err := s.insertFile(a[0], data, false)
		if err != nil {
			return err
		}
		if n > 1 {
			fmt.Fprintf(out, "fita inserida (%d blocos); digite LOAD \"\" (ou LOAD \"NOME\") no ZX81\n", n)
		} else {
			fmt.Fprintln(out, `fita inserida; digite LOAD "" no ZX81`)
		}
		return nil
	}
	if err := s.loadProgram(a[0], data); err != nil {
		return err
	}
	fmt.Fprintln(out, "programa carregado:", filepath.Base(a[0]))
	return nil
}

// loadProgram loads a program file (.p, .81, .p81 or BASIC text) by the
// tape and waits until the ZX81 has it.
func (s *Session) loadProgram(file string, data []byte) error {
	if _, err := s.insertFile(file, data, true); err != nil {
		return err
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
	if s.m.ZX80ROM() {
		return s.saveZX80(a[0], out)
	}
	name = strings.ToUpper(name)
	for _, r := range name {
		if zx81.ChordForRune(r) == "" || r == '"' {
			return fmt.Errorf("o nome %q tem caracteres que o ZX81 não aceita num SAVE", name)
		}
	}
	s.lastSave = nil
	s.saveCmd = true
	defer s.flashSave()()
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

// saveZX80 is the save command on a ZX80: SAVE is the E key and takes no
// name; the file is the .o the ROM writes.
func (s *Session) saveZX80(file string, out io.Writer) error {
	s.lastSave = nil
	s.saveCmd = true
	defer s.flashSave()()
	s.waitBoot()
	if err := s.queueChord("E", KeyHold); err != nil {
		return err
	}
	if err := s.queueChord("ENTER", KeyHold); err != nil {
		return err
	}
	if !s.waitUntil(func() bool { return s.lastSave != nil }, 5000) {
		return errors.New(`o SAVE não aconteceu (o ZX80 estava esperando um comando, com o cursor K?)`)
	}
	if err := os.WriteFile(file, s.lastSave.prog, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "SAVE do ZX80 gravado em %s (%d bytes)\n", file, len(s.lastSave.prog))
	return nil
}

// zx80Memory is the ZX80 memory from 4000h to E_LINE, as a .o file.
func (s *Session) zx80Memory() []byte {
	eline := int(s.m.Peek(0x400A)) | int(s.m.Peek(0x400B))<<8
	if eline < basic.ZX80ProgStart || eline > 0xFFFF {
		return nil
	}
	o := make([]byte, eline-basic.ZX80SysVars)
	for i := range o {
		o[i] = s.m.Peek(uint16(basic.ZX80SysVars + i))
	}
	return o
}

// requireZX81 refuses the BASIC tools that only know the ZX81 BASIC.
func (s *Session) requireZX81(cmd string) error {
	if s.m.ZX80ROM() {
		return fmt.Errorf("%s ainda só funciona com o BASIC do ZX81 (esta máquina é um ZX80)", cmd)
	}
	return nil
}

func cmdList(s *Session, a []string, out io.Writer) error {
	var prog []byte
	switch len(a) {
	case 0:
		if s.m.ZX80ROM() {
			o := s.zx80Memory()
			if o == nil {
				return errors.New("E_LINE inválido na memória do ZX80")
			}
			text, err := basic.ListZX80(o)
			if text == "" && err == nil {
				text = "(nenhum programa)\n"
			}
			fmt.Fprint(out, text)
			return err
		}
		prog = basic.ProgramFromMemory(s.m)
	case 1:
		data, err := os.ReadFile(a[0])
		if err != nil {
			return err
		}
		if basic.IsZX80File(a[0]) {
			text, err := basic.ListZX80(data)
			fmt.Fprint(out, text)
			return err
		}
		if prog, err = basic.ProgramFromP(data); err != nil {
			return err
		}
	default:
		return usageErr("list [programa.p|.o]")
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
