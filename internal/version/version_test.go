package version

import "testing"

func TestNumber(t *testing.T) {
	if got := Number(); got != "1.1.1" {
		t.Fatalf("Number() = %q, want 1.1.1", got)
	}
}
