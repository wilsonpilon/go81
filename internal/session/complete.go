// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// What each argument of a command can be completed with.
const (
	argFile    = "file"
	argDir     = "dir"
	argCommand = "command"
)

// argKinds lists, per command, what each argument is. The last entry
// repeats for the following arguments. A word list is written "a|b|c".
var argKinds = map[string][]string{
	"help":       {argCommand},
	"load":       {argFile, "noauto"},
	"list":       {argFile},
	"make":       {argFile, argFile},
	"save":       {argFile, ""},
	"script":     {argFile},
	"screenshot": {argFile, "none|small|normal|large|full"},
	"cd":         {argDir},
	"ls":         {argDir},
	"machine":    {"perfil=|zx81|ts1000|ram=|ntsc=on|ntsc=off|wrx=on|wrx=off|sound=on|sound=off|rom="},
	"clock":      {"on|off"},
	"link":       {"start|socket|pipe|stop|status", ""},
	"history":    {""},
}

// Complete returns the possible completions of a command line, each one a
// whole new line. It is used by the shell (Tab), the terminal console and
// the window console.
func Complete(r *Registry, line string) []string {
	// Complete only the last command of "cmd1; cmd2".
	parts := SplitCommands(line)
	head := strings.Join(parts[:len(parts)-1], ";")
	if head != "" {
		head += ";"
	}
	cur := parts[len(parts)-1]
	lead := cur[:len(cur)-len(strings.TrimLeft(cur, " \t"))]
	if head != "" && lead == "" {
		lead = " "
	}
	cur = strings.TrimLeft(cur, " \t")
	prefix := head + lead

	name, rest, hasSpace := strings.Cut(cur, " ")
	if !hasSpace {
		var out []string
		for _, n := range commandNames(r) {
			if strings.HasPrefix(n, strings.ToLower(name)) {
				out = append(out, prefix+n+" ")
			}
		}
		return out
	}
	cmd := r.Lookup(name)
	if cmd == nil || cmd.Raw {
		return nil
	}
	// Split the arguments; the last one (possibly empty) is being typed.
	args, err := SplitArgs(rest)
	if err != nil { // open quote: the last argument is everything after it
		i := strings.LastIndex(rest, `"`)
		args, _ = SplitArgs(rest[:i])
		args = append(args, rest[i+1:])
	} else if rest == "" || strings.HasSuffix(rest, " ") {
		args = append(args, "")
	}
	word := args[len(args)-1]
	kinds := argKinds[cmd.Name]
	if len(kinds) == 0 {
		return nil
	}
	kind := kinds[min(len(args)-1, len(kinds)-1)]

	var cands []string
	switch kind {
	case "":
		return nil
	case argCommand:
		for _, n := range commandNames(r) {
			if strings.HasPrefix(n, strings.ToLower(word)) {
				cands = append(cands, n)
			}
		}
	case argFile, argDir:
		cands = completePath(word, kind == argDir)
	default:
		for _, w := range strings.Split(kind, "|") {
			if strings.HasPrefix(w, strings.ToLower(word)) {
				cands = append(cands, w)
			}
		}
	}
	// Rebuild the line: everything before the word being typed + candidate.
	before := strings.TrimSuffix(cur, word)
	if strings.HasSuffix(before, `"`) {
		before = strings.TrimSuffix(before, `"`)
	}
	out := make([]string, 0, len(cands))
	for _, c := range cands {
		q := c
		if strings.ContainsAny(c, " \t;") {
			q = `"` + c + `"`
		}
		end := " "
		if strings.HasSuffix(c, "/") || strings.HasSuffix(c, `\`) || strings.HasSuffix(c, "=") {
			end = "" // a folder or "ram=": keep typing
			if strings.HasSuffix(q, `"`) {
				q = strings.TrimSuffix(q, `"`)
			}
		}
		out = append(out, prefix+before+q+end)
	}
	return out
}

func commandNames(r *Registry) []string {
	var names []string
	for n := range r.byName {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// completePath lists files (or only folders) starting with word.
func completePath(word string, onlyDirs bool) []string {
	dir, base := filepath.Split(word)
	look := dir
	if look == "" {
		look = "."
	}
	ents, err := os.ReadDir(look)
	if err != nil {
		return nil
	}
	sep := "/"
	if strings.Contains(word, `\`) {
		sep = `\`
	}
	var out []string
	for _, e := range ents {
		n := e.Name()
		if !strings.HasPrefix(strings.ToLower(n), strings.ToLower(base)) {
			continue
		}
		if strings.HasPrefix(n, ".") && !strings.HasPrefix(base, ".") {
			continue
		}
		if e.IsDir() {
			out = append(out, dir+n+sep)
		} else if !onlyDirs {
			out = append(out, dir+n)
		}
	}
	return out
}

// CommonPrefix returns the longest common prefix of the lines.
func CommonPrefix(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	p := lines[0]
	for _, l := range lines[1:] {
		for !strings.HasPrefix(l, p) {
			p = p[:len(p)-1]
		}
	}
	return p
}
