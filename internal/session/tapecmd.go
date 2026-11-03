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

const tapeUsage = "tape [status | new | insert <arquivo> [som] | save <arquivo> | play | stop | rewind | origem tzx|wav | destino tzx|wav|som | ir N | apagar N | mover N M | renomear N NOME | rapida sim|nao | automatica sim|nao]"

// blockArg reads the block numbers of a tape subcommand (counted from 1)
// as indexes from 0; want is len(a) with them.
func blockArg(a []string, want int) ([]int, error) {
	if len(a) != want {
		return nil, usageErr(tapeUsage)
	}
	var n []int
	for _, s := range a[1:] {
		v, err := ParseNumber(s)
		if err != nil || v < 1 {
			return nil, fmt.Errorf("%q não é um número de bloco (o primeiro é 1)", s)
		}
		n = append(n, v-1)
	}
	return n, nil
}

func yesNo(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "sim", "s", "on", "yes", "1", "liga", "ligada":
		return true, nil
	case "nao", "não", "n", "off", "no", "0", "desliga", "desligada":
		return false, nil
	}
	return false, fmt.Errorf("%q: use sim ou nao", s)
}

func ligada(on bool) string {
	if on {
		return "ligada"
	}
	return "desligada"
}

// flashSave makes SAVE go to the session for the save command, whatever
// the tape output is, and gives back what puts it as it was.
func (s *Session) flashSave() func() {
	d := s.m.Tape
	out := d.Out
	d.Out = true
	return func() {
		s.saveCmd = false
		d.Out = out
	}
}

// outputNames are the names of tape destino.
var outputNames = map[tape.Output]string{
	tape.OutTape:  "o gerenciador de fitas (TZX): o SAVE vira um bloco da fita, na hora",
	tape.OutWav:   "o carregador WAV: o SAVE é gravado como som, em tempo real (tape save x.wav grava o som)",
	tape.OutAudio: "a saída de áudio: o SAVE toca em tempo real (com o som pelo vídeo ligado), sem gravar",
}

// tapeSource is tape origem: the tape manager or the WAV loader (EightyOne:
// Options → Tape → Source).
func (s *Session) tapeSource(arg string, out io.Writer) error {
	d := s.m.Tape
	switch strings.ToLower(arg) {
	case "tzx", "fita", "gerenciador":
		d.StopTape()
		if d.Wav != nil {
			d.Wav.Stop()
		}
		d.In = true
		fmt.Fprintln(out, "origem da fita: o gerenciador de fitas (TZX)")
	case "wav", "som":
		if d.Wav == nil {
			return fmt.Errorf("não há som de fita no carregador WAV: tape insert fita.wav som")
		}
		d.StopTape()
		d.In = false
		fmt.Fprintln(out, "origem da fita: o carregador WAV (o som, em tempo real)")
	default:
		return usageErr("tape origem tzx|wav")
	}
	return nil
}

// tapeDestination is tape destino (EightyOne: Options → Tape → Destination).
func (s *Session) tapeDestination(arg string, out io.Writer) error {
	var o tape.Output
	switch strings.ToLower(arg) {
	case "tzx", "fita", "gerenciador":
		o = tape.OutTape
	case "wav":
		o = tape.OutWav
	case "som", "audio", "áudio":
		o = tape.OutAudio
	default:
		return usageErr("tape destino tzx|wav|som")
	}
	s.m.Tape.SetOutput(o)
	s.tapeOut, s.tapeOutSet = o, true
	fmt.Fprintln(out, "destino do SAVE: "+outputNames[o])
	return nil
}

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
	case "origem", "source":
		if len(a) != 2 {
			return usageErr("tape origem tzx|wav")
		}
		return s.tapeSource(a[1], out)
	case "destino", "destination":
		if len(a) != 2 {
			return usageErr("tape destino tzx|wav|som")
		}
		return s.tapeDestination(a[1], out)
	case "ir", "select", "goto":
		n, err := blockArg(a, 2)
		if err != nil {
			return err
		}
		if err := d.Player.SelectBlock(n[0]); err != nil {
			return err
		}
		fmt.Fprintf(out, "bloco atual: %d (o próximo LOAD começa nele)\n", n[0]+1)
	case "apagar", "delete":
		n, err := blockArg(a, 2)
		if err != nil {
			return err
		}
		if err := d.Player.DeleteBlock(n[0]); err != nil {
			return err
		}
		fmt.Fprintf(out, "bloco %d apagado; a fita tem %d\n", n[0]+1, len(d.Player.Blocks))
	case "mover", "move":
		n, err := blockArg(a, 3)
		if err != nil {
			return err
		}
		if err := d.Player.MoveBlock(n[0], n[1]); err != nil {
			return err
		}
		fmt.Fprintf(out, "bloco %d agora é o %d\n", n[0]+1, n[1]+1)
	case "renomear", "rename":
		if len(a) != 3 {
			return usageErr("tape renomear N NOME")
		}
		n, err := blockArg(a[:2], 2)
		if err != nil {
			return err
		}
		if err := d.Player.RenameProgram(n[0], a[2]); err != nil {
			return err
		}
		fmt.Fprintf(out, "bloco %d: %s\n", n[0]+1, BlockText(d.Player.Blocks[n[0]]))
	case "rapida", "rápida", "flash", "automatica", "automática", "autostart":
		if len(a) != 2 {
			return usageErr("tape " + sub + " sim|nao")
		}
		on, err := yesNo(a[1])
		if err != nil {
			return err
		}
		if strings.HasPrefix(sub, "r") || sub == "flash" {
			d.Player.FlashLoad = on
			fmt.Fprintln(out, "carga rápida:", ligada(on))
		} else {
			d.Player.AutoStart = on
			fmt.Fprintln(out, "partida automática:", ligada(on))
		}
	case "new", "nova":
		d.Player.Stop(false)
		d.Player.NewTape()
		d.Player.CurBlock = 0
		d.In = true
		if d.Rec != nil {
			d.Rec.Clear()
		}
		fmt.Fprintln(out, "fita nova (vazia); o SAVE do ZX81 grava nela (tape save grava o arquivo)")
	case "insert", "inserir":
		if len(a) == 3 && (strings.EqualFold(a[2], "som") || strings.EqualFold(a[2], "sound")) {
			return s.insertWavSound(a[1], out)
		}
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
		if d.Wav != nil {
			d.Wav.Rewind()
		}
		fmt.Fprintln(out, "fita no começo")
	default:
		return usageErr(tapeUsage)
	}
	return nil
}

// insertWavSound makes a WAV the input of the deck, played as sound in
// real time: the ROM reads the bits itself, as from a cassette (EightyOne's
// WAV loader). Slow, but it loads what the conversion does not (turbo
// loaders, tapes in poor shape).
func (s *Session) insertWavSound(path string, out io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	wp, err := tape.NewWavPlayer(data)
	if err != nil {
		return fmt.Errorf("%s: %v", filepath.Base(path), err)
	}
	s.queue, s.hold, s.gap = nil, 0, 0
	s.m.Tape.InsertWav(wp)
	_, length := wp.Position()
	fmt.Fprintf(out, "som da fita %s (%.0f s) na entrada do ZX81; digite LOAD \"\" (carrega em tempo real)", filepath.Base(path), length)
	fmt.Fprintln(out)
	return nil
}

// tapeStatus lists the blocks of the tape, marking the current one.
func (s *Session) tapeStatus() string {
	p := s.m.Tape.Player
	var sb strings.Builder
	if w := s.m.Tape.Wav; w != nil && !s.m.Tape.In {
		at, length := w.Position()
		state := "parado"
		if w.Playing() {
			state = "tocando"
		}
		fmt.Fprintf(&sb, "som de fita (WAV) %s: %.1f de %.1f s; o SAVE vai para a fita abaixo", state, at, length)
		sb.WriteByte(10)
	}
	switch r := s.m.Tape.Rec; s.m.Tape.Output() {
	case tape.OutWav:
		state := ""
		if r.Recording() {
			state = ", gravando"
		}
		fmt.Fprintf(&sb, "o SAVE é gravado como som: %.1f s%s (tape save x.wav)\n", r.Seconds(), state)
	case tape.OutAudio:
		sb.WriteString("o SAVE vai para a saída de áudio, em tempo real\n")
	}
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
	case ".wav":
		data = tape.SaveWAV(blocks)
		if r := s.m.Tape.Rec; s.m.Tape.Output() == tape.OutWav && r.Seconds() > 0 {
			data = r.WAV() // what the ROM saved, as recorded
		}
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
		return "", fmt.Errorf("%s: a fita se grava como .tzx, .t81, .wav, .p ou .p81", filepath.Base(path))
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
