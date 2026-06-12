package cli_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/cli"
)

func TestQAInvocationPublicWorkflowUsesCallerSuppliedArgv(t *testing.T) {
	t.Parallel()

	argv := []string{"dib", "deploy", "--token=dib_fake_secret_value", "manifest.yaml"}
	invocation, err := cli.FromOSArgs(argv)
	if err != nil {
		t.Fatalf("FromOSArgs returned unexpected error: %v", err)
	}

	if got := invocation.Program(); got != "dib" {
		t.Fatalf("Program() = %q, want dib", got)
	}
	if got := invocation.Args(); !reflect.DeepEqual(got, []string{"deploy", "--token=dib_fake_secret_value", "manifest.yaml"}) {
		t.Fatalf("Args() = %q, want caller-supplied user args", got)
	}

	argv[0] = "mutated-program"
	argv[1] = "mutated-command"
	returnedArgs := invocation.Args()
	returnedArgs[1] = "mutated-token"

	if got := invocation.Program(); got != "dib" {
		t.Fatalf("Program() changed after argv mutation: %q", got)
	}
	if got := invocation.Args(); !reflect.DeepEqual(got, []string{"deploy", "--token=dib_fake_secret_value", "manifest.yaml"}) {
		t.Fatalf("Args() changed after caller mutation: %q", got)
	}

	reused := cli.FromArgs(invocation.Program(), invocation.Args())
	if got := reused.Program(); got != "dib" {
		t.Fatalf("reused Program() = %q, want dib", got)
	}
	if got := reused.Args(); !reflect.DeepEqual(got, []string{"deploy", "--token=dib_fake_secret_value", "manifest.yaml"}) {
		t.Fatalf("reused Args() = %q, want copied invocation args", got)
	}
}

func TestQAInvocationInvalidFullArgvDiagnosticsAreTypedAndValueFree(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		argv []string
	}{
		{name: "nil argv", argv: nil},
		{name: "empty argv", argv: []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			invocation, err := cli.FromOSArgs(tc.argv)
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
			for _, forbidden := range []string{"dib_fake_secret_value", "dib_fake_token_value", "--token"} {
				if strings.Contains(err.Error(), forbidden) || strings.Contains(invocationErr.Problem(), forbidden) {
					t.Fatalf("diagnostic leaked %q: error=%q problem=%q", forbidden, err.Error(), invocationErr.Problem())
				}
			}
		})
	}
}
