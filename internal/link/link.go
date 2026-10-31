// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package link is Go81's communication port: a TCP server that lets other
// programs (editors, test scripts, IDEs) drive the emulator.
//
// Two protocols run side by side:
//
//   - Go81 text protocol (default port 8181): one shell command per line.
//     The reply is the command output followed by a line "ok" or
//     "err <mensagem>". Output lines that start with "ok", "err" or "." get a
//     "." in front (dot-stuffing); clients remove it. A line starting with
//     "{" is a JSON request {"cmd":"..."} and gets one JSON line back:
//     {"ok":true,"output":"..."} or {"ok":false,"output":"...","error":"..."}.
//   - ZRCP (default port 10000): a subset of ZEsarUX's remote command
//     protocol, so tools made for ZEsarUX can talk to Go81 (see zrcp.go).
//
// The text protocol also runs over stdin/stdout (ServeStream), a Unix
// socket (ListenUnix) and a Windows named pipe (ListenPipe). TCP listens on
// 127.0.0.1 only and the pipe accepts only the current user: the port
// controls the whole machine.
package link

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/version"
)

// Default ports.
const (
	DefaultPort     = 8181
	DefaultZRCPPort = 10000
)

// Server is one listening protocol server.
type Server struct {
	ln     net.Listener
	sess   *session.Session
	handle func(*Server, net.Conn)
	name   string // for status: "Go81 em 127.0.0.1:8181"
	clean  func() // runs after Close (removes a socket file)

	mu    sync.Mutex
	conns map[net.Conn]struct{}
	wg    sync.WaitGroup
}

// Listen starts the Go81 text protocol on a TCP address (e.g.
// "127.0.0.1:8181").
func Listen(sess *session.Session, addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return serve(sess, ln, textHandler, "Go81 em "+ln.Addr().String()), nil
}

// ListenZRCP starts the ZRCP-compatible server on a TCP address.
func ListenZRCP(sess *session.Session, addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return serve(sess, ln, (*Server).serveZRCP, "ZRCP (ZEsarUX) em "+ln.Addr().String()), nil
}

// ListenUnix starts the Go81 text protocol on a Unix socket (Linux, macOS
// and Windows 10 or newer). A stale socket file is replaced.
func ListenUnix(sess *session.Session, path string) (*Server, error) {
	if st, err := os.Stat(path); err == nil && st.Mode()&os.ModeSocket != 0 || err == nil && st.Size() == 0 {
		os.Remove(path)
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	s := serve(sess, ln, textHandler, "Go81 no socket "+path)
	s.clean = func() { os.Remove(path) }
	return s, nil
}

// PipePath turns a pipe name ("go81") into a Windows pipe path
// (\\.\pipe\go81); a full path is kept.
func PipePath(name string) string {
	if strings.HasPrefix(name, `\\`) {
		return name
	}
	return `\\.\pipe\` + name
}

func textHandler(s *Server, c net.Conn) { ServeStream(s.sess, c, c) }

func serve(sess *session.Session, ln net.Listener, h func(*Server, net.Conn), name string) *Server {
	s := &Server{ln: ln, sess: sess, handle: h, name: name, conns: map[net.Conn]struct{}{}}
	s.wg.Add(1)
	go s.accept()
	return s
}

// Addr returns the listening address.
func (s *Server) Addr() string { return s.ln.Addr().String() }

// String describes the server ("Go81 em 127.0.0.1:8181").
func (s *Server) String() string { return s.name }

// Close stops the server and closes its connections.
func (s *Server) Close() error {
	err := s.ln.Close()
	s.mu.Lock()
	for c := range s.conns {
		c.Close()
	}
	s.mu.Unlock()
	s.wg.Wait()
	if s.clean != nil {
		s.clean()
	}
	return err
}

func (s *Server) accept() {
	defer s.wg.Done()
	for {
		c, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		s.conns[c] = struct{}{}
		s.mu.Unlock()
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handle(s, c)
			c.Close()
			s.mu.Lock()
			delete(s.conns, c)
			s.mu.Unlock()
		}()
	}
}

// Greeting is the first line of the text protocol.
func Greeting() string { return "go81 link " + version.Number() }

type jsonRequest struct {
	Cmd string `json:"cmd"`
}

type jsonReply struct {
	OK     bool   `json:"ok"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// ServeStream runs the Go81 text protocol on any stream: a connection, or
// stdin/stdout ("go81 link -stdio"). It returns at end of input or quit.
func ServeStream(sess *session.Session, in io.Reader, out io.Writer) {
	w := bufio.NewWriter(out)
	fmt.Fprintln(w, Greeting())
	w.Flush()
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "{") {
			var req jsonRequest
			var rep jsonReply
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				rep.Error = "JSON inválido: " + err.Error()
			} else {
				out, err := run(sess, req.Cmd)
				rep.Output = out
				switch {
				case errors.Is(err, session.ErrQuit):
					rep.OK = true
					writeJSON(w, rep)
					return
				case err != nil:
					rep.Error = err.Error()
				default:
					rep.OK = true
				}
			}
			writeJSON(w, rep)
			continue
		}
		out, err := run(sess, line)
		writeStuffed(w, out)
		switch {
		case errors.Is(err, session.ErrQuit):
			fmt.Fprintln(w, "ok tchau")
			w.Flush()
			return
		case err != nil:
			fmt.Fprintln(w, "err", strings.ReplaceAll(err.Error(), "\n", " "))
		default:
			fmt.Fprintln(w, "ok")
		}
		if w.Flush() != nil {
			return
		}
	}
}

// run executes a line that may hold several commands separated by ';'.
func run(sess *session.Session, line string) (string, error) {
	var out bytes.Buffer
	for _, cmd := range session.SplitCommands(line) {
		if err := sess.Exec(cmd, &out); err != nil {
			return out.String(), err
		}
	}
	return out.String(), nil
}

func writeJSON(w *bufio.Writer, rep jsonReply) {
	b, _ := json.Marshal(rep)
	w.Write(b)
	w.WriteByte('\n')
	w.Flush()
}

// NeedsStuffing reports whether an output line must get a "." in front.
func NeedsStuffing(l string) bool {
	return l == "ok" || l == "err" || strings.HasPrefix(l, "ok ") ||
		strings.HasPrefix(l, "err ") || strings.HasPrefix(l, ".")
}

func writeStuffed(w io.Writer, out string) {
	out = strings.TrimSuffix(out, "\n")
	if out == "" {
		return
	}
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSuffix(l, "\r")
		if NeedsStuffing(l) {
			l = "." + l
		}
		fmt.Fprintln(w, l)
	}
}
