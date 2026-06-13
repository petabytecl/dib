package command_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/flags"
)

type boundaryContextKey struct{}

func TestRouteBoundaryPackagesExplicitCallerInputs(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	originalStdin := os.Stdin
	t.Cleanup(func() {
		os.Args = originalArgs
		os.Stdout = originalStdout
		os.Stderr = originalStderr
		os.Stdin = originalStdin
	})

	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout-*")
	if err != nil {
		t.Fatalf("CreateTemp stdout returned unexpected error: %v", err)
	}
	stderrFile, err := os.CreateTemp(t.TempDir(), "stderr-*")
	if err != nil {
		t.Fatalf("CreateTemp stderr returned unexpected error: %v", err)
	}
	stdinFile, err := os.CreateTemp(t.TempDir(), "stdin-*")
	if err != nil {
		t.Fatalf("CreateTemp stdin returned unexpected error: %v", err)
	}
	if _, err := stdinFile.WriteString("ambient stdin"); err != nil {
		t.Fatalf("stdin WriteString returned unexpected error: %v", err)
	}
	if _, err := stdinFile.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("stdin Seek returned unexpected error: %v", err)
	}
	os.Args = []string{"dib", "ambient", "--wrong"}
	os.Stdout = stdoutFile
	os.Stderr = stderrFile
	os.Stdin = stdinFile
	t.Setenv("DIB_COMMAND_NAME", "ambient")
	t.Setenv("COLUMNS", "12")
	t.Chdir(t.TempDir())

	root := mustFlagRoutingTree(t)
	ctx := context.WithValue(context.Background(), boundaryContextKey{}, "trace-123")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	args := []string{"--verbose", "deploy", "apply", "--dry-run", "manifest.yaml"}

	boundary, err := root.RouteBoundary(ctx, args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("RouteBoundary returned unexpected error: %v", err)
	}
	args[0] = "--mutated"

	if got := boundary.Context().Value(boundaryContextKey{}); got != "trace-123" {
		t.Fatalf("Context value = %v, want trace-123", got)
	}
	if got := boundary.Args(); !reflect.DeepEqual(got, []string{"--verbose", "deploy", "apply", "--dry-run", "manifest.yaml"}) {
		t.Fatalf("Args() = %q, want explicit caller args", got)
	}
	result, ok := boundary.Result()
	if !ok {
		t.Fatal("Result() returned ok=false")
	}
	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("Result().PathNames() = %q, want routed path", got)
	}
	if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"manifest.yaml"}) {
		t.Fatalf("Result().RemainingArgs() = %q, want manifest.yaml", got)
	}
	snapshot, ok := result.FlagSnapshot()
	if !ok {
		t.Fatal("Result().FlagSnapshot() returned ok=false")
	}
	assertFlagValues(t, snapshot, "verbose", []any{true}, true, []string{"--verbose"})
	assertFlagValues(t, snapshot, "dry-run", []any{true}, true, []string{"--dry-run"})

	if got, ok := boundary.Stdout(); !ok || got != io.Writer(&stdout) {
		t.Fatalf("Stdout() = (%T, %v), want supplied stdout writer", got, ok)
	}
	if got, ok := boundary.Stderr(); !ok || got != io.Writer(&stderr) {
		t.Fatalf("Stderr() = (%T, %v), want supplied stderr writer", got, ok)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("RouteBoundary wrote to supplied writers: stdout=%q stderr=%q", stdout.String(), stderr.String())
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

func TestBoundaryPreservesCanceledContextWithoutPolicy(t *testing.T) {
	root := mustRoutingTree(t)
	result, err := root.Route([]string{"deploy", "apply"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	boundary := command.NewBoundary(ctx, result, []string{"deploy", "apply"}, discardWriter{}, discardWriter{})

	if got := boundary.Context().Err(); !errors.Is(got, context.Canceled) {
		t.Fatalf("Context().Err() = %v, want context.Canceled", got)
	}
	if result, ok := boundary.Result(); !ok || !reflect.DeepEqual(result.PathNames(), []string{"dib", "deploy", "apply"}) {
		t.Fatalf("Result() = (%q, %v), want routed result despite canceled context", result.PathNames(), ok)
	}
}

func TestBoundaryRetainsWritersWithoutWriting(t *testing.T) {
	root := mustRoutingTree(t)
	result, err := root.Route([]string{"deploy", "apply"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	stdout := &countingWriter{}
	stderr := &countingWriter{}

	boundary := command.NewBoundary(context.Background(), result, []string{"deploy", "apply"}, stdout, stderr)

	if got, ok := boundary.Stdout(); !ok || got != io.Writer(stdout) {
		t.Fatalf("Stdout() = (%T, %v), want supplied stdout writer", got, ok)
	}
	if got, ok := boundary.Stderr(); !ok || got != io.Writer(stderr) {
		t.Fatalf("Stderr() = (%T, %v), want supplied stderr writer", got, ok)
	}
	if stdout.writes != 0 || stderr.writes != 0 {
		t.Fatalf("NewBoundary wrote to writers: stdout writes=%d stderr writes=%d", stdout.writes, stderr.writes)
	}
}

func TestRouteBoundaryPreservesTypedRoutingAndFlagErrors(t *testing.T) {
	t.Run("unknown command", func(t *testing.T) {
		boundary, err := mustRoutingTree(t).RouteBoundary(context.Background(), []string{"deploy", "missing"}, discardWriter{}, discardWriter{})
		if err == nil {
			t.Fatal("RouteBoundary returned nil error")
		}
		if !errors.Is(err, command.ErrUnknownCommand) {
			t.Fatalf("error does not satisfy ErrUnknownCommand: %v", err)
		}
		var unknown *command.UnknownCommandError
		if !errors.As(err, &unknown) {
			t.Fatalf("error does not expose *command.UnknownCommandError: %T", err)
		}
		if _, ok := boundary.Result(); ok {
			t.Fatal("failed RouteBoundary returned a result")
		}
	})

	t.Run("flag parse failure", func(t *testing.T) {
		boundary, err := mustFlagRoutingTree(t).RouteBoundary(context.Background(), []string{"deploy", "apply", "--missing"}, discardWriter{}, discardWriter{})
		if err == nil {
			t.Fatal("RouteBoundary returned nil error")
		}
		if !errors.Is(err, flags.ErrUnknownFlag) {
			t.Fatalf("error does not satisfy flags.ErrUnknownFlag: %v", err)
		}
		var parseErr *flags.ParseError
		if !errors.As(err, &parseErr) {
			t.Fatalf("error does not expose *flags.ParseError: %T", err)
		}
		if _, ok := boundary.Result(); ok {
			t.Fatal("failed RouteBoundary returned a result")
		}
	})
}

func TestBoundaryKeepsOrdinaryCallerErrorsSeparate(t *testing.T) {
	boundary, err := mustRoutingTree(t).RouteBoundary(context.Background(), []string{"deploy", "apply"}, discardWriter{}, discardWriter{})
	if err != nil {
		t.Fatalf("RouteBoundary returned unexpected error: %v", err)
	}
	if _, ok := boundary.Result(); !ok {
		t.Fatal("RouteBoundary returned no result")
	}

	callerErr := errors.New("caller execution failure")
	if !errors.Is(callerErr, callerErr) {
		t.Fatal("ordinary caller error is not inspectable as itself")
	}
	if errors.Is(callerErr, command.ErrUnknownCommand) {
		t.Fatalf("ordinary caller error was converted to command.ErrUnknownCommand: %v", callerErr)
	}
	if errors.Is(callerErr, command.ErrFlagComposition) {
		t.Fatalf("ordinary caller error was converted to command.ErrFlagComposition: %v", callerErr)
	}
	var parseErr *flags.ParseError
	if errors.As(callerErr, &parseErr) {
		t.Fatalf("ordinary caller error was converted to *flags.ParseError: %v", callerErr)
	}
}

func TestBoundaryAccessorsAreDefensiveAndReusable(t *testing.T) {
	root := mustRoutingTree(t)
	result, err := root.Route([]string{"deploy", "apply", "manifest.yaml"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	args := []string{"deploy", "apply", "manifest.yaml"}
	boundary := command.NewBoundary(context.Background(), result, args, discardWriter{}, discardWriter{})
	args[2] = "mutated"

	gotArgs := boundary.Args()
	gotArgs[0] = "mutated"
	if got := boundary.Args(); !reflect.DeepEqual(got, []string{"deploy", "apply", "manifest.yaml"}) {
		t.Fatalf("Args() leaked mutable slice: %q", got)
	}

	gotResult, ok := boundary.Result()
	if !ok {
		t.Fatal("Result() returned ok=false")
	}
	path := gotResult.Path()
	path[0] = mustDefinition(t, "mutated")
	again, ok := boundary.Result()
	if !ok {
		t.Fatal("second Result() returned ok=false")
	}
	if got := again.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("Result() leaked mutable route state: %q", got)
	}

	const runs = 32
	errs := make(chan string, runs)
	var wg sync.WaitGroup
	for range runs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, ok := boundary.Result()
			if !ok {
				errs <- "missing result"
				return
			}
			if !reflect.DeepEqual(result.PathNames(), []string{"dib", "deploy", "apply"}) {
				errs <- "unexpected path"
				return
			}
			if !reflect.DeepEqual(boundary.Args(), []string{"deploy", "apply", "manifest.yaml"}) {
				errs <- "unexpected args"
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func TestBoundaryZeroValueIsAbsentState(t *testing.T) {
	var boundary command.Boundary

	if boundary.Context() != nil {
		t.Fatal("zero Boundary Context() returned non-nil context")
	}
	if got := boundary.Args(); got != nil {
		t.Fatalf("zero Boundary Args() = %q, want nil", got)
	}
	if _, ok := boundary.Result(); ok {
		t.Fatal("zero Boundary Result() returned ok=true")
	}
	if _, ok := boundary.Stdout(); ok {
		t.Fatal("zero Boundary Stdout() returned ok=true")
	}
	if _, ok := boundary.Stderr(); ok {
		t.Fatal("zero Boundary Stderr() returned ok=true")
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

type countingWriter struct {
	writes int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.writes++
	return len(p), nil
}
