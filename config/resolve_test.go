package config_test

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/petabytecl/dib/config"
)

// resolveTestSet builds a config set for resolve tests.
func resolveTestSet(t *testing.T) config.Set {
	t.Helper()
	set, err := config.NewSet(
		config.String("level", "default-level", "log level"),
		config.String("output", "default-output", "output path"),
		config.Int("count", 5, "retry count"),
		config.Bool("verbose", false, "verbose mode"),
		config.Define("token", config.KindString, "api token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	return set
}

// resolveNormalizedTestSet builds a normalized config set for resolve tests.
func resolveNormalizedTestSet(t *testing.T) config.Set {
	t.Helper()
	set, err := config.NewNormalizedSet(
		normalizeConfigSeparators,
		config.String("log-level", "default-level", "log level"),
		config.Bool("debug", false, "debug mode"),
		config.Define("secret-key", config.KindString, "api secret", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}
	return set
}

// -- Precedence pair tests --

func TestResolvePrecedenceAdjacentPairs(t *testing.T) {
	t.Parallel()

	set := resolveTestSet(t)

	makeExplicit := func(value string) config.Snapshot {
		s, err := config.NewExplicitSnapshot(set, config.Assignment{Key: "level", Value: value})
		if err != nil {
			t.Fatalf("NewExplicitSnapshot: %v", err)
		}
		return s
	}
	makeFlag := func(value string) config.Snapshot {
		s, err := config.NewFlagSnapshot(set, []config.FlagValue{
			{ConfigKey: "level", ExplicitlySet: true, Value: value},
		})
		if err != nil {
			t.Fatalf("NewFlagSnapshot: %v", err)
		}
		return s
	}
	makeEnv := func(value string) config.Snapshot {
		s, err := config.NewEnvSnapshot(set, func(name string) (string, bool) {
			if name == "LEVEL" {
				return value, true
			}
			return "", false
		}, []config.EnvBinding{config.BindEnv("level", "LEVEL")})
		if err != nil {
			t.Fatalf("NewEnvSnapshot: %v", err)
		}
		return s
	}
	makeJSON := func(value string) config.Snapshot {
		s, err := config.LoadJSON(set, strings.NewReader(`{"level":"`+value+`"}`))
		if err != nil {
			t.Fatalf("LoadJSON: %v", err)
		}
		return s
	}
	empty := config.Snapshot{}

	tests := []struct {
		name           string
		explicit       config.Snapshot
		flag           config.Snapshot
		env            config.Snapshot
		jsonSrc        config.Snapshot
		wantLevel      string
		wantProvenance string
	}{
		{
			name:           "explicit beats flag",
			explicit:       makeExplicit("explicit-level"),
			flag:           makeFlag("flag-level"),
			env:            empty,
			jsonSrc:        empty,
			wantLevel:      "explicit-level",
			wantProvenance: config.SourceExplicit,
		},
		{
			name:           "flag beats env",
			explicit:       empty,
			flag:           makeFlag("flag-level"),
			env:            makeEnv("env-level"),
			jsonSrc:        empty,
			wantLevel:      "flag-level",
			wantProvenance: config.SourceFlagBinding,
		},
		{
			name:           "env beats JSON",
			explicit:       empty,
			flag:           empty,
			env:            makeEnv("env-level"),
			jsonSrc:        makeJSON("json-level"),
			wantLevel:      "env-level",
			wantProvenance: config.SourceEnv,
		},
		{
			name:           "JSON beats default",
			explicit:       empty,
			flag:           empty,
			env:            empty,
			jsonSrc:        makeJSON("json-level"),
			wantLevel:      "json-level",
			wantProvenance: config.SourceJSON,
		},
		{
			name:           "default when no higher source",
			explicit:       empty,
			flag:           empty,
			env:            empty,
			jsonSrc:        empty,
			wantLevel:      "default-level",
			wantProvenance: config.SourceDefault,
		},
		{
			name:           "explicit beats all",
			explicit:       makeExplicit("explicit-level"),
			flag:           makeFlag("flag-level"),
			env:            makeEnv("env-level"),
			jsonSrc:        makeJSON("json-level"),
			wantLevel:      "explicit-level",
			wantProvenance: config.SourceExplicit,
		},
		{
			name:           "lower tier with value does not override higher tier",
			explicit:       makeExplicit("explicit-level"),
			flag:           empty,
			env:            makeEnv("env-level"),
			jsonSrc:        makeJSON("json-level"),
			wantLevel:      "explicit-level",
			wantProvenance: config.SourceExplicit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := config.Resolve(set, tt.explicit, tt.flag, tt.env, tt.jsonSrc)
			v, ok := result.Lookup("level")
			if !ok {
				t.Fatal("Lookup(level) returned false")
			}
			got, hasValue := v.Value()
			if !hasValue || got != tt.wantLevel {
				t.Fatalf("Value() = %#v, %v; want %q, true", got, hasValue, tt.wantLevel)
			}
			if p := v.Provenance(); p != tt.wantProvenance {
				t.Fatalf("Provenance() = %q; want %q", p, tt.wantProvenance)
			}
		})
	}
}

// -- Zero-value Snapshot boundary --

func TestResolveZeroValueSnapshotsBehaviorEqualsDefaultSnapshot(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	resolved := config.Resolve(set, config.Snapshot{}, config.Snapshot{}, config.Snapshot{}, config.Snapshot{})
	defaults := set.DefaultSnapshot()

	for _, key := range []string{"level", "output", "count", "verbose"} {
		rv, rok := resolved.Lookup(key)
		dv, dok := defaults.Lookup(key)
		if rok != dok {
			t.Fatalf("Lookup(%q): resolved ok=%v, defaults ok=%v", key, rok, dok)
		}
		rVal, rHas := rv.Value()
		dVal, dHas := dv.Value()
		if rHas != dHas {
			t.Fatalf("key %q: resolved hasValue=%v, defaults hasValue=%v", key, rHas, dHas)
		}
		if !reflect.DeepEqual(rVal, dVal) {
			t.Fatalf("key %q: resolved value=%#v, defaults value=%#v", key, rVal, dVal)
		}
		if rv.Provenance() != dv.Provenance() {
			t.Fatalf("key %q: resolved provenance=%q, defaults provenance=%q", key, rv.Provenance(), dv.Provenance())
		}
	}
}

// -- Flag binding explicit-set behavior --

func TestResolveFlagDefaultDoesNotOverrideEnv(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	// Flag with ExplicitlySet=false should not inject any value.
	flagSnap, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "level", ExplicitlySet: false},
	})
	if err != nil {
		t.Fatalf("NewFlagSnapshot: %v", err)
	}
	envSnap, err := config.NewEnvSnapshot(set, func(name string) (string, bool) {
		if name == "LEVEL" {
			return "env-level", true
		}
		return "", false
	}, []config.EnvBinding{config.BindEnv("level", "LEVEL")})
	if err != nil {
		t.Fatalf("NewEnvSnapshot: %v", err)
	}

	result := config.Resolve(set, config.Snapshot{}, flagSnap, envSnap, config.Snapshot{})
	v, ok := result.Lookup("level")
	if !ok {
		t.Fatal("Lookup(level) returned false")
	}
	got, _ := v.Value()
	if got != "env-level" {
		t.Fatalf("Value() = %q; want env-level (flag default must not override env)", got)
	}
	if p := v.Provenance(); p != config.SourceEnv {
		t.Fatalf("Provenance() = %q; want %q", p, config.SourceEnv)
	}
}

func TestResolveFlagNotInFlagValuesDefersTolowerTiers(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	// Flag binding for "output" only; "level" absent from flag tier → should use env.
	flagSnap, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "output", ExplicitlySet: true, Value: "flag-output"},
	})
	if err != nil {
		t.Fatalf("NewFlagSnapshot: %v", err)
	}
	envSnap, err := config.NewEnvSnapshot(set, func(name string) (string, bool) {
		if name == "LEVEL" {
			return "env-level", true
		}
		return "", false
	}, []config.EnvBinding{config.BindEnv("level", "LEVEL")})
	if err != nil {
		t.Fatalf("NewEnvSnapshot: %v", err)
	}

	result := config.Resolve(set, config.Snapshot{}, flagSnap, envSnap, config.Snapshot{})
	lv, _ := result.Lookup("level")
	if got, _ := lv.Value(); got != "env-level" {
		t.Fatalf("level Value() = %q; want env-level", got)
	}
}

func TestResolveFlagExplicitBeatsEnvAndJSON(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	flagSnap, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "level", ExplicitlySet: true, Value: "flag-level"},
	})
	if err != nil {
		t.Fatalf("NewFlagSnapshot: %v", err)
	}
	envSnap, err := config.NewEnvSnapshot(set, func(name string) (string, bool) {
		if name == "LEVEL" {
			return "env-level", true
		}
		return "", false
	}, []config.EnvBinding{config.BindEnv("level", "LEVEL")})
	if err != nil {
		t.Fatalf("NewEnvSnapshot: %v", err)
	}
	jsonSnap, err := config.LoadJSON(set, strings.NewReader(`{"level":"json-level"}`))
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}

	result := config.Resolve(set, config.Snapshot{}, flagSnap, envSnap, jsonSnap)
	v, _ := result.Lookup("level")
	if got, _ := v.Value(); got != "flag-level" {
		t.Fatalf("Value() = %q; want flag-level", got)
	}
	if p := v.Provenance(); p != config.SourceFlagBinding {
		t.Fatalf("Provenance() = %q; want %q", p, config.SourceFlagBinding)
	}
}

// -- Absent / zero distinction --

func TestResolveAbsentKeyWithNoDefaultReturnsNoValue(t *testing.T) {
	t.Parallel()
	set, err := config.NewSet(
		config.Define("unset", config.KindString, "no default key"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	result := config.Resolve(set, config.Snapshot{}, config.Snapshot{}, config.Snapshot{}, config.Snapshot{})
	v, ok := result.Lookup("unset")
	if !ok {
		t.Fatal("Lookup(unset) returned false")
	}
	_, hasValue := v.Value()
	if hasValue {
		t.Fatal("Value() hasValue=true; want false for key with no default and no source")
	}
	if p := v.Provenance(); p != "" {
		t.Fatalf("Provenance() = %q; want empty for absent value", p)
	}
}

func TestResolveExplicitZeroValueBeatsLowerTierNonZeroValue(t *testing.T) {
	t.Parallel()
	set, err := config.NewSet(
		config.Int("count", 10, "retry count"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	explicit, err := config.NewExplicitSnapshot(set, config.Assignment{Key: "count", Value: 0})
	if err != nil {
		t.Fatalf("NewExplicitSnapshot: %v", err)
	}
	jsonSnap, err := config.LoadJSON(set, strings.NewReader(`{"count":99}`))
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}

	result := config.Resolve(set, explicit, config.Snapshot{}, config.Snapshot{}, jsonSnap)
	v, _ := result.Lookup("count")
	got, hasValue := v.Value()
	if !hasValue || got != 0 {
		t.Fatalf("Value() = %#v, %v; want 0, true (explicit zero beats JSON non-zero)", got, hasValue)
	}
	if p := v.Provenance(); p != config.SourceExplicit {
		t.Fatalf("Provenance() = %q; want %q", p, config.SourceExplicit)
	}
}

func TestResolveEmptyEnvStringCounts(t *testing.T) {
	t.Parallel()
	set, err := config.NewSet(
		config.String("level", "default", "log level"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	envSnap, err := config.NewEnvSnapshot(set, func(name string) (string, bool) {
		if name == "LEVEL" {
			return "", true // empty string is present
		}
		return "", false
	}, []config.EnvBinding{config.BindEnv("level", "LEVEL")})
	if err != nil {
		t.Fatalf("NewEnvSnapshot: %v", err)
	}
	jsonSnap, err := config.LoadJSON(set, strings.NewReader(`{"level":"json-value"}`))
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}

	result := config.Resolve(set, config.Snapshot{}, config.Snapshot{}, envSnap, jsonSnap)
	v, _ := result.Lookup("level")
	got, hasValue := v.Value()
	if !hasValue || got != "" {
		t.Fatalf("Value() = %#v, %v; want empty string, true (empty env string is present)", got, hasValue)
	}
	if p := v.Provenance(); p != config.SourceEnv {
		t.Fatalf("Provenance() = %q; want %q", p, config.SourceEnv)
	}
}

// -- Provenance labels --

func TestResolveProvenanceLabelsAreCanonical(t *testing.T) {
	t.Parallel()

	set := resolveTestSet(t)
	canonicalLabels := map[string]bool{
		config.SourceDefault:     true,
		config.SourceExplicit:    true,
		config.SourceFlagBinding: true,
		config.SourceEnv:         true,
		config.SourceJSON:        true,
		"":                       true, // absent value with no default
	}

	explicit, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "level", Value: "explicit-level"})
	flag, _ := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "output", ExplicitlySet: true, Value: "flag-output"},
	})
	env, _ := config.NewEnvSnapshot(set, func(name string) (string, bool) {
		if name == "COUNT" {
			return "3", true
		}
		return "", false
	}, []config.EnvBinding{config.BindEnv("count", "COUNT")})
	jsonSnap, _ := config.LoadJSON(set, strings.NewReader(`{"verbose":true}`))

	result := config.Resolve(set, explicit, flag, env, jsonSnap)

	for _, key := range []string{"level", "output", "count", "verbose", "token"} {
		v, ok := result.Lookup(key)
		if !ok {
			t.Fatalf("Lookup(%q) returned false", key)
		}
		p := v.Provenance()
		if !canonicalLabels[p] {
			t.Fatalf("key %q: Provenance() = %q; not in canonical vocabulary", key, p)
		}
	}
}

// -- NewFlagSnapshot error cases --

func TestNewFlagSnapshotUnknownKey(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	_, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "nonexistent", ExplicitlySet: true, Value: "value"},
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error for unknown key")
	}
	if !errors.Is(err, config.ErrUnknownSourceKey) {
		t.Fatalf("errors.Is(err, ErrUnknownSourceKey) = false; want true; err=%v", err)
	}
	var se *config.SourceError
	if !errors.As(err, &se) {
		t.Fatalf("errors.As(*SourceError) = false; want true")
	}
	if se.Source() != config.SourceFlagBinding {
		t.Fatalf("SourceError.Source() = %q; want %q", se.Source(), config.SourceFlagBinding)
	}
}

func TestNewFlagSnapshotDuplicateBinding(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	_, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "level", ExplicitlySet: true, Value: "first"},
		{ConfigKey: "level", ExplicitlySet: true, Value: "second"},
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error for duplicate binding")
	}
	if !errors.Is(err, config.ErrDuplicateBinding) {
		t.Fatalf("errors.Is(err, ErrDuplicateBinding) = false; want true; err=%v", err)
	}
	var se *config.SourceError
	if !errors.As(err, &se) {
		t.Fatalf("errors.As(*SourceError) = false; want true")
	}
	if se.Key() == "" {
		t.Fatalf("SourceError.Key() is empty; want colliding key")
	}
	if se.Source() != config.SourceFlagBinding {
		t.Fatalf("SourceError.Source() = %q; want %q", se.Source(), config.SourceFlagBinding)
	}
}

func TestNewFlagSnapshotNoPartialStateOnError(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	snap, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "level", ExplicitlySet: true, Value: "first"},
		{ConfigKey: "output", ExplicitlySet: true, Value: "second-output"},
		{ConfigKey: "level", ExplicitlySet: true, Value: "duplicate"},
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error for duplicate binding")
	}
	// Returned snapshot should be zero value (not partial).
	if _, ok := snap.Lookup("level"); ok {
		t.Fatal("Lookup on returned snapshot returned ok=true; want zero-value Snapshot on error")
	}
}

func TestNewFlagSnapshotKindMismatch(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	_, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "count", ExplicitlySet: true, Value: "not-an-int"},
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error for kind mismatch")
	}
	if !errors.Is(err, config.ErrSourceConversion) {
		t.Fatalf("errors.Is(err, ErrSourceConversion) = false; want true; err=%v", err)
	}
	var se *config.SourceError
	if !errors.As(err, &se) {
		t.Fatalf("errors.As(*SourceError) = false; want true")
	}
	if se.Source() != config.SourceFlagBinding {
		t.Fatalf("SourceError.Source() = %q; want %q", se.Source(), config.SourceFlagBinding)
	}
}

func TestNewFlagSnapshotSensitiveRedaction(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	// Attempt a kind mismatch on a sensitive key. The error must not echo the fake corpus.
	for _, sensitiveValue := range []any{
		42, // wrong kind for token (string), value used as diagnostic
	} {
		_, err := config.NewFlagSnapshot(set, []config.FlagValue{
			{ConfigKey: "token", ExplicitlySet: true, Value: sensitiveValue},
		})
		if err == nil {
			t.Fatal("NewFlagSnapshot returned nil error for sensitive key kind mismatch")
		}
		errStr := err.Error()
		for _, corpus := range []string{
			"dib_fake_secret_value", "dib_fake_password_value", "dib_fake_token_value",
		} {
			if strings.Contains(errStr, corpus) {
				t.Fatalf("error string contains sensitive corpus %q: %s", corpus, errStr)
			}
		}
		var se *config.SourceError
		if errors.As(err, &se) && !se.Redacted() {
			t.Fatalf("SourceError.Redacted() = false for sensitive key; want true")
		}
	}
}

// Check that an explicit binding with a sensitive value does not leak in errors.
func TestNewFlagSnapshotSensitiveBindingErrorDoesNotEchoCorpus(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	// Duplicate binding on sensitive key must not echo the corpus even if the value were present.
	_, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "token", ExplicitlySet: true, Value: "dib_fake_token_value"},
		{ConfigKey: "token", ExplicitlySet: true, Value: "dib_fake_token_value"},
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error for duplicate sensitive binding")
	}
	errStr := err.Error()
	for _, corpus := range []string{
		"dib_fake_secret_value", "dib_fake_password_value", "dib_fake_token_value",
	} {
		if strings.Contains(errStr, corpus) {
			t.Fatalf("error string contains sensitive corpus %q: %s", corpus, errStr)
		}
	}
}

// -- Immutability --

func TestResolveSourceSnapshotImmutability(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	explicit, err := config.NewExplicitSnapshot(set,
		config.Assignment{Key: "level", Value: "explicit-level"},
	)
	if err != nil {
		t.Fatalf("NewExplicitSnapshot: %v", err)
	}
	envLookup := func(name string) (string, bool) {
		if name == "LEVEL" {
			return "env-level", true
		}
		return "", false
	}
	env, err := config.NewEnvSnapshot(set, envLookup, []config.EnvBinding{config.BindEnv("level", "LEVEL")})
	if err != nil {
		t.Fatalf("NewEnvSnapshot: %v", err)
	}

	result1 := config.Resolve(set, explicit, config.Snapshot{}, env, config.Snapshot{})

	// Resolve again with same snapshots to prove reuse doesn't mutate source.
	result2 := config.Resolve(set, explicit, config.Snapshot{}, env, config.Snapshot{})

	v1, _ := result1.Lookup("level")
	v2, _ := result2.Lookup("level")
	g1, _ := v1.Value()
	g2, _ := v2.Value()
	if g1 != g2 {
		t.Fatalf("second Resolve returned different level %q vs %q", g2, g1)
	}
	if g1 != "explicit-level" {
		t.Fatalf("Value() = %q; want explicit-level", g1)
	}
}

func TestResolveReusableSourceSnapshots(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	explicit, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "level", Value: "e"})
	env, _ := config.NewEnvSnapshot(set, func(name string) (string, bool) {
		return "", false
	}, nil)
	jsonSnap, _ := config.LoadJSON(set, strings.NewReader(`{"verbose":true}`))

	// Calling Resolve multiple times with the same snapshots must return identical results.
	r1 := config.Resolve(set, explicit, config.Snapshot{}, env, jsonSnap)
	r2 := config.Resolve(set, explicit, config.Snapshot{}, env, jsonSnap)

	for _, key := range []string{"level", "verbose", "output"} {
		v1, _ := r1.Lookup(key)
		v2, _ := r2.Lookup(key)
		val1, has1 := v1.Value()
		val2, has2 := v2.Value()
		if has1 != has2 || !reflect.DeepEqual(val1, val2) {
			t.Fatalf("key %q: first %#v/%v vs second %#v/%v", key, val1, has1, val2, has2)
		}
		if v1.Provenance() != v2.Provenance() {
			t.Fatalf("key %q: provenance %q vs %q", key, v1.Provenance(), v2.Provenance())
		}
	}
}

// -- Concurrent reuse --

func TestResolveConcurrentSourceReuse(t *testing.T) {
	t.Parallel()
	set := resolveTestSet(t)

	explicit, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "level", Value: "concurrent-level"})
	env, _ := config.NewEnvSnapshot(set, func(name string) (string, bool) { return "", false }, nil)
	jsonSnap, _ := config.LoadJSON(set, strings.NewReader(`{"verbose":true}`))

	const goroutines = 10
	results := make([]config.Snapshot, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func() {
			defer wg.Done()
			results[i] = config.Resolve(set, explicit, config.Snapshot{}, env, jsonSnap)
		}()
	}
	wg.Wait()

	for i, r := range results {
		v, ok := r.Lookup("level")
		if !ok {
			t.Fatalf("goroutine %d: Lookup(level) returned false", i)
		}
		got, _ := v.Value()
		if got != "concurrent-level" {
			t.Fatalf("goroutine %d: level = %q; want concurrent-level", i, got)
		}
	}
}

// -- FlagValue ExplicitlySet=false with normalized set --

func TestNewFlagSnapshotNormalizedSetDuplicateBinding(t *testing.T) {
	t.Parallel()
	set := resolveNormalizedTestSet(t)

	// "log-level" and "log_level" normalize to the same key → duplicate binding.
	_, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "log-level", ExplicitlySet: true, Value: "first"},
		{ConfigKey: "log_level", ExplicitlySet: true, Value: "second"},
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error for normalized duplicate binding")
	}
	if !errors.Is(err, config.ErrDuplicateBinding) {
		t.Fatalf("errors.Is(err, ErrDuplicateBinding) = false; want true; err=%v", err)
	}
}

// -- SourceFlagBinding constant --

func TestSourceFlagBindingConstant(t *testing.T) {
	t.Parallel()
	if config.SourceFlagBinding != "flag binding" {
		t.Fatalf("SourceFlagBinding = %q; want \"flag binding\"", config.SourceFlagBinding)
	}
}
