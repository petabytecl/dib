package migration_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func Example_standardFlagConcepts() {
	set, err := flags.NewSet(
		flags.String("name", "guest", "name to greet"),
		flags.Int("count", 1, "number of greetings"),
		flags.Bool("verbose", false, "show details"),
	)
	if err != nil {
		panic(err)
	}

	snapshot, err := set.Parse([]string{"--name", "Ada", "--count=2", "--verbose", "input.txt"})
	if err != nil {
		panic(err)
	}

	name, _ := snapshot.Lookup("name")
	count, _ := snapshot.Lookup("count")
	verbose, _ := snapshot.Lookup("verbose")

	fmt.Println(name.Values()[0], count.Values()[0], verbose.Explicit())
	fmt.Println(snapshot.RemainingArgs())
	// Output:
	// Ada 2 true
	// [input.txt]
}

func TestStandardFlagConceptsUseExplicitInputsAndTypedFailures(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	os.Args = []string{"ambient", "--name", "wrong"}
	t.Cleanup(func() { os.Args = originalArgs })

	set, err := flags.NewSet(
		flags.String("name", "guest", "name"),
		flags.Int("count", 1, "count"),
		flags.Bool("verbose", false, "verbose"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	tests := []struct {
		name          string
		args          []string
		wantName      []any
		wantCount     []any
		wantVerbose   []any
		wantRemaining []string
	}{
		{
			name:          "explicit args override defaults",
			args:          []string{"--name=Ada", "--count", "3", "--verbose", "file.txt"},
			wantName:      []any{"Ada"},
			wantCount:     []any{3},
			wantVerbose:   []any{true},
			wantRemaining: []string{"file.txt"},
		},
		{
			name:          "defaults are captured without process globals",
			args:          nil,
			wantName:      []any{"guest"},
			wantCount:     []any{1},
			wantVerbose:   []any{false},
			wantRemaining: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := set.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			assertFlagValues(t, snapshot, "name", tt.wantName, len(tt.args) > 0 && tt.wantName[0] != "guest")
			assertFlagValues(t, snapshot, "count", tt.wantCount, len(tt.args) > 0 && tt.wantCount[0] != 1)
			assertFlagValues(t, snapshot, "verbose", tt.wantVerbose, len(tt.args) > 0 && tt.wantVerbose[0] == true)
			if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, tt.wantRemaining) {
				t.Fatalf("RemainingArgs() = %#v, want %#v", got, tt.wantRemaining)
			}
		})
	}
}

func TestStandardFlagConceptsInspectParseErrors(t *testing.T) {
	set, err := flags.NewSet(
		flags.String("name", "guest", "name"),
		flags.Int("count", 1, "count"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		wantCategory error
		wantToken    string
	}{
		{name: "unknown flag", args: []string{"--missing"}, wantCategory: flags.ErrUnknownFlag, wantToken: "--missing"},
		{name: "missing value", args: []string{"--name"}, wantCategory: flags.ErrMissingValue, wantToken: "--name"},
		{name: "conversion failure", args: []string{"--count", "many"}, wantCategory: flags.ErrConversion, wantToken: "--count"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := set.Parse(tt.args)
			if err == nil {
				t.Fatal("Parse returned nil error")
			}
			if !errors.Is(err, tt.wantCategory) {
				t.Fatalf("error %v does not satisfy %v", err, tt.wantCategory)
			}
			var parseErr *flags.ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error does not expose *flags.ParseError: %T", err)
			}
			if got := parseErr.Token(); got != tt.wantToken {
				t.Fatalf("Token() = %q, want %q", got, tt.wantToken)
			}
			if got := snapshot.RemainingArgs(); got != nil {
				t.Fatalf("failed parse returned non-zero remaining args: %#v", got)
			}
			if _, ok := snapshot.Lookup("name"); ok {
				t.Fatal("failed parse returned non-zero snapshot state")
			}
		})
	}
}

func assertFlagValues(t *testing.T, snapshot flags.Snapshot, name string, want []any, wantExplicit bool) {
	t.Helper()
	state, ok := snapshot.Lookup(name)
	if !ok {
		t.Fatalf("Lookup(%q) returned ok=false", name)
	}
	if got := state.Values(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Lookup(%q).Values() = %#v, want %#v", name, got, want)
	}
	if got := state.Explicit(); got != wantExplicit {
		t.Fatalf("Lookup(%q).Explicit() = %t, want %t", name, got, wantExplicit)
	}
}
