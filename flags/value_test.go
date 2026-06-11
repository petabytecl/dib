package flags_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/petabytecl/dib/flags"
)

func TestDefinitionsParseBuiltInValues(t *testing.T) {
	tests := []struct {
		name string
		def  flags.Definition
		raw  string
		want any
	}{
		{name: "string", def: flags.String("config", "", "config"), raw: "dev.json", want: "dev.json"},
		{name: "bool", def: flags.Bool("verbose", false, "verbose"), raw: "true", want: true},
		{name: "int", def: flags.Int("workers", 0, "workers"), raw: "4", want: 4},
		{name: "int64", def: flags.Int64("limit", 0, "limit"), raw: "64", want: int64(64)},
		{name: "uint", def: flags.Uint("retries", 0, "retries"), raw: "3", want: uint(3)},
		{name: "uint64", def: flags.Uint64("bytes", 0, "bytes"), raw: "1024", want: uint64(1024)},
		{name: "float64", def: flags.Float64("ratio", 0, "ratio"), raw: "0.5", want: 0.5},
		{name: "duration", def: flags.Duration("timeout", 0, "timeout"), raw: "5s", want: 5 * time.Second},
		{name: "string list", def: flags.StringList("tag", nil, "tags"), raw: "alpha", want: []string{"alpha"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.def.Parse(tt.raw)
			if err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDefinitionParseConversionErrorIsInspectable(t *testing.T) {
	def := flags.Int("workers", 1, "workers")

	_, err := def.Parse("not-an-int")
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("errors.Is(err, ErrConversion) = false; err=%v", err)
	}

	var valueErr *flags.ValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("error does not expose *flags.ValueError: %T", err)
	}
	if valueErr.Name() != "workers" || valueErr.Kind() != flags.KindInt {
		t.Fatalf("ValueError context = name %q kind %v", valueErr.Name(), valueErr.Kind())
	}
}

func TestStringListDefaultsAreDefensivelyCopied(t *testing.T) {
	defaults := []string{"alpha"}
	def := flags.StringList("tag", defaults, "tags")
	defaults[0] = "caller-mutated"

	got := def.Default().([]string)
	got[0] = "returned-mutated"

	if want := []string{"alpha"}; !reflect.DeepEqual(def.Default(), want) {
		t.Fatalf("Default() = %#v, want %#v", def.Default(), want)
	}
}

func TestCustomDefinitionsValidateKindMetadata(t *testing.T) {
	tests := []struct {
		name string
		def  flags.Definition
	}{
		{
			name: "unknown kind",
			def: flags.Custom("value", flags.Kind(99), "default", "value", flags.ParserFunc(func(raw string) (any, error) {
				return raw, nil
			})),
		},
		{
			name: "mismatched default",
			def: flags.Custom("workers", flags.KindInt, "not-an-int", "workers", flags.ParserFunc(func(raw string) (any, error) {
				return 1, nil
			})),
		},
		{
			name: "mutable default with wrong kind",
			def: flags.Custom("data", flags.KindString, map[string]string{"key": "value"}, "data", flags.ParserFunc(func(raw string) (any, error) {
				return raw, nil
			})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := flags.NewSet(tt.def)
			if err == nil {
				t.Fatal("NewSet returned nil error")
			}
			if !errors.Is(err, flags.ErrInvalidDefinition) {
				t.Fatalf("errors.Is(err, ErrInvalidDefinition) = false; err=%v", err)
			}
		})
	}
}

func TestCustomParseRejectsMismatchedResultKind(t *testing.T) {
	set, err := flags.NewSet(flags.Custom("workers", flags.KindInt, 1, "workers", flags.ParserFunc(func(raw string) (any, error) {
		return "not-an-int", nil
	})))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	def, _ := set.Lookup("workers")
	_, err = def.Parse("2")
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("errors.Is(err, ErrConversion) = false; err=%v", err)
	}
}

func TestCustomStringListParseResultIsDefensivelyCopied(t *testing.T) {
	parserResult := []string{"alpha"}
	set, err := flags.NewSet(flags.Custom("tag", flags.KindStringList, []string{}, "tags", flags.ParserFunc(func(raw string) (any, error) {
		return parserResult, nil
	})))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	def, _ := set.Lookup("tag")
	value, err := def.Parse("alpha")
	values := mustParseValue[[]string](t, value, err)
	values[0] = "returned-mutated"
	if parserResult[0] != "alpha" {
		t.Fatalf("Parse result aliases parser-owned slice: %#v", parserResult)
	}

	parserResult[0] = "parser-mutated"

	value, err = def.Parse("alpha")
	fresh := mustParseValue[[]string](t, value, err)
	if want := []string{"parser-mutated"}; !reflect.DeepEqual(fresh, want) {
		t.Fatalf("Parse() = %#v, want %#v", fresh, want)
	}
}

func mustParseValue[T any](t *testing.T, value any, err error) T {
	t.Helper()
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	typed, ok := value.(T)
	if !ok {
		t.Fatalf("Parse returned %T, want requested type", value)
	}
	return typed
}
