// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

// Converting between BASIC text (the zxtext2p format) and programs, without
// the emulated ZX81: txt2p and p2txt in the shell and the consoles, go81
// txt2p / p2txt on the command line, Ferramentas → Converter in the window.
// (The hidden value of each number still comes from the ROM routine, run on
// a Z80 of its own, so the bytes are the ones the ZX81 would store.)

// TextToPFile converts a BASIC text file into a .p file (out "" = the same
// name with .p) and says what it did.
func TextToPFile(in, out string, opt basic.POptions) (string, error) {
	text, err := os.ReadFile(in)
	if err != nil {
		return "", err
	}
	p, err := basic.TextToP(string(text), opt)
	if err != nil {
		return "", fmt.Errorf("%s: %v", in, err)
	}
	if out == "" {
		out = strings.TrimSuffix(in, filepath.Ext(in)) + ".p"
	}
	if err := os.WriteFile(out, p, 0o644); err != nil {
		return "", err
	}
	lines, _ := basic.Lines(mustProgram(p))
	return fmt.Sprintf("%s → %s (%d linhas, %d bytes)", in, out, len(lines), len(p)), nil
}

func mustProgram(p []byte) []byte {
	prog, _ := basic.ProgramFromP(p)
	return prog
}

// ProgramText lists a program file as text: .p, .81 and .p81 (ZX81) or .o
// and .80 (ZX80). It also says how many bytes became "\{XX}" (bytes that
// are not text, such as machine code in a REM): zxtext2p does not read
// those, only Go81.
func ProgramText(in string) (text string, raw int, err error) {
	return programText(in, false)
}

// ProgramTextReadable is ProgramText for people to read: keywords in a REM
// or a string spelled out (basic.TextReadable). It does not always go back
// to the same .p.
func ProgramTextReadable(in string) (text string, raw int, err error) {
	return programText(in, true)
}

func programText(in string, readable bool) (text string, raw int, err error) {
	data, err := os.ReadFile(in)
	if err != nil {
		return "", 0, err
	}
	if looksLikeText(data) {
		return "", 0, fmt.Errorf("isto já é um programa em texto (formato do zxtext2p ou .AZX do Basic Dignified), não um .p: para fazer o .p, use txt2p")
	}
	if tape.IsTapeFile(in) {
		progs, err := TapePrograms(data)
		if err != nil {
			return "", 0, fmt.Errorf("%s: %v", filepath.Base(in), err)
		}
		if len(progs) == 1 {
			text, err = listP(progs[0].P, readable)
			return text, strings.Count(text, `\{`), err
		}
		// Several programs: each one after a "#" line (a comment for
		// txt2p), so the text can be read and split again.
		var sb strings.Builder
		for i, pr := range progs {
			t, err := listP(pr.P, readable)
			if err != nil {
				return sb.String(), 0, fmt.Errorf("programa %s: %v", pr.Label(), err)
			}
			if i > 0 {
				sb.WriteString("\n")
			}
			fmt.Fprintf(&sb, "# programa %s (%d de %d)\n%s", pr.Label(), i+1, len(progs), t)
		}
		text = sb.String()
		return text, strings.Count(text, `\{`), nil
	}
	switch strings.ToLower(filepath.Ext(in)) {
	case ".o", ".80":
		text, err = basic.ListZX80(data)
	case ".p81":
		_, prog, ok := tape.SplitSaved(data)
		if !ok {
			return "", 0, fmt.Errorf("%s: não é um .p81", in)
		}
		text, err = listP(prog, readable)
	default:
		text, err = listP(data, readable)
	}
	return text, strings.Count(text, `\{`), err
}

// FirstTapeProgram returns the first ZX81 program of a .tzx or .t81 tape,
// as a .p (without its name).
func FirstTapeProgram(data []byte) ([]byte, error) {
	progs, err := TapePrograms(data)
	if err != nil {
		return nil, err
	}
	return progs[0].P, nil
}

// TapeProgram is one ZX81 program of a tape: its name (as text, "" when it
// has none) and the program as a .p.
type TapeProgram struct {
	Name string
	P    []byte
}

// Label is the name to show: the program name or "(sem nome)".
func (p TapeProgram) Label() string {
	if p.Name == "" {
		return "(sem nome)"
	}
	return p.Name
}

// TapePrograms returns the ZX81 programs of a .tzx, .t81 or .wav tape, in
// tape order.
func TapePrograms(data []byte) ([]TapeProgram, error) {
	blocks, err := tape.LoadTape(data)
	if err != nil {
		return nil, err
	}
	var progs []TapeProgram
	for _, b := range blocks {
		if b.ID != tape.BlockGeneral {
			continue
		}
		if name, prog, ok := tape.SplitSaved(b.Data[:min(b.DataLen, len(b.Data))]); ok {
			progs = append(progs, TapeProgram{Name: name, P: prog})
		}
	}
	if len(progs) == 0 {
		return nil, fmt.Errorf("não há programa do ZX81 na fita")
	}
	return progs, nil
}

func listP(p []byte, readable bool) (string, error) {
	if readable {
		return basic.ListPReadable(p)
	}
	return basic.ListP(p)
}

// looksLikeText reports whether a file is BASIC text rather than a
// program: printable ASCII lines, the first one starting with a number. (A
// .p starts with the system variables: VERSN 0, then binary addresses.)
func looksLikeText(data []byte) bool {
	t := strings.TrimLeft(string(data), " \t\r\n\uFEFF")
	if t == "" || t[0] < '0' || t[0] > '9' {
		return false
	}
	for _, b := range data {
		if b < ' ' && b != '\n' && b != '\r' && b != '\t' || b == 0x7F {
			return false
		}
	}
	return true
}

// PToTextFile converts a program file into BASIC text (out "" = the same
// name with .txt) and says what it did.
func PToTextFile(in, out string) (string, error) {
	return pToTextFile(in, out, false)
}

// PToTextFileReadable is PToTextFile with ProgramTextReadable.
func PToTextFileReadable(in, out string) (string, error) {
	return pToTextFile(in, out, true)
}

func pToTextFile(in, out string, readable bool) (string, error) {
	if tape.IsTapeFile(in) {
		if msg, done, err := tapeToTextFiles(in, out, readable); done {
			return msg, err
		}
	}
	text, raw, err := programText(in, readable)
	if err != nil && text == "" {
		return "", fmt.Errorf("%s: %v", in, err)
	}
	if out == "" {
		out = strings.TrimSuffix(in, filepath.Ext(in)) + ".txt"
	}
	if err := os.WriteFile(out, []byte(text), 0o644); err != nil {
		return "", err
	}
	msg := fmt.Sprintf("%s → %s (%d linhas)", in, out, strings.Count(text, "\n"))
	return msg + convertNotes(readable, raw), nil
}

// tapeToTextFiles writes each program of a tape with more than one as its
// own text, named after the tape (or out) and the program: fita-JOGO.txt.
// done is false for a tape of one program (written as any other file).
func tapeToTextFiles(in, out string, readable bool) (msg string, done bool, err error) {
	data, err := os.ReadFile(in)
	if err != nil {
		return "", true, err
	}
	progs, err := TapePrograms(data)
	if err != nil {
		return "", true, fmt.Errorf("%s: %v", in, err)
	}
	if len(progs) < 2 {
		return "", false, nil
	}
	if out == "" {
		out = in
	}
	base := strings.TrimSuffix(out, filepath.Ext(out))
	used := map[string]bool{}
	var sb strings.Builder
	raw := 0
	for i, pr := range progs {
		text, err := listP(pr.P, readable)
		if err != nil {
			return sb.String(), true, fmt.Errorf("%s: programa %s: %v", in, pr.Label(), err)
		}
		raw += strings.Count(text, `\{`)
		name := fileSafe(pr.Name)
		if name == "" {
			name = fmt.Sprint(i + 1)
		}
		unique := name
		for k := 2; used[strings.ToLower(unique)]; k++ {
			unique = fmt.Sprintf("%s-%d", name, k)
		}
		used[strings.ToLower(unique)] = true
		file := base + "-" + unique + ".txt"
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			return sb.String(), true, err
		}
		fmt.Fprintf(&sb, "%s: programa %s → %s (%d linhas)\n", in, pr.Label(), file, strings.Count(text, "\n"))
	}
	fmt.Fprintf(&sb, "%d programas", len(progs))
	return sb.String() + convertNotes(readable, raw), true, nil
}

// convertNotes are the notes p2txt adds after what it wrote.
func convertNotes(readable bool, raw int) string {
	msg := ""
	if readable {
		msg += "\nlegível: palavras-chave dentro de REM e de textos escritas por extenso; o texto pode não voltar ao mesmo .p"
	}
	if raw > 0 {
		msg += fmt.Sprintf("\naviso: %d byte(s) sem forma de texto escritos como \\{XX} (código de máquina num REM?): o Go81 lê de volta, o zxtext2p não", raw)
	}
	return msg
}

func registerConvert(r *Registry) {
	r.Register(&Command{Name: "txt2p", Usage: "txt2p <texto.txt> [programa.p] [run=N] [1k]",
		Help: "converte um programa BASIC em texto (formato do zxtext2p) num .p, sem usar o ZX81", Run: cmdTxt2p})
	r.Register(&Command{Name: "p2txt", Usage: "p2txt <programa.p> [texto.txt] [legivel]",
		Help: "converte um .p (.81, .p81, .o, ou os programas de uma fita .tzx/.t81/.wav, um texto para cada) em texto no formato do zxtext2p, que o txt2p transforma de volta (legivel: palavras-chave em REM e textos por extenso)", Run: cmdP2txt})
}

func cmdTxt2p(s *Session, a []string, out io.Writer) error {
	usage := "txt2p <texto.txt> [programa.p] [run=N] [1k]"
	if len(a) < 1 {
		return usageErr(usage)
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
			return usageErr(usage)
		}
	}
	msg, err := TextToPFile(a[0], dest, opt)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, msg)
	return nil
}

func cmdP2txt(s *Session, a []string, out io.Writer) error {
	usage := "p2txt <programa.p> [texto.txt] [legivel]"
	if len(a) < 1 {
		return usageErr(usage)
	}
	dest, readable := "", false
	for _, arg := range a[1:] {
		switch {
		case strings.EqualFold(arg, "legivel") || strings.EqualFold(arg, "legível"):
			readable = true
		case dest == "":
			dest = arg
		default:
			return usageErr(usage)
		}
	}
	msg, err := pToTextFile(a[0], dest, readable)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, msg)
	return nil
}
