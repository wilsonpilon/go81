// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"io"
	"strings"

	"github.com/wilsonpilon/go81/internal/media/tape"
)

// The tape deck (EightyOne's tape manager without its window): what is on
// the tape, a new empty tape, a program put on the tape without loading it,
// and the Play, Stop and rewind buttons.

const tapeUsage = "tape [status | new | insert <programa> | play | stop | rewind]"

func registerTape(r *Registry) {
	r.Register(&Command{Name: "tape", Aliases: []string{"fita"}, Usage: tapeUsage,
		Help: "o toca-fitas: o que há na fita, fita nova, pôr um programa na fita sem carregar, play, stop e rebobinar",
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
		fmt.Fprintln(out, "fita nova (vazia); o SAVE do ZX81 continua gravando no Go81")
	case "insert", "inserir":
		if len(a) != 2 {
			return usageErr(tapeUsage)
		}
		return cmdLoad(s, []string{a[1], "noauto"}, out)
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
		return fmt.Sprintf("pausa de %d ms", b.Pause)
	}
	return fmt.Sprintf("bloco %02Xh", b.ID)
}
