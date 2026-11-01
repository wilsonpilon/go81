// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// Event is something that happened in the ZX81, sent to the listeners
// (the Link's "events on").
type Event struct {
	Kind string // "report", "save", "load"
	Text string // "0/10", the program name, ...
}

// romReport is the ROM routine that shows the report (0/10, 2/30...) at
// the end of a command or program; the same in the three ZX81 ROMs.
const romReport = 0x06D1

// Subscribe adds a listener and returns the function that removes it.
// Listeners are called with the session locked: they must not block or
// call the session.
func (s *Session) Subscribe(f func(Event)) (cancel func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listeners == nil {
		s.listeners = map[int]func(Event){}
	}
	s.nextListener++
	id := s.nextListener
	s.listeners[id] = f
	s.hookSteps()
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.listeners, id)
		s.hookSteps()
	}
}

// emit sends an event to the listeners. Caller holds s.mu.
func (s *Session) emit(kind, text string) {
	for _, f := range s.listeners {
		f(Event{Kind: kind, Text: text})
	}
}

// hookSteps watches the instructions only while someone listens (it costs
// a little speed). Caller holds s.mu.
func (s *Session) hookSteps() {
	if len(s.listeners) == 0 {
		s.m.StepHook = nil
		return
	}
	m := s.m
	m.StepHook = func() {
		if m.CPU.PC == romReport {
			s.emit("report", reportText(m))
		}
	}
}

// reportText is the report the ROM is about to show, e.g. "0/10".
func reportText(m *zx81.Machine) string {
	code := int(m.Peek(0x4000) + 1) // ERR_NR holds the code minus one (FF = 0, OK)
	line := int(m.Peek(0x4007)) | int(m.Peek(0x4008))<<8
	if line >= 0xFFFE { // a command typed directly
		line = 0
	}
	c := '?'
	if i := 0x1C + code; code >= 0 && i < len(zx81.TextChars) {
		c = zx81.TextChars[i]
	}
	return fmt.Sprintf("%c/%d", c, line)
}
