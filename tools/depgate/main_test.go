package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDepgateAllowsStdlibOnlyFixture(t *testing.T) {
	t.Skip("ATDD red phase: remove this skip while implementing Story 1.4")

	binary := buildDepgateBinary(t)
	result := runDepgateFixture(t, binary, "stdlib-only")

	if result.err != nil {
		t.Fatalf("depgate returned error for stdlib-only fixture: %v\n%s", result.err, result.output)
	}
	if strings.Contains(result.output, "non-standard import") {
		t.Fatalf("depgate reported a dependency violation for stdlib-only fixture:\n%s", result.output)
	}
}

func TestDepgateReportsNonStandardTestImports(t *testing.T) {
	t.Skip("ATDD red phase: remove this skip while implementing Story 1.4")

	binary := buildDepgateBinary(t)
	result := runDepgateFixture(t, binary, "non-stdlib-test-import")

	if result.err == nil {
		t.Fatalf("depgate succeeded for fixture with non-standard test imports:\n%s", result.output)
	}

	for _, want := range []string{
		"non-standard import:",
		"package=",
		"import=",
		"example.com/external/notstdlib",
		"example.com/external/also-notstdlib",
	} {
		if !strings.Contains(result.output, want) {
			t.Fatalf("depgate output missing %q:\n%s", want, result.output)
		}
	}
}

func TestDepgateReportsEveryViolationDeterministically(t *testing.T) {
	t.Skip("ATDD red phase: remove this skip while implementing Story 1.4")

	binary := buildDepgateBinary(t)
	first := runDepgateFixture(t, binary, "non-stdlib-test-import")
	second := runDepgateFixture(t, binary, "non-stdlib-test-import")

	if first.err == nil || second.err == nil {
		t.Fatalf("depgate should fail both runs for non-standard imports\nfirst:\n%s\nsecond:\n%s", first.output, second.output)
	}
	if first.output != second.output {
		t.Fatalf("depgate diagnostics are not deterministic\nfirst:\n%s\nsecond:\n%s", first.output, second.output)
	}
	if got := strings.Count(first.output, "non-standard import:"); got != 2 {
		t.Fatalf("depgate reported %d violations, want 2:\n%s", got, first.output)
	}
}

type depgateRunResult struct {
	output string
	err    error
}

func buildDepgateBinary(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "depgate")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Env = append(cmd.Environ(), "GOFLAGS=-buildvcs=false")

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build depgate binary: %v\n%s", err, output.String())
	}

	return binary
}

func runDepgateFixture(t *testing.T, binary string, fixture string) depgateRunResult {
	t.Helper()

	cmd := exec.Command(binary)
	cmd.Dir = filepath.Join("testdata", fixture)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	return depgateRunResult{
		output: output.String(),
		err:    err,
	}
}
