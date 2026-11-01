// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package man is the manual built into Go81: a page for every command of
// the shell and of the command line, for the BASIC keywords of the ZX81 and
// of the ZX80, for the system variables, the error codes and the labels of
// the ZX81 ROM, and a few reference pages. The idea comes from fwMSX's man
// (same author); the pages are data compiled into the program.
//
//	man print          the first page called "print" (see Sections for the order)
//	man basic print    the page in one section
//	man -k som         pages that mention a word
//	man -l [section]   list
package man

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"
)

// Topic is one page.
type Topic struct {
	Name    string   // as typed: "load", "PRINT", "FP-CALC", "D_FILE"
	Aliases []string // other names
	Section string   // see Sections
	Usage   string   // syntax line ("" = none)
	Summary string   // one line
	Page    string   // the text (may be empty for a bare ROM label)
	SeeAlso []string // other pages ("section name" or "name")
}

// Section is a part of the manual.
type Section struct {
	ID    string
	Title string
}

// Sections in the order "man name" looks through them.
var Sections = []Section{
	{"basic", "BASIC do ZX81: palavras-chave e funções"},
	{"shell", "Comandos do shell, dos consoles e do Link"},
	{"cli", "Comandos da linha de comando (go81 ...)"},
	{"basic80", "BASIC do ZX80"},
	{"sysvar", "Variáveis de sistema do ZX81"},
	{"rom", "ROM do ZX81 (edição 3): rotinas e rótulos"},
	{"erro", "Códigos de erro (relatórios) do ZX81 e do ZX80"},
	{"ref", "Referência: caracteres, teclado, memória, formatos"},
}

var (
	mu     sync.Mutex
	topics []*Topic
	built  bool
)

// add registers pages (called by the content files when the index is built).
func add(ts ...*Topic) { topics = append(topics, ts...) }

// providers fill the index on first use (the ROM tables need the ROM).
var providers []func()

func build() {
	mu.Lock()
	defer mu.Unlock()
	if built {
		return
	}
	keyMaps()
	for _, p := range providers {
		p()
	}
	sort.SliceStable(topics, func(i, j int) bool {
		si, sj := sectionIndex(topics[i].Section), sectionIndex(topics[j].Section)
		if si != sj {
			return si < sj
		}
		return strings.ToLower(topics[i].Name) < strings.ToLower(topics[j].Name)
	})
	built = true
}

func sectionIndex(id string) int {
	for i, s := range Sections {
		if s.ID == id {
			return i
		}
	}
	return len(Sections)
}

// IsSection reports whether s is a section id.
func IsSection(s string) bool { return sectionIndex(strings.ToLower(s)) < len(Sections) }

// Topics returns every page, by section and name.
func Topics() []*Topic {
	build()
	return topics
}

func norm(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }

func (t *Topic) names() []string { return append([]string{t.Name}, t.Aliases...) }

// Find returns the pages called name (by name or alias, upper and lower case
// the same), in section order (in a section, the exact name before an
// alias); section "" = any.
func Find(section, name string) []*Topic {
	n := norm(name)
	type hit struct {
		t    *Topic
		rank int
	}
	var hits []hit
	for _, t := range Topics() {
		if section != "" && t.Section != section {
			continue
		}
		rank := -1
		if norm(t.Name) == n {
			rank = 0
		} else {
			for _, a := range t.Aliases {
				if norm(a) == n {
					rank = 1
					break
				}
			}
		}
		if rank >= 0 {
			hits = append(hits, hit{t, sectionIndex(t.Section)*2 + rank})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].rank < hits[j].rank })
	out := make([]*Topic, len(hits))
	for i, h := range hits {
		out[i] = h.t
	}
	return out
}

// Resolve is Find, and else the single page whose name starts with name.
func Resolve(section, name string) []*Topic {
	if r := Find(section, name); len(r) > 0 {
		return r
	}
	n := norm(name)
	var pre []*Topic
	for _, t := range Topics() {
		if (section == "" || t.Section == section) && strings.HasPrefix(norm(t.Name), n) {
			pre = append(pre, t)
		}
	}
	if len(pre) == 1 {
		return pre
	}
	return nil
}

// Search returns the pages whose name, summary or text contain the word.
// Names first, then summaries, then texts.
func Search(word string) []*Topic {
	w := norm(word)
	if w == "" {
		return nil
	}
	var a, b, c []*Topic
	for _, t := range Topics() {
		switch {
		case containsAny(t.names(), w):
			a = append(a, t)
		case strings.Contains(strings.ToLower(t.Summary), w):
			b = append(b, t)
		case strings.Contains(strings.ToLower(t.Page), w) || strings.Contains(strings.ToLower(t.Usage), w):
			c = append(c, t)
		}
	}
	return append(append(a, b...), c...)
}

func containsAny(names []string, w string) bool {
	for _, n := range names {
		if strings.Contains(norm(n), w) {
			return true
		}
	}
	return false
}

// Ref is how to ask for a page: "man basic PRINT".
func (t *Topic) Ref() string { return t.Section + " " + t.Name }

// Format writes a page, wrapped to width columns.
func (t *Topic) Format(width int) string {
	var sb strings.Builder
	title := fmt.Sprintf("%s(%s)", t.Name, t.Section)
	sb.WriteString(title)
	if t.Summary != "" {
		sb.WriteString(" — " + t.Summary)
	}
	sb.WriteString("\n")
	if t.Usage != "" {
		sb.WriteString("\nUSO\n")
		for _, u := range strings.Split(t.Usage, "\n") {
			sb.WriteString("    " + u + "\n")
		}
	}
	if len(t.Aliases) > 0 {
		sb.WriteString("\nOUTROS NOMES\n    " + strings.Join(t.Aliases, ", ") + "\n")
	}
	if p := strings.Trim(t.Page, "\n"); p != "" {
		sb.WriteString("\n")
		for _, l := range Wrap(p, width) {
			sb.WriteString(l + "\n")
		}
	}
	if len(t.SeeAlso) > 0 {
		sb.WriteString("\nVEJA TAMBÉM\n    " + strings.Join(t.SeeAlso, ", ") + "\n")
	}
	return sb.String()
}

// Wrap breaks text into lines of at most width columns, keeping each
// line's indentation. Lines that start with two spaces or more (examples,
// tables) are not wrapped.
func Wrap(text string, width int) []string {
	if width < 20 {
		width = 20
	}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		indent := len(line) - len(strings.TrimLeftFunc(line, unicode.IsSpace))
		if indent >= 2 || len([]rune(line)) <= width {
			out = append(out, strings.TrimRight(line, " "))
			continue
		}
		pre := line[:indent]
		cur := pre
		for _, w := range strings.Fields(line) {
			if cur != pre && len([]rune(cur))+1+len([]rune(w)) > width {
				out = append(out, cur)
				cur = pre
			}
			if cur != pre {
				cur += " "
			}
			cur += w
		}
		out = append(out, cur)
	}
	return out
}

// List returns the index of a section ("" = all), one line per page.
func List(section string) string {
	var sb strings.Builder
	for _, s := range Sections {
		if section != "" && s.ID != section {
			continue
		}
		fmt.Fprintf(&sb, "%s — %s\n", s.ID, s.Title)
		for _, t := range Topics() {
			if t.Section != s.ID {
				continue
			}
			sum := t.Summary
			if sum == "" {
				sum = "(rótulo da ROM)"
			}
			fmt.Fprintf(&sb, "  %-14s %s\n", t.Name, sum)
		}
	}
	return sb.String()
}

// Intro is the page shown by "man" alone.
func Intro() string {
	var sb strings.Builder
	sb.WriteString("man — o manual do Go81\n\n")
	sb.WriteString("    man <nome>            a página (comando, palavra do BASIC, rotina da ROM...)\n")
	sb.WriteString("    man <seção> <nome>    a página de uma seção (ex.: man basic run)\n")
	sb.WriteString("    man -k <palavra>      procura nas páginas\n")
	sb.WriteString("    man -l [seção]        lista as páginas\n\nSEÇÕES\n")
	counts := map[string]int{}
	for _, t := range Topics() {
		counts[t.Section]++
	}
	for _, s := range Sections {
		fmt.Fprintf(&sb, "    %-8s %s (%d)\n", s.ID, s.Title, counts[s.ID])
	}
	sb.WriteString("\nComece por: man intro, man shell, man basic, man memoria, man D_FILE, man FP-CALC.\n")
	return sb.String()
}

// Lookup answers a "man" command line (the words after "man") and returns
// the text to show.
func Lookup(args []string, width int) (string, error) {
	if len(args) == 0 {
		return Intro(), nil
	}
	switch args[0] {
	case "-k":
		if len(args) < 2 {
			return "", fmt.Errorf("uso: man -k <palavra>")
		}
		r := Search(strings.Join(args[1:], " "))
		if len(r) == 0 {
			return "", fmt.Errorf("nenhuma página fala de %q", strings.Join(args[1:], " "))
		}
		var sb strings.Builder
		for _, t := range r {
			fmt.Fprintf(&sb, "%-22s %s\n", t.Ref(), t.Summary)
		}
		return sb.String(), nil
	case "-l":
		sec := ""
		if len(args) > 1 {
			sec = strings.ToLower(args[1])
			if !IsSection(sec) {
				return "", fmt.Errorf("seção desconhecida: %s (%s)", args[1], sectionIDs())
			}
		}
		return List(sec), nil
	}
	if len(args) == 1 && IsSection(args[0]) {
		return List(strings.ToLower(args[0])), nil
	}
	sec, name := "", strings.Join(args, " ")
	if len(args) > 1 && IsSection(args[0]) {
		sec, name = strings.ToLower(args[0]), strings.Join(args[1:], " ")
	}
	r := Resolve(sec, name)
	if len(r) == 0 {
		if s := Search(name); len(s) > 0 {
			var sb strings.Builder
			fmt.Fprintf(&sb, "não há página %q; páginas que falam disso:\n", name)
			for i, t := range s {
				if i == 10 {
					fmt.Fprintf(&sb, "... (%d no total: man -k %s)\n", len(s), name)
					break
				}
				fmt.Fprintf(&sb, "  man %-20s %s\n", t.Ref(), t.Summary)
			}
			return sb.String(), nil
		}
		return "", fmt.Errorf("não há página %q (man -l lista todas)", name)
	}
	text := r[0].Format(width)
	if len(r) > 1 {
		var others []string
		for _, t := range r[1:] {
			others = append(others, "man "+t.Ref())
		}
		text += "\nTAMBÉM COM ESTE NOME\n    " + strings.Join(others, "\n    ") + "\n"
	}
	return text, nil
}

func sectionIDs() string {
	var ids []string
	for _, s := range Sections {
		ids = append(ids, s.ID)
	}
	return strings.Join(ids, ", ")
}
