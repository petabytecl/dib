package pkg_test

import (
	"testing"

	_ "example.com/external/also-notstdlib"
	_ "example.com/external/notstdlib"

	"example.com/dib-depgate-nonstdlib-fixture/pkg"
)

func TestNormalizeName(t *testing.T) {
	got := pkg.NormalizeName("  dib  ")
	if got != "dib" {
		t.Fatalf("NormalizeName() = %q, want %q", got, "dib")
	}
}
