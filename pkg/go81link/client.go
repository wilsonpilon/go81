// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package go81link is a client for Go81's communication port (Link).
//
//	c, err := go81link.Dial("127.0.0.1:8181")      // TCP ("go81 link")
//	c, err := go81link.Dial("pipe:go81")           // Windows pipe ("go81 link -pipe go81")
//	c, err := go81link.Dial("unix:/tmp/go81.sock") // Unix socket ("go81 link -socket ...")
//	out, err := c.Cmd(`typeln 10P"OLA"`)
//	out, err = c.Cmd("screen")
//
// Every Go81 shell command works here ("help" lists them). NewClient talks
// over any stream, e.g. the stdin/stdout of a "go81 link -stdio" process.
package go81link

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// DefaultAddr is where "go81 link" listens by default.
const DefaultAddr = "127.0.0.1:8181"

// CommandError is a command that ran and failed ("err ..." reply).
type CommandError struct {
	Cmd, Msg string
}

func (e *CommandError) Error() string { return e.Cmd + ": " + e.Msg }

// Client is a connection to a Go81 Link server. It is not safe for
// concurrent use; open one client per goroutine.
type Client struct {
	w        io.Writer
	closer   io.Closer
	r        *bufio.Reader
	Greeting string // e.g. "go81 link 1.7.4" ("... auth" when a token is needed)
	events   chan Event
}

// Event is something that happened in the ZX81, after EnableEvents:
// Kind "report" (Text "0/10"), "save" (Text = program name) or "load".
type Event struct {
	Kind, Text string
}

// NeedsAuth reports whether the server asks for a token (Auth).
func (c *Client) NeedsAuth() bool { return strings.HasSuffix(c.Greeting, " auth") }

// Auth sends the token of a protected server.
func (c *Client) Auth(token string) error {
	_, err := c.Cmd("auth " + token)
	return err
}

// EnableEvents asks the server for events. They arrive on Events while
// commands run (and on the next Cmd or WaitEvent if nothing else is read).
func (c *Client) EnableEvents() error {
	_, err := c.Cmd("events on")
	return err
}

// Events returns the channel of events (buffered; old events are dropped
// if nobody reads them).
func (c *Client) Events() <-chan Event { return c.events }

// WaitEvent reads the connection until an event arrives (use it when no
// command is running).
func (c *Client) WaitEvent() (Event, error) {
	select {
	case e := <-c.events:
		return e, nil
	default:
	}
	for {
		l, err := c.readLine()
		if err != nil {
			return Event{}, err
		}
		if e, ok := parseEvent(l); ok {
			return e, nil
		}
	}
}

func parseEvent(l string) (Event, bool) {
	if !strings.HasPrefix(l, "! ") {
		return Event{}, false
	}
	kind, text, _ := strings.Cut(strings.TrimPrefix(l, "! "), " ")
	return Event{Kind: kind, Text: text}, true
}

func (c *Client) push(e Event) {
	select {
	case c.events <- e:
	default:
		select { // full: drop the oldest
		case <-c.events:
		default:
		}
		c.events <- e
	}
}

// Dial connects to a Go81 Link server. addr is a TCP address
// ("127.0.0.1:8181"), "unix:<socket file>" or "pipe:<name>" (Windows,
// \\.\pipe\<name>).
func Dial(addr string) (*Client, error) {
	var conn net.Conn
	var err error
	switch {
	case strings.HasPrefix(addr, "unix:"):
		conn, err = net.DialTimeout("unix", strings.TrimPrefix(addr, "unix:"), 5*time.Second)
	case strings.HasPrefix(addr, "pipe:"):
		conn, err = dialPipe(strings.TrimPrefix(addr, "pipe:"))
	default:
		conn, err = net.DialTimeout("tcp", addr, 5*time.Second)
	}
	if err != nil {
		return nil, err
	}
	return NewClient(conn, conn, conn)
}

// NewClient starts a client on any stream: it reads the server's greeting
// from r and sends commands to w. closer (optional) is closed by Close.
func NewClient(r io.Reader, w io.Writer, closer io.Closer) (*Client, error) {
	c := &Client{w: w, closer: closer, r: bufio.NewReader(r), events: make(chan Event, 64)}
	g, err := c.readLine()
	if err == nil && !strings.HasPrefix(g, "go81 link") {
		err = fmt.Errorf("go81link: o outro lado não é um servidor Go81 (%q)", g)
	}
	if err != nil {
		if closer != nil {
			closer.Close()
		}
		return nil, err
	}
	c.Greeting = g
	return c, nil
}

// pipePath turns "go81" into \\.\pipe\go81 (a full path is kept).
func pipePath(name string) string {
	if strings.HasPrefix(name, `\\`) {
		return name
	}
	return `\\.\pipe\` + name
}

func (c *Client) readLine() (string, error) {
	l, err := c.r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(l, "\r\n"), nil
}

// Cmd runs one command line (several commands may be joined with ';') and
// returns its output. A failed command returns a *CommandError together
// with the output produced before the failure.
func (c *Client) Cmd(line string) (string, error) {
	if strings.ContainsAny(line, "\r\n") {
		return "", errors.New(`go81link: o comando não pode ter quebra de linha (use \n em type)`)
	}
	if _, err := io.WriteString(c.w, line+"\n"); err != nil {
		return "", err
	}
	var out strings.Builder
	for {
		l, err := c.readLine()
		if err != nil {
			return out.String(), err
		}
		if e, ok := parseEvent(l); ok {
			c.push(e)
			continue
		}
		switch {
		case l == "ok" || strings.HasPrefix(l, "ok "):
			return out.String(), nil
		case l == "err" || strings.HasPrefix(l, "err "):
			return out.String(), &CommandError{Cmd: line, Msg: strings.TrimPrefix(strings.TrimPrefix(l, "err"), " ")}
		}
		out.WriteString(strings.TrimPrefix(l, "."))
		out.WriteByte('\n')
	}
}

// Close ends the connection (the server keeps running).
func (c *Client) Close() error {
	if c.closer == nil {
		return nil
	}
	return c.closer.Close()
}
