package flags_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestParseShortGroupBooleanMembers(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("all", false, "all", flags.Shorthand("a")),
		flags.Bool("brief", false, "brief", flags.Shorthand("b")),
		flags.Bool("color", false, "color", flags.Shorthand("c")),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"pos-a", "-abc", "pos-b"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	for _, tc := range []struct {
		name      string
		shorthand string
	}{
		{name: "all", shorthand: "-a"},
		{name: "brief", shorthand: "-b"},
		{name: "color", shorthand: "-c"},
	} {
		state, ok := snapshot.Lookup(tc.name)
		if !ok {
			t.Fatalf("snapshot missing %s state", tc.name)
		}
		if got := state.Values(); !reflect.DeepEqual(got, []any{true}) {
			t.Fatalf("%s Values() = %#v, want true", tc.name, got)
		}
		if !state.Explicit() {
			t.Fatalf("%s was not marked explicit", tc.name)
		}
		occurrences := state.Occurrences()
		if len(occurrences) != 1 {
			t.Fatalf("%s occurrence count = %d, want 1", tc.name, len(occurrences))
		}
		if got := occurrences[0].Spelling(); got != tc.shorthand {
			t.Fatalf("%s occurrence Spelling() = %q, want %q", tc.name, got, tc.shorthand)
		}
		if got := occurrences[0].NormalizedName(); got != tc.name {
			t.Fatalf("%s occurrence NormalizedName() = %q, want %q", tc.name, got, tc.name)
		}
		if got := occurrences[0].Definition().Name(); got != tc.name {
			t.Fatalf("%s occurrence Definition().Name() = %q, want %q", tc.name, got, tc.name)
		}
	}
	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"pos-a", "pos-b"}) {
		t.Fatalf("RemainingArgs() = %#v, want positionals", got)
	}
}

func TestParseShortGroupFinalRequiredValue(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("all", false, "all", flags.Shorthand("a")),
		flags.Int("count", 0, "count", flags.Shorthand("n")),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		wantArgs     []string
		wantCount    []any
		wantSpelling string
	}{
		{
			name:         "attached",
			args:         []string{"-an10", "pos"},
			wantArgs:     []string{"pos"},
			wantCount:    []any{10},
			wantSpelling: "-n",
		},
		{
			name:         "separate",
			args:         []string{"-an", "10", "pos"},
			wantArgs:     []string{"pos"},
			wantCount:    []any{10},
			wantSpelling: "-n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := set.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			all, ok := snapshot.Lookup("all")
			if !ok || !all.Explicit() || !reflect.DeepEqual(all.Values(), []any{true}) {
				t.Fatalf("all state = %#v/%v, want explicit true", all, ok)
			}
			count, ok := snapshot.Lookup("count")
			if !ok {
				t.Fatal("snapshot missing count state")
			}
			if got := count.Values(); !reflect.DeepEqual(got, tt.wantCount) {
				t.Fatalf("count Values() = %#v, want %#v", got, tt.wantCount)
			}
			if !count.Explicit() {
				t.Fatal("count was not marked explicit")
			}
			occurrences := count.Occurrences()
			if len(occurrences) != 1 || occurrences[0].Spelling() != tt.wantSpelling || occurrences[0].NormalizedName() != "count" {
				t.Fatalf("count Occurrences() = %#v, want canonical -n occurrence", occurrences)
			}
			if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, tt.wantArgs) {
				t.Fatalf("RemainingArgs() = %#v, want %#v", got, tt.wantArgs)
			}
		})
	}
}

func TestParseShortGroupNonFinalNoOptionDefault(t *testing.T) {
	set, err := flags.NewSet(
		flags.Int("level", 0, "level", flags.Shorthand("l"), flags.NoOptionDefault(7)),
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"-lv"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	level, ok := snapshot.Lookup("level")
	if !ok {
		t.Fatal("snapshot missing level state")
	}
	if got := level.Default(); got != 0 {
		t.Fatalf("level Default() = %#v, want configured default 0", got)
	}
	if got := level.Values(); !reflect.DeepEqual(got, []any{7}) {
		t.Fatalf("level Values() = %#v, want no-option value 7", got)
	}
	if !level.Explicit() {
		t.Fatal("level no-option value was not marked explicit")
	}
	occurrences := level.Occurrences()
	if len(occurrences) != 1 || occurrences[0].Spelling() != "-l" || occurrences[0].NormalizedName() != "level" {
		t.Fatalf("level Occurrences() = %#v, want one -l occurrence", occurrences)
	}

	verbose, ok := snapshot.Lookup("verbose")
	if !ok || !verbose.Explicit() || !reflect.DeepEqual(verbose.Values(), []any{true}) {
		t.Fatalf("verbose state = %#v/%v, want explicit true", verbose, ok)
	}
}

func TestParseShortGroupFinalNoOptionDefaultWithoutExplicitValue(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("all", false, "all", flags.Shorthand("a")),
		flags.Int("level", 0, "level", flags.Shorthand("l"), flags.NoOptionDefault(7)),
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantArgs []string
	}{
		{
			name:     "end of args",
			args:     []string{"-al"},
			wantArgs: nil,
		},
		{
			name:     "terminator",
			args:     []string{"-al", "--", "-v"},
			wantArgs: []string{"-v"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := set.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			all, ok := snapshot.Lookup("all")
			if !ok || !all.Explicit() || !reflect.DeepEqual(all.Values(), []any{true}) {
				t.Fatalf("all state = %#v/%v, want explicit true", all, ok)
			}
			level, ok := snapshot.Lookup("level")
			if !ok {
				t.Fatal("snapshot missing level state")
			}
			if got := level.Values(); !reflect.DeepEqual(got, []any{7}) {
				t.Fatalf("level Values() = %#v, want no-option value 7", got)
			}
			if !level.Explicit() {
				t.Fatal("level no-option value was not marked explicit")
			}
			occurrences := level.Occurrences()
			if len(occurrences) != 1 || occurrences[0].Spelling() != "-l" || occurrences[0].NormalizedName() != "level" {
				t.Fatalf("level Occurrences() = %#v, want one -l occurrence", occurrences)
			}
			verbose, ok := snapshot.Lookup("verbose")
			if !ok || verbose.Explicit() {
				t.Fatalf("verbose state = %#v/%v, want default after terminator", verbose, ok)
			}
			if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, tt.wantArgs) {
				t.Fatalf("RemainingArgs() = %#v, want %#v", got, tt.wantArgs)
			}
		})
	}
}

func TestParseShortGroupDiagnosticsAreTyped(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("all", false, "all", flags.Shorthand("a")),
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
			name:         "unknown member",
			args:         []string{"-ax"},
			wantCategory: flags.ErrUnknownFlag,
			wantToken:    "-ax",
			wantName:     "x",
		},
		{
			name:         "invalid non-final required member",
			args:         []string{"-nva"},
			wantCategory: flags.ErrInvalidGroup,
			wantToken:    "-n",
			wantName:     "n",
			wantDef:      "count",
		},
		{
			name:         "missing final value",
			args:         []string{"-an"},
			wantCategory: flags.ErrMissingValue,
			wantToken:    "-an",
			wantName:     "n",
			wantDef:      "count",
		},
		{
			name:         "conversion",
			args:         []string{"-andib_fake_password_value"},
			wantCategory: flags.ErrConversion,
			wantToken:    "-an",
			wantName:     "n",
			wantDef:      "count",
			wantValueErr: true,
		},
		{
			name:         "duplicate within group",
			args:         []string{"-vv"},
			wantCategory: flags.ErrDuplicateValue,
			wantToken:    "-vv",
			wantName:     "v",
			wantDef:      "verbose",
		},
		{
			name:         "duplicate across long and group",
			args:         []string{"--verbose", "-av"},
			wantCategory: flags.ErrDuplicateValue,
			wantToken:    "-av",
			wantName:     "v",
			wantDef:      "verbose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := set.Parse(tt.args)
			if err == nil {
				t.Fatal("Parse returned nil error")
			}
			if snapshot.RemainingArgs() != nil {
				t.Fatalf("failed parse returned partial remaining args: %#v", snapshot.RemainingArgs())
			}
			if !errors.Is(err, tt.wantCategory) {
				t.Fatalf("error does not satisfy %v: %v", tt.wantCategory, err)
			}
			var parseErr *flags.ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error does not expose *flags.ParseError: %T", err)
			}
			if parseErr.Category() != tt.wantCategory {
				t.Fatalf("ParseError.Category() = %v, want %v", parseErr.Category(), tt.wantCategory)
			}
			if parseErr.Token() != tt.wantToken || parseErr.Name() != tt.wantName {
				t.Fatalf("ParseError context token/name = %q/%q, want %q/%q", parseErr.Token(), parseErr.Name(), tt.wantToken, tt.wantName)
			}
			if strings.Contains(parseErr.Token(), "dib_fake_password_value") || strings.Contains(err.Error(), "dib_fake_password_value") {
				t.Fatalf("group diagnostic leaked raw attached value: %v", err)
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
			}
		})
	}
}

func TestParseShortGroupTerminatorProtectsLaterTokens(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("all", false, "all", flags.Shorthand("a")),
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"--", "-av", "pos"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	all, ok := snapshot.Lookup("all")
	if !ok || all.Explicit() {
		t.Fatalf("all state = %#v/%v, want default after terminator", all, ok)
	}
	verbose, ok := snapshot.Lookup("verbose")
	if !ok || verbose.Explicit() {
		t.Fatalf("verbose state = %#v/%v, want default after terminator", verbose, ok)
	}
	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"-av", "pos"}) {
		t.Fatalf("RemainingArgs() = %#v, want terminator tail", got)
	}
}

func TestParseShortGroupRepeatableValuesAccumulate(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("all", false, "all", flags.Shorthand("a")),
		flags.StringList("tag", nil, "tags", flags.Shorthand("t"), flags.Repeatable()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"-atone", "-t", "two", "pos"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	all, ok := snapshot.Lookup("all")
	if !ok || !all.Explicit() || !reflect.DeepEqual(all.Values(), []any{true}) {
		t.Fatalf("all state = %#v/%v, want explicit true", all, ok)
	}
	tag, ok := snapshot.Lookup("tag")
	if !ok {
		t.Fatal("snapshot missing tag state")
	}
	if got := tag.Values(); !reflect.DeepEqual(got, []any{"one", "two"}) {
		t.Fatalf("tag Values() = %#v, want accumulated grouped and standalone values", got)
	}
	occurrences := tag.Occurrences()
	if len(occurrences) != 2 || occurrences[0].Spelling() != "-t" || occurrences[1].Spelling() != "-t" {
		t.Fatalf("tag Occurrences() = %#v, want two canonical -t occurrences", occurrences)
	}
	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"pos"}) {
		t.Fatalf("RemainingArgs() = %#v, want pos", got)
	}
}

func TestParseShortGroupKeepsShorthandIndependentFromLongNormalization(t *testing.T) {
	normalizer := flags.NameNormalizer(func(name string) string {
		return strings.ReplaceAll(name, "_", "-")
	})
	set, err := flags.NewNormalizedSet(
		normalizer,
		flags.Bool("log-level", false, "log level", flags.Shorthand("l")),
		flags.Bool("verbose", false, "verbose", flags.Shorthand("v")),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"-lv", "--log_level"})
	if err == nil {
		t.Fatal("Parse returned nil error for duplicate normalized long name after group")
	}
	if !errors.Is(err, flags.ErrDuplicateValue) {
		t.Fatalf("error does not satisfy ErrDuplicateValue: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Token() != "--log_level" || parseErr.Name() != "log_level" || parseErr.NormalizedName() != "log-level" {
		t.Fatalf("ParseError context = token %q name %q normalized %q, want normalized long context", parseErr.Token(), parseErr.Name(), parseErr.NormalizedName())
	}
	if _, ok := snapshot.Lookup("log-level"); ok {
		t.Fatalf("failed parse exposed partial grouped snapshot: %#v", snapshot)
	}

	snapshot, err = set.Parse([]string{"-log_level"})
	if err == nil {
		t.Fatal("Parse returned nil error for shorthand-looking normalized long name")
	}
	if !errors.Is(err, flags.ErrUnknownFlag) {
		t.Fatalf("error does not satisfy ErrUnknownFlag: %v", err)
	}
	if !errors.As(err, &parseErr) {
		t.Fatalf("error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Name() != "o" || parseErr.NormalizedName() != "" {
		t.Fatalf("group member context name/normalized = %q/%q, want o/empty", parseErr.Name(), parseErr.NormalizedName())
	}
	if _, ok := snapshot.Lookup("log-level"); ok {
		t.Fatalf("failed shorthand-looking parse exposed partial snapshot: %#v", snapshot)
	}
}

func FuzzParseShortGroups(f *testing.F) {
	set, err := flags.NewSet(
		flags.Bool("all", false, "all", flags.Shorthand("a")),
		flags.Bool("brief", false, "brief", flags.Shorthand("b")),
		flags.Int("count", 0, "count", flags.Shorthand("n")),
		flags.String("name", "", "name", flags.Shorthand("s"), flags.NoOptionDefault("default")),
	)
	if err != nil {
		f.Fatalf("NewSet returned unexpected error: %v", err)
	}

	for _, seed := range []string{"-ab", "-an10", "-an", "-sba", "-nbs", "-ax", "-andib_fake_token_value"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, token string) {
		_, err1 := set.Parse([]string{token, "10", "pos"})
		_, err2 := set.Parse([]string{token, "10", "pos"})
		if (err1 == nil) != (err2 == nil) {
			t.Fatalf("Parse determinism mismatch for %q: %v vs %v", token, err1, err2)
		}
		if err1 == nil {
			return
		}
		if err1.Error() != err2.Error() {
			t.Fatalf("Parse diagnostic changed for %q: %q vs %q", token, err1.Error(), err2.Error())
		}
		var parseErr *flags.ParseError
		if !errors.As(err1, &parseErr) {
			t.Fatalf("error does not expose *flags.ParseError for %q: %T", token, err1)
		}
	})
}
