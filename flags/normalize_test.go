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
		lookups    []lookupCase
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
			lookups: []lookupCase{
				{spelling: "log-level", wantName: "log-level"},
				{spelling: "log_level", wantName: "log_level"},
				{spelling: "log.level", wantName: "log.level"},
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
			lookups: []lookupCase{
				{spelling: "log-level", wantName: "log-level"},
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
			lookups: []lookupCase{
				{spelling: "log-level", wantName: "log-level"},
				{spelling: "log_level", wantName: "log-level"},
				{spelling: "log.level", wantName: "log-level"},
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
			lookups: []lookupCase{
				{spelling: "log_level", wantName: "log-level"},
			},
			notLookups: []string{"l"},
		},
		{
			name: "raw lookup names must be valid before normalization",
			build: func(t *testing.T) flags.Set {
				t.Helper()
				set, err := flags.NewNormalizedSet(
					flags.NameNormalizer(func(string) string { return "log-level" }),
					flags.Bool("log-level", false, "log level"),
				)
				if err != nil {
					t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
				}
				return set
			},
			lookups: []lookupCase{
				{spelling: "log-level", wantName: "log-level"},
			},
			notLookups: []string{"", "-log-level", "log level"},
		},
		{
			name: "normalizer cannot map a registered shorthand to a long-name alias",
			build: func(t *testing.T) flags.Set {
				t.Helper()
				set, err := flags.NewNormalizedSet(
					flags.NameNormalizer(func(name string) string {
						if name == "l" {
							return "log-level"
						}
						return strings.NewReplacer("_", "-", ".", "-").Replace(name)
					}),
					flags.Bool("log-level", false, "log level", flags.Shorthand("l")),
				)
				if err != nil {
					t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
				}
				return set
			},
			lookups: []lookupCase{
				{spelling: "log_level", wantName: "log-level"},
			},
			notLookups: []string{"l"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := tt.build(t)
			for _, lookup := range tt.lookups {
				def, ok := set.Lookup(lookup.spelling)
				if !ok {
					t.Fatalf("Lookup(%q) returned false", lookup.spelling)
				}
				if got := def.Name(); got != lookup.wantName {
					t.Fatalf("Lookup(%q).Name() = %q, want %q", lookup.spelling, got, lookup.wantName)
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

type lookupCase struct {
	spelling string
	wantName string
}

func TestNameNormalizationValidationErrors(t *testing.T) {
	normalizeSeparators := flags.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})

	tests := []struct {
		name                 string
		build                func(t *testing.T) error
		want                 error
		wantName             string
		wantNormalized       string
		wantCollidingNames   []string
		rejectExactDuplicate bool
	}{
		{
			name: "new normalized set rejects normalized collision",
			build: func(t *testing.T) error {
				t.Helper()
				_, err := flags.NewNormalizedSet(
					normalizeSeparators,
					flags.String("log-level", "info", "first"),
					flags.String("log_level", "debug", "second"),
				)
				return err
			},
			want:                 flags.ErrDuplicateNormalizedName,
			wantNormalized:       "log-level",
			wantCollidingNames:   []string{"log-level", "log_level"},
			rejectExactDuplicate: true,
		},
		{
			name: "with normalizer rejects existing exact names that normalize together",
			build: func(t *testing.T) error {
				t.Helper()
				base, err := flags.NewSet(
					flags.String("log-level", "info", "first"),
					flags.String("log_level", "debug", "second"),
				)
				if err != nil {
					t.Fatalf("NewSet returned unexpected error: %v", err)
				}
				_, err = base.WithNormalizer(normalizeSeparators)
				return err
			},
			want:                 flags.ErrDuplicateNormalizedName,
			wantNormalized:       "log-level",
			wantCollidingNames:   []string{"log-level", "log_level"},
			rejectExactDuplicate: true,
		},
		{
			name: "empty normalized definition key is invalid",
			build: func(t *testing.T) error {
				t.Helper()
				_, err := flags.NewNormalizedSet(
					flags.NameNormalizer(func(string) string { return "" }),
					flags.String("log-level", "info", "log level"),
				)
				return err
			},
			want:     flags.ErrInvalidDefinition,
			wantName: "log-level",
		},
		{
			name: "leading hyphen normalized definition key is invalid",
			build: func(t *testing.T) error {
				t.Helper()
				_, err := flags.NewNormalizedSet(
					flags.NameNormalizer(func(string) string { return "-log-level" }),
					flags.String("log-level", "info", "log level"),
				)
				return err
			},
			want:           flags.ErrInvalidDefinition,
			wantName:       "log-level",
			wantNormalized: "-log-level",
		},
		{
			name: "whitespace normalized definition key is invalid",
			build: func(t *testing.T) error {
				t.Helper()
				_, err := flags.NewNormalizedSet(
					flags.NameNormalizer(func(string) string { return "log level" }),
					flags.String("log-level", "info", "log level"),
				)
				return err
			},
			want:           flags.ErrInvalidDefinition,
			wantName:       "log-level",
			wantNormalized: "log level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.build(t)
			if err == nil {
				t.Fatal("operation returned nil error")
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.want, err)
			}
			if tt.rejectExactDuplicate && errors.Is(err, flags.ErrDuplicateName) {
				t.Fatalf("normalized collision should not be reported as exact duplicate name: %v", err)
			}

			var definitionErr *flags.DefinitionError
			if !errors.As(err, &definitionErr) {
				t.Fatalf("error does not expose *flags.DefinitionError: %T", err)
			}
			if tt.wantName != "" {
				if got := definitionErr.Name(); got != tt.wantName {
					t.Fatalf("DefinitionError.Name() = %q, want %q", got, tt.wantName)
				}
			}
			if got := definitionErr.NormalizedName(); got != tt.wantNormalized {
				t.Fatalf("DefinitionError.NormalizedName() = %q, want %q", got, tt.wantNormalized)
			}
			if len(tt.wantCollidingNames) > 0 {
				gotNames := map[string]bool{
					definitionErr.Name():          true,
					definitionErr.CollidingName(): true,
				}
				for _, want := range tt.wantCollidingNames {
					if !gotNames[want] {
						t.Fatalf("collision diagnostics missing %q: name=%q colliding=%q", want, definitionErr.Name(), definitionErr.CollidingName())
					}
				}
			}
		})
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
