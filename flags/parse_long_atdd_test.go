package flags_test

import "testing"

func TestATDDLongFlagValuesPreserveSourceAndRemainingArgs(t *testing.T) {
	runConsumerContract(t, "long flag values preserve source and remaining args", `package flagsconsumer_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func wordSeparatorNormalizer(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(name)
}

func TestLongFlagValuesPreserveSourceAndRemainingArgs(t *testing.T) {
	set, err := flags.NewNormalizedSet(
		flags.NameNormalizer(wordSeparatorNormalizer),
		flags.String("log-level", "info", "log level"),
		flags.Int("workers", 1, "worker count"),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	snapshot, err := set.Parse([]string{"pos-a", "--log_level=debug", "--workers", "4", "pos-b"})
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	logState, ok := snapshot.Lookup("log-level")
	if !ok {
		t.Fatal("snapshot missing canonical log-level state")
	}
	if got := logState.Values(); !reflect.DeepEqual(got, []any{"debug"}) {
		t.Fatalf("log-level Values() = %#v, want debug", got)
	}
	if !logState.Explicit() {
		t.Fatal("log-level state was not marked explicit")
	}
	logOccurrences := logState.Occurrences()
	if len(logOccurrences) != 1 {
		t.Fatalf("log-level occurrence count = %d, want 1", len(logOccurrences))
	}
	if got := logOccurrences[0].Spelling(); got != "--log_level" {
		t.Fatalf("log-level source spelling = %q, want --log_level", got)
	}
	if got := logOccurrences[0].Definition().Name(); got != "log-level" {
		t.Fatalf("log-level occurrence Definition().Name() = %q, want log-level", got)
	}

	workersState, ok := snapshot.Lookup("workers")
	if !ok {
		t.Fatal("snapshot missing workers state")
	}
	if got := workersState.Values(); !reflect.DeepEqual(got, []any{4}) {
		t.Fatalf("workers Values() = %#v, want 4", got)
	}
	workerOccurrences := workersState.Occurrences()
	if len(workerOccurrences) != 1 || workerOccurrences[0].Spelling() != "--workers" {
		t.Fatalf("workers occurrences = %#v, want one --workers occurrence", workerOccurrences)
	}

	gotRemaining := snapshot.RemainingArgs()
	if want := []string{"pos-a", "pos-b"}; !reflect.DeepEqual(gotRemaining, want) {
		t.Fatalf("RemainingArgs() = %#v, want %#v", gotRemaining, want)
	}
	gotRemaining[0] = "caller-mutated"
	if fresh := snapshot.RemainingArgs(); !reflect.DeepEqual(fresh, []string{"pos-a", "pos-b"}) {
		t.Fatalf("RemainingArgs returned mutable snapshot storage: %#v", fresh)
	}
}
`)
}

func TestATDDBooleanLongFlagsParsePresenceAndExplicitValues(t *testing.T) {
	runConsumerContract(t, "boolean long flags parse presence and explicit values", `package flagsconsumer_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestBooleanLongFlagsParsePresenceAndExplicitValues(t *testing.T) {
	set, err := flags.NewSet(flags.Bool("verbose", false, "verbose output"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	presence, err := set.Parse([]string{"--verbose"})
	if err != nil {
		t.Fatalf("Parse(--verbose) returned unexpected error: %v", err)
	}
	presenceState, ok := presence.Lookup("verbose")
	if !ok {
		t.Fatal("presence snapshot missing verbose state")
	}
	if got := presenceState.Values(); !reflect.DeepEqual(got, []any{true}) {
		t.Fatalf("--verbose Values() = %#v, want true", got)
	}
	if !presenceState.Explicit() {
		t.Fatal("--verbose state was not marked explicit")
	}
	if occurrences := presenceState.Occurrences(); len(occurrences) != 1 || occurrences[0].Spelling() != "--verbose" {
		t.Fatalf("--verbose occurrences = %#v, want one --verbose occurrence", occurrences)
	}

	explicitFalse, err := set.Parse([]string{"--verbose=false"})
	if err != nil {
		t.Fatalf("Parse(--verbose=false) returned unexpected error: %v", err)
	}
	falseState, ok := explicitFalse.Lookup("verbose")
	if !ok {
		t.Fatal("explicit false snapshot missing verbose state")
	}
	if got := falseState.Values(); !reflect.DeepEqual(got, []any{false}) {
		t.Fatalf("--verbose=false Values() = %#v, want false", got)
	}

	_, err = set.Parse([]string{"--verbose=maybe"})
	if err == nil {
		t.Fatal("Parse(--verbose=maybe) returned nil error")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("invalid bool error does not satisfy ErrConversion: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("invalid bool error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Token() != "--verbose" || parseErr.Name() != "verbose" {
		t.Fatalf("ParseError context token/name = %q/%q, want --verbose/verbose", parseErr.Token(), parseErr.Name())
	}
	var valueErr *flags.ValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("invalid bool error does not expose *flags.ValueError: %T", err)
	}
	if valueErr.Name() != "verbose" || valueErr.Kind() != flags.KindBool {
		t.Fatalf("ValueError context = name %q kind %v, want verbose bool", valueErr.Name(), valueErr.Kind())
	}
}
`)
}

func TestATDDUnknownLongFlagsExposeLookupContext(t *testing.T) {
	runConsumerContract(t, "unknown long flags expose lookup context", `package flagsconsumer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func wordSeparatorNormalizer(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(name)
}

func TestUnknownLongFlagsExposeLookupContext(t *testing.T) {
	set, err := flags.NewNormalizedSet(
		flags.NameNormalizer(wordSeparatorNormalizer),
		flags.String("log-level", "info", "log level"),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--unknown_flag=value"})
	if err == nil {
		t.Fatal("Parse returned nil error for unknown long flag")
	}
	if !errors.Is(err, flags.ErrUnknownFlag) {
		t.Fatalf("unknown flag error does not satisfy ErrUnknownFlag: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("unknown flag error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Token() != "--unknown_flag" {
		t.Fatalf("ParseError.Token() = %q, want --unknown_flag", parseErr.Token())
	}
	if parseErr.Name() != "unknown_flag" {
		t.Fatalf("ParseError.Name() = %q, want unknown_flag", parseErr.Name())
	}
	if parseErr.NormalizedName() != "unknown-flag" {
		t.Fatalf("ParseError.NormalizedName() = %q, want unknown-flag", parseErr.NormalizedName())
	}
	if def, ok := parseErr.Definition(); ok {
		t.Fatalf("unknown flag ParseError.Definition() = (%q, true), want false", def.Name())
	}
}
`)
}

func TestATDDMissingRequiredLongValuesAreInspectableAndLeaveSetReusable(t *testing.T) {
	runConsumerContract(t, "missing required long values are inspectable and leave set reusable", `package flagsconsumer_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestMissingRequiredLongValuesAreInspectableAndLeaveSetReusable(t *testing.T) {
	set, err := flags.NewSet(
		flags.String("config", "default.json", "config path"),
		flags.Bool("verbose", false, "verbose output"),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	for _, args := range [][]string{
		{"--config"},
		{"--config", "--verbose"},
		{"--config", "--"},
	} {
		_, err := set.Parse(args)
		if err == nil {
			t.Fatalf("Parse(%#v) returned nil error", args)
		}
		if !errors.Is(err, flags.ErrMissingValue) {
			t.Fatalf("Parse(%#v) error does not satisfy ErrMissingValue: %v", args, err)
		}
		var parseErr *flags.ParseError
		if !errors.As(err, &parseErr) {
			t.Fatalf("Parse(%#v) error does not expose *flags.ParseError: %T", args, err)
		}
		if parseErr.Token() != "--config" || parseErr.Name() != "config" {
			t.Fatalf("Parse(%#v) context token/name = %q/%q, want --config/config", args, parseErr.Token(), parseErr.Name())
		}
		def, ok := parseErr.Definition()
		if !ok || def.Name() != "config" {
			t.Fatalf("Parse(%#v) Definition() = (%q, %v), want config true", args, def.Name(), ok)
		}
	}

	defaultSnapshot := set.DefaultSnapshot()
	defaultState, ok := defaultSnapshot.Lookup("config")
	if !ok {
		t.Fatal("DefaultSnapshot missing config state")
	}
	if defaultState.Explicit() {
		t.Fatal("failed parses mutated reusable set default snapshot explicit state")
	}
	if got := defaultState.Values(); !reflect.DeepEqual(got, []any{"default.json"}) {
		t.Fatalf("default config Values() = %#v, want default.json", got)
	}

	goodSnapshot, err := set.Parse([]string{"--config=prod.json"})
	if err != nil {
		t.Fatalf("Parse after failures returned unexpected error: %v", err)
	}
	goodState, _ := goodSnapshot.Lookup("config")
	if got := goodState.Values(); !reflect.DeepEqual(got, []any{"prod.json"}) {
		t.Fatalf("config Values() after retry = %#v, want prod.json", got)
	}
}
`)
}

func TestATDDDuplicateSingleValueLongFlagsAreInspectable(t *testing.T) {
	runConsumerContract(t, "duplicate single-value long flags are inspectable", `package flagsconsumer_test

import (
	"errors"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestDuplicateSingleValueLongFlagsAreInspectable(t *testing.T) {
	set, err := flags.NewSet(flags.String("config", "default.json", "config path"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--config=first.json", "--config", "second.json"})
	if err == nil {
		t.Fatal("Parse returned nil error for duplicate single-value flag")
	}
	if !errors.Is(err, flags.ErrDuplicateValue) {
		t.Fatalf("duplicate error does not satisfy ErrDuplicateValue: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("duplicate error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Token() != "--config" || parseErr.Name() != "config" {
		t.Fatalf("duplicate context token/name = %q/%q, want --config/config", parseErr.Token(), parseErr.Name())
	}
	def, ok := parseErr.Definition()
	if !ok || def.Name() != "config" {
		t.Fatalf("duplicate Definition() = (%q, %v), want config true", def.Name(), ok)
	}
}
`)
}

func TestATDDExactAndNormalizedLongNamesParseSafely(t *testing.T) {
	runConsumerContract(t, "exact and normalized long names parse safely", `package flagsconsumer_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func wordSeparatorNormalizer(name string) string {
	if name == "l" {
		return "log-level"
	}
	return strings.NewReplacer("_", "-", ".", "-").Replace(name)
}

func TestExactAndNormalizedLongNamesParseSafely(t *testing.T) {
	exact, err := flags.NewSet(
		flags.String("log-level", "info", "hyphen"),
		flags.String("log_level", "debug", "underscore"),
		flags.String("log.level", "warn", "dot"),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	exactSnapshot, err := exact.Parse([]string{"--log-level=error", "--log_level=trace", "--log.level=warn"})
	if err != nil {
		t.Fatalf("exact Parse returned unexpected error: %v", err)
	}
	for name, want := range map[string]string{
		"log-level": "error",
		"log_level": "trace",
		"log.level": "warn",
	} {
		state, ok := exactSnapshot.Lookup(name)
		if !ok {
			t.Fatalf("exact snapshot missing %q", name)
		}
		if got := state.Values(); !reflect.DeepEqual(got, []any{want}) {
			t.Fatalf("exact %s Values() = %#v, want %q", name, got, want)
		}
	}

	normalized, err := flags.NewNormalizedSet(
		flags.NameNormalizer(wordSeparatorNormalizer),
		flags.Bool("log-level", false, "log level", flags.Shorthand("l")),
	)
	if err != nil {
		t.Fatalf("NewNormalizedSet returned unexpected error: %v", err)
	}
	normalizedSnapshot, err := normalized.Parse([]string{"--log_level"})
	if err != nil {
		t.Fatalf("normalized Parse returned unexpected error: %v", err)
	}
	state, ok := normalizedSnapshot.Lookup("log-level")
	if !ok {
		t.Fatal("normalized snapshot missing canonical log-level state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{true}) {
		t.Fatalf("normalized log-level Values() = %#v, want true", got)
	}
	occurrences := state.Occurrences()
	if len(occurrences) != 1 || occurrences[0].Spelling() != "--log_level" || occurrences[0].Definition().Name() != "log-level" {
		t.Fatalf("normalized occurrences = %#v, want --log_level mapped to canonical log-level", occurrences)
	}

	_, err = normalized.Parse([]string{"--l"})
	if err == nil {
		t.Fatal("Parse(--l) returned nil error for shorthand-only spelling")
	}
	if !errors.Is(err, flags.ErrUnknownFlag) {
		t.Fatalf("Parse(--l) error does not satisfy ErrUnknownFlag: %v", err)
	}
}
`)
}

func TestATDDNoPrefixedLongNamesAreOrdinaryNames(t *testing.T) {
	runConsumerContract(t, "no-prefixed long names are ordinary names", `package flagsconsumer_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestNoPrefixedLongNamesAreOrdinaryNames(t *testing.T) {
	registeredNoFlag, err := flags.NewSet(flags.Bool("no-color", false, "disable color"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	snapshot, err := registeredNoFlag.Parse([]string{"--no-color"})
	if err != nil {
		t.Fatalf("Parse registered --no-color returned unexpected error: %v", err)
	}
	state, ok := snapshot.Lookup("no-color")
	if !ok {
		t.Fatal("snapshot missing no-color state")
	}
	if got := state.Values(); !reflect.DeepEqual(got, []any{true}) {
		t.Fatalf("no-color Values() = %#v, want true", got)
	}

	colorOnly, err := flags.NewSet(flags.Bool("color", true, "enable color"))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	_, err = colorOnly.Parse([]string{"--no-color"})
	if err == nil {
		t.Fatal("Parse unregistered --no-color returned nil error")
	}
	if !errors.Is(err, flags.ErrUnknownFlag) {
		t.Fatalf("unregistered --no-color error does not satisfy ErrUnknownFlag: %v", err)
	}
}
`)
}

func TestATDDSensitiveConversionErrorsDoNotLeakAttachedValues(t *testing.T) {
	runConsumerContract(t, "sensitive conversion errors do not leak attached values", `package flagsconsumer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/flags"
)

func TestSensitiveConversionErrorsDoNotLeakAttachedValues(t *testing.T) {
	parserErr := errors.New("parser rejected dib_fake_secret_value")
	set, err := flags.NewSet(flags.Custom("token", flags.KindString, "", "token", flags.ParserFunc(func(raw string) (any, error) {
		return nil, parserErr
	}), flags.Sensitive()))
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	_, err = set.Parse([]string{"--token=dib_fake_secret_value"})
	if err == nil {
		t.Fatal("Parse returned nil error for sensitive conversion failure")
	}
	if !errors.Is(err, flags.ErrConversion) {
		t.Fatalf("sensitive conversion error does not satisfy ErrConversion: %v", err)
	}
	if errors.Is(err, parserErr) {
		t.Fatalf("sensitive conversion error exposes parser cause: %v", err)
	}
	if strings.Contains(err.Error(), "dib_fake_secret_value") {
		t.Fatalf("sensitive raw value leaked in error: %v", err)
	}
	var parseErr *flags.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("sensitive conversion error does not expose *flags.ParseError: %T", err)
	}
	if parseErr.Token() != "--token" || parseErr.Name() != "token" {
		t.Fatalf("sensitive ParseError token/name = %q/%q, want --token/token", parseErr.Token(), parseErr.Name())
	}
	if strings.Contains(parseErr.Token(), "dib_fake_secret_value") {
		t.Fatalf("ParseError.Token leaked sensitive value: %q", parseErr.Token())
	}
}
`)
}
