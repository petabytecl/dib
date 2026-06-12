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

func TestLintPassesCleanFormattedGoFiles(t *testing.T) {
	root := t.TempDir()
	writeLintFile(t, root, "pkg/pkg.go", "package pkg\n\nfunc Name() string {\n\treturn \"dib\"\n}\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(context.Background(), root, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("run() wrote diagnostics for clean files:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("run() wrote unexpected stderr:\n%s", stderr.String())
	}
}

func TestLintReportsUnformattedFilesDeterministically(t *testing.T) {
	root := t.TempDir()
	writeLintFile(t, root, "b/b.go", "package b\n\nfunc B() string{return \"b\"}\n")
	writeLintFile(t, root, "a/a.go", "package a\n\nfunc A() string{return \"a\"}\n")

	var stdout bytes.Buffer
	exitCode := run(context.Background(), root, &stdout, &bytes.Buffer{})

	if exitCode != lintViolationExit {
		t.Fatalf("run() exit code = %d, want %d", exitCode, lintViolationExit)
	}

	got := stdout.String()
	want := "lint: a/a.go: not gofmt-formatted\nlint: b/b.go: not gofmt-formatted\n"
	if got != want {
		t.Fatalf("lint diagnostics mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestLintSkipsRepositoryMetadataAndStoryArtifacts(t *testing.T) {
	root := t.TempDir()
	writeLintFile(t, root, "pkg/pkg.go", "package pkg\n")
	writeLintFile(t, root, ".agents/bad.go", "package agents\n\nfunc Bad() string{return \"bad\"}\n")
	writeLintFile(t, root, ".codex/bad.go", "package codex\n\nfunc Bad() string{return \"bad\"}\n")
	writeLintFile(t, root, ".git/bad.go", "package git\n\nfunc Bad() string{return \"bad\"}\n")
	writeLintFile(t, root, "_bmad/bad.go", "package bmad\n\nfunc Bad() string{return \"bad\"}\n")
	writeLintFile(t, root, "_bmad-output/bad.go", "package bmad\n\nfunc Bad() string{return \"bad\"}\n")

	var stdout bytes.Buffer
	exitCode := run(context.Background(), root, &stdout, &bytes.Buffer{})

	if exitCode != 0 {
		t.Fatalf("run() exit code = %d, want 0\nstdout:\n%s", exitCode, stdout.String())
	}
}

func TestLintSeparatesExecutionFailuresFromLintViolations(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(context.Background(), filepath.Join(t.TempDir(), "missing"), &stdout, &stderr)

	if exitCode != executionFailureExit {
		t.Fatalf("run() exit code = %d, want %d", exitCode, executionFailureExit)
	}
	if stdout.Len() != 0 {
		t.Fatalf("run() wrote lint diagnostics for execution failure:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "lint execution error:") {
		t.Fatalf("stderr = %q, want execution error diagnostic", stderr.String())
	}
}

func TestLintReportsInvalidGoSourceAsExecutionFailure(t *testing.T) {
	root := t.TempDir()
	writeLintFile(t, root, "pkg/broken.go", "package pkg\n\nfunc Broken( {\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(context.Background(), root, &stdout, &stderr)

	if exitCode != executionFailureExit {
		t.Fatalf("run() exit code = %d, want %d", exitCode, executionFailureExit)
	}
	if stdout.Len() != 0 {
		t.Fatalf("run() wrote lint diagnostics for invalid Go source:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "lint execution error: walk") ||
		!strings.Contains(stderr.String(), "format") ||
		!strings.Contains(stderr.String(), "broken.go") {
		t.Fatalf("stderr = %q, want formatted execution error for invalid Go source", stderr.String())
	}
}

func TestLintCommandRunsFromRepositoryRoot(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	command := exec.Command("go", "run", "./tools/lint")
	command.Dir = repoRoot
	command.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "go-build"))

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("go run ./tools/lint failed: %v\n%s", err, string(output))
	}
	if len(output) != 0 {
		t.Fatalf("go run ./tools/lint wrote unexpected output:\n%s", string(output))
	}
}

func writeLintFile(t *testing.T, root string, name string, content string) {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
