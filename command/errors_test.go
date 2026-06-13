package command_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/flags"
)

func TestUnknownCommandErrorParentPathIsDefensive(t *testing.T) {
	root := mustRoutingTree(t)

	_, err := root.Route([]string{"deploy", "missing"})
	if err == nil {
		t.Fatal("Route returned nil error")
	}

	var unknown *command.UnknownCommandError
	if !errors.As(err, &unknown) {
		t.Fatalf("error does not expose UnknownCommandError: %T", err)
	}

	parent := unknown.ParentPath()
	parent[0] = "mutated"
	if got := unknown.ParentPath(); !reflect.DeepEqual(got, []string{"dib", "deploy"}) {
		t.Fatalf("ParentPath() leaked mutable slice: %q", got)
	}
}

func TestTokenConflictErrorParentPathIsDefensive(t *testing.T) {
	root := mustDefinition(t, "dib")
	deploy := mustDefinition(t, "deploy", command.Aliases("run"))
	status := mustDefinition(t, "status", command.Aliases("run"))

	_, err := root.WithChildren(deploy, status)
	if err == nil {
		t.Fatal("WithChildren returned nil error")
	}

	var conflict *command.TokenConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("error does not expose TokenConflictError: %T", err)
	}

	parent := conflict.ParentPath()
	parent[0] = "mutated"
	if got := conflict.ParentPath(); !reflect.DeepEqual(got, []string{"dib"}) {
		t.Fatalf("ParentPath() leaked mutable slice: %q", got)
	}
}

func TestAliasErrorParentPathIsDefensive(t *testing.T) {
	_, err := command.NewDefinition("deploy", command.Aliases(" "))
	if err == nil {
		t.Fatal("NewDefinition returned nil error")
	}

	var aliasErr *command.AliasError
	if !errors.As(err, &aliasErr) {
		t.Fatalf("error does not expose AliasError: %T", err)
	}

	parent := aliasErr.ParentPath()
	parent = append(parent, "mutated")
	if got := aliasErr.ParentPath(); len(got) != 0 {
		t.Fatalf("ParentPath() leaked mutable slice: %q (after local append %q)", got, parent)
	}
}

func TestFlagCompositionErrorAccessorsAreDefensive(t *testing.T) {
	apply := mustDefinition(t, "apply", command.LocalFlags(flags.Bool("verbose", false, "")))

	_, err := command.NewDefinition(
		"dib",
		command.InheritedFlags(flags.Bool("verbose", false, "")),
		command.Children(apply),
	)
	if err == nil {
		t.Fatal("NewDefinition returned nil error")
	}

	var composition *command.FlagCompositionError
	if !errors.As(err, &composition) {
		t.Fatalf("error does not expose *command.FlagCompositionError: %T", err)
	}
	if got := composition.Scope(); got != "available" {
		t.Fatalf("Scope() = %q, want %q", got, "available")
	}
	path := composition.Path()
	path[0] = "mutated"
	if got := composition.Path(); !reflect.DeepEqual(got, []string{"dib", "apply"}) {
		t.Fatalf("Path() leaked mutable slice: %q", got)
	}
}
