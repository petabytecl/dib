package command_test

import (
	"reflect"
	"testing"

	"github.com/petabytecl/dib/command"
)

func TestRouteResultExposesDefensivePathDefinitions(t *testing.T) {
	root := mustRoutingTree(t)

	result, err := root.Route([]string{"deploy", "apply"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}

	path := result.Path()
	if got := pathNames(path); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("Path() names = %q, want %q", got, []string{"dib", "deploy", "apply"})
	}

	replacement, err := command.NewDefinition("replacement")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}
	path[0] = replacement

	if got := pathNames(result.Path()); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("Path() leaked mutable slice: %q", got)
	}
}

func TestRouteResultExposesDefensiveMatchTokens(t *testing.T) {
	root := mustRoutingTree(t)

	result, err := root.Route([]string{"ship", "push"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}

	tokens := result.MatchTokens()
	if !reflect.DeepEqual(tokens, []string{"dib", "ship", "push"}) {
		t.Fatalf("MatchTokens() = %q, want %q", tokens, []string{"dib", "ship", "push"})
	}
	tokens[0] = "mutated"
	if got := result.MatchTokens(); !reflect.DeepEqual(got, []string{"dib", "ship", "push"}) {
		t.Fatalf("MatchTokens() leaked mutable slice: %q", got)
	}
}

func pathNames(path []command.Definition) []string {
	names := make([]string, len(path))
	for i, def := range path {
		names[i] = def.Name()
	}
	return names
}
