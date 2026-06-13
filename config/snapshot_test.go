package config_test

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/petabytecl/dib/config"
)

func TestDefaultSnapshotResolvesDefaultsAndNotFound(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("empty-string", "", "empty string"),
		config.Bool("disabled", false, "disabled"),
		config.Int("zero", 0, "zero"),
		config.StringList("empty-list", []string{}, "empty list"),
		config.Define("no-default", config.KindString, "no default"),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	snapshot := set.DefaultSnapshot()
	tests := []struct {
		key      string
		want     any
		hasValue bool
	}{
		{key: "empty-string", want: "", hasValue: true},
		{key: "disabled", want: false, hasValue: true},
		{key: "zero", want: 0, hasValue: true},
		{key: "empty-list", want: []string{}, hasValue: true},
		{key: "no-default", hasValue: false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			value, ok := snapshot.Lookup(tt.key)
			if !ok {
				t.Fatalf("Lookup(%q) returned false", tt.key)
			}
			got, hasValue := value.Value()
			if hasValue != tt.hasValue {
				t.Fatalf("Value() presence = %v, want %v", hasValue, tt.hasValue)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Value() = %#v, want %#v", got, tt.want)
			}
			if hasValue {
				if got := value.Provenance(); got != config.SourceDefault {
					t.Fatalf("Provenance() = %q, want %q", got, config.SourceDefault)
				}
			} else if got := value.Provenance(); got != "" {
				t.Fatalf("Provenance() = %q, want empty for no-default value", got)
			}
		})
	}

	if value, ok := snapshot.Lookup("missing"); ok {
		t.Fatalf("Lookup(missing) = (%#v, true), want false", value)
	}
}

func TestDefaultSnapshotIsReusableAndDefensive(t *testing.T) {
	t.Parallel()

	defaultTags := []string{"alpha"}
	set, err := config.NewSet(config.StringList("tags", defaultTags, "tag list"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	defaultTags[0] = "mutated"

	snapshot := set.DefaultSnapshot()
	value, ok := snapshot.Lookup("tags")
	if !ok {
		t.Fatal("Lookup(tags) returned false")
	}
	tags, hasValue := value.Value()
	if !hasValue {
		t.Fatal("Value() returned hasValue=false")
	}
	tags.([]string)[0] = "changed"

	again, ok := snapshot.Lookup("tags")
	if !ok {
		t.Fatal("second Lookup(tags) returned false")
	}
	got, _ := again.Value()
	if !reflect.DeepEqual(got, []string{"alpha"}) {
		t.Fatalf("snapshot leaked mutable value alias: %#v", got)
	}

	def, ok := again.Definition()
	if !ok {
		t.Fatal("Definition() returned ok=false")
	}
	defDefault, _ := def.Default()
	if !reflect.DeepEqual(defDefault, []string{"alpha"}) {
		t.Fatalf("Definition default mutated during snapshot resolution: %#v", defDefault)
	}
}

func TestDefaultSnapshotUsesSetNormalizer(t *testing.T) {
	t.Parallel()

	normalizeSeparators := config.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})
	set, err := config.NewNormalizedSet(normalizeSeparators, config.String("log-level", "info", "log level"))
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	snapshot := set.DefaultSnapshot()
	value, ok := snapshot.Lookup("log_level")
	if !ok {
		t.Fatal("Lookup(log_level) returned false")
	}
	got, hasValue := value.Value()
	if !hasValue || got != "info" {
		t.Fatalf("Value() = %#v, %v; want info, true", got, hasValue)
	}
}

func TestDefaultSnapshotConcurrentReuse(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(config.StringList("tags", []string{"alpha", "beta"}, "tag list"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	snapshot := set.DefaultSnapshot()

	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, ok := snapshot.Lookup("tags")
			if !ok {
				t.Error("Lookup(tags) returned false")
				return
			}
			got, ok := value.Value()
			if !ok || !reflect.DeepEqual(got, []string{"alpha", "beta"}) {
				t.Errorf("Value() = %#v, %v; want alpha/beta default", got, ok)
				return
			}
			got.([]string)[0] = "local mutation"
		}()
	}
	wg.Wait()
}
