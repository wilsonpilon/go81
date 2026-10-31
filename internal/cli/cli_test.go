// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Main([]string{"--version"}, &out, &errb); code != ExitOK {
		t.Fatalf("exit %d", code)
	}
	if !strings.HasPrefix(out.String(), "go81 ") {
		t.Fatalf("unexpected output %q", out.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Main([]string{"nope"}, &out, &errb); code != ExitUsage {
		t.Fatalf("exit %d, want %d", code, ExitUsage)
	}
}

func TestRunScreenshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tela.png")
	var out, errb bytes.Buffer
	code := Main([]string{"run", "--frames", "160", "--screenshot", path}, &out, &errb)
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 320 || b.Dy() != 240 {
		t.Fatalf("screenshot is %v, want 320x240", b)
	}
	// The boot screen has the inverse K cursor at the bottom left:
	// display column 0, row 23 -> image (32..39, 208..215); the display file
	// starts 24 lines below the top of a "normal" border image.
	dark := 0
	for y := 208; y < 216; y++ {
		for x := 32; x < 40; x++ {
			if r, _, _, _ := img.At(x, y).RGBA(); r < 0x8000 {
				dark++
			}
		}
	}
	if dark < 20 {
		t.Fatalf("cursor not found at the bottom left (%d dark pixels)", dark)
	}
}

func TestRunBadRAM(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Main([]string{"run", "--ram", "3"}, &out, &errb); code != ExitUsage {
		t.Fatalf("exit %d, want %d", code, ExitUsage)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	prog := filepath.Join(dir, "oi.p")
	var out, errb bytes.Buffer
	code := Main([]string{"run", "--frames", "900",
		"--keys", "1,0,P,SHIFT+P,O,I,SHIFT+P,ENTER,S,SHIFT+P,O,I,SHIFT+P,ENTER",
		"--save", prog}, &out, &errb)
	if code != ExitOK {
		t.Fatalf("save: exit %d: %s", code, errb.String())
	}
	data, err := os.ReadFile(prog)
	if err != nil || len(data) < 100 {
		t.Fatalf("saved file: %d bytes, %v", len(data), err)
	}
	// NXTLIN must point into the program, as a real SAVE leaves it.
	if nxt := int(data[0x4029-0x4009]) | int(data[0x402A-0x4009])<<8; nxt < 0x4009 || nxt > 0x4009+len(data) {
		t.Fatalf("NXTLIN %04x outside the program", nxt)
	}
	shot := filepath.Join(dir, "oi.png")
	code = Main([]string{"run", "--frames", "600", "--keys-start", "450", "--keys", "R,ENTER",
		"--screenshot", shot, prog}, &out, &errb)
	if code != ExitOK {
		t.Fatalf("load: exit %d: %s", code, errb.String())
	}
}
