package main

import (
	"bytes"
	"testing"
)

func TestRunGreetsByName(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := run(&out, &errOut, []string{"greeter", "hello", "--name", "Ada"})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, errOut.String())
	}
	if got, want := out.String(), "hello, Ada\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestRunSupportsBooleanFlags(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := run(&out, &errOut, []string{"greeter", "hello", "--name", "Ada", "--shout"})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, errOut.String())
	}
	if got, want := out.String(), "HELLO, ADA\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}
