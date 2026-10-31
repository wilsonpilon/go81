// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package link

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/pkg/go81link"
)

func startText(t *testing.T) (*session.Session, *Server) {
	t.Helper()
	sess, err := session.New(session.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	sess.Exec("run 150", nil)
	srv, err := Listen(sess, "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { srv.Close() })
	return sess, srv
}

func TestTextProtocolWithClient(t *testing.T) {
	_, srv := startText(t)
	c, err := go81link.Dial(srv.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if !strings.HasPrefix(c.Greeting, "go81 link ") {
		t.Fatalf("greeting %q", c.Greeting)
	}
	if _, err := c.Cmd("typeln P2*21"); err != nil {
		t.Fatal(err)
	}
	out, err := c.Cmd("screen")
	if err != nil || !strings.Contains(out, "|42 ") {
		t.Fatalf("screen: %v\n%s", err, out)
	}
	// Dot-stuffing: an output line "ok" must not end the reply.
	out, err = c.Cmd("echo ok; echo .x; echo fim")
	if err != nil || out != "ok\n.x\nfim\n" {
		t.Fatalf("stuffing: %q %v", out, err)
	}
	_, err = c.Cmd("peek")
	var ce *go81link.CommandError
	if !errors.As(err, &ce) || !strings.Contains(ce.Msg, "uso: peek") {
		t.Fatalf("error reply: %v", err)
	}
	// The connection still works after an error.
	if out, err = c.Cmd("peek 0"); err != nil || !strings.HasPrefix(out, "0000: D3") {
		t.Fatalf("peek after error: %q %v", out, err)
	}
}

func TestJSONRequest(t *testing.T) {
	_, srv := startText(t)
	conn, err := net.Dial("tcp", srv.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)
	r.ReadString('\n') // greeting
	fmt.Fprintln(conn, `{"cmd":"peek 0"}`)
	l, _ := r.ReadString('\n')
	var rep struct {
		OK     bool
		Output string
		Error  string
	}
	if err := json.Unmarshal([]byte(l), &rep); err != nil || !rep.OK || !strings.HasPrefix(rep.Output, "0000: D3") {
		t.Fatalf("json reply %q: %v", l, err)
	}
	fmt.Fprintln(conn, `{"cmd":"xyz"}`)
	l, _ = r.ReadString('\n')
	json.Unmarshal([]byte(l), &rep)
	if rep.OK || !strings.Contains(rep.Error, "desconhecido") {
		t.Fatalf("json error %q", l)
	}
}

// zrcp sends one command and returns the text before the next prompt.
func zrcp(t *testing.T, conn net.Conn, r *bufio.Reader, cmd string) string {
	t.Helper()
	fmt.Fprintf(conn, "%s\n", cmd)
	return readUntilPrompt(t, conn, r)
}

func readUntilPrompt(t *testing.T, conn net.Conn, r *bufio.Reader) string {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	var sb strings.Builder
	for !strings.HasSuffix(sb.String(), ZRCPPrompt) {
		b, err := r.ReadByte()
		if err != nil {
			t.Fatalf("read: %v (got %q)", err, sb.String())
		}
		sb.WriteByte(b)
	}
	return strings.TrimSuffix(sb.String(), ZRCPPrompt)
}

func TestZRCP(t *testing.T) {
	sess, err := session.New(session.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	sess.Exec("run 150", nil)
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
	if w := readUntilPrompt(t, conn, r); w != ZRCPWelcome {
		t.Fatalf("welcome %q", w)
	}
	cases := []struct{ cmd, want string }{
		{"about", "ZEsarUX remote command protocol"},
		{"get-current-machine", "ZX81"},
		{"read-memory 0 4", "D3FD01FF"},
		{"write-memory 30000 1 $FF 10H", ""},
		{"read-memory 30000 3", "01FF10"},
		{"xyz", "Unknown command"},
		{"set-register BC=1234H", ""},
	}
	for _, c := range cases {
		got := zrcp(t, conn, r, c.cmd)
		if c.cmd == "set-register BC=1234H" {
			if !strings.Contains(got, " BC=1234 ") {
				t.Errorf("set-register: %q", got)
			}
			continue
		}
		if got != c.want {
			t.Errorf("%s = %q, want %q", c.cmd, got, c.want)
		}
	}
	re := regexp.MustCompile(`^PC=[0-9a-f]{4} SP=[0-9a-f]{4} AF=[0-9a-f]{4} BC=1234 HL=[0-9a-f]{4} DE=[0-9a-f]{4} IX=[0-9a-f]{4} IY=4000 AF'=[0-9a-f]{4} BC'=[0-9a-f]{4} HL'=[0-9a-f]{4} DE'=[0-9a-f]{4} I=1e R=[0-9a-f]{2}  F=[SZ5H3PNC-]{8} F'=[SZ5H3PNC-]{8} MEMPTR=0000 IM[012] IFF[1-][2-] VPS: 50 MMU=0{32}$`)
	if g := zrcp(t, conn, r, "get-registers"); !re.MatchString(g) {
		t.Errorf("get-registers = %q", g)
	}
	zrcp(t, conn, r, "send-keys-string 100 p2*21")
	ocr := zrcp(t, conn, r, "get-ocr")
	if !strings.Contains(ocr, "PRINT 2*21") || strings.Count(ocr, "\n") != 24 {
		t.Errorf("get-ocr = %q", ocr)
	}
	fmt.Fprintln(conn, "quit")
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	l, _ := r.ReadString('\n')
	if l != "Sayonara baby\n" {
		t.Errorf("quit = %q", l)
	}
}

func TestPortsCommand(t *testing.T) {
	sess, _ := session.New(session.DefaultSettings())
	Register(sess)
	defer sess.StopClock()
	var out strings.Builder
	if err := sess.Exec("link start 0 off", &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "link ligado: Go81 em 127.0.0.1:") || !sess.Live() {
		t.Fatalf("link start: %q live=%v", out.String(), sess.Live())
	}
	if err := sess.Exec("link stop", &out); err != nil {
		t.Fatal(err)
	}
	if err := sess.Exec("link stop", &out); err == nil {
		t.Fatal("second stop should fail")
	}
}
