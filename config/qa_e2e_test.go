package config_test

import (
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
