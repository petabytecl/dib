package flags_test

import (
	"errors"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestNewSetValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		defs      []flags.Definition
		want      error
		flagName  string
		shorthand string
	}{
		{
			name: "empty name",
			defs: []flags.Definition{
				flags.String("", "", "empty"),
			},
			want: flags.ErrInvalidDefinition,
		},
		{
			name: "whitespace name",
			defs: []flags.Definition{
				flags.String("config path", "", "contains whitespace"),
			},
			want:     flags.ErrInvalidDefinition,
			flagName: "config path",
		},
		{
			name: "leading hyphen",
			defs: []flags.Definition{
				flags.String("-config", "", "starts with hyphen"),
			},
			want:     flags.ErrInvalidDefinition,
			flagName: "-config",
		},
		{
			name: "duplicate long name",
			defs: []flags.Definition{
				flags.String("config", "dev.json", "first"),
				flags.Bool("config", false, "second"),
			},
			want:     flags.ErrDuplicateName,
			flagName: "config",
		},
		{
			name: "duplicate shorthand",
			defs: []flags.Definition{
				flags.String("config", "", "first", flags.Shorthand("c")),
				flags.Bool("color", false, "second", flags.Shorthand("c")),
			},
			want:      flags.ErrDuplicateShorthand,
			shorthand: "c",
		},
		{
			name: "empty shorthand",
			defs: []flags.Definition{
				flags.String("config", "", "config", flags.Shorthand("")),
			},
			want:     flags.ErrInvalidShorthand,
			flagName: "config",
		},
		{
			name: "multi-rune shorthand",
			defs: []flags.Definition{
				flags.String("config", "", "config", flags.Shorthand("cc")),
			},
			want:      flags.ErrInvalidShorthand,
			flagName:  "config",
			shorthand: "cc",
		},
		{
			name: "invalid no-option default",
			defs: []flags.Definition{
				flags.Int("workers", 1, "workers", flags.NoOptionDefault("not-an-int")),
			},
			want:     flags.ErrInvalidNoOptionDefault,
			flagName: "workers",
		},
		{
			name: "nil option",
			defs: []flags.Definition{
				flags.String("config", "dev.json", "config", nil),
			},
			want:     flags.ErrInvalidDefinition,
			flagName: "config",
		},
		{
			name: "typed nil parser",
			defs: []flags.Definition{
				flags.Custom("value", flags.KindString, "", "value", flags.ParserFunc(nil)),
			},
			want:     flags.ErrInvalidDefinition,
			flagName: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := flags.NewSet(tt.defs...)
			if err == nil {
				t.Fatal("NewSet returned nil error")
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.want, err)
			}

			var definitionErr *flags.DefinitionError
			if !errors.As(err, &definitionErr) {
				t.Fatalf("error does not expose *flags.DefinitionError: %T", err)
			}
			if got := definitionErr.Name(); got != tt.flagName {
				t.Fatalf("DefinitionError.Name() = %q, want %q", got, tt.flagName)
			}
			if got := definitionErr.Shorthand(); got != tt.shorthand {
				t.Fatalf("DefinitionError.Shorthand() = %q, want %q", got, tt.shorthand)
			}
		})
	}
}

func TestSetWithReturnsIndependentSet(t *testing.T) {
	base, err := flags.NewSet(flags.String("config", "dev.json", "config"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	derived, err := base.With(flags.Bool("verbose", false, "verbose"))
	if err != nil {
		t.Fatalf("With returned unexpected error: %v", err)
	}

	if _, ok := base.Lookup("verbose"); ok {
		t.Fatal("base set includes derived definition")
	}
	if _, ok := derived.Lookup("verbose"); !ok {
		t.Fatal("derived set is missing new definition")
	}
}
