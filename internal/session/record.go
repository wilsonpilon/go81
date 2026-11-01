// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/wav"
)

// recording writes the sound of every frame the ZX81 runs to a .wav file:
// the sound of emulated time, so turbo or a script that runs frames
// without waiting records the same as real time.
type recording struct {
	path string
	w    *wav.Writer
	err  error // the write that stopped the recording
}

func registerRecord(r *Registry) {
	r.Register(&Command{Name: "record", Aliases: []string{"gravar-som"}, Usage: "record <arquivo.wav> | stop",
		Help: "grava o som do ZX81 (o som pelo vídeo) num arquivo .wav; sem argumento, mostra a gravação", Run: cmdRecord})
}

// recordFrame adds the sound of the frame just run. Caller holds s.mu.
func (s *Session) recordFrame() {
	r := s.rec
	if r == nil || r.err != nil {
		return
	}
	if err := r.w.Write(s.m.Audio()); err != nil {
		r.err = err
		return
	}
	if s.frames%50 == 0 { // keep the file valid about once a second
		if err := r.w.Sync(); err != nil {
			r.err = err
		}
	}
}

// StartRecording starts writing the sound to a .wav file.
func (s *Session) StartRecording(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.startRecording(path)
}

func (s *Session) startRecording(path string) error {
	if s.rec != nil {
		return fmt.Errorf("já gravando em %s (record stop para parar)", s.rec.path)
	}
	if !strings.EqualFold(filepath.Ext(path), ".wav") {
		return errors.New("o arquivo do som deve ter a extensão .wav")
	}
	w, err := wav.Create(path, zx81.SampleRate, zx81.AudioChannels)
	if err != nil {
		return err
	}
	s.rec = &recording{path: path, w: w}
	return nil
}

// StopRecording ends the recording and returns a description of it ("" if
// nothing was being recorded).
func (s *Session) StopRecording() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopRecording()
}

func (s *Session) stopRecording() (string, error) {
	r := s.rec
	if r == nil {
		return "", nil
	}
	s.rec = nil
	err := r.w.Close()
	if r.err != nil {
		err = r.err
	}
	return fmt.Sprintf("%s: %.1f s de som", r.path, r.w.Seconds()), err
}

// Recording returns the file being recorded and its length in seconds
// ("" when there is none).
func (s *Session) Recording() (string, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rec == nil {
		return "", 0
	}
	return s.rec.path, s.rec.w.Seconds()
}

func cmdRecord(s *Session, a []string, out io.Writer) error {
	switch {
	case len(a) == 0:
		if s.rec == nil {
			fmt.Fprintln(out, "nenhuma gravação de som")
			return nil
		}
		fmt.Fprintf(out, "gravando em %s: %.1f s\n", s.rec.path, s.rec.w.Seconds())
		return s.rec.err
	case len(a) == 1 && strings.EqualFold(a[0], "stop"):
		msg, err := s.stopRecording()
		if msg == "" {
			fmt.Fprintln(out, "nenhuma gravação de som")
			return nil
		}
		fmt.Fprintln(out, "gravado", msg)
		return err
	case len(a) == 1:
		if err := s.startRecording(a[0]); err != nil {
			return err
		}
		fmt.Fprintln(out, "gravando o som em", a[0])
		if !s.set.Sound {
			fmt.Fprintln(out, "aviso: o som pelo vídeo está desligado (machine sound=on); o arquivo terá silêncio")
		}
		return nil
	}
	return usageErr("record <arquivo.wav> | stop")
}
