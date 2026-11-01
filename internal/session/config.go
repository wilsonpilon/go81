// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// The configuration file (config.toml) holds the default machine, window
// and Link options and named machine profiles. Options given on the command
// line win over it. Example: see ConfigTemplate.

// MachineConfig is a machine description; empty fields keep the value
// they would have otherwise.
type MachineConfig struct {
	ROM   string `toml:"rom,omitempty"`
	RAM   *int   `toml:"ram,omitempty"`
	NTSC  *bool  `toml:"ntsc,omitempty"`
	WRX   *bool  `toml:"wrx,omitempty"`
	Sound *bool  `toml:"som,omitempty"`
}

// Config is the whole configuration file.
type Config struct {
	Machine MachineConfig `toml:"maquina"`
	Window  struct {
		Zoom   int    `toml:"zoom"`
		Border string `toml:"borda"`
	} `toml:"janela"`
	Link struct {
		Port  *int   `toml:"porta"`
		ZRCP  *int   `toml:"zrcp"`
		Token string `toml:"token"`
	} `toml:"link"`
	Profiles map[string]MachineConfig `toml:"perfis"`

	// Path is where the configuration was read from ("" = no file).
	Path string `toml:"-"`
}

// ConfigPath is the configuration file: $GO81_CONFIG if set, else
// %APPDATA%\go81\config.toml (Windows) or ~/.config/go81/config.toml.
func ConfigPath() string {
	if p := os.Getenv("GO81_CONFIG"); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "go81", "config.toml")
}

// LoadConfig reads a configuration file. A missing file is an empty
// configuration (not an error).
func LoadConfig(path string) (*Config, error) {
	c := &Config{}
	if path == "" {
		return c, nil
	}
	md, err := toml.DecodeFile(path, c)
	if errors.Is(err, fs.ErrNotExist) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %v", path, err)
	}
	if und := md.Undecoded(); len(und) > 0 {
		keys := make([]string, len(und))
		for i, k := range und {
			keys[i] = k.String()
		}
		return nil, fmt.Errorf("%s: opções desconhecidas: %s", path, strings.Join(keys, ", "))
	}
	c.Path = path
	return c, nil
}

// apply copies the fields that are set.
func (mc MachineConfig) apply(s *Settings) error {
	if mc.ROM != "" {
		s.ROM = mc.ROM
	}
	if mc.RAM != nil {
		r, err := checkRAM(*mc.RAM)
		if err != nil {
			return err
		}
		s.RAMK = r
	}
	if mc.NTSC != nil {
		s.NTSC = *mc.NTSC
	}
	if mc.WRX != nil {
		s.WRX = *mc.WRX
	}
	if mc.Sound != nil {
		s.Sound = *mc.Sound
	}
	return nil
}

func checkRAM(k int) (int, error) {
	switch k {
	case 0, 1:
		return 0, nil
	case 2, 4, 8, 16, 32, 48:
		return k, nil
	}
	return 0, fmt.Errorf("RAM inválida: %d (0, 1, 2, 4, 8, 16, 32 ou 48)", k)
}

// MachineSettings returns the machine of the configuration, with a profile
// on top ("" = none).
func (c *Config) MachineSettings(profile string) (Settings, error) {
	s := DefaultSettings()
	if err := c.Machine.apply(&s); err != nil {
		return s, fmt.Errorf("[maquina]: %v", err)
	}
	if profile == "" {
		return s, nil
	}
	p, ok := c.Profiles[profile]
	if !ok {
		return s, fmt.Errorf("perfil desconhecido: %s (perfis: %s)", profile, strings.Join(c.ProfileNames(), ", "))
	}
	if err := p.apply(&s); err != nil {
		return s, fmt.Errorf("[perfis.%s]: %v", profile, err)
	}
	return s, nil
}

// ProfileNames lists the profiles, sorted.
func (c *Config) ProfileNames() []string {
	var n []string
	for k := range c.Profiles {
		n = append(n, k)
	}
	sort.Strings(n)
	if len(n) == 0 {
		return []string{"(nenhum)"}
	}
	return n
}

// ConfigTemplate is written by "go81 config init".
const ConfigTemplate = `# Go81 — configuração
# As opções da linha de comando valem por cima destas.

[maquina]           # a máquina de sempre
rom  = "zx81.edition3"
ram  = 16           # 0 (1K), 2, 4, 8, 16, 32 ou 48
ntsc = false        # true = TS1000 / ZX81 americano
wrx  = false        # alta resolução WRX
som  = false        # "Video Sound" do EightyOne

[janela]
zoom  = 2           # 1 a 4
borda = "normal"    # none, small, normal, large, full

[link]
porta = 8181
zrcp  = 10000       # -1 desliga
# token = "segredo" # se definido, os clientes enviam "auth segredo" antes dos comandos

# Perfis: go81 gui -profile ts1000   (ou, no shell: machine perfil=ts1000)
[perfis.ts1000]
ram  = 2
ntsc = true

[perfis.zx81-1k]
ram = 0

[perfis.wrx]
ram = 16
wrx = true
`

// SaveDefaultMachine writes the [maquina] section of the configuration
// file with these settings, keeping the rest of the file (and its
// comments). The file is created if needed.
func SaveDefaultMachine(path string, s Settings) error {
	if path == "" {
		return errors.New("não há pasta de configuração")
	}
	section := fmt.Sprintf("[maquina]\nrom  = %q\nram  = %d\nntsc = %v\nwrx  = %v\nsom  = %v\n",
		s.ROM, s.RAMK, s.NTSC, s.WRX, s.Sound)
	old, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		old = []byte(ConfigTemplate)
	} else if err != nil {
		return err
	}
	var out []string
	lines := strings.Split(strings.ReplaceAll(string(old), "\r\n", "\n"), "\n")
	done, skipping := false, false
	for _, l := range lines {
		t := strings.TrimSpace(l)
		isHeader := strings.HasPrefix(t, "[")
		switch {
		case isHeader && strings.HasPrefix(t, "[maquina]"):
			out = append(out, strings.TrimRight(section, "\n"))
			done, skipping = true, true
			continue
		case isHeader:
			if skipping {
				out = append(out, "")
			}
			skipping = false
		case skipping:
			continue
		}
		out = append(out, l)
	}
	if !done {
		out = append([]string{strings.TrimRight(section, "\n"), ""}, out...)
	}
	text := strings.Join(out, "\n")
	// Check that the result is still valid before writing it.
	var check Config
	if _, err := toml.Decode(text, &check); err != nil {
		return fmt.Errorf("o arquivo de configuração ficaria inválido: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0o644)
}
