package pkg

import "testing"

func TestName(t *testing.T) {
	if got := Name(); got != "resolved" {
		t.Fatalf("Name() = %q, want %q", got, "resolved")
	}
}
