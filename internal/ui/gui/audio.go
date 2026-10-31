// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !nogui

package gui

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"

	"github.com/wilsonpilon/go81/internal/emu/zx81"
)

// Audio buffering: the emulator produces one frame of samples every 1/50 s
// on its own timer, the sound card pulls them on its clock. The queue absorbs
// the difference; when it grows too long the oldest samples are dropped.
const (
	bytesPerSecond = zx81.SampleRate * zx81.AudioChannels * 2
	maxQueue       = bytesPerSecond * 15 / 100 // 150 ms
	keepQueue      = bytesPerSecond * 6 / 100  // 60 ms
)

type audioOut struct {
	mu     sync.Mutex
	queue  []byte
	player *oto.Player
}

// newAudio opens the sound card. It returns nil if there is no sound device.
func newAudio() *audioOut {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   zx81.SampleRate,
		ChannelCount: zx81.AudioChannels,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   40 * time.Millisecond,
	})
	if err != nil {
		return nil
	}
	<-ready
	a := &audioOut{}
	a.player = ctx.NewPlayer(a)
	a.player.Play()
	return a
}

// Read feeds the sound card; missing samples are silence.
func (a *audioOut) Read(p []byte) (int, error) {
	a.mu.Lock()
	n := copy(p, a.queue)
	a.queue = a.queue[n:]
	a.mu.Unlock()
	for i := n; i < len(p); i++ {
		p[i] = 0
	}
	return len(p), nil
}

// push queues one frame of samples.
func (a *audioOut) push(samples []int16) {
	if a == nil {
		return
	}
	b := make([]byte, 2*len(samples))
	for i, s := range samples {
		binary.LittleEndian.PutUint16(b[2*i:], uint16(s))
	}
	a.mu.Lock()
	a.queue = append(a.queue, b...)
	if len(a.queue) > maxQueue {
		drop := len(a.queue) - keepQueue
		drop -= drop % 4 // whole stereo samples
		a.queue = append([]byte(nil), a.queue[drop:]...)
	}
	a.mu.Unlock()
}
