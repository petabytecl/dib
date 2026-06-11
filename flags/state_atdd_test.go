package flags_test

import "testing"

func TestATDDIndependentFlagSetsIgnoreAmbientProcessState(t *testing.T) {
	runConsumerContract(t, "independent sets", `package flagsconsumer_test

import (
	"os"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestIndependentFlagSetsIgnoreAmbientProcessState(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	t.Cleanup(func() { os.Args = originalArgs })

	os.Args = []string{"dib", "--config=ambient"}
	t.Setenv("DIB_CONFIG", "ambient")

	left, err := flags.NewSet(flags.String("config", "left.json", "config"))
	if err != nil {
		t.Fatalf("left NewSet returned unexpected error: %v", err)
	}
	right, err := flags.NewSet(flags.String("config", "right.json", "config"))
	if err != nil {
		t.Fatalf("right NewSet returned unexpected error: %v", err)
	}

	leftSnapshot := left.DefaultSnapshot()
	rightSnapshot := right.DefaultSnapshot()

	leftValue, ok := leftSnapshot.Lookup("config")
	if !ok {
		t.Fatal("left snapshot missing config")
	}
	rightValue, ok := rightSnapshot.Lookup("config")
	if !ok {
		t.Fatal("right snapshot missing config")
	}

	if leftValue.Explicit() || rightValue.Explicit() {
		t.Fatalf("default snapshots must not mark ambient process values explicit: left=%v right=%v", leftValue.Explicit(), rightValue.Explicit())
	}
	if got := leftValue.Default(); got != "left.json" {
		t.Fatalf("left default = %q, want left.json", got)
	}
	if got := rightValue.Default(); got != "right.json" {
		t.Fatalf("right default = %q, want right.json", got)
	}
}
`)
}

func TestATDDDerivedFlagSetsDoNotMutateOriginalsOrLeakAliases(t *testing.T) {
	runConsumerContract(t, "immutable derivation", `package flagsconsumer_test

import (
	"reflect"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestDerivedFlagSetsDoNotMutateOriginalsOrLeakAliases(t *testing.T) {
	defaultTags := []string{"alpha"}
	base, err := flags.NewSet(flags.StringList("tag", defaultTags, "tags", flags.Repeatable()))
	if err != nil {
		t.Fatalf("base NewSet returned unexpected error: %v", err)
	}

	defaultTags[0] = "caller-mutated"

	derived, err := base.With(flags.String("mode", "dev", "mode"))
	if err != nil {
		t.Fatalf("With returned unexpected error: %v", err)
	}
	if _, ok := base.Lookup("mode"); ok {
		t.Fatal("base set was mutated by With")
	}
	if _, ok := derived.Lookup("mode"); !ok {
		t.Fatal("derived set is missing mode")
	}

	baseTag, _ := base.Lookup("tag")
	if !reflect.DeepEqual(baseTag.Default(), []string{"alpha"}) {
		t.Fatalf("base tag default aliased caller slice: %#v", baseTag.Default())
	}

	returnedDefault := baseTag.Default().([]string)
	returnedDefault[0] = "returned-mutated"
	baseTagAgain, _ := base.Lookup("tag")
	if !reflect.DeepEqual(baseTagAgain.Default(), []string{"alpha"}) {
		t.Fatalf("base tag default aliased returned slice: %#v", baseTagAgain.Default())
	}

	definitions := derived.Definitions()
	definitions[0] = flags.String("other", "value", "other")
	if _, ok := derived.Lookup("other"); ok {
		t.Fatal("Definitions returned mutable storage from derived set")
	}

	snapshot := derived.DefaultSnapshot()
	tagState, _ := snapshot.Lookup("tag")
	values := tagState.Values()
	values[0] = "snapshot-mutated"

	freshSnapshot := derived.DefaultSnapshot()
	freshTagState, _ := freshSnapshot.Lookup("tag")
	if got := freshTagState.Values()[0]; got != "alpha" {
		t.Fatalf("snapshot values aliased mutable storage: %q", got)
	}
}
`)
}

func TestATDDValueAndDiagnosticFoundationIsMachineReadable(t *testing.T) {
	runConsumerContract(t, "value diagnostics", `package flagsconsumer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestValueAndDiagnosticFoundationIsMachineReadable(t *testing.T) {
	parserErr := errors.New("parser rejected value")
	parser := flags.ParserFunc(func(raw string) (any, error) {
		return nil, parserErr
	})

	set, err := flags.NewSet(flags.Custom("endpoint", flags.KindString, "http://localhost", "endpoint", parser, flags.Sensitive()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	def, ok := set.Lookup("endpoint")
	if !ok {
		t.Fatal("endpoint definition missing")
	}
	if got := def.Arity(); got != flags.ArityRequired {
		t.Fatalf("Arity() = %v, want ArityRequired", got)
	}
	if !def.Sensitive() {
		t.Fatal("custom endpoint should preserve sensitivity metadata")
	}

	_, err = def.Parse("dib_fake_secret_value")
	if err == nil {
		t.Fatal("Parse returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("conversion error does not satisfy ErrConversion: %v", err)
	}
	if !errors.Is(err, parserErr) {
		t.Fatalf("conversion error does not wrap parser error: %v", err)
	}
	if strings.Contains(err.Error(), "dib_fake_secret_value") {
		t.Fatalf("sensitive raw value leaked in error: %v", err)
	}

	var valueErr *flags.ValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("error does not expose *flags.ValueError: %T", err)
	}
	if valueErr.Name() != "endpoint" || valueErr.Kind() != flags.KindString {
		t.Fatalf("ValueError context = name %q kind %v", valueErr.Name(), valueErr.Kind())
	}

	snapshot := set.DefaultSnapshot()
	state, ok := snapshot.Lookup("endpoint")
	if !ok {
		t.Fatal("endpoint missing from snapshot")
	}
	if state.Explicit() {
		t.Fatal("default snapshot marked endpoint explicit")
	}
	if state.Arity() != flags.ArityRequired {
		t.Fatalf("snapshot arity = %v, want ArityRequired", state.Arity())
	}
}
`)
}
