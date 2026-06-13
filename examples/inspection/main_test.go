package main

import (
	"bytes"
	"testing"
)

func TestRunResolvesWithoutExecutingAHandler(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := run(&out, &errOut, []string{
		"auditctl",
		"events",
		"list",
		"--tenant", "acme",
		"--limit", "25",
		"severity=error",
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, errOut.String())
	}
	want := "path=auditctl events list\n" +
		"remaining=severity=error\n" +
		"tenant=acme limit=25\n"
	if got := out.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}
