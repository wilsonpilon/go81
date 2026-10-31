// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package shell

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wilsonpilon/go81/internal/session"
)

func newSession(t *testing.T) *session.Session {
	t.Helper()
	s, err := session.New(session.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestComplete(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "jogo.p"), nil, 0o644)
	os.WriteFile(filepath.Join(dir, "jogo dois.p"), nil, 0o644)
	os.Mkdir(filepath.Join(dir, "fitas"), 0o755)
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)

	c := NewConsole(newSession(t))
	cases := []struct {
		in, want string
		choices  []string
	}{
		{"pe", "peek ", nil},
		{"help sc", "help scr", nil}, // screen, screenshot, script
		{"help scr", "help scr", []string{"screen", "screenshot", "script"}}, // no growth: show choices
		{"load jogo", "load jogo", []string{`"jogo dois.p"`, "jogo.p"}},
		{"load jogo.", "load jogo.p ", nil},
		{"load jogo.p no", "load jogo.p noauto ", nil},
		{"cd f", "cd fitas/", nil},
		{"machine ra", "machine ram=", nil},
		{"run 10; vers", "run 10; version ", nil},
		{"xyz a", "xyz a", nil},
	}
	for _, k := range cases {
		got, ch := c.Complete(k.in)
		if got != k.want || !reflect.DeepEqual(ch, k.choices) {
			t.Errorf("Complete(%q) = %q %q; want %q %q", k.in, got, ch, k.want, k.choices)
		}
	}
}

func TestConsoleHistory(t *testing.T) {
	c := NewConsole(newSession(t))
	done := make(chan Result, 1)
	for _, l := range []string{"echo um", "echo dois"} {
		c.Run(l, func(r Result) { done <- r })
		if r := <-done; r.Output != strings.TrimPrefix(l, "echo ")+"\n" {
			t.Fatalf("%s -> %q", l, r.Output)
		}
	}
	if p := c.Prev("rascunho"); p != "echo dois" {
		t.Fatalf("Prev = %q", p)
	}
	if p := c.Prev(""); p != "echo um" {
		t.Fatalf("Prev 2 = %q", p)
	}
	if p := c.Prev(""); p != "echo um" {
		t.Fatalf("Prev at top = %q", p)
	}
	c.Next("")
	if n := c.Next(""); n != "rascunho" {
		t.Fatalf("Next back to draft = %q", n)
	}
	c.Run("history", func(r Result) { done <- r })
	if r := <-done; !strings.Contains(r.Output, "1  echo um") || !strings.Contains(r.Output, "2  echo dois") {
		t.Fatalf("history = %q", r.Output)
	}
	c.Run("quit", func(r Result) { done <- r })
	if r := <-done; !r.Quit {
		t.Fatal("quit not reported")
	}
}

func TestRunScriptMode(t *testing.T) {
	var out, errb bytes.Buffer
	err := Run(newSession(t), strings.NewReader("echo a\nhistory\nbad\necho b\n"), &out, &errb, Options{})
	if err == nil || !strings.Contains(err.Error(), "linha 3") {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(out.String(), "a\n") || strings.Contains(out.String(), "b\n") {
		t.Fatalf("out = %q", out.String())
	}
}
