package cli_test

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/cli"
)

func TestFromOSArgsUsesExplicitFullArgv(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	t.Cleanup(func() {
		os.Args = originalArgs
	})
	os.Args = []string{"ambient", "ignored", "--secret=ambient"}

	invocation, err := cli.FromOSArgs([]string{"dib", "deploy", "--verbose"})
	if err != nil {
		t.Fatalf("FromOSArgs returned unexpected error: %v", err)
	}

	if got := invocation.Program(); got != "dib" {
		t.Fatalf("Program() = %q, want dib", got)
	}
	if got := invocation.Args(); !reflect.DeepEqual(got, []string{"deploy", "--verbose"}) {
		t.Fatalf("Args() = %q, want explicit caller args", got)
	}
}

func TestFromArgsPreservesCallerProgram(t *testing.T) {
	tests := []struct {
		name    string
		program string
		args    []string
	}{
		{
			name:    "named program",
			program: "dib",
			args:    []string{"deploy"},
		},
		{
			name:    "empty program is caller owned",
			program: "",
			args:    []string{"deploy"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			invocation := cli.FromArgs(tc.program, tc.args)

			if got := invocation.Program(); got != tc.program {
				t.Fatalf("Program() = %q, want %q", got, tc.program)
			}
			if got := invocation.Args(); !reflect.DeepEqual(got, tc.args) {
				t.Fatalf("Args() = %q, want %q", got, tc.args)
			}
		})
	}
}

func TestInvocationDefensiveCopies(t *testing.T) {
	t.Run("full argv constructor", func(t *testing.T) {
		argv := []string{"dib", "deploy", "--verbose"}
		invocation, err := cli.FromOSArgs(argv)
		if err != nil {
			t.Fatalf("FromOSArgs returned unexpected error: %v", err)
		}
		argv[0] = "mutated-program"
		argv[1] = "mutated-arg"

		if got := invocation.Program(); got != "dib" {
			t.Fatalf("Program() = %q, want dib", got)
		}
		if got := invocation.Args(); !reflect.DeepEqual(got, []string{"deploy", "--verbose"}) {
			t.Fatalf("Args() = %q, want original args", got)
		}
	})

	t.Run("stripped args constructor", func(t *testing.T) {
		args := []string{"deploy"}
		invocation := cli.FromArgs("dib", args)
		args[0] = "mutated"

		if got := invocation.Args(); !reflect.DeepEqual(got, []string{"deploy"}) {
			t.Fatalf("Args() = %q, want original args", got)
		}
	})

	t.Run("args accessor", func(t *testing.T) {
		invocation := cli.FromArgs("dib", []string{"deploy", "--verbose"})
		got := invocation.Args()
		got[0] = "mutated"

		if again := invocation.Args(); !reflect.DeepEqual(again, []string{"deploy", "--verbose"}) {
			t.Fatalf("Args() leaked mutable state: %q", again)
		}
	})
}

func TestFromOSArgsEmptyFullArgvReturnsTypedError(t *testing.T) {
	invocation, err := cli.FromOSArgs(nil)
	if err == nil {
		t.Fatal("FromOSArgs returned nil error")
	}
	if !reflect.DeepEqual(invocation, cli.Invocation{}) {
		t.Fatalf("FromOSArgs returned invocation %#v, want zero value", invocation)
	}
	if !errors.Is(err, cli.ErrInvalidInvocation) {
		t.Fatalf("error does not satisfy ErrInvalidInvocation: %v", err)
	}
	var invocationErr *cli.InvocationError
	if !errors.As(err, &invocationErr) {
		t.Fatalf("error does not expose *cli.InvocationError: %T", err)
	}
	if got := invocationErr.Category(); !errors.Is(got, cli.ErrInvalidInvocation) {
		t.Fatalf("Category() = %v, want ErrInvalidInvocation", got)
	}
	if got := invocationErr.Problem(); got != "missing program" {
		t.Fatalf("Problem() = %q, want missing program", got)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("Error() echoed argv contents: %q", err.Error())
	}
}
