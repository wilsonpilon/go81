// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Command is one command of the registry.
type Command struct {
	Name    string
	Aliases []string
	Usage   string // e.g. "peek <endereço> [quantidade]"
	Help    string // one line, in Portuguese
	// Raw commands receive the rest of the line unparsed as args[0].
	Raw bool
	Run func(s *Session, args []string, out io.Writer) error
}

// Registry holds the commands.
type Registry struct {
	byName map[string]*Command
	list   []*Command
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{byName: map[string]*Command{}} }

// Register adds (or replaces) a command.
func (r *Registry) Register(c *Command) {
	r.byName[strings.ToLower(c.Name)] = c
	for _, a := range c.Aliases {
		r.byName[strings.ToLower(a)] = c
	}
	for i, old := range r.list {
		if old.Name == c.Name {
			r.list[i] = c
			return
		}
	}
	r.list = append(r.list, c)
}

// Lookup finds a command by name or alias (case insensitive).
func (r *Registry) Lookup(name string) *Command { return r.byName[strings.ToLower(name)] }

// Commands returns the commands sorted by name.
func (r *Registry) Commands() []*Command {
	l := append([]*Command(nil), r.list...)
	sort.Slice(l, func(i, j int) bool { return l[i].Name < l[j].Name })
	return l
}

// splitCommand returns the first word of a line and the rest.
func splitCommand(line string) (string, string) {
	line = strings.TrimSpace(line)
	i := strings.IndexFunc(line, unicode.IsSpace)
	if i < 0 {
		return line, ""
	}
	return line[:i], strings.TrimSpace(line[i:])
}

// SplitArgs splits a line into words; "double quotes" group words and \"
// inside quotes is a quote.
func SplitArgs(s string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inQuote, have := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inQuote && c == '\\' && i+1 < len(s) && s[i+1] == '"':
			cur.WriteByte('"')
			i++
		case c == '"':
			inQuote = !inQuote
			have = true
		case !inQuote && (c == ' ' || c == '\t'):
			if have {
				args = append(args, cur.String())
				cur.Reset()
				have = false
			}
		default:
			cur.WriteByte(c)
			have = true
		}
	}
	if inQuote {
		return nil, fmt.Errorf("aspas não fechadas")
	}
	if have {
		args = append(args, cur.String())
	}
	return args, nil
}

// SplitCommands splits "cmd1; cmd2" on semicolons outside quotes.
func SplitCommands(s string) []string {
	var out []string
	inQuote, start := false, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			inQuote = !inQuote
		case ';':
			if !inQuote {
				out = append(out, s[start:i])
				start = i + 1
			}
		}
	}
	return append(out, s[start:])
}

// ParseNumber reads a number with the same rules as ZEsarUX's ZRCP:
// decimal by default; hexadecimal with a $ or 0x prefix or an H suffix;
// binary with a % suffix; 'c' is the character code of c.
func ParseNumber(s string) (int, error) {
	t := strings.TrimSpace(s)
	if t == "" {
		return 0, fmt.Errorf("número vazio")
	}
	if (t[0] == '\'' || t[0] == '"') && len(t) >= 2 {
		return int(t[1]), nil
	}
	neg := false
	if t[0] == '-' {
		neg, t = true, t[1:]
	}
	var v int64
	var err error
	lower := strings.ToLower(t)
	switch {
	case strings.HasPrefix(t, "$"):
		v, err = strconv.ParseInt(t[1:], 16, 64)
	case strings.HasPrefix(lower, "0x"):
		v, err = strconv.ParseInt(t[2:], 16, 64)
	case strings.HasSuffix(lower, "h") && isHex(t[:len(t)-1]):
		v, err = strconv.ParseInt(t[:len(t)-1], 16, 64)
	case strings.HasSuffix(t, "%"):
		v, err = strconv.ParseInt(t[:len(t)-1], 2, 64)
	default:
		v, err = strconv.ParseInt(t, 10, 64)
	}
	if err != nil {
		return 0, fmt.Errorf("número inválido: %s", s)
	}
	if neg {
		v = -v
	}
	return int(v), nil
}

func isHex(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !unicode.Is(unicode.ASCII_Hex_Digit, c) {
			return false
		}
	}
	return true
}

// parseAddr parses a 16-bit address.
func parseAddr(s string) (uint16, error) {
	v, err := ParseNumber(s)
	if err != nil {
		return 0, err
	}
	if v < 0 || v > 0xFFFF {
		return 0, fmt.Errorf("endereço fora de 0..65535: %s", s)
	}
	return uint16(v), nil
}

// parseOnOff reads on/off, sim/não, 1/0, true/false.
func parseOnOff(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "on", "sim", "s", "1", "true", "yes", "ligado":
		return true, nil
	case "off", "nao", "não", "n", "0", "false", "no", "desligado":
		return false, nil
	}
	return false, fmt.Errorf("use on ou off: %s", s)
}
