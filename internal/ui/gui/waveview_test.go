// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/wilsonpilon/go81/internal/media/tape"
)

// renderWave draws a column per sample range, from the lowest to the
// highest sample, and the cursor where the tape plays.
func TestRenderWave(t *testing.T) {
	wv := &tape.SampleWave{S: []byte{128, 255, 0, 128}, Pos: 2}
	img := renderWave(wv, 0, 4, 4, 256, 2, 0)
	if img.RGBAAt(1, 0) != waveLine || img.RGBAAt(1, 200) != waveBack {
		t.Fatal("column 1 (255) not drawn at the top")
	}
	if img.RGBAAt(2, 0) != waveCursor {
		t.Fatal("no cursor at sample 2")
	}
	if img.RGBAAt(0, 0) != waveBack || img.RGBAAt(0, 127) != waveLine {
		t.Fatal("column 0 (silence) wrong")
	}
	if renderWave(nil, 0, 0, 10, 10, -1, 0) == nil {
		t.Fatal("no image for no sound")
	}
}

// The waveform window shows the tape as sound, then the WAV put in the
// input, following where it plays.
func TestWaveWindowScreenshot(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(theme.DefaultTheme())
	sess := tapeWithPrograms(t)
	e := newEmulator(sess, 2)
	wv, name, _ := e.pickWave(waveSourceAuto)
	if name != waveSourceNames[waveSourceTape] || wv == nil || wv.Len() < 44100 {
		t.Fatalf("auto source: %q, %v", name, wv)
	}
	wav := filepath.Join(t.TempDir(), "fita.wav")
	for _, line := range []string{`tape save "` + wav + `"`, "reset", "run 150", `tape insert "` + wav + `" som`, `typeln J""`, "run 120"} {
		if err := sess.Exec(line, io.Discard); err != nil {
			t.Fatalf("%s: %v", line, err)
		}
	}
	wv, name, thresh := e.pickWave(waveSourceAuto)
	if name != waveSourceNames[waveSourceIn] || wv.Position() <= 0 || thresh == 0 {
		t.Fatalf("WAV source: %q pos %d thresh %d", name, wv.Position(), thresh)
	}
	win := e.waveWindow(a, func() {})
	win.Show()
	if dir := os.Getenv("GO81_SCREENSHOTS"); dir != "" {
		out, _ := filepath.Abs(dir)
		if err := savePNG(filepath.Join(out, "gui-fita-som.png"), win.Canvas().Capture()); err != nil {
			t.Fatal(err)
		}
	}
	win.Close()
}
