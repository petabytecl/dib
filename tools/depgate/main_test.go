package main

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDepgateFixtures(t *testing.T) {
	binary := buildDepgateBinary(t)

	tests := []struct {
		name       string
		fixture    string
		wantErr    bool
		wantOutput []string
	}{
		{
			name:    "stdlib-only",
			fixture: "stdlib-only",
		},
		{
			name:    "non-stdlib test imports",
			fixture: "non-stdlib-test-import",
			wantErr: true,
			wantOutput: []string{
				"non-standard import:",
				"package=",
				"import=",
				"example.com/external/notstdlib",
				"example.com/external/also-notstdlib",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runDepgateFixture(t, binary, tt.fixture)

			if tt.wantErr && result.err == nil {
				t.Fatalf("depgate succeeded unexpectedly:\n%s", result.output)
			}
			if !tt.wantErr && result.err != nil {
				t.Fatalf("depgate returned unexpected error: %v\n%s", result.err, result.output)
			}
			if !tt.wantErr && strings.Contains(result.output, "non-standard import") {
				t.Fatalf("depgate reported a dependency violation unexpectedly:\n%s", result.output)
			}
			for _, want := range tt.wantOutput {
				if !strings.Contains(result.output, want) {
					t.Fatalf("depgate output missing %q:\n%s", want, result.output)
				}
			}
		})
	}
}

func TestDepgateReportsEveryViolationDeterministically(t *testing.T) {
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

func TestRunSeparatesExecutionFailureFromDependencyViolations(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(context.Background(), filepath.Join(t.TempDir(), "missing"), &stdout, &stderr)

	if exitCode != executionFailureExit {
		t.Fatalf("run() exit code = %d, want %d", exitCode, executionFailureExit)
	}
	if stdout.Len() != 0 {
		t.Fatalf("run() wrote dependency diagnostics to stdout for execution failure:\n%s", stdout.String())
	}
	if got := stderr.String(); !strings.Contains(got, "depgate execution error:") {
		t.Fatalf("run() stderr = %q, want execution error diagnostic", got)
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
