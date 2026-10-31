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
	"strings"
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

	"github.com/wilsonpilon/go81/assets"
	"github.com/wilsonpilon/go81/internal/emu/zx81"
	"github.com/wilsonpilon/go81/internal/media/tape"
	"github.com/wilsonpilon/go81/internal/version"
)

// Available reports whether this build has the graphical interface.
const Available = true

// Options configure the window.
type Options struct {
	Config   zx81.Config
	ROMName  string // name shown in the hardware dialog
	RAMK     int    // RAM pack in K (0 = 1K internal)
	TapeName string // program to load at start (optional)
	TapeData []byte
	Zoom     int
	Border   int
	NoSound  bool // do not open the sound card

	// For automation: after ExitAfter frames, save the window picture to
	// CaptureFile (if set) and close.
	ExitAfter   int
	CaptureFile string
}

type emulator struct {
	mu      sync.Mutex
	m       *zx81.Machine
	cfg     zx81.Config
	romName string
	ramK    int
	pressed map[string]bool
	paused  bool
	turbo   bool
	border  int
	fps     float64
	saving  []byte
	audio   *audioOut

	// onSaved is called (from the emulation goroutine) when the program
	// running in the ZX81 finishes a SAVE.
	onSaved func(name string, prog []byte)
}

// newMachine replaces the machine (hard reset) with a new configuration.
// The caller holds e.mu.
func (e *emulator) newMachine(cfg zx81.Config) error {
	m, err := zx81.New(cfg)
	if err != nil {
		return err
	}
	// The tape deck is always the SAVE output, so SAVE goes to a file.
	m.Tape = tape.NewDeck(3250000)
	m.Tape.Out = true
	m.SaveByte = func(b byte) { e.saving = append(e.saving, b) }
	m.SaveEnd = func() {
		if len(e.saving) == 0 {
			return
		}
		name, prog, ok := tape.SplitSaved(e.saving)
		e.saving = nil
		if ok && e.onSaved != nil {
			e.onSaved(name, prog)
		}
	}
	e.m, e.cfg = m, cfg
	e.pressed = map[string]bool{}
	return nil
}

func (e *emulator) insert(name string, data []byte) error {
	block, err := tape.ProgramBlock(name, data)
	if err != nil {
		return err
	}
	e.m.InsertTape(block, true)
	return nil
}

// description is the machine summary for the status bar. The caller holds e.mu.
func (e *emulator) description() string {
	ram := e.ramK
	if ram == 0 {
		ram = 1
	}
	s := fmt.Sprintf("ZX81 %dK", ram)
	if e.cfg.NTSC {
		s += " NTSC"
	} else {
		s += " PAL"
	}
	if e.cfg.HiRes == zx81.HiResWRX {
		s += " · WRX"
	}
	if e.cfg.VSyncSound {
		s += " · som"
	}
	return s
}

// Run opens the window and blocks until it is closed.
func Run(opts Options) error {
	e := &emulator{romName: opts.ROMName, ramK: opts.RAMK, border: opts.Border}
	if err := e.newMachine(opts.Config); err != nil {
		return err
	}
	if opts.TapeData != nil {
		if err := e.insert(opts.TapeName, opts.TapeData); err != nil {
			return err
		}
	}
	if opts.Zoom < 1 {
		opts.Zoom = 2
	}
	if !opts.NoSound {
		e.audio = newAudio()
	}

	a := app.NewWithID("io.github.wilsonpilon.go81")
	w := a.NewWindow("Go81 " + version.Number())

	screen := canvas.NewImageFromImage(e.m.Screen(e.border))
	screen.FillMode = canvas.ImageFillContain
	screen.ScaleMode = canvas.ImageScalePixels
	zoom := opts.Zoom
	setZoom := func(z int) {
		zoom = z
		e.mu.Lock()
		r := e.m.Screen(e.border).Bounds()
		e.mu.Unlock()
		screen.SetMinSize(fyne.NewSize(float32(r.Dx()*z), float32(r.Dy()*z)))
		w.Resize(fyne.NewSize(float32(r.Dx()*z), float32(r.Dy()*z+30)))
	}
	status := widget.NewLabel("")
	w.SetContent(container.NewBorder(nil, status, nil, nil, screen))
	setZoom(zoom)

	if dc, ok := w.Canvas().(desktop.Canvas); ok {
		dc.SetOnKeyDown(func(k *fyne.KeyEvent) { e.key(string(k.Name), true) })
		dc.SetOnKeyUp(func(k *fyne.KeyEvent) { e.key(string(k.Name), false) })
	}

	e.onSaved = func(name string, prog []byte) {
		fyne.Do(func() { askSave(w, name, prog) })
	}

	w.SetMainMenu(e.menu(a, w, func() int { return zoom }, setZoom))
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
	})

	w.SetOnClosed(func() { close(stop) })
	w.ShowAndRun()
	return runErr
}

// askSave asks where to write a program the ZX81 just saved.
func askSave(w fyne.Window, name string, prog []byte) {
	d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		defer wc.Close()
		if _, err := wc.Write(prog); err != nil {
			dialog.ShowError(err, w)
		}
	}, w)
	file := strings.TrimSpace(name)
	if file == "" {
		file = "programa"
	}
	d.SetFileName(file + ".p")
	d.SetFilter(storage.NewExtensionFileFilter([]string{".p"}))
	d.Show()
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

func framePeriod(ntsc bool) time.Duration {
	if ntsc {
		return time.Second / 60
	}
	return time.Second / 50
}

// loop runs one frame per 1/50 s (1/60 in NTSC), sends the sound to the
// sound card and hands the picture to show.
func (e *emulator) loop(stop chan struct{}, show func(image.Image, string)) {
	e.mu.Lock()
	ntsc := e.cfg.NTSC
	e.mu.Unlock()
	tick := time.NewTicker(framePeriod(ntsc))
	defer tick.Stop()
	frames, last := 0, time.Now()
	for {
		select {
		case <-stop:
			return
		case <-tick.C:
		}
		e.mu.Lock()
		if e.cfg.NTSC != ntsc {
			ntsc = e.cfg.NTSC
			tick.Reset(framePeriod(ntsc))
		}
		var audio []int16
		if !e.paused {
			n := 1
			if e.turbo {
				n = 10
			}
			for i := 0; i < n; i++ {
				e.m.RunFrame()
			}
			frames += n
			if !e.turbo && e.cfg.VSyncSound {
				audio = append([]int16(nil), e.m.Audio()...)
			}
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
		line := fmt.Sprintf("%s · %.0f quadros/s%s", e.description(), e.fps, state)
		e.mu.Unlock()
		if audio != nil {
			e.audio.push(audio)
		}
		show(img, line)
	}
}

// hardwareDialog lets the user change the machine (applies with a reset).
func (e *emulator) hardwareDialog(w fyne.Window, setZoom func(int), zoom func() int) {
	ramLabels := []string{"1K (só a interna)", "2K", "4K", "8K", "16K", "32K", "48K"}
	ramValues := []int{0, 2, 4, 8, 16, 32, 48}

	e.mu.Lock()
	cfg, romName, ramK := e.cfg, e.romName, e.ramK
	e.mu.Unlock()

	roms := assets.ROMNames()
	known := false
	for _, r := range roms {
		if r == romName {
			known = true
		}
	}
	if !known && romName != "" {
		roms = append(roms, romName)
	}
	romSel := widget.NewSelect(roms, nil)
	romSel.SetSelected(romName)
	ramSel := widget.NewSelect(ramLabels, nil)
	for i, v := range ramValues {
		if v == ramK {
			ramSel.SetSelectedIndex(i)
		}
	}
	ntsc := widget.NewCheck("NTSC — TS1000 / ZX81 americano (60 Hz)", nil)
	ntsc.SetChecked(cfg.NTSC)
	wrx := widget.NewCheck("Alta resolução WRX", nil)
	wrx.SetChecked(cfg.HiRes == zx81.HiResWRX)
	snd := widget.NewCheck("Som pelo sinal de vídeo (Video Sound)", nil)
	snd.SetChecked(cfg.VSyncSound)

	items := []*widget.FormItem{
		widget.NewFormItem("ROM", romSel),
		widget.NewFormItem("Memória", ramSel),
		widget.NewFormItem("", ntsc),
		widget.NewFormItem("", wrx),
		widget.NewFormItem("", snd),
	}
	dialog.ShowForm("Configuração do hardware", "Aplicar e reiniciar", "Cancelar", items, func(ok bool) {
		if !ok {
			return
		}
		name := romSel.Selected
		rom := cfg.ROM
		if name != romName {
			var err error
			if rom, err = loadROM(name); err != nil {
				dialog.ShowError(err, w)
				return
			}
		}
		newCfg := zx81.DefaultConfig(rom)
		k := ramValues[max(ramSel.SelectedIndex(), 0)]
		newCfg.RAMTop = zx81.RAMTopForPack(k)
		newCfg.NTSC = ntsc.Checked
		if wrx.Checked {
			newCfg.HiRes = zx81.HiResWRX
		}
		newCfg.VSyncSound = snd.Checked
		e.mu.Lock()
		err := e.newMachine(newCfg)
		if err == nil {
			e.romName, e.ramK = name, k
		}
		e.mu.Unlock()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		setZoom(zoom())
	}, w)
}

func loadROM(name string) ([]byte, error) {
	if st, err := os.Stat(name); err == nil && !st.IsDir() {
		return os.ReadFile(name)
	}
	return assets.ROM(name)
}

func (e *emulator) menu(a fyne.App, w fyne.Window, zoom func() int, setZoom func(int)) *fyne.MainMenu {
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

	hw := fyne.NewMenuItem("Configuração do hardware…", func() { e.hardwareDialog(w, setZoom, zoom) })
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
	sound := fyne.NewMenuItem("Som (Video Sound)", nil)
	sound.Checked = e.cfg.VSyncSound
	machineMenu := fyne.NewMenu("Máquina", hw, fyne.NewMenuItemSeparator(), reset, pause, turbo, sound)
	var mainMenu *fyne.MainMenu
	sound.Action = func() {
		e.mu.Lock()
		on := !e.cfg.VSyncSound
		e.cfg.VSyncSound = on
		e.m.SetVSyncSound(on)
		e.mu.Unlock()
		sound.Checked = on
		machineMenu.Refresh()
		if on && e.audio == nil {
			dialog.ShowInformation("Som", "Nenhuma saída de som disponível.", w)
		}
	}

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
			setZoom(zoom())
		}))
	}
	view := fyne.NewMenu("Exibir", append(append(zooms, fyne.NewMenuItemSeparator()), borders...)...)

	keys := fyne.NewMenuItem("Teclado…", func() { dialog.ShowInformation("Teclado do ZX81", KeyHelp, w) })
	about := fyne.NewMenuItem("Sobre…", func() {
		dialog.ShowInformation("Sobre o Go81", "Go81 "+version.Full()+
			"\n\nEmulador de ZX81 em Go, compatível com o EightyOne.\nLicença GPL v3 ou posterior.", w)
	})

	mainMenu = fyne.NewMainMenu(
		fyne.NewMenu("Arquivo", open, shot, fyne.NewMenuItemSeparator(), quit),
		machineMenu,
		view,
		fyne.NewMenu("Ajuda", keys, about),
	)
	return mainMenu
}
