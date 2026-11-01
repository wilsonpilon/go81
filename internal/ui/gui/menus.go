// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"io"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/session"
)

// The keyboard pictures come from EightyOne (KBPics, GPL v2 or later).
var (
	//go:embed teclado-zx81.png
	keyboardZX81PNG []byte
	//go:embed teclado-ts1000.png
	keyboardTS1000PNG []byte
	//go:embed teclado-ts1500.png
	keyboardTS1500PNG []byte
	//go:embed teclado-tk85.png
	keyboardTK85PNG []byte
	//go:embed teclado-r470.png
	keyboardR470PNG []byte
)

// todoItem is a menu item of EightyOne that Go81 does not have yet: it
// says so in a dialog.
func todoItem(w fyne.Window, label string) *fyne.MenuItem {
	return fyne.NewMenuItem(label, func() { showTodo(w, label) })
}

func showTodo(w fyne.Window, label string) {
	name := strings.TrimSuffix(label, "…")
	dialog.ShowInformation("A implementar",
		"«"+name+"»\n\nEsta opção do EightyOne ainda não foi implementada no Go81.", w)
}

// todoMenu is a submenu whose items are all to be done.
func todoMenu(w fyne.Window, label string, items ...string) *fyne.MenuItem {
	m := fyne.NewMenuItem(label, nil)
	var children []*fyne.MenuItem
	for _, it := range items {
		if it == "-" {
			children = append(children, fyne.NewMenuItemSeparator())
			continue
		}
		children = append(children, todoItem(w, it))
	}
	m.ChildMenu = fyne.NewMenu("", children...)
	return m
}

// zoomImage is a picture that doubles its size with a double click (and
// goes back with another).
type zoomImage struct {
	widget.BaseWidget
	img    *canvas.Image
	size   fyne.Size // at 1×
	big    bool
	onZoom func()
}

func newZoomImage(img image.Image, onZoom func()) *zoomImage {
	b := img.Bounds()
	z := &zoomImage{img: canvas.NewImageFromImage(img), size: fyne.NewSize(float32(b.Dx()), float32(b.Dy())), onZoom: onZoom}
	z.img.FillMode = canvas.ImageFillContain
	z.img.ScaleMode = canvas.ImageScalePixels
	z.img.SetMinSize(z.size)
	z.ExtendBaseWidget(z)
	return z
}

func (z *zoomImage) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(z.img) }

func (z *zoomImage) DoubleTapped(*fyne.PointEvent) {
	z.big = !z.big
	s := z.size
	if z.big {
		s = fyne.NewSize(s.Width*2, s.Height*2)
	}
	z.img.SetMinSize(s)
	z.Refresh()
	if z.onZoom != nil {
		z.onZoom()
	}
}

// Tapped is needed for DoubleTapped to be called.
func (z *zoomImage) Tapped(*fyne.PointEvent) {}

// keyboardPicture is the keyboard of the model; for the ZX81, the TS1000
// one on the American (NTSC) machines.
func keyboardPicture(model string, ntsc bool) (image.Image, error) {
	data := keyboardZX81PNG
	switch {
	case model == "ts1500":
		data = keyboardTS1500PNG
	case model == "tk85":
		data = keyboardTK85PNG
	case model == "r470":
		data = keyboardR470PNG
	case ntsc:
		data = keyboardTS1000PNG
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// showKeyboardHelp opens Ajuda → Teclado: the picture of the keyboard
// (double click enlarges it) and how the PC keys map to it.
func (e *emulator) showKeyboardHelp(a fyne.App) {
	set := e.sess.Settings()
	img, err := keyboardPicture(set.Model, set.NTSC)
	if err != nil {
		return
	}
	title := "Teclado do ZX81"
	if m, err := session.FindModel(set.Model); err == nil && m.ID != "zx81" {
		title = "Teclado do " + m.Name
	}
	win := a.NewWindow(title)
	pic := newZoomImage(img, nil)
	pic.onZoom = func() { win.Resize(win.Content().MinSize()) }
	hint := widget.NewLabel("Clique duas vezes na imagem para aumentar ou diminuir.")
	hint.Alignment = fyne.TextAlignCenter
	help := widget.NewLabel(KeyHelp)
	win.SetContent(container.NewVBox(pic, hint, widget.NewSeparator(), help))
	win.Show()
}

// askNumbers asks for addresses (decimal, $hex or 0x hex) in a form.
func askNumbers(w fyne.Window, title string, labels, defaults []string, done func([]int)) {
	entries := make([]*widget.Entry, len(labels))
	var items []*widget.FormItem
	for i, l := range labels {
		entries[i] = widget.NewEntry()
		entries[i].SetText(defaults[i])
		items = append(items, widget.NewFormItem(l, entries[i]))
	}
	dialog.ShowForm(title, "OK", "Cancelar", items, func(ok bool) {
		if !ok {
			return
		}
		nums := make([]int, len(entries))
		for i, en := range entries {
			n, err := session.ParseNumber(strings.TrimSpace(en.Text))
			if err != nil || n < 0 || n > 0xFFFF+1 {
				dialog.ShowError(fmt.Errorf("%s: número inválido: %s", labels[i], en.Text), w)
				return
			}
			nums[i] = n
		}
		done(nums)
	}, w)
}

// loadMemoryBlock reads a file into the memory at an address.
func (e *emulator) loadMemoryBlock(w fyne.Window) {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			return
		}
		data, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		askNumbers(w, "Carregar bloco de memória", []string{"Endereço"}, []string{"16384"}, func(n []int) {
			if n[0]+len(data) > 0x10000 {
				dialog.ShowError(fmt.Errorf("o bloco (%d bytes) passa do fim da memória", len(data)), w)
				return
			}
			e.sess.Do(func(m *zx81.Machine) {
				for i, b := range data {
					m.Poke(uint16(n[0]+i), b)
				}
			})
			e.flash(fmt.Sprintf("%d bytes carregados em %d ($%04X)", len(data), n[0], n[0]))
		})
	}, w)
	d.Show()
}

// saveMemoryBlock writes a piece of the memory to a file.
func (e *emulator) saveMemoryBlock(w fyne.Window) {
	askNumbers(w, "Salvar bloco de memória", []string{"Endereço", "Tamanho"}, []string{"16384", "1024"}, func(n []int) {
		start, size := n[0], n[1]
		if size < 1 || start+size > 0x10000 {
			dialog.ShowError(fmt.Errorf("bloco inválido: %d bytes a partir de %d", size, start), w)
			return
		}
		data := make([]byte, size)
		e.sess.Do(func(m *zx81.Machine) {
			for i := range data {
				data[i] = m.Peek(uint16(start + i))
			}
		})
		d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			defer wc.Close()
			if _, err := wc.Write(data); err != nil {
				dialog.ShowError(err, w)
			}
		}, w)
		d.SetFileName(fmt.Sprintf("memoria-%04X.bin", start))
		d.Show()
	})
}

// invert turns the picture into inverse video (View → Inverse Video).
func invert(img *image.RGBA) {
	p := img.Pix
	for i := 0; i+3 < len(p); i += 4 {
		p[i], p[i+1], p[i+2] = 255-p[i], 255-p[i+1], 255-p[i+2]
	}
}
