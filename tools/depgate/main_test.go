package main

import (
	"bytes"
	"context"
	"os"
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
		denyOutput []string
	}{
		{
			name:    "stdlib-only",
			fixture: "stdlib-only",
		},
		{
			name:    "resolved non-stdlib import",
			fixture: "resolved-nonstdlib-import",
			wantErr: true,
			wantOutput: []string{
				"non-standard import:",
				"package=example.com/dib-depgate-resolved-fixture/pkg",
				"import=example.com/external/resolved",
			},
			denyOutput: []string{
				"package=example.com/external/resolved import=example.com/external/resolved",
			},
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
			for _, denied := range tt.denyOutput {
				if strings.Contains(result.output, denied) {
					t.Fatalf("depgate output included denied diagnostic %q:\n%s", denied, result.output)
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

func TestDepgateDisablesWorkspaceMode(t *testing.T) {
	binary := buildDepgateBinary(t)
	workspace := t.TempDir()

	writeFile(t, workspace, "go.work", `go 1.26

use (
	./app
	./external
)
`)
	writeFile(t, workspace, "app/go.mod", `module example.com/dib-depgate-workspace-fixture

go 1.26

require example.com/external/workspace v0.0.0

replace example.com/external/workspace => ../external
`)
	writeFile(t, workspace, "app/pkg/pkg.go", `package pkg

import "example.com/external/workspace"

func Name() string {
	return workspace.Name()
}
`)
	writeFile(t, workspace, "external/go.mod", `module example.com/external/workspace

go 1.26
`)
	writeFile(t, workspace, "external/workspace.go", `package workspace

func Name() string {
	return "workspace"
}
`)

	cmd := exec.Command(binary)
	cmd.Dir = filepath.Join(workspace, "app")

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err == nil {
		t.Fatalf("depgate succeeded unexpectedly in workspace fixture:\n%s", output.String())
	}

	got := output.String()
	for _, want := range []string{
		"non-standard import:",
		"package=example.com/dib-depgate-workspace-fixture/pkg",
		"import=example.com/external/workspace",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("depgate workspace output missing %q:\n%s", want, got)
		}
	}
}

func TestCollectViolationsIncludesDepsErrors(t *testing.T) {
	packages := []listedPackage{
		{
			ImportPath: "example.com/dib-depgate-fixture/pkg.test",
			Module:     &listedModule{Main: true},
			DepsErrors: []*listedError{
				{
					ImportStack: []string{"example.com/dib-depgate-fixture/pkg_test"},
					Err:         "no required module provides package example.com/external/missing; to add it:\n\tgo get example.com/external/missing",
				},
			},
		},
	}

	violations := collectViolations(packages)
	if len(violations) != 1 {
		t.Fatalf("collectViolations() returned %d violations, want 1: %#v", len(violations), violations)
	}

	want := violation{
		Package: "example.com/dib-depgate-fixture/pkg_test",
		Import:  "example.com/external/missing",
	}
	if violations[0] != want {
		t.Fatalf("collectViolations()[0] = %#v, want %#v", violations[0], want)
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

func runDepgateFixture(t *testing.T, binary, fixture string) depgateRunResult {
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

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create parent directory for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}
