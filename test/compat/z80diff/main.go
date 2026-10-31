// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Command z80diff compares the Go Z80 core with the original EightyOne core.
//
// It builds test/compat/z80ref/harness.c against the EightyOne sources in
// 81emul/ (needs gcc), runs the same cases on both cores and reports any
// difference. With -golden it also writes the expected results, which the
// normal Go tests use when 81emul/ and gcc are not available.
//
//	go run ./test/compat/z80diff                 # compare
//	go run ./test/compat/z80diff -per 50 -runs 500
//	go run ./test/compat/z80diff -golden internal/emu/z80/testdata/eightyone.txt.gz
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

	"github.com/wilsonpilon/go81/internal/emu/z80/refcase"
)

func main() {
	var (
		root    = flag.String("root", ".", "repository root")
		cc      = flag.String("cc", "gcc", "C compiler")
		ref     = flag.String("ref", "", "use this harness binary instead of building one")
		seed    = flag.Uint64("seed", 81, "random seed")
		per     = flag.Int("per", 20, "random states per opcode")
		runs    = flag.Int("runs", 200, "random code runs")
		runLen  = flag.Int("runlen", 2000, "instructions per run")
		golden  = flag.String("golden", "", "write expected results to this .txt.gz file")
		maxShow = flag.Int("show", 5, "mismatches to show in detail")
	)
	flag.Parse()

	exe := *ref
	if exe == "" {
		var err error
		exe, err = build(*root, *cc)
		if err != nil {
			fail(err)
		}
	}

	cases := refcase.Cases(*seed, *per, *runs, *runLen)
	fmt.Printf("z80diff: %d cases\n", len(cases))

	want, err := runRef(exe, cases, false)
	if err != nil {
		fail(err)
	}

	r := refcase.NewRunner()
	bad := 0
	for i, tc := range cases {
		got, _ := r.Run(tc, false)
		if got == want[i] {
			continue
		}
		bad++
		if bad <= *maxShow {
			show(exe, r, tc, want[i], got)
		}
	}
	if bad > 0 {
		fmt.Printf("z80diff: %d of %d cases DIFFER\n", bad, len(cases))
		os.Exit(1)
	}
	fmt.Println("z80diff: all cases match EightyOne")

	if *golden != "" {
		if err := writeGolden(*golden, cases, want); err != nil {
			fail(err)
		}
		fmt.Println("z80diff: golden written to", *golden)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "z80diff:", err)
	os.Exit(2)
}

func build(root, cc string) (string, error) {
	src := filepath.Join(root, "81emul", "source", "Source", "z80")
	if _, err := os.Stat(filepath.Join(src, "z80_ops.c")); err != nil {
		return "", fmt.Errorf("EightyOne sources not found in %s", src)
	}
	dir := filepath.Join(root, "test", "compat", "z80ref")
	out := filepath.Join(os.TempDir(), "go81-z80ref.exe")
	cmd := exec.Command(cc, "-O2", "-w",
		"-I", filepath.Join(dir, "shim"), "-I", src,
		"-o", out,
		filepath.Join(dir, "harness.c"),
		filepath.Join(src, "z80.c"), filepath.Join(src, "z80_ops.c"))
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("building harness: %w", err)
	}
	return out, nil
}

func runRef(exe string, cases []refcase.Case, verbose bool) ([]string, error) {
	var in strings.Builder
	for _, tc := range cases {
		in.WriteString(tc.String())
		in.WriteByte('\n')
	}
	args := []string{}
	if verbose {
		args = append(args, "-v")
	}
	cmd := exec.Command(exe, args...)
	cmd.Stdin = strings.NewReader(in.String())
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running harness: %w", err)
	}
	var lines []string
	for _, l := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	if !verbose && len(lines) != len(cases) {
		return nil, fmt.Errorf("harness returned %d lines for %d cases", len(lines), len(cases))
	}
	return lines, nil
}

func show(exe string, r *refcase.Runner, tc refcase.Case, want, got string) {
	fmt.Printf("\ncase  %s\nwant  %s\ngot   %s\n", tc, want, got)
	refLog, err := runRef(exe, []refcase.Case{tc}, true)
	if err == nil {
		fmt.Println("EightyOne bus log:")
		for _, l := range refLog {
			if strings.HasPrefix(l, "  ") {
				fmt.Println(l)
			}
		}
	}
	_, goLog := r.Run(tc, true)
	fmt.Print("Go bus log:\n", goLog)
}

func writeGolden(path string, cases []refcase.Case, want []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw, _ := gzip.NewWriterLevel(f, gzip.BestCompression)
	w := bufio.NewWriter(zw)
	for i, tc := range cases {
		fmt.Fprintf(w, "%s\t%s\n", tc, want[i])
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return zw.Close()
}
