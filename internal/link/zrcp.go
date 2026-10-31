// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package link

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/wilsonpilon/go81/internal/emu/z80"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/version"
)

// ZRCP texts, as in ZEsarUX (src/zrcp/remote.c).
const (
	ZRCPWelcome = "Welcome to ZEsarUX remote command protocol (ZRCP)\nWrite help for available commands\n"
	ZRCPPrompt  = "\ncommand> "
)

type zrcpCommand struct {
	names  []string
	params string
	help   string
	run    func(s *Server, params string) string
}

// zrcpCommands is the supported subset. Replies follow ZEsarUX byte for byte.
var zrcpCommands []zrcpCommand

func init() {
	zrcpCommands = []zrcpCommand{
		{[]string{"about"}, "", "Shows about message", func(*Server, string) string {
			return "ZEsarUX remote command protocol"
		}},
		{[]string{"get-version"}, "", "Shows emulator version", func(*Server, string) string {
			return version.Number()
		}},
		{[]string{"get-current-machine", "gcm"}, "", "Returns current machine name", func(s *Server, _ string) string {
			if s.sess.Settings().NTSC {
				return "TS1000"
			}
			return "ZX81"
		}},
		{[]string{"get-registers", "gr"}, "", "Get CPU registers", func(s *Server, _ string) string {
			return s.zrcpRegisters()
		}},
		{[]string{"set-register", "sr"}, "register=value", "Changes register value. Example: set-register DE=3344H", zrcpSetRegister},
		{[]string{"read-memory"}, "[address] [length]", "Dumps memory at address", zrcpReadMemory},
		{[]string{"write-memory", "wm"}, "address value", "Writes a sequence of bytes starting at desired address on memory. Bytes must be separated by one space each one", zrcpWriteMemory},
		{[]string{"hard-reset-cpu"}, "", "Hard resets the machine", func(s *Server, _ string) string {
			return s.zrcpExec("reset")
		}},
		{[]string{"send-keys-string"}, "time string", "Simulates sending some keys on parameter string. Every key is separated in time by a non-press time. Time is in miliseconds, a normal value for Basic writing is 100 miliseconds", zrcpSendKeys},
		{[]string{"get-ocr"}, "", "Get OCR output text", func(s *Server, _ string) string {
			return s.zrcpOCR()
		}},
		{[]string{"smartload", "sl"}, "file", "Smart-loads a file", func(s *Server, p string) string {
			a, err := session.SplitArgs(p)
			if err != nil || len(a) < 1 {
				return "ERROR. No parameter set"
			}
			var out bytes.Buffer
			if err := s.sess.Exec(fmt.Sprintf("load %q", a[0]), &out); err != nil {
				return "ERROR. " + err.Error()
			}
			return ""
		}},
		{[]string{"view-basic"}, "", "Gets Basic program listing", func(s *Server, _ string) string {
			var out bytes.Buffer
			if err := s.sess.Exec("list", &out); err != nil {
				return "ERROR. " + err.Error()
			}
			return strings.TrimSuffix(out.String(), "\n")
		}},
		{[]string{"help"}, "[command]", "Shows help screen or command help", zrcpHelp},
	}
}

func lookupZRCP(name string) *zrcpCommand {
	for i := range zrcpCommands {
		for _, n := range zrcpCommands[i].names {
			if n == name {
				return &zrcpCommands[i]
			}
		}
	}
	return nil
}

func (s *Server) serveZRCP(c net.Conn) {
	w := bufio.NewWriter(c)
	w.WriteString(ZRCPWelcome)
	sc := bufio.NewScanner(c)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for {
		w.WriteString(ZRCPPrompt)
		if w.Flush() != nil || !sc.Scan() {
			return
		}
		line := strings.TrimSpace(sc.Text())
		name, params, _ := strings.Cut(line, " ")
		params = strings.TrimLeft(params, " ")
		switch name {
		case "":
			continue
		case "quit", "exit", "logout":
			w.WriteString("Sayonara baby\n")
			w.Flush()
			return
		}
		cmd := lookupZRCP(name)
		if cmd == nil {
			w.WriteString("Unknown command")
			continue
		}
		w.WriteString(cmd.run(s, params))
	}
}

// zrcpExec runs a Go81 command and turns an error into a ZRCP error.
func (s *Server) zrcpExec(line string) string {
	var out bytes.Buffer
	if err := s.sess.Exec(line, &out); err != nil {
		return "ERROR. " + err.Error()
	}
	return ""
}

// zrcpNumber parses a number as ZEsarUX's parse_string_to_number does
// (errors become 0).
func zrcpNumber(t string) int {
	v, _ := session.ParseNumber(t)
	return v
}

func zrcpFlags(f byte) string { return session.FlagString(f) }

// ZRCPRegisters formats the registers like ZEsarUX's print_registers.
// Go81 does not emulate MEMPTR (neither does EightyOne), so it is 0000.
func ZRCPRegisters(c *z80.CPU, vps int) string {
	iff := []byte("--")
	if c.IFF1 != 0 {
		iff[0] = '1'
	}
	if c.IFF2 != 0 {
		iff[1] = '2'
	}
	return fmt.Sprintf("PC=%04x SP=%04x AF=%04x BC=%04x HL=%04x DE=%04x IX=%04x IY=%04x AF'=%04x BC'=%04x HL'=%04x DE'=%04x I=%02x R=%02x  "+
		"F=%s F'=%s MEMPTR=%04x IM%d IFF%s VPS: %d "+
		"MMU=%s",
		c.PC, c.SP, c.AF(), c.BC(), c.HL(), c.DE(), c.IX, c.IY, c.AF_, c.BC_, c.HL_, c.DE_, c.I, c.RegR(),
		zrcpFlags(c.F), zrcpFlags(byte(c.AF_)), 0, c.IM, iff, vps, strings.Repeat("0000", 8))
}

func (s *Server) zrcpRegisters() string {
	var r string
	vps := 50
	if s.sess.Settings().NTSC {
		vps = 60
	}
	s.sess.Do(func(m *zx81.Machine) { r = ZRCPRegisters(m.CPU, vps) })
	return r
}

func zrcpSetRegister(s *Server, p string) string {
	name, val, ok := strings.Cut(strings.TrimSpace(p), "=")
	if !ok || val == "" {
		return "Error changing register"
	}
	v, err := session.ParseNumber(val)
	if err != nil {
		return "Error changing register"
	}
	w, b := uint16(v), byte(v)
	bad := false
	s.sess.Do(func(m *zx81.Machine) {
		c := m.CPU
		switch strings.ToUpper(name) {
		case "PC":
			c.PC = w
		case "SP":
			c.SP = w
		case "AF":
			c.SetAF(w)
		case "BC":
			c.SetBC(w)
		case "DE":
			c.SetDE(w)
		case "HL":
			c.SetHL(w)
		case "IX":
			c.IX = w
		case "IY":
			c.IY = w
		case "AF'":
			c.AF_ = w
		case "BC'":
			c.BC_ = w
		case "DE'":
			c.DE_ = w
		case "HL'":
			c.HL_ = w
		case "A":
			c.A = b
		case "F":
			c.F = b
		case "B":
			c.B = b
		case "C":
			c.C = b
		case "D":
			c.D = b
		case "E":
			c.E = b
		case "H":
			c.H = b
		case "L":
			c.L = b
		case "I":
			c.I = b
		case "R":
			c.R, c.R7 = uint16(b&0x7f), b&0x80
		default:
			bad = true
		}
	})
	if bad {
		return "Error changing register"
	}
	return s.zrcpRegisters()
}

func zrcpReadMemory(s *Server, p string) string {
	f := strings.Fields(p)
	start, length := 0, 0
	if len(f) > 0 {
		start, length = zrcpNumber(f[0]), 1
		if len(f) > 1 {
			length = zrcpNumber(f[1])
		}
	}
	if length <= 0 {
		length = 65536 // whole memory zone
	}
	var sb strings.Builder
	s.sess.Do(func(m *zx81.Machine) {
		for i := 0; i < length; i++ {
			fmt.Fprintf(&sb, "%02X", m.Peek(uint16(start+i)))
		}
	})
	return sb.String()
}

func zrcpWriteMemory(s *Server, p string) string {
	f := strings.Fields(p)
	if len(f) == 0 {
		return "ERROR. No parameters set"
	}
	addr := zrcpNumber(f[0])
	s.sess.Do(func(m *zx81.Machine) {
		for i, t := range f[1:] {
			m.Poke(uint16(addr+i), byte(zrcpNumber(t)))
		}
	})
	return ""
}

func zrcpSendKeys(s *Server, p string) string {
	ms, text, ok := strings.Cut(p, " ")
	if !ok || text == "" {
		return "ERROR. Needs two parameters minimum"
	}
	hold := zrcpNumber(ms) / 20 // 20 ms per frame
	if hold < 1 {
		hold = 1
	}
	var sb strings.Builder
	for _, r := range text {
		if zx81.ChordForRune(r) != "" {
			sb.WriteRune(r)
		}
	}
	if err := s.sess.Type(sb.String(), hold); err != nil {
		return "ERROR. " + err.Error()
	}
	return ""
}

// zrcpOCR returns the text screen in plain ASCII: block graphics become '#',
// one line per row.
func (s *Server) zrcpOCR() string {
	var lines []string
	s.sess.Do(func(m *zx81.Machine) { lines = m.ScreenText() })
	var sb strings.Builder
	for _, l := range lines {
		for _, r := range l {
			if r > 126 {
				r = '#'
			}
			sb.WriteRune(r)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func zrcpHelp(s *Server, p string) string {
	p = strings.TrimSpace(p)
	var sb strings.Builder
	if p != "" {
		c := lookupZRCP(p)
		if c == nil {
			return "Unknown command"
		}
		sb.WriteString("Syntax: " + c.names[0])
		for _, a := range c.names[1:] {
			sb.WriteString("|" + a)
		}
		if c.params != "" {
			sb.WriteString(" " + c.params)
		}
		sb.WriteString("\n" + c.help + "\n")
		return sb.String()
	}
	sb.WriteString("Available commands (Go81 ZRCP subset):\n")
	for _, c := range zrcpCommands {
		fmt.Fprintf(&sb, "%-20s %s\n", strings.Join(c.names, "|"), c.help)
	}
	sb.WriteString("quit|exit|logout     Closes connection\n")
	return sb.String()
}
