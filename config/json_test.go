package config_test

import (
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/petabytecl/dib/config"
)

func TestJSONReaderSnapshotStrictAndPermissiveModes(t *testing.T) {
	t.Parallel()

	set := mustSourceTestSet(t)
	snapshot, err := config.LoadJSON(
		set,
		strings.NewReader(`{"log-level":"debug","debug":false,"workers":0,"limit64":64,"retries":3,"bytes":128,"timeout":"150ms","tags":["alpha","beta"],"ratio":"0.5"}`),
		config.JSONReaderLabel("inline"),
	)
	if err != nil {
		t.Fatalf("LoadJSON returned unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want any
	}{
		{key: "log-level", want: "debug"},
		{key: "debug", want: false},
		{key: "workers", want: 0},
		{key: "limit64", want: int64(64)},
		{key: "retries", want: uint(3)},
		{key: "bytes", want: uint64(128)},
		{key: "timeout", want: 150 * time.Millisecond},
		{key: "tags", want: []string{"alpha", "beta"}},
		{key: "ratio", want: 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			value, ok := snapshot.Lookup(tt.key)
			if !ok {
				t.Fatalf("Lookup(%q) returned false", tt.key)
			}
			got, hasValue := value.Value()
			if !hasValue || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Value() = %#v, %v; want %#v, true", got, hasValue, tt.want)
			}
			if got := value.Provenance(); got != config.SourceJSON {
				t.Fatalf("Provenance() = %q, want JSON", got)
			}
			if got := value.Source().JSONReaderLabel(); got != "inline" {
				t.Fatalf("Source().JSONReaderLabel() = %q, want inline", got)
			}
		})
	}

	_, err = config.LoadJSON(set, strings.NewReader(`{"unknown":true}`))
	if err == nil {
		t.Fatal("strict LoadJSON returned nil error for unknown key")
	}
	if !errors.Is(err, config.ErrUnknownSourceKey) {
		t.Fatalf("errors.Is(err, ErrUnknownSourceKey) = false; err=%v", err)
	}

	snapshot, err = config.LoadJSON(set, strings.NewReader(`{"unknown":true,"workers":2}`), config.JSONPermissive())
	if err != nil {
		t.Fatalf("permissive LoadJSON returned unexpected error: %v", err)
	}
	workers, ok := snapshot.Lookup("workers")
	if !ok {
		t.Fatal("Lookup(workers) returned false")
	}
	gotWorkers, hasWorkers := workers.Value()
	if !hasWorkers || gotWorkers != 2 {
		t.Fatalf("workers Value() = %#v, %v; want 2, true", gotWorkers, hasWorkers)
	}
}

func TestJSONPathSnapshotUsesCallerSuppliedPathAndFixtures(t *testing.T) {
	t.Parallel()

	set := mustSourceTestSet(t)
	snapshot, err := config.LoadJSONFile(set, "testdata/json/valid.json")
	if err != nil {
		t.Fatalf("LoadJSONFile returned unexpected error: %v", err)
	}
	logLevel, ok := snapshot.Lookup("log-level")
	if !ok {
		t.Fatal("Lookup(log-level) returned false")
	}
	got, hasValue := logLevel.Value()
	if !hasValue || got != "file-debug" {
		t.Fatalf("log-level Value() = %#v, %v; want file-debug, true", got, hasValue)
	}
	if got := logLevel.Source().JSONPath(); got != "testdata/json/valid.json" {
		t.Fatalf("Source().JSONPath() = %q, want fixture path", got)
	}
}

func TestJSONDiagnosticsAreTypedAndRedacted(t *testing.T) {
	t.Parallel()

	set := mustSourceTestSet(t)
	tests := []struct {
		name       string
		load       func() error
		want       error
		wantKey    string
		wantPath   string
		wantSecret string
		wantIs     error
	}{
		{
			name: "read failure",
			load: func() error {
				_, err := config.LoadJSON(set, failingReader{err: errControlledRead}, config.JSONReaderLabel("failing"))
				return err
			},
			want:   config.ErrSourceRead,
			wantIs: errControlledRead,
		},
		{
			name: "malformed JSON fixture",
			load: func() error {
				_, err := config.LoadJSONFile(set, "testdata/json/malformed.json")
				return err
			},
			want:     config.ErrJSONDecode,
			wantPath: "testdata/json/malformed.json",
		},
		{
			name: "non object JSON fixture",
			load: func() error {
				_, err := config.LoadJSONFile(set, "testdata/json/non-object.json")
				return err
			},
			want:     config.ErrJSONDecode,
			wantPath: "testdata/json/non-object.json",
		},
		{
			name: "trailing JSON data",
			load: func() error {
				_, err := config.LoadJSON(set, strings.NewReader(`{"workers":1} {"debug":true}`))
				return err
			},
			want: config.ErrJSONDecode,
		},
		{
			name: "unknown key fixture",
			load: func() error {
				_, err := config.LoadJSONFile(set, "testdata/json/unknown-key.json")
				return err
			},
			want:     config.ErrUnknownSourceKey,
			wantKey:  "unregistered",
			wantPath: "testdata/json/unknown-key.json",
		},
		{
			name: "bad type fixture",
			load: func() error {
				_, err := config.LoadJSONFile(set, "testdata/json/bad-type.json")
				return err
			},
			want:     config.ErrSourceConversion,
			wantKey:  "workers",
			wantPath: "testdata/json/bad-type.json",
		},
		{
			name: "fractional int rejected",
			load: func() error {
				_, err := config.LoadJSON(set, strings.NewReader(`{"workers":1.5}`))
				return err
			},
			want:    config.ErrSourceConversion,
			wantKey: "workers",
		},
		{
			name: "negative uint rejected",
			load: func() error {
				_, err := config.LoadJSON(set, strings.NewReader(`{"retries":-1}`))
				return err
			},
			want:    config.ErrSourceConversion,
			wantKey: "retries",
		},
		{
			name: "sensitive type failure redacts raw JSON value",
			load: func() error {
				_, err := config.LoadJSONFile(set, "testdata/json/sensitive-value.json")
				return err
			},
			want:       config.ErrSourceConversion,
			wantKey:    "token",
			wantPath:   "testdata/json/sensitive-value.json",
			wantSecret: "dib_fake_token_value",
		},
		{
			name: "missing path preserves not-exist inspection",
			load: func() error {
				_, err := config.LoadJSONFile(set, "testdata/json/missing.json")
				return err
			},
			want:     config.ErrSourceRead,
			wantPath: "testdata/json/missing.json",
			wantIs:   os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.load()
			if err == nil {
				t.Fatal("load returned nil error")
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.want, err)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Fatalf("errors.Is(err, %v) = false; err=%v", tt.wantIs, err)
			}
			var sourceErr *config.SourceError
			if !errors.As(err, &sourceErr) {
				t.Fatalf("error does not expose *config.SourceError: %T", err)
			}
			if got := sourceErr.Source(); got != config.SourceJSON {
				t.Fatalf("SourceError.Source() = %q, want JSON", got)
			}
			if got := sourceErr.Key(); got != tt.wantKey {
				t.Fatalf("SourceError.Key() = %q, want %q", got, tt.wantKey)
			}
			if got := sourceErr.JSONPath(); got != tt.wantPath {
				t.Fatalf("SourceError.JSONPath() = %q, want %q", got, tt.wantPath)
			}
			if tt.wantSecret != "" && strings.Contains(err.Error(), tt.wantSecret) {
				t.Fatalf("sensitive JSON error leaked raw value: %v", err)
			}
		})
	}
}

func TestJSONDiagnosticsUseDeterministicKeyOrder(t *testing.T) {
	t.Parallel()

	set := mustSourceTestSet(t)
	_, err := config.LoadJSON(set, strings.NewReader(`{"z-unknown":true,"a-unknown":true}`))
	if err == nil {
		t.Fatal("LoadJSON returned nil error")
	}
	if !errors.Is(err, config.ErrUnknownSourceKey) {
		t.Fatalf("errors.Is(err, ErrUnknownSourceKey) = false; err=%v", err)
	}
	var sourceErr *config.SourceError
	if !errors.As(err, &sourceErr) {
		t.Fatalf("error does not expose *config.SourceError: %T", err)
	}
	if got := sourceErr.Key(); got != "a-unknown" {
		t.Fatalf("SourceError.Key() = %q, want deterministic first key a-unknown", got)
	}
}

func TestJSONSnapshotDefensiveStringLists(t *testing.T) {
	t.Parallel()

	set := mustSourceTestSet(t)
	snapshot, err := config.LoadJSON(set, strings.NewReader(`{"tags":["alpha"]}`))
	if err != nil {
		t.Fatalf("LoadJSON returned unexpected error: %v", err)
	}
	tags, ok := snapshot.Lookup("tags")
	if !ok {
		t.Fatal("Lookup(tags) returned false")
	}
	got, hasValue := tags.Value()
	if !hasValue {
		t.Fatal("tags Value() returned hasValue=false")
	}
	got.([]string)[0] = "changed"
	again, _ := tags.Value()
	if !reflect.DeepEqual(again, []string{"alpha"}) {
		t.Fatalf("Value() leaked mutable slice alias: %#v", again)
	}
}

func mustSourceTestSet(t *testing.T) config.Set {
	t.Helper()

	set, err := config.NewSet(
		config.String("log-level", "info", "log level"),
		config.Bool("debug", false, "debug"),
		config.Int("workers", 1, "workers"),
		config.Int64("limit64", 1, "limit64"),
		config.Uint("retries", 0, "retries"),
		config.Uint64("bytes", 0, "bytes"),
		config.Duration("timeout", time.Second, "timeout"),
		config.StringList("tags", nil, "tags"),
		config.Float64("ratio", 0, "ratio"),
		config.Define("token", config.KindInt, "token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	return set
}

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

var errControlledRead = errors.New("controlled read failure")

var _ io.Reader = failingReader{}
