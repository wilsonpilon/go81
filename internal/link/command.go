// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package link

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/wilsonpilon/go81/internal/session"
)

// DefaultPipe is the default Windows pipe name (\\.\pipe\go81).
const DefaultPipe = "go81"

// Ports holds the servers started by the "link" command or the CLI.
type Ports struct {
	mu      sync.Mutex
	servers []*Server
	tcp     bool
}

// Start opens the TCP text port and, if zrcpPort >= 0, the ZRCP port. Port 0
// picks a free port (useful in tests).
func (p *Ports) Start(sess *session.Session, port, zrcpPort int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.tcp {
		return fmt.Errorf("as portas TCP já estão ligadas")
	}
	t, err := Listen(sess, fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	p.servers = append(p.servers, t)
	if zrcpPort >= 0 {
		z, err := ListenZRCP(sess, fmt.Sprintf("127.0.0.1:%d", zrcpPort))
		if err != nil {
			return err
		}
		p.servers = append(p.servers, z)
	}
	p.tcp = true
	return nil
}

// AddSocket opens a Unix socket.
func (p *Ports) AddSocket(sess *session.Session, path string) error {
	s, err := ListenUnix(sess, path)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.servers = append(p.servers, s)
	p.mu.Unlock()
	return nil
}

// AddPipe opens a Windows named pipe.
func (p *Ports) AddPipe(sess *session.Session, name string) error {
	s, err := ListenPipe(sess, name)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.servers = append(p.servers, s)
	p.mu.Unlock()
	return nil
}

// Any reports whether some port is open.
func (p *Ports) Any() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.servers) > 0
}

// Stop closes every port. Connections are closed in the background,
// because a connection may be waiting for the session that called Stop.
func (p *Ports) Stop() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.servers) == 0 {
		return false
	}
	list := p.servers
	p.servers, p.tcp = nil, false
	go func() {
		for _, s := range list {
			s.Close()
		}
	}()
	return true
}

// Status describes the open ports.
func (p *Ports) Status() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.servers) == 0 {
		return "link desligado"
	}
	names := make([]string, len(p.servers))
	for i, s := range p.servers {
		names[i] = s.String()
	}
	return "link ligado: " + strings.Join(names, ", ")
}

// Register adds the "link" command to the session and returns the ports it
// controls.
func Register(sess *session.Session) *Ports {
	p := &Ports{}
	sess.Registry().Register(&session.Command{
		Name:  "link",
		Usage: "link [start [porta] [porta-zrcp|off] | socket <arquivo> | pipe [nome] | stop | status]",
		Help:  "porta de comunicação: TCP (Go81 8181 e ZRCP 10000), socket ou pipe",
		Run: func(s *session.Session, a []string, out io.Writer) error {
			op := "status"
			if len(a) > 0 {
				op = a[0]
			}
			switch op {
			case "start", "on":
				port, zport := DefaultPort, DefaultZRCPPort
				if len(a) > 1 {
					v, err := session.ParseNumber(a[1])
					if err != nil || v < 0 || v > 65535 {
						return fmt.Errorf("porta inválida: %s", a[1])
					}
					port = v
				}
				if len(a) > 2 {
					if a[2] == "off" {
						zport = -1
					} else {
						v, err := session.ParseNumber(a[2])
						if err != nil || v < 0 || v > 65535 {
							return fmt.Errorf("porta inválida: %s", a[2])
						}
						zport = v
					}
				}
				if err := p.Start(s, port, zport); err != nil {
					return err
				}
			case "socket":
				if len(a) != 2 {
					return fmt.Errorf("uso: link socket <arquivo>")
				}
				if err := p.AddSocket(s, a[1]); err != nil {
					return err
				}
			case "pipe":
				name := DefaultPipe
				if len(a) > 1 {
					name = a[1]
				}
				if err := p.AddPipe(s, name); err != nil {
					return err
				}
			case "stop", "off":
				if !p.Stop() {
					return fmt.Errorf("o link não está ligado")
				}
			case "status":
			default:
				return fmt.Errorf("uso: link [start|socket|pipe|stop|status]")
			}
			if op != "stop" && op != "off" && op != "status" {
				s.Unlocked(s.StartClock) // the ZX81 runs in real time
			}
			fmt.Fprintln(out, p.Status())
			return nil
		},
	})
	return p
}
