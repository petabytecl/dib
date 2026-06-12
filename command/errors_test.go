package command_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/command"
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
