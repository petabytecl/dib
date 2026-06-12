package config_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/petabytecl/dib/config"
)

func TestEnvSnapshotUsesInjectedLookupAndTracksMetadata(t *testing.T) {
	t.Parallel()

	set, err := config.NewNormalizedSet(
		normalizeConfigSeparators,
		config.String("log-level", "info", "log level"),
		config.Bool("debug", false, "debug"),
		config.Int("workers", 1, "workers"),
		config.Int64("limit64", 1, "limit64"),
		config.Uint("retries", 0, "retries"),
		config.Uint64("bytes", 0, "bytes"),
		config.Float64("ratio", 0, "ratio"),
		config.Duration("timeout", time.Second, "timeout"),
		config.StringList("tags", nil, "tags"),
		config.String("empty", "fallback", "empty"),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	env := map[string]string{
		"DIB_LOG_LEVEL": "debug",
		"DIB_DEBUG":     "false",
		"DIB_WORKERS":   "0",
		"DIB_LIMIT64":   "64",
		"DIB_RETRIES":   "3",
		"DIB_BYTES":     "128",
		"DIB_RATIO":     "0.25",
		"DIB_TIMEOUT":   "250ms",
		"DIB_TAGS":      "alpha,beta",
		"DIB_EMPTY":     "",
	}
	snapshot, err := config.NewEnvSnapshot(
		set,
		mapLookup(env),
		[]config.EnvBinding{
			config.MapEnv("log_level"),
			config.MapEnv("debug"),
			config.MapEnv("workers"),
			config.MapEnv("limit64"),
			config.MapEnv("retries"),
			config.MapEnv("bytes"),
			config.MapEnv("ratio"),
			config.MapEnv("timeout"),
			config.MapEnv("tags"),
			config.BindEnv("empty", "DIB_EMPTY"),
		},
		config.EnvPrefix("DIB_"),
		config.EnvKeyReplacer(strings.NewReplacer("-", "_")),
	)
	if err != nil {
		t.Fatalf("NewEnvSnapshot returned unexpected error: %v", err)
	}

	tests := []struct {
		key     string
		want    any
		envName string
	}{
		{key: "log-level", want: "debug", envName: "DIB_LOG_LEVEL"},
		{key: "debug", want: false, envName: "DIB_DEBUG"},
		{key: "workers", want: 0, envName: "DIB_WORKERS"},
		{key: "limit64", want: int64(64), envName: "DIB_LIMIT64"},
		{key: "retries", want: uint(3), envName: "DIB_RETRIES"},
		{key: "bytes", want: uint64(128), envName: "DIB_BYTES"},
		{key: "ratio", want: 0.25, envName: "DIB_RATIO"},
		{key: "timeout", want: 250 * time.Millisecond, envName: "DIB_TIMEOUT"},
		{key: "tags", want: []string{"alpha", "beta"}, envName: "DIB_TAGS"},
		{key: "empty", want: "", envName: "DIB_EMPTY"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			value, ok := snapshot.Lookup(tt.key)
			if !ok {
				t.Fatalf("Lookup(%q) returned false", tt.key)
			}
			got, hasValue := value.Value()
			if !hasValue || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Value() = %#v, %v; want %#v, true", got, hasValue, tt.want)
			}
			if got := value.Provenance(); got != config.SourceEnv {
				t.Fatalf("Provenance() = %q, want env", got)
			}
			if got := value.Source().EnvName(); got != tt.envName {
				t.Fatalf("Source().EnvName() = %q, want %q", got, tt.envName)
			}
		})
	}
}

func TestEnvSnapshotAbsentVariablesAreNotErrors(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("present", "", "present"),
		config.String("absent", "", "absent"),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	snapshot, err := config.NewEnvSnapshot(
		set,
		mapLookup(map[string]string{"PRESENT": "value"}),
		[]config.EnvBinding{config.BindEnv("present", "PRESENT"), config.BindEnv("absent", "ABSENT")},
	)
	if err != nil {
		t.Fatalf("NewEnvSnapshot returned unexpected error: %v", err)
	}
	if value, ok := snapshot.Lookup("present"); !ok {
		t.Fatal("Lookup(present) returned false")
	} else if got, hasValue := value.Value(); !hasValue || got != "value" {
		t.Fatalf("present Value() = %#v, %v; want value, true", got, hasValue)
	}
	if value, ok := snapshot.Lookup("absent"); !ok {
		t.Fatal("Lookup(absent) returned false for registered absent binding")
	} else if got, hasValue := value.Value(); hasValue || got != nil {
		t.Fatalf("absent Value() = %#v, %v; want nil, false", got, hasValue)
	}
}

func TestEnvSnapshotDiagnosticsAreTypedAndRedacted(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.Int("workers", 1, "workers"),
		config.Define("token", config.KindInt, "token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	tests := []struct {
		name       string
		lookup     config.EnvLookup
		bindings   []config.EnvBinding
		want       error
		wantKey    string
		wantEnv    string
		wantSecret string
	}{
		{
			name:     "nil lookup",
			lookup:   nil,
			bindings: []config.EnvBinding{config.BindEnv("workers", "WORKERS")},
			want:     config.ErrInvalidSource,
		},
		{
			name:     "unknown key",
			lookup:   mapLookup(nil),
			bindings: []config.EnvBinding{config.BindEnv("missing", "MISSING")},
			want:     config.ErrUnknownSourceKey,
			wantKey:  "missing",
			wantEnv:  "MISSING",
		},
		{
			name:     "empty explicit env name",
			lookup:   mapLookup(nil),
			bindings: []config.EnvBinding{config.BindEnv("workers", "")},
			want:     config.ErrInvalidSource,
			wantKey:  "workers",
		},
		{
			name:     "invalid conversion",
			lookup:   mapLookup(map[string]string{"WORKERS": "many"}),
			bindings: []config.EnvBinding{config.BindEnv("workers", "WORKERS")},
			want:     config.ErrSourceConversion,
			wantKey:  "workers",
			wantEnv:  "WORKERS",
		},
		{
			name:       "sensitive conversion redacts raw env value",
			lookup:     mapLookup(map[string]string{"TOKEN": "dib_fake_password_value"}),
			bindings:   []config.EnvBinding{config.BindEnv("token", "TOKEN")},
			want:       config.ErrSourceConversion,
			wantKey:    "token",
			wantEnv:    "TOKEN",
			wantSecret: "dib_fake_password_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.NewEnvSnapshot(set, tt.lookup, tt.bindings)
			if err == nil {
				t.Fatal("NewEnvSnapshot returned nil error")
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.want, err)
			}
			var sourceErr *config.SourceError
			if !errors.As(err, &sourceErr) {
				t.Fatalf("error does not expose *config.SourceError: %T", err)
			}
			if got := sourceErr.Source(); got != config.SourceEnv {
				t.Fatalf("SourceError.Source() = %q, want env", got)
			}
			if got := sourceErr.Key(); got != tt.wantKey {
				t.Fatalf("SourceError.Key() = %q, want %q", got, tt.wantKey)
			}
			if got := sourceErr.EnvName(); got != tt.wantEnv {
				t.Fatalf("SourceError.EnvName() = %q, want %q", got, tt.wantEnv)
			}
			if tt.wantSecret != "" && strings.Contains(err.Error(), tt.wantSecret) {
				t.Fatalf("sensitive env error leaked raw value: %v", err)
			}
		})
	}
}

func mapLookup(values map[string]string) config.EnvLookup {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
