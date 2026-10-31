// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"testing"

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/emu/zx81/zx81ref"
)

func TestAudioQueue(t *testing.T) {
	a := &audioOut{}
	a.push([]int16{1, -1, 2, -2})
	p := make([]byte, 12)
	if n, _ := a.Read(p); n != 12 {
		t.Fatalf("Read returned %d", n)
	}
	want := []byte{1, 0, 0xff, 0xff, 2, 0, 0xfe, 0xff, 0, 0, 0, 0} // then silence
	for i := range want {
		if p[i] != want[i] {
			t.Fatalf("byte %d = %x, want %x", i, p[i], want[i])
		}
	}
	// A long backlog is trimmed to keepQueue, keeping whole samples.
	big := make([]int16, maxQueue) // 2*maxQueue bytes
	a.push(big)
	if len(a.queue) != keepQueue {
		t.Fatalf("queue not trimmed: %d bytes", len(a.queue))
	}
	if len(a.queue)%4 != 0 {
		t.Fatalf("queue cut mid-sample: %d bytes", len(a.queue))
	}
	var nilOut *audioOut
	nilOut.push([]int16{1}) // no sound card: must not panic
}

func TestSaveReachesWindow(t *testing.T) {
	rom, _ := assets.ROM("zx81.edition3")
	e := &emulator{}
	if err := e.newMachine(zx81.DefaultConfig(rom)); err != nil {
		t.Fatal(err)
	}
	var gotName string
	var gotProg []byte
	e.onSaved = func(name string, prog []byte) { gotName, gotProg = name, prog }

	keys, end := zx81ref.Type(150, 6, 6, 60,
		"1", "0", "P", "SHIFT+P", "O", "I", "SHIFT+P", "ENTER",
		"S", "SHIFT+P", "J", "O", "G", "O", "SHIFT+P", "ENTER")
	ev := 0
	for f := 0; f < end+100; f++ {
		for ev < len(keys) && keys[ev].Frame <= f {
			e.m.Keyboard[keys[ev].Row&7] = byte(keys[ev].Bits)
			ev++
		}
		e.m.RunFrame()
	}
	if gotName != "JOGO" {
		t.Fatalf("saved name %q", gotName)
	}
	if len(gotProg) < 100 {
		t.Fatalf("saved program has %d bytes", len(gotProg))
	}
}
