package pkg

import "testing"

func TestNormalizeName(t *testing.T) {
	got := NormalizeName("  dib  ")
	if got != "dib" {
		t.Fatalf("NormalizeName() = %q, want %q", got, "dib")
	}
}
