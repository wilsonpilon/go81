// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package basic_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/basic"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
)

// TestEmptyTemplates: the templates are what the ROM saves on 16K and 1K.
func TestEmptyTemplates(t *testing.T) {
	for _, ram := range []int{16, 0} {
		set := session.DefaultSettings()
		set.RAMK = ram
		s, err := session.New(set)
		if err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(t.TempDir(), "vazio.p")
		for _, c := range []string{"run 200", "save " + file} {
			if err := s.Exec(c, io.Discard); err != nil {
				t.Fatal(err)
			}
		}
		got, _ := os.ReadFile(file)
		if want := basic.EmptyP(ram == 0); !bytes.Equal(got, want) {
			t.Errorf("%dK: ROM saved\n% X\ntemplate\n% X", ram, got, want)
		}
	}
}

func TestParseLine(t *testing.T) {
	cases := []struct {
		in   string
		want []byte
	}{
		{`10 PRINT "OI"`, []byte{0xF5, 0x0B, 0x34, 0x2E, 0x0B}},
		{`20 go to 10`, []byte{0xEC, 0x1D, 0x1C, 0x7E, 0x84, 0x20, 0, 0, 0}},
		{`20 GOSUB 1`, []byte{0xED, 0x1D, 0x7E, 0x81, 0, 0, 0, 0}},
		{`30 LET PRINTER=A1`, []byte{0xF1, 0x35, 0x37, 0x2E, 0x33, 0x39, 0x2A, 0x37, 0x14, 0x26, 0x1D}},
		{`40 IF A<=B THEN STOP`, []byte{0xFA, 0x26, 0xDB, 0x27, 0xDE, 0xE3}},
		{`50 REM  oi PRINT`, []byte{0xEA, 0x00, 0x34, 0x2E, 0x00, 0x35, 0x37, 0x2E, 0x33, 0x39}},
		{`60 PRINT "\"\::%A#"`, []byte{0xF5, 0x0B, 0xC0, 0x80, 0xA6, 0x0C, 0x0B}},
		{`70 PRINT CHR$ 65;INKEY$;PI`, []byte{0xF5, 0xD6, 0x22, 0x21, 0x7E, 0x87, 0x02, 0, 0, 0, 0x19, 0x41, 0x19, 0x42}},
		{`80 LET A=1\{7E}\{81}\{00}\{00}\{00}\{01}`, []byte{0xF1, 0x26, 0x14, 0x1D, 0x7E, 0x81, 0, 0, 0, 1}},
		{`90 POKE\  16514,0`, []byte{0xF4, 0x00, 0x1D, 0x22, 0x21, 0x1D, 0x20, 0x7E, 0x8F, 0x01, 0x04, 0, 0, 0x1A, 0x1C, 0x7E, 0, 0, 0, 0, 0}},
	}
	for _, c := range cases {
		l, err := basic.ParseLine(c.in)
		if err != nil {
			t.Errorf("%s: %v", c.in, err)
			continue
		}
		if !bytes.Equal(l.Body, c.want) {
			t.Errorf("%s\n got % X\nwant % X", c.in, l.Body, c.want)
		}
	}
	for _, bad := range []string{"PRINT 1", "16384 CLS", `10 PRINT "!"`, `10 PRINT \zz`, `10 LET A=1E99`} {
		if _, err := basic.ParseLine(bad); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
}

func TestParseTextErrors(t *testing.T) {
	_, err := basic.ParseText("# comentário\n10 CLS\n\n5 CLS\n")
	if err == nil || !strings.Contains(err.Error(), "linha 4") {
		t.Fatalf("out of order: %v", err)
	}
	lines, err := basic.ParseText("10 PRINT \\\n\"OI\"\n")
	if err != nil || len(lines) != 1 || basic.Text(lines[0].Body) != `PRINT "OI"` {
		t.Fatalf("continuation: %v %v", lines, err)
	}
}

// roundTrip lists a program area and loads the text back.
func roundTrip(t *testing.T, name string, prog []byte) {
	t.Helper()
	lines, err := basic.Lines(prog)
	if err != nil {
		return // a program with tricks the lister cannot follow
	}
	text := basic.List(lines)
	back, err := basic.ParseText(text)
	if err != nil {
		t.Errorf("%s: %v\n%s", name, err, text)
		return
	}
	if len(back) != len(lines) {
		t.Errorf("%s: %d lines back, want %d", name, len(back), len(lines))
		return
	}
	for i := range lines {
		if back[i].Num != lines[i].Num || !bytes.Equal(back[i].Body, lines[i].Body) {
			t.Errorf("%s line %d:\n text %s\n orig % X\n back % X", name, lines[i].Num,
				basic.Text(lines[i].Body), lines[i].Body, back[i].Body)
			return
		}
	}
}

func TestRoundTripSpecialLines(t *testing.T) {
	roundTrip(t, "testLines", flatten(testLines()))
	// A REM ending in spaces, a number with a value that is not its digits.
	extra := [][]byte{
		line(1, basic.TokREM, 0x26, 0, 0),
		line(2, 0xF5, 0x1C, basic.NumMarker, 0x81, 0, 0, 0, 0), // "0" that is 1
		line(3, 0xF5, 0x00, 0x26),                              // stored space
	}
	roundTrip(t, "extra", flatten(extra))
}

// TestRoundTripRealPrograms lists and reloads every .p of the local
// reference folders (skipped when they are not on this computer).
func TestRoundTripRealPrograms(t *testing.T) {
	var files []string
	for _, dir := range []string{"../../81emul", "../../resource"} {
		filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.EqualFold(filepath.Ext(p), ".p") {
				files = append(files, p)
			}
			return nil
		})
	}
	if len(files) == 0 {
		t.Skip("sem programas de referência (81emul/, resource/)")
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		prog, err := basic.ProgramFromP(data)
		if err != nil {
			continue
		}
		roundTrip(t, filepath.Base(f), prog)
	}
	t.Logf("%d programas", len(files))
}

// runInZX81 loads a .p in the emulated ZX81, optionally types keys, and
// returns the screen.
func runInZX81(t *testing.T, p []byte, keys ...string) []string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "t.p")
	os.WriteFile(file, p, 0o644)
	s, _ := session.New(session.DefaultSettings())
	cmds := append([]string{"load " + file}, keys...)
	for _, c := range cmds {
		if err := s.Exec(c, io.Discard); err != nil {
			t.Fatal(err)
		}
	}
	var rows []string
	s.Do(func(m *zx81.Machine) { rows = m.ScreenText() })
	return rows
}

const sample = `# um programa de teste
10 REM OLA
20 FOR I=1 TO 3
30 PRINT I*1.5;" ";
40 NEXT I
50 PRINT
60 PRINT "FIM"; 2**3
`

func TestMadeProgramRuns(t *testing.T) {
	p, err := basic.TextToP(sample, basic.POptions{})
	if err != nil {
		t.Fatal(err)
	}
	// It loads and lists like the source (ROM LIST).
	rows := runInZX81(t, p, "typeln K")
	want := []string{"  10 REM OLA", "  20 FOR I=1 TO 3", `  30 PRINT I*1.5;" ";`, "  40 NEXT I", "  50 PRINT", `  60 PRINT "FIM";2**3`}
	for i, w := range want {
		got := strings.TrimRight(rows[i], " ")
		got = strings.Replace(got, ">", " ", 1) // the current line has ">"
		if got != w {
			t.Errorf("LIST row %d = %q, want %q", i, got, w)
		}
	}
	// It runs.
	rows = runInZX81(t, p, "typeln R", "run 100")
	if r := strings.TrimRight(rows[0], " "); r != "1.5 3 4.5" {
		t.Errorf("RUN row 0 = %q", r)
	}
	if r := strings.TrimRight(rows[1], " "); r != "FIM8" {
		t.Errorf("RUN row 1 = %q", r)
	}
}

func TestAutorun(t *testing.T) {
	p, err := basic.TextToP(sample, basic.POptions{AutoRun: 20})
	if err != nil {
		t.Fatal(err)
	}
	rows := runInZX81(t, p, "run 100") // no RUN typed
	if r := strings.TrimRight(rows[0], " "); r != "1.5 3 4.5" {
		t.Errorf("autorun row 0 = %q", r)
	}
	if _, err := basic.TextToP(sample, basic.POptions{AutoRun: 15}); err == nil {
		t.Error("autorun on a missing line should fail")
	}
	// Without autorun, LOAD just stops.
	p, _ = basic.TextToP(sample, basic.POptions{})
	rows = runInZX81(t, p, "run 100")
	if strings.HasPrefix(rows[0], "1.5") {
		t.Error("the program ran without autorun")
	}
}

// TestCollapsedOn16K: a program with a collapsed display (as saved on a 1K
// ZX81) works on 16K. The first CLS expands the display one space at a
// time, which takes about two seconds on a ZX81 (and here): the listing
// shows after that.
func TestCollapsedOn16K(t *testing.T) {
	p, err := basic.TextToP("10 PRINT 1\n", basic.POptions{Collapsed: true})
	if err != nil {
		t.Fatal(err)
	}
	rows := runInZX81(t, p, "typeln K", "run 200")
	if !strings.Contains(rows[0], "10>PRINT 1") && !strings.Contains(rows[0], "10 PRINT 1") {
		t.Fatalf("listing not shown:\n%s", strings.Join(rows, "\n"))
	}
}
