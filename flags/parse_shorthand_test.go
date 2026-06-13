package flags_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestParseShorthandValuesAndBooleanPresence(t *testing.T) {
	set, err := flags.NewSet(
		flags.String("name", "default", "name", flags.Shorthand("n")),
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		wantName   []any
		wantArgs   []string
		wantToken  string
		wantLookup string
	}{
		{
			name:       "separate value",
			args:       []string{"pos-a", "-n", "alice", "pos-b"},
			wantName:   []any{"alice"},
			wantArgs:   []string{"pos-a", "pos-b"},
			wantToken:  "-n",
			wantLookup: "name",
		},
		{
			name:       "equals value",
			args:       []string{"-n=bob", "pos"},
			wantName:   []any{"bob"},
			wantArgs:   []string{"pos"},
			wantToken:  "-n",
			wantLookup: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := set.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			state, ok := snapshot.Lookup("name")
			if !ok {
				t.Fatal("snapshot missing name state")
			}
			if got := state.Values(); !reflect.DeepEqual(got, tt.wantName) {
				t.Fatalf("name Values() = %#v, want %#v", got, tt.wantName)
			}
			if !state.Explicit() {
				t.Fatal("name state was not marked explicit")
			}
			occurrences := state.Occurrences()
			if len(occurrences) != 1 {
				t.Fatalf("name occurrence count = %d, want 1", len(occurrences))
			}
			if got := occurrences[0].Spelling(); got != tt.wantToken {
				t.Fatalf("occurrence Spelling() = %q, want %q", got, tt.wantToken)
			}
			if got := occurrences[0].NormalizedName(); got != tt.wantLookup {
				t.Fatalf("occurrence NormalizedName() = %q, want %q", got, tt.wantLookup)
			}
			if got := occurrences[0].Definition().Name(); got != "name" {
				t.Fatalf("occurrence Definition().Name() = %q, want name", got)
			}
			if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, tt.wantArgs) {
				t.Fatalf("RemainingArgs() = %#v, want %#v", got, tt.wantArgs)
			}

			verbose, ok := snapshot.Lookup("verbose")
			if !ok {
				t.Fatal("snapshot missing verbose state")
			}
			if verbose.Explicit() {
				t.Fatal("verbose default was incorrectly marked explicit")
			}
			if got := verbose.Values(); !reflect.DeepEqual(got, []any{false}) {
				t.Fatalf("verbose Values() = %#v, want default false", got)
			}
		})
	}

	snapshot, err := set.Parse([]string{"-v"})
	if err != nil {
		t.Fatalf("Parse(-v) returned unexpected error: %v", err)
	}
	verbose, ok := snapshot.Lookup("verbose")
	if !ok {
		t.Fatal("snapshot missing verbose state")
	}
	if got := verbose.Values(); !reflect.DeepEqual(got, []any{true}) {
		t.Fatalf("verbose Values() = %#v, want true", got)
	}
	if !verbose.Explicit() {
		t.Fatal("verbose state was not marked explicit")
	}
	occurrences := verbose.Occurrences()
	if len(occurrences) != 1 || occurrences[0].Spelling() != "-v" || occurrences[0].NormalizedName() != "verbose" || occurrences[0].Definition().Name() != "verbose" {
		t.Fatalf("verbose Occurrences() = %#v, want one canonical -v occurrence", occurrences)
	}
}

func TestParseShorthandDiagnosticsAreTyped(t *testing.T) {
	set, err := flags.NewSet(
		flags.Int("count", 0, "count", flags.Shorthand("n")),
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		wantCategory error
		wantToken    string
		wantName     string
		wantDef      string
		wantValueErr bool
	}{
		{
			name:         "unknown",
			args:         []string{"-x"},
			wantCategory: flags.ErrUnknownFlag,
			wantToken:    "-x",
			wantName:     "x",
		},
		{
			name:         "unknown with attached value",
			args:         []string{"-x=dib_fake_secret_value"},
			wantCategory: flags.ErrUnknownFlag,
			wantToken:    "-x",
			wantName:     "x",
		},
		{
			name:         "missing value",
			args:         []string{"-n"},
			wantCategory: flags.ErrMissingValue,
			wantToken:    "-n",
			wantName:     "n",
			wantDef:      "count",
		},
		{
			name:         "terminator is not value",
			args:         []string{"-n", "--"},
			wantCategory: flags.ErrMissingValue,
			wantToken:    "-n",
			wantName:     "n",
			wantDef:      "count",
		},
		{
			name:         "conversion",
			args:         []string{"-n=dib_fake_password_value"},
			wantCategory: flags.ErrConversion,
			wantToken:    "-n",
			wantName:     "n",
			wantDef:      "count",
			wantValueErr: true,
		},
		{
			name:         "duplicate across long and short",
			args:         []string{"--count=1", "-n=2"},
			wantCategory: flags.ErrDuplicateValue,
			wantToken:    "-n",
			wantName:     "n",
			wantDef:      "count",
		},
		{
			name:         "duplicate repeated short",
			args:         []string{"-n=1", "-n=2"},
			wantCategory: flags.ErrDuplicateValue,
			wantToken:    "-n",
			wantName:     "n",
			wantDef:      "count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := set.Parse(tt.args)
			if err == nil {
				t.Fatal("Parse returned nil error")
			}
			if !errors.Is(err, tt.wantCategory) {
				t.Fatalf("error does not satisfy %v: %v", tt.wantCategory, err)
			}
			var parseErr *flags.ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error does not expose *flags.ParseError: %T", err)
			}
			if !errors.Is(parseErr.Category(), tt.wantCategory) {
				t.Fatalf("ParseError.Category() = %v, want %v", parseErr.Category(), tt.wantCategory)
			}
			if parseErr.Token() != tt.wantToken || parseErr.Name() != tt.wantName {
				t.Fatalf("ParseError context token/name = %q/%q, want %q/%q", parseErr.Token(), parseErr.Name(), tt.wantToken, tt.wantName)
			}
			if strings.Contains(parseErr.Token(), "dib_fake_secret_value") || strings.Contains(err.Error(), "dib_fake_secret_value") {
				t.Fatalf("unknown shorthand diagnostic leaked attached raw value: %v", err)
			}
			def, ok := parseErr.Definition()
			if tt.wantDef == "" {
				if ok {
					t.Fatalf("ParseError.Definition() = (%q, true), want false", def.Name())
				}
			} else if !ok || def.Name() != tt.wantDef {
				t.Fatalf("ParseError.Definition() = (%q, %v), want %q true", def.Name(), ok, tt.wantDef)
			}
			if tt.wantValueErr {
				var valueErr *flags.ValueError
				if !errors.As(err, &valueErr) {
					t.Fatalf("error does not expose *flags.ValueError: %T", err)
				}
				if valueErr.Name() != tt.wantDef || valueErr.Kind() != flags.KindInt {
					t.Fatalf("ValueError context = name %q kind %v, want %s int", valueErr.Name(), valueErr.Kind(), tt.wantDef)
				}
				if strings.Contains(err.Error(), "dib_fake_password_value") {
					t.Fatalf("raw value leaked in parse error: %v", err)
				}
			}
		})
	}

	defaultSnapshot := set.DefaultSnapshot()
	count, ok := defaultSnapshot.Lookup("count")
	if !ok {
		t.Fatal("DefaultSnapshot missing count state")
	}
	if count.Explicit() {
		t.Fatal("failed parses mutated reusable set default snapshot")
	}
}

func TestParseShorthandRepeatableValuesAccumulate(t *testing.T) {
	set, err := flags.NewSet(flags.StringList("tag", nil, "tags", flags.Shorthand("t"), flags.Repeatable()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"-t=one", "-t", "two", "pos"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	state, ok := snapshot.Lookup("tag")
	if !ok {
		t.Fatal("snapshot missing tag state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{"one", "two"}) {
		t.Fatalf("tag Values() = %#v, want accumulated short values", got)
	}
	occurrences := state.Occurrences()
	if len(occurrences) != 2 || occurrences[0].Spelling() != "-t" || occurrences[1].Spelling() != "-t" {
		t.Fatalf("tag Occurrences() = %#v, want two -t occurrences", occurrences)
	}
	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"pos"}) {
		t.Fatalf("RemainingArgs() = %#v, want pos", got)
	}
}

func TestParseShorthandFailedParseDoesNotAffectLaterParse(t *testing.T) {
	set, err := flags.NewSet(flags.Int("count", 0, "count", flags.Shorthand("n")))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"-n=1", "-n=2"})
	if err == nil {
		t.Fatal("Parse returned nil error for duplicate shorthand")
	}
	if !errors.Is(err, flags.ErrDuplicateValue) {
		t.Fatalf("duplicate shorthand error does not satisfy ErrDuplicateValue: %v", err)
	}

	snapshot, err := set.Parse([]string{"-n=3", "pos"})
	if err != nil {
		t.Fatalf("Parse after failed parse returned unexpected error: %v", err)
	}
	state, ok := snapshot.Lookup("count")
	if !ok {
		t.Fatal("snapshot missing count state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{3}) {
		t.Fatalf("count Values() = %#v, want 3", got)
	}
	if !state.Explicit() {
		t.Fatal("count state was not marked explicit")
	}
	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"pos"}) {
		t.Fatalf("RemainingArgs() = %#v, want pos", got)
	}
}

func TestParseShorthandLookupIsIndependentFromLongNameNormalization(t *testing.T) {
	normalizer := flags.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})
	set, err := flags.NewNormalizedSet(
		normalizer,
		flags.String("log-level", "info", "log level", flags.Shorthand("l")),
		flags.String("letter", "default", "letter"),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"-l=debug"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	state, ok := snapshot.Lookup("log-level")
	if !ok {
		t.Fatal("snapshot missing log-level state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{"debug"}) {
		t.Fatalf("log-level Values() = %#v, want debug", got)
	}
	occurrences := state.Occurrences()
	if len(occurrences) != 1 || occurrences[0].NormalizedName() != "log-level" || occurrences[0].Definition().Name() != "log-level" {
		t.Fatalf("log-level Occurrences() = %#v, want canonical log-level lookup", occurrences)
	}

	_, err = set.Parse([]string{"-log_level=debug"})
	if err == nil {
		t.Fatal("Parse returned nil error for multi-rune shorthand-looking token")
	}
	if !errors.Is(err, flags.ErrUnknownFlag) {
		t.Fatalf("multi-rune shorthand-looking error does not satisfy ErrUnknownFlag: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("multi-rune shorthand-looking error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Name() != "log_level" || parseErr.NormalizedName() != "" {
		t.Fatalf("multi-rune shorthand-looking context name/normalized = %q/%q, want log_level/empty", parseErr.Name(), parseErr.NormalizedName())
	}
}

func TestParseShorthandSensitiveValuesDoNotLeakInErrors(t *testing.T) {
	parserErr := errors.New("parser rejected dib_fake_token_value")
	set, err := flags.NewSet(flags.Custom("token", flags.KindString, "", "token", flags.ParserFunc(func(raw string) (any, error) {
		return nil, parserErr
	}), flags.Shorthand("t"), flags.Sensitive()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"-t", "dib_fake_token_value"})
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
		t.Fatalf("raw value leaked in parse error: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Token() != "-t" || parseErr.Name() != "t" {
		t.Fatalf("ParseError context token/name = %q/%q, want -t/t", parseErr.Token(), parseErr.Name())
	}
}

// TestParseShorthandNoOptionDefault verifies that a short flag with NoOptionDefault uses the
// configured fallback when no explicit value follows: at end of args or before --.
// Short-flag parsing uses stopBeforeLong=false, so a following long-flag token is consumed
// as the value rather than triggering the no-option default.
func TestParseShorthandNoOptionDefault(t *testing.T) {
	set, err := flags.NewSet(
		flags.String("output", "stdout", "output", flags.Shorthand("o"), flags.NoOptionDefault("file.out")),
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	cases := []struct {
		name      string
		args      []string
		wantValue any
		wantArgs  []string
	}{
		{
			name:      "short flag at end of args applies no-option default",
			args:      []string{"-o"},
			wantValue: "file.out",
			wantArgs:  nil,
		},
		{
			name:      "short flag before double-dash applies no-option default",
			args:      []string{"-o", "--", "passthrough"},
			wantValue: "file.out",
			wantArgs:  []string{"passthrough"},
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
		})
	}
}

func TestParseShorthandTerminatorProtectsLaterTokens(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose", flags.Shorthand("v")))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"--", "-v", "pos"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	verbose, ok := snapshot.Lookup("verbose")
	if !ok {
		t.Fatal("snapshot missing verbose state")
	}
	if verbose.Explicit() {
		t.Fatal("verbose was marked explicit after terminator")
	}
	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"-v", "pos"}) {
		t.Fatalf("RemainingArgs() = %#v, want terminator tail", got)
	}
}
