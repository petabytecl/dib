package command_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestRouteBoundaryPassesHelpRequestsWithoutRendering(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	t.Cleanup(func() {
		os.Args = originalArgs
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	})

	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout-*")
	if err != nil {
		t.Fatalf("CreateTemp stdout returned unexpected error: %v", err)
	}
	stderrFile, err := os.CreateTemp(t.TempDir(), "stderr-*")
	if err != nil {
		t.Fatalf("CreateTemp stderr returned unexpected error: %v", err)
	}
	os.Stdout = stdoutFile
	os.Stderr = stderrFile
	os.Args = []string{"ambient", "--help"}
	t.Setenv("DIB_WIDTH", "10")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	boundary, err := mustHelpTree(t).RouteBoundary(context.Background(), []string{"deploy", "apply", "--help"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("RouteBoundary returned nil error")
	}
	if !errors.Is(err, flags.ErrHelpRequest) {
		t.Fatalf("RouteBoundary error does not satisfy flags.ErrHelpRequest: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("RouteBoundary error does not expose *flags.ParseError: %T", err)
	}
	if _, ok := boundary.Result(); ok {
		t.Fatal("help-request RouteBoundary returned a result")
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("RouteBoundary rendered help or diagnostics: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	if err := stdoutFile.Sync(); err != nil {
		t.Fatalf("stdout Sync returned unexpected error: %v", err)
	}
	if err := stderrFile.Sync(); err != nil {
		t.Fatalf("stderr Sync returned unexpected error: %v", err)
	}
	assertEmptyFile(t, stdoutFile)
	assertEmptyFile(t, stderrFile)
}

func TestRouteBoundaryLeavesRenderingUnderCallerControl(t *testing.T) {
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	t.Cleanup(func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	})

	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout-*")
	if err != nil {
		t.Fatalf("CreateTemp stdout returned unexpected error: %v", err)
	}
	stderrFile, err := os.CreateTemp(t.TempDir(), "stderr-*")
	if err != nil {
		t.Fatalf("CreateTemp stderr returned unexpected error: %v", err)
	}
	os.Stdout = stdoutFile
	os.Stderr = stderrFile

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	boundary, err := mustHelpTree(t).RouteBoundary(
		context.Background(),
		[]string{"ship", "apply", "--cluster", "prod", "manifest.yaml"},
		&stdout,
		&stderr,
	)
	if err != nil {
		t.Fatalf("RouteBoundary returned unexpected error: %v", err)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("RouteBoundary wrote before caller rendering: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}

	result, ok := boundary.Result()
	if !ok {
		t.Fatal("RouteBoundary returned no result")
	}
	stdoutWriter, ok := boundary.Stdout()
	if !ok {
		t.Fatal("Stdout() returned ok=false")
	}
	if err := result.WriteUsage(stdoutWriter); err != nil {
		t.Fatalf("Result.WriteUsage returned unexpected error: %v", err)
	}
	if got, want := stdout.String(), "Usage:\n  dib deploy apply <manifest> [flags]\n"; got != want {
		t.Fatalf("caller-rendered usage = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Fatalf("caller-controlled usage render wrote stderr: %q", stderr.String())
	}
	if err := stdoutFile.Sync(); err != nil {
		t.Fatalf("stdout Sync returned unexpected error: %v", err)
	}
	if err := stderrFile.Sync(); err != nil {
		t.Fatalf("stderr Sync returned unexpected error: %v", err)
	}
	assertEmptyFile(t, stdoutFile)
	assertEmptyFile(t, stderrFile)
}

func TestRouteBoundaryPreservesTerminatorPassthrough(t *testing.T) {
	boundary, err := mustFlagRoutingTree(t).RouteBoundary(
		context.Background(),
		[]string{"--verbose", "deploy", "apply", "--", "--dry-run", "manifest.yaml"},
		io.Discard,
		io.Discard,
	)
	if err != nil {
		t.Fatalf("RouteBoundary returned unexpected error: %v", err)
	}

	result, ok := boundary.Result()
	if !ok {
		t.Fatal("RouteBoundary returned no result")
	}
	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("PathNames() = %q, want routed command path", got)
	}
	if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"--dry-run", "manifest.yaml"}) {
		t.Fatalf("RemainingArgs() = %q, want terminator passthrough args", got)
	}
	snapshot, ok := result.FlagSnapshot()
	if !ok {
		t.Fatal("FlagSnapshot() returned ok=false")
	}
	assertFlagValues(t, snapshot, "verbose", []any{true}, true, []string{"--verbose"})
	assertFlagValues(t, snapshot, "dry-run", []any{false}, false, nil)
}
