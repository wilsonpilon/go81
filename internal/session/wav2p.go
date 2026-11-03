// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/internal/media/tape"
)

// wav2p: the programs of a tape recorded as sound (.wav, from a cassette)
// become .p files, without the emulated ZX81. The sound is read the way the
// ROM's LOAD reads it (EightyOne's WAV loader).

// WavToPFiles writes each program of the WAV as a .p (or .o, zx80) in dir
// ("" = the folder of the WAV), named after the program, and says what it
// did.
func WavToPFiles(in, dir string, zx80 bool) (string, error) {
	data, err := os.ReadFile(in)
	if err != nil {
		return "", err
	}
	blocks, err := tape.LoadWAV(data, zx80)
	if err != nil {
		return "", fmt.Errorf("%s: %v", in, err)
	}
	if dir == "" {
		dir = filepath.Dir(in)
	}
	base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
	used := map[string]bool{}
	var sb strings.Builder
	n := 0
	for _, b := range blocks {
		if b.ID != tape.BlockGeneral {
			continue
		}
		n++
		prog, name, ext := b.Data, "", ".o"
		if !zx80 {
			var ok bool
			if name, prog, ok = tape.SplitSaved(b.Data); !ok {
				continue
			}
			ext = ".p"
		}
		file := fileSafe(name)
		if file == "" {
			file = fmt.Sprintf("%s-%d", base, n)
		}
		unique := file
		for k := 2; used[strings.ToLower(unique)]; k++ {
			unique = fmt.Sprintf("%s-%d", file, k)
		}
		used[strings.ToLower(unique)] = true
		out := filepath.Join(dir, unique+ext)
		if err := os.WriteFile(out, prog, 0o644); err != nil {
			return sb.String(), err
		}
		if name == "" {
			name = "(sem nome)"
		}
		fmt.Fprintf(&sb, "%s: programa %s → %s (%d bytes)\n", filepath.Base(in), name, out, len(prog))
	}
	if n == 0 {
		return "", fmt.Errorf("%s: nenhum programa encontrado", in)
	}
	return strings.TrimSuffix(sb.String(), "\n"), nil
}

// fileSafe makes a ZX81 program name a file name: letters and digits stay,
// the rest becomes "_".
func fileSafe(name string) string {
	var sb strings.Builder
	for _, r := range strings.TrimSpace(name) {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			sb.WriteRune(r)
		default:
			sb.WriteByte('_')
		}
	}
	return strings.Trim(sb.String(), "_")
}

func registerWav(r *Registry) {
	r.Register(&Command{Name: "wav2p", Usage: "wav2p <fita.wav> [pasta] [zx80]",
		Help: "lê uma fita gravada em som (.wav) e grava cada programa num .p (ou .o), sem usar o ZX81", Run: cmdWav2p})
}

func cmdWav2p(s *Session, a []string, out io.Writer) error {
	usage := "wav2p <fita.wav> [pasta] [zx80]"
	if len(a) < 1 {
		return usageErr(usage)
	}
	dir, zx80 := "", false
	for _, arg := range a[1:] {
		switch {
		case strings.EqualFold(arg, "zx80"):
			zx80 = true
		case dir == "":
			dir = arg
		default:
			return usageErr(usage)
		}
	}
	msg, err := WavToPFiles(a[0], dir, zx80)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, msg)
	return nil
}
