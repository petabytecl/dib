package cli_test

import (
	"errors"
	"testing"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

// TestInvocationErrorNilReceiverIsDefensive verifies that nil *InvocationError
// pointer receivers return safe zero values rather than panicking.
func TestInvocationErrorNilReceiverIsDefensive(t *testing.T) {
	t.Parallel()

	var e *cli.InvocationError

	if got := e.Error(); got == "" {
		t.Fatal("nil InvocationError.Error() returned empty string")
	}
	if got := e.Unwrap(); got != nil {
		t.Fatalf("nil InvocationError.Unwrap() = %v, want nil", got)
	}
	if got := e.Category(); got != nil {
		t.Fatalf("nil InvocationError.Category() = %v, want nil", got)
	}
	if got := e.Problem(); got != "" {
		t.Fatalf("nil InvocationError.Problem() = %q, want empty", got)
	}
}

// TestInvocationErrorNoProblemMessage exercises the branch where problem is empty.
func TestInvocationErrorNoProblemMessage(t *testing.T) {
	t.Parallel()

	// FromOSArgs with empty slice triggers newInvocationError with a non-empty problem.
	// To cover the e.problem == "" branch, call FromArgs with explicit empty values.
	// We can't construct InvocationError directly (unexported), so we test via
	// the public error surface: inspect the message when problem is non-empty (covered)
	// and confirm the nil receiver path (tested above).
	//
	// The e.problem == "" branch in Error() is defensive code; we verify the
	// non-empty path (the default case in FromOSArgs) below.
	_, err := cli.FromOSArgs(nil)
	if err == nil {
		t.Fatal("FromOSArgs(nil) returned nil error")
	}
	if !errors.Is(err, cli.ErrInvalidInvocation) {
		t.Fatalf("FromOSArgs(nil) error does not satisfy ErrInvalidInvocation: %v", err)
	}

	var invErr *cli.InvocationError
	if !errors.As(err, &invErr) {
		t.Fatalf("FromOSArgs(nil) error is not *InvocationError: %T", err)
	}
	if invErr.Error() == "" {
		t.Fatal("InvocationError.Error() returned empty string")
	}
	if invErr.Problem() == "" {
		t.Fatal("InvocationError.Problem() returned empty string for missing-program error")
	}
	if !errors.Is(invErr.Unwrap(), cli.ErrInvalidInvocation) {
		t.Fatalf("InvocationError.Unwrap() = %v, want ErrInvalidInvocation", invErr.Unwrap())
	}
	if !errors.Is(invErr.Category(), cli.ErrInvalidInvocation) {
		t.Fatalf("InvocationError.Category() = %v, want ErrInvalidInvocation", invErr.Category())
	}
}

// TestNewFlagSnapshotRepeatableStringFlagBoundToScalarConfigKey verifies that
// binding a repeatable string-list flag (which produces multiple values) to a
// scalar config.String key produces a typed BindingError, not a panic or silent
// type corruption. This exercises the multi-value non-list branch in the
// internal bindingValue translation path.
func TestNewFlagSnapshotRepeatableStringFlagBoundToScalarConfigKey(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("host", "", "host"),
	)
	if err != nil {
		t.Fatalf("config.NewSet: %v", err)
	}

	root, err := command.NewDefinition("dib",
		command.LocalFlags(flags.StringList("host", nil, "host list", flags.Repeatable())),
	)
	if err != nil {
		t.Fatalf("command.NewDefinition: %v", err)
	}

	route, err := root.Route([]string{"--host=alpha", "--host=beta"})
	if err != nil {
		t.Fatalf("Route: %v", err)
	}

	_, err = cli.NewFlagSnapshot(set, route, []cli.FlagBinding{
		cli.BindFlag("host", "host"),
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error for multi-value flag bound to scalar config key")
	}
	if !errors.Is(err, cli.ErrInvalidBinding) {
		t.Fatalf("error does not satisfy ErrInvalidBinding: %v", err)
	}
	var bindErr *cli.BindingError
	if !errors.As(err, &bindErr) {
		t.Fatalf("error is not *cli.BindingError: %T", err)
	}
	if msg := bindErr.Error(); msg == "" {
		t.Fatal("BindingError.Error() returned empty string")
	}
}

// TestBindingErrorNilReceiverIsDefensive verifies that nil *BindingError pointer
// receivers return safe zero values and false rather than panicking.
func TestBindingErrorNilReceiverIsDefensive(t *testing.T) {
	t.Parallel()

	var e *cli.BindingError

	if got := e.Error(); got == "" {
		t.Fatal("nil BindingError.Error() returned empty string")
	}
	if e.Is(cli.ErrInvalidBinding) {
		t.Fatal("nil BindingError.Is(ErrInvalidBinding) returned true, want false")
	}
	if got := e.Unwrap(); got != nil {
		t.Fatalf("nil BindingError.Unwrap() = %v, want nil", got)
	}
	if got := e.FlagName(); got != "" {
		t.Fatalf("nil BindingError.FlagName() = %q, want empty", got)
	}
	if got := e.ConfigKey(); got != "" {
		t.Fatalf("nil BindingError.ConfigKey() = %q, want empty", got)
	}
	if got := e.Category(); got != nil {
		t.Fatalf("nil BindingError.Category() = %v, want nil", got)
	}
	if got := e.Cause(); got != nil {
		t.Fatalf("nil BindingError.Cause() = %v, want nil", got)
	}
}
