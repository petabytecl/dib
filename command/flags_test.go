package command_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/flags"
)

func TestRouteComposesInheritedAndLocalFlags(t *testing.T) {
	root := mustFlagRoutingTree(t)

	result, err := root.Route([]string{"--verbose", "deploy", "--cluster", "prod", "apply", "--dry-run", "manifest.yaml"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("PathNames() = %q, want %q", got, []string{"dib", "deploy", "apply"})
	}
	if got := result.MatchTokens(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("MatchTokens() = %q, want %q", got, []string{"dib", "deploy", "apply"})
	}
	if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"manifest.yaml"}) {
		t.Fatalf("RemainingArgs() = %q, want %q", got, []string{"manifest.yaml"})
	}

	set, ok := result.Flags()
	if !ok {
		t.Fatal("Flags() returned ok=false")
	}
	if got := flagNames(set.Definitions()); !reflect.DeepEqual(got, []string{"verbose", "cluster", "retries", "tag", "dry-run"}) {
		t.Fatalf("Flags().Definitions() names = %q, want %q", got, []string{"verbose", "cluster", "retries", "tag", "dry-run"})
	}

	snapshot, ok := result.FlagSnapshot()
	if !ok {
		t.Fatal("FlagSnapshot() returned ok=false")
	}
	assertFlagValues(t, snapshot, "verbose", []any{true}, true, []string{"--verbose"})
	assertFlagValues(t, snapshot, "cluster", []any{"prod"}, true, []string{"--cluster"})
	assertFlagValues(t, snapshot, "dry-run", []any{true}, true, []string{"--dry-run"})
}

func TestRouteKeepsSiblingAndAncestorLocalFlagsIsolated(t *testing.T) {
	root := mustFlagRoutingTree(t)

	tests := []struct {
		name string
		args []string
	}{
		{name: "sibling local does not parse", args: []string{"status", "--dry-run"}},
		{name: "ancestor local does not leak", args: []string{"deploy", "apply", "--target", "prod"}},
		{name: "nested sibling local does not parse", args: []string{"deploy", "plan", "--dry-run"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := root.Route(tt.args)
			if err == nil {
				t.Fatal("Route returned nil error")
			}
			if !errors.Is(err, flags.ErrUnknownFlag) {
				t.Fatalf("error does not satisfy flags.ErrUnknownFlag: %v", err)
			}
			var parseErr *flags.ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error does not expose *flags.ParseError: %T", err)
			}
			if got := result.PathNames(); len(got) != 0 {
				t.Fatalf("failed Route returned non-zero path: %q", got)
			}
			if _, ok := result.FlagSnapshot(); ok {
				t.Fatal("failed Route exposed a flag snapshot")
			}
		})
	}
}

func TestRouteFlagCompositionConflictsAreInspectable(t *testing.T) {
	tests := []struct {
		name      string
		root      func(t *testing.T) (command.Definition, error)
		wantCause error
	}{
		{
			name: "duplicate long name",
			root: func(t *testing.T) (command.Definition, error) {
				apply := mustDefinition(t, "apply", command.LocalFlags(flags.Bool("verbose", false, "")))
				return command.NewDefinition("dib", command.InheritedFlags(flags.Bool("verbose", false, "")), command.Children(apply))
			},
			wantCause: flags.ErrDuplicateName,
		},
		{
			name: "duplicate shorthand",
			root: func(t *testing.T) (command.Definition, error) {
				apply := mustDefinition(t, "apply", command.LocalFlags(flags.Bool("dry-run", false, "", flags.Shorthand("v"))))
				return command.NewDefinition("dib", command.InheritedFlags(flags.Bool("verbose", false, "", flags.Shorthand("v"))), command.Children(apply))
			},
			wantCause: flags.ErrDuplicateShorthand,
		},
		{
			name: "duplicate normalized name",
			root: func(t *testing.T) (command.Definition, error) {
				apply := mustDefinition(t, "apply", command.LocalFlags(flags.Bool("dry_run", false, "")))
				return command.NewDefinition(
					"dib",
					command.FlagNormalizer(func(name string) string { return strings.ReplaceAll(name, "_", "-") }),
					command.InheritedFlags(flags.Bool("dry-run", false, "")),
					command.Children(apply),
				)
			},
			wantCause: flags.ErrDuplicateNormalizedName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.root(t)
			if err == nil {
				t.Fatal("NewDefinition returned nil error")
			}
			if !errors.Is(err, command.ErrFlagComposition) {
				t.Fatalf("error does not satisfy command.ErrFlagComposition: %v", err)
			}
			if !errors.Is(err, tt.wantCause) {
				t.Fatalf("error does not satisfy %v: %v", tt.wantCause, err)
			}
			var composition *command.FlagCompositionError
			if !errors.As(err, &composition) {
				t.Fatalf("error does not expose *command.FlagCompositionError: %T", err)
			}
			if got := composition.Path(); !reflect.DeepEqual(got, []string{"dib", "apply"}) {
				t.Fatalf("Path() = %q, want %q", got, []string{"dib", "apply"})
			}
			var definitionErr *flags.DefinitionError
			if !errors.As(err, &definitionErr) {
				t.Fatalf("error does not expose *flags.DefinitionError: %T", err)
			}
		})
	}
}

func TestRoutePreservesParserBoundariesAndTypedFlagFailures(t *testing.T) {
	root := mustFlagRoutingTree(t)

	t.Run("interspersed flags and terminator passthrough", func(t *testing.T) {
		result, err := root.Route([]string{"deploy", "--cluster=prod", "apply", "--", "--dry-run", "literal"})
		if err != nil {
			t.Fatalf("Route returned unexpected error: %v", err)
		}
		if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
			t.Fatalf("PathNames() = %q, want %q", got, []string{"dib", "deploy", "apply"})
		}
		if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"--dry-run", "literal"}) {
			t.Fatalf("RemainingArgs() = %q, want %q", got, []string{"--dry-run", "literal"})
		}
		snapshot, ok := result.FlagSnapshot()
		if !ok {
			t.Fatal("FlagSnapshot() returned ok=false")
		}
		assertFlagValues(t, snapshot, "cluster", []any{"prod"}, true, []string{"--cluster"})
	})

	tests := []struct {
		name      string
		args      []string
		wantCause error
		wantToken string
	}{
		{name: "help request", args: []string{"deploy", "apply", "--help"}, wantCause: flags.ErrHelpRequest, wantToken: "--help"},
		{name: "unknown flag", args: []string{"deploy", "apply", "--unknown"}, wantCause: flags.ErrUnknownFlag, wantToken: "--unknown"},
		{name: "missing value", args: []string{"deploy", "apply", "--cluster"}, wantCause: flags.ErrMissingValue, wantToken: "--cluster"},
		{name: "conversion failure", args: []string{"deploy", "apply", "--retries", "nope"}, wantCause: flags.ErrConversion, wantToken: "--retries"},
		{name: "duplicate single value", args: []string{"deploy", "apply", "--cluster", "prod", "--cluster", "dev"}, wantCause: flags.ErrDuplicateValue, wantToken: "--cluster"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := root.Route(tt.args)
			if err == nil {
				t.Fatal("Route returned nil error")
			}
			if !errors.Is(err, tt.wantCause) {
				t.Fatalf("error does not satisfy %v: %v", tt.wantCause, err)
			}
			var parseErr *flags.ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error does not expose *flags.ParseError: %T", err)
			}
			if got := parseErr.Token(); got != tt.wantToken {
				t.Fatalf("Token() = %q, want %q", got, tt.wantToken)
			}
			if got := result.PathNames(); len(got) != 0 {
				t.Fatalf("failed Route returned non-zero path: %q", got)
			}
		})
	}
}

func TestRouteDistinguishesCommandTokensFromFlagSyntax(t *testing.T) {
	apply := mustDefinition(t, "apply")
	deploy := mustDefinition(t, "deploy", command.Children(apply))
	root := mustDefinition(
		t,
		"dib",
		command.InheritedFlags(flags.Bool("apply", false, "")),
		command.Children(deploy),
	)

	t.Run("command token matching flag name still routes as command", func(t *testing.T) {
		result, err := root.Route([]string{"deploy", "apply"})
		if err != nil {
			t.Fatalf("Route returned unexpected error: %v", err)
		}
		if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
			t.Fatalf("PathNames() = %q, want %q", got, []string{"dib", "deploy", "apply"})
		}
		snapshot, ok := result.FlagSnapshot()
		if !ok {
			t.Fatal("FlagSnapshot() returned ok=false")
		}
		assertFlagValues(t, snapshot, "apply", []any{false}, false, nil)
	})

	t.Run("registered flag-like token parses through flags package", func(t *testing.T) {
		result, err := root.Route([]string{"--apply", "deploy", "apply"})
		if err != nil {
			t.Fatalf("Route returned unexpected error: %v", err)
		}
		if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
			t.Fatalf("PathNames() = %q, want %q", got, []string{"dib", "deploy", "apply"})
		}
		snapshot, ok := result.FlagSnapshot()
		if !ok {
			t.Fatal("FlagSnapshot() returned ok=false")
		}
		assertFlagValues(t, snapshot, "apply", []any{true}, true, []string{"--apply"})
	})

	t.Run("unknown flag-like token fails as flag parse error", func(t *testing.T) {
		result, err := root.Route([]string{"--missing", "deploy", "apply"})
		if err == nil {
			t.Fatal("Route returned nil error")
		}
		if !errors.Is(err, flags.ErrUnknownFlag) {
			t.Fatalf("error does not satisfy flags.ErrUnknownFlag: %v", err)
		}
		var parseErr *flags.ParseError
		if !errors.As(err, &parseErr) {
			t.Fatalf("error does not expose *flags.ParseError: %T", err)
		}
		if got := parseErr.Token(); got != "--missing" {
			t.Fatalf("Token() = %q, want %q", got, "--missing")
		}
		if got := result.PathNames(); len(got) != 0 {
			t.Fatalf("failed Route returned non-zero path: %q", got)
		}
	})

	t.Run("flag-like positional after terminator remains unparsed", func(t *testing.T) {
		result, err := root.Route([]string{"deploy", "apply", "--", "--apply"})
		if err != nil {
			t.Fatalf("Route returned unexpected error: %v", err)
		}
		if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
			t.Fatalf("PathNames() = %q, want %q", got, []string{"dib", "deploy", "apply"})
		}
		if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"--apply"}) {
			t.Fatalf("RemainingArgs() = %q, want %q", got, []string{"--apply"})
		}
		snapshot, ok := result.FlagSnapshot()
		if !ok {
			t.Fatal("FlagSnapshot() returned ok=false")
		}
		assertFlagValues(t, snapshot, "apply", []any{false}, false, nil)
	})
}

func TestRouteFlagSnapshotsAreDefensiveAndReusable(t *testing.T) {
	root := mustFlagRoutingTree(t)
	args := []string{"deploy", "apply", "--tag", "one", "--tag", "two"}

	result, err := root.Route(args)
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	args[0] = "mutated"

	set, ok := result.Flags()
	if !ok {
		t.Fatal("Flags() returned ok=false")
	}
	defs := set.Definitions()
	defs[0] = flags.Bool("mutated", false, "")
	if got := flagNames(mustResultFlags(t, result).Definitions()); !reflect.DeepEqual(got, []string{"verbose", "cluster", "retries", "tag", "dry-run"}) {
		t.Fatalf("Flags() leaked mutable definitions: %q", got)
	}

	snapshot, ok := result.FlagSnapshot()
	if !ok {
		t.Fatal("FlagSnapshot() returned ok=false")
	}
	state, ok := snapshot.Lookup("tag")
	if !ok {
		t.Fatal("Lookup(tag) returned ok=false")
	}
	values := state.Values()
	values[0] = "mutated"
	assertFlagValues(t, snapshot, "tag", []any{"one", "two"}, true, []string{"--tag", "--tag"})

	again, err := root.Route([]string{"ship", "push", "-v", "--dry-run"})
	if err != nil {
		t.Fatalf("second Route returned unexpected error: %v", err)
	}
	if got := again.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("second PathNames() = %q, want %q", got, []string{"dib", "deploy", "apply"})
	}
	secondSnapshot, ok := again.FlagSnapshot()
	if !ok {
		t.Fatal("second FlagSnapshot() returned ok=false")
	}
	assertFlagValues(t, secondSnapshot, "verbose", []any{true}, true, []string{"-v"})
	assertFlagValues(t, secondSnapshot, "dry-run", []any{true}, true, []string{"--dry-run"})
}

func mustFlagRoutingTree(t *testing.T) command.Definition {
	t.Helper()

	apply := mustDefinition(
		t,
		"apply",
		command.Aliases("push"),
		command.LocalFlags(
			flags.Bool("dry-run", false, ""),
		),
	)
	plan := mustDefinition(t, "plan")
	deploy := mustDefinition(
		t,
		"deploy",
		command.Aliases("ship"),
		command.LocalFlags(flags.String("target", "", "")),
		command.InheritedFlags(
			flags.String("cluster", "", ""),
			flags.Int("retries", 0, ""),
			flags.StringList("tag", nil, "", flags.Repeatable()),
		),
		command.Children(apply, plan),
	)
	status := mustDefinition(t, "status")
	return mustDefinition(
		t,
		"dib",
		command.InheritedFlags(flags.Bool("verbose", false, "", flags.Shorthand("v"))),
		command.Children(deploy, status),
	)
}

func mustResultFlags(t *testing.T, result command.Result) flags.Set {
	t.Helper()

	set, ok := result.Flags()
	if !ok {
		t.Fatal("Flags() returned ok=false")
	}
	return set
}

func flagNames(definitions []flags.Definition) []string {
	names := make([]string, len(definitions))
	for i, definition := range definitions {
		names[i] = definition.Name()
	}
	return names
}

func assertFlagValues(t *testing.T, snapshot flags.Snapshot, name string, values []any, explicit bool, spellings []string) {
	t.Helper()

	state, ok := snapshot.Lookup(name)
	if !ok {
		t.Fatalf("Lookup(%q) returned ok=false", name)
	}
	if got := state.Values(); !reflect.DeepEqual(got, values) {
		t.Fatalf("Lookup(%q).Values() = %#v, want %#v", name, got, values)
	}
	if got := state.Explicit(); got != explicit {
		t.Fatalf("Lookup(%q).Explicit() = %v, want %v", name, got, explicit)
	}
	occurrences := state.Occurrences()
	if len(occurrences) != len(spellings) {
		t.Fatalf("Lookup(%q).Occurrences() length = %d, want %d", name, len(occurrences), len(spellings))
	}
	for i, occurrence := range occurrences {
		if got := occurrence.Spelling(); got != spellings[i] {
			t.Fatalf("Lookup(%q).Occurrences()[%d].Spelling() = %q, want %q", name, i, got, spellings[i])
		}
	}
}
