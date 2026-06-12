package config_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/config"
)

func TestSetExactLookupAndDefinitionsAreDeterministic(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("log-level", "info", "hyphen"),
		config.String("log_level", "debug", "underscore"),
		config.String("log.level", "warn", "dot"),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	if got := set.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}
	for _, name := range []string{"log-level", "log_level", "log.level"} {
		def, ok := set.Lookup(name)
		if !ok {
			t.Fatalf("Lookup(%q) returned false", name)
		}
		if got := def.Name(); got != name {
			t.Fatalf("Lookup(%q).Name() = %q, want %q", name, got, name)
		}
	}

	defs := set.Definitions()
	gotNames := []string{defs[0].Name(), defs[1].Name(), defs[2].Name()}
	if !reflect.DeepEqual(gotNames, []string{"log-level", "log_level", "log.level"}) {
		t.Fatalf("Definitions order = %#v", gotNames)
	}
	defs[0] = config.String("mutated", "", "")
	if def, ok := set.Lookup("log-level"); !ok || def.Name() != "log-level" {
		t.Fatalf("Definitions() leaked mutable definition storage")
	}
}

func TestSetValidationErrorsAreInspectable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		defs           []config.Definition
		want           error
		wantKey        string
		wantKind       config.Kind
		wantProvenance string
	}{
		{
			name: "empty key",
			defs: []config.Definition{config.String("", "", "empty")},
			want: config.ErrInvalidDefinition,
		},
		{
			name:    "whitespace key",
			defs:    []config.Definition{config.String("log level", "info", "contains whitespace")},
			want:    config.ErrInvalidDefinition,
			wantKey: "log level",
		},
		{
			name:    "leading hyphen key",
			defs:    []config.Definition{config.String("-log-level", "info", "starts with hyphen")},
			want:    config.ErrInvalidDefinition,
			wantKey: "-log-level",
		},
		{
			name: "duplicate exact key",
			defs: []config.Definition{
				config.String("log-level", "info", "first"),
				config.Bool("log-level", false, "second"),
			},
			want:    config.ErrDuplicateKey,
			wantKey: "log-level",
		},
		{
			name: "invalid default type",
			defs: []config.Definition{
				config.Define("workers", config.KindInt, "workers", config.Default("not-an-int")),
			},
			want:           config.ErrInvalidDefault,
			wantKey:        "workers",
			wantKind:       config.KindInt,
			wantProvenance: config.SourceDefault,
		},
		{
			name: "nil option",
			defs: []config.Definition{
				config.Define("endpoint", config.KindString, "endpoint", nil),
			},
			want:    config.ErrInvalidDefinition,
			wantKey: "endpoint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.NewSet(tt.defs...)
			if err == nil {
				t.Fatal("NewSet returned nil error")
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.want, err)
			}

			var definitionErr *config.DefinitionError
			if !errors.As(err, &definitionErr) {
				t.Fatalf("error does not expose *config.DefinitionError: %T", err)
			}
			if got := definitionErr.Key(); got != tt.wantKey {
				t.Fatalf("DefinitionError.Key() = %q, want %q", got, tt.wantKey)
			}
			if tt.wantKind != config.KindString {
				if got := definitionErr.Kind(); got != tt.wantKind {
					t.Fatalf("DefinitionError.Kind() = %v, want %v", got, tt.wantKind)
				}
			}
			if got := definitionErr.Provenance(); got != tt.wantProvenance {
				t.Fatalf("DefinitionError.Provenance() = %q, want %q", got, tt.wantProvenance)
			}
		})
	}
}

func TestNormalizedSetLookupAndCollisions(t *testing.T) {
	t.Parallel()

	normalizeSeparators := config.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})

	set, err := config.NewNormalizedSet(
		normalizeSeparators,
		config.String("log-level", "info", "log level"),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}
	for _, spelling := range []string{"log-level", "log_level", "log.level"} {
		def, ok := set.Lookup(spelling)
		if !ok {
			t.Fatalf("Lookup(%q) returned false", spelling)
		}
		if got := def.Name(); got != "log-level" {
			t.Fatalf("Lookup(%q).Name() = %q, want log-level", spelling, got)
		}
	}
	for _, spelling := range []string{"", "-log-level", "log level"} {
		if def, ok := set.Lookup(spelling); ok {
			t.Fatalf("Lookup(%q) = (%q, true), want false", spelling, def.Name())
		}
	}

	_, err = config.NewNormalizedSet(
		normalizeSeparators,
		config.String("log-level", "info", "first"),
		config.String("log_level", "debug", "second"),
	)
	if err == nil {
		t.Fatal("NewNormalizedSet returned nil error for collision")
	}
	if !errors.Is(err, config.ErrDuplicateNormalizedKey) {
		t.Fatalf("errors.Is(err, ErrDuplicateNormalizedKey) = false; err=%v", err)
	}
	if errors.Is(err, config.ErrDuplicateKey) {
		t.Fatalf("normalized collision should not be reported as exact duplicate: %v", err)
	}
	var definitionErr *config.DefinitionError
	if !errors.As(err, &definitionErr) {
		t.Fatalf("error does not expose *config.DefinitionError: %T", err)
	}
	if got := definitionErr.NormalizedKey(); got != "log-level" {
		t.Fatalf("DefinitionError.NormalizedKey() = %q, want log-level", got)
	}
	gotKeys := map[string]bool{definitionErr.Key(): true, definitionErr.CollidingKey(): true}
	for _, want := range []string{"log-level", "log_level"} {
		if !gotKeys[want] {
			t.Fatalf("collision diagnostics missing %q: key=%q colliding=%q", want, definitionErr.Key(), definitionErr.CollidingKey())
		}
	}
}

func TestNormalizedSetRejectsInvalidNormalizedKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		normalizer     config.NameNormalizer
		wantNormalized string
	}{
		{
			name:           "empty normalized key",
			normalizer:     func(string) string { return "" },
			wantNormalized: "",
		},
		{
			name:           "leading hyphen normalized key",
			normalizer:     func(string) string { return "-log-level" },
			wantNormalized: "-log-level",
		},
		{
			name:           "whitespace normalized key",
			normalizer:     func(string) string { return "log level" },
			wantNormalized: "log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.NewNormalizedSet(tt.normalizer, config.String("log-level", "info", "log level"))
			if err == nil {
				t.Fatal("NewNormalizedSet returned nil error")
			}
			if !errors.Is(err, config.ErrInvalidDefinition) {
				t.Fatalf("errors.Is(err, ErrInvalidDefinition) = false; err=%v", err)
			}
			var definitionErr *config.DefinitionError
			if !errors.As(err, &definitionErr) {
				t.Fatalf("error does not expose *config.DefinitionError: %T", err)
			}
			if got := definitionErr.Key(); got != "log-level" {
				t.Fatalf("DefinitionError.Key() = %q, want log-level", got)
			}
			if got := definitionErr.NormalizedKey(); got != tt.wantNormalized {
				t.Fatalf("DefinitionError.NormalizedKey() = %q, want %q", got, tt.wantNormalized)
			}
		})
	}
}

func TestSetDerivationDoesNotMutateOriginals(t *testing.T) {
	t.Parallel()

	normalizeSeparators := config.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})

	base, err := config.NewSet(config.String("log-level", "info", "log level"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	normalized, err := base.WithNormalizer(normalizeSeparators)
	if err != nil {
		t.Fatalf("WithNormalizer returned unexpected error: %v", err)
	}
	if _, ok := base.Lookup("log_level"); ok {
		t.Fatal("WithNormalizer mutated the original exact-name set")
	}
	if def, ok := normalized.Lookup("log_level"); !ok || def.Name() != "log-level" {
		t.Fatalf("normalized Lookup(log_level) = (%q, %v), want canonical log-level", def.Name(), ok)
	}

	derived, err := normalized.With(config.Bool("debug", false, "debug mode"))
	if err != nil {
		t.Fatalf("With returned unexpected error: %v", err)
	}
	if _, ok := normalized.Lookup("debug"); ok {
		t.Fatal("With mutated the source normalized set")
	}
	if _, ok := derived.Lookup("debug"); !ok {
		t.Fatal("derived set is missing debug")
	}
	if def, ok := derived.Lookup("log.level"); !ok || def.Name() != "log-level" {
		t.Fatalf("derived Lookup(log.level) = (%q, %v), want canonical log-level", def.Name(), ok)
	}
}
