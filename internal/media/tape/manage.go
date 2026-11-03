// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later
//
// The block editing of EightyOne's tape manager window (tzx/TZXMan.cpp:
// delete, move up and down, edit the name) — Copyright (C) 2003-2025
// Michael D Wynne (GPL v2 or later).

package tape

import "fmt"

// Editing a tape, as the tape manager window does: the blocks are
// numbered from 0 here (the commands and the window count from 1).

func (p *Player) checkBlock(i int) error {
	if i < 0 || i >= len(p.Blocks) {
		return fmt.Errorf("não há bloco %d (a fita tem %d)", i+1, len(p.Blocks))
	}
	return nil
}

// SelectBlock makes block i the current one: the next LOAD starts there.
func (p *Player) SelectBlock(i int) error {
	if err := p.checkBlock(i); err != nil {
		return err
	}
	p.Stop(false)
	p.CurBlock = i
	return nil
}

// DeleteBlock takes block i off the tape.
func (p *Player) DeleteBlock(i int) error {
	if err := p.checkBlock(i); err != nil {
		return err
	}
	p.Stop(false)
	p.Blocks = append(p.Blocks[:i], p.Blocks[i+1:]...)
	if p.CurBlock > i {
		p.CurBlock--
	}
	p.CurBlock = max(0, min(p.CurBlock, len(p.Blocks)))
	return nil
}

// MoveBlock moves block i to position j (the others make room).
func (p *Player) MoveBlock(i, j int) error {
	if err := p.checkBlock(i); err != nil {
		return err
	}
	if err := p.checkBlock(j); err != nil {
		return err
	}
	p.Stop(false)
	b := p.Blocks[i]
	p.Blocks = append(p.Blocks[:i], p.Blocks[i+1:]...)
	p.Blocks = append(p.Blocks[:j], append([]Block{b}, p.Blocks[j:]...)...)
	// The current block stays the same block.
	switch c := p.CurBlock; {
	case c == i:
		p.CurBlock = j
	case i < c && j >= c:
		p.CurBlock--
	case i > c && j <= c:
		p.CurBlock++
	}
	return nil
}

// RenameProgram changes the name of the ZX81 program in block i (the name
// SAVE wrote before the program, which LOAD "NOME" looks for).
func (p *Player) RenameProgram(i int, name string) error {
	if err := p.checkBlock(i); err != nil {
		return err
	}
	b := &p.Blocks[i]
	if b.ID != BlockGeneral {
		return fmt.Errorf("o bloco %d não é um programa", i+1)
	}
	_, prog, ok := SplitSaved(b.Data[:min(b.DataLen, len(b.Data))])
	if !ok {
		return fmt.Errorf("o bloco %d não é um programa do ZX81 com nome", i+1)
	}
	zx := ZX81Name(name)
	if len(zx) == 0 || len([]rune(name)) != len(zx) {
		return fmt.Errorf("o nome %q tem caracteres que o ZX81 não tem", name)
	}
	data := append(zx, prog...)
	b.Data, b.DataLen, b.TOTD = data, len(data), len(data)*8
	return nil
}
