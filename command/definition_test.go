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

	derived, err := deploy.
		WithDescription("deploy workloads").
		WithAliases("dep", "release")
	if err != nil {
		t.Fatalf("WithAliases returned unexpected error: %v", err)
	}
	derived = derived.WithUsage("deploy <command>")
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

func TestDefinitionRejectsInvalidAliases(t *testing.T) {
	tests := []struct {
		name    string
		command string
		aliases []string
	}{
		{name: "blank alias", command: "deploy", aliases: []string{" "}},
		{name: "self alias", command: "deploy", aliases: []string{"deploy"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := command.NewDefinition(tt.command, command.Aliases(tt.aliases...))
			if err == nil {
				t.Fatal("NewDefinition returned nil error")
			}
			if !errors.Is(err, command.ErrInvalidCommandAlias) {
				t.Fatalf("error does not satisfy ErrInvalidCommandAlias: %v", err)
			}
			var aliasErr *command.AliasError
			if !errors.As(err, &aliasErr) {
				t.Fatalf("error does not expose AliasError: %T", err)
			}
			if got := aliasErr.Command(); got != tt.command {
				t.Fatalf("Command() = %q, want %q", got, tt.command)
			}
		})
	}
}

func TestDefinitionRejectsAliasTokenCollisions(t *testing.T) {
	t.Run("duplicate aliases on one command", func(t *testing.T) {
		_, err := command.NewDefinition("deploy", command.Aliases("ship", "ship"))
		if err == nil {
			t.Fatal("NewDefinition returned nil error")
		}
		if !errors.Is(err, command.ErrDuplicateCommandToken) {
			t.Fatalf("error does not satisfy ErrDuplicateCommandToken: %v", err)
		}
		var conflict *command.TokenConflictError
		if !errors.As(err, &conflict) {
			t.Fatalf("error does not expose TokenConflictError: %T", err)
		}
		if got := conflict.Token(); got != "ship" {
			t.Fatalf("Token() = %q, want %q", got, "ship")
		}
		if got := conflict.FirstCommand(); got != "deploy" {
			t.Fatalf("FirstCommand() = %q, want %q", got, "deploy")
		}
		if got := conflict.CollidingCommand(); got != "deploy" {
			t.Fatalf("CollidingCommand() = %q, want %q", got, "deploy")
		}
	})

	tests := []struct {
		name          string
		children      func(t *testing.T) []command.Definition
		wantToken     string
		wantFirst     string
		wantColliding string
	}{
		{
			name: "duplicate child names",
			children: func(t *testing.T) []command.Definition {
				return []command.Definition{
					mustDefinition(t, "deploy"),
					mustDefinition(t, "deploy"),
				}
			},
			wantToken:     "deploy",
			wantFirst:     "deploy",
			wantColliding: "deploy",
		},
		{
			name: "alias matches sibling child name",
			children: func(t *testing.T) []command.Definition {
				deploy := mustDefinition(t, "deploy", command.Aliases("status"))
				status := mustDefinition(t, "status")
				return []command.Definition{deploy, status}
			},
			wantToken:     "status",
			wantFirst:     "status",
			wantColliding: "deploy",
		},
		{
			name: "alias matches sibling alias",
			children: func(t *testing.T) []command.Definition {
				deploy := mustDefinition(t, "deploy", command.Aliases("run"))
				status := mustDefinition(t, "status", command.Aliases("run"))
				return []command.Definition{deploy, status}
			},
			wantToken:     "run",
			wantFirst:     "deploy",
			wantColliding: "status",
		},
		{
			name: "cross alias cycle",
			children: func(t *testing.T) []command.Definition {
				apply := mustDefinition(t, "apply", command.Aliases("plan"))
				plan := mustDefinition(t, "plan", command.Aliases("apply"))
				return []command.Definition{apply, plan}
			},
			wantToken:     "plan",
			wantFirst:     "plan",
			wantColliding: "apply",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := mustDefinition(t, "dib")
			_, err := root.WithChildren(tt.children(t)...)
			if err == nil {
				t.Fatal("WithChildren returned nil error")
			}
			if !errors.Is(err, command.ErrDuplicateCommandToken) {
				t.Fatalf("error does not satisfy ErrDuplicateCommandToken: %v", err)
			}
			var conflict *command.TokenConflictError
			if !errors.As(err, &conflict) {
				t.Fatalf("error does not expose TokenConflictError: %T", err)
			}
			if got := conflict.ParentPath(); !reflect.DeepEqual(got, []string{"dib"}) {
				t.Fatalf("ParentPath() = %q, want %q", got, []string{"dib"})
			}
			if got := conflict.Token(); got != tt.wantToken {
				t.Fatalf("Token() = %q, want %q", got, tt.wantToken)
			}
			if got := conflict.FirstCommand(); got != tt.wantFirst {
				t.Fatalf("FirstCommand() = %q, want %q", got, tt.wantFirst)
			}
			if got := conflict.CollidingCommand(); got != tt.wantColliding {
				t.Fatalf("CollidingCommand() = %q, want %q", got, tt.wantColliding)
			}
		})
	}
}

func TestDefinitionDerivationValidationDoesNotMutateOriginals(t *testing.T) {
	apply := mustDefinition(t, "apply")
	plan := mustDefinition(t, "plan")
	deploy := mustDefinition(t, "deploy", command.Aliases("ship"))

	if _, err := deploy.WithAliases("deploy"); err == nil {
		t.Fatal("WithAliases returned nil error")
	}
	if got := deploy.Aliases(); !reflect.DeepEqual(got, []string{"ship"}) {
		t.Fatalf("failed WithAliases mutated original aliases: %q", got)
	}

	derived, err := deploy.WithChildren(apply)
	if err != nil {
		t.Fatalf("WithChildren(apply) returned unexpected error: %v", err)
	}
	if _, err := derived.WithChildren(apply, mustWithAliases(t, plan, "apply")); err == nil {
		t.Fatal("WithChildren returned nil error")
	}
	if got := derived.Children(); len(got) != 1 || got[0].Name() != "apply" {
		t.Fatalf("failed WithChildren mutated previous derived children: %v", got)
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

func mustDefinition(t *testing.T, name string, options ...command.Option) command.Definition {
	t.Helper()

	definition, err := command.NewDefinition(name, options...)
	if err != nil {
		t.Fatalf("NewDefinition(%q) returned unexpected error: %v", name, err)
	}
	return definition
}

func mustWithAliases(t *testing.T, d command.Definition, aliases ...string) command.Definition {
	t.Helper()

	derived, err := d.WithAliases(aliases...)
	if err != nil {
		t.Fatalf("WithAliases returned unexpected error: %v", err)
	}
	return derived
}
