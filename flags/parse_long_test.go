package flags_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestParseLongFlagsCoversContractMatrix(t *testing.T) {
	set, err := flags.NewSet(
		flags.String("empty", "default", "empty"),
		flags.Bool("verbose", false, "verbose"),
		flags.StringList("tag", nil, "tags", flags.Repeatable()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"--empty=", "pos-a", "--verbose=false", "--tag=one", "--tag", "two", "--", "--unknown", "pos-b"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	empty, ok := snapshot.Lookup("empty")
	if !ok {
		t.Fatal("snapshot missing empty state")
	}
	if got := empty.Values(); !reflect.DeepEqual(got, []any{""}) {
		t.Fatalf("empty Values() = %#v, want attached empty string", got)
	}

	verbose, ok := snapshot.Lookup("verbose")
	if !ok {
		t.Fatal("snapshot missing verbose state")
	}
	if got := verbose.Values(); !reflect.DeepEqual(got, []any{false}) {
		t.Fatalf("verbose Values() = %#v, want false", got)
	}

	tag, ok := snapshot.Lookup("tag")
	if !ok {
		t.Fatal("snapshot missing tag state")
	}
	if got := tag.Values(); !reflect.DeepEqual(got, []any{"one", "two"}) {
		t.Fatalf("tag Values() = %#v, want accumulated values", got)
	}
	if occurrences := tag.Occurrences(); len(occurrences) != 2 || occurrences[0].Spelling() != "--tag" || occurrences[1].Spelling() != "--tag" {
		t.Fatalf("tag Occurrences() = %#v, want two --tag occurrences", occurrences)
	}

	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"pos-a", "--unknown", "pos-b"}) {
		t.Fatalf("RemainingArgs() = %#v, want positionals plus terminator tail", got)
	}
}

func TestParseLongSnapshotAccessorsAreDefensive(t *testing.T) {
	set, err := flags.NewSet(flags.StringList("tag", []string{"default"}, "tags", flags.Repeatable()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"--tag=one", "pos"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	remaining := snapshot.RemainingArgs()
	remaining[0] = "mutated"
	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"pos"}) {
		t.Fatalf("RemainingArgs exposed mutable storage: %#v", got)
	}

	state, _ := snapshot.Lookup("tag")
	values := state.Values()
	values[0] = "mutated"
	freshState, _ := snapshot.Lookup("tag")
	if got := freshState.Values(); !reflect.DeepEqual(got, []any{"one"}) {
		t.Fatalf("Values exposed mutable storage: %#v", got)
	}

	occurrences := state.Occurrences()
	occurrences[0] = flags.ValueOccurrence{}
	freshOccurrences := state.Occurrences()
	if len(freshOccurrences) != 1 || freshOccurrences[0].Spelling() != "--tag" {
		t.Fatalf("Occurrences exposed mutable storage: %#v", freshOccurrences)
	}
}

func TestParseLongErrorsDoNotEchoRawValues(t *testing.T) {
	set, err := flags.NewSet(flags.Int("workers", 1, "workers"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--workers=dib_fake_password_value"})
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("error does not satisfy ErrConversion: %v", err)
	}
	if strings.Contains(err.Error(), "dib_fake_password_value") {
		t.Fatalf("raw value leaked in parse error: %v", err)
	}
}

func TestParseLongBooleanExplicitTrue(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"--verbose=true"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	state, ok := snapshot.Lookup("verbose")
	if !ok {
		t.Fatal("snapshot missing verbose state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{true}) {
		t.Fatalf("verbose Values() = %#v, want explicit true", got)
	}
	if !state.Explicit() {
		t.Fatal("verbose state was not marked explicit")
	}
	occurrences := state.Occurrences()
	if len(occurrences) != 1 || occurrences[0].Spelling() != "--verbose" || occurrences[0].Definition().Name() != "verbose" {
		t.Fatalf("verbose Occurrences() = %#v, want one canonical --verbose occurrence", occurrences)
	}
}

func TestParseLongNormalizedOccurrenceExposesLookupKey(t *testing.T) {
	normalizer := flags.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})
	set, err := flags.NewNormalizedSet(normalizer, flags.String("log-level", "info", "log level"))
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"--log_level=debug"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	state, ok := snapshot.Lookup("log-level")
	if !ok {
		t.Fatal("snapshot missing log-level state")
	}
	occurrences := state.Occurrences()
	if len(occurrences) != 1 {
		t.Fatalf("log-level occurrence count = %d, want 1", len(occurrences))
	}
	if got := occurrences[0].NormalizedName(); got != "log-level" {
		t.Fatalf("occurrence NormalizedName() = %q, want log-level", got)
	}
	if got := occurrences[0].Spelling(); got != "--log_level" {
		t.Fatalf("occurrence Spelling() = %q, want --log_level", got)
	}
}

func TestParseLongDuplicateSingleValuePrecedesConversion(t *testing.T) {
	set, err := flags.NewSet(flags.Int("workers", 1, "workers"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--workers=2", "--workers=not-an-int"})
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrDuplicateValue) {
		t.Fatalf("duplicate second occurrence error does not satisfy ErrDuplicateValue: %v", err)
	}
	if errors.Is(err, flags.ErrConversion) {
		t.Fatalf("duplicate second occurrence unexpectedly exposed conversion: %v", err)
	}
}

func TestParseLongDuplicateNormalizedNameKeepsRawContext(t *testing.T) {
	normalizer := flags.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})
	set, err := flags.NewNormalizedSet(normalizer, flags.String("log-level", "info", "log level"))
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--log-level=debug", "--log_level=trace"})
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrDuplicateValue) {
		t.Fatalf("duplicate normalized error does not satisfy ErrDuplicateValue: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("duplicate normalized error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Token() != "--log_level" || parseErr.Name() != "log_level" {
		t.Fatalf("duplicate normalized context token/name = %q/%q, want --log_level/log_level", parseErr.Token(), parseErr.Name())
	}
	if parseErr.NormalizedName() != "log-level" {
		t.Fatalf("duplicate normalized NormalizedName() = %q, want log-level", parseErr.NormalizedName())
	}
	def, ok := parseErr.Definition()
	if !ok || def.Name() != "log-level" {
		t.Fatalf("duplicate normalized Definition() = (%q, %v), want log-level true", def.Name(), ok)
	}
}

// TestParseLongNoOptionDefault verifies that a long flag with NoOptionDefault uses the
// configured fallback when no explicit value follows: at end of args, before --, or before
// another long-flag token (long-flag parsing uses stopBeforeLong=true to avoid consuming
// a flag token as a value).
func TestParseLongNoOptionDefault(t *testing.T) {
	set, err := flags.NewSet(
		flags.String("output", "stdout", "output", flags.NoOptionDefault("file.out")),
		flags.Bool("verbose", false, "verbose"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	cases := []struct {
		name        string
		args        []string
		wantValue   any
		wantArgs    []string
		wantVerbose bool
	}{
		{
			name:        "long flag at end of args applies no-option default",
			args:        []string{"--output"},
			wantValue:   "file.out",
			wantArgs:    nil,
			wantVerbose: false,
		},
		{
			name:        "long flag before double-dash applies no-option default",
			args:        []string{"--output", "--", "passthrough"},
			wantValue:   "file.out",
			wantArgs:    []string{"passthrough"},
			wantVerbose: false,
		},
		{
			name:        "long flag before another long-flag token applies no-option default",
			args:        []string{"--output", "--verbose"},
			wantValue:   "file.out",
			wantArgs:    nil,
			wantVerbose: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := set.Parse(tc.args)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			state, ok := snapshot.Lookup("output")
			if !ok {
				t.Fatal("snapshot missing output state")
			}
			if got := state.Values(); !reflect.DeepEqual(got, []any{tc.wantValue}) {
				t.Fatalf("output Values() = %#v, want %#v", got, []any{tc.wantValue})
			}
			if !state.Explicit() {
				t.Fatal("output Explicit() = false; no-option default should mark the flag explicit")
			}
			if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, tc.wantArgs) {
				t.Fatalf("RemainingArgs() = %#v, want %#v", got, tc.wantArgs)
			}
			verbose, ok := snapshot.Lookup("verbose")
			if !ok {
				t.Fatal("snapshot missing verbose state")
			}
			if got := verbose.Explicit(); got != tc.wantVerbose {
				t.Fatalf("verbose Explicit() = %v, want %v", got, tc.wantVerbose)
			}
		})
	}
}

func TestParseLongSeparateSensitiveValuesDoNotLeakInErrors(t *testing.T) {
	parserErr := errors.New("parser rejected dib_fake_token_value")
	set, err := flags.NewSet(flags.Custom("token", flags.KindString, "", "token", flags.ParserFunc(func(raw string) (any, error) {
		return nil, parserErr
	}), flags.Sensitive()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--token", "dib_fake_token_value"})
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("error does not satisfy ErrConversion: %v", err)
	}
	if errors.Is(err, parserErr) {
		t.Fatalf("sensitive parser cause leaked through parse error: %v", err)
	}
	if strings.Contains(err.Error(), "dib_fake_token_value") {
		t.Fatalf("separate raw value leaked in parse error: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Token() != "--token" || strings.Contains(parseErr.Token(), "dib_fake_token_value") {
		t.Fatalf("ParseError.Token() = %q, want redacted --token", parseErr.Token())
	}
}
