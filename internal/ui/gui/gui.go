// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

// Package gui is the Fyne window of the emulator.
package gui

import (
	"bytes"
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
	"github.com/wilsonpilon/go81/internal/session"
	"github.com/wilsonpilon/go81/internal/version"
)

// Available reports whether this build has the graphical interface.
const Available = true

// Options configure the window.
type Options struct {
	// Session is the machine. The window runs its frames in real time, so
	// its commands (console, Link) wait for the window.
	Session *session.Session
	// ConfigPath is the configuration file ("" = the default one), where
	// "Salvar esta máquina como padrão" writes.
	ConfigPath string
	Zoom       int
	Border     int
	NoSound    bool // do not open the sound card
	Console    bool // open the command console at start

	// For automation: after ExitAfter frames, save the window picture to
	// CaptureFile (if set) and close.
	ExitAfter   int
	CaptureFile string
}

// emulator is the window state that is not the machine.
type emulator struct {
	sess *session.Session

	mu      sync.Mutex
	pressed map[string]bool
	ctrl    bool // a Control key is held down

	configPath string // the configuration file

	message string    // shown in the status bar for a few seconds
	msgEnd  time.Time // when the message goes away
	turbo   bool
	border  int
	fps     float64
	audio   *audioOut

	// onSaved is called when the program running in the ZX81 finishes a
	// SAVE (not the console's save command, which writes its own file).
	onSaved func(name string, prog []byte)
}

func newEmulator(sess *session.Session, border int) *emulator {
	e := &emulator{sess: sess, border: border, pressed: map[string]bool{}}
	sess.OnSave = func(name string, prog []byte) {
		if e.onSaved != nil {
			e.onSaved(name, prog)
		}
	}
	return e
}

// Run opens the window and blocks until it is closed.
func Run(opts Options) error {
	sess := opts.Session
	e := newEmulator(sess, opts.Border)
	sess.Drive()
	if opts.Zoom < 1 {
		opts.Zoom = 2
	}
	if !opts.NoSound {
		e.audio = newAudio()
	}

	a := app.NewWithID("io.github.wilsonpilon.go81")
	w := a.NewWindow("Go81 " + version.Number())

	screen := canvas.NewImageFromImage(e.picture())
	screen.FillMode = canvas.ImageFillContain
	screen.ScaleMode = canvas.ImageScalePixels
	zoom := opts.Zoom
	setZoom := func(z int) {
		zoom = z
		r := e.picture().Bounds()
		screen.SetMinSize(fyne.NewSize(float32(r.Dx()*z), float32(r.Dy()*z)))
		w.Resize(fyne.NewSize(float32(r.Dx()*z), float32(r.Dy()*z+30)))
	}
	status := widget.NewLabel("")
	w.SetContent(container.NewBorder(nil, status, nil, nil, screen))
	setZoom(zoom)

	con := newConsoleWindow(a, sess, func() { a.Quit() })
	if dc, ok := w.Canvas().(desktop.Canvas); ok {
		dc.SetOnKeyDown(func(k *fyne.KeyEvent) {
			if k.Name == fyne.KeyF2 { // console (also in the Tools menu)
				con.show()
				return
			}
			e.key(string(k.Name), true)
		})
		dc.SetOnKeyUp(func(k *fyne.KeyEvent) { e.key(string(k.Name), false) })
	}

	e.onSaved = func(name string, prog []byte) {
		fyne.Do(func() { askSave(w, name, prog) })
	}

	e.configPath = opts.ConfigPath
	if e.configPath == "" {
		e.configPath = session.ConfigPath()
	}
	w.SetMainMenu(e.menu(a, w, con, func() int { return zoom }, setZoom))
	// Release every key when the window loses focus, so none gets stuck.
	a.Lifecycle().SetOnExitedForeground(e.releaseAll)
	if opts.Console {
		a.Lifecycle().SetOnStarted(con.show)
	}

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

	w.SetOnClosed(func() { close(stop); a.Quit() })
	w.ShowAndRun()
	return runErr
}

// picture returns the TV picture with the current border.
func (e *emulator) picture() *image.RGBA {
	e.mu.Lock()
	b := e.border
	e.mu.Unlock()
	var img *image.RGBA
	e.sess.Do(func(m *zx81.Machine) { img = m.Screen(b) })
	return img
}

// flash shows a message in the status bar for a few seconds.
func (e *emulator) flash(msg string) {
	e.mu.Lock()
	e.message, e.msgEnd = msg, time.Now().Add(4*time.Second)
	e.mu.Unlock()
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
	e.pressed = map[string]bool{}
	e.mu.Unlock()
	e.sess.SetKeys([8]byte{})
}

// key updates the ZX81 keyboard from the set of PC keys held down.
func (e *emulator) key(name string, down bool) {
	if name == "LeftControl" || name == "RightControl" {
		e.mu.Lock()
		e.ctrl = down
		e.mu.Unlock()
		return
	}
	if _, ok := keyMap[name]; !ok {
		return
	}
	e.mu.Lock()
	if down && e.ctrl { // Ctrl+V, Ctrl+C...: window shortcuts, not ZX81 keys
		e.mu.Unlock()
		return
	}
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
	e.mu.Unlock()
	e.sess.SetKeys(rows)
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
	ntsc := e.sess.Settings().NTSC
	tick := time.NewTicker(framePeriod(ntsc))
	defer tick.Stop()
	frames, last := 0, time.Now()
	for {
		select {
		case <-stop:
			return
		case <-tick.C:
		}
		set := e.sess.Settings()
		if set.NTSC != ntsc {
			ntsc = set.NTSC
			tick.Reset(framePeriod(ntsc))
		}
		e.mu.Lock()
		n, turbo, border := 1, e.turbo, e.border
		e.mu.Unlock()
		if turbo {
			n = 10
		}
		ran := e.sess.Tick(n)
		if ran {
			frames += n
		}
		var audio []int16
		var img *image.RGBA
		playing := false
		e.sess.Do(func(m *zx81.Machine) {
			if ran && !turbo && set.Sound {
				audio = append([]int16(nil), m.Audio()...)
			}
			img = m.Screen(border)
			playing = m.Tape != nil && m.Tape.Player.Playing
		})
		state := ""
		switch {
		case e.sess.Paused():
			state = " · PAUSADO"
		case turbo:
			state = " · TURBO"
		}
		if playing {
			state += " · fita tocando"
		}
		e.mu.Lock()
		if d := time.Since(last); d >= time.Second {
			e.fps = float64(frames) / d.Seconds()
			frames, last = 0, time.Now()
		}
		line := fmt.Sprintf("%s · %s · %.0f quadros/s%s", e.sess.Description(), e.sess.Status().Text(), e.fps, state)
		if e.message != "" && time.Now().Before(e.msgEnd) {
			line = e.message
		}
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

	set := e.sess.Settings()
	roms := assets.ROMNames()
	known := false
	for _, r := range roms {
		if r == set.ROM {
			known = true
		}
	}
	if !known && set.ROM != "" {
		roms = append(roms, set.ROM)
	}
	romSel := widget.NewSelect(roms, nil)
	romSel.SetSelected(set.ROM)
	ramSel := widget.NewSelect(ramLabels, nil)
	for i, v := range ramValues {
		if v == set.RAMK {
			ramSel.SetSelectedIndex(i)
		}
	}
	ntsc := widget.NewCheck("NTSC — TS1000 / ZX81 americano (60 Hz)", nil)
	ntsc.SetChecked(set.NTSC)
	wrx := widget.NewCheck("Alta resolução WRX", nil)
	wrx.SetChecked(set.WRX)
	snd := widget.NewCheck("Som pelo sinal de vídeo (Video Sound)", nil)
	snd.SetChecked(set.Sound)

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
		newSet := session.Settings{
			ROM:   romSel.Selected,
			RAMK:  ramValues[max(ramSel.SelectedIndex(), 0)],
			NTSC:  ntsc.Checked,
			WRX:   wrx.Checked,
			Sound: snd.Checked,
		}
		if err := e.sess.Configure(newSet); err != nil {
			dialog.ShowError(err, w)
			return
		}
		setZoom(zoom())
	}, w)
}

// runCommand runs a shell command in the background and shows its output
// (or error) in a dialog.
func (e *emulator) runCommand(w fyne.Window, title, line string) {
	go func() {
		var out bytes.Buffer
		err := e.sess.Exec(line, &out)
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			dialog.ShowInformation(title, strings.TrimSpace(out.String()), w)
		})
	}()
}

func (e *emulator) menu(a fyne.App, w fyne.Window, con *consoleWindow, zoom func() int, setZoom func(int)) *fyne.MainMenu {
	open := fyne.NewMenuItem("Abrir programa…", func() {
		d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			defer r.Close()
			data, err := io.ReadAll(r)
			if err == nil {
				err = e.sess.InsertProgram(r.URI().Name(), data, true)
			}
			if err != nil {
				dialog.ShowError(err, w)
			}
		}, w)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".p", ".P", ".81", ".p81", ".P81", ".bas", ".BAS", ".txt", ".TXT", ".b81"}))
		d.Show()
	})
	shot := fyne.NewMenuItem("Salvar tela (PNG, BMP, PBM ou TXT)…", func() {
		e.mu.Lock()
		border := e.border
		e.mu.Unlock()
		d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			defer wc.Close()
			format, err := session.ScreenFormat(wc.URI().Name())
			if err == nil {
				e.sess.Do(func(m *zx81.Machine) { err = session.EncodeScreen(wc, format, m, border) })
			}
			if err != nil {
				dialog.ShowError(err, w)
			}
		}, w)
		d.SetFileName("go81.png")
		d.SetFilter(storage.NewExtensionFileFilter(session.ScreenFormats))
		d.Show()
	})
	quit := fyne.NewMenuItem("Sair", func() { a.Quit() })
	quit.IsQuit = true

	// Colar: a BASIC program is loaded, any other text is typed.
	paste := fyne.NewMenuItem("Colar (programa BASIC ou texto)", func() {
		msg, err := e.sess.Paste(a.Clipboard().Content())
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		e.flash(msg)
	})
	paste.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyV, Modifier: fyne.KeyModifierShortcutDefault}
	copyText := fyne.NewMenuItem("Copiar tela como texto", func() {
		var lines []string
		e.sess.Do(func(m *zx81.Machine) { lines = m.ScreenText() })
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " ")
		}
		a.Clipboard().SetContent(strings.Join(lines, "\n") + "\n")
	})
	copyText.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyC, Modifier: fyne.KeyModifierShortcutDefault}
	// Menu shortcuts must also be registered on the window to work.
	w.Canvas().AddShortcut(paste.Shortcut, func(fyne.Shortcut) { paste.Action() })
	w.Canvas().AddShortcut(copyText.Shortcut, func(fyne.Shortcut) { copyText.Action() })
	pasteKeys := fyne.NewMenuItem("Colar como teclas (digitar tudo)", func() {
		e.flash(e.sess.PasteKeys(a.Clipboard().Content()))
	})
	editMenu := fyne.NewMenu("Editar", paste, pasteKeys, fyne.NewMenuItemSeparator(), copyText)

	hw := fyne.NewMenuItem("Configuração do hardware…", func() { e.hardwareDialog(w, setZoom, zoom) })
	reset := fyne.NewMenuItem("Reiniciar (desligar e ligar)", e.sess.Reset)
	pause := fyne.NewMenuItem("Pausar / continuar", func() { e.sess.SetPaused(!e.sess.Paused()) })
	pause.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF5}
	turbo := fyne.NewMenuItem("Turbo (10×)", func() {
		e.mu.Lock()
		e.turbo = !e.turbo
		e.mu.Unlock()
	})
	turbo.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF6}
	sound := fyne.NewMenuItem("Som (Video Sound)", nil)
	sound.Checked = e.sess.Settings().Sound
	saveDefault := fyne.NewMenuItem("Salvar esta máquina como padrão", func() {
		if err := session.SaveDefaultMachine(e.configPath, e.sess.Settings()); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Configuração", "Máquina padrão gravada em\n"+e.configPath+
			"\n\n"+e.sess.Description(), w)
	})
	machineMenu := fyne.NewMenu("Máquina", hw, saveDefault, fyne.NewMenuItemSeparator(), reset, pause, turbo, sound)
	sound.Action = func() {
		on := !e.sess.Settings().Sound
		e.sess.SetSound(on)
		sound.Checked = on
		machineMenu.Refresh()
		if on && e.audio == nil {
			dialog.ShowInformation("Som", "Nenhuma saída de som disponível.", w)
		}
	}

	var zooms []*fyne.MenuItem
	for z := 1; z <= 4; z++ {
		zooms = append(zooms, fyne.NewMenuItem(fmt.Sprintf("Zoom %d×", z), func() { setZoom(z) }))
	}
	borderNames := []string{"Sem borda", "Borda pequena", "Borda normal", "Borda grande", "Borda total"}
	var borders []*fyne.MenuItem
	for b, name := range borderNames {
		borders = append(borders, fyne.NewMenuItem(name, func() {
			e.mu.Lock()
			e.border = b
			e.mu.Unlock()
			setZoom(zoom())
		}))
	}
	view := fyne.NewMenu("Exibir", append(append(zooms, fyne.NewMenuItemSeparator()), borders...)...)

	console := fyne.NewMenuItem("Console de comandos", con.show)
	console.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF2}
	linkOn := fyne.NewMenuItem("Ligar a porta de comunicação (Link)", func() {
		e.runCommand(w, "Porta de comunicação", "link start")
	})
	linkOff := fyne.NewMenuItem("Desligar a porta de comunicação", func() {
		e.runCommand(w, "Porta de comunicação", "link stop")
	})
	linkStatus := fyne.NewMenuItem("Estado da porta de comunicação", func() {
		e.runCommand(w, "Porta de comunicação", "link status")
	})
	tools := fyne.NewMenu("Ferramentas", console, fyne.NewMenuItemSeparator(), linkOn, linkOff, linkStatus)

	keys := fyne.NewMenuItem("Teclado…", func() { dialog.ShowInformation("Teclado do ZX81", KeyHelp, w) })
	about := fyne.NewMenuItem("Sobre…", func() {
		dialog.ShowInformation("Sobre o Go81", "Go81 "+version.Full()+
			"\n\nEmulador de ZX81 em Go, compatível com o EightyOne.\nLicença GPL v3 ou posterior.", w)
	})

	return fyne.NewMainMenu(
		fyne.NewMenu("Arquivo", open, shot, fyne.NewMenuItemSeparator(), quit),
		editMenu,
		machineMenu,
		view,
		tools,
		fyne.NewMenu("Ajuda", keys, about),
	)
}
