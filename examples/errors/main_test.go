package main

import (
	"bytes"
	"testing"
)

const expectedUsageErrorExitCode = 2

func TestRunReportsParseErrorsAsStructuredDiagnostics(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := run(&out, &errOut, []string{"workerctl", "run", "--workers", "many"})

	if code != expectedUsageErrorExitCode {
		t.Fatalf("exit code = %d, want %d", code, expectedUsageErrorExitCode)
	}
	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", out.String())
	}
	want := "error=flags: flag value conversion failed at \"--workers\" for \"workers\"\n" +
		"parse_error token=\"--workers\" name=\"workers\" category=\"flag value conversion failed\"\n"
	if got := errOut.String(); got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestRunReportsUnknownCommandsAsStructuredDiagnostics(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := run(&out, &errOut, []string{"workerctl", "status"})

	if code != expectedUsageErrorExitCode {
		t.Fatalf("exit code = %d, want %d", code, expectedUsageErrorExitCode)
	}
	want := "error=command: unknown command \"status\"\n" +
		"unknown_command token=\"status\" parent=\"workerctl\"\n"
	if got := errOut.String(); got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}
