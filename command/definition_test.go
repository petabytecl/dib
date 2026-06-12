package command_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/command"
)

func TestNewDefinitionAcceptsStableName(t *testing.T) {
	definition, err := command.NewDefinition(
		"serve",
		command.Description("run the service"),
		command.Aliases("server", "svc"),
		command.Usage("serve [flags]"),
	)
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	if got := definition.Name(); got != "serve" {
		t.Fatalf("Name() = %q, want %q", got, "serve")
	}
	if got := definition.Description(); got != "run the service" {
		t.Fatalf("Description() = %q, want %q", got, "run the service")
	}
	if got := definition.Aliases(); !reflect.DeepEqual(got, []string{"server", "svc"}) {
		t.Fatalf("Aliases() = %q, want %q", got, []string{"server", "svc"})
	}
	if got := definition.Usage(); got != "serve [flags]" {
		t.Fatalf("Usage() = %q, want %q", got, "serve [flags]")
	}
}

func TestNewDefinitionRejectsBlankNames(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "spaces", in: "   \t\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := command.NewDefinition(tt.in)
			if err == nil {
				t.Fatal("NewDefinition returned nil error")
			}
			var nameErr *command.NameError
			if !errors.As(err, &nameErr) {
				t.Fatalf("error does not expose NameError: %T", err)
			}
		})
	}
}

func TestDefinitionDerivationDoesNotMutateOriginalsOrLeakAliases(t *testing.T) {
	apply, err := command.NewDefinition("apply")
	if err != nil {
		t.Fatalf("NewDefinition(apply) returned unexpected error: %v", err)
	}
	destroy, err := command.NewDefinition("destroy")
	if err != nil {
		t.Fatalf("NewDefinition(destroy) returned unexpected error: %v", err)
	}
	deploy, err := command.NewDefinition("deploy", command.Aliases("ship"))
	if err != nil {
		t.Fatalf("NewDefinition(deploy) returned unexpected error: %v", err)
	}

	derived := deploy.
		WithDescription("deploy workloads").
		WithAliases("dep", "release").
		WithUsage("deploy <command>")
	derived, err = derived.WithChildren(apply)
	if err != nil {
		t.Fatalf("WithChildren returned unexpected error: %v", err)
	}

	if deploy.Description() != "" || deploy.Usage() != "" {
		t.Fatalf("original definition metadata mutated: description=%q usage=%q", deploy.Description(), deploy.Usage())
	}
	if got := deploy.Aliases(); !reflect.DeepEqual(got, []string{"ship"}) {
		t.Fatalf("original Aliases() = %q, want %q", got, []string{"ship"})
	}
	if got := deploy.Children(); len(got) != 0 {
		t.Fatalf("original Children() length = %d, want 0", len(got))
	}

	aliases := derived.Aliases()
	aliases[0] = "mutated"
	if got := derived.Aliases(); !reflect.DeepEqual(got, []string{"dep", "release"}) {
		t.Fatalf("derived Aliases() leaked mutable slice: %q", got)
	}

	children := derived.Children()
	children[0] = destroy
	if got := derived.Children(); len(got) != 1 || got[0].Name() != "apply" {
		t.Fatalf("derived Children() leaked mutable slice: %v", got)
	}
}

func TestDefinitionRejectsBlankChildNames(t *testing.T) {
	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition(dib) returned unexpected error: %v", err)
	}

	_, err = root.WithChildren(command.Definition{})
	if err == nil {
		t.Fatal("WithChildren returned nil error")
	}
	var nameErr *command.NameError
	if !errors.As(err, &nameErr) {
		t.Fatalf("error does not expose NameError: %T", err)
	}
}

func ExampleNewDefinition() {
	definition, err := command.NewDefinition("serve")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(definition.Name())
	// Output: serve
}
