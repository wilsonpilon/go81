// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne tzx/TZXMan.cpp (tape manager window logic) and
// wavtape/WavCInterface.cpp — Copyright (C) 2003-2025 Michael D Wynne
// (GPL v2 or later).

package tape

// Deck is EightyOne's tape manager without its window: the Play and Stop
// buttons, "auto start" and the hooks the ROM patches call.
type Deck struct {
	Player *Player

	// In and Out are emulator.TZXin / TZXout: the tape manager is the input
	// (instead of a WAV file) and the output for SAVE.
	In, Out bool

	// Wav is the WAV loader, the input when In is off (wavdeck.go).
	Wav *WavPlayer
	// Rec is the WAV loader as the output, when Out is off (wavrec.go).
	Rec *WavRecorder

	outChosen bool // SetOutput was called: putting a tape in keeps Out

	playDown, stopDown bool
	previous           bool
}

// NewDeck creates a deck with an empty tape. EightyOne's defaults are
// "flash load" and "auto start" on.
func NewDeck(clockSpeed int) *Deck {
	return &Deck{Player: NewPlayer(clockSpeed), stopDown: true, previous: true}
}

func (d *Deck) playClick() {
	p := d.Player
	if p.block(p.CurBlock).ID == BlockPause {
		p.CurBlock++
	}
	if !p.Start() {
		return
	}
	if d.playDown {
		return
	}
	d.playDown = true
	d.stopDown = false
}

func (d *Deck) stopClick() {
	if d.stopDown {
		return
	}
	d.Player.Stop(false)
	d.playDown = false
}

// Play presses Play (also used by the LOAD patch when auto start is on).
func (d *Deck) Play() { d.playClick() }

// StopTape presses Stop.
func (d *Deck) StopTape() { d.stopClick() }

// ClockTick is TTZX::ClockTick for the ZX81.
func (d *Deck) ClockTick(tstates int) {
	if d.Wav != nil && !d.In { // WavLoad->ClockTick
		d.Wav.w.tick(tstates)
	}
	p := d.Player
	if p.Playing != d.previous {
		d.previous = p.Playing
		d.playDown = d.previous
		d.stopDown = !d.previous
	}
	if !p.Playing {
		return
	}
	if p.ClockTick(tstates) {
		a := p.CurBlock + 1
		rowCount := len(p.Blocks) + 2
		if a >= rowCount {
			a = rowCount - 1
		}
		p.CurBlock = a - 1
	}
}

// WavStart is called by the ROM LOAD patch.
func (d *Deck) WavStart() {
	switch {
	case d.In:
		if d.Player.AutoStart {
			d.playClick()
		}
	case d.Wav != nil:
		d.Wav.Start()
	}
}

// WavStop is called by the ROM patch at the end of LOAD/SAVE.
func (d *Deck) WavStop() {
	if d.playDown && d.Player.AutoStart {
		d.stopClick()
	}
	if d.Wav != nil {
		d.Wav.Stop()
	}
	if d.Rec != nil {
		d.Rec.recording = false
	}
}

// EarState is the tape signal seen on the EAR input.
func (d *Deck) EarState() bool {
	if d.In {
		return d.Player.GetEarState()
	}
	if d.Wav != nil {
		return d.Wav.w.ear()
	}
	return false
}

// FlashSaveable reports whether SAVE goes straight to the tape manager.
func (d *Deck) FlashSaveable() bool { return d.Out }

// Load puts a whole tape in the deck (a .tzx or .t81: EightyOne's
// LoadFile, then UpdateTable(true)): the first block is the current one.
func (d *Deck) Load(blocks []Block) {
	p := d.Player
	p.Blocks = append(p.Blocks[:0], blocks...)
	p.CurBlock = 0
	d.Wav = nil
	d.In = true
	d.keepOutput()
}

// Insert puts a new tape in the deck with one program block, as opening a
// .p file in EightyOne's main window does.
func (d *Deck) Insert(block []byte) {
	p := d.Player
	p.NewTape()
	i := p.AddGeneral(block)
	p.Blocks[i].Pause = 3000
	p.CurBlock = 0 // TZX->UpdateTable(true)
	d.Wav = nil
	d.In = true
	d.keepOutput()
}
