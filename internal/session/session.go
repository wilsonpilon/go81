// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package session holds one emulated machine and the command registry that
// every front-end uses: the shell, the Link server (and, later, the GUI and
// TUI). A command does the same thing wherever it is typed.
package session

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
)

// Settings describe the machine.
type Settings struct {
	ROM   string // embedded ROM name or file path
	RAMK  int    // RAM pack in K (0 = 1K internal only)
	NTSC  bool
	WRX   bool
	Sound bool // EightyOne "Video Sound"
}

// DefaultSettings is a stock 16K ZX81.
func DefaultSettings() Settings { return Settings{ROM: "zx81.edition3", RAMK: 16} }

// Key timing: each key is held for a few frames and released; after ENTER
// the ROM needs about a second to redraw the screen.
const (
	KeyHold  = 5
	KeyGap   = 5
	EnterGap = 60
	// BootFrames is how long the ROM takes to clear memory and show the
	// K cursor after power-on; keys pressed before that are lost.
	BootFrames = 150
)

// ErrQuit is returned by the quit command.
var ErrQuit = errors.New("quit")

type keyPress struct {
	rows  [8]byte
	enter bool
	hold  int
}

// Session is one machine plus its state. All exported methods are safe for
// concurrent use; commands run with the session locked.
type Session struct {
	mu  sync.Mutex
	set Settings
	m   *zx81.Machine

	queue []keyPress
	hold  int
	gap   int

	saving   []byte
	lastSave *savedProgram

	frames  int64 // frames run since the machine was created
	bootEnd int64 // frame where the last power-on finishes
	paused  bool
	live    bool // something runs frames in real time (clock or front-end)
	driven  bool // a front-end (window, terminal) runs the frames
	stopLiv chan struct{}

	// OnSave, if set, is called (with the session locked) when the
	// program in the ZX81 finishes a SAVE.
	OnSave  func(name string, prog []byte)
	saveCmd bool // the save command is running: it handles the program

	reg *Registry
}

type savedProgram struct {
	name string
	prog []byte
}

// New creates a session.
func New(set Settings) (*Session, error) {
	s := &Session{reg: NewRegistry()}
	if err := s.configure(set); err != nil {
		return nil, err
	}
	registerBuiltins(s.reg)
	return s, nil
}

// Registry returns the command registry (front-ends may add commands).
func (s *Session) Registry() *Registry { return s.reg }

// LoadROM returns an embedded ROM or reads a ROM file.
func LoadROM(name string) ([]byte, error) {
	if st, err := os.Stat(name); err == nil && !st.IsDir() {
		return os.ReadFile(name)
	}
	return assets.ROM(name)
}

// configure builds a new machine (hard reset). The caller holds s.mu or
// the session is not shared yet.
func (s *Session) configure(set Settings) error {
	rom, err := LoadROM(set.ROM)
	if err != nil {
		return err
	}
	cfg := zx81.DefaultConfig(rom)
	cfg.RAMTop = zx81.RAMTopForPack(set.RAMK)
	cfg.NTSC = set.NTSC
	if set.WRX {
		cfg.HiRes = zx81.HiResWRX
	}
	cfg.VSyncSound = set.Sound
	m, err := zx81.New(cfg)
	if err != nil {
		return err
	}
	m.Tape = tape.NewDeck(3250000)
	m.Tape.Out = true // SAVE goes to the session (see the save command)
	m.SaveByte = func(b byte) { s.saving = append(s.saving, b) }
	m.SaveEnd = func() {
		if len(s.saving) == 0 {
			return
		}
		if name, prog, ok := tape.SplitSaved(s.saving); ok {
			s.lastSave = &savedProgram{name: name, prog: prog}
			if s.OnSave != nil && !s.saveCmd {
				s.OnSave(name, prog)
			}
		}
		s.saving = nil
	}
	s.m, s.set = m, set
	s.queue, s.hold, s.gap = nil, 0, 0
	s.frames, s.bootEnd = 0, BootFrames
	return nil
}

// Do runs f with the session locked and gives it the machine.
func (s *Session) Do(f func(m *zx81.Machine)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f(s.m)
}

// Settings returns the current settings.
func (s *Session) Settings() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set
}

// frame runs one frame, pressing and releasing queued keys. Caller holds s.mu.
func (s *Session) frame() {
	switch {
	case s.hold > 0:
		s.hold--
		if s.hold == 0 {
			s.m.SetKeys([8]byte{})
		}
	case s.gap > 0:
		s.gap--
	case len(s.queue) > 0:
		k := s.queue[0]
		s.queue = s.queue[1:]
		s.m.SetKeys(k.rows)
		s.hold = k.hold
		s.gap = KeyGap
		if k.enter {
			s.gap += EnterGap
		}
	}
	s.m.RunFrame()
	s.frames++
}

// keysPending reports whether keys are still being typed. Caller holds s.mu.
func (s *Session) keysPending() bool { return len(s.queue) > 0 || s.hold > 0 || s.gap > 0 }

// queueChord queues a key chord. Caller holds s.mu.
func (s *Session) queueChord(chord string, hold int) error {
	rows, err := zx81.ParseChord(chord)
	if err != nil {
		return err
	}
	if hold < 1 {
		hold = KeyHold
	}
	s.queue = append(s.queue, keyPress{rows: rows, enter: strings.EqualFold(chord, "ENTER"), hold: hold})
	return nil
}

// waitBoot waits until the ZX81 has finished powering on. Caller holds s.mu.
func (s *Session) waitBoot() {
	end := s.bootEnd
	s.waitUntil(func() bool { return s.frames >= end }, int(end-s.frames)+1)
}

// powerCycle resets the machine. Caller holds s.mu.
func (s *Session) powerCycle() {
	s.m.Reset()
	s.queue, s.hold, s.gap = nil, 0, 0
	s.bootEnd = s.frames + BootFrames
}

// queueText queues the keys that type text ("\n" is ENTER). Caller holds s.mu.
func (s *Session) queueText(text string, hold int) error {
	for _, r := range text {
		c := zx81.ChordForRune(r)
		if c == "" {
			return fmt.Errorf("o caractere %q não existe no teclado do ZX81", r)
		}
		if err := s.queueChord(c, hold); err != nil {
			return err
		}
	}
	return nil
}

// waitUntil runs the machine until cond is true or max frames pass. In live
// mode the clock runs the frames and waitUntil only waits. Caller holds s.mu;
// in live mode the lock is released while waiting.
func (s *Session) waitUntil(cond func() bool, max int) bool {
	if !s.live {
		for i := 0; i < max; i++ {
			if cond() {
				return true
			}
			s.frame()
		}
		return cond()
	}
	start := s.frames
	// Safety net for a clock that stops (front-end closed or stuck).
	deadline := time.Now().Add(time.Duration(max)*60*time.Millisecond + 5*time.Second)
	for !cond() {
		if s.frames-start >= int64(max) || s.paused || time.Now().After(deadline) {
			return false
		}
		s.mu.Unlock()
		time.Sleep(5 * time.Millisecond)
		s.mu.Lock()
	}
	return true
}

// StartClock runs the machine in real time (50 or 60 frames per second)
// until StopClock. Commands still work meanwhile.
func (s *Session) StartClock() {
	s.mu.Lock()
	if s.live { // already running (or a front-end runs the frames)
		s.mu.Unlock()
		return
	}
	s.live = true
	stop := make(chan struct{})
	s.stopLiv = stop
	ntsc := s.set.NTSC
	s.mu.Unlock()

	period := time.Second / 50
	if ntsc {
		period = time.Second / 60
	}
	go func() {
		t := time.NewTicker(period)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
			}
			s.mu.Lock()
			if !s.paused {
				s.frame()
			}
			s.mu.Unlock()
		}
	}()
}

// StopClock stops the real-time clock.
func (s *Session) StopClock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.live && !s.driven {
		close(s.stopLiv)
		s.live = false
	}
}

// Type types text on the ZX81 keyboard ("\n" is ENTER), holding each key
// for hold frames, and returns when the last key is released.
func (s *Session) Type(text string, hold int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitBoot()
	if err := s.queueText(text, hold); err != nil {
		return err
	}
	s.waitUntil(func() bool { return !s.keysPending() }, 1000000)
	return nil
}

// Unlocked runs f with the session unlocked. Only a command's Run (which is
// called with the session locked) may use it, e.g. to call StartClock.
func (s *Session) Unlocked(f func()) {
	s.mu.Unlock()
	defer s.mu.Lock()
	f()
}

// Live reports whether the real-time clock is running.
func (s *Session) Live() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.live
}

// Exec parses and runs one command line, writing its output to out.
func (s *Session) Exec(line string, out io.Writer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.execLocked(line, out)
}

// execLocked is Exec with s.mu already held (used by scripts).
func (s *Session) execLocked(line string, out io.Writer) error {
	name, rest := splitCommand(line)
	if name == "" || strings.HasPrefix(name, "#") {
		return nil
	}
	cmd := s.reg.Lookup(name)
	if cmd == nil {
		return fmt.Errorf("comando desconhecido: %s (digite help)", name)
	}
	var args []string
	if cmd.Raw {
		args = []string{rest}
	} else {
		var err error
		if args, err = SplitArgs(rest); err != nil {
			return err
		}
	}
	return cmd.Run(s, args, out)
}
