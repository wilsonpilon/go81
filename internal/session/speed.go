// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"fmt"
	"io"
	"strings"
)

// The emulation speed (EightyOne: Options → Speed), in percent of a real
// ZX81: the window and the terminal run SpeedFrames frames at each tick of
// 50 (or 60) per second. Turbo (F6) is still 10×, over any speed.

// Speeds are the ones the menus offer; speed accepts any from MinSpeed to
// MaxSpeed.
var Speeds = []int{10, 25, 50, 100, 200, 400, 800, 1600, 3200}

const (
	MinSpeed = 10
	MaxSpeed = 3200
)

// Speed is the emulation speed in percent (100 = a real ZX81).
func (s *Session) Speed() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.speedPercent()
}

func (s *Session) speedPercent() int {
	if s.speed == 0 {
		return 100
	}
	return s.speed
}

// SetSpeed sets the emulation speed in percent.
func (s *Session) SetSpeed(p int) error {
	if p < MinSpeed || p > MaxSpeed {
		return fmt.Errorf("a velocidade vai de %d%% a %d%%", MinSpeed, MaxSpeed)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.speed, s.speedAcc = p, 0
	return nil
}

// SpeedFrames is how many frames to run at this tick of the front-end: one
// at 100%, two at 200%, one every other tick at 50%.
func (s *Session) SpeedFrames() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.speedAcc += s.speedPercent()
	n := s.speedAcc / 100
	s.speedAcc %= 100
	return n
}

func registerSpeed(r *Registry) {
	r.Register(&Command{Name: "speed", Aliases: []string{"velocidade"}, Usage: "speed [porcento]",
		Help: "a velocidade da emulação em % de um ZX81 de verdade (10 a 3200; 100 = normal); sem número, mostra a atual",
		Run:  cmdSpeed})
}

func cmdSpeed(s *Session, a []string, out io.Writer) error {
	if len(a) > 1 {
		return usageErr("speed [porcento]")
	}
	// Commands run with s.mu held: the fields, not Speed/SetSpeed.
	if len(a) == 1 {
		n, err := ParseNumber(strings.TrimSuffix(a[0], "%"))
		if err != nil {
			return usageErr("speed [porcento]")
		}
		if n < MinSpeed || n > MaxSpeed {
			return fmt.Errorf("a velocidade vai de %d%% a %d%%", MinSpeed, MaxSpeed)
		}
		s.speed, s.speedAcc = n, 0
	}
	p := s.speedPercent()
	fmt.Fprintf(out, "velocidade %d%%%s\n", p, speedNote(p, s.driven))
	return nil
}

func speedNote(p int, driven bool) string {
	switch {
	case !driven:
		return " (vale na janela e no terminal; os comandos do shell rodam os quadros que pedem)"
	case p != 100:
		return " (sem som fora de 100%)"
	}
	return ""
}
