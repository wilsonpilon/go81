// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Ported from EightyOne tzx/tzxevent.cpp, tzx/tzxEventGeneral.cpp and
// tzx/TZXFILE.cpp — Copyright (C) 2003-2025 Michael D Wynne (GPL v2 or later).

// Package tape plays ZX81 tapes the way EightyOne's TZX player does,
// including its "flash load" mode used with the ROM LOAD patch.
package tape

// TZX block IDs used here.
const (
	BlockGeneral = 0x19
	BlockPause   = 0x20
	BlockText    = 0x30
)

// ByteEmpty means no byte is waiting for the flash loader (TZX_BYTE_EMPTY).
const ByteEmpty = -1

// zx81SymDefD is the pulse table EightyOne uses for ZX81 data in general
// blocks (tzxadd.cpp AddGeneralBlock): symbol 0 = 4 pulses, 1 = 9 pulses.
var zx81SymDefD = []uint16{
	3, 530, 520, 530, 520, 530, 520, 530, 4689,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	3, 530, 520, 530, 520, 530, 520, 530, 520, 530,
	520, 530, 520, 530, 520, 530, 520, 530, 4689,
}

// Block is one tape block.
type Block struct {
	ID    byte
	Pause int // milliseconds
	Data  []byte
	Text  string
	Raw   []byte // a block the player does not use, as in the file (after the ID)

	// General block header.
	TOTP, NPP, ASP int
	TOTD, NPD, ASD int
	DataLen        int
	SymDefP        []uint16
	PRLE           []uint16
	SymDefD        []uint16
}

// Player is EightyOne's TTZXFile (the parts used by the ZX81 family).
type Player struct {
	Blocks     []Block
	CurBlock   int
	startBlock int

	Playing         bool
	Stopping        bool
	blockInProgress bool
	FlashLoad       bool
	AutoStart       bool
	FlashLoadable   int
	EarState        int
	Pause           int
	curBlockLen     int
	curBlockProg    int

	// EventCounter is TZXEventCounter and Byte is TZXByte; the ROM patch
	// reads and clears them.
	EventCounter int
	Byte         int

	// ClockSpeed is the machine clock in Hz (3250000 for the ZX81).
	ClockSpeed int

	pauseP int // static p in EventDoPause

	// EventGeneral statics.
	gData       int // index into block data (C: pointer, may be -1)
	gPRIndex    int
	gPulse      int // index into SymDefP or SymDefD
	gPulseInD   bool
	gSymSize    byte
	gSymMask    byte
	gSymShift   byte
	gBitCount   int8 // unsigned char in C, compared with <= 0 after subtraction
	gPulseCount byte
	gUsedBits   int
	gBit        int
	gByte       int
	gPRCount    int
	gTOTP       int
	gPhase      int
	gDataLen    int
}

// NewPlayer returns an empty player for a machine running at clockSpeed Hz.
func NewPlayer(clockSpeed int) *Player {
	return &Player{ClockSpeed: clockSpeed, Byte: ByteEmpty, FlashLoad: true, AutoStart: true}
}

// block returns block i, or an empty block past the end (EightyOne reads a
// zeroed entry of its fixed-size array there).
func (p *Player) block(i int) *Block {
	if i >= 0 && i < len(p.Blocks) {
		return &p.Blocks[i]
	}
	return &Block{}
}

// scale is TZXSCALE(x): pulse lengths are given at 3.5 MHz (float maths).
func (p *Player) scale(x uint16) float32 {
	return float32(x) / (float32(3500000) / float32(p.ClockSpeed))
}

func (p *Player) addCounter(x uint16) {
	p.EventCounter = int(float32(p.EventCounter) + p.scale(x))
}

// NewTape clears the tape and adds EightyOne's text block (NewTZX).
func (p *Player) NewTape() {
	p.Blocks = p.Blocks[:0]
	p.Blocks = append(p.Blocks, Block{ID: BlockText, Text: "Created using the EightyOne emulator"})
	p.CurBlock = 1
	p.AutoStart = true
}

// AddGeneral appends a ZX81 data block (AddGeneralBlock) and returns its index.
func (p *Player) AddGeneral(data []byte) int {
	sym := make([]uint16, len(zx81SymDefD))
	copy(sym, zx81SymDefD)
	p.Blocks = append(p.Blocks, Block{
		ID: BlockGeneral, Pause: p.Pause,
		TOTD: len(data) * 8, NPD: 19, ASD: 2, DataLen: len(data),
		SymDefD: sym, Data: append([]byte(nil), data...),
	})
	return len(p.Blocks) - 1
}

// Record puts a block the ZX81 saved on the tape at the current position,
// with a pause of 5 s, and moves past it (TZXMan.cpp RecStopCheck).
func (p *Player) Record(data []byte) {
	if len(data) < 2 {
		return
	}
	i := p.AddGeneral(data)
	b := p.Blocks[i]
	b.Pause = 5000
	at := max(0, min(p.CurBlock, i))
	copy(p.Blocks[at+1:], p.Blocks[at:i])
	p.Blocks[at] = b
	p.CurBlock = at + 1
}

// Start is TTZXFile::Start.
func (p *Player) Start() bool {
	if p.Playing {
		return false
	}
	p.startBlock = p.CurBlock
	if p.CurBlock == len(p.Blocks) {
		p.CurBlock = 0
	}
	p.Playing = true
	p.blockInProgress = false
	p.FlashLoadable = 0
	p.EventCounter = p.ClockSpeed / 4
	p.Byte = ByteEmpty
	p.EarState = 0
	p.Stopping = false
	p.curBlockLen = 0
	p.curBlockProg = 0
	p.Pause = 0
	return true
}

// Stop is TTZXFile::Stop.
func (p *Player) Stop(err bool) {
	if !p.Playing || p.Stopping {
		return
	}
	p.Stopping = true
	p.blockInProgress = false
	p.EventCounter = 0
	p.FlashLoadable = 0
	p.Pause = 10
	if !err && p.curBlockLen != 0 && p.curBlockProg > p.curBlockLen-10 {
		p.nextBlock()
	}
}

func (p *Player) stopNow() {
	p.Playing = false
	p.FlashLoadable = 0
	p.EarState = 0
	p.Stopping = false
	p.EventCounter = 0
}

// ClockTick advances the tape by tstates; it reports whether a new block
// was reached.
func (p *Player) ClockTick(tstates int) bool {
	newBlock := false
	if !p.Playing {
		return false
	}
	p.EventCounter -= tstates
	for p.Playing && p.EventCounter <= 0 {
		if p.stateChange() {
			newBlock = true
		}
	}
	return newBlock
}

// GetEarState returns the tape signal.
func (p *Player) GetEarState() bool {
	if !p.Playing {
		return false
	}
	return p.EarState != 0
}

func (p *Player) stateChange() bool {
	if p.Pause != 0 {
		return p.doPause()
	}
	if p.Stopping {
		p.stopNow()
	}
	switch p.block(p.CurBlock).ID {
	case BlockGeneral:
		return p.eventGeneral()
	case BlockPause:
		b := p.block(p.CurBlock)
		p.Pause = b.Pause
		if p.Pause == 0 {
			p.Stop(true)
		}
		p.nextBlock()
		return true
	default:
		p.nextBlock()
		return true
	}
}

func (p *Player) doPause() bool {
	const initial = 1
	left := p.Pause - initial*3
	if left < 0 {
		left = 0
	}
	if p.pauseP < 2 {
		p.EarState = boolInt(p.EarState == 0)
		p.EventCounter = 3500 * initial * (p.pauseP + 1)
		p.pauseP++
	} else {
		p.EarState = 0
		p.EventCounter = 3500 * left
		p.Pause = 0
		p.pauseP = 0
	}
	return false
}

func (p *Player) nextBlock() {
	p.blockInProgress = false
	p.CurBlock++
	if p.CurBlock == len(p.Blocks) {
		p.Stop(true)
	}
	if p.CurBlock == p.startBlock {
		p.Stop(true)
	}
	p.curBlockLen = 0
	p.curBlockProg = 0
}

// u16 reads a table entry; past the end it returns 0 (C would read beyond).
func u16(t []uint16, i int) uint16 {
	if i >= 0 && i < len(t) {
		return t[i]
	}
	return 0
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// setEar is the SetEarState macro (ZX81: the level is inverted afterwards).
func (p *Player) setEar(list []uint16) {
	switch u16(list, p.gPulse) {
	case 0:
		p.gPulse++
	case 1:
		p.EarState = boolInt(p.EarState == 0)
		p.gPulse++
	case 2:
		p.EarState = 0
		p.gPulse++
	case 3:
		p.EarState = 1
		p.gPulse++
	}
	p.EarState = boolInt(p.EarState == 0)
}

func (p *Player) dataAt(b *Block, i int) int {
	if i >= 0 && i < len(b.Data) {
		return int(b.Data[i])
	}
	return 0
}

// eventGeneral is TTZXFile::EventGeneral.
func (p *Player) eventGeneral() bool {
	b := p.block(p.CurBlock)
	if !p.blockInProgress {
		p.gPhase = 0
		p.gData = -1
		p.gDataLen = b.DataLen
		p.gTOTP = b.TOTP
		if p.FlashLoad {
			p.gData++
		}
		syms := b.ASD
		if syms != 0 {
			syms--
		}
		p.gSymSize = 0
		for syms != 0 {
			p.gSymSize++
			syms >>= 1
		}
		p.gSymMask = byte(1<<p.gSymSize) - 1
		p.gSymShift = 8 - p.gSymSize
		p.gSymMask <<= p.gSymShift
		p.gUsedBits = 8 - (p.gDataLen*8 - b.TOTD*int(p.gSymSize))
		p.blockInProgress = true
		if p.gTOTP == 0 {
			p.gPhase = 2
		}
		if p.FlashLoad {
			p.gPhase = 5
		}
		p.Byte = ByteEmpty
		p.FlashLoadable = 1
	}

	if p.gPhase == 5 {
		p.EarState = 1
		p.EventCounter = 4
		if p.Byte == ByteEmpty {
			p.gDataLen--
			if p.gDataLen == -1 {
				p.EarState = 0
				p.Byte = ByteEmpty
				p.Pause = b.Pause / 100
				p.curBlockProg = 0
				p.curBlockLen = 0
				p.nextBlock()
				return true
			}
			p.Byte = p.dataAt(b, p.gData)
			p.gData++
		}
	}

	if p.gPhase == 0 {
		p.gPRIndex = 0
		p.gPRCount = int(u16(b.PRLE, 1))
		p.gPulse = int(u16(b.PRLE, 0)) * b.NPP
		p.gPulseInD = false
		p.setEar(b.SymDefP)
		p.gPulseCount = 1
		p.gPhase = 1
	}

	if p.gPhase == 1 {
		if u16(b.SymDefP, p.gPulse) != 0 && int(p.gPulseCount) < b.NPP {
			p.addCounter(u16(b.SymDefP, p.gPulse))
			p.EarState = boolInt(p.EarState == 0)
			p.gPulse++
			p.gPulseCount++
			return false
		}
		p.gPRCount--
		if p.gPRCount == 0 {
			p.gPRIndex += 2
			p.gPRCount = int(u16(b.PRLE, p.gPRIndex+1))
			p.gTOTP--
			if p.gTOTP == 0 {
				p.gPhase = 2
				return false
			}
		}
		p.gPulse = int(u16(b.PRLE, p.gPRIndex)) * b.NPP
		p.setEar(b.SymDefP)
		p.gPulseCount = 1
		return false
	}

	if p.gPhase == 2 {
		if p.gDataLen == 0 {
			p.Byte = ByteEmpty
			p.Pause = b.Pause
			p.curBlockLen = 0
			p.curBlockProg = 0
			p.nextBlock()
			return true
		}
		p.curBlockLen = p.gDataLen
		p.curBlockProg = 0
		p.gByte = p.dataAt(b, p.gData)<<8 | p.dataAt(b, p.gData+1)
		p.gBitCount = 8
		p.gBit = (p.gByte & int(p.gSymMask)) >> p.gSymShift
		p.gPulse = p.gBit * b.NPD
		p.gPulseInD = true
		p.setEar(b.SymDefD)
		p.gPulseCount = 1
		p.gPhase = 3
	}

	if p.gPhase == 3 {
		if u16(b.SymDefD, p.gPulse) != 0 && int(p.gPulseCount) < b.NPD {
			p.addCounter(u16(b.SymDefD, p.gPulse))
			p.EarState = boolInt(p.EarState == 0)
			p.gPulse++
			p.gPulseCount++
			return false
		}
		p.gBitCount -= int8(p.gSymSize)
		p.gPulseCount = 1
		if p.gBitCount <= 0 {
			p.gData++
			p.gByte = p.dataAt(b, p.gData)<<8 | p.dataAt(b, p.gData+1)
			p.curBlockProg++
			if p.gDataLen == 2 {
				p.gBitCount += int8(p.gUsedBits)
			} else {
				p.gBitCount += 8
			}
			p.gDataLen--
			if p.gDataLen == 0 {
				p.Byte = ByteEmpty
				p.Pause = b.Pause
				p.curBlockLen = 0
				p.curBlockProg = 0
				p.nextBlock()
				return true
			}
		} else {
			p.gByte <<= p.gSymSize
		}
		p.gBit = (p.gByte & int(p.gSymMask)) >> p.gSymShift
		p.gPulse = p.gBit * b.NPD
		p.setEar(b.SymDefD)
		return false
	}
	return false
}
