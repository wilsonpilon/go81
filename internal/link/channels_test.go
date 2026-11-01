// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package link

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
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

func TestTokenAndEvents(t *testing.T) {
	sess := newSess(t)
	srv, err := Listen(sess, "127.0.0.1:0", Options{Token: "segredo"})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c, err := go81link.Dial(srv.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if !c.NeedsAuth() {
		t.Fatalf("greeting %q should ask for auth", c.Greeting)
	}
	if _, err := c.Cmd("peek 0"); err == nil || !strings.Contains(err.Error(), "autenticação") {
		t.Fatalf("command before auth: %v", err)
	}
	if err := c.Auth("errado"); err == nil {
		t.Fatal("wrong token accepted")
	}
	if err := c.Auth("segredo"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnableEvents(); err != nil {
		t.Fatal(err)
	}
	// An output line starting with "!" is stuffed, not taken as an event.
	if out, err := c.Cmd("echo !nao e evento"); err != nil || out != "!nao e evento\n" {
		t.Fatalf("stuffing: %q %v", out, err)
	}
	for _, cmd := range []string{"run 150", `typeln 10P"OI"`, "typeln R", "run 50"} {
		if _, err := c.Cmd(cmd); err != nil {
			t.Fatal(cmd, err)
		}
	}
	got := map[string]bool{}
	timeout := time.After(5 * time.Second)
	for !got["report 0/10"] {
		select {
		case e := <-c.Events():
			got[e.Kind+" "+e.Text] = true
		case <-timeout:
			t.Fatalf("no report event, got %v", got)
		}
	}
}

func TestEventsJSON(t *testing.T) {
	sess := newSess(t)
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	go ServeStream(sess, inR, outW)
	r := bufio.NewReader(outR)
	r.ReadString('\n')
	fmt.Fprintln(inW, `{"cmd":"events on"}`)
	r.ReadString('\n')
	fmt.Fprintln(inW, `{"cmd":"run 150; typeln P1"}`)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		l, err := r.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(l, `"event":"report"`) && strings.Contains(l, `"data":"0/0"`) {
			inW.Close()
			return
		}
	}
	t.Fatal("no JSON report event")
}

func TestZRCPMoreCommands(t *testing.T) {
	sess := newSess(t)
	srv, err := ListenZRCP(sess, "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	conn, err := net.Dial("tcp", srv.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)
	readUntilPrompt(t, conn, r)
	shot := filepath.Join(t.TempDir(), "tela.txt")
	cases := []struct{ cmd, want string }{
		{"evaluate 2+3*4", "14"},
		{"e peek(0)", "211"},
		{"evaluate", "Error. No expression"},
		{"evaluate 1/0", "Error. divisão por zero"},
		{"save-screen " + shot, ""},
		{"reset-cpu", ""},
		{"exit-cpu-step", "Error. You are not in step to step mode"},
	}
	for _, c := range cases {
		if got := zrcp(t, conn, r, c.cmd); got != c.want {
			t.Errorf("%s = %q, want %q", c.cmd, got, c.want)
		}
	}
	if _, err := os.Stat(shot); err != nil {
		t.Errorf("save-screen: %v", err)
	}
	if os := zrcp(t, conn, r, "get-os"); os == "" || os == "Unknown command" {
		t.Errorf("get-os = %q", os)
	}
	// enter-cpu-step pauses and changes the prompt.
	fmt.Fprintln(conn, "enter-cpu-step")
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var sb strings.Builder
	for !strings.HasSuffix(sb.String(), ZRCPStepPrompt) {
		b, err := r.ReadByte()
		if err != nil {
			t.Fatalf("no step prompt: %q", sb.String())
		}
		sb.WriteByte(b)
	}
	if !sess.Paused() {
		t.Error("enter-cpu-step did not pause")
	}
	fmt.Fprintln(conn, "exit-cpu-step")
	readUntilPrompt(t, conn, r)
	if sess.Paused() {
		t.Error("exit-cpu-step did not resume")
	}
}
