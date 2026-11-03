// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tape new empties the tape, tape insert puts a program on it without
// loading it, and LOAD "" in the ZX81 takes it from there.
func TestTapeCommand(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "fita.txt")
	if err := os.WriteFile(in, []byte("10 PRINT \"FITA\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newBooted(t)
	mustExec(t, s, "tape new")
	if out := mustExec(t, s, "tape"); !strings.Contains(out, "1 bloco(s)") {
		t.Fatalf("tape = %q", out)
	}
	mustExec(t, s, `txt2p "`+in+`"`)
	if out := mustExec(t, s, `tape insert "`+filepath.Join(dir, "fita.p")+`"`); !strings.Contains(out, `LOAD ""`) {
		t.Fatalf("tape insert = %q", out)
	}
	if out := mustExec(t, s, "tape"); !strings.Contains(out, "programa FITA") {
		t.Fatalf("tape = %q", out)
	}
	mustExec(t, s, `typeln J""`) // LOAD "" (J is LOAD in K mode)
	mustExec(t, s, "run 400")
	if out := mustExec(t, s, "list"); !strings.Contains(out, `10 PRINT "FITA"`) {
		t.Fatalf("after LOAD: %q", out)
	}
	if err := s.Exec("tape voar", nil); err == nil {
		t.Fatal("tape voar accepted")
	}
}

// A tape saved as sound (.wav) comes back: load reads its programs (as
// tape blocks, flash load), wav2p writes them as .p, and "tape insert ...
// som" plays it to the ZX81 in real time, the ROM reading the bits.
func TestTapeWAV(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "oi.txt")
	os.WriteFile(in, []byte("10 PRINT \"OI DO WAV\"\n"), 0o644)
	s := newBooted(t)
	mustExec(t, s, `txt2p "`+in+`"`)
	mustExec(t, s, `load "`+filepath.Join(dir, "oi.p")+`"`)
	wav := filepath.Join(dir, "fita.wav")
	mustExec(t, s, `tape save "`+wav+`"`)
	out := mustExec(t, s, `wav2p "`+wav+`" "`+dir+`"`)
	if !strings.Contains(out, "programa OI") {
		t.Fatalf("wav2p: %q", out)
	}
	orig, _ := os.ReadFile(filepath.Join(dir, "oi.p"))
	back, _ := os.ReadFile(filepath.Join(dir, "OI.p"))
	if !bytes.Equal(orig, back) {
		t.Fatal("wav2p: the .p differs")
	}
	if text, _, err := ProgramText(wav); err != nil || text != "10 PRINT \"OI DO WAV\"\n" {
		t.Fatalf("p2txt of the WAV: %q %v", text, err)
	}
	mustExec(t, s, "reset")
	mustExec(t, s, `load "`+wav+`"`)
	if out := mustExec(t, s, "list"); !strings.Contains(out, `OI DO WAV`) {
		t.Fatalf("load .wav: %q", out)
	}
	// In real time: about 6 s of sound for this program.
	mustExec(t, s, "reset")
	mustExec(t, s, "run 150")
	if out := mustExec(t, s, `tape insert "`+wav+`" som`); !strings.Contains(out, "tempo real") {
		t.Fatalf("tape insert som: %q", out)
	}
	mustExec(t, s, `typeln J""`)
	for i := 0; i < 40; i++ {
		mustExec(t, s, "run 50")
		if out := mustExec(t, s, "list"); strings.Contains(out, `OI DO WAV`) {
			if st := mustExec(t, s, "tape"); !strings.Contains(st, "som de fita (WAV)") {
				t.Fatalf("tape status: %q", st)
			}
			return
		}
	}
	t.Fatalf("the ROM did not load the WAV in real time:\n%s", mustExec(t, s, "tape"))
}

// What the ZX81 SAVEs goes on the tape; tape save writes it as .tzx and
// .t81; a tape of two programs loads the second by its name; load and
// p2txt read .tzx and .t81.
func TestTapeFiles(t *testing.T) {
	dir := t.TempDir()
	s := newBooted(t)
	in := filepath.Join(dir, "dois.txt")
	os.WriteFile(in, []byte("10 PRINT \"DOIS\"\n"), 0o644)
	mustExec(t, s, `load "`+in+`"`)
	// A new tape, two SAVEs.
	mustExec(t, s, "tape new")
	mustExec(t, s, `save "`+filepath.Join(dir, "a.p")+`" DOIS`)
	mustExec(t, s, `save "`+filepath.Join(dir, "b.p")+`" OUTRO`)
	if out := mustExec(t, s, "tape"); !strings.Contains(out, "programa DOIS") || !strings.Contains(out, "programa OUTRO") || !strings.Contains(out, "3 bloco(s)") {
		t.Fatalf("tape after SAVE:\n%s", out)
	}
	tzx, t81 := filepath.Join(dir, "fita.tzx"), filepath.Join(dir, "fita.t81")
	mustExec(t, s, `tape save "`+tzx+`"`)
	mustExec(t, s, `tape save "`+t81+`"`)
	mustExec(t, s, `tape save "`+filepath.Join(dir, "primeiro.p")+`"`)
	first, _ := os.ReadFile(filepath.Join(dir, "primeiro.p"))
	saved, _ := os.ReadFile(filepath.Join(dir, "a.p"))
	if len(first) == 0 || !bytes.Equal(first, saved) {
		t.Fatal("tape save .p: not the first program")
	}
	// Both files list the two programs, and the editor reads the first.
	for _, f := range []string{tzx, t81} {
		text, _, err := ProgramText(f)
		want := "# programa DOIS (1 de 2)\n10 PRINT \"DOIS\"\n\n# programa OUTRO (2 de 2)\n10 PRINT \"DOIS\"\n"
		if err != nil || text != want {
			t.Fatalf("%s: %q %v", f, text, err)
		}
		if lines, err := ReadProgramLines(f); err != nil || len(lines) != 1 {
			t.Fatalf("editor %s: %v", f, err)
		}
	}
	// p2txt writes one text per program, and txt2p reads the text back.
	out := mustExec(t, s, `p2txt "`+tzx+`"`)
	if !strings.Contains(out, "2 programas") {
		t.Fatalf("p2txt of a tape: %q", out)
	}
	for _, name := range []string{"fita-DOIS.txt", "fita-OUTRO.txt"} {
		if text, err := os.ReadFile(filepath.Join(dir, name)); err != nil || string(text) != "10 PRINT \"DOIS\"\n" {
			t.Fatalf("%s: %q %v", name, text, err)
		}
	}
	mustExec(t, s, `p2txt "`+t81+`" "`+filepath.Join(dir, "lista.txt")+`"`)
	if _, err := os.Stat(filepath.Join(dir, "lista-OUTRO.txt")); err != nil {
		t.Fatal(err)
	}
	// A new ZX81, the tape in: LOAD "OUTRO" skips DOIS.
	mustExec(t, s, "reset")
	mustExec(t, s, "run 150")
	if out := mustExec(t, s, `tape insert "`+tzx+`"`); !strings.Contains(out, "3 blocos") {
		t.Fatalf("tape insert: %q", out)
	}
	mustExec(t, s, `typeln J"OUTRO"`)
	mustExec(t, s, "run 900")
	if out := mustExec(t, s, "list"); !strings.Contains(out, `10 PRINT "DOIS"`) {
		t.Fatalf("after LOAD \"OUTRO\": %q", out)
	}
	if out := mustExec(t, s, "tape"); !strings.Contains(out, "> 3") && !strings.Contains(out, "fita parada") {
		t.Fatalf("tape after LOAD:\n%s", out)
	}
	// load of a .t81 loads its first program by itself.
	mustExec(t, s, "reset")
	mustExec(t, s, `load "`+t81+`"`)
	if out := mustExec(t, s, "list"); !strings.Contains(out, `10 PRINT "DOIS"`) {
		t.Fatalf("load .t81: %q", out)
	}
}

// tape destino wav: SAVE runs in real time and is recorded as sound; the
// sound reads back as the same program. tape origem and speed too.
func TestTapeDestination(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "som.txt")
	os.WriteFile(in, []byte("10 PRINT \"SOM REAL\"\n"), 0o644)
	s := newBooted(t)
	mustExec(t, s, `load "`+in+`"`)
	if out := mustExec(t, s, "tape destino wav"); !strings.Contains(out, "carregador WAV") {
		t.Fatalf("tape destino: %q", out)
	}
	mustExec(t, s, `typeln S"SOM"`) // SAVE "SOM"
	for i := 0; i < 30 && !strings.Contains(mustExec(t, s, "tape"), "gravando"); i++ {
		mustExec(t, s, "run 10")
	}
	for i := 0; i < 60 && strings.Contains(mustExec(t, s, "tape"), "gravando"); i++ {
		mustExec(t, s, "run 50")
	}
	st := mustExec(t, s, "tape")
	// 2 blocks: the program load put on the tape; the SAVE is only sound.
	if strings.Contains(st, "gravando") || !strings.Contains(st, "gravado como som") || !strings.Contains(st, "2 bloco(s)") {
		t.Fatalf("tape after SAVE:\n%s", st)
	}
	wav := filepath.Join(dir, "real.wav")
	mustExec(t, s, `tape save "`+wav+`"`)
	out := mustExec(t, s, `wav2p "`+wav+`" "`+dir+`"`)
	if !strings.Contains(out, "programa SOM") {
		t.Fatalf("wav2p of the recorded SAVE: %q", out)
	}
	if text, _, err := ProgramText(filepath.Join(dir, "SOM.p")); err != nil || text != "10 PRINT \"SOM REAL\"\n" {
		t.Fatalf("recorded program: %q %v", text, err)
	}
	// The save command still writes the file (flash), and the choice stays.
	mustExec(t, s, `save "`+filepath.Join(dir, "x.p")+`"`)
	if out := mustExec(t, s, "tape"); !strings.Contains(out, "gravado como som") {
		t.Fatalf("destination lost: %q", out)
	}
	mustExec(t, s, "tape destino tzx")
	if err := s.Exec("tape origem wav", nil); err == nil {
		t.Fatal("tape origem wav without a WAV")
	}
	mustExec(t, s, "tape origem tzx")
	if out := mustExec(t, s, "speed 50"); !strings.Contains(out, "50%") {
		t.Fatalf("speed: %q", out)
	}
	if n := s.SpeedFrames() + s.SpeedFrames(); n != 1 {
		t.Fatalf("50%%: %d frames in two ticks", n)
	}
	mustExec(t, s, "speed 300")
	if n := s.SpeedFrames(); n != 3 {
		t.Fatalf("300%%: %d", n)
	}
	if err := s.Exec("speed 5", nil); err == nil {
		t.Fatal("speed 5 accepted")
	}
}

// The tape manager commands: go to, move, rename, delete, flash load.
func TestTapeEditing(t *testing.T) {
	dir := t.TempDir()
	s := newBooted(t)
	in := filepath.Join(dir, "um.txt")
	os.WriteFile(in, []byte("10 PRINT \"UM\"\n"), 0o644)
	mustExec(t, s, `load "`+in+`"`)
	// Build a tape of three programs with SAVE.
	mustExec(t, s, "tape new")
	for _, n := range []string{"UM", "DOIS", "TRES"} {
		mustExec(t, s, `save "`+filepath.Join(dir, n+".p")+`" `+n)
	}
	st := mustExec(t, s, "tape")
	if !strings.Contains(st, "4 bloco(s)") {
		t.Fatalf("tape:\n%s", st)
	}
	// SAVE records at the current block: UM, DOIS, TRES, then the text block.
	mustExec(t, s, "tape mover 3 1") // TRES first
	mustExec(t, s, `tape renomear 1 "ULTIMO"`)
	mustExec(t, s, "tape apagar 2") // UM
	mustExec(t, s, "tape rapida nao")
	mustExec(t, s, "tape ir 2")
	st = mustExec(t, s, "tape")
	if !strings.Contains(st, "  1  programa ULTIMO") || strings.Contains(st, "programa UM") || !strings.Contains(st, "> 2  programa DOIS") {
		t.Fatalf("tape after editing:\n%s", st)
	}
	if err := s.Exec("tape apagar 9", nil); err == nil {
		t.Fatal("tape apagar 9 accepted")
	}
	if err := s.Exec("tape renomear 3 X", nil); err == nil {
		t.Fatal("renamed a text block")
	}
	// LOAD "ULTIMO" finds the renamed program (TRES), loading at normal speed.
	mustExec(t, s, "tape rapida sim")
	tzx := filepath.Join(dir, "editada.tzx")
	mustExec(t, s, `tape save "`+tzx+`"`)
	mustExec(t, s, "reset")
	mustExec(t, s, "run 150")
	mustExec(t, s, `tape insert "`+tzx+`"`)
	mustExec(t, s, `typeln J"ULTIMO"`)
	mustExec(t, s, "run 900")
	if out := mustExec(t, s, "list"); !strings.Contains(out, `PRINT "`) {
		t.Fatalf("LOAD \"ULTIMO\": %q", out)
	}
}
