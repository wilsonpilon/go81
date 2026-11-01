// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
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

func TestScreenKeyboard(t *testing.T) {
	sess, err := session.New(session.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	e := newEmulator(sess, 0)
	changes := 0
	e.onScreenKeys = func() { changes++ }
	rows := func() [8]byte {
		e.mu.Lock()
		defer e.mu.Unlock()
		return e.rowsLocked()
	}
	chord := func(c string) [8]byte {
		r, err := zx81.ParseChord(c)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	// SHIFT locks on the click and stays down while P is held...
	e.screenKey("SHIFT", true)
	e.screenKey("SHIFT", false)
	e.screenKey("P", true)
	if rows() != chord("SHIFT+P") || !e.screenKeyDown("SHIFT") || !e.screenKeyDown("P") {
		t.Fatalf("SHIFT+P: %v", rows())
	}
	// ...and goes up with P.
	e.screenKey("P", false)
	if rows() != ([8]byte{}) || e.screenKeyDown("SHIFT") {
		t.Fatalf("after P: %v", rows())
	}
	// A second click on SHIFT unlocks it.
	e.screenKey("SHIFT", true)
	e.screenKey("SHIFT", true)
	if e.screenKeyDown("SHIFT") {
		t.Fatal("SHIFT still locked")
	}
	// PC keys and screen keys add up; losing the focus releases only the PC ones.
	e.key("A", true)
	e.screenKey("ENTER", true)
	want := chord("A")
	for i, b := range chord("ENTER") {
		want[i] |= b
	}
	if rows() != want {
		t.Fatalf("A+ENTER: %v", rows())
	}
	e.releaseAll()
	if rows() != chord("ENTER") {
		t.Fatalf("after releaseAll: %v", rows())
	}
	if changes != 7 {
		t.Fatalf("%d notifications", changes)
	}
}

func TestSaveReachesWindow(t *testing.T) {
	sess, err := session.New(session.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	e := newEmulator(sess, 0)
	var gotName string
	var gotProg []byte
	e.onSaved = func(name string, prog []byte) { gotName, gotProg = name, prog }

	// A SAVE typed in the ZX81 reaches the window...
	for _, l := range []string{`typeln 10P"OI"`, `typeln S"JOGO"`, "run 300"} {
		if err := sess.Exec(l, io.Discard); err != nil {
			t.Fatal(err)
		}
	}
	if gotName != "JOGO" || len(gotProg) < 100 {
		t.Fatalf("saved %q, %d bytes", gotName, len(gotProg))
	}
	// ...but the console's save command writes its own file.
	gotName = ""
	p := filepath.Join(t.TempDir(), "x.p")
	if err := sess.Exec("save "+p, io.Discard); err != nil {
		t.Fatal(err)
	}
	if gotName != "" {
		t.Fatalf("save command also opened the window dialog (%q)", gotName)
	}
}
