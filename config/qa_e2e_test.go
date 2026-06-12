package config_test

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/petabytecl/dib/config"
)

func TestQAConfigPublicWorkflowCoversDefaultsNormalizationAndNotFound(t *testing.T) {
	t.Parallel()

	normalizeSeparators := config.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})

	set, err := config.NewNormalizedSet(
		normalizeSeparators,
		config.String("log-level", "info", "log level"),
		config.Bool("debug", false, "debug mode"),
		config.StringList("tags", []string{"alpha", "beta"}, "tag list"),
		config.Define("token", config.KindString, "api token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	def, ok := set.Lookup("log_level")
	if !ok {
		t.Fatal("Lookup(log_level) returned false")
	}
	if got := def.Name(); got != "log-level" {
		t.Fatalf("Lookup(log_level).Name() = %q, want log-level", got)
	}

	snapshot := set.DefaultSnapshot()
	logLevel, ok := snapshot.Lookup("log.level")
	if !ok {
		t.Fatal("snapshot.Lookup(log.level) returned false")
	}
	gotLogLevel, hasLogLevel := logLevel.Value()
	if !hasLogLevel || gotLogLevel != "info" {
		t.Fatalf("log-level Value() = %#v, %v; want info, true", gotLogLevel, hasLogLevel)
	}
	if got := logLevel.Provenance(); got != config.SourceDefault {
		t.Fatalf("log-level Provenance() = %q, want %q", got, config.SourceDefault)
	}

	debug, ok := snapshot.Lookup("debug")
	if !ok {
		t.Fatal("snapshot.Lookup(debug) returned false")
	}
	gotDebug, hasDebug := debug.Value()
	if !hasDebug || gotDebug != false {
		t.Fatalf("debug Value() = %#v, %v; want false, true", gotDebug, hasDebug)
	}

	tags, ok := snapshot.Lookup("tags")
	if !ok {
		t.Fatal("snapshot.Lookup(tags) returned false")
	}
	gotTags, hasTags := tags.Value()
	if !hasTags || !reflect.DeepEqual(gotTags, []string{"alpha", "beta"}) {
		t.Fatalf("tags Value() = %#v, %v; want alpha/beta, true", gotTags, hasTags)
	}
	gotTags.([]string)[0] = "mutated"
	gotTagsAgain, _ := tags.Value()
	if !reflect.DeepEqual(gotTagsAgain, []string{"alpha", "beta"}) {
		t.Fatalf("Value() leaked mutable slice alias: %#v", gotTagsAgain)
	}

	token, ok := snapshot.Lookup("token")
	if !ok {
		t.Fatal("snapshot.Lookup(token) returned false for registered no-default key")
	}
	if got, hasValue := token.Value(); hasValue || got != nil {
		t.Fatalf("token Value() = %#v, %v; want nil, false", got, hasValue)
	}
	if got := token.Provenance(); got != "" {
		t.Fatalf("token Provenance() = %q, want empty for no-default key", got)
	}
	tokenDef, ok := token.Definition()
	if !ok {
		t.Fatal("token Definition() returned false")
	}
	if !tokenDef.Sensitive() {
		t.Fatal("token Definition().Sensitive() = false, want true")
	}

	if missing, ok := snapshot.Lookup("missing"); ok {
		t.Fatalf("snapshot.Lookup(missing) = (%#v, true), want false", missing)
	}
}

func TestQAConfigSetupErrorsCoverUnknownKindsAndNormalizedCollisions(t *testing.T) {
	t.Parallel()

	_, err := config.NewSet(config.Define("mode", config.Kind(99), "mode"))
	if err == nil {
		t.Fatal("NewSet returned nil error for unknown kind")
	}
	if !errors.Is(err, config.ErrInvalidDefinition) {
		t.Fatalf("errors.Is(err, ErrInvalidDefinition) = false; err=%v", err)
	}
	var definitionErr *config.DefinitionError
	if !errors.As(err, &definitionErr) {
		t.Fatalf("error does not expose *config.DefinitionError: %T", err)
	}
	if got := definitionErr.Key(); got != "mode" {
		t.Fatalf("DefinitionError.Key() = %q, want mode", got)
	}
	if got := definitionErr.Kind(); got != config.Kind(99) {
		t.Fatalf("DefinitionError.Kind() = %v, want unknown kind", got)
	}

	normalizeSeparators := config.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})
	_, err = config.NewSet(
		config.String("log-level", "info", "hyphen"),
		config.String("log_level", "debug", "underscore"),
	)
	if err != nil {
		t.Fatalf("NewSet exact spellings returned unexpected error: %v", err)
	}
	_, err = config.NewNormalizedSet(
		normalizeSeparators,
		config.String("log-level", "info", "hyphen"),
		config.String("log_level", "debug", "underscore"),
	)
	if err == nil {
		t.Fatal("NewNormalizedSet returned nil error for normalized collision")
	}
	if !errors.Is(err, config.ErrDuplicateNormalizedKey) {
		t.Fatalf("errors.Is(err, ErrDuplicateNormalizedKey) = false; err=%v", err)
	}
	if !errors.As(err, &definitionErr) {
		t.Fatalf("error does not expose *config.DefinitionError: %T", err)
	}
	if got := definitionErr.NormalizedKey(); got != "log-level" {
		t.Fatalf("DefinitionError.NormalizedKey() = %q, want log-level", got)
	}
}

func TestQAConfigSourceWorkflowCoversExplicitEnvAndJSONBoundaries(t *testing.T) {
	t.Parallel()

	normalizeSeparators := config.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})
	set, err := config.NewNormalizedSet(
		normalizeSeparators,
		config.String("log-level", "info", "log level"),
		config.Bool("debug", false, "debug mode"),
		config.Int("workers", 1, "worker count"),
		config.Uint("retries", 0, "retry count"),
		config.Float64("ratio", 0, "ratio"),
		config.Duration("timeout", time.Second, "timeout"),
		config.StringList("tags", []string{"default"}, "tag list"),
		config.String("empty", "fallback", "empty value"),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	sourceTags := []string{"explicit"}
	explicitSnapshot, err := config.NewExplicitSnapshot(set,
		config.Assignment{Key: "log_level", Value: "debug"},
		config.Assignment{Key: "log-level", Value: "warn"},
		config.Assignment{Key: "tags", Value: sourceTags},
	)
	if err != nil {
		t.Fatalf("NewExplicitSnapshot returned unexpected error: %v", err)
	}
	sourceTags[0] = "mutated"
	assertConfigValue(t, explicitSnapshot, "log.level", "warn", config.SourceExplicit)
	assertConfigValue(t, explicitSnapshot, "tags", []string{"explicit"}, config.SourceExplicit)
	explicitValue, _ := explicitSnapshot.Lookup("log-level")
	if got := explicitValue.Source().Key(); got != "log-level" {
		t.Fatalf("explicit Source().Key() = %q, want log-level", got)
	}
	if got := explicitValue.Source().Label(); got != config.SourceExplicit {
		t.Fatalf("explicit Source().Label() = %q, want %q", got, config.SourceExplicit)
	}

	envSnapshot, err := config.NewEnvSnapshot(
		set,
		mapLookup(map[string]string{
			"DIB_LOG_LEVEL": "env-debug",
			"DIB_DEBUG":     "true",
			"DIB_WORKERS":   "0",
			"DIB_EMPTY":     "",
		}),
		[]config.EnvBinding{
			config.MapEnv("log_level"),
			config.MapEnv("debug"),
			config.MapEnv("workers"),
			config.BindEnv("empty", "DIB_EMPTY"),
		},
		config.EnvPrefix("DIB_"),
		config.EnvKeyReplacer(strings.NewReplacer("-", "_")),
	)
	if err != nil {
		t.Fatalf("NewEnvSnapshot returned unexpected error: %v", err)
	}
	assertConfigValue(t, envSnapshot, "log-level", "env-debug", config.SourceEnv)
	assertConfigValue(t, envSnapshot, "debug", true, config.SourceEnv)
	assertConfigValue(t, envSnapshot, "workers", 0, config.SourceEnv)
	assertConfigValue(t, envSnapshot, "empty", "", config.SourceEnv)
	envValue, _ := envSnapshot.Lookup("log-level")
	if got := envValue.Source().EnvName(); got != "DIB_LOG_LEVEL" {
		t.Fatalf("env Source().EnvName() = %q, want DIB_LOG_LEVEL", got)
	}

	jsonSnapshot, err := config.LoadJSON(
		set,
		strings.NewReader(`{"log-level":"json-debug","timeout":"250ms","tags":["json","reader"],"unknown":true}`),
		config.JSONPermissive(),
		config.JSONReaderLabel("qa-inline"),
	)
	if err != nil {
		t.Fatalf("LoadJSON returned unexpected error: %v", err)
	}
	assertConfigValue(t, jsonSnapshot, "log-level", "json-debug", config.SourceJSON)
	assertConfigValue(t, jsonSnapshot, "timeout", 250*time.Millisecond, config.SourceJSON)
	assertConfigValue(t, jsonSnapshot, "tags", []string{"json", "reader"}, config.SourceJSON)
	jsonValue, _ := jsonSnapshot.Lookup("timeout")
	if got := jsonValue.Source().JSONReaderLabel(); got != "qa-inline" {
		t.Fatalf("JSON Source().JSONReaderLabel() = %q, want qa-inline", got)
	}

	fileSnapshot, err := config.LoadJSONFile(set, "testdata/json/valid.json")
	if err != nil {
		t.Fatalf("LoadJSONFile returned unexpected error: %v", err)
	}
	assertConfigValue(t, fileSnapshot, "log-level", "file-debug", config.SourceJSON)
	fileValue, _ := fileSnapshot.Lookup("log-level")
	if got := fileValue.Source().JSONPath(); got != "testdata/json/valid.json" {
		t.Fatalf("JSON file Source().JSONPath() = %q, want testdata/json/valid.json", got)
	}
}

func TestQAConfigSourceDiagnosticsCoverCriticalFailuresAndRedaction(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.Int("workers", 1, "worker count"),
		config.Define("token", config.KindInt, "api token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	tests := []struct {
		name       string
		run        func() error
		want       error
		wantSource string
		wantKey    string
		wantSecret string
		wantIs     error
	}{
		{
			name: "explicit unknown key",
			run: func() error {
				_, err := config.NewExplicitSnapshot(set, config.Assignment{Key: "missing", Value: 1})
				return err
			},
			want:       config.ErrUnknownSourceKey,
			wantSource: config.SourceExplicit,
			wantKey:    "missing",
		},
		{
			name: "env conversion failure",
			run: func() error {
				_, err := config.NewEnvSnapshot(set, mapLookup(map[string]string{"WORKERS": "many"}), []config.EnvBinding{
					config.BindEnv("workers", "WORKERS"),
				})
				return err
			},
			want:       config.ErrSourceConversion,
			wantSource: config.SourceEnv,
			wantKey:    "workers",
		},
		{
			name: "JSON strict unknown key",
			run: func() error {
				_, err := config.LoadJSON(set, strings.NewReader(`{"missing":true}`))
				return err
			},
			want:       config.ErrUnknownSourceKey,
			wantSource: config.SourceJSON,
			wantKey:    "missing",
		},
		{
			name: "JSON missing path preserves file inspection",
			run: func() error {
				_, err := config.LoadJSONFile(set, "testdata/json/does-not-exist.json")
				return err
			},
			want:       config.ErrSourceRead,
			wantSource: config.SourceJSON,
			wantIs:     os.ErrNotExist,
		},
		{
			name: "sensitive explicit conversion redacts raw value",
			run: func() error {
				_, err := config.NewExplicitSnapshot(set, config.Assignment{Key: "token", Value: "dib_fake_secret_value"})
				return err
			},
			want:       config.ErrSourceConversion,
			wantSource: config.SourceExplicit,
			wantKey:    "token",
			wantSecret: "dib_fake_secret_value",
		},
		{
			name: "sensitive env conversion redacts raw value",
			run: func() error {
				_, err := config.NewEnvSnapshot(set, mapLookup(map[string]string{"TOKEN": "dib_fake_password_value"}), []config.EnvBinding{
					config.BindEnv("token", "TOKEN"),
				})
				return err
			},
			want:       config.ErrSourceConversion,
			wantSource: config.SourceEnv,
			wantKey:    "token",
			wantSecret: "dib_fake_password_value",
		},
		{
			name: "sensitive JSON conversion redacts raw value",
			run: func() error {
				_, err := config.LoadJSON(set, strings.NewReader(`{"token":"dib_fake_token_value"}`))
				return err
			},
			want:       config.ErrSourceConversion,
			wantSource: config.SourceJSON,
			wantKey:    "token",
			wantSecret: "dib_fake_token_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil {
				t.Fatal("source operation returned nil error")
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.want, err)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.wantIs, err)
			}
			var sourceErr *config.SourceError
			if !errors.As(err, &sourceErr) {
				t.Fatalf("error does not expose *config.SourceError: %T", err)
			}
			if got := sourceErr.Source(); got != tt.wantSource {
				t.Fatalf("SourceError.Source() = %q, want %q", got, tt.wantSource)
			}
			if got := sourceErr.Key(); got != tt.wantKey {
				t.Fatalf("SourceError.Key() = %q, want %q", got, tt.wantKey)
			}
			if tt.wantSecret != "" && strings.Contains(err.Error(), tt.wantSecret) {
				t.Fatalf("sensitive source error leaked raw value: %v", err)
			}
			if tt.wantSecret != "" && !sourceErr.Redacted() {
				t.Fatal("SourceError.Redacted() = false, want true")
			}
		})
	}
}

func TestQAConfigResolutionWorkflowCoversFlagBindingAndFullPrecedence(t *testing.T) {
	t.Parallel()

	normalizeSeparators := config.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})
	set, err := config.NewNormalizedSet(
		normalizeSeparators,
		config.String("log-level", "info", "log level"),
		config.Int("workers", 1, "worker count"),
		config.Bool("debug", false, "debug mode"),
		config.Duration("timeout", time.Second, "timeout"),
		config.String("output", "default-output", "output path"),
		config.Float64("ratio", 1.5, "sampling ratio"),
		config.Define("token", config.KindString, "api token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet: %v", err)
	}

	// Explicit wins for "log-level".
	explicitSnap, err := config.NewExplicitSnapshot(set,
		config.Assignment{Key: "log-level", Value: "explicit-level"},
	)
	if err != nil {
		t.Fatalf("NewExplicitSnapshot: %v", err)
	}

	// Flag wins for "workers" and sensitive "token"; ExplicitlySet=false for "output" injects nothing.
	flagSnap, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "workers", ExplicitlySet: true, Value: 4},
		{ConfigKey: "output", ExplicitlySet: false},
		{ConfigKey: "token", ExplicitlySet: true, Value: "dib_fake_token_value"},
	})
	if err != nil {
		t.Fatalf("NewFlagSnapshot: %v", err)
	}

	// Env wins for "debug" and "output" (flag did not inject "output").
	envSnap, err := config.NewEnvSnapshot(
		set,
		mapLookup(map[string]string{
			"DIB_DEBUG":  "true",
			"DIB_OUTPUT": "env-output",
		}),
		[]config.EnvBinding{
			config.BindEnv("debug", "DIB_DEBUG"),
			config.BindEnv("output", "DIB_OUTPUT"),
		},
	)
	if err != nil {
		t.Fatalf("NewEnvSnapshot: %v", err)
	}

	// JSON wins for "timeout"; "log-level" and "workers" from JSON lose to higher tiers.
	jsonSnap, err := config.LoadJSON(
		set,
		strings.NewReader(`{"log-level":"json-level","workers":99,"timeout":"500ms","output":"json-output"}`),
		config.JSONPermissive(),
	)
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}

	result := config.Resolve(set, explicitSnap, flagSnap, envSnap, jsonSnap)

	// Explicit tier wins for "log-level".
	assertConfigValue(t, result, "log-level", "explicit-level", config.SourceExplicit)

	// Flag tier wins for "workers"; verify Source().Key() is canonical.
	assertConfigValue(t, result, "workers", 4, config.SourceFlagBinding)
	workersVal, _ := result.Lookup("workers")
	if got := workersVal.Source().Key(); got != "workers" {
		t.Fatalf("workers Source().Key() = %q; want workers", got)
	}

	// Env tier wins for "debug"; Source().EnvName() reflects the binding.
	assertConfigValue(t, result, "debug", true, config.SourceEnv)
	debugVal, _ := result.Lookup("debug")
	if got := debugVal.Source().EnvName(); got != "DIB_DEBUG" {
		t.Fatalf("debug Source().EnvName() = %q; want DIB_DEBUG", got)
	}

	// JSON tier wins for "timeout" (no higher source provided it).
	assertConfigValue(t, result, "timeout", 500*time.Millisecond, config.SourceJSON)

	// Env tier wins for "output": flag ExplicitlySet=false → no flag value; env beats JSON.
	assertConfigValue(t, result, "output", "env-output", config.SourceEnv)

	// Default tier wins for "ratio" (no source provided it).
	assertConfigValue(t, result, "ratio", 1.5, config.SourceDefault)

	// Flag tier wins for "token" (sensitive, explicitly set).
	tokenVal, ok := result.Lookup("token")
	if !ok {
		t.Fatal("Lookup(token) returned false")
	}
	gotToken, hasToken := tokenVal.Value()
	if !hasToken || gotToken != "dib_fake_token_value" {
		t.Fatalf("token Value() = %#v, %v; want dib_fake_token_value, true", gotToken, hasToken)
	}
	if p := tokenVal.Provenance(); p != config.SourceFlagBinding {
		t.Fatalf("token Provenance() = %q; want %q", p, config.SourceFlagBinding)
	}
	if !tokenVal.Source().Redacted() {
		t.Fatal("token Source().Redacted() = false; want true for sensitive key resolved via flag binding")
	}

	// Normalized key lookup works on the resolved snapshot.
	logLevelByUnderscore, ok := result.Lookup("log_level")
	if !ok {
		t.Fatal("Lookup(log_level) on resolved snapshot returned false")
	}
	if got, _ := logLevelByUnderscore.Value(); got != "explicit-level" {
		t.Fatalf("log_level via resolved Lookup = %q; want explicit-level", got)
	}

	// Resolved snapshot is reusable: a second Resolve with the same inputs returns identical results.
	result2 := config.Resolve(set, explicitSnap, flagSnap, envSnap, jsonSnap)
	for _, key := range []string{"log-level", "workers", "debug", "timeout", "output", "ratio"} {
		v1, _ := result.Lookup(key)
		v2, _ := result2.Lookup(key)
		val1, _ := v1.Value()
		val2, _ := v2.Value()
		if !reflect.DeepEqual(val1, val2) || v1.Provenance() != v2.Provenance() {
			t.Fatalf("key %q: second Resolve returned different results (%#v/%q vs %#v/%q)",
				key, val1, v1.Provenance(), val2, v2.Provenance())
		}
	}
}

func TestQAConfigFlagBindingDiagnosticsCoverErrorCategories(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.Int("workers", 1, "worker count"),
		config.String("log-level", "info", "log level"),
		config.Define("secret-int", config.KindInt, "secret integer", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	tests := []struct {
		name       string
		run        func() error
		wantIs     error
		wantSource string
		wantKey    string
		wantSecret string
	}{
		{
			name: "unknown config key",
			run: func() error {
				_, err := config.NewFlagSnapshot(set, []config.FlagValue{
					{ConfigKey: "nonexistent", ExplicitlySet: true, Value: "value"},
				})
				return err
			},
			wantIs:     config.ErrUnknownSourceKey,
			wantSource: config.SourceFlagBinding,
			wantKey:    "nonexistent",
		},
		{
			name: "kind mismatch for non-sensitive key",
			run: func() error {
				_, err := config.NewFlagSnapshot(set, []config.FlagValue{
					{ConfigKey: "workers", ExplicitlySet: true, Value: "not-an-int"},
				})
				return err
			},
			wantIs:     config.ErrSourceConversion,
			wantSource: config.SourceFlagBinding,
			wantKey:    "workers",
		},
		{
			name: "kind mismatch for sensitive key redacts raw value",
			run: func() error {
				_, err := config.NewFlagSnapshot(set, []config.FlagValue{
					{ConfigKey: "secret-int", ExplicitlySet: true, Value: "dib_fake_secret_value"},
				})
				return err
			},
			wantIs:     config.ErrSourceConversion,
			wantSource: config.SourceFlagBinding,
			wantKey:    "secret-int",
			wantSecret: "dib_fake_secret_value",
		},
		{
			name: "duplicate binding colliding key is reported",
			run: func() error {
				_, err := config.NewFlagSnapshot(set, []config.FlagValue{
					{ConfigKey: "workers", ExplicitlySet: true, Value: 2},
					{ConfigKey: "workers", ExplicitlySet: true, Value: 3},
				})
				return err
			},
			wantIs:     config.ErrDuplicateBinding,
			wantSource: config.SourceFlagBinding,
			wantKey:    "workers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.run()
			if err == nil {
				t.Fatal("NewFlagSnapshot returned nil error")
			}
			if !errors.Is(err, tt.wantIs) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.wantIs, err)
			}
			var sourceErr *config.SourceError
			if !errors.As(err, &sourceErr) {
				t.Fatalf("error does not expose *config.SourceError: %T", err)
			}
			if got := sourceErr.Source(); got != tt.wantSource {
				t.Fatalf("SourceError.Source() = %q; want %q", got, tt.wantSource)
			}
			if got := sourceErr.Key(); got != tt.wantKey {
				t.Fatalf("SourceError.Key() = %q; want %q", got, tt.wantKey)
			}
			if tt.wantSecret != "" {
				if strings.Contains(err.Error(), tt.wantSecret) {
					t.Fatalf("sensitive source error leaked raw value: %v", err)
				}
				if !sourceErr.Redacted() {
					t.Fatal("SourceError.Redacted() = false; want true for sensitive key")
				}
			}
		})
	}
}

func TestQAConfigTypedGettersWorkflowCoversAllKinds(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("sval", "hello", "string"),
		config.Bool("bval", true, "bool"),
		config.Int("ival", 42, "int"),
		config.Int64("i64val", int64(64), "int64"),
		config.Uint("uval", uint(7), "uint"),
		config.Uint64("u64val", uint64(128), "uint64"),
		config.Float64("fval", 3.14, "float64"),
		config.Duration("dval", 5*time.Second, "duration"),
		config.StringList("lval", []string{"x", "y"}, "string-list"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	snap := config.Resolve(set, config.Snapshot{}, config.Snapshot{}, config.Snapshot{}, config.Snapshot{})

	if s, err := snap.GetString("sval"); err != nil || s != "hello" {
		t.Fatalf("GetString(sval) = %q, %v; want hello, nil", s, err)
	}
	if v, err := snap.GetBool("bval"); err != nil || !v {
		t.Fatalf("GetBool(bval) = %v, %v; want true, nil", v, err)
	}
	if v, err := snap.GetInt("ival"); err != nil || v != 42 {
		t.Fatalf("GetInt(ival) = %d, %v; want 42, nil", v, err)
	}
	if v, err := snap.GetInt64("i64val"); err != nil || v != 64 {
		t.Fatalf("GetInt64(i64val) = %d, %v; want 64, nil", v, err)
	}
	if v, err := snap.GetUint("uval"); err != nil || v != 7 {
		t.Fatalf("GetUint(uval) = %d, %v; want 7, nil", v, err)
	}
	if v, err := snap.GetUint64("u64val"); err != nil || v != 128 {
		t.Fatalf("GetUint64(u64val) = %d, %v; want 128, nil", v, err)
	}
	if v, err := snap.GetFloat64("fval"); err != nil || v != 3.14 {
		t.Fatalf("GetFloat64(fval) = %f, %v; want 3.14, nil", v, err)
	}
	if v, err := snap.GetDuration("dval"); err != nil || v != 5*time.Second {
		t.Fatalf("GetDuration(dval) = %v, %v; want 5s, nil", v, err)
	}
	if v, err := snap.GetStringList("lval"); err != nil || len(v) != 2 || v[0] != "x" || v[1] != "y" {
		t.Fatalf("GetStringList(lval) = %v, %v; want [x y], nil", v, err)
	}

	for _, key := range []string{"sval", "bval", "ival", "i64val", "uval", "u64val", "fval", "dval", "lval"} {
		if !snap.IsSet(key) {
			t.Fatalf("IsSet(%q) = false, want true for default-valued key", key)
		}
	}
}

func TestQAConfigTypedGettersResolvedPresenceStates(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.Define("absent", config.KindString, "no source"),
		config.Int("default-zero", 0, "default zero"),
		config.String("empty-default", "", "empty default"),
		config.Int("explicit-zero", 99, "explicit zero"),
		config.String("empty-env", "fallback", "empty env"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	explicitSnap, err := config.NewExplicitSnapshot(set, config.Assignment{Key: "explicit-zero", Value: 0})
	if err != nil {
		t.Fatalf("NewExplicitSnapshot: %v", err)
	}
	envSnap, err := config.NewEnvSnapshot(set,
		mapLookup(map[string]string{"EMPTY_ENV": ""}),
		[]config.EnvBinding{config.BindEnv("empty-env", "EMPTY_ENV")},
	)
	if err != nil {
		t.Fatalf("NewEnvSnapshot: %v", err)
	}

	snap := config.Resolve(set, explicitSnap, config.Snapshot{}, envSnap, config.Snapshot{})

	_, err = snap.GetString("absent")
	if err == nil || !errors.Is(err, config.ErrKeyAbsent) {
		t.Fatalf("GetString(absent): errors.Is(ErrKeyAbsent) = false; err=%v", err)
	}
	if snap.IsSet("absent") {
		t.Fatal("IsSet(absent) = true, want false for registered key with no value")
	}

	if got, err := snap.GetInt("default-zero"); err != nil || got != 0 {
		t.Fatalf("GetInt(default-zero) = %d, %v; want 0, nil", got, err)
	}
	if !snap.IsSet("default-zero") {
		t.Fatal("IsSet(default-zero) = false, want true for zero default")
	}

	if got, err := snap.GetString("empty-default"); err != nil || got != "" {
		t.Fatalf("GetString(empty-default) = %q, %v; want empty, nil", got, err)
	}
	if !snap.IsSet("empty-default") {
		t.Fatal("IsSet(empty-default) = false, want true for empty default")
	}

	if got, err := snap.GetInt("explicit-zero"); err != nil || got != 0 {
		t.Fatalf("GetInt(explicit-zero) = %d, %v; want 0, nil", got, err)
	}
	if !snap.IsSet("explicit-zero") {
		t.Fatal("IsSet(explicit-zero) = false, want true for explicit zero")
	}
	explicitValue, _ := snap.Lookup("explicit-zero")
	if got := explicitValue.Provenance(); got != config.SourceExplicit {
		t.Fatalf("explicit-zero Provenance() = %q, want %q", got, config.SourceExplicit)
	}

	if got, err := snap.GetString("empty-env"); err != nil || got != "" {
		t.Fatalf("GetString(empty-env) = %q, %v; want empty, nil", got, err)
	}
	if !snap.IsSet("empty-env") {
		t.Fatal("IsSet(empty-env) = false, want true for empty env value")
	}
	envValue, _ := snap.Lookup("empty-env")
	if got := envValue.Provenance(); got != config.SourceEnv {
		t.Fatalf("empty-env Provenance() = %q, want %q", got, config.SourceEnv)
	}

	if snap.IsSet("unregistered") {
		t.Fatal("IsSet(unregistered) = true, want false")
	}
	_, err = snap.GetString("unregistered")
	if err == nil || !errors.Is(err, config.ErrKeyNotFound) {
		t.Fatalf("GetString(unregistered): errors.Is(ErrKeyNotFound) = false; err=%v", err)
	}
}

func TestQAConfigGetterDiagnosticsCoverAbsenceAndKindMismatch(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.Bool("flag", true, "flag"),
		config.Define("absent", config.KindString, "no default"),
		config.Define("sensitive-absent", config.KindString, "sensitive", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "flag", Value: true})

	// ErrKeyNotFound
	_, err = snap.GetString("nonexistent")
	if err == nil || !errors.Is(err, config.ErrKeyNotFound) {
		t.Fatalf("ErrKeyNotFound: errors.Is = false; err=%v", err)
	}
	var getErr *config.GetError
	if !errors.As(err, &getErr) {
		t.Fatalf("ErrKeyNotFound: error does not expose *config.GetError: %T", err)
	}
	if getErr.Key() != "nonexistent" {
		t.Fatalf("ErrKeyNotFound: Key() = %q, want nonexistent", getErr.Key())
	}

	// ErrKeyAbsent
	defaultSnap := set.DefaultSnapshot()
	_, err = defaultSnap.GetString("absent")
	if err == nil || !errors.Is(err, config.ErrKeyAbsent) {
		t.Fatalf("ErrKeyAbsent: errors.Is = false; err=%v", err)
	}
	if !errors.As(err, &getErr) {
		t.Fatalf("ErrKeyAbsent: error does not expose *config.GetError: %T", err)
	}
	if getErr.Key() != "absent" {
		t.Fatalf("ErrKeyAbsent: Key() = %q, want absent", getErr.Key())
	}

	// ErrGetConversion (bool key, asked for string)
	_, err = snap.GetString("flag")
	if err == nil || !errors.Is(err, config.ErrGetConversion) {
		t.Fatalf("ErrGetConversion: errors.Is = false; err=%v", err)
	}
	if !errors.As(err, &getErr) {
		t.Fatalf("ErrGetConversion: error does not expose *config.GetError: %T", err)
	}
	if getErr.Kind() != config.KindBool {
		t.Fatalf("ErrGetConversion: Kind() = %v, want KindBool", getErr.Kind())
	}
	if getErr.WantKind() != config.KindString {
		t.Fatalf("ErrGetConversion: WantKind() = %v, want KindString", getErr.WantKind())
	}

	// Sensitive absent key: Redacted() true, corpus absent from error string
	_, err = defaultSnap.GetString("sensitive-absent")
	if err == nil || !errors.Is(err, config.ErrKeyAbsent) {
		t.Fatalf("sensitive absent: errors.Is(ErrKeyAbsent) = false; err=%v", err)
	}
	if !errors.As(err, &getErr) {
		t.Fatalf("sensitive absent: error does not expose *config.GetError: %T", err)
	}
	if !getErr.Redacted() {
		t.Fatal("sensitive absent: Redacted() = false, want true")
	}
	corpus := []string{"dib_fake_secret_value", "dib_fake_password_value", "dib_fake_token_value"}
	for _, raw := range corpus {
		if strings.Contains(err.Error(), raw) {
			t.Fatalf("sensitive absent error leaked corpus value %q: %v", raw, err)
		}
	}
}

func TestQAConfigProvenanceReportsExplainWinningSourcesWithoutValues(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("log-level", "info", "log level"),
		config.Int("workers", 1, "workers"),
		config.Bool("debug", false, "debug"),
		config.Define("absent", config.KindString, "absent"),
		config.Define("token", config.KindString, "token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	explicit, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "log-level", Value: "explicit-level"})
	flag, _ := config.NewFlagSnapshot(set, []config.FlagValue{{ConfigKey: "workers", ExplicitlySet: true, Value: 8}})
	env, _ := config.NewEnvSnapshot(set, mapLookup(map[string]string{"DEBUG": "true"}), []config.EnvBinding{config.BindEnv("debug", "DEBUG")})
	jsonSnap, _ := config.LoadJSON(set, strings.NewReader(`{"token":"dib_fake_token_value"}`), config.JSONReaderLabel("qa-inline"))

	snapshot := config.Resolve(set, explicit, flag, env, jsonSnap)
	report := snapshot.SourceReport()
	if got, want := len(report), 5; got != want {
		t.Fatalf("SourceReport length = %d, want %d", got, want)
	}
	wantSources := []string{config.SourceExplicit, config.SourceFlagBinding, config.SourceEnv, "", config.SourceJSON}
	for i, want := range wantSources {
		if got := report[i].SourceLabel(); got != want {
			t.Fatalf("report[%d].SourceLabel() = %q, want %q", i, got, want)
		}
	}
	if report[3].IsSet() {
		t.Fatal("absent report entry IsSet() = true, want false")
	}
	if !report[4].Redacted() {
		t.Fatal("token report entry Redacted() = false, want true")
	}
	var rendered bytes.Buffer
	if err := snapshot.WriteSourceReport(&rendered); err != nil {
		t.Fatalf("WriteSourceReport: %v", err)
	}
	for _, forbidden := range []string{"explicit-level", "dib_fake_token_value"} {
		if strings.Contains(rendered.String(), forbidden) {
			t.Fatalf("source report leaked raw value %q:\n%s", forbidden, rendered.String())
		}
	}
}

func TestQAConfigDiagnosticsDistinguishSourceAndCategory(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(config.Int("workers", 1, "workers"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	_, err = config.NewEnvSnapshot(set, mapLookup(map[string]string{"WORKERS": "many"}), []config.EnvBinding{config.BindEnv("workers", "WORKERS")})
	if err == nil {
		t.Fatal("NewEnvSnapshot returned nil error")
	}
	diag, ok := config.InspectDiagnostic(err)
	if !ok {
		t.Fatal("InspectDiagnostic returned ok=false")
	}
	if diag.Category() != config.ErrSourceConversion {
		t.Fatalf("Category() = %v, want ErrSourceConversion", diag.Category())
	}
	if diag.SourceLabel() != config.SourceEnv {
		t.Fatalf("SourceLabel() = %q, want env", diag.SourceLabel())
	}
	if diag.EnvName() != "WORKERS" {
		t.Fatalf("EnvName() = %q, want WORKERS", diag.EnvName())
	}
	var rendered bytes.Buffer
	if err := config.WriteDiagnostic(&rendered, err); err != nil {
		t.Fatalf("WriteDiagnostic: %v", err)
	}
	if !strings.Contains(rendered.String(), `category="config source value conversion failure"`) ||
		!strings.Contains(rendered.String(), `source="env"`) {
		t.Fatalf("rendered diagnostic did not distinguish source and category:\n%s", rendered.String())
	}
}

func TestQAConfigProvenanceRenderingRedactsSensitiveCorpus(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("ordinary", "ordinary-default", "ordinary"),
		config.Define("token", config.KindString, "token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	explicit, _ := config.NewExplicitSnapshot(set,
		config.Assignment{Key: "ordinary", Value: "ordinary-runtime-value"},
		config.Assignment{Key: "token", Value: "dib_fake_secret_value"},
	)
	snapshot := config.Resolve(set, explicit, config.Snapshot{}, config.Snapshot{}, config.Snapshot{})

	var report bytes.Buffer
	if err := snapshot.WriteSourceReport(&report); err != nil {
		t.Fatalf("WriteSourceReport: %v", err)
	}
	_, err = config.NewExplicitSnapshot(set, config.Assignment{Key: "token", Value: 42})
	if err == nil {
		t.Fatal("NewExplicitSnapshot returned nil error for sensitive mismatch")
	}
	var diagnostic bytes.Buffer
	if err := config.WriteDiagnostic(&diagnostic, err); err != nil {
		t.Fatalf("WriteDiagnostic: %v", err)
	}
	combined := report.String() + diagnostic.String()
	for _, forbidden := range []string{
		"ordinary-default",
		"ordinary-runtime-value",
		"dib_fake_secret_value",
		"dib_fake_password_value",
		"dib_fake_token_value",
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("rendered provenance/diagnostic output leaked %q:\n%s", forbidden, combined)
		}
	}
}

func assertConfigValue(t *testing.T, snapshot config.Snapshot, key string, want any, wantSource string) {
	t.Helper()

	value, ok := snapshot.Lookup(key)
	if !ok {
		t.Fatalf("Lookup(%q) returned false", key)
	}
	got, hasValue := value.Value()
	if !hasValue || !reflect.DeepEqual(got, want) {
		t.Fatalf("%s Value() = %#v, %v; want %#v, true", key, got, hasValue, want)
	}
	if got := value.Provenance(); got != wantSource {
		t.Fatalf("%s Provenance() = %q, want %q", key, got, wantSource)
	}
	if got := value.Source().Label(); got != wantSource {
		t.Fatalf("%s Source().Label() = %q, want %q", key, got, wantSource)
	}
}
