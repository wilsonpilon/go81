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
	"compress/gzip"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/zx81/zx81ref"
)

func main() {
	var (
		root   = flag.String("root", ".", "repository root")
		cc     = flag.String("cc", "gcc", "C compiler")
		cxx    = flag.String("cxx", "g++", "C++ compiler")
		ref    = flag.String("ref", "", "use this harness binary instead of building one")
		rom    = flag.String("rom", "81emul/exe/ROMs/zx81.edition3.rom", "ZX81 ROM")
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

	var goldenLines []string
	bad := 0
	for _, s := range zx81ref.Scenarios() {
		if *only != "" && s.Name != *only {
			continue
		}
		want, err := runRef(exe, *rom, s)
		if err != nil {
			fail(err)
		}
		got, err := zx81ref.Run(romData, s)
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
	exe := filepath.Join(tmp, "go81-zx81ref.exe")
	args := append([]string{"-O2", "-w", "-fpermissive"}, inc...)
	args = append(args, "-o", exe, filepath.Join(dir, "harness.cpp"), filepath.Join(src, "zx81", "zx81.cpp"))
	args = append(args, objs...)
	if err := run(cxx, args...); err != nil {
		return "", err
	}
	return exe, nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func runRef(exe, rom string, s zx81ref.Scenario) ([]string, error) {
	cmd := exec.Command(exe, append([]string{rom}, s.Args()...)...)
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
