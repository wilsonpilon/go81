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
