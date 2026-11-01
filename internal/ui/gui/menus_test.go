// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"image"
	"testing"
)

func TestKeyboardPictures(t *testing.T) {
	for _, model := range []string{"", "zx81", "ts1000", "ts1500", "tk85", "r470", "zx80"} {
		for _, ntsc := range []bool{false, true} {
			img, err := keyboardPicture(model, ntsc)
			if err != nil {
				t.Fatal(err)
			}
			if b := img.Bounds(); b.Dx() != 505 || b.Dy() != 223 {
				t.Fatalf("%s ntsc=%v: %v", model, ntsc, b)
			}
		}
	}
}

func TestInvert(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	copy(img.Pix, []byte{0, 0, 0, 255, 255, 200, 10, 255})
	invert(img)
	want := []byte{255, 255, 255, 255, 0, 55, 245, 255}
	for i := range want {
		if img.Pix[i] != want[i] {
			t.Fatalf("pix %v, want %v", img.Pix, want)
		}
	}
}
