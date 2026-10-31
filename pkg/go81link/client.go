// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package go81link is a client for Go81's communication port (Link).
//
//	c, err := go81link.Dial("127.0.0.1:8181")
//	out, err := c.Cmd(`typeln P"OLA"`)
//	out, err = c.Cmd("screen")
//
// Every Go81 shell command works here ("help" lists them).
package go81link

import (
	"bufio"
	"errors"
	"fmt"
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
	conn     net.Conn
	r        *bufio.Reader
	Greeting string // e.g. "go81 link 1.6.1"
}

// Dial connects to a Go81 Link server.
func Dial(addr string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	c := &Client{conn: conn, r: bufio.NewReader(conn)}
	g, err := c.readLine()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !strings.HasPrefix(g, "go81 link") {
		conn.Close()
		return nil, fmt.Errorf("go81link: %s não é um servidor Go81 (%q)", addr, g)
	}
	c.Greeting = g
	return c, nil
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
		return "", errors.New("go81link: o comando não pode ter quebra de linha (use \\n em type)")
	}
	if _, err := fmt.Fprintf(c.conn, "%s\n", line); err != nil {
		return "", err
	}
	var out strings.Builder
	for {
		l, err := c.readLine()
		if err != nil {
			return out.String(), err
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
func (c *Client) Close() error { return c.conn.Close() }
