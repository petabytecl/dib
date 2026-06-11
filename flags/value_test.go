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
		{name: "string list", def: flags.StringList("tag", nil, "tags"), raw: "alpha", want: "alpha"},
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
