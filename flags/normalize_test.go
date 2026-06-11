package flags_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestNameNormalizationLookupModes(t *testing.T) {
	normalizeSeparators := flags.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})

	tests := []struct {
		name       string
		build      func(t *testing.T) flags.Set
		lookups    map[string]string
		notLookups []string
	}{
		{
			name: "exact names remain distinct by default",
			build: func(t *testing.T) flags.Set {
				t.Helper()
				set, err := flags.NewSet(
					flags.String("log-level", "info", "log level"),
					flags.String("log_level", "debug", "log level"),
					flags.String("log.level", "warn", "log level"),
				)
				if err != nil {
					t.Fatalf("NewSet returned unexpected error: %v", err)
				}
				return set
			},
			lookups: map[string]string{
				"log-level": "log-level",
				"log_level": "log_level",
				"log.level": "log.level",
			},
		},
		{
			name: "nil normalizer preserves exact matching",
			build: func(t *testing.T) flags.Set {
				t.Helper()
				set, err := flags.NewNormalizedSet(nil, flags.String("log-level", "info", "log level"))
				if err != nil {
					t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
				}
				return set
			},
			lookups: map[string]string{
				"log-level": "log-level",
			},
			notLookups: []string{"log_level"},
		},
		{
			name: "configured normalizer resolves equivalent spellings",
			build: func(t *testing.T) flags.Set {
				t.Helper()
				set, err := flags.NewNormalizedSet(normalizeSeparators, flags.String("log-level", "info", "log level"))
				if err != nil {
					t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
				}
				return set
			},
			lookups: map[string]string{
				"log-level": "log-level",
				"log_level": "log-level",
				"log.level": "log-level",
			},
		},
		{
			name: "long-name normalization does not create shorthand aliases",
			build: func(t *testing.T) flags.Set {
				t.Helper()
				set, err := flags.NewNormalizedSet(
					normalizeSeparators,
					flags.Bool("log-level", false, "log level", flags.Shorthand("l")),
				)
				if err != nil {
					t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
				}
				return set
			},
			lookups: map[string]string{
				"log_level": "log-level",
			},
			notLookups: []string{"l"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := tt.build(t)
			for spelling, wantName := range tt.lookups {
				def, ok := set.Lookup(spelling)
				if !ok {
					t.Fatalf("Lookup(%q) returned false", spelling)
				}
				if got := def.Name(); got != wantName {
					t.Fatalf("Lookup(%q).Name() = %q, want %q", spelling, got, wantName)
				}
			}
			for _, spelling := range tt.notLookups {
				if def, ok := set.Lookup(spelling); ok {
					t.Fatalf("Lookup(%q) = (%q, true), want false", spelling, def.Name())
				}
			}
		})
	}
}

func TestNameNormalizationCollisionsAreTyped(t *testing.T) {
	normalizeSeparators := flags.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})

	_, err := flags.NewNormalizedSet(
		normalizeSeparators,
		flags.String("log-level", "info", "first"),
		flags.String("log_level", "debug", "second"),
	)
	if err == nil {
		t.Fatal("NewNormalizedSet returned nil error")
	}
	if !errors.Is(err, flags.ErrDuplicateNormalizedName) {
		t.Fatalf("errors.Is(err, ErrDuplicateNormalizedName) = false; err=%v", err)
	}
	if errors.Is(err, flags.ErrDuplicateName) {
		t.Fatalf("normalized collision should not be reported as exact duplicate name: %v", err)
	}

	var definitionErr *flags.DefinitionError
	if !errors.As(err, &definitionErr) {
		t.Fatalf("error does not expose *flags.DefinitionError: %T", err)
	}
	gotNames := map[string]bool{
		definitionErr.Name():          true,
		definitionErr.CollidingName(): true,
	}
	for _, want := range []string{"log-level", "log_level"} {
		if !gotNames[want] {
			t.Fatalf("collision diagnostics missing %q: name=%q colliding=%q", want, definitionErr.Name(), definitionErr.CollidingName())
		}
	}
	if got := definitionErr.NormalizedName(); got != "log-level" {
		t.Fatalf("NormalizedName() = %q, want log-level", got)
	}
}

func TestNameNormalizerOutputMustRemainAValidFlagName(t *testing.T) {
	_, err := flags.NewNormalizedSet(
		flags.NameNormalizer(func(string) string { return "" }),
		flags.String("log-level", "info", "log level"),
	)
	if err == nil {
		t.Fatal("NewNormalizedSet returned nil error")
	}
	if !errors.Is(err, flags.ErrInvalidDefinition) {
		t.Fatalf("errors.Is(err, ErrInvalidDefinition) = false; err=%v", err)
	}

	var definitionErr *flags.DefinitionError
	if !errors.As(err, &definitionErr) {
		t.Fatalf("error does not expose *flags.DefinitionError: %T", err)
	}
	if got := definitionErr.Name(); got != "log-level" {
		t.Fatalf("DefinitionError.Name() = %q, want log-level", got)
	}
	if got := definitionErr.NormalizedName(); got != "" {
		t.Fatalf("DefinitionError.NormalizedName() = %q, want empty string", got)
	}
}

func TestWithNormalizerAndWithKeepSetsIndependent(t *testing.T) {
	normalizeSeparators := flags.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})

	base, err := flags.NewSet(flags.String("log-level", "info", "log level"))
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

	derived, err := normalized.With(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("With returned unexpected error: %v", err)
	}
	if _, ok := normalized.Lookup("verbose"); ok {
		t.Fatal("With mutated the source normalized set")
	}
	if _, ok := derived.Lookup("verbose"); !ok {
		t.Fatal("derived set is missing verbose")
	}
	if def, ok := derived.Lookup("log.level"); !ok || def.Name() != "log-level" {
		t.Fatalf("derived Lookup(log.level) = (%q, %v), want canonical log-level", def.Name(), ok)
	}
}
