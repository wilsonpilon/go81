// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Command zx81diff compares the Go ZX81 machine with EightyOne's zx81.cpp.
//
// It builds test/compat/zx81ref/harness.cpp against the EightyOne sources in
// 81emul/ (needs gcc/g++), runs the standard scenarios on both and reports
// the first frame that differs. With -golden it writes the expected output
// used by the normal Go tests.
//
//	go run ./test/compat/zx81diff
//	go run ./test/compat/zx81diff -golden internal/emu/zx81/testdata/eightyone.txt.gz
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/zx81/zx81ref"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

func main() {
	var (
		root   = flag.String("root", ".", "repository root")
		cc     = flag.String("cc", "gcc", "C compiler")
		cxx    = flag.String("cxx", "g++", "C++ compiler")
		ref    = flag.String("ref", "", "use this harness binary instead of building one")
		rom    = flag.String("rom", "81emul/exe/ROMs/zx81.edition3.rom", "ZX81 ROM")
		rom80  = flag.String("rom80", "81emul/exe/ROMs/zx80.rom", "ZX80 ROM (for the zx80-* scenarios)")
		only   = flag.String("only", "", "run only the scenario with this name")
		golden = flag.String("golden", "", "write expected results to this .txt.gz file")
	)
	flag.Parse()

	romData, err := os.ReadFile(*rom)
	if err != nil {
		fail(err)
	}
	exe := *ref
	if exe == "" {
		if exe, err = build(*root, *cc, *cxx); err != nil {
			fail(err)
		}
	}

	tapeScen, err := zx81ref.TapeScenarios(romData)
	if err != nil {
		fail(err)
	}
	scenarios := append(append(zx81ref.Scenarios(), tapeScen...), zx81ref.ZX80Scenarios()...)
	rom80Data, err := os.ReadFile(*rom80)
	if err != nil {
		fail(err)
	}
	tape80, err := zx81ref.ZX80TapeScenarios(rom80Data)
	if err != nil {
		fail(err)
	}
	scenarios = append(scenarios, tape80...)
	tmpDir, err := os.MkdirTemp("", "go81-zx81diff")
	if err != nil {
		fail(err)
	}
	defer os.RemoveAll(tmpDir)

	var goldenLines []string
	bad := 0
	for _, s := range scenarios {
		if *only != "" && s.Name != *only {
			continue
		}
		tapePath := ""
		if s.TapeData != nil {
			tapePath = filepath.Join(tmpDir, s.TapeName)
			if err := os.WriteFile(tapePath, s.TapeData, 0o644); err != nil {
				fail(err)
			}
		}
		romPath, romBytes := *rom, romData
		if s.ZX80 {
			romPath, romBytes = *rom80, rom80Data
		}
		want, err := runRef(exe, romPath, s, tapePath)
		if err != nil {
			fail(err)
		}
		got, err := zx81ref.Run(romBytes, s)
		if err != nil {
			fail(err)
		}
		diff := -1
		for i := range want {
			if i >= len(got) || got[i] != want[i] {
				diff = i
				break
			}
		}
		if diff < 0 && len(got) != len(want) {
			diff = len(want)
		}
		if diff >= 0 {
			bad++
			fmt.Printf("%-12s DIFFERS at frame %d\n", s.Name, diff)
			if diff < len(want) {
				fmt.Printf("  want %s\n", want[diff])
			}
			if diff < len(got) {
				fmt.Printf("  got  %s\n", got[diff])
			}
			continue
		}
		fmt.Printf("%-12s ok (%d frames)\n", s.Name, len(want))
		for _, l := range want {
			goldenLines = append(goldenLines, s.Name+"\t"+l)
		}
	}
	if *only == "" {
		bad += checkTapeFiles(exe, *rom, tmpDir, scenarios)
	}
	if bad > 0 {
		fmt.Printf("zx81diff: %d scenario(s) differ\n", bad)
		os.Exit(1)
	}
	fmt.Println("zx81diff: all scenarios match EightyOne")
	if *golden != "" {
		if err := writeGolden(*golden, goldenLines); err != nil {
			fail(err)
		}
		fmt.Println("zx81diff: golden written to", *golden)
	}
}

// checkTapeFiles reads each .tzx/.t81 tape of the scenarios with
// EightyOne's LoadFile and writes it back with its SaveFile as .tzx and
// .t81; Go81 must write the same bytes from the same tape. It returns the
// number of differences.
func checkTapeFiles(exe, rom, tmp string, scenarios []zx81ref.Scenario) int {
	bad := 0
	seen := map[string]bool{}
	for _, s := range scenarios {
		if s.TapeData == nil || !tape.IsTapeFile(s.TapeName) || seen[s.TapeName] {
			continue
		}
		seen[s.TapeName] = true
		in := filepath.Join(tmp, s.TapeName)
		if err := os.WriteFile(in, s.TapeData, 0o644); err != nil {
			fail(err)
		}
		blocks, err := tape.LoadTape(s.TapeData)
		if err != nil {
			fail(err)
		}
		for _, ext := range []string{".tzx", ".t81"} {
			out := filepath.Join(tmp, "resave"+ext)
			cmd := exec.Command(exe, rom, "1", "tape="+in, "resave="+out)
			if b, err := cmd.CombinedOutput(); err != nil {
				fail(fmt.Errorf("resave %s: %v %s", s.TapeName, err, b))
			}
			want, err := os.ReadFile(out)
			if err != nil {
				fail(err)
			}
			got := tape.SaveTZX(blocks)
			if ext == ".t81" {
				got = tape.SaveT81(blocks)
			}
			name := s.TapeName + " -> " + ext
			if !bytes.Equal(got, want) {
				bad++
				i := 0
				for i < len(got) && i < len(want) && got[i] == want[i] {
					i++
				}
				fmt.Printf("%-12s DIFFERS at byte %d (Go81 %d bytes, EightyOne %d)\n", name, i, len(got), len(want))
				continue
			}
			fmt.Printf("%-12s ok (%d bytes, same as EightyOne's SaveFile)\n", name, len(want))
		}
	}
	return bad
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "zx81diff:", err)
	os.Exit(2)
}

func build(root, cc, cxx string) (string, error) {
	src := filepath.Join(root, "81emul", "source", "Source")
	if _, err := os.Stat(filepath.Join(src, "zx81", "zx81.cpp")); err != nil {
		return "", fmt.Errorf("EightyOne sources not found in %s", src)
	}
	dir := filepath.Join(root, "test", "compat", "zx81ref")
	tmp := os.TempDir()
	inc := []string{"-I", filepath.Join(dir, "shim"), "-I", src,
		"-I", filepath.Join(src, "zx81"), "-I", filepath.Join(src, "z80")}
	var objs []string
	for _, f := range []string{"z80.c", "z80_ops.c"} {
		o := filepath.Join(tmp, "go81-zx81ref-"+strings.TrimSuffix(f, ".c")+".o")
		args := append([]string{"-O2", "-w", "-c"}, inc...)
		args = append(args, filepath.Join(src, "z80", f), "-o", o)
		if err := run(cc, args...); err != nil {
			return "", err
		}
		objs = append(objs, o)
	}
	// TZXFILE.cpp ends with the tape block editor dialogs (VCL forms); use a
	// copy without them.
	tzxFile, err := filteredTZXFile(filepath.Join(src, "tzx", "TZXFILE.cpp"), tmp)
	if err != nil {
		return "", err
	}
	inc = append(inc, "-I", filepath.Join(src, "tzx"))
	exe := filepath.Join(tmp, "go81-zx81ref.exe")
	args := append([]string{"-O2", "-w", "-fpermissive"}, inc...)
	args = append(args, "-o", exe,
		filepath.Join(dir, "harness.cpp"), filepath.Join(dir, "tape.cpp"), filepath.Join(dir, "sound.cpp"),
		filepath.Join(src, "zx81", "zx81.cpp"), filepath.Join(src, "zx81", "rompatch.cpp"), tzxFile)
	for _, f := range []string{"tzxevent", "tzxEventGeneral", "tzxEventMisc", "tzxEventTurbo", "tzxadd", "tzxload", "tzxsave"} {
		args = append(args, filepath.Join(src, "tzx", f+".cpp"))
	}
	args = append(args, objs...)
	if err := run(cxx, args...); err != nil {
		return "", err
	}
	return exe, nil
}

func filteredTZXFile(path, tmp string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	s := string(b)
	i := strings.Index(s, "void TTZXFile::EditBlock")
	if i < 0 {
		return "", fmt.Errorf("%s: EditBlock not found", path)
	}
	var out []string
	for _, l := range strings.Split(s[:i], "\n") {
		if !strings.Contains(l, `#include "tzxEdit`) {
			out = append(out, l)
		}
	}
	dst := filepath.Join(tmp, "go81-TZXFILE.cpp")
	return dst, os.WriteFile(dst, []byte(strings.Join(out, "\n")), 0o644)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func runRef(exe, rom string, s zx81ref.Scenario, tapePath string) ([]string, error) {
	cmd := exec.Command(exe, append([]string{rom}, s.Args(tapePath)...)...)
	cmd.Stdin = strings.NewReader(s.KeyInput())
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("harness (%s): %w", s.Name, err)
	}
	var lines []string
	for _, l := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

func writeGolden(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw, _ := gzip.NewWriterLevel(f, gzip.BestCompression)
	w := bufio.NewWriter(zw)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return zw.Close()
}
