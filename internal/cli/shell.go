// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/wilsonpilon/go81/internal/link"
	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/shell"
)

// stdin is the shell input (tests replace it).
var stdin io.Reader = os.Stdin

// stdinIsTerminal reports whether the shell talks to a person.
var stdinIsTerminal = func() bool {
	st, err := os.Stdin.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

// machineOpts are the machine options shared by gui, tui, shell and link,
// plus -config and -profile. The machine comes from the configuration file,
// then the profile, then the options given on the command line.
type machineOpts struct {
	fs      *flag.FlagSet
	model   *string
	lowRAM  *bool
	m1not   *bool
	fpfix   *bool
	rom     *string
	ram     *int
	ntsc    *bool
	wrx     *bool
	sound   *bool
	config  *string
	profile *string
	token   *string
}

// linkToken returns the Link token: the option if typed, else the
// configuration.
func (o *machineOpts) linkToken(cfg *session.Config) string {
	if o.given("token") {
		return *o.token
	}
	return cfg.Link.Token
}

func machineFlags(fs *flag.FlagSet) *machineOpts {
	return &machineOpts{
		fs:      fs,
		model:   fs.String("model", "zx81", "modelo: zx81, ts1000, ts1500, tk85, r470 ou zx80 (traz a ROM, a RAM e o NTSC dele)"),
		lowRAM:  fs.Bool("lowram", false, "RAM em 8K-16K"),
		m1not:   fs.Bool("m1not", false, "modificação M1NOT (código em 32K-48K)"),
		fpfix:   fs.Bool("fpfix", false, "correção do ponto flutuante (ROM edição 1)"),
		rom:     fs.String("rom", "zx81.edition3", "ROM embutida (veja \"go81 roms\") ou caminho de um arquivo .rom"),
		ram:     fs.Int("ram", 16, "RAM pack em K: 0 (nenhum), 2, 4, 8, 16, 32 ou 48 (soma aos 16K do TS1500, TK85 e R-470)"),
		ntsc:    fs.Bool("ntsc", false, "ZX81 americano / TS1000 (60 Hz)"),
		wrx:     fs.Bool("wrx", false, "alta resolução WRX"),
		sound:   fs.Bool("sound", false, "som pelo vídeo (Video Sound do EightyOne)"),
		token:   fs.String("token", "", "senha da porta de comunicação (clientes enviam auth <senha>)"),
		config:  fs.String("config", "", "arquivo de configuração (padrão: "+session.ConfigPath()+")"),
		profile: fs.String("profile", "", "perfil de máquina do arquivo de configuração"),
	}
}

// given reports whether an option was typed on the command line.
func (o *machineOpts) given(name string) bool {
	found := false
	o.fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// load returns the machine settings and the configuration.
func (o *machineOpts) load() (session.Settings, *session.Config, error) {
	path := *o.config
	if path == "" {
		path = session.ConfigPath()
	} else if _, err := os.Stat(path); err != nil {
		return session.Settings{}, nil, err
	}
	cfg, err := session.LoadConfig(path)
	if err != nil {
		return session.Settings{}, nil, err
	}
	set, err := cfg.MachineSettings(*o.profile)
	if err != nil {
		return set, cfg, err
	}
	if o.given("model") {
		if set, err = set.WithModel(*o.model); err != nil {
			return set, cfg, err
		}
	}
	if o.given("rom") {
		set.ROM = *o.rom
	}
	if o.given("ram") {
		if set.RAMK, err = session.CheckPack(*o.ram); err != nil {
			return set, cfg, err
		}
	}
	if o.given("lowram") {
		set.LowRAM = *o.lowRAM
	}
	if o.given("m1not") {
		set.M1Not = *o.m1not
	}
	if o.given("fpfix") {
		set.FPFix = *o.fpfix
	}
	if o.given("ntsc") {
		set.NTSC = *o.ntsc
	}
	if o.given("wrx") {
		set.WRX = *o.wrx
	}
	if o.given("sound") {
		set.Sound = *o.sound
	}
	return set, cfg, nil
}

// linkPorts returns the Link ports: the options if typed, else the
// configuration, else the defaults.
func linkPorts(o *machineOpts, cfg *session.Config, port, zrcp *int) (int, int) {
	p, z := *port, *zrcp
	if !o.given("port") && cfg.Link.Port != nil {
		p = *cfg.Link.Port
	}
	if !o.given("zrcp") && cfg.Link.ZRCP != nil {
		z = *cfg.Link.ZRCP
	}
	return p, z
}

func cmdShell(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("shell", flag.ContinueOnError)
	fs.SetOutput(stderr)
	mo := machineFlags(fs)
	var (
		cmds     = fs.String("c", "", "executa estes comandos (separados por ;) e sai")
		linkOn   = fs.Bool("link", false, "liga também a porta de comunicação (link start)")
		port     = fs.Int("port", link.DefaultPort, "porta do protocolo Go81 (com -link)")
		zrcpPort = fs.Int("zrcp", link.DefaultZRCPPort, "porta ZRCP compatível com ZEsarUX (com -link; -1 desliga)")
	)
	fs.Usage = func() {
		fmt.Fprint(stderr, `Uso: go81 shell [opções] [script.g81]

Abre o shell do Go81: cada linha é um comando (help lista todos).
Com um script, executa o arquivo e para no primeiro erro.

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
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "go81 shell: só um script por vez")
		return ExitUsage
	}
	set, cfg, err := mo.load()
	if err != nil {
		fmt.Fprintln(stderr, "go81 shell:", err)
		return ExitUsage
	}
	*port, *zrcpPort = linkPorts(mo, cfg, port, zrcpPort)
	sess, err := session.New(set)
	if err != nil {
		fmt.Fprintln(stderr, "go81 shell:", err)
		return ExitError
	}
	ports := link.Register(sess)
	ports.Token = mo.linkToken(cfg)
	defer ports.Stop()
	defer sess.StopClock()
	if *linkOn {
		if err := ports.Start(sess, *port, *zrcpPort); err != nil {
			fmt.Fprintln(stderr, "go81 shell:", err)
			return ExitError
		}
		sess.StartClock()
		fmt.Fprintln(stdout, ports.Status())
	}

	var runErr error
	switch {
	case *cmds != "":
		runErr = shell.ExecLine(sess, *cmds, stdout)
		if errors.Is(runErr, session.ErrQuit) {
			runErr = nil
		}
	case fs.NArg() == 1:
		runErr = sess.Exec(fmt.Sprintf("script %q", fs.Arg(0)), stdout)
		if errors.Is(runErr, session.ErrQuit) {
			runErr = nil
		}
	default:
		term := stdinIsTerminal()
		runErr = shell.Run(sess, stdin, stdout, stderr, shell.Options{Interactive: term, LineEditor: term && stdin == io.Reader(os.Stdin)})
	}
	if runErr != nil {
		fmt.Fprintln(stderr, "erro:", runErr)
		return ExitError
	}
	return ExitOK
}

func cmdLink(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("link", flag.ContinueOnError)
	fs.SetOutput(stderr)
	mo := machineFlags(fs)
	var (
		port     = fs.Int("port", link.DefaultPort, "porta TCP do protocolo Go81 (-1 desliga)")
		zrcpPort = fs.Int("zrcp", link.DefaultZRCPPort, "porta ZRCP compatível com ZEsarUX (-1 desliga)")
		stdio    = fs.Bool("stdio", false, "protocolo Go81 pela entrada e saída padrão (as portas TCP só ligam se pedidas)")
		socket   = fs.String("socket", "", "também escuta neste socket Unix (arquivo)")
		pipe     = fs.String("pipe", "", `também escuta neste pipe nomeado do Windows (ex.: go81 = \\.\pipe\go81)`)
	)
	fs.Usage = func() {
		fmt.Fprint(stderr, linkUsage)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "go81 link: só um programa por vez")
		return ExitUsage
	}
	set, cfg, err := mo.load()
	if err != nil {
		fmt.Fprintln(stderr, "go81 link:", err)
		return ExitUsage
	}
	*port, *zrcpPort = linkPorts(mo, cfg, port, zrcpPort)
	// With -stdio, stdout belongs to the protocol: messages go to stderr,
	// and TCP opens only when asked for.
	msg := stdout
	if *stdio {
		msg = stderr
		if !mo.given("port") {
			*port = -1
		}
		if !mo.given("zrcp") {
			*zrcpPort = -1
		}
	}
	sess, err := session.New(set)
	if err != nil {
		fmt.Fprintln(stderr, "go81 link:", err)
		return ExitError
	}
	ports := link.Register(sess)
	ports.Token = mo.linkToken(cfg)
	if fs.NArg() == 1 {
		if err := sess.Exec(fmt.Sprintf("load %q", fs.Arg(0)), msg); err != nil {
			fmt.Fprintln(stderr, "go81 link:", err)
			return ExitError
		}
	}
	fail := func(err error) int {
		ports.Stop()
		fmt.Fprintln(stderr, "go81 link:", err)
		return ExitError
	}
	if *port >= 0 {
		if err := ports.Start(sess, *port, *zrcpPort); err != nil {
			return fail(err)
		}
	}
	if *socket != "" {
		if err := ports.AddSocket(sess, *socket); err != nil {
			return fail(err)
		}
	}
	if *pipe != "" {
		if err := ports.AddPipe(sess, *pipe); err != nil {
			return fail(err)
		}
	}
	if !*stdio && !ports.Any() {
		fmt.Fprintln(stderr, "go81 link: nenhuma porta ligada (use -port, -socket, -pipe ou -stdio)")
		return ExitUsage
	}
	sess.StartClock()
	defer sess.StopClock()
	defer ports.Stop()

	if *stdio {
		if ports.Any() {
			fmt.Fprintln(msg, ports.Status())
		}
		link.ServeStream(sess, stdin, stdout, link.Options{Token: ports.Token})
		return ExitOK
	}
	fmt.Fprintln(msg, ports.Status())
	fmt.Fprintln(msg, "Ctrl+C encerra.")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Fprintln(msg, "link encerrado")
	return ExitOK
}

const linkUsage = `Uso: go81 link [opções] [programa.p]

Roda o ZX81 sem janela, em tempo real, controlado pela porta de
comunicação. Ctrl+C encerra (com -stdio, o fim da entrada).

  Protocolo Go81: uma linha por comando do shell; resposta = saída + "ok"
  ou "err <mensagem>". Teste com: telnet 127.0.0.1 8181
  ZRCP: subconjunto do protocolo do ZEsarUX (get-registers, read-memory...)

Canais (podem ser usados juntos):
  TCP em 127.0.0.1 (padrão: 8181 e ZRCP 10000)
  -stdio            entrada/saída padrão, para quem abre o go81 como processo
  -socket <arquivo> socket Unix (Linux e Windows 10+)
  -pipe <nome>      pipe nomeado do Windows, só o usuário atual conecta

Opções:
`
