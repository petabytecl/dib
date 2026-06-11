package flags_test

import "testing"

func TestATDDExactFlagNamesRemainDistinctByDefault(t *testing.T) {
	t.Skip("ATDD RED: remove this skip during Story 2.2 implementation to verify exact matching remains the default")

	runConsumerContract(t, "exact names by default", `package flagsconsumer_test

import (
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestExactFlagNamesRemainDistinctByDefault(t *testing.T) {
	set, err := flags.NewSet(
		flags.String("log-level", "info", "hyphenated log level"),
		flags.String("log_level", "debug", "underscored log level"),
		flags.String("log.level", "warn", "dotted log level"),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	for _, name := range []string{"log-level", "log_level", "log.level"} {
		def, ok := set.Lookup(name)
		if !ok {
			t.Fatalf("Lookup(%q) returned false", name)
		}
		if got := def.Name(); got != name {
			t.Fatalf("Lookup(%q).Name() = %q, want %q", name, got, name)
		}

		state, ok := set.DefaultSnapshot().Lookup(name)
		if !ok {
			t.Fatalf("DefaultSnapshot().Lookup(%q) returned false", name)
		}
		if got := state.Values()[0]; got == "" {
			t.Fatalf("snapshot value for %q is empty", name)
		}
	}
}
`)
}

func TestATDDConfiguredNormalizerResolvesCanonicalDefinitions(t *testing.T) {
	t.Skip("ATDD RED: remove this skip during Story 2.2 implementation after adding the explicit normalizer API")

	runConsumerContract(t, "configured normalizer resolves canonical definitions", `package flagsconsumer_test

import (
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func wordSeparatorNormalizer(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(name)
}

func TestConfiguredNormalizerResolvesCanonicalDefinitions(t *testing.T) {
	set, err := flags.NewNormalizedSet(
		flags.NameNormalizer(wordSeparatorNormalizer),
		flags.String("log-level", "info", "log level"),
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
			t.Fatalf("Lookup(%q).Name() = %q, want canonical log-level", spelling, got)
		}
	}

	snapshot := set.DefaultSnapshot()
	state, ok := snapshot.Lookup("log-level")
	if !ok {
		t.Fatal("DefaultSnapshot missing canonical log-level state")
	}
	if got := state.Default(); got != "info" {
		t.Fatalf("canonical snapshot default = %q, want info", got)
	}
}
`)
}

func TestATDDNormalizationCollisionsAreInspectable(t *testing.T) {
	t.Skip("ATDD RED: remove this skip during Story 2.2 implementation after adding typed normalization collision diagnostics")

	runConsumerContract(t, "normalization collisions are inspectable", `package flagsconsumer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func wordSeparatorNormalizer(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(name)
}

func TestNormalizationCollisionsAreInspectable(t *testing.T) {
	_, err := flags.NewNormalizedSet(
		flags.NameNormalizer(wordSeparatorNormalizer),
		flags.String("log-level", "info", "first"),
		flags.String("log_level", "debug", "second"),
	)
	if err == nil {
		t.Fatal("NewNormalizedSet returned nil error for normalized name collision")
	}
	if !errors.Is(err, flags.ErrDuplicateNormalizedName) {
		t.Fatalf("errors.Is(err, ErrDuplicateNormalizedName) = false; err=%v", err)
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
`)
}

func TestATDDNormalizedDerivationDoesNotMutateOriginalSets(t *testing.T) {
	t.Skip("ATDD RED: remove this skip during Story 2.2 implementation after normalized derivation preserves immutability")

	runConsumerContract(t, "normalized derivation does not mutate originals", `package flagsconsumer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func wordSeparatorNormalizer(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(name)
}

func TestNormalizedDerivationDoesNotMutateOriginalSets(t *testing.T) {
	base, err := flags.NewSet(flags.String("log-level", "info", "log level"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	normalized, err := base.WithNormalizer(flags.NameNormalizer(wordSeparatorNormalizer))
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
		t.Fatal("With mutated the normalized source set")
	}
	if _, ok := derived.Lookup("verbose"); !ok {
		t.Fatal("derived normalized set is missing verbose")
	}
	if def, ok := derived.Lookup("log.level"); !ok || def.Name() != "log-level" {
		t.Fatalf("derived Lookup(log.level) = (%q, %v), want canonical log-level", def.Name(), ok)
	}

	_, err = normalized.With(flags.String("log_level", "debug", "collision"))
	if !errors.Is(err, flags.ErrDuplicateNormalizedName) {
		t.Fatalf("With normalized collision error = %v, want ErrDuplicateNormalizedName", err)
	}
}
`)
}

func TestATDDLongNameNormalizationDoesNotCreateShorthandAliases(t *testing.T) {
	t.Skip("ATDD RED: remove this skip during Story 2.2 implementation after shorthand and long-name indexes remain independent")

	runConsumerContract(t, "normalization does not create shorthand aliases", `package flagsconsumer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func wordSeparatorNormalizer(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(name)
}

func TestLongNameNormalizationDoesNotCreateShorthandAliases(t *testing.T) {
	set, err := flags.NewNormalizedSet(
		flags.NameNormalizer(wordSeparatorNormalizer),
		flags.Bool("log-level", false, "log level", flags.Shorthand("l")),
		flags.Bool("lookup", false, "lookup", flags.Shorthand("k")),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}
	if _, ok := set.Lookup("l"); ok {
		t.Fatal("long-name Lookup(l) resolved a shorthand alias")
	}
	if def, ok := set.Lookup("log_level"); !ok || def.Name() != "log-level" {
		t.Fatalf("Lookup(log_level) = (%q, %v), want canonical log-level", def.Name(), ok)
	}

	_, err = flags.NewNormalizedSet(
		flags.NameNormalizer(wordSeparatorNormalizer),
		flags.Bool("log-level", false, "log level", flags.Shorthand("l")),
		flags.Bool("other", false, "other", flags.Shorthand("l")),
	)
	if !errors.Is(err, flags.ErrDuplicateShorthand) {
		t.Fatalf("duplicate shorthand error = %v, want ErrDuplicateShorthand", err)
	}
	if errors.Is(err, flags.ErrDuplicateNormalizedName) {
		t.Fatalf("duplicate shorthand should not be reported as a normalized long-name collision: %v", err)
	}
}
`)
}
