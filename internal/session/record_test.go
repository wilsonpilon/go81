// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

func TestRecord(t *testing.T) {
	set := DefaultSettings()
	set.Sound = true
	s, err := New(set)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "som.wav")
	if out := mustExec(t, s, "record"); !strings.Contains(out, "nenhuma") {
		t.Fatalf("status = %q", out)
	}
	if out := mustExec(t, s, "record "+path); strings.Contains(out, "aviso") {
		t.Fatalf("start = %q", out)
	}
	if err := s.Exec("record outro.wav", new(strings.Builder)); err == nil {
		t.Fatal("second recording accepted")
	}
	mustExec(t, s, "run 100") // 2 s in PAL
	if out := mustExec(t, s, "record"); !strings.Contains(out, "2.0 s") {
		t.Fatalf("status = %q", out)
	}
	if out := mustExec(t, s, "record stop"); !strings.Contains(out, "2.0 s de som") {
		t.Fatalf("stop = %q", out)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data := int(binary.LittleEndian.Uint32(b[40:]))
	if want := 100 * zx81.SampleRate / 50 * zx81.AudioChannels * 2; data != want || len(b) != 44+want {
		t.Fatalf("data %d, file %d, want %d", data, len(b), want)
	}
	loud := 0
	for i := 44; i+1 < len(b); i += 2 {
		if int16(binary.LittleEndian.Uint16(b[i:])) != 0 {
			loud++
		}
	}
	if loud == 0 {
		t.Fatal("the ZX81 makes noise with the video sound on, but the file is silent")
	}
	if out := mustExec(t, s, "record stop"); !strings.Contains(out, "nenhuma") {
		t.Fatalf("second stop = %q", out)
	}
	if err := s.Exec("record som.mp3", new(strings.Builder)); err == nil {
		t.Fatal(".mp3 accepted")
	}
	s2 := newBooted(t)
	if out := mustExec(t, s2, "record "+filepath.Join(t.TempDir(), "mudo.wav")); !strings.Contains(out, "aviso") {
		t.Fatalf("no warning without sound: %q", out)
	}
	s2.StopRecording()
}
