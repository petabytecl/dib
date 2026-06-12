package config_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/config"
)

func TestQAConfigPublicWorkflowCoversDefaultsNormalizationAndNotFound(t *testing.T) {
	t.Parallel()

	normalizeSeparators := config.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})

	set, err := config.NewNormalizedSet(
		normalizeSeparators,
		config.String("log-level", "info", "log level"),
		config.Bool("debug", false, "debug mode"),
		config.StringList("tags", []string{"alpha", "beta"}, "tag list"),
		config.Define("token", config.KindString, "api token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	def, ok := set.Lookup("log_level")
	if !ok {
		t.Fatal("Lookup(log_level) returned false")
	}
	if got := def.Name(); got != "log-level" {
		t.Fatalf("Lookup(log_level).Name() = %q, want log-level", got)
	}

	snapshot := set.DefaultSnapshot()
	logLevel, ok := snapshot.Lookup("log.level")
	if !ok {
		t.Fatal("snapshot.Lookup(log.level) returned false")
	}
	gotLogLevel, hasLogLevel := logLevel.Value()
	if !hasLogLevel || gotLogLevel != "info" {
		t.Fatalf("log-level Value() = %#v, %v; want info, true", gotLogLevel, hasLogLevel)
	}
	if got := logLevel.Provenance(); got != config.SourceDefault {
		t.Fatalf("log-level Provenance() = %q, want %q", got, config.SourceDefault)
	}

	debug, ok := snapshot.Lookup("debug")
	if !ok {
		t.Fatal("snapshot.Lookup(debug) returned false")
	}
	gotDebug, hasDebug := debug.Value()
	if !hasDebug || gotDebug != false {
		t.Fatalf("debug Value() = %#v, %v; want false, true", gotDebug, hasDebug)
	}

	tags, ok := snapshot.Lookup("tags")
	if !ok {
		t.Fatal("snapshot.Lookup(tags) returned false")
	}
	gotTags, hasTags := tags.Value()
	if !hasTags || !reflect.DeepEqual(gotTags, []string{"alpha", "beta"}) {
		t.Fatalf("tags Value() = %#v, %v; want alpha/beta, true", gotTags, hasTags)
	}
	gotTags.([]string)[0] = "mutated"
	gotTagsAgain, _ := tags.Value()
	if !reflect.DeepEqual(gotTagsAgain, []string{"alpha", "beta"}) {
		t.Fatalf("Value() leaked mutable slice alias: %#v", gotTagsAgain)
	}

	token, ok := snapshot.Lookup("token")
	if !ok {
		t.Fatal("snapshot.Lookup(token) returned false for registered no-default key")
	}
	if got, hasValue := token.Value(); hasValue || got != nil {
		t.Fatalf("token Value() = %#v, %v; want nil, false", got, hasValue)
	}
	if got := token.Provenance(); got != "" {
		t.Fatalf("token Provenance() = %q, want empty for no-default key", got)
	}
	tokenDef, ok := token.Definition()
	if !ok {
		t.Fatal("token Definition() returned false")
	}
	if !tokenDef.Sensitive() {
		t.Fatal("token Definition().Sensitive() = false, want true")
	}

	if missing, ok := snapshot.Lookup("missing"); ok {
		t.Fatalf("snapshot.Lookup(missing) = (%#v, true), want false", missing)
	}
}

func TestQAConfigSetupErrorsCoverUnknownKindsAndNormalizedCollisions(t *testing.T) {
	t.Parallel()

	_, err := config.NewSet(config.Define("mode", config.Kind(99), "mode"))
	if err == nil {
		t.Fatal("NewSet returned nil error for unknown kind")
	}
	if !errors.Is(err, config.ErrInvalidDefinition) {
		t.Fatalf("errors.Is(err, ErrInvalidDefinition) = false; err=%v", err)
	}
	var definitionErr *config.DefinitionError
	if !errors.As(err, &definitionErr) {
		t.Fatalf("error does not expose *config.DefinitionError: %T", err)
	}
	if got := definitionErr.Key(); got != "mode" {
		t.Fatalf("DefinitionError.Key() = %q, want mode", got)
	}
	if got := definitionErr.Kind(); got != config.Kind(99) {
		t.Fatalf("DefinitionError.Kind() = %v, want unknown kind", got)
	}

	normalizeSeparators := config.NameNormalizer(func(name string) string {
		return strings.NewReplacer("_", "-", ".", "-").Replace(name)
	})
	_, err = config.NewSet(
		config.String("log-level", "info", "hyphen"),
		config.String("log_level", "debug", "underscore"),
	)
	if err != nil {
		t.Fatalf("NewSet exact spellings returned unexpected error: %v", err)
	}
	_, err = config.NewNormalizedSet(
		normalizeSeparators,
		config.String("log-level", "info", "hyphen"),
		config.String("log_level", "debug", "underscore"),
	)
	if err == nil {
		t.Fatal("NewNormalizedSet returned nil error for normalized collision")
	}
	if !errors.Is(err, config.ErrDuplicateNormalizedKey) {
		t.Fatalf("errors.Is(err, ErrDuplicateNormalizedKey) = false; err=%v", err)
	}
	if !errors.As(err, &definitionErr) {
		t.Fatalf("error does not expose *config.DefinitionError: %T", err)
	}
	if got := definitionErr.NormalizedKey(); got != "log-level" {
		t.Fatalf("DefinitionError.NormalizedKey() = %q, want log-level", got)
	}
}
