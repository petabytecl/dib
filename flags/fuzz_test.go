package flags_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

// FuzzParse verifies broad parser invariants against arbitrary multi-token inputs.
// It covers long flags, short flags, shorthand groups, repeatable values, no-option
// defaults, normalization spellings, positionals, --, help requests, and error diagnostics.
// Seeds are independently written clean-room inputs; no seeds are copied from Go flag,
// pflag, Cobra, Viper, or other parser projects.
func FuzzParse(f *testing.F) {
	normalizer := flags.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})
	set, err := flags.NewNormalizedSet(
		normalizer,
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
		flags.String("name", "default", "name", flags.Shorthand("n")),
		flags.Int("count", 0, "count", flags.Shorthand("c")),
		flags.String("log-level", "info", "log level"),
		flags.String("output", "stdout", "output", flags.Shorthand("o"), flags.NoOptionDefault("file.out")),
		flags.StringList("tag", nil, "tags", flags.Repeatable(), flags.Shorthand("t")),
		flags.Int("secret-count", 0, "secret count", flags.Sensitive()),
	)
	if err != nil {
		f.Fatalf("NewSet: %v", err)
	}
	wantDefinitions := set.Definitions()

	// Each seed is a newline-separated list of CLI args.
	f.Add("")                                      // empty input
	f.Add("--verbose")                             // boolean long flag
	f.Add("--name=alice")                          // attached string value
	f.Add("--name\nalice")                         // separate string value
	f.Add("--count=42")                            // attached int value
	f.Add("--output")                              // long flag uses no-option default
	f.Add("--tag=a\n--tag=b\n--tag=c")             // repeatable long flags
	f.Add("-v")                                    // boolean short flag
	f.Add("-n=bob")                                // attached short value
	f.Add("-n\nbob")                               // separate short value
	f.Add("-o")                                    // short flag uses no-option default
	f.Add("-t\nalpha\n-t\nbeta")                   // repeatable short flags
	f.Add("-vn=bob")                               // boolean group then attached value
	f.Add("-vn\nbob")                              // boolean group then separate value
	f.Add("pos-a\n--verbose\npos-b")               // interspersed positionals
	f.Add("--verbose\n--\n--unknown\npos")         // double-dash passthrough
	f.Add("--help")                                // unregistered --help
	f.Add("-h")                                    // unregistered -h
	f.Add("--unknown")                             // unknown long flag
	f.Add("--count=notanint")                      // conversion failure
	f.Add("--name=x\n--name=y")                    // duplicate single-value flag
	f.Add("--secret-count\ndib_fake_secret_value") // sensitive conversion value
	f.Add("--log_level=debug")                     // normalized long spelling

	f.Fuzz(func(t *testing.T, raw string) {
		var args []string
		if raw != "" {
			args = strings.Split(raw, "\n")
		}

		snapshot1, err1 := set.Parse(args)

		// Invariant: no panic (guaranteed by reaching this line).
		assertDefinitionsUnchanged(t, set, wantDefinitions)

		// Invariant: failed parse returns zero-value snapshot.
		if err1 != nil {
			if _, ok := snapshot1.Lookup("verbose"); ok {
				t.Error("failed parse exposed non-zero Lookup state for 'verbose'")
			}
			if snapshot1.RemainingArgs() != nil {
				t.Errorf("failed parse returned non-nil RemainingArgs: %v", snapshot1.RemainingArgs())
			}
			// Invariant: parse failure is a typed *flags.ParseError.
			var pe *flags.ParseError
			if !errors.As(err1, &pe) {
				t.Fatalf("parse error is not *flags.ParseError: %T: %v", err1, err1)
			}
			if hasSensitiveSecretCountValue(args) {
				if strings.Contains(pe.Token(), "dib_fake_secret_value") || strings.Contains(err1.Error(), "dib_fake_secret_value") {
					t.Errorf("sensitive conversion leaked raw value: token=%q err=%v", pe.Token(), err1)
				}
			}
		}

		// Invariant: deterministic — re-parse of the same args must agree on success/failure.
		_, err2 := set.Parse(args)
		if (err1 == nil) != (err2 == nil) {
			t.Errorf("re-parse produced different error outcome: first=%v, second=%v", err1, err2)
		}
		assertDefinitionsUnchanged(t, set, wantDefinitions)

		// Invariant: all-positional input (no arg starts with -) must succeed.
		allPositional := true
		for _, a := range args {
			if strings.HasPrefix(a, "-") {
				allPositional = false
				break
			}
		}
		if allPositional && err1 != nil {
			t.Errorf("all-positional args produced unexpected error: %v", err1)
		}

		// Invariant: successful parse RemainingArgs is defensively copied.
		if err1 == nil {
			ra1 := snapshot1.RemainingArgs()
			if ra1 != nil {
				ra1[0] = "mutated"
				if got := snapshot1.RemainingArgs(); len(got) > 0 && got[0] == "mutated" {
					t.Error("RemainingArgs exposed mutable storage after successful parse")
				}
			}

			for _, def := range wantDefinitions {
				state, ok := snapshot1.Lookup(def.Name())
				if !ok {
					t.Fatalf("successful parse snapshot missing %q state", def.Name())
				}

				values := state.Values()
				if len(values) > 0 {
					values[0] = "mutated"
					got, _ := snapshot1.Lookup(def.Name())
					if gotValues := got.Values(); len(gotValues) > 0 && gotValues[0] == "mutated" {
						t.Fatalf("%s Values exposed mutable storage after successful parse", def.Name())
					}
				}

				occurrences := state.Occurrences()
				if len(occurrences) > 0 {
					occurrences[0] = flags.ValueOccurrence{}
					got, _ := snapshot1.Lookup(def.Name())
					if gotOccurrences := got.Occurrences(); len(gotOccurrences) > 0 && gotOccurrences[0].Spelling() == "" {
						t.Fatalf("%s Occurrences exposed mutable storage after successful parse", def.Name())
					}
				}
			}
		}
	})
}

func assertDefinitionsUnchanged(t *testing.T, set flags.Set, want []flags.Definition) {
	t.Helper()

	got := set.Definitions()
	if len(got) != len(want) {
		t.Fatalf("Set definitions length changed: got %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i].Name() != want[i].Name() ||
			got[i].Kind() != want[i].Kind() ||
			got[i].Usage() != want[i].Usage() ||
			got[i].Hidden() != want[i].Hidden() ||
			got[i].Sensitive() != want[i].Sensitive() ||
			got[i].Deprecated() != want[i].Deprecated() ||
			got[i].RepeatPolicy() != want[i].RepeatPolicy() ||
			got[i].Arity() != want[i].Arity() {
			t.Fatalf("Set definition %d metadata changed: got %#v, want %#v", i, got[i], want[i])
		}
		gotShort, gotOK := got[i].Shorthand()
		wantShort, wantOK := want[i].Shorthand()
		if gotShort != wantShort || gotOK != wantOK {
			t.Fatalf("Set definition %d shorthand changed: got %q/%v, want %q/%v", i, gotShort, gotOK, wantShort, wantOK)
		}
		if gotDefault, wantDefault := got[i].Default(), want[i].Default(); !reflect.DeepEqual(gotDefault, wantDefault) {
			t.Fatalf("Set definition %d default changed: got %#v, want %#v", i, gotDefault, wantDefault)
		}
		gotNoOption, gotNoOptionOK := got[i].NoOptionDefault()
		wantNoOption, wantNoOptionOK := want[i].NoOptionDefault()
		if gotNoOptionOK != wantNoOptionOK || !reflect.DeepEqual(gotNoOption, wantNoOption) {
			t.Fatalf("Set definition %d no-option default changed: got %#v/%v, want %#v/%v",
				i, gotNoOption, gotNoOptionOK, wantNoOption, wantNoOptionOK)
		}
	}
}

func hasSensitiveSecretCountValue(args []string) bool {
	for i, arg := range args {
		if arg == "--secret-count" && i+1 < len(args) && args[i+1] == "dib_fake_secret_value" {
			return true
		}
		if strings.HasPrefix(arg, "--secret-count=") && strings.TrimPrefix(arg, "--secret-count=") == "dib_fake_secret_value" {
			return true
		}
	}
	return false
}

// FuzzParseBoundary verifies parse boundary invariants against arbitrary inputs:
// no panic, RemainingArgs non-nil when args follow --, all-positional input succeeds,
// and the Set is reusable across parse runs.
func FuzzParseBoundary(f *testing.F) {
	set, err := flags.NewSet(
		flags.Bool("verbose", false, "verbose"),
		flags.String("name", "", "name"),
	)
	if err != nil {
		f.Fatalf("NewSet: %v", err)
	}

	// Each seed is a newline-separated list of CLI args.
	f.Add("")                                     // empty string slice
	f.Add("--")                                   // -- alone
	f.Add("--\n--verbose\n-v")                    // -- followed by flag-like args
	f.Add("pos-a\n--verbose\npos-b\n--name\nfoo") // interspersed positionals and flags
	f.Add("--help")                               // --help alone (unregistered)
	f.Add("--name")                               // flag followed by missing value
	f.Add("--name=attached")                      // flag with attached value

	f.Fuzz(func(t *testing.T, raw string) {
		var args []string
		if raw != "" {
			args = strings.Split(raw, "\n")
		}

		snapshot, err := set.Parse(args)

		if err == nil {
			// When args follow --, RemainingArgs must not be nil.
			for i, a := range args {
				if a == "--" && i+1 < len(args) {
					if snapshot.RemainingArgs() == nil {
						t.Error("RemainingArgs returned nil when args follow --")
					}
					break
				}
			}
		}

		// Parse must succeed when every arg is a positional (no leading -).
		allPositional := true
		for _, a := range args {
			if strings.HasPrefix(a, "-") {
				allPositional = false
				break
			}
		}
		if allPositional && err != nil {
			t.Errorf("all-positional args produced unexpected error: %v", err)
		}

		// Set is reusable: a second parse of the same args must agree on success/failure.
		_, err2 := set.Parse(args)
		if (err == nil) != (err2 == nil) {
			t.Errorf("re-parse produced different error outcome: first=%v, second=%v", err, err2)
		}
	})
}
