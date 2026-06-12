package config_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/petabytecl/dib/config"
)

// --- Typed getter value tests ---

func TestGetStringReturnedValueAndProvenance(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.String("log-level", "info", "level"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "log-level", Value: "debug"})
	got, err := snap.GetString("log-level")
	if err != nil {
		t.Fatalf("GetString returned unexpected error: %v", err)
	}
	if got != "debug" {
		t.Fatalf("GetString() = %q, want debug", got)
	}
	v, ok := snap.Lookup("log-level")
	if !ok {
		t.Fatal("Lookup returned false")
	}
	if v.Provenance() != config.SourceExplicit {
		t.Fatalf("Provenance() = %q, want %q", v.Provenance(), config.SourceExplicit)
	}
}

func TestGetBoolReturnedValue(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Bool("debug", false, "debug"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "debug", Value: true})
	got, err := snap.GetBool("debug")
	if err != nil {
		t.Fatalf("GetBool returned unexpected error: %v", err)
	}
	if !got {
		t.Fatal("GetBool() = false, want true")
	}
}

func TestGetIntReturnedValue(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Int("workers", 1, "workers"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "workers", Value: 4})
	got, err := snap.GetInt("workers")
	if err != nil {
		t.Fatalf("GetInt returned unexpected error: %v", err)
	}
	if got != 4 {
		t.Fatalf("GetInt() = %d, want 4", got)
	}
}

func TestGetInt64ReturnedValue(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Int64("limit", 0, "limit"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "limit", Value: int64(999)})
	got, err := snap.GetInt64("limit")
	if err != nil {
		t.Fatalf("GetInt64 returned unexpected error: %v", err)
	}
	if got != 999 {
		t.Fatalf("GetInt64() = %d, want 999", got)
	}
}

func TestGetUintReturnedValue(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Uint("retries", 0, "retries"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "retries", Value: uint(3)})
	got, err := snap.GetUint("retries")
	if err != nil {
		t.Fatalf("GetUint returned unexpected error: %v", err)
	}
	if got != 3 {
		t.Fatalf("GetUint() = %d, want 3", got)
	}
}

func TestGetUint64ReturnedValue(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Uint64("bytes", 0, "bytes"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "bytes", Value: uint64(1024)})
	got, err := snap.GetUint64("bytes")
	if err != nil {
		t.Fatalf("GetUint64 returned unexpected error: %v", err)
	}
	if got != 1024 {
		t.Fatalf("GetUint64() = %d, want 1024", got)
	}
}

func TestGetFloat64ReturnedValue(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Float64("ratio", 0, "ratio"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "ratio", Value: 1.5})
	got, err := snap.GetFloat64("ratio")
	if err != nil {
		t.Fatalf("GetFloat64 returned unexpected error: %v", err)
	}
	if got != 1.5 {
		t.Fatalf("GetFloat64() = %f, want 1.5", got)
	}
}

func TestGetDurationReturnedValue(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Duration("timeout", time.Second, "timeout"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "timeout", Value: 250 * time.Millisecond})
	got, err := snap.GetDuration("timeout")
	if err != nil {
		t.Fatalf("GetDuration returned unexpected error: %v", err)
	}
	if got != 250*time.Millisecond {
		t.Fatalf("GetDuration() = %v, want 250ms", got)
	}
}

func TestGetStringListReturnedValue(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.StringList("tags", []string{"a", "b"}, "tags"))
	snap := set.DefaultSnapshot()
	got, err := snap.GetStringList("tags")
	if err != nil {
		t.Fatalf("GetStringList returned unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("GetStringList() = %v, want [a b]", got)
	}
}

func TestGetStringListReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.StringList("tags", []string{"x", "y"}, "tags"))
	snap := set.DefaultSnapshot()
	got, err := snap.GetStringList("tags")
	if err != nil {
		t.Fatalf("GetStringList returned unexpected error: %v", err)
	}
	got[0] = "mutated"
	got2, err := snap.GetStringList("tags")
	if err != nil {
		t.Fatalf("second GetStringList returned unexpected error: %v", err)
	}
	if got2[0] != "x" {
		t.Fatalf("GetStringList defensive copy broken: second call returned %v, want original", got2)
	}
}

// --- Error path tests ---

func TestGetKindMismatchReturnsGetConversionError(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Bool("debug", false, "debug"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "debug", Value: true})
	_, err := snap.GetString("debug")
	if err == nil {
		t.Fatal("GetString on KindBool key returned nil error")
	}
	if !errors.Is(err, config.ErrGetConversion) {
		t.Fatalf("errors.Is(err, ErrGetConversion) = false; err=%v", err)
	}
	var getErr *config.GetError
	if !errors.As(err, &getErr) {
		t.Fatalf("error does not expose *config.GetError: %T", err)
	}
	if got := getErr.Key(); got != "debug" {
		t.Fatalf("GetError.Key() = %q, want debug", got)
	}
	if got := getErr.Kind(); got != config.KindBool {
		t.Fatalf("GetError.Kind() = %v, want KindBool", got)
	}
	if got := getErr.WantKind(); got != config.KindString {
		t.Fatalf("GetError.WantKind() = %v, want KindString", got)
	}
}

func TestGetAbsentRegisteredKeyReturnsKeyAbsentError(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Define("token", config.KindString, "token"))
	snap := set.DefaultSnapshot()
	_, err := snap.GetString("token")
	if err == nil {
		t.Fatal("GetString on absent key returned nil error")
	}
	if !errors.Is(err, config.ErrKeyAbsent) {
		t.Fatalf("errors.Is(err, ErrKeyAbsent) = false; err=%v", err)
	}
}

func TestGetUnregisteredKeyReturnsKeyNotFoundError(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.String("log-level", "info", "level"))
	snap := set.DefaultSnapshot()
	_, err := snap.GetString("nonexistent")
	if err == nil {
		t.Fatal("GetString on unregistered key returned nil error")
	}
	if !errors.Is(err, config.ErrKeyNotFound) {
		t.Fatalf("errors.Is(err, ErrKeyNotFound) = false; err=%v", err)
	}
}

// --- IsSet tests ---

func TestIsSetDistinguishesAbsentFromZero(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(
		config.Define("absent", config.KindInt, "no default"),
		config.Int("zero", 0, "explicit zero default"),
	)
	snap := set.DefaultSnapshot()
	if snap.IsSet("absent") {
		t.Fatal("IsSet(absent) = true, want false for no-default key")
	}
	if !snap.IsSet("zero") {
		t.Fatal("IsSet(zero) = false, want true for zero-default key")
	}
}

func TestIsSetReturnsTrueForExplicitZeroInt(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Define("count", config.KindInt, "count"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "count", Value: 0})
	if !snap.IsSet("count") {
		t.Fatal("IsSet(count) = false, want true for explicitly set zero int")
	}
}

func TestIsSetReturnsTrueForEmptyStringFromExplicit(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.String("name", "default", "name"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "name", Value: ""})
	if !snap.IsSet("name") {
		t.Fatal("IsSet(name) = false, want true for explicitly set empty string")
	}
}

func TestIsSetReturnsFalseForAbsentKey(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Define("token", config.KindString, "token"))
	snap := set.DefaultSnapshot()
	if snap.IsSet("token") {
		t.Fatal("IsSet(token) = true, want false for registered key with no default and no source")
	}
}

func TestIsSetReturnsFalseForUnregisteredKey(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.String("log-level", "info", "level"))
	snap := set.DefaultSnapshot()
	if snap.IsSet("nonexistent") {
		t.Fatal("IsSet(nonexistent) = true, want false for unregistered key")
	}
}

// --- Source and provenance tests ---

func TestGetValueFromDefaultSource(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Int("workers", 5, "workers"))
	snap := set.DefaultSnapshot()
	got, err := snap.GetInt("workers")
	if err != nil {
		t.Fatalf("GetInt returned unexpected error: %v", err)
	}
	if got != 5 {
		t.Fatalf("GetInt() = %d, want 5", got)
	}
	v, _ := snap.Lookup("workers")
	if v.Provenance() != config.SourceDefault {
		t.Fatalf("Provenance() = %q, want %q", v.Provenance(), config.SourceDefault)
	}
}

func TestGetValueFromAllSources(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(config.String("key", "default-val", "test key"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	tests := []struct {
		name       string
		snap       func() config.Snapshot
		wantSource string
		wantValue  string
	}{
		{
			name: config.SourceExplicit,
			snap: func() config.Snapshot {
				s, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "key", Value: "explicit-val"})
				return s
			},
			wantSource: config.SourceExplicit,
			wantValue:  "explicit-val",
		},
		{
			name: config.SourceFlagBinding,
			snap: func() config.Snapshot {
				s, _ := config.NewFlagSnapshot(set, []config.FlagValue{
					{ConfigKey: "key", ExplicitlySet: true, Value: "flag-val"},
				})
				return s
			},
			wantSource: config.SourceFlagBinding,
			wantValue:  "flag-val",
		},
		{
			name: config.SourceEnv,
			snap: func() config.Snapshot {
				s, _ := config.NewEnvSnapshot(set,
					mapLookup(map[string]string{"KEY": "env-val"}),
					[]config.EnvBinding{config.BindEnv("key", "KEY")},
				)
				return s
			},
			wantSource: config.SourceEnv,
			wantValue:  "env-val",
		},
		{
			name: config.SourceJSON,
			snap: func() config.Snapshot {
				s, _ := config.LoadJSON(set, strings.NewReader(`{"key":"json-val"}`))
				return s
			},
			wantSource: config.SourceJSON,
			wantValue:  "json-val",
		},
		{
			name:       config.SourceDefault,
			snap:       func() config.Snapshot { return set.DefaultSnapshot() },
			wantSource: config.SourceDefault,
			wantValue:  "default-val",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			snap := tt.snap()
			got, err := snap.GetString("key")
			if err != nil {
				t.Fatalf("GetString returned unexpected error: %v", err)
			}
			if got != tt.wantValue {
				t.Fatalf("GetString() = %q, want %q", got, tt.wantValue)
			}
			v, _ := snap.Lookup("key")
			if v.Provenance() != tt.wantSource {
				t.Fatalf("Provenance() = %q, want %q", v.Provenance(), tt.wantSource)
			}
		})
	}
}

// --- Zero and empty value counts-as-set tests ---

func TestGetExplicitZeroValueCountsAsSet(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Int("count", 99, "count"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "count", Value: 0})
	got, err := snap.GetInt("count")
	if err != nil {
		t.Fatalf("GetInt returned unexpected error: %v", err)
	}
	if got != 0 {
		t.Fatalf("GetInt() = %d, want 0", got)
	}
	if !snap.IsSet("count") {
		t.Fatal("IsSet(count) = false, want true for explicitly set zero")
	}
}

func TestGetEmptyEnvStringCountsAsSet(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.String("prefix", "default", "prefix"))
	snap, _ := config.NewEnvSnapshot(set,
		mapLookup(map[string]string{"PREFIX": ""}),
		[]config.EnvBinding{config.BindEnv("prefix", "PREFIX")},
	)
	got, err := snap.GetString("prefix")
	if err != nil {
		t.Fatalf("GetString returned unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("GetString() = %q, want empty", got)
	}
	if !snap.IsSet("prefix") {
		t.Fatal("IsSet(prefix) = false, want true for empty env string")
	}
}

func TestGetStringListEmptySliceCountsAsSet(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.StringList("tags", []string{"default"}, "tags"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "tags", Value: []string{}})
	got, err := snap.GetStringList("tags")
	if err != nil {
		t.Fatalf("GetStringList returned unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("GetStringList() = %v, want empty slice", got)
	}
	if !snap.IsSet("tags") {
		t.Fatal("IsSet(tags) = false, want true for explicitly set empty slice")
	}
}

// --- Sensitive key redaction tests ---

func TestGetSensitiveKeyAbsentError(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Define("api-key", config.KindString, "api key", config.Sensitive()))
	snap := set.DefaultSnapshot()
	_, err := snap.GetString("api-key")
	if err == nil {
		t.Fatal("GetString on absent sensitive key returned nil error")
	}
	if !errors.Is(err, config.ErrKeyAbsent) {
		t.Fatalf("errors.Is(err, ErrKeyAbsent) = false; err=%v", err)
	}
	var getErr *config.GetError
	if !errors.As(err, &getErr) {
		t.Fatalf("error does not expose *config.GetError: %T", err)
	}
	if !getErr.Redacted() {
		t.Fatal("GetError.Redacted() = false, want true for sensitive key")
	}
	corpus := []string{"dib_fake_secret_value", "dib_fake_password_value", "dib_fake_token_value"}
	for _, raw := range corpus {
		if strings.Contains(err.Error(), raw) {
			t.Fatalf("sensitive absent error leaked corpus value %q: %v", raw, err)
		}
	}
}

func TestGetSensitiveKeyKindMismatchError(t *testing.T) {
	t.Parallel()
	set, _ := config.NewSet(config.Define("secret", config.KindInt, "secret", config.Sensitive()))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "secret", Value: 42})
	_, err := snap.GetString("secret")
	if err == nil {
		t.Fatal("GetString on KindInt sensitive key returned nil error")
	}
	if !errors.Is(err, config.ErrGetConversion) {
		t.Fatalf("errors.Is(err, ErrGetConversion) = false; err=%v", err)
	}
	var getErr *config.GetError
	if !errors.As(err, &getErr) {
		t.Fatalf("error does not expose *config.GetError: %T", err)
	}
	if !getErr.Redacted() {
		t.Fatal("GetError.Redacted() = false, want true for sensitive key")
	}
	corpus := []string{"dib_fake_secret_value", "dib_fake_password_value", "dib_fake_token_value"}
	for _, raw := range corpus {
		if strings.Contains(err.Error(), raw) {
			t.Fatalf("sensitive mismatch error leaked corpus value %q: %v", raw, err)
		}
	}
}

// --- Sentinel and accessor tests ---

func TestGetErrorSentinelCategories(t *testing.T) {
	t.Parallel()

	set, _ := config.NewSet(
		config.Bool("flag", true, "flag"),
		config.Define("absent-key", config.KindString, "no default"),
	)
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "flag", Value: true})

	tests := []struct {
		name   string
		run    func() error
		wantIs error
	}{
		{
			name: "ErrKeyNotFound",
			run: func() error {
				_, err := snap.GetString("nonexistent")
				return err
			},
			wantIs: config.ErrKeyNotFound,
		},
		{
			name: "ErrKeyAbsent",
			run: func() error {
				defaultSnap := set.DefaultSnapshot()
				_, err := defaultSnap.GetString("absent-key")
				return err
			},
			wantIs: config.ErrKeyAbsent,
		},
		{
			name: "ErrGetConversion",
			run: func() error {
				_, err := snap.GetString("flag")
				return err
			},
			wantIs: config.ErrGetConversion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.run()
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !errors.Is(err, tt.wantIs) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.wantIs, err)
			}
		})
	}
}

func TestGetErrorAccessors(t *testing.T) {
	t.Parallel()

	set, _ := config.NewSet(config.Bool("verbose", false, "verbose"))
	snap, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "verbose", Value: true})

	_, err := snap.GetString("verbose")
	if err == nil {
		t.Fatal("GetString on KindBool returned nil error")
	}

	var getErr *config.GetError
	if !errors.As(err, &getErr) {
		t.Fatalf("error does not expose *config.GetError: %T", err)
	}

	if got := getErr.Key(); got != "verbose" {
		t.Fatalf("Key() = %q, want verbose", got)
	}
	if got := getErr.Kind(); got != config.KindBool {
		t.Fatalf("Kind() = %v, want KindBool", got)
	}
	if got := getErr.WantKind(); got != config.KindString {
		t.Fatalf("WantKind() = %v, want KindString", got)
	}
	if got := getErr.SourceLabel(); got != config.SourceExplicit {
		t.Fatalf("SourceLabel() = %q, want %q", got, config.SourceExplicit)
	}
	if got := getErr.Redacted(); got != false {
		t.Fatalf("Redacted() = %v, want false", got)
	}
	if got := getErr.Category(); got != config.ErrGetConversion {
		t.Fatalf("Category() = %v, want ErrGetConversion", got)
	}
}

func TestGetStringListNilVsEmpty(t *testing.T) {
	t.Parallel()

	// nil-default key (no default registered) — absent, returns ErrKeyAbsent
	set, _ := config.NewSet(
		config.Define("nil-key", config.KindStringList, "no default"),
		config.StringList("empty-key", []string{}, "empty default"),
	)
	snap := set.DefaultSnapshot()

	// nil-default key (no default) returns ErrKeyAbsent — no panic
	_, err := snap.GetStringList("nil-key")
	if err == nil {
		t.Fatal("GetStringList on no-default key returned nil error")
	}
	if !errors.Is(err, config.ErrKeyAbsent) {
		t.Fatalf("errors.Is(err, ErrKeyAbsent) = false; err=%v", err)
	}

	// empty default key returns value with len 0 — no panic
	got, err := snap.GetStringList("empty-key")
	if err != nil {
		t.Fatalf("GetStringList on empty-default key returned unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("GetStringList() = %v, want empty", got)
	}

	// explicit empty slice — no panic
	snap2, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "nil-key", Value: []string{}})
	got2, err := snap2.GetStringList("nil-key")
	if err != nil {
		t.Fatalf("GetStringList on explicit empty slice returned unexpected error: %v", err)
	}
	if len(got2) != 0 {
		t.Fatalf("GetStringList() = %v, want empty slice", got2)
	}
}
