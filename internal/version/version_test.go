// Go81 — Copyright (C) 2026 Wilson "Barney" Pilon
// SPDX-License-Identifier: GPL-3.0-or-later

package version

import "testing"

func TestNumber(t *testing.T) {
	if got := Number(); got != "1.3.1" {
		t.Fatalf("Number() = %q, want 1.3.1", got)
	}
}
