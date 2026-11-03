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

// The tape deck (EightyOne's tape manager without its window): what is on
// the tape, a new empty tape, a program or a whole tape (.tzx, .t81) put in
// without loading it, the tape saved as a file, and the Play, Stop and
// rewind buttons. What the ZX81 SAVEs goes on the tape too.

const tapeUsage = "tape [status | new | insert <arquivo> | save <arquivo> | play | stop | rewind]"

func registerTape(r *Registry) {
	r.Register(&Command{Name: "tape", Aliases: []string{"fita"}, Usage: tapeUsage,
		Help: "o toca-fitas: o que há na fita, fita nova, pôr um programa ou uma fita (.tzx, .t81) sem carregar, gravar a fita, play, stop e rebobinar",
		Run:  cmdTape})
}

func cmdTape(s *Session, a []string, out io.Writer) error {
	sub := "status"
	if len(a) > 0 {
		sub = strings.ToLower(a[0])
	}
	d := s.m.Tape
	switch sub {
	case "status", "estado":
		if len(a) > 1 {
			return usageErr(tapeUsage)
		}
		fmt.Fprint(out, s.tapeStatus())
	case "new", "nova":
		d.Player.Stop(false)
		d.Player.NewTape()
		d.Player.CurBlock = 0
		d.In = true
		fmt.Fprintln(out, "fita nova (vazia); o SAVE do ZX81 grava nela (tape save grava o arquivo)")
	case "insert", "inserir":
		if len(a) != 2 {
			return usageErr(tapeUsage)
		}
		return cmdLoad(s, []string{a[1], "noauto"}, out)
	case "save", "gravar", "salvar":
		if len(a) != 2 {
			return usageErr(tapeUsage)
		}
		msg, err := s.saveTape(a[1])
		if err != nil {
			return err
		}
		fmt.Fprintln(out, msg)
	case "play":
		d.In = true
		d.Play()
		fmt.Fprintln(out, "play")
	case "stop":
		d.StopTape()
		fmt.Fprintln(out, "stop")
	case "rewind", "rebobinar":
		d.StopTape()
		d.Player.CurBlock = 0
		fmt.Fprintln(out, "fita no começo")
	default:
		return usageErr(tapeUsage)
	}
	return nil
}

// tapeStatus lists the blocks of the tape, marking the current one.
func (s *Session) tapeStatus() string {
	p := s.m.Tape.Player
	var sb strings.Builder
	state := "parada"
	if p.Playing {
		state = "tocando"
	}
	fmt.Fprintf(&sb, "fita %s, %d bloco(s)\n", state, len(p.Blocks))
	for i, b := range p.Blocks {
		mark := "  "
		if i == p.CurBlock {
			mark = "> "
		}
		fmt.Fprintf(&sb, "%s%d  %s\n", mark, i+1, BlockText(b))
	}
	return sb.String()
}

// saveTape writes the tape: .tzx or .t81 with every block, or the first
// program as .p (without its name) or .p81 (with it), as EightyOne's
// SaveFile.
func (s *Session) saveTape(path string) (string, error) {
	blocks := s.m.Tape.Player.Blocks
	ext := strings.ToLower(filepath.Ext(path))
	var data []byte
	switch ext {
	case ".tzx":
		data = tape.SaveTZX(blocks)
	case ".t81":
		data = tape.SaveT81(blocks)
	case ".p", ".81", ".p81", ".o", ".80":
		for _, b := range blocks {
			if b.ID != tape.BlockGeneral {
				continue
			}
			data = b.Data[:min(b.DataLen, len(b.Data))]
			if ext != ".p81" && !s.m.ZX80ROM() {
				_, prog, ok := tape.SplitSaved(data)
				if !ok {
					return "", fmt.Errorf("o primeiro programa da fita não tem nome")
				}
				data = prog
			}
			break
		}
		if data == nil {
			return "", fmt.Errorf("não há programa na fita")
		}
	default:
		return "", fmt.Errorf("%s: a fita se grava como .tzx, .t81, .p ou .p81", filepath.Base(path))
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("fita gravada em %s (%d bytes)", path, len(data)), nil
}

// BlockText describes a tape block: a ZX81 program shows its name and size.
func BlockText(b tape.Block) string {
	switch b.ID {
	case tape.BlockText:
		return "texto: " + b.Text
	case tape.BlockGeneral:
		if name, prog, ok := tape.SplitSaved(b.Data); ok {
			if name == "" {
				name = "(sem nome)"
			}
			return fmt.Sprintf("programa %s, %d bytes", name, len(prog))
		}
		return fmt.Sprintf("dados, %d bytes", len(b.Data))
	case tape.BlockPause:
		if b.Pause == 0 {
			return "pausa: pare a fita"
		}
		return fmt.Sprintf("pausa de %d ms", b.Pause)
	case tape.BlockMessage:
		return "mensagem: " + b.Text
	case tape.BlockGStart:
		return "grupo: " + b.Text
	case tape.BlockArchive:
		return "informações do arquivo"
	case tape.BlockHWType:
		return "tipo de máquina"
	case tape.BlockCustom:
		return "dados próprios: " + b.Text
	case tape.BlockROM, tape.BlockTurbo, tape.BlockTone, tape.BlockPulse, tape.BlockData, tape.BlockDRec, tape.BlockCSW:
		return fmt.Sprintf("bloco %02Xh (Spectrum ou som; pulado no ZX81)", b.ID)
	}
	return fmt.Sprintf("bloco %02Xh", b.ID)
}
