// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Per-game configuration: the machine a program needs (a WRX game, a 1K
// game, a TS1000 program...). When a program is loaded, the machine is
// changed if there is
//
//   - a file <name>.toml beside the program (it goes with the game), or
//   - a section [jogos."name.p"] (or [jogos.name]) in config.toml.
//
// Both have the keys of [maquina], and perfil = "name" to start from a
// profile. Loading a program without its own configuration brings back
// the machine there was before the last game configuration.

// GameConfig is the machine of one game.
type GameConfig struct {
	Profile string `toml:"perfil,omitempty"`
	MachineConfig
}

// GameFile is the per-game file beside a program: jogo.p → jogo.toml.
func GameFile(program string) string {
	return strings.TrimSuffix(program, filepath.Ext(program)) + ".toml"
}

// FindGame returns the configuration of a program and where it came from
// (ok = false if it has none).
func (c *Config) FindGame(program string) (gc GameConfig, from string, ok bool, err error) {
	side := GameFile(program)
	md, err := toml.DecodeFile(side, &gc)
	switch {
	case err == nil:
		if und := md.Undecoded(); len(und) > 0 {
			return gc, side, false, fmt.Errorf("%s: opções desconhecidas: %s", side, und[0].String())
		}
		return gc, filepath.Base(side), true, nil
	case !errors.Is(err, fs.ErrNotExist):
		return gc, side, false, fmt.Errorf("%s: %v", side, err)
	}
	base := strings.ToLower(filepath.Base(program))
	for _, key := range []string{base, strings.TrimSuffix(base, filepath.Ext(base))} {
		for k, v := range c.Games {
			if strings.ToLower(k) == key {
				return v, fmt.Sprintf("[jogos.%q] em %s", k, filepath.Base(c.Path)), true, nil
			}
		}
	}
	return gc, "", false, nil
}

// GameSettings is the machine for a game: cur (or the profile) with the
// fields of the game on top.
func (c *Config) GameSettings(gc GameConfig, cur Settings) (Settings, error) {
	set := cur
	if gc.Profile != "" {
		var err error
		if set, err = c.MachineSettings(gc.Profile); err != nil {
			return cur, err
		}
	}
	if err := gc.apply(&set); err != nil {
		return cur, err
	}
	return set, nil
}

// gameMachine changes the machine for the program about to be loaded, as
// its configuration says (see above), and says so in out.
func (s *Session) gameMachine(program string, out io.Writer) error {
	s.lastProgram = program
	cfg, err := LoadConfig(ConfigPath())
	if err != nil {
		return err
	}
	gc, from, ok, err := cfg.FindGame(program)
	if err != nil {
		return err
	}
	want := s.set
	switch {
	case ok:
		base := s.set
		if s.gameBase != nil {
			base = *s.gameBase
		}
		if want, err = cfg.GameSettings(gc, base); err != nil {
			return fmt.Errorf("%s: %v", from, err)
		}
		if s.gameBase == nil {
			b := s.set
			s.gameBase = &b
		}
		s.gameFrom = from
	case s.gameBase != nil: // back to the machine before the last game
		want, s.gameBase, s.gameFrom = *s.gameBase, nil, ""
		from = "a máquina de antes"
	default:
		s.gameFrom = ""
		return nil
	}
	if want == s.set {
		return nil
	}
	if err := s.configure(want); err != nil {
		return err
	}
	fmt.Fprintf(out, "máquina (%s): %s\n", from, s.description())
	return nil
}

// GameMachine is gameMachine for the windows: it returns the message
// ("" if the machine stayed the same).
func (s *Session) GameMachine(program string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var b strings.Builder
	err := s.gameMachine(program, &b)
	return strings.TrimSpace(b.String()), err
}

func registerGameConfig(r *Registry) {
	r.Register(&Command{Name: "jogo", Aliases: []string{"game"}, Usage: "jogo [salvar | apagar]",
		Help: "mostra a configuração de máquina do último programa carregado; salvar grava a máquina atual em <programa>.toml, ao lado dele",
		Run:  cmdGame})
}

func cmdGame(s *Session, a []string, out io.Writer) error {
	if s.lastProgram == "" {
		return errors.New("nenhum programa foi carregado de um arquivo (load)")
	}
	side := GameFile(s.lastProgram)
	switch {
	case len(a) == 0:
		fmt.Fprintln(out, "programa:", s.lastProgram)
		if s.gameFrom != "" {
			fmt.Fprintf(out, "configuração: %s\n", s.gameFrom)
		} else {
			fmt.Fprintf(out, "sem configuração própria (jogo salvar cria %s)\n", filepath.Base(side))
		}
		fmt.Fprintln(out, "máquina:", s.description())
		return nil
	case len(a) == 1 && (a[0] == "salvar" || a[0] == "save"):
		if err := os.WriteFile(side, []byte(gameText(s.set)), 0o644); err != nil {
			return err
		}
		s.gameFrom = filepath.Base(side)
		if s.gameBase == nil {
			b := s.set
			s.gameBase = &b
		}
		fmt.Fprintln(out, "máquina deste jogo gravada em", side)
		return nil
	case len(a) == 1 && (a[0] == "apagar" || a[0] == "delete"):
		if err := os.Remove(side); err != nil {
			return err
		}
		s.gameFrom = ""
		fmt.Fprintln(out, "apagado:", side)
		return nil
	}
	return usageErr("jogo [salvar | apagar]")
}

// gameText is a per-game file with the whole machine.
func gameText(set Settings) string {
	model := set.Model
	if model == "" {
		model = "zx81"
	}
	return fmt.Sprintf("# Go81 — a máquina deste jogo (as chaves de [maquina] do config.toml)\n"+
		"modelo = %q\nrom  = %q\nram  = %d\nntsc = %v\nwrx  = %v\nsom  = %v\nram_8k_16k  = %v\nm1not       = %v\ncorrecao_pf = %v\n",
		model, set.ROM, set.RAMK, set.NTSC, set.WRX, set.Sound, set.LowRAM, set.M1Not, set.FPFix)
}
