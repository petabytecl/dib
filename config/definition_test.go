package config_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/petabytecl/dib/config"
)

func TestDefinitionsExposeMetadataAndDefensiveDefaults(t *testing.T) {
	t.Parallel()

	tagsDefault := []string{"alpha", "beta"}
	defs := []struct {
		def     config.Definition
		name    string
		kind    config.Kind
		usage   string
		want    any
		hasWant bool
	}{
		{def: config.String("endpoint", "http://localhost", "service endpoint"), name: "endpoint", kind: config.KindString, usage: "service endpoint", want: "http://localhost", hasWant: true},
		{def: config.Bool("debug", false, "debug mode"), name: "debug", kind: config.KindBool, usage: "debug mode", want: false, hasWant: true},
		{def: config.Int("workers", 4, "worker count"), name: "workers", kind: config.KindInt, usage: "worker count", want: 4, hasWant: true},
		{def: config.Int64("limit64", int64(64), "limit"), name: "limit64", kind: config.KindInt64, usage: "limit", want: int64(64), hasWant: true},
		{def: config.Uint("retries", uint(3), "retry count"), name: "retries", kind: config.KindUint, usage: "retry count", want: uint(3), hasWant: true},
		{def: config.Uint64("bytes", uint64(1024), "byte limit"), name: "bytes", kind: config.KindUint64, usage: "byte limit", want: uint64(1024), hasWant: true},
		{def: config.Float64("ratio", 0.5, "ratio"), name: "ratio", kind: config.KindFloat64, usage: "ratio", want: 0.5, hasWant: true},
		{def: config.Duration("timeout", 5*time.Second, "timeout"), name: "timeout", kind: config.KindDuration, usage: "timeout", want: 5 * time.Second, hasWant: true},
		{def: config.StringList("tags", tagsDefault, "tag list"), name: "tags", kind: config.KindStringList, usage: "tag list", want: []string{"alpha", "beta"}, hasWant: true},
		{def: config.Define("token", config.KindString, "api token", config.Sensitive()), name: "token", kind: config.KindString, usage: "api token"},
	}
	tagsDefault[0] = "mutated"

	for _, tt := range defs {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.def.Name(); got != tt.name {
				t.Fatalf("Name() = %q, want %q", got, tt.name)
			}
			if got := tt.def.Kind(); got != tt.kind {
				t.Fatalf("Kind() = %v, want %v", got, tt.kind)
			}
			if got := tt.def.Usage(); got != tt.usage {
				t.Fatalf("Usage() = %q, want %q", got, tt.usage)
			}
			gotDefault, hasDefault := tt.def.Default()
			if hasDefault != tt.hasWant {
				t.Fatalf("Default() presence = %v, want %v", hasDefault, tt.hasWant)
			}
			if !reflect.DeepEqual(gotDefault, tt.want) {
				t.Fatalf("Default() = %#v, want %#v", gotDefault, tt.want)
			}
		})
	}

	token := defs[len(defs)-1].def
	if !token.Sensitive() {
		t.Fatal("Sensitive() = false, want true")
	}

	tags := defs[8].def
	firstDefault, _ := tags.Default()
	firstDefault.([]string)[0] = "changed"
	secondDefault, _ := tags.Default()
	if !reflect.DeepEqual(secondDefault, []string{"alpha", "beta"}) {
		t.Fatalf("Default() leaked mutable slice alias: %#v", secondDefault)
	}
}

func TestKindStringVocabulary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind config.Kind
		want string
	}{
		{kind: config.KindString, want: "string"},
		{kind: config.KindBool, want: "bool"},
		{kind: config.KindInt, want: "int"},
		{kind: config.KindInt64, want: "int64"},
		{kind: config.KindUint, want: "uint"},
		{kind: config.KindUint64, want: "uint64"},
		{kind: config.KindFloat64, want: "float64"},
		{kind: config.KindDuration, want: "duration"},
		{kind: config.KindStringList, want: "string-list"},
		{kind: config.Kind(99), want: "unknown"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Fatalf("%v.String() = %q, want %q", int(tt.kind), got, tt.want)
		}
	}
}
