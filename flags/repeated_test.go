package flags_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/petabytecl/dib/flags"
)

func TestParseRepeatableValuesAccumulateAcrossSpellings(t *testing.T) {
	set, err := flags.NewSet(
		flags.Bool("all", false, "all", flags.Shorthand("a")),
		flags.Bool("brief", false, "brief", flags.Shorthand("b")),
		flags.StringList("tag", nil, "tags", flags.Shorthand("t"), flags.Repeatable()),
		flags.Int("level", 0, "level", flags.Shorthand("l"), flags.NoOptionDefault(7), flags.Repeatable()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"--tag=one", "-t", "two", "-atomega", "pos", "--level=9", "-bl"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	tag, ok := snapshot.Lookup("tag")
	if !ok {
		t.Fatal("snapshot missing tag state")
	}
	if got := tag.Values(); !reflect.DeepEqual(got, []any{"one", "two", "omega"}) {
		t.Fatalf("tag Values() = %#v, want command-line ordered accumulation", got)
	}
	assertOccurrences(t, tag, []string{"--tag", "-t", "-t"}, "tag")

	level, ok := snapshot.Lookup("level")
	if !ok {
		t.Fatal("snapshot missing level state")
	}
	if got := level.Values(); !reflect.DeepEqual(got, []any{9, 7}) {
		t.Fatalf("level Values() = %#v, want explicit then no-option accumulation", got)
	}
	assertOccurrences(t, level, []string{"--level", "-l"}, "level")

	if got := snapshot.RemainingArgs(); !reflect.DeepEqual(got, []string{"pos"}) {
		t.Fatalf("RemainingArgs() = %#v, want preserved positional", got)
	}
}

func TestParseRepeatableBuiltInValuesAccumulateByKind(t *testing.T) {
	tests := []struct {
		name string
		def  flags.Definition
		args []string
		want []any
	}{
		{
			name: "string",
			def:  flags.String("config", "", "config", flags.Repeatable()),
			args: []string{"--config=dev.json", "--config=prod.json"},
			want: []any{"dev.json", "prod.json"},
		},
		{
			name: "bool",
			def:  flags.Bool("verbose", false, "verbose", flags.Repeatable()),
			args: []string{"--verbose", "--verbose=false"},
			want: []any{true, false},
		},
		{
			name: "int64",
			def:  flags.Int64("limit", 0, "limit", flags.Repeatable()),
			args: []string{"--limit=64", "--limit=128"},
			want: []any{int64(64), int64(128)},
		},
		{
			name: "uint",
			def:  flags.Uint("retries", 0, "retries", flags.Repeatable()),
			args: []string{"--retries=3", "--retries=5"},
			want: []any{uint(3), uint(5)},
		},
		{
			name: "uint64",
			def:  flags.Uint64("bytes", 0, "bytes", flags.Repeatable()),
			args: []string{"--bytes=1024", "--bytes=2048"},
			want: []any{uint64(1024), uint64(2048)},
		},
		{
			name: "float64",
			def:  flags.Float64("ratio", 0, "ratio", flags.Repeatable()),
			args: []string{"--ratio=0.5", "--ratio=0.75"},
			want: []any{0.5, 0.75},
		},
		{
			name: "duration",
			def:  flags.Duration("timeout", 0, "timeout", flags.Repeatable()),
			args: []string{"--timeout=5s", "--timeout=1m"},
			want: []any{5 * time.Second, time.Minute},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, err := flags.NewSet(tt.def)
			if err != nil {
				t.Fatalf("NewSet returned unexpected error: %v", err)
			}

			snapshot, err := set.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			state, ok := snapshot.Lookup(tt.def.Name())
			if !ok {
				t.Fatalf("snapshot missing %s state", tt.def.Name())
			}
			if got := state.Values(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("%s Values() = %#v, want accumulated values %#v", tt.def.Name(), got, tt.want)
			}
			assertOccurrences(t, state, []string{"--" + tt.def.Name(), "--" + tt.def.Name()}, tt.def.Name())
		})
	}
}

func TestParseRepeatableCustomValuesAppendOnePerOccurrence(t *testing.T) {
	set, err := flags.NewSet(flags.Custom("label", flags.KindString, "", "label", flags.ParserFunc(func(raw string) (any, error) {
		return strings.ToUpper(raw), nil
	}), flags.Repeatable()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"--label=one", "--label=two"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	state, ok := snapshot.Lookup("label")
	if !ok {
		t.Fatal("snapshot missing label state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{"ONE", "TWO"}) {
		t.Fatalf("label Values() = %#v, want one parsed custom value per occurrence", got)
	}
	assertOccurrences(t, state, []string{"--label", "--label"}, "label")
}

func TestParseCustomStringListFlattensValuesAndRecordsOccurrences(t *testing.T) {
	set, err := flags.NewSet(flags.Custom("tag", flags.KindStringList, []string{}, "tags", flags.ParserFunc(func(raw string) (any, error) {
		if raw == "" {
			return []string{}, nil
		}
		return strings.Split(raw, ","), nil
	}), flags.Repeatable()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"--tag=alpha,beta", "--tag=gamma"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	state, ok := snapshot.Lookup("tag")
	if !ok {
		t.Fatal("snapshot missing tag state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{"alpha", "beta", "gamma"}) {
		t.Fatalf("tag Values() = %#v, want flattened effective values", got)
	}
	assertOccurrences(t, state, []string{"--tag", "--tag"}, "tag")
}

func TestParseDuplicateSingleValuePrecedesSecondConversionAcrossSpellings(t *testing.T) {
	normalizer := flags.NameNormalizer(func(name string) string {
		return strings.ReplaceAll(name, "_", "-")
	})

	tests := []struct {
		name           string
		set            flags.Set
		args           []string
		wantToken      string
		wantName       string
		wantNormalized string
		wantDefinition string
	}{
		{
			name:           "long",
			set:            mustNewSet(t, flags.Int("workers", 1, "workers", flags.Shorthand("w"))),
			args:           []string{"--workers=2", "--workers=not-an-int"},
			wantToken:      "--workers",
			wantName:       "workers",
			wantNormalized: "workers",
			wantDefinition: "workers",
		},
		{
			name:           "short",
			set:            mustNewSet(t, flags.Int("workers", 1, "workers", flags.Shorthand("w"))),
			args:           []string{"-w=2", "-w=not-an-int"},
			wantToken:      "-w",
			wantName:       "w",
			wantNormalized: "workers",
			wantDefinition: "workers",
		},
		{
			name:           "grouped final shorthand",
			set:            mustNewSet(t, flags.Bool("all", false, "all", flags.Shorthand("a")), flags.Int("workers", 1, "workers", flags.Shorthand("w"))),
			args:           []string{"--workers=2", "-awnot-an-int"},
			wantToken:      "-aw",
			wantName:       "w",
			wantNormalized: "workers",
			wantDefinition: "workers",
		},
		{
			name:           "normalized long",
			set:            mustNewNormalizedSet(t, normalizer, flags.Int("worker-count", 1, "workers")),
			args:           []string{"--worker-count=2", "--worker_count=not-an-int"},
			wantToken:      "--worker_count",
			wantName:       "worker_count",
			wantNormalized: "worker-count",
			wantDefinition: "worker-count",
		},
		{
			name:           "mixed short then long",
			set:            mustNewSet(t, flags.Int("workers", 1, "workers", flags.Shorthand("w"))),
			args:           []string{"-w=2", "--workers=not-an-int"},
			wantToken:      "--workers",
			wantName:       "workers",
			wantNormalized: "workers",
			wantDefinition: "workers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.set.Parse(tt.args)
			if err == nil {
				t.Fatal("Parse returned nil error")
			}
			if !errors.Is(err, flags.ErrDuplicateValue) {
				t.Fatalf("error does not satisfy ErrDuplicateValue: %v", err)
			}
			if errors.Is(err, flags.ErrConversion) {
				t.Fatalf("duplicate second occurrence unexpectedly exposed conversion: %v", err)
			}

			var parseErr *flags.ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error does not expose *flags.ParseError: %T", err)
			}
			if parseErr.Token() != tt.wantToken || parseErr.Name() != tt.wantName || parseErr.NormalizedName() != tt.wantNormalized {
				t.Fatalf("ParseError context token/name/normalized = %q/%q/%q, want %q/%q/%q",
					parseErr.Token(), parseErr.Name(), parseErr.NormalizedName(), tt.wantToken, tt.wantName, tt.wantNormalized)
			}
			def, ok := parseErr.Definition()
			if !ok || def.Name() != tt.wantDefinition {
				t.Fatalf("ParseError.Definition() = (%q, %v), want %q true", def.Name(), ok, tt.wantDefinition)
			}
		})
	}
}

func TestParseCustomFailurePreservesNonSensitiveInspection(t *testing.T) {
	parserCause := errors.New("caller parser cause")
	set, err := flags.NewSet(flags.Custom("mode", flags.KindString, "", "mode", flags.ParserFunc(func(raw string) (any, error) {
		return nil, fmt.Errorf("caller rejected %q: %w", raw, parserCause)
	})))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--mode=bad"})
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("error does not satisfy ErrConversion: %v", err)
	}
	if !errors.Is(err, parserCause) {
		t.Fatalf("error does not preserve caller parser cause: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Token() != "--mode" || parseErr.Name() != "mode" || parseErr.NormalizedName() != "mode" {
		t.Fatalf("ParseError context token/name/normalized = %q/%q/%q, want --mode/mode/mode",
			parseErr.Token(), parseErr.Name(), parseErr.NormalizedName())
	}
	var valueErr *flags.ValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("error does not expose *flags.ValueError: %T", err)
	}
	if valueErr.Name() != "mode" || valueErr.Kind() != flags.KindString {
		t.Fatalf("ValueError context = name %q kind %v, want mode string", valueErr.Name(), valueErr.Kind())
	}
}

func TestParseSensitiveCustomFailureRedactsRawValueAndCause(t *testing.T) {
	const secret = "dib_fake_token_value"
	parserCause := errors.New("caller parser cause")
	set, err := flags.NewSet(flags.Custom("token", flags.KindString, "", "token", flags.ParserFunc(func(raw string) (any, error) {
		return nil, fmt.Errorf("caller rejected %s: %w", raw, parserCause)
	}), flags.Sensitive()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--token", secret})
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("error does not satisfy ErrConversion: %v", err)
	}
	if errors.Is(err, parserCause) {
		t.Fatalf("sensitive custom parser cause leaked through error inspection: %v", err)
	}
	if strings.Contains(err.Error(), secret) || strings.Contains(fmt.Sprintf("%+v", err), secret) {
		t.Fatalf("sensitive raw value leaked through error text: %v", err)
	}
	var valueErr *flags.ValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("error does not expose redacted *flags.ValueError context: %T", err)
	}
	if valueErr.Name() != "token" || valueErr.Kind() != flags.KindString {
		t.Fatalf("ValueError context = name %q kind %v, want token string", valueErr.Name(), valueErr.Kind())
	}
}

func TestParseCustomMutableResultsDoNotLeakAcrossRuns(t *testing.T) {
	parserResult := []string{"alpha"}
	set, err := flags.NewSet(flags.Custom("tag", flags.KindStringList, []string{}, "tags", flags.ParserFunc(func(raw string) (any, error) {
		return parserResult, nil
	}), flags.Repeatable()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	first, err := set.Parse([]string{"--tag=ignored"})
	if err != nil {
		t.Fatalf("first Parse returned unexpected error: %v", err)
	}
	firstState, _ := first.Lookup("tag")
	firstValues := firstState.Values()
	firstValues[0] = "caller-mutated"
	parserResult[0] = "beta"

	second, err := set.Parse([]string{"--tag=ignored"})
	if err != nil {
		t.Fatalf("second Parse returned unexpected error: %v", err)
	}
	freshFirstState, _ := first.Lookup("tag")
	if got := freshFirstState.Values(); !reflect.DeepEqual(got, []any{"alpha"}) {
		t.Fatalf("first snapshot Values() = %#v after later mutations, want alpha", got)
	}
	secondState, _ := second.Lookup("tag")
	if got := secondState.Values(); !reflect.DeepEqual(got, []any{"beta"}) {
		t.Fatalf("second snapshot Values() = %#v, want beta from independent parse run", got)
	}
}

func TestParseCustomRejectsMismatchedResultKindThroughParseContract(t *testing.T) {
	set, err := flags.NewSet(flags.Custom("workers", flags.KindInt, 1, "workers", flags.ParserFunc(func(raw string) (any, error) {
		return "not-an-int", nil
	})))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--workers=2"})
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("error does not satisfy ErrConversion: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error does not expose *flags.ParseError: %T", err)
	}
	var valueErr *flags.ValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("error does not expose *flags.ValueError: %T", err)
	}
	if valueErr.Name() != "workers" || valueErr.Kind() != flags.KindInt {
		t.Fatalf("ValueError context = name %q kind %v, want workers int", valueErr.Name(), valueErr.Kind())
	}
}

func assertOccurrences(t *testing.T, state flags.ValueState, wantSpellings []string, wantDefinition string) {
	t.Helper()

	if !state.Explicit() {
		t.Fatalf("%s state was not marked explicit", wantDefinition)
	}
	occurrences := state.Occurrences()
	if len(occurrences) != len(wantSpellings) {
		t.Fatalf("occurrence count = %d, want %d: %#v", len(occurrences), len(wantSpellings), occurrences)
	}
	for i, wantSpelling := range wantSpellings {
		if got := occurrences[i].Spelling(); got != wantSpelling {
			t.Fatalf("occurrence %d Spelling() = %q, want %q", i, got, wantSpelling)
		}
		if got := occurrences[i].NormalizedName(); got != wantDefinition {
			t.Fatalf("occurrence %d NormalizedName() = %q, want %q", i, got, wantDefinition)
		}
		if got := occurrences[i].Definition().Name(); got != wantDefinition {
			t.Fatalf("occurrence %d Definition().Name() = %q, want %q", i, got, wantDefinition)
		}
	}
}

func mustNewSet(t *testing.T, defs ...flags.Definition) flags.Set {
	t.Helper()

	set, err := flags.NewSet(defs...)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	return set
}

func mustNewNormalizedSet(t *testing.T, normalizer flags.NameNormalizer, defs ...flags.Definition) flags.Set {
	t.Helper()

	set, err := flags.NewNormalizedSet(normalizer, defs...)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}
	return set
}
