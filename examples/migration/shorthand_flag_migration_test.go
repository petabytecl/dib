package migration_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func Example_shorthandFlagMigration() {
	set, err := flags.NewSet(
		flags.Bool("all", false, "select all", flags.Shorthand("a")),
		flags.Bool("verbose", false, "verbose output", flags.Shorthand("v")),
		flags.Int("level", 0, "verbosity level", flags.Shorthand("l"), flags.NoOptionDefault(1)),
		flags.String("output", "stdout", "output path", flags.Shorthand("o"), flags.NoOptionDefault("stdout")),
		flags.StringList("tag", nil, "tag", flags.Shorthand("t"), flags.Repeatable()),
	)
	if err != nil {
		panic(err)
	}

	snapshot, err := set.Parse([]string{
		"--output=result.txt",
		"-alv",
		"--tag", "alpha",
		"-t", "beta",
		"input.txt",
		"--",
		"--literal",
	})
	if err != nil {
		panic(err)
	}

	level, _ := snapshot.Lookup("level")
	tags, _ := snapshot.Lookup("tag")
	fmt.Println(level.Values()[0], tags.Values())
	fmt.Println(snapshot.RemainingArgs())
	// Output:
	// 1 [alpha beta]
	// [input.txt --literal]
}

func TestShorthandFlagMigrationCoversPflagStyleConcepts(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("all", false, "select all", flags.Shorthand("a")),
		flags.Bool("verbose", false, "verbose output", flags.Shorthand("v")),
		flags.Int("level", 0, "level", flags.Shorthand("l"), flags.NoOptionDefault(1)),
		flags.String("output", "stdout", "output path", flags.Shorthand("o"), flags.NoOptionDefault("stdout")),
		flags.StringList("tag", nil, "tag", flags.Shorthand("t"), flags.Repeatable()),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	snapshot, err := set.Parse([]string{
		"before",
		"--output", "report.txt",
		"-alv",
		"--tag=alpha",
		"-t", "beta",
		"after",
		"--",
		"--not-a-flag",
		"-x",
	})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	assertFlagValues(t, snapshot, "all", []any{true}, true)
	assertFlagValues(t, snapshot, "verbose", []any{true}, true)
	assertFlagValues(t, snapshot, "level", []any{1}, true)
	assertFlagValues(t, snapshot, "output", []any{"report.txt"}, true)
	assertFlagValues(t, snapshot, "tag", []any{"alpha", "beta"}, true)
	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"before", "after", "--not-a-flag", "-x"}) {
		t.Fatalf("RemainingArgs() = %#v, want interspersed positionals and passthrough", got)
	}
}

func TestShorthandFlagMigrationIntentionalDifferences(t *testing.T) {
	set, err := flags.NewNormalizedSet(
		func(name string) string { return strings.ReplaceAll(name, "_", "-") },
		flags.Bool("dry-run", false, "preview", flags.Shorthand("d")),
		flags.Bool("no-cache", false, "disable cache"),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		wantCategory error
	}{
		{
			name:         "undefined no-prefix long name is ordinary",
			args:         []string{"--no-color"},
			wantCategory: flags.ErrUnknownFlag,
		},
		{
			name:         "unregistered long help is caller controlled",
			args:         []string{"--help"},
			wantCategory: flags.ErrHelpRequest,
		},
		{
			name:         "unregistered short help is caller controlled",
			args:         []string{"-h"},
			wantCategory: flags.ErrHelpRequest,
		},
		{
			name:         "shorthand lookup is not long-name normalization",
			args:         []string{"-D"},
			wantCategory: flags.ErrUnknownFlag,
		},
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
			if _, ok := snapshot.Lookup("dry-run"); ok {
				t.Fatal("failed parse returned non-zero snapshot")
			}
		})
	}

	snapshot, err := set.Parse([]string{"--dry_run", "--no-cache"})
	if err != nil {
		t.Fatalf("Parse valid normalized long and defined no-* flag: %v", err)
	}
	assertFlagValues(t, snapshot, "dry-run", []any{true}, true)
	assertFlagValues(t, snapshot, "no-cache", []any{true}, true)

	shorthandSnapshot, err := set.Parse([]string{"-d"})
	if err != nil {
		t.Fatalf("Parse valid shorthand: %v", err)
	}
	assertFlagValues(t, shorthandSnapshot, "dry-run", []any{true}, true)
}
