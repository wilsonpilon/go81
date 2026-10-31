// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package session

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// ScreenFormats are the file types the screen can be saved as.
var ScreenFormats = []string{".png", ".bmp", ".pbm", ".txt"}

// ScreenFormat returns the format for a file name (by its extension), or
// an error listing the formats.
func ScreenFormat(name string) (string, error) {
	ext := strings.ToLower(filepath.Ext(name))
	for _, f := range ScreenFormats {
		if ext == f {
			return f, nil
		}
	}
	return "", fmt.Errorf("formato de tela desconhecido %q: use %s", ext, strings.Join(ScreenFormats, ", "))
}

// EncodeScreen writes the screen in a format (".png", ".bmp", ".pbm" or
// ".txt"). The text format is the 32x24 characters (block graphics as
// Unicode quadrants) and ignores the border.
func EncodeScreen(w io.Writer, format string, m *zx81.Machine, border int) error {
	switch format {
	case ".txt":
		return writeText(w, m.ScreenText())
	case ".png":
		return png.Encode(w, m.Screen(border))
	case ".bmp":
		return bmp.Encode(w, m.Screen(border))
	case ".pbm":
		return writePBM(w, m.Screen(border))
	}
	return fmt.Errorf("formato de tela desconhecido: %s", format)
}

// SaveScreen writes the screen to a file, in the format of its extension.
func SaveScreen(path string, m *zx81.Machine, border int) error {
	format, err := ScreenFormat(path)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := EncodeScreen(f, format, m, border); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func writeText(w io.Writer, lines []string) error {
	bw := bufio.NewWriter(w)
	for _, l := range lines {
		bw.WriteString(strings.TrimRight(l, " "))
		bw.WriteByte('\n')
	}
	return bw.Flush()
}

// writePBM writes a binary PBM (P4): 1 bit per pixel, 1 = black.
func writePBM(w io.Writer, img *image.RGBA) error {
	b := img.Bounds()
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "P4\n%d %d\n", b.Dx(), b.Dy())
	row := make([]byte, (b.Dx()+7)/8)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		clear(row)
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if int(c.R)+int(c.G)+int(c.B) < 3*128 {
				i := x - b.Min.X
				row[i/8] |= 0x80 >> (i % 8)
			}
		}
		bw.Write(row)
	}
	return bw.Flush()
}
