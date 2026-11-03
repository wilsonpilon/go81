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
	"net/url"
	"os"
	"path/filepath"
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

	"github.com/wilsonpilon/go81/internal/basic"
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
	Zoom       float64
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
	pressed map[string]bool // PC keys held down
	ctrl    bool            // a Control key is held down

	// The on-screen keyboard: ZX81 keys held with the mouse, and SHIFT,
	// which stays down until the next key is released.
	screenKeys map[string]bool
	shiftLock  bool
	// onScreenKeys is called (not on the Fyne thread) when they change.
	onScreenKeys func()

	configPath string // the configuration file

	prefs           fyne.Preferences // where the recent files are kept
	recent          []string         // recent files, newest first
	onRecentChanged func()           // rebuilds the "Abrir recente" menu

	message         string    // shown in the status bar for a few seconds
	msgEnd          time.Time // when the message goes away
	turbo           bool
	syncTape        func() // the check marks of Opções → Fita
	border          int
	fps             float64
	audio           *audioOut
	inverse         bool                    // inverse video (Exibir menu)
	keymap          *kbOverlay              // the keyboard map over the screen (F4)
	editor          *editorView             // the BASIC editor in place of the screen (F7)
	fkeys           map[fyne.KeyName]func() // F keys of menu items (Fyne runs menu shortcuts only with Ctrl/Alt)
	onKeyMapChanged func()                  // marks the Mapa do teclado menu
	status          *widget.Label           // the status bar

	// onSaved is called when the program running in the ZX81 finishes a
	// SAVE (not the console's save command, which writes its own file).
	onSaved func(name string, prog []byte)
}

func newEmulator(sess *session.Session, border int) *emulator {
	e := &emulator{sess: sess, border: border, pressed: map[string]bool{}, screenKeys: map[string]bool{}}
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
	km := newKbOverlay()
	e.keymap = km
	ev := newEditorView(e, screen)
	e.editor = ev
	ev.w = w
	zoom := opts.Zoom
	setZoom := func(z float64) {
		zoom = z
		r := e.picture().Bounds()
		wd, ht := float32(float64(r.Dx())*z), float32(float64(r.Dy())*z)
		screen.SetMinSize(fyne.NewSize(wd, ht))
		ev.img.SetMinSize(fyne.NewSize(wd, ht))
		set := e.sess.Settings()
		km.setModel(set.Model, set.NTSC)
		km.resize(wd)
		if !w.FullScreen() {
			w.Resize(fyne.NewSize(wd, ht+30))
		}
	}
	status := widget.NewLabel("")
	e.status = status
	w.SetContent(container.NewBorder(nil, status, nil, nil, container.NewStack(screen, ev.img, km.box)))
	setZoom(zoom)

	e.configPath = opts.ConfigPath
	if e.configPath == "" {
		e.configPath = session.ConfigPath()
	}
	e.prefs = a.Preferences()
	ev.keys = e.prefs.Bool(prefEditorKeys) && ev.kb != nil
	e.recent = e.prefs.StringList(prefRecent)
	km.fade = e.prefs.FloatWithFallback(prefMapFade, 0.5)
	km.setMode(e.prefs.IntWithFallback(prefMapMode, mapOff))

	con := newConsoleWindow(a, sess, func() { a.Quit() })
	var kb *keyboardWindow
	// The F keys of the window work in it and in the on-screen keyboard.
	onKey := func(k *fyne.KeyEvent, down bool) {
		if k.Name == desktop.KeyShiftLeft || k.Name == desktop.KeyShiftRight {
			km.setShift(down) // the keyboard map can show while SHIFT is held
			ev.shift = down
		}
		if k.Name == desktop.KeyControlLeft || k.Name == desktop.KeyControlRight {
			km.setCtrl(down, e.cursorNow()) // the keys of the ZX81's cursor
			ev.ctrl = down
		}
		if down && ev.scrollKey(k.Name) { // Ctrl+arrows scroll the editor's listing
			return
		}
		switch k.Name {
		case fyne.KeyF4: // keyboard map over the screen
			if down {
				e.setKeyMapMode(km.mode + 1)
			}
		case fyne.KeyF1: // picture of the keyboard (Help menu), as in EightyOne
			if down {
				e.showKeyboardHelp(a)
			}
		case fyne.KeyF2: // console (also in the Tools menu)
			if down {
				con.show()
			}
		case fyne.KeyF3: // on-screen keyboard
			if down {
				kb.show()
			}
		case fyne.KeyF11: // full screen
			if down {
				w.SetFullScreen(!w.FullScreen())
			}
		case fyne.KeyF5, fyne.KeyF6: // pause, turbo (Máquina menu)
			if f := e.fkeys[k.Name]; down && f != nil {
				f()
			}
		case fyne.KeyF7: // BASIC editor in place of the screen
			if down {
				ev.toggle()
			}
		case fyne.KeyF8: // editor: direct typing or the ZX81 keys
			if down {
				ev.toggleKeys()
			}
		default:
			if ev.handlesKey(k.Name) {
				if down {
					ev.keyDown(k.Name)
				}
				return // typed into the editor (SetOnTypedRune/Key)
			}
			e.key(string(k.Name), down)
		}
	}
	kb = newKeyboardWindow(a, e, onKey)
	if dc, ok := w.Canvas().(desktop.Canvas); ok {
		dc.SetOnKeyDown(func(k *fyne.KeyEvent) { onKey(k, true) })
		dc.SetOnKeyUp(func(k *fyne.KeyEvent) { onKey(k, false) })
	}
	w.Canvas().SetOnTypedRune(ev.typedRune)
	w.Canvas().SetOnTypedKey(ev.typedKey)
	// A program dropped on the window is loaded (the first one, if many).
	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, u := range uris {
			if u.Scheme() == "file" && isProgram(u.Path()) {
				e.openFile(w, u.Path())
				return
			}
		}
		if len(uris) > 0 {
			e.flash("Não é um programa do ZX81: " + uris[0].Name())
		}
	})

	e.onSaved = func(name string, prog []byte) {
		fyne.Do(func() { askSave(w, name, prog) })
	}

	w.SetMainMenu(e.menu(a, w, con, kb, func() float64 { return zoom }, setZoom))
	// Release every key when the window loses focus, so none gets stuck.
	a.Lifecycle().SetOnExitedForeground(func() {
		e.releaseAll()
		km.setShift(false)
		km.setCtrl(false, 0)
	})
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
			if km.ctrl {
				km.setCursor(e.cursorNow()) // follow K → L, SHIFT+ENTER → F...
			}
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

	w.SetOnClosed(func() {
		e.sess.StopRecording() // finish the .wav
		close(stop)
		a.Quit()
	})
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

// prefRecent is the preference that keeps the recent files.
const prefRecent = "recentes"

// openFile loads a program from the disk (recent files, drag and drop).
func (e *emulator) openFile(w fyne.Window, path string) {
	if session.IsSnapshot(path) { // the whole machine (.z81)
		if err := e.sess.LoadSnapshot(path); err != nil {
			dialog.ShowError(err, w)
			return
		}
		e.rememberFile(path)
		e.flash("Snapshot carregado: " + filepath.Base(path))
		return
	}
	data, err := os.ReadFile(path)
	var machine string
	if err == nil {
		machine, err = e.sess.GameMachine(path) // the machine of this game, if it has one
	}
	if err == nil {
		err = e.sess.InsertProgram(filepath.Base(path), data, true)
	}
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	e.rememberFile(path)
	if machine != "" {
		e.flash("Carregando " + filepath.Base(path) + " · " + machine)
		return
	}
	e.flash("Carregando " + filepath.Base(path))
}

// rememberFile puts a file at the top of the recent files.
func (e *emulator) rememberFile(path string) {
	e.setRecent(addRecent(e.recent, path))
}

func (e *emulator) setRecent(list []string) {
	e.recent = list
	if e.prefs != nil {
		e.prefs.SetStringList(prefRecent, list)
	}
	if e.onRecentChanged != nil {
		e.onRecentChanged()
	}
}

// rememberView writes the zoom and the border in the configuration file, so
// the next window opens the same way.
func (e *emulator) rememberView(zoom float64) {
	e.mu.Lock()
	b := e.border
	e.mu.Unlock()
	if err := session.SaveWindow(e.configPath, zoom, b); err != nil {
		e.flash("Não deu para gravar o zoom e a borda: " + err.Error())
	}
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
	ext := ".p"
	if name == "" { // only the ZX80 saves without a name: a .o file
		ext = ".o"
	}
	d.SetFileName(file + ext)
	d.SetFilter(storage.NewExtensionFileFilter([]string{ext}))
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

// releaseAll releases every PC key (the on-screen keyboard keeps its own).
func (e *emulator) releaseAll() {
	e.mu.Lock()
	e.pressed = map[string]bool{}
	rows := e.rowsLocked()
	e.mu.Unlock()
	e.sess.SetKeys(rows)
}

// rowsLocked is the ZX81 keyboard matrix for the keys held down on the PC
// and on the screen. Caller holds e.mu.
func (e *emulator) rowsLocked() [8]byte {
	var chords []string
	for k := range e.pressed {
		chords = append(chords, keyMap[k])
	}
	for k := range e.screenKeys {
		chords = append(chords, k)
	}
	if e.shiftLock {
		chords = append(chords, "SHIFT")
	}
	var rows [8]byte
	for _, c := range chords {
		r, err := zx81.ParseChord(c)
		if err != nil {
			continue
		}
		for i := range rows {
			rows[i] |= r[i]
		}
	}
	return rows
}

// screenKey presses or releases a key of the on-screen keyboard. A click on
// SHIFT locks it (or unlocks it); releasing any other key unlocks it.
func (e *emulator) screenKey(name string, down bool) {
	e.mu.Lock()
	switch {
	case name == "SHIFT":
		if down {
			e.shiftLock = !e.shiftLock
		}
	case down:
		e.screenKeys[name] = true
	default:
		delete(e.screenKeys, name)
		e.shiftLock = false
	}
	rows := e.rowsLocked()
	notify := e.onScreenKeys
	e.mu.Unlock()
	e.sess.SetKeys(rows)
	if notify != nil {
		notify()
	}
}

// screenKeyDown reports whether a key of the on-screen keyboard is held.
func (e *emulator) screenKeyDown(name string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if name == "SHIFT" {
		return e.shiftLock
	}
	return e.screenKeys[name]
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
	rows := e.rowsLocked()
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
		turbo, border, inverse := e.turbo, e.border, e.inverse
		e.mu.Unlock()
		n, speed := e.sess.SpeedFrames(), e.sess.Speed()
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
			if ran && n == 1 && !turbo && speed == 100 && set.Sound {
				audio = append([]int16(nil), m.Audio()...)
			}
			img = m.Screen(border)
			if inverse {
				invert(img)
			}
			playing = m.Tape != nil && m.Tape.Player.Playing
		})
		state := ""
		switch {
		case e.sess.Paused():
			state = " · PAUSADO"
		case turbo:
			state = " · TURBO"
		case speed != 100:
			state = fmt.Sprintf(" · %d%%", speed)
		}
		if playing {
			state += " · fita tocando"
		}
		if rec, secs := e.sess.Recording(); rec != "" {
			state += fmt.Sprintf(" · ● gravando som %.0f s", secs)
		}
		e.mu.Lock()
		if d := time.Since(last); d >= time.Second {
			e.fps = float64(frames) / d.Seconds()
			frames, last = 0, time.Now()
			if sync := e.syncTape; sync != nil {
				fyne.Do(sync) // Opções → Fita follows the tape put in
			}
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

func (e *emulator) menu(a fyne.App, w fyne.Window, con *consoleWindow, kb *keyboardWindow, zoom func() float64, setZoom func(float64)) *fyne.MainMenu {
	open := fyne.NewMenuItem("Abrir programa…", func() {
		d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			defer r.Close()
			if r.URI().Scheme() == "file" { // from the disk: also its game configuration
				e.openFile(w, r.URI().Path())
				return
			}
			data, err := io.ReadAll(r)
			if err == nil {
				err = e.sess.InsertProgram(r.URI().Name(), data, true)
			}
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
		}, w)
		var exts []string
		for _, x := range programExts {
			exts = append(exts, x, strings.ToUpper(x))
		}
		d.SetFilter(storage.NewExtensionFileFilter(exts))
		d.Show()
	})
	open.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierShortcutDefault}
	w.Canvas().AddShortcut(open.Shortcut, func(fyne.Shortcut) { open.Action() })
	recent := fyne.NewMenuItem("Abrir recente", nil)
	recent.ChildMenu = fyne.NewMenu("")
	var mainMenu *fyne.MainMenu
	e.onRecentChanged = func() {
		var items []*fyne.MenuItem
		for _, p := range e.recent {
			items = append(items, fyne.NewMenuItem(p, func() { e.openFile(w, p) }))
		}
		if len(items) == 0 {
			none := fyne.NewMenuItem("(nenhum)", nil)
			none.Disabled = true
			items = append(items, none)
		} else {
			items = append(items, fyne.NewMenuItemSeparator(),
				fyne.NewMenuItem("Limpar a lista", func() { e.setRecent(nil) }))
		}
		recent.ChildMenu.Items = items
		if mainMenu != nil {
			mainMenu.Refresh()
		}
	}
	e.onRecentChanged()
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
		if e.editor.on { // into the BASIC editor
			e.editor.ed.Paste(a.Clipboard().Content())
			e.editor.draw()
			return
		}
		msg, err := e.sess.Paste(a.Clipboard().Content())
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		e.flash(msg)
	})
	paste.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyV, Modifier: fyne.KeyModifierShortcutDefault}
	copyText := fyne.NewMenuItem("Copiar tela como texto", func() {
		if e.editor.on { // the program of the editor, as text
			a.Clipboard().SetContent(basic.List(e.editor.ed.Lines))
			e.flash("Programa copiado como texto")
			return
		}
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

	// The menus follow EightyOne (File, View, Control, Options, Tools, Help);
	// what Go81 does not have yet says "A implementar". The shortcuts are the
	// ones of Go81 (F2 console, F3 keyboard, F5 pause, F6 turbo, F11 full
	// screen), with F1 for the keyboard picture as in EightyOne.
	saveProg := fyne.NewMenuItem("Salvar programa…", func() {
		d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			wc.Close()
			if wc.URI().Scheme() != "file" {
				dialog.ShowError(fmt.Errorf("só dá para gravar num arquivo do disco"), w)
				return
			}
			e.runCommand(w, "Salvar programa", `save "`+wc.URI().Path()+`"`)
		}, w)
		ext := e.sess.ProgramExt()
		d.SetFileName("programa" + ext)
		d.SetFilter(storage.NewExtensionFileFilter([]string{ext}))
		d.Show()
	})
	gameItem := fyne.NewMenuItem("Gravar a máquina para este jogo", func() {
		e.runCommand(w, "Máquina deste jogo", "jogo salvar")
	})
	fileMenu := fyne.NewMenu("Arquivo",
		open, recent, saveProg, gameItem,
		fyne.NewMenuItemSeparator(),
		e.tapeOpenItem(w), e.tapeSaveItem(w), e.tapeNewItem(w), e.tapeStatusItem(w),
		fyne.NewMenuItemSeparator(),
		todoItem(w, "Drives…"),
		fyne.NewMenuItemSeparator(),
		e.snapshotLoadItem(w), e.snapshotSaveItem(w),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Carregar bloco de memória…", func() { e.loadMemoryBlock(w) }),
		fyne.NewMenuItem("Salvar bloco de memória…", func() { e.saveMemoryBlock(w) }),
		shot,
		fyne.NewMenuItemSeparator(),
		quit)

	// Exibir: the zoom and the border chosen here are remembered in the
	// configuration file ([janela]).
	var zooms, borders []*fyne.MenuItem
	refresh := func() {
		if mainMenu != nil {
			mainMenu.Refresh()
		}
	}
	markView := func() {
		e.mu.Lock()
		b := e.border
		e.mu.Unlock()
		for i, it := range zooms {
			it.Checked = float64(i+1) == zoom()
		}
		for i, it := range borders {
			it.Checked = i == b
		}
		refresh()
	}
	for z := 1; z <= 4; z++ {
		zooms = append(zooms, fyne.NewMenuItem(fmt.Sprintf("%d×", z), func() {
			setZoom(float64(z))
			markView()
			e.rememberView(float64(z))
		}))
	}
	// Any zoom from 1 to 8, fractions too (EightyOne: View → Zoom → User defined).
	userZoom := fyne.NewMenuItem("Definido pelo usuário…", func() {
		askText(w, "Zoom definido pelo usuário", []string{fmt.Sprintf("Zoom (%g a %g, ex.: 2,5)", session.MinZoom, session.MaxZoom)},
			[]string{session.FormatZoom(zoom())}, func(v []string) {
				z, err := session.ParseZoom(v[0])
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				setZoom(z)
				markView()
				e.rememberView(z)
			})
	})
	full := fyne.NewMenuItem("Tela cheia", func() { w.SetFullScreen(!w.FullScreen()) })
	full.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF11}
	zoomItem := fyne.NewMenuItem("Zoom", nil)
	zoomItem.ChildMenu = fyne.NewMenu("", append(zooms, fyne.NewMenuItemSeparator(),
		userZoom, full)...)
	borderNames := []string{"Sem borda (256×192)", "Pequena (264×200)", "Normal (320×240)", "Grande (400×300)", "Quadro inteiro (414×313)"}
	for b, name := range borderNames {
		borders = append(borders, fyne.NewMenuItem(name, func() {
			e.mu.Lock()
			e.border = b
			e.mu.Unlock()
			setZoom(zoom())
			markView()
			e.rememberView(zoom())
		}))
	}
	borderItem := fyne.NewMenuItem("Tamanho da borda", nil)
	borderItem.ChildMenu = fyne.NewMenu("", borders...)
	inverse := fyne.NewMenuItem("Vídeo inverso", nil)
	inverse.Action = func() {
		e.mu.Lock()
		e.inverse = !e.inverse
		inverse.Checked = e.inverse
		e.mu.Unlock()
		refresh()
	}
	annotations := todoMenu(w, "Anotações",
		"Selecionar todas as colorações", "Remover todas as colorações", "-",
		"Pulso de sincronismo vertical", "Rotina de vídeo da ROM", "Pulso de sincronismo horizontal", "Back porch",
		"Resposta à NMI (interrupção não mascarável)", "Estados de espera da resposta à NMI", "Rotina de serviço da NMI",
		"Recursão da rotina de serviço da NMI", "Instrução atravessando a NMI", "Estados de espera da instrução atravessando a NMI",
		"Resposta à INT (interrupção mascarável)", "Rotina de serviço da INT", "Z80 parado (HALT)",
		"Início das instruções do programa do usuário")
	statusBar := fyne.NewMenuItem("Barra de estado", nil)
	statusBar.Checked = true
	statusBar.Action = func() {
		statusBar.Checked = !statusBar.Checked
		if statusBar.Checked {
			e.status.Show()
		} else {
			e.status.Hide()
		}
		refresh()
	}
	keyboard := fyne.NewMenuItem("Teclado do ZX81 na tela", kb.show)
	others := func() []fyne.Window {
		return otherWindows(a, w, func(x fyne.Window) bool { return x == kb.w && !kb.visible })
	}
	gather := fyne.NewMenuItem("Reunir janelas", func() { e.flash(windowsMessage("Janelas reunidas", gatherWindows(others()))) })
	front := fyne.NewMenuItem("Trazer janelas para a frente", func() { e.flash(windowsMessage("Janelas trazidas para a frente", bringToFront(others()))) })
	keyboard.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF3}
	mapItem := e.keyMapMenu(refresh)
	editorItem := fyne.NewMenuItem("Editor de BASIC (F7 alterna com o emulador)", func() { e.editor.toggle() })
	editorItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF7}
	editorMenu := fyne.NewMenuItem("Editor de BASIC", nil)
	editorMenu.ChildMenu = fyne.NewMenu("", append([]*fyne.MenuItem{editorItem, fyne.NewMenuItemSeparator()}, e.editor.fileItems()...)...)
	view := fyne.NewMenu("Exibir", zoomItem, borderItem, inverse, annotations, statusBar,
		fyne.NewMenuItemSeparator(),
		gather, front, e.splashItem(w),
		fyne.NewMenuItemSeparator(),
		keyboard, mapItem, fyne.NewMenuItemSeparator(), editorMenu)
	markView()

	// Máquina (EightyOne's Control menu).
	softReset := fyne.NewMenuItem("Reiniciar (linha RESET)", func() {
		e.sess.Do(func(m *zx81.Machine) { m.CPU.Reset() })
	})
	hardReset := fyne.NewMenuItem("Desligar e ligar", e.sess.Reset)
	pause := fyne.NewMenuItem("Pausar / continuar", func() { e.sess.SetPaused(!e.sess.Paused()) })
	pause.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF5}
	nmi := fyne.NewMenuItem("NMI", func() {
		e.sess.Do(func(m *zx81.Machine) { m.CPU.NMI() })
	})
	turbo := fyne.NewMenuItem("Turbo (10×)", nil)
	turbo.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF6}
	e.fkeys = map[fyne.KeyName]func(){fyne.KeyF5: func() { pause.Action() }, fyne.KeyF6: func() { turbo.Action() }}
	turbo.Action = func() {
		e.mu.Lock()
		e.turbo = !e.turbo
		turbo.Checked = e.turbo
		e.mu.Unlock()
		refresh()
	}
	machineMenu := fyne.NewMenu("Máquina",
		softReset, hardReset, pause, nmi, turbo,
		fyne.NewMenuItemSeparator(),
		todoMenu(w, "Joystick",
			"Escolher o controle do joystick 1", "Escolher o controle do joystick 2",
			"Conectar o joystick 1", "Conectar o joystick 2",
			"Tiro automático no joystick 1", "Tiro automático no joystick 2"),
		fyne.NewMenuItemSeparator(),
		todoItem(w, "Reiniciar o modo HRG da Memotech"), todoItem(w, "Reiniciar a alta resolução da Quicksilva"),
		todoItem(w, "Reiniciar a saída de voz"),
		fyne.NewMenuItemSeparator(),
		todoItem(w, "Placa de caracteres Quicksilva"), todoItem(w, "Cores Chroma"), todoItem(w, "Cores Spectra"),
		fyne.NewMenuItemSeparator(),
		todoItem(w, "Ligar o Memocalc da Memotech"), todoItem(w, "Ligar o montador Z80 da Memotech"),
		fyne.NewMenuItemSeparator(),
		todoItem(w, "Memória do divIDE (jumper E fechado)"), todoItem(w, "Memória da ZXCF (jumper UPLOAD fechado)"),
		todoItem(w, "ROM do Simple IDE"),
		fyne.NewMenuItemSeparator(),
		todoItem(w, "Proteger a RAM de 8K–16K contra escrita"))

	// Opções.
	saveDefault := fyne.NewMenuItem("Salvar esta máquina como padrão", func() {
		if err := session.SaveDefaultMachine(e.configPath, e.sess.Settings()); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Configuração", "Máquina padrão gravada em\n"+e.configPath+
			"\n\n"+e.sess.Description(), w)
	})
	restore := fyne.NewMenuItem("Voltar à máquina de fábrica", func() {
		dialog.ShowConfirm("Máquina de fábrica", "Reiniciar como ZX81 de 16K, ROM edição 3, PAL?", func(ok bool) {
			if !ok {
				return
			}
			if err := e.sess.Configure(session.DefaultSettings()); err != nil {
				dialog.ShowError(err, w)
				return
			}
			setZoom(zoom())
		}, w)
	})
	configItems := []*fyne.MenuItem{saveDefault, restore, e.deleteConfigItem(w, func() { setZoom(zoom()) })}
	if cfg, err := session.LoadConfig(e.configPath); err == nil && len(cfg.Profiles) > 0 {
		configItems = append(configItems, fyne.NewMenuItemSeparator())
		for _, name := range cfg.ProfileNames() {
			configItems = append(configItems, fyne.NewMenuItem("Perfil "+name, func() {
				e.runCommand(w, "Perfil "+name, "machine perfil="+name)
			}))
		}
	}
	configItem := fyne.NewMenuItem("Configuração", nil)
	configItem.ChildMenu = fyne.NewMenu("", configItems...)
	tapeItem, syncTape := e.tapeOptions(w, refresh)
	e.mu.Lock()
	e.syncTape = func() {
		if syncTape() {
			refresh()
		}
	}
	e.mu.Unlock()
	sound := fyne.NewMenuItem("Som pelo vídeo (Video Sound)", nil)
	sound.Checked = e.sess.Settings().Sound
	sound.Action = func() {
		on := !e.sess.Settings().Sound
		e.sess.SetSound(on)
		sound.Checked = on
		refresh()
		if on && e.audio == nil {
			dialog.ShowInformation("Som", "Nenhuma saída de som disponível.", w)
		}
	}
	hw := fyne.NewMenuItem("Hardware…", func() { e.hardwareDialog(w, setZoom, zoom) })
	optionsMenu := fyne.NewMenu("Opções",
		configItem,
		fyne.NewMenuItemSeparator(),
		tapeItem,
		todoMenu(w, "Portas", "Impressora (porta Centronics)", "Modem Timex TS2050"),
		sound,
		todoItem(w, "Tela…"), hw, todoItem(w, "Teclado…"), todoItem(w, "Som…"),
		e.speedItem(w), todoItem(w, "Emulação de TV…"))

	// Ferramentas.
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
	tools := fyne.NewMenu("Ferramentas",
		console,
		fyne.NewMenuItemSeparator(),
		todoItem(w, "Memória ao vivo…"), todoItem(w, "Depurador…"), todoItem(w, "Saída da ZX Printer…"),
		todoItem(w, "Saída de som…"), e.tapeManagerItem(a, w), e.wavLoaderItem(w),
		e.recordMenu(w),
		e.basicMenu(a, w),
		fyne.NewMenuItemSeparator(),
		linkOn, linkOff, linkStatus)

	// Ajuda.
	openURL := func(u string) func() {
		return func() {
			p, err := url.Parse(u)
			if err == nil {
				err = a.OpenURL(p)
			}
			if err != nil {
				dialog.ShowError(err, w)
			}
		}
	}
	const repo = "https://github.com/wilsonpilon/go81"
	keys := fyne.NewMenuItem("Teclado", func() { e.showKeyboardHelp(a) })
	keys.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF1}
	about := fyne.NewMenuItem("Sobre…", func() {
		dialog.ShowInformation("Sobre o Go81", "Go81 "+version.Full()+
			"\n\nEmulador de ZX81 em Go, compatível com o EightyOne.\nLicença GPL v3 ou posterior."+
			"\n\nImagens do teclado: EightyOne (GPL v2 ou posterior).", w)
	})
	helpMenu := fyne.NewMenu("Ajuda",
		keys,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Manual embutido (man)…", func() { showMan(a, "") }),
		fyne.NewMenuItem("Manual do Go81 (site)", openURL(repo+"/blob/main/MANUAL.md")),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exemplos", openURL(repo+"/tree/main/examples")),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Histórico de versões", openURL(repo+"/blob/main/RELEASE.md")),
		fyne.NewMenuItem("Procurar atualizações", openURL(repo+"/releases")),
		fyne.NewMenuItemSeparator(),
		about)

	mainMenu = fyne.NewMainMenu(fileMenu, editMenu, view, machineMenu, optionsMenu, tools, helpMenu)
	return mainMenu
}
