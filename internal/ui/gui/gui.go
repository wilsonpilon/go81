// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

// Package gui is the Fyne window of the emulator.
package gui

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
	"github.com/wilsonpilon/go81/internal/version"
)

// Available reports whether this build has the graphical interface.
const Available = true

// Options configure the window.
type Options struct {
	Config      zx81.Config
	TapeName    string // program to load at start (optional)
	TapeData    []byte
	Zoom        int
	Border      int
	Description string // shown in the status bar, e.g. "ZX81 16K PAL"

	// For automation: after ExitAfter frames, save the window picture to
	// CaptureFile (if set) and close.
	ExitAfter   int
	CaptureFile string
}

type emulator struct {
	mu      sync.Mutex
	m       *zx81.Machine
	cfg     zx81.Config
	pressed map[string]bool
	paused  bool
	turbo   bool
	border  int
	fps     float64
}

// Run opens the window and blocks until it is closed.
func Run(opts Options) error {
	m, err := zx81.New(opts.Config)
	if err != nil {
		return err
	}
	e := &emulator{m: m, cfg: opts.Config, pressed: map[string]bool{}, border: opts.Border}
	if opts.TapeData != nil {
		if err := e.insert(opts.TapeName, opts.TapeData); err != nil {
			return err
		}
	}
	if opts.Zoom < 1 {
		opts.Zoom = 2
	}

	a := app.NewWithID("io.github.wilsonpilon.go81")
	w := a.NewWindow("Go81 " + version.Number())

	screen := canvas.NewImageFromImage(m.Screen(e.border))
	screen.FillMode = canvas.ImageFillContain
	screen.ScaleMode = canvas.ImageScalePixels
	setZoom := func(z int) {
		r := m.Screen(e.border).Bounds()
		screen.SetMinSize(fyne.NewSize(float32(r.Dx()*z), float32(r.Dy()*z)))
		w.Resize(fyne.NewSize(float32(r.Dx()*z), float32(r.Dy()*z+30)))
	}
	status := widget.NewLabel("")
	w.SetContent(container.NewBorder(nil, status, nil, nil, screen))
	setZoom(opts.Zoom)

	if dc, ok := w.Canvas().(desktop.Canvas); ok {
		dc.SetOnKeyDown(func(k *fyne.KeyEvent) { e.key(string(k.Name), true) })
		dc.SetOnKeyUp(func(k *fyne.KeyEvent) { e.key(string(k.Name), false) })
	}

	w.SetMainMenu(e.menu(a, w, setZoom))
	// Release every key when the window loses focus, so none gets stuck.
	a.Lifecycle().SetOnExitedForeground(e.releaseAll)

	stop := make(chan struct{})
	var runErr error
	shown := 0
	go e.loop(stop, func(img image.Image, line string) {
		fyne.Do(func() {
			screen.Image = img
			screen.Refresh()
			status.SetText(line)
			shown++
			if opts.ExitAfter > 0 && shown == opts.ExitAfter {
				if opts.CaptureFile != "" {
					runErr = savePNG(opts.CaptureFile, w.Canvas().Capture())
				}
				a.Quit()
			}
		})
	}, opts.Description)

	w.SetOnClosed(func() { close(stop) })
	w.ShowAndRun()
	return runErr
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (e *emulator) insert(name string, data []byte) error {
	block, err := tape.ProgramBlock(name, data)
	if err != nil {
		return err
	}
	e.m.InsertTape(block, true)
	return nil
}

// releaseAll releases every key.
func (e *emulator) releaseAll() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pressed = map[string]bool{}
	e.m.SetKeys([8]byte{})
}

// key updates the ZX81 keyboard from the set of PC keys held down.
func (e *emulator) key(name string, down bool) {
	if _, ok := keyMap[name]; !ok {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if down {
		e.pressed[name] = true
	} else {
		delete(e.pressed, name)
	}
	var rows [8]byte
	for k := range e.pressed {
		r, err := zx81.ParseChord(keyMap[k])
		if err != nil {
			continue
		}
		for i := range rows {
			rows[i] |= r[i]
		}
	}
	e.m.SetKeys(rows)
}

// loop runs one frame per 1/50 s (1/60 in NTSC) and hands the picture to show.
func (e *emulator) loop(stop chan struct{}, show func(image.Image, string), desc string) {
	period := time.Second / 50
	if e.cfg.NTSC {
		period = time.Second / 60
	}
	tick := time.NewTicker(period)
	defer tick.Stop()
	frames, last := 0, time.Now()
	for {
		select {
		case <-stop:
			return
		case <-tick.C:
		}
		e.mu.Lock()
		n := 1
		if e.turbo {
			n = 10
		}
		if !e.paused {
			for i := 0; i < n; i++ {
				e.m.RunFrame()
			}
			frames += n
		}
		img := e.m.Screen(e.border)
		state := ""
		switch {
		case e.paused:
			state = " · PAUSADO"
		case e.turbo:
			state = " · TURBO"
		}
		if e.m.Tape != nil && e.m.Tape.Player.Playing {
			state += " · fita tocando"
		}
		if d := time.Since(last); d >= time.Second {
			e.fps = float64(frames) / d.Seconds()
			frames, last = 0, time.Now()
		}
		line := fmt.Sprintf("%s · %.0f quadros/s%s", desc, e.fps, state)
		e.mu.Unlock()
		show(img, line)
	}
}

func (e *emulator) menu(a fyne.App, w fyne.Window, setZoom func(int)) *fyne.MainMenu {
	open := fyne.NewMenuItem("Abrir programa…", func() {
		d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			defer r.Close()
			data, err := io.ReadAll(r)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			e.mu.Lock()
			err = e.insert(r.URI().Name(), data)
			e.mu.Unlock()
			if err != nil {
				dialog.ShowError(err, w)
			}
		}, w)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".p", ".P", ".81", ".p81", ".P81"}))
		d.Show()
	})
	shot := fyne.NewMenuItem("Salvar tela como PNG…", func() {
		e.mu.Lock()
		img := e.m.Screen(e.border)
		e.mu.Unlock()
		d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			defer wc.Close()
			if err := png.Encode(wc, img); err != nil {
				dialog.ShowError(err, w)
			}
		}, w)
		d.SetFileName("go81.png")
		d.Show()
	})
	quit := fyne.NewMenuItem("Sair", func() { a.Quit() })
	quit.IsQuit = true

	reset := fyne.NewMenuItem("Reiniciar (desligar e ligar)", func() {
		e.mu.Lock()
		e.m.Reset()
		e.mu.Unlock()
	})
	pause := fyne.NewMenuItem("Pausar / continuar", func() {
		e.mu.Lock()
		e.paused = !e.paused
		e.mu.Unlock()
	})
	pause.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF5}
	turbo := fyne.NewMenuItem("Turbo (10×)", func() {
		e.mu.Lock()
		e.turbo = !e.turbo
		e.mu.Unlock()
	})
	turbo.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF6}

	var zooms []*fyne.MenuItem
	for z := 1; z <= 4; z++ {
		z := z
		zooms = append(zooms, fyne.NewMenuItem(fmt.Sprintf("Zoom %d×", z), func() { setZoom(z) }))
	}
	borderNames := []string{"Sem borda", "Borda pequena", "Borda normal", "Borda grande", "Borda total"}
	var borders []*fyne.MenuItem
	for b, name := range borderNames {
		b := b
		borders = append(borders, fyne.NewMenuItem(name, func() {
			e.mu.Lock()
			e.border = b
			e.mu.Unlock()
			setZoom(2)
		}))
	}
	view := fyne.NewMenu("Exibir", append(append(zooms, fyne.NewMenuItemSeparator()), borders...)...)

	keys := fyne.NewMenuItem("Teclado…", func() { dialog.ShowInformation("Teclado do ZX81", KeyHelp, w) })
	about := fyne.NewMenuItem("Sobre…", func() {
		dialog.ShowInformation("Sobre o Go81", "Go81 "+version.Full()+
			"\n\nEmulador de ZX81 em Go, compatível com o EightyOne.\nLicença GPL v3 ou posterior.", w)
	})

	return fyne.NewMainMenu(
		fyne.NewMenu("Arquivo", open, shot, fyne.NewMenuItemSeparator(), quit),
		fyne.NewMenu("Máquina", reset, pause, turbo),
		view,
		fyne.NewMenu("Ajuda", keys, about),
	)
}
