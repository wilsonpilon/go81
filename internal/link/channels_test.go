// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package link

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/pkg/go81link"
)

func newSess(t *testing.T) *session.Session {
	t.Helper()
	s, err := session.New(session.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// talk runs a few commands through a client and checks the answers.
func talk(t *testing.T, c *go81link.Client) {
	t.Helper()
	defer c.Close()
	if _, err := c.Cmd("poke 30000 42"); err != nil {
		t.Fatal(err)
	}
	out, err := c.Cmd("peek 30000")
	if err != nil || !strings.HasPrefix(out, "7530: 2A (42)") {
		t.Fatalf("peek: %q %v", out, err)
	}
}

func TestStdioStream(t *testing.T) {
	sess := newSess(t)
	inR, inW := io.Pipe()   // client -> server
	outR, outW := io.Pipe() // server -> client
	done := make(chan struct{})
	go func() {
		ServeStream(sess, inR, outW)
		outW.Close()
		close(done)
	}()
	c, err := go81link.NewClient(outR, inW, inW)
	if err != nil {
		t.Fatal(err)
	}
	talk(t, c) // Close ends the input: the server returns
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ServeStream did not return at end of input")
	}
}

func TestUnixSocket(t *testing.T) {
	path := filepath.Join(t.TempDir(), "go81.sock")
	srv, err := ListenUnix(newSess(t), path)
	if err != nil {
		t.Skip("sem suporte a socket Unix:", err)
	}
	defer srv.Close()
	c, err := go81link.Dial("unix:" + path)
	if err != nil {
		t.Fatal(err)
	}
	talk(t, c)
}

func TestNamedPipe(t *testing.T) {
	if !PipesSupported {
		t.Skip("pipes nomeados só no Windows")
	}
	name := fmt.Sprintf("go81-teste-%d", time.Now().UnixNano())
	srv, err := ListenPipe(newSess(t), name)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	if !strings.Contains(srv.String(), `\.\pipe\`+name) {
		t.Fatalf("name %q", srv.String())
	}
	c, err := go81link.Dial("pipe:" + name)
	if err != nil {
		t.Fatal(err)
	}
	talk(t, c)
}

func TestPortsAllChannels(t *testing.T) {
	sess := newSess(t)
	p := Register(sess)
	defer sess.StopClock()
	var out strings.Builder
	sock := filepath.Join(t.TempDir(), "s.sock")
	cmds := []string{"link start 0 off", "link socket " + sock}
	if PipesSupported {
		cmds = append(cmds, fmt.Sprintf("link pipe go81-teste-%d", time.Now().UnixNano()))
	}
	for _, c := range cmds {
		if err := sess.Exec(c, &out); err != nil {
			t.Fatalf("%s: %v", c, err)
		}
	}
	st := p.Status()
	if !strings.Contains(st, "Go81 em 127.0.0.1:") || !strings.Contains(st, "socket") ||
		PipesSupported && !strings.Contains(st, "pipe") {
		t.Fatalf("status %q", st)
	}
	if err := sess.Exec("link start", &out); err == nil {
		t.Fatal("TCP twice should fail")
	}
	if err := sess.Exec("link stop", &out); err != nil {
		t.Fatal(err)
	}
	if p.Any() {
		t.Fatal("ports still open")
	}
}
