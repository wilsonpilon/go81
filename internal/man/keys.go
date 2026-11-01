// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package man

import (
	"fmt"

	"github.com/wilsonpilon/go81/assets"
)

// Where each keyword is on the keyboard, read from the keyboard tables of
// the ROM itself (so it cannot be wrong): on the K cursor a letter key
// gives the keyword letter+C0h; the SHIFT table gives the signs and some
// keywords; on the ZX81 the FUNCTION table (SHIFT+ENTER, cursor F) gives
// the functions.

// matrixKeys is the order of the keys in the ROM tables.
var matrixKeys = []string{"Z", "X", "C", "V", "A", "S", "D", "F", "G", "Q", "W", "E", "R", "T",
	"1", "2", "3", "4", "5", "0", "9", "8", "7", "6", "P", "O", "I", "U", "Y", "ENTER",
	"L", "K", "J", "H", "SPACE", ".", "M", "N", "B"}

// keyTable returns, for each code, how to type it.
func keyTable(romName string, shiftAt, functAt int) map[byte]string {
	rom, err := assets.ROM(romName)
	if err != nil {
		return nil
	}
	keys := map[byte]string{}
	for _, k := range matrixKeys {
		if len(k) == 1 && k[0] >= 'A' && k[0] <= 'Z' {
			code := byte(0x26+k[0]-'A') + 0xC0
			keys[code] = fmt.Sprintf("tecla %s com o cursor K (início do comando)", k)
		}
	}
	for i, k := range matrixKeys {
		if c := rom[shiftAt+i]; c >= 0xC0 || c == 0x40 || c == 0x41 || c == 0x42 {
			if _, ok := keys[c]; !ok {
				keys[c] = "SHIFT+" + k
			}
		}
	}
	if functAt > 0 {
		for i, k := range matrixKeys {
			c := rom[functAt+i]
			if c == 0x78 { // no function on this key
				continue
			}
			if _, ok := keys[c]; !ok {
				keys[c] = fmt.Sprintf("FUNCTION (SHIFT+ENTER, cursor F) e depois %s", k)
			}
		}
	}
	return keys
}

var zx81KeyMap, zx80KeyMap map[byte]string

func keyMaps() {
	zx81KeyMap = keyTable("zx81.edition3", 0x00A5, 0x00CC) // K-SHIFT, K-FUNCT
	zx80KeyMap = keyTable("zx80", 0x006C+39, 0)
}
