package main

import (
	"bytes"
	"testing"
)

func TestRunShowsConfigPrecedence(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	lookup := mapLookup(map[string]string{
		"DIB_REGION":  "eu-south",
		"DIB_WORKERS": "3",
	})

	code := run(&out, &errOut, []string{"deployctl", "deploy", "--workers", "4"}, lookup)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, errOut.String())
	}
	want := "route=deployctl deploy\n" +
		"region=eu-south workers=4 format=json\n" +
		"sources region=env workers=flag binding format=JSON\n"
	if got := out.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
