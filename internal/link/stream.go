// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package link

import (
	"bufio"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/wilsonpilon/go81/internal/session"
)

// Options of a Go81-protocol server.
type Options struct {
	// Token, if set, must be sent with "auth <token>" before any other
	// command. The greeting then ends in " auth".
	Token string
}

func optsOf(opt []Options) Options {
	if len(opt) > 0 {
		return opt[0]
	}
	return Options{}
}

// stream is one client of the text protocol.
type stream struct {
	sess   *session.Session
	opt    Options
	mu     sync.Mutex // the writer is shared by replies and events
	w      *bufio.Writer
	authed bool

	events     chan session.Event
	cancel     func()
	jsonEvents bool
	done       chan struct{}
}

// ServeStream runs the Go81 text protocol on any stream: a connection, or
// stdin/stdout ("go81 link -stdio"). It returns at end of input or quit.
//
// Besides the shell commands, a connection understands:
//
//	auth <token>      when the server has a token (greeting "go81 link X.Y.Z auth")
//	events on|off     start/stop events: lines "! report 0/10", "! save NOME", "! load"
//	                  (or JSON {"event":"report","data":"0/10"} if asked in JSON)
func ServeStream(sess *session.Session, in io.Reader, out io.Writer, opt ...Options) {
	st := &stream{sess: sess, opt: optsOf(opt), w: bufio.NewWriter(out), done: make(chan struct{})}
	st.authed = st.opt.Token == ""
	defer st.stopEvents()
	g := Greeting()
	if !st.authed {
		g += " auth"
	}
	st.write(func(w *bufio.Writer) { fmt.Fprintln(w, g) })
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		isJSON := strings.HasPrefix(line, "{")
		cmd := line
		if isJSON {
			var req jsonRequest
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				st.replyJSON(jsonReply{Error: "JSON inválido: " + err.Error()})
				continue
			}
			cmd = req.Cmd
		}
		out, err := st.handle(cmd, isJSON)
		if isJSON {
			rep := jsonReply{OK: err == nil || errors.Is(err, session.ErrQuit), Output: out}
			if err != nil && !errors.Is(err, session.ErrQuit) {
				rep.Error = err.Error()
			}
			st.replyJSON(rep)
		} else {
			st.write(func(w *bufio.Writer) {
				writeStuffed(w, out)
				switch {
				case errors.Is(err, session.ErrQuit):
					fmt.Fprintln(w, "ok tchau")
				case err != nil:
					fmt.Fprintln(w, "err", strings.ReplaceAll(err.Error(), "\n", " "))
				default:
					fmt.Fprintln(w, "ok")
				}
			})
		}
		if errors.Is(err, session.ErrQuit) {
			return
		}
	}
}

// handle runs a line: the connection's own commands or shell commands.
func (st *stream) handle(line string, isJSON bool) (string, error) {
	word, rest, _ := strings.Cut(line, " ")
	word = strings.ToLower(word)
	if word == "auth" {
		if st.opt.Token == "" {
			return "", nil
		}
		if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(rest)), []byte(st.opt.Token)) != 1 {
			return "", errors.New("token errado")
		}
		st.authed = true
		return "", nil
	}
	if !st.authed {
		if word == "quit" || word == "exit" {
			return "", session.ErrQuit
		}
		return "", errors.New("autenticação necessária: envie auth <token>")
	}
	if word == "events" {
		switch strings.ToLower(strings.TrimSpace(rest)) {
		case "on":
			st.startEvents(isJSON)
			return "eventos ligados\n", nil
		case "off":
			st.stopEvents()
			return "eventos desligados\n", nil
		}
		return "", errors.New("uso: events on|off")
	}
	return run(st.sess, line)
}

func (st *stream) write(f func(w *bufio.Writer)) {
	st.mu.Lock()
	defer st.mu.Unlock()
	f(st.w)
	st.w.Flush()
}

func (st *stream) replyJSON(rep jsonReply) {
	b, _ := json.Marshal(rep)
	st.write(func(w *bufio.Writer) { w.Write(b); w.WriteByte('\n') })
}

// startEvents subscribes to the session; events are written by their own
// goroutine, between the lines of the replies.
func (st *stream) startEvents(isJSON bool) {
	st.stopEvents()
	ch := make(chan session.Event, 64)
	st.events, st.jsonEvents = ch, isJSON
	st.cancel = st.sess.Subscribe(func(e session.Event) {
		select {
		case ch <- e:
		default: // a slow client loses events rather than slowing the ZX81
		}
	})
	asJSON := isJSON
	go func() {
		for e := range ch {
			if asJSON {
				b, _ := json.Marshal(struct {
					Event string `json:"event"`
					Data  string `json:"data"`
				}{e.Kind, e.Text})
				st.write(func(w *bufio.Writer) { w.Write(b); w.WriteByte('\n') })
				continue
			}
			line := strings.TrimSpace("! " + e.Kind + " " + e.Text)
			st.write(func(w *bufio.Writer) { fmt.Fprintln(w, line) })
		}
	}()
}

func (st *stream) stopEvents() {
	if st.cancel != nil {
		st.cancel()
		st.cancel = nil
		close(st.events)
	}
}
