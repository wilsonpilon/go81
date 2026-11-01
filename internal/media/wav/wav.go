// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

// Package wav writes 16-bit PCM WAVE files.
package wav

import (
	"bufio"
	"encoding/binary"
	"errors"
	"os"
)

const headerSize = 44

// Writer writes samples to a .wav file. The sizes in the header are kept
// up to date by Sync, so a file that is never closed (the program ends
// abruptly) is still valid up to the last Sync.
type Writer struct {
	f        *os.File
	w        *bufio.Writer
	rate     int
	channels int
	bytes    int64 // sample data written
}

// Create starts a file with the sample rate and the number of channels.
func Create(path string, rate, channels int) (*Writer, error) {
	if rate <= 0 || channels < 1 || channels > 2 {
		return nil, errors.New("wav: formato inválido")
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	wr := &Writer{f: f, w: bufio.NewWriterSize(f, 64<<10), rate: rate, channels: channels}
	if _, err := wr.w.Write(wr.header()); err != nil {
		f.Close()
		return nil, err
	}
	return wr, nil
}

// header is the RIFF header for the data written so far.
func (wr *Writer) header() []byte {
	h := make([]byte, headerSize)
	le := binary.LittleEndian
	copy(h[0:], "RIFF")
	le.PutUint32(h[4:], uint32(36+wr.bytes))
	copy(h[8:], "WAVEfmt ")
	le.PutUint32(h[16:], 16) // size of the fmt chunk
	le.PutUint16(h[20:], 1)  // PCM
	le.PutUint16(h[22:], uint16(wr.channels))
	le.PutUint32(h[24:], uint32(wr.rate))
	le.PutUint32(h[28:], uint32(wr.rate*wr.channels*2)) // bytes per second
	le.PutUint16(h[32:], uint16(wr.channels*2))         // bytes per frame
	le.PutUint16(h[34:], 16)                            // bits per sample
	copy(h[36:], "data")
	le.PutUint32(h[40:], uint32(wr.bytes))
	return h
}

// Write adds samples (interleaved when stereo).
func (wr *Writer) Write(samples []int16) error {
	var b [2]byte
	for _, s := range samples {
		binary.LittleEndian.PutUint16(b[:], uint16(s))
		if _, err := wr.w.Write(b[:]); err != nil {
			return err
		}
	}
	wr.bytes += int64(2 * len(samples))
	return nil
}

// Seconds is the length of the sound written so far.
func (wr *Writer) Seconds() float64 {
	return float64(wr.bytes) / float64(wr.rate*wr.channels*2)
}

// Sync writes the buffered samples and the header with the current sizes.
func (wr *Writer) Sync() error {
	if err := wr.w.Flush(); err != nil {
		return err
	}
	if _, err := wr.f.WriteAt(wr.header(), 0); err != nil {
		return err
	}
	return nil
}

// Close finishes the file.
func (wr *Writer) Close() error {
	err := wr.Sync()
	if cerr := wr.f.Close(); err == nil {
		err = cerr
	}
	return err
}
