package flags_test

import "testing"

func TestATDDFlagDefinitionsExposeMetadata(t *testing.T) {
	t.Skip("ATDD RED: remove this skip during Story 2.1 implementation, then make the public flags API satisfy the contract")

	runConsumerContract(t, "definition metadata", `package flagsconsumer_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/petabytecl/dib/flags"
)

func TestDefinitionsExposeMetadata(t *testing.T) {
	set, err := flags.NewSet(
		flags.String("config", "dev.json", "configuration file", flags.Shorthand("c"), flags.Hidden(), flags.Deprecated("use config-file"), flags.Sensitive()),
		flags.Bool("verbose", false, "enable verbose logging", flags.Shorthand("v"), flags.NoOptionDefault(true)),
		flags.Int("workers", 2, "worker count"),
		flags.Int64("limit64", int64(64), "limit"),
		flags.Uint("retries", uint(3), "retry count"),
		flags.Uint64("bytes", uint64(1024), "byte limit"),
		flags.Float64("ratio", 0.5, "ratio"),
		flags.Duration("timeout", 5*time.Second, "timeout"),
		flags.StringList("tag", []string{"alpha"}, "tags", flags.Repeatable()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	cases := []struct {
		name string
		kind flags.Kind
	}{
		{name: "config", kind: flags.KindString},
		{name: "verbose", kind: flags.KindBool},
		{name: "workers", kind: flags.KindInt},
		{name: "limit64", kind: flags.KindInt64},
		{name: "retries", kind: flags.KindUint},
		{name: "bytes", kind: flags.KindUint64},
		{name: "ratio", kind: flags.KindFloat64},
		{name: "timeout", kind: flags.KindDuration},
		{name: "tag", kind: flags.KindStringList},
	}

	if got := set.Len(); got != len(cases) {
		t.Fatalf("Len() = %d, want %d", got, len(cases))
	}
	if got := len(set.Definitions()); got != len(cases) {
		t.Fatalf("Definitions() length = %d, want %d", got, len(cases))
	}

	for _, tt := range cases {
		def, ok := set.Lookup(tt.name)
		if !ok {
			t.Fatalf("Lookup(%q) returned false", tt.name)
		}
		if got := def.Kind(); got != tt.kind {
			t.Fatalf("%s Kind() = %v, want %v", tt.name, got, tt.kind)
		}
		if def.Usage() == "" {
			t.Fatalf("%s Usage() is empty", tt.name)
		}
	}

	config, _ := set.Lookup("config")
	shorthand, ok := config.Shorthand()
	if !ok || shorthand != "c" {
		t.Fatalf("config shorthand = %q, %v; want c, true", shorthand, ok)
	}
	if !config.Hidden() || !config.Sensitive() || config.Deprecated() != "use config-file" {
		t.Fatalf("config metadata not preserved: hidden=%v sensitive=%v deprecated=%q", config.Hidden(), config.Sensitive(), config.Deprecated())
	}

	verbose, _ := set.Lookup("verbose")
	noOptionDefault, ok := verbose.NoOptionDefault()
	if !ok || noOptionDefault != true {
		t.Fatalf("verbose no-option default = %#v, %v; want true, true", noOptionDefault, ok)
	}

	tag, _ := set.Lookup("tag")
	if got := tag.RepeatPolicy(); got != flags.RepeatAccumulated {
		t.Fatalf("tag RepeatPolicy() = %v, want RepeatAccumulated", got)
	}
	if !reflect.DeepEqual(tag.Default(), []string{"alpha"}) {
		t.Fatalf("tag Default() = %#v, want []string{alpha}", tag.Default())
	}

	timeout, _ := set.Lookup("timeout")
	if got := timeout.Default(); got != 5*time.Second {
		t.Fatalf("timeout Default() = %#v, want 5s", got)
	}
}
`)
}

func TestATDDFlagSetValidationErrorsAreInspectable(t *testing.T) {
	t.Skip("ATDD RED: remove this skip during Story 2.1 implementation, then make validation expose typed errors")

	runConsumerContract(t, "validation errors", `package flagsconsumer_test

import (
	"errors"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestValidationErrorsAreInspectable(t *testing.T) {
	tests := []struct {
		name      string
		build     func() error
		sentinel  error
		flagName  string
		shorthand string
	}{
		{
			name: "empty long name",
			build: func() error {
				_, err := flags.NewSet(flags.String("", "", "missing name"))
				return err
			},
			sentinel: flags.ErrInvalidDefinition,
		},
		{
			name: "duplicate long name",
			build: func() error {
				_, err := flags.NewSet(flags.String("config", "", "first"), flags.Bool("config", false, "second"))
				return err
			},
			sentinel: flags.ErrDuplicateName,
			flagName: "config",
		},
		{
			name: "duplicate shorthand",
			build: func() error {
				_, err := flags.NewSet(flags.String("config", "", "first", flags.Shorthand("c")), flags.Bool("color", false, "second", flags.Shorthand("c")))
				return err
			},
			sentinel: flags.ErrDuplicateShorthand,
			shorthand: "c",
		},
		{
			name: "invalid shorthand",
			build: func() error {
				_, err := flags.NewSet(flags.String("config", "", "config", flags.Shorthand("cc")))
				return err
			},
			sentinel: flags.ErrInvalidShorthand,
			flagName: "config",
			shorthand: "cc",
		},
		{
			name: "invalid no-option default",
			build: func() error {
				_, err := flags.NewSet(flags.Int("workers", 1, "workers", flags.NoOptionDefault("not-an-int")))
				return err
			},
			sentinel: flags.ErrInvalidNoOptionDefault,
			flagName: "workers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.build()
			if err == nil {
				t.Fatal("NewSet returned nil error")
			}
			if !errors.Is(err, tt.sentinel) {
				t.Fatalf("errors.Is(%T, %T) = false; err=%v", err, tt.sentinel, err)
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
`)
}
