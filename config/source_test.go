package config_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/petabytecl/dib/config"
)

func TestExplicitSnapshotLastWriterWinsAndDefensiveValues(t *testing.T) {
	t.Parallel()

	inputTags := []string{"alpha"}
	set, err := config.NewNormalizedSet(
		normalizeConfigSeparators,
		config.String("log-level", "info", "log level"),
		config.StringList("tags", nil, "tags"),
		config.Duration("timeout", time.Second, "timeout"),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	snapshot, err := config.NewExplicitSnapshot(set,
		config.Assignment{Key: "log_level", Value: "debug"},
		config.Assignment{Key: "log-level", Value: "warn"},
		config.Assignment{Key: "tags", Value: inputTags},
		config.Assignment{Key: "timeout", Value: 2 * time.Second},
	)
	if err != nil {
		t.Fatalf("NewExplicitSnapshot returned unexpected error: %v", err)
	}
	inputTags[0] = "mutated"

	logLevel, ok := snapshot.Lookup("log.level")
	if !ok {
		t.Fatal("Lookup(log.level) returned false")
	}
	gotLogLevel, hasLogLevel := logLevel.Value()
	if !hasLogLevel || gotLogLevel != "warn" {
		t.Fatalf("Value() = %#v, %v; want warn, true", gotLogLevel, hasLogLevel)
	}
	if got := logLevel.Provenance(); got != config.SourceExplicit {
		t.Fatalf("Provenance() = %q, want %q", got, config.SourceExplicit)
	}
	source := logLevel.Source()
	if got := source.Key(); got != "log-level" {
		t.Fatalf("Source().Key() = %q, want canonical log-level", got)
	}
	if got := source.Label(); got != config.SourceExplicit {
		t.Fatalf("Source().Label() = %q, want explicit setter", got)
	}

	tags, ok := snapshot.Lookup("tags")
	if !ok {
		t.Fatal("Lookup(tags) returned false")
	}
	gotTags, hasTags := tags.Value()
	if !hasTags || !reflect.DeepEqual(gotTags, []string{"alpha"}) {
		t.Fatalf("tags Value() = %#v, %v; want alpha, true", gotTags, hasTags)
	}
	gotTags.([]string)[0] = "changed"
	gotAgain, _ := tags.Value()
	if !reflect.DeepEqual(gotAgain, []string{"alpha"}) {
		t.Fatalf("Value() leaked mutable slice alias: %#v", gotAgain)
	}

	timeout, ok := snapshot.Lookup("timeout")
	if !ok {
		t.Fatal("Lookup(timeout) returned false")
	}
	gotTimeout, hasTimeout := timeout.Value()
	if !hasTimeout || gotTimeout != 2*time.Second {
		t.Fatalf("timeout Value() = %#v, %v; want 2s, true", gotTimeout, hasTimeout)
	}
}

func TestExplicitSnapshotValidationErrorsAreTypedAndNoPartialSnapshot(t *testing.T) {
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
		assigns    []config.Assignment
		want       error
		wantKey    string
		wantKind   config.Kind
		wantSecret string
	}{
		{
			name:    "unknown key",
			assigns: []config.Assignment{{Key: "missing", Value: 1}},
			want:    config.ErrUnknownSourceKey,
			wantKey: "missing",
		},
		{
			name:     "type mismatch",
			assigns:  []config.Assignment{{Key: "workers", Value: "many"}},
			want:     config.ErrSourceConversion,
			wantKey:  "workers",
			wantKind: config.KindInt,
		},
		{
			name:       "sensitive raw value redacted",
			assigns:    []config.Assignment{{Key: "token", Value: "dib_fake_secret_value"}},
			want:       config.ErrSourceConversion,
			wantKey:    "token",
			wantKind:   config.KindInt,
			wantSecret: "dib_fake_secret_value",
		},
		{
			name: "invalid later write prevents partial use",
			assigns: []config.Assignment{
				{Key: "workers", Value: 2},
				{Key: "workers", Value: "many"},
			},
			want:     config.ErrSourceConversion,
			wantKey:  "workers",
			wantKind: config.KindInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := config.NewExplicitSnapshot(set, tt.assigns...)
			if err == nil {
				t.Fatal("NewExplicitSnapshot returned nil error")
			}
			if value, ok := snapshot.Lookup("workers"); ok {
				t.Fatalf("error returned usable partial snapshot value: %#v", value)
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.want, err)
			}
			var sourceErr *config.SourceError
			if !errors.As(err, &sourceErr) {
				t.Fatalf("error does not expose *config.SourceError: %T", err)
			}
			if got := sourceErr.Key(); got != tt.wantKey {
				t.Fatalf("SourceError.Key() = %q, want %q", got, tt.wantKey)
			}
			if got := sourceErr.Source(); got != config.SourceExplicit {
				t.Fatalf("SourceError.Source() = %q, want explicit setter", got)
			}
			if tt.wantKind != config.KindString {
				if got := sourceErr.Kind(); got != tt.wantKind {
					t.Fatalf("SourceError.Kind() = %v, want %v", got, tt.wantKind)
				}
			}
			if tt.wantSecret != "" && strings.Contains(err.Error(), tt.wantSecret) {
				t.Fatalf("sensitive source error leaked raw value: %v", err)
			}
		})
	}
}

func TestExplicitSnapshotZeroValuesAreSetValues(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("empty-string", "fallback", ""),
		config.Bool("disabled", true, ""),
		config.Int("zero", 9, ""),
		config.StringList("empty-list", []string{"fallback"}, ""),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	snapshot, err := config.NewExplicitSnapshot(set,
		config.Assignment{Key: "empty-string", Value: ""},
		config.Assignment{Key: "disabled", Value: false},
		config.Assignment{Key: "zero", Value: 0},
		config.Assignment{Key: "empty-list", Value: []string{}},
	)
	if err != nil {
		t.Fatalf("NewExplicitSnapshot returned unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want any
	}{
		{key: "empty-string", want: ""},
		{key: "disabled", want: false},
		{key: "zero", want: 0},
		{key: "empty-list", want: []string{}},
	}
	for _, tt := range tests {
		value, ok := snapshot.Lookup(tt.key)
		if !ok {
			t.Fatalf("Lookup(%q) returned false", tt.key)
		}
		got, hasValue := value.Value()
		if !hasValue || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("%s Value() = %#v, %v; want %#v, true", tt.key, got, hasValue, tt.want)
		}
	}
}

func normalizeConfigSeparators(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(name)
}
