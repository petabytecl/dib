package flags_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/flags"
)

// TestParseBoundaryInterspersedPositionals verifies positional args accumulate in relative order
// regardless of how flags and positionals are interleaved.
func TestParseBoundaryInterspersedPositionals(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose"),
		flags.String("name", "default", "name"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	cases := []struct {
		name          string
		args          []string
		wantRemaining []string
	}{
		{
			name:          "positionals only",
			args:          []string{"pos-a", "pos-b", "pos-c"},
			wantRemaining: []string{"pos-a", "pos-b", "pos-c"},
		},
		{
			name:          "flags only no positionals",
			args:          []string{"--verbose"},
			wantRemaining: nil,
		},
		{
			name:          "positional then flag then positional",
			args:          []string{"pos-a", "--verbose", "pos-b"},
			wantRemaining: []string{"pos-a", "pos-b"},
		},
		{
			name:          "multiple positionals between flags",
			args:          []string{"--verbose", "pos-1", "pos-2", "--name", "foo", "pos-3"},
			wantRemaining: []string{"pos-1", "pos-2", "pos-3"},
		},
		{
			name:          "interleaved flags and positionals keep relative order",
			args:          []string{"first", "--verbose", "second", "third"},
			wantRemaining: []string{"first", "second", "third"},
		},
		{
			name:          "no args produces nil remaining",
			args:          []string{},
			wantRemaining: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := set.Parse(tc.args)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, tc.wantRemaining) {
				t.Fatalf("RemainingArgs() = %#v, want %#v", got, tc.wantRemaining)
			}
		})
	}
}

// TestParseBoundaryTerminator verifies -- stops flag parsing and preserves all subsequent tokens.
func TestParseBoundaryTerminator(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	cases := []struct {
		name          string
		args          []string
		wantRemaining []string
	}{
		{
			name:          "double-dash alone",
			args:          []string{"--"},
			wantRemaining: nil,
		},
		{
			name:          "double-dash followed by flag-like tokens",
			args:          []string{"--", "--verbose", "-v"},
			wantRemaining: []string{"--verbose", "-v"},
		},
		{
			name:          "double-dash followed by ordinary strings",
			args:          []string{"--", "a", "b", "c"},
			wantRemaining: []string{"a", "b", "c"},
		},
		{
			name:          "double-dash as first arg",
			args:          []string{"--", "first", "second"},
			wantRemaining: []string{"first", "second"},
		},
		{
			name:          "double-dash after parsed flags",
			args:          []string{"--verbose", "--", "extra"},
			wantRemaining: []string{"extra"},
		},
		{
			name:          "flags before double-dash and flag-like after",
			args:          []string{"--verbose", "--", "--unknown", "pos-b"},
			wantRemaining: []string{"--unknown", "pos-b"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := set.Parse(tc.args)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, tc.wantRemaining) {
				t.Fatalf("RemainingArgs() = %#v, want %#v", got, tc.wantRemaining)
			}
		})
	}
}

// TestParseBoundaryPassthrough verifies args after -- that look like invalid flags are preserved verbatim.
func TestParseBoundaryPassthrough(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose"),
		flags.String("name", "", "name"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	cases := []struct {
		name          string
		args          []string
		wantRemaining []string
	}{
		{
			name:          "unknown flag after double-dash appears verbatim",
			args:          []string{"--", "--unknown-flag"},
			wantRemaining: []string{"--unknown-flag"},
		},
		{
			name:          "missing-value-like arg after double-dash appears verbatim",
			args:          []string{"--", "--name"},
			wantRemaining: []string{"--name"},
		},
		{
			name:          "mixed ordinary and flag-like after double-dash",
			args:          []string{"--", "arg1", "--flag", "-x", "arg2"},
			wantRemaining: []string{"arg1", "--flag", "-x", "arg2"},
		},
		{
			name:          "second double-dash after first double-dash appears verbatim",
			args:          []string{"--", "--", "-"},
			wantRemaining: []string{"--", "-"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := set.Parse(tc.args)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, tc.wantRemaining) {
				t.Fatalf("RemainingArgs() = %#v, want %#v", got, tc.wantRemaining)
			}
		})
	}
}

// TestParseBoundaryFailedParses verifies every parse failure category returns a typed error
// and a zero-value Snapshot so callers cannot accidentally use partial state.
func TestParseBoundaryFailedParses(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
		flags.String("name", "", "name", flags.Shorthand("n")),
		flags.Int("count", 0, "count"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	cases := []struct {
		name         string
		args         []string
		wantCategory error
	}{
		{
			name:         "unknown long flag",
			args:         []string{"--unknown"},
			wantCategory: flags.ErrUnknownFlag,
		},
		{
			name:         "unknown shorthand",
			args:         []string{"-z"},
			wantCategory: flags.ErrUnknownFlag,
		},
		{
			name:         "missing value for long flag",
			args:         []string{"--name"},
			wantCategory: flags.ErrMissingValue,
		},
		{
			name:         "missing value for short flag",
			args:         []string{"-n"},
			wantCategory: flags.ErrMissingValue,
		},
		{
			name:         "invalid value conversion",
			args:         []string{"--count=notanint"},
			wantCategory: flags.ErrConversion,
		},
		{
			name:         "invalid group non-final required member",
			args:         []string{"-nv"},
			wantCategory: flags.ErrInvalidGroup,
		},
		{
			name:         "duplicate single-value long flag",
			args:         []string{"--name=foo", "--name=bar"},
			wantCategory: flags.ErrDuplicateValue,
		},
		{
			name:         "duplicate across long and short spellings",
			args:         []string{"--name=foo", "-n", "bar"},
			wantCategory: flags.ErrDuplicateValue,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, parseErr := set.Parse(tc.args)
			if parseErr == nil {
				t.Fatal("Parse returned nil error, expected error")
			}

			if !errors.Is(parseErr, tc.wantCategory) {
				t.Errorf("errors.Is(err, %v) = false; got %v", tc.wantCategory, parseErr)
			}

			var pe *flags.ParseError
			if !errors.As(parseErr, &pe) {
				t.Fatalf("errors.As(err, **ParseError) = false; got %T: %v", parseErr, parseErr)
			}

			// Zero-value snapshot must be returned on parse failure.
			if _, ok := snapshot.Lookup("verbose"); ok {
				t.Error("expected zero-value snapshot on parse error, but Lookup found state for 'verbose'")
			}
			if _, ok := snapshot.Lookup("name"); ok {
				t.Error("expected zero-value snapshot on parse error, but Lookup found state for 'name'")
			}
			if got := snapshot.RemainingArgs(); got != nil {
				t.Errorf("expected nil RemainingArgs on parse error, got %v", got)
			}
		})
	}
}

// TestParseBoundaryDeterministicSnapshot verifies simultaneous assertions on a complex parse
// that combines interspersed positionals, flags, and a -- terminator.
func TestParseBoundaryDeterministicSnapshot(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
		flags.String("name", "default", "name", flags.Shorthand("n")),
		flags.Int("count", 1, "count"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	args := []string{"pos-1", "--verbose", "--name", "alice", "pos-2", "--count=5", "--", "--flag-like", "pos-3"}
	snapshot, err := set.Parse(args)
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	verbose, ok := snapshot.Lookup("verbose")
	if !ok {
		t.Fatal("snapshot missing verbose state")
	}
	if got := verbose.Values(); !reflect.DeepEqual(got, []any{true}) {
		t.Fatalf("verbose Values() = %#v, want [true]", got)
	}
	if !verbose.Explicit() {
		t.Fatal("verbose Explicit() = false, want true")
	}
	if occ := verbose.Occurrences(); len(occ) != 1 || occ[0].Spelling() != "--verbose" {
		t.Fatalf("verbose Occurrences() = %#v, want one --verbose", occ)
	}

	name, ok := snapshot.Lookup("name")
	if !ok {
		t.Fatal("snapshot missing name state")
	}
	if got := name.Values(); !reflect.DeepEqual(got, []any{"alice"}) {
		t.Fatalf("name Values() = %#v, want [alice]", got)
	}
	if !name.Explicit() {
		t.Fatal("name Explicit() = false, want true")
	}

	count, ok := snapshot.Lookup("count")
	if !ok {
		t.Fatal("snapshot missing count state")
	}
	if got := count.Values(); !reflect.DeepEqual(got, []any{5}) {
		t.Fatalf("count Values() = %#v, want [5]", got)
	}
	if !count.Explicit() {
		t.Fatal("count Explicit() = false, want true")
	}
	if occ := count.Occurrences(); len(occ) != 1 || occ[0].Spelling() != "--count" {
		t.Fatalf("count Occurrences() = %#v, want one --count", occ)
	}

	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"pos-1", "pos-2", "--flag-like", "pos-3"}) {
		t.Fatalf("RemainingArgs() = %#v, want interspersed positionals plus terminator tail", got)
	}
}

// TestParseHelpLongUnregistered verifies --help produces ErrHelpRequest when no help flag is defined.
func TestParseHelpLongUnregistered(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	_, parseErr := set.Parse([]string{"--help"})
	if parseErr == nil {
		t.Fatal("Parse returned nil error for unregistered --help, want ErrHelpRequest")
	}
	if !errors.Is(parseErr, flags.ErrHelpRequest) {
		t.Fatalf("errors.Is(err, ErrHelpRequest) = false; got %v", parseErr)
	}

	var pe *flags.ParseError
	if !errors.As(parseErr, &pe) {
		t.Fatalf("errors.As(err, **ParseError) = false; got %T", parseErr)
	}
	if pe.Token() != "--help" {
		t.Errorf("ParseError.Token() = %q, want --help", pe.Token())
	}
	if pe.Name() != "help" {
		t.Errorf("ParseError.Name() = %q, want help", pe.Name())
	}
	if pe.Category() != flags.ErrHelpRequest {
		t.Errorf("ParseError.Category() = %v, want ErrHelpRequest", pe.Category())
	}
}

// TestParseHelpShortUnregistered verifies -h produces ErrHelpRequest when no h shorthand is defined.
func TestParseHelpShortUnregistered(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	_, parseErr := set.Parse([]string{"-h"})
	if parseErr == nil {
		t.Fatal("Parse returned nil error for unregistered -h, want ErrHelpRequest")
	}
	if !errors.Is(parseErr, flags.ErrHelpRequest) {
		t.Fatalf("errors.Is(err, ErrHelpRequest) = false; got %v", parseErr)
	}

	var pe *flags.ParseError
	if !errors.As(parseErr, &pe) {
		t.Fatalf("errors.As(err, **ParseError) = false; got %T", parseErr)
	}
	if pe.Token() != "--help" {
		t.Errorf("ParseError.Token() = %q, want --help (normalized from -h)", pe.Token())
	}
	if pe.Name() != "help" {
		t.Errorf("ParseError.Name() = %q, want help", pe.Name())
	}
	if pe.Category() != flags.ErrHelpRequest {
		t.Errorf("ParseError.Category() = %v, want ErrHelpRequest", pe.Category())
	}
}

// TestParseHelpLongRegistered verifies --help parses normally when a help flag is registered.
func TestParseHelpLongRegistered(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("help", false, "show help"),
		flags.Bool("verbose", false, "verbose"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	snapshot, err := set.Parse([]string{"--help"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error for registered --help: %v", err)
	}
	state, ok := snapshot.Lookup("help")
	if !ok {
		t.Fatal("snapshot missing help state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{true}) {
		t.Fatalf("help Values() = %#v, want [true]", got)
	}
	if !state.Explicit() {
		t.Fatal("help Explicit() = false when registered --help is parsed, want true")
	}
}

// TestParseHelpShortRegistered verifies -h parses normally when h shorthand is registered.
func TestParseHelpShortRegistered(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose", flags.Shorthand("h")),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	snapshot, err := set.Parse([]string{"-h"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error for registered -h shorthand: %v", err)
	}
	state, ok := snapshot.Lookup("verbose")
	if !ok {
		t.Fatal("snapshot missing verbose state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{true}) {
		t.Fatalf("verbose Values() via -h = %#v, want [true]", got)
	}
}

// TestParseHelpAfterTerminator verifies --help after -- is treated as a passthrough arg, not a help request.
func TestParseHelpAfterTerminator(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	snapshot, err := set.Parse([]string{"--", "--help"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"--help"}) {
		t.Fatalf("RemainingArgs() = %#v, want [--help] as passthrough", got)
	}
}

// TestParseHelpRequestZeroValueSnapshot verifies that --help and -h (unregistered) return a zero-value
// Snapshot so callers cannot accidentally use partial state after a help request.
func TestParseHelpRequestZeroValueSnapshot(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	cases := []struct {
		name string
		args []string
	}{
		{"--help unregistered", []string{"--help"}},
		{"-h unregistered", []string{"-h"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, parseErr := set.Parse(tc.args)
			if parseErr == nil {
				t.Fatal("Parse returned nil error, expected ErrHelpRequest")
			}
			if _, ok := snapshot.Lookup("verbose"); ok {
				t.Error("zero-value snapshot should have no Lookup state after help request")
			}
			if got := snapshot.RemainingArgs(); got != nil {
				t.Errorf("zero-value snapshot RemainingArgs() = %v, want nil", got)
			}
		})
	}
}

// TestParseHelpRequestNormalizedName verifies that ParseError.NormalizedName() exposes the correct
// normalized name for both --help (long) and -h (short) help requests.
func TestParseHelpRequestNormalizedName(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	cases := []struct {
		name           string
		args           []string
		wantNormalized string
	}{
		{"--help long form", []string{"--help"}, "help"},
		{"-h short form", []string{"-h"}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, parseErr := set.Parse(tc.args)
			var pe *flags.ParseError
			if !errors.As(parseErr, &pe) {
				t.Fatalf("errors.As(err, **ParseError) = false; got %T: %v", parseErr, parseErr)
			}
			if got := pe.NormalizedName(); got != tc.wantNormalized {
				t.Errorf("ParseError.NormalizedName() = %q, want %q", got, tc.wantNormalized)
			}
		})
	}
}

// TestParseHelpRequestDefinitionAccessor verifies that ParseError.Definition() returns (_, false)
// for help requests, since no matching definition exists in the set.
func TestParseHelpRequestDefinitionAccessor(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"--help", []string{"--help"}},
		{"-h", []string{"-h"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, parseErr := set.Parse(tc.args)
			var pe *flags.ParseError
			if !errors.As(parseErr, &pe) {
				t.Fatalf("errors.As(err, **ParseError) = false; got %T: %v", parseErr, parseErr)
			}
			if _, hasDef := pe.Definition(); hasDef {
				t.Errorf("ParseError.Definition() hasDefinition = true for help request, want false")
			}
		})
	}
}

// TestParseHelpLongAttachedValueUnregistered verifies that --help=anything triggers ErrHelpRequest
// when no help flag is registered, matching the behavior of --help without an attached value.
func TestParseHelpLongAttachedValueUnregistered(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	_, parseErr := set.Parse([]string{"--help=anything"})
	if parseErr == nil {
		t.Fatal("Parse returned nil error for --help=anything with no registered help flag")
	}
	if !errors.Is(parseErr, flags.ErrHelpRequest) {
		t.Fatalf("errors.Is(err, ErrHelpRequest) = false; got %v", parseErr)
	}

	var pe *flags.ParseError
	if !errors.As(parseErr, &pe) {
		t.Fatalf("errors.As(err, **ParseError) = false; got %T", parseErr)
	}
	if pe.Name() != "help" {
		t.Errorf("ParseError.Name() = %q, want help", pe.Name())
	}
}

// TestParseHelpShortInGroupDoesNotTriggerHelpRequest verifies that -h appearing as a member
// of a shorthand group (e.g., -vh) returns ErrUnknownFlag, not ErrHelpRequest.
// Help detection applies only when -h is the sole token; grouped shorthands take the group path.
func TestParseHelpShortInGroupDoesNotTriggerHelpRequest(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	_, parseErr := set.Parse([]string{"-vh"})
	if parseErr == nil {
		t.Fatal("Parse returned nil error for -vh with unregistered h, want error")
	}
	if errors.Is(parseErr, flags.ErrHelpRequest) {
		t.Error("errors.Is(err, ErrHelpRequest) = true for -vh in group, want false; grouped -h is not a help request")
	}
	if !errors.Is(parseErr, flags.ErrUnknownFlag) {
		t.Errorf("expected ErrUnknownFlag for grouped -h; got %v", parseErr)
	}
}

// TestSetReusableAfterHelpRequest verifies that the Set can be used for a normal successful parse
// after it previously returned ErrHelpRequest.
func TestSetReusableAfterHelpRequest(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose"),
		flags.String("name", "default", "name"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	_, helpErr := set.Parse([]string{"--help"})
	if !errors.Is(helpErr, flags.ErrHelpRequest) {
		t.Fatalf("expected ErrHelpRequest, got: %v", helpErr)
	}

	snapshot, err := set.Parse([]string{"--verbose", "--name", "alice"})
	if err != nil {
		t.Fatalf("Parse after help request returned unexpected error: %v", err)
	}
	if state, ok := snapshot.Lookup("verbose"); !ok || !state.Explicit() {
		t.Error("verbose not found or not explicit in post-help-request parse")
	}
	if state, ok := snapshot.Lookup("name"); !ok || !reflect.DeepEqual(state.Values(), []any{"alice"}) {
		t.Errorf("name values incorrect in post-help-request parse: %v", snapshot)
	}
}
