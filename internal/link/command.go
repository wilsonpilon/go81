// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package link

import (
	"fmt"
	"io"
	"sync"

	"github.com/wilsonpilon/go81/internal/session"
)

// Ports holds the servers started by the "link" command.
type Ports struct {
	mu         sync.Mutex
	text, zrcp *Server
}

// Start opens the text port and, if zrcpPort > 0, the ZRCP port, and starts
// the real-time clock. Port 0 picks a free port (useful in tests).
func (p *Ports) Start(sess *session.Session, port, zrcpPort int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.text != nil {
		return fmt.Errorf("o link já está ligado em %s", p.text.Addr())
	}
	t, err := Listen(sess, fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	var z *Server
	if zrcpPort >= 0 {
		if z, err = ListenZRCP(sess, fmt.Sprintf("127.0.0.1:%d", zrcpPort)); err != nil {
			t.Close()
			return err
		}
	}
	p.text, p.zrcp = t, z
	return nil
}

// Stop closes the ports. Connections are closed in the background, because
// a connection may be waiting for the session that called Stop.
func (p *Ports) Stop() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.text == nil {
		return false
	}
	t, z := p.text, p.zrcp
	p.text, p.zrcp = nil, nil
	go func() {
		t.Close()
		if z != nil {
			z.Close()
		}
	}()
	return true
}

// Status describes the open ports.
func (p *Ports) Status() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.text == nil {
		return "link desligado"
	}
	s := "link ligado: Go81 em " + p.text.Addr()
	if p.zrcp != nil {
		s += ", ZRCP (ZEsarUX) em " + p.zrcp.Addr()
	}
	return s
}

// Register adds the "link" command to the session and returns the ports it
// controls.
func Register(sess *session.Session) *Ports {
	p := &Ports{}
	sess.Registry().Register(&session.Command{
		Name:  "link",
		Usage: "link [start [porta] [porta-zrcp|off] | stop | status]",
		Help:  "liga a porta de comunicação (Go81 8181 e ZRCP 10000) e o relógio",
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
				s.Unlocked(s.StartClock)
			case "stop", "off":
				if !p.Stop() {
					return fmt.Errorf("o link não está ligado")
				}
			case "status":
			default:
				return fmt.Errorf("uso: link [start|stop|status]")
			}
			fmt.Fprintln(out, p.Status())
			return nil
		},
	})
	return p
}
