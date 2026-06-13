package config_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/config"
)

func TestSourceReportCoversWinningSourcesAndAbsentKeys(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("from-default", "default-value", "default"),
		config.Define("from-explicit", config.KindString, "explicit"),
		config.Define("from-flag", config.KindInt, "flag"),
		config.Define("from-env", config.KindBool, "env"),
		config.Define("from-json", config.KindString, "json"),
		config.Define("absent", config.KindString, "absent"),
		config.Define("secret", config.KindString, "secret", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	explicit, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "from-explicit", Value: "explicit-value"})
	flag, _ := config.NewFlagSnapshot(set, []config.FlagValue{{ConfigKey: "from-flag", ExplicitlySet: true, Value: 7}})
	env, _ := config.NewEnvSnapshot(set, mapLookup(map[string]string{"FROM_ENV": "true"}), []config.EnvBinding{config.BindEnv("from-env", "FROM_ENV")})
	jsonSnap, _ := config.LoadJSON(set, strings.NewReader(`{"from-json":"json-value","secret":"dib_fake_secret_value"}`), config.JSONReaderLabel("inline-config"))

	report := config.Resolve(set, explicit, flag, env, jsonSnap).SourceReport()
	got := reportRows(report)
	want := []sourceReportRow{
		{key: "from-default", kind: config.KindString, set: true, source: config.SourceDefault},
		{key: "from-explicit", kind: config.KindString, set: true, source: config.SourceExplicit},
		{key: "from-flag", kind: config.KindInt, set: true, source: config.SourceFlagBinding},
		{key: "from-env", kind: config.KindBool, set: true, source: config.SourceEnv, envName: "FROM_ENV"},
		{key: "from-json", kind: config.KindString, set: true, source: config.SourceJSON, jsonReader: "inline-config"},
		{key: "absent", kind: config.KindString, source: ""},
		{key: "secret", kind: config.KindString, set: true, source: config.SourceJSON, jsonReader: "inline-config", redacted: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SourceReport rows = %#v, want %#v", got, want)
	}
	assertReportLabelsClosed(t, report)
}

func TestSourceReportRenderingIsDeterministicValueFreeAndWriterBound(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("visible", "ordinary-value", "visible"),
		config.Define("token", config.KindString, "token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	explicit, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "visible", Value: "render-me-not"})
	env, _ := config.NewEnvSnapshot(set, mapLookup(map[string]string{"TOKEN": "dib_fake_password_value"}), []config.EnvBinding{config.BindEnv("token", "TOKEN")})
	snapshot := config.Resolve(set, explicit, config.Snapshot{}, env, config.Snapshot{})

	var first bytes.Buffer
	if err := snapshot.WriteSourceReport(&first); err != nil {
		t.Fatalf("WriteSourceReport: %v", err)
	}
	var second bytes.Buffer
	if err := snapshot.WriteSourceReport(&second); err != nil {
		t.Fatalf("WriteSourceReport second: %v", err)
	}
	if first.String() != second.String() {
		t.Fatalf("WriteSourceReport not deterministic:\nfirst=%s\nsecond=%s", first.String(), second.String())
	}
	for _, forbidden := range []string{"ordinary-value", "render-me-not", "dib_fake_secret_value", "dib_fake_password_value", "dib_fake_token_value"} {
		if strings.Contains(first.String(), forbidden) {
			t.Fatalf("rendered source report leaked %q:\n%s", forbidden, first.String())
		}
	}
	if !strings.Contains(first.String(), `key="visible"`) || !strings.Contains(first.String(), `source="explicit setter"`) {
		t.Fatalf("rendered report omitted useful source context:\n%s", first.String())
	}
	if !strings.Contains(first.String(), `key="token"`) || !strings.Contains(first.String(), `env="TOKEN"`) || !strings.Contains(first.String(), `redacted=true`) {
		t.Fatalf("rendered report omitted sensitive source metadata:\n%s", first.String())
	}

	err = snapshot.WriteSourceReport(failingWriter{err: errControlledWrite})
	if !errors.Is(err, errControlledWrite) {
		t.Fatalf("WriteSourceReport writer error = %v, want %v", err, errControlledWrite)
	}
}

func TestSourceReportIncludesJSONPathMetadata(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.String("log-level", "info", "log level"),
		config.Bool("debug", false, "debug"),
		config.Int("workers", 1, "workers"),
		config.Uint("retries", 0, "retries"),
		config.Duration("timeout", 0, "timeout"),
		config.StringList("tags", nil, "tags"),
		config.Float64("ratio", 0, "ratio"),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	snapshot, err := config.LoadJSONFile(set, "testdata/json/valid.json")
	if err != nil {
		t.Fatalf("LoadJSONFile: %v", err)
	}

	report := snapshot.SourceReport()
	if got, want := len(report), 7; got != want {
		t.Fatalf("SourceReport length = %d, want %d", got, want)
	}
	for _, entry := range report {
		if !entry.IsSet() {
			t.Fatalf("entry %q IsSet() = false, want true", entry.Key())
		}
		if got := entry.SourceLabel(); got != config.SourceJSON {
			t.Fatalf("entry %q SourceLabel() = %q, want JSON", entry.Key(), got)
		}
		if got := entry.JSONPath(); got != "testdata/json/valid.json" {
			t.Fatalf("entry %q JSONPath() = %q, want fixture path", entry.Key(), got)
		}
		if got := entry.JSONReaderLabel(); got != "" {
			t.Fatalf("entry %q JSONReaderLabel() = %q, want empty for file source", entry.Key(), got)
		}
	}
}

func TestSourceReportPublicFormattingNeverLeaksRawValues(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(config.Define("token", config.KindString, "token", config.Sensitive()))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	explicit, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "token", Value: "dib_fake_token_value"})
	report := explicit.SourceReport()
	if len(report) != 1 {
		t.Fatalf("SourceReport length = %d, want 1", len(report))
	}
	for _, rendered := range []string{fmt.Sprint(report[0]), fmt.Sprintf("%#v", report[0])} {
		assertNoCorpus(t, rendered)
	}
}

func TestReportAndDiagnosticNilWritersReturnErrors(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(config.String("visible", "value", "visible"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	snapshot := set.DefaultSnapshot()
	if err := snapshot.WriteSourceReport(nil); err == nil {
		t.Fatal("WriteSourceReport(nil) returned nil error")
	}
	if err := config.WriteDiagnostic(nil, errors.New("not config")); err == nil {
		t.Fatal("WriteDiagnostic(nil, err) returned nil error")
	}
}

func TestUnsupportedDiagnosticIsRenderedWithoutClassification(t *testing.T) {
	t.Parallel()

	var rendered bytes.Buffer
	if err := config.WriteDiagnostic(&rendered, errors.New("external failure")); err != nil {
		t.Fatalf("WriteDiagnostic unsupported error: %v", err)
	}
	if got, want := rendered.String(), "config diagnostic: unsupported\n"; got != want {
		t.Fatalf("unsupported diagnostic rendering = %q, want %q", got, want)
	}
	if _, ok := config.InspectDiagnostic(errors.New("external failure")); ok {
		t.Fatal("InspectDiagnostic returned ok=true for non-config error")
	}
	if _, ok := config.InspectDiagnostic(nil); ok {
		t.Fatal("InspectDiagnostic(nil) returned ok=true")
	}
}

func TestInspectDiagnosticClassifiesConfigErrors(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(
		config.Int("workers", 1, "workers"),
		config.Define("absent", config.KindString, "absent"),
		config.Bool("enabled", false, "enabled"),
		config.Define("token", config.KindInt, "token", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}

	_, sourceReadErr := config.LoadJSONFile(set, "testdata/json/does-not-exist.json")
	_, decodeErr := config.LoadJSON(set, strings.NewReader(`{"workers":1} {"enabled":true}`))
	_, sourceConversionErr := config.NewEnvSnapshot(set, mapLookup(map[string]string{"WORKERS": "many"}), []config.EnvBinding{config.BindEnv("workers", "WORKERS")})
	_, duplicateBindingErr := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "workers", ExplicitlySet: true, Value: 1},
		{ConfigKey: "workers", ExplicitlySet: true, Value: 2},
	})
	_, absentErr := set.DefaultSnapshot().GetString("absent")
	_, notFoundErr := set.DefaultSnapshot().GetString("missing")
	flagSnap, _ := config.NewFlagSnapshot(set, []config.FlagValue{{ConfigKey: "enabled", ExplicitlySet: true, Value: true}})
	_, getConversionErr := flagSnap.GetString("enabled")
	_, defaultErr := config.NewSet(config.Define("bad-default", config.KindInt, "bad", config.Default("not-an-int")))

	tests := []struct {
		name       string
		err        error
		category   error
		key        string
		kind       config.Kind
		wantKind   config.Kind
		source     string
		jsonPath   string
		envName    string
		hasCause   bool
		wantErrors []error
	}{
		{name: "source read", err: sourceReadErr, category: config.ErrSourceRead, kind: config.Kind(-1), source: config.SourceJSON, jsonPath: "testdata/json/does-not-exist.json", hasCause: true, wantErrors: []error{config.ErrSourceRead, os.ErrNotExist}},
		{name: "json decode", err: decodeErr, category: config.ErrJSONDecode, kind: config.Kind(-1), source: config.SourceJSON, hasCause: true, wantErrors: []error{config.ErrJSONDecode}},
		{name: "source conversion", err: sourceConversionErr, category: config.ErrSourceConversion, key: "workers", kind: config.KindInt, source: config.SourceEnv, envName: "WORKERS", hasCause: true, wantErrors: []error{config.ErrSourceConversion}},
		{name: "duplicate flag binding", err: duplicateBindingErr, category: config.ErrDuplicateBinding, key: "workers", kind: config.KindInt, source: config.SourceFlagBinding, wantErrors: []error{config.ErrDuplicateBinding}},
		{name: "getter absent", err: absentErr, category: config.ErrKeyAbsent, key: "absent", kind: config.KindString, wantKind: config.KindString, wantErrors: []error{config.ErrKeyAbsent}},
		{name: "getter not found", err: notFoundErr, category: config.ErrKeyNotFound, key: "missing", kind: config.KindString, wantKind: config.KindString, wantErrors: []error{config.ErrKeyNotFound}},
		{name: "getter conversion", err: getConversionErr, category: config.ErrGetConversion, key: "enabled", kind: config.KindBool, wantKind: config.KindString, source: config.SourceFlagBinding, wantErrors: []error{config.ErrGetConversion}},
		{name: "invalid default", err: defaultErr, category: config.ErrInvalidDefault, key: "bad-default", kind: config.KindInt, source: config.SourceDefault, wantErrors: []error{config.ErrInvalidDefault}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag, ok := config.InspectDiagnostic(tt.err)
			if !ok {
				t.Fatalf("InspectDiagnostic returned ok=false for %T", tt.err)
			}
			if !errors.Is(diag.Category(), tt.category) {
				t.Fatalf("Category() = %v, want %v", diag.Category(), tt.category)
			}
			if diag.Key() != tt.key || diag.Kind() != tt.kind || diag.WantKind() != tt.wantKind || diag.SourceLabel() != tt.source {
				t.Fatalf("diagnostic = key %q kind %v want %v source %q", diag.Key(), diag.Kind(), diag.WantKind(), diag.SourceLabel())
			}
			if diag.EnvName() != tt.envName || diag.JSONPath() != tt.jsonPath {
				t.Fatalf("diagnostic metadata = env %q path %q", diag.EnvName(), diag.JSONPath())
			}
			if diag.HasSafeCause() != tt.hasCause {
				t.Fatalf("HasSafeCause() = %v, want %v", diag.HasSafeCause(), tt.hasCause)
			}
			for _, want := range tt.wantErrors {
				if !errors.Is(tt.err, want) {
					t.Fatalf("original error no longer satisfies errors.Is(%v): %v", want, tt.err)
				}
			}
		})
	}
}

func TestWriteDiagnosticIsDeterministicValueFreeAndWriterBound(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(config.Define("token", config.KindInt, "token", config.Sensitive()))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	_, sourceErr := config.NewEnvSnapshot(set, mapLookup(map[string]string{"TOKEN": "dib_fake_password_value"}), []config.EnvBinding{config.BindEnv("token", "TOKEN")})
	var sourceTyped *config.SourceError
	if !errors.As(sourceErr, &sourceTyped) {
		t.Fatalf("source error does not expose *SourceError: %T", sourceErr)
	}
	if !sourceTyped.Redacted() {
		t.Fatal("SourceError.Redacted() = false, want true")
	}

	diag, ok := config.InspectDiagnostic(sourceErr)
	if !ok {
		t.Fatal("InspectDiagnostic returned ok=false")
	}
	if !diag.Redacted() {
		t.Fatal("Diagnostic.Redacted() = false, want true")
	}

	var first bytes.Buffer
	if err := config.WriteDiagnostic(&first, sourceErr); err != nil {
		t.Fatalf("WriteDiagnostic: %v", err)
	}
	var second bytes.Buffer
	if err := config.WriteDiagnostic(&second, sourceErr); err != nil {
		t.Fatalf("WriteDiagnostic second: %v", err)
	}
	if first.String() != second.String() {
		t.Fatalf("WriteDiagnostic not deterministic:\nfirst=%s\nsecond=%s", first.String(), second.String())
	}
	for _, rendered := range []string{sourceErr.Error(), fmt.Sprint(diag), fmt.Sprintf("%#v", diag), first.String()} {
		assertNoCorpus(t, rendered)
	}
	if !strings.Contains(first.String(), `category="config source value conversion failure"`) ||
		!strings.Contains(first.String(), `source="env"`) ||
		!strings.Contains(first.String(), `env="TOKEN"`) ||
		!strings.Contains(first.String(), `redacted=true`) {
		t.Fatalf("rendered diagnostic omitted structured context:\n%s", first.String())
	}

	err = config.WriteDiagnostic(failingWriter{err: errControlledWrite}, sourceErr)
	if !errors.Is(err, errControlledWrite) {
		t.Fatalf("WriteDiagnostic writer error = %v, want %v", err, errControlledWrite)
	}
}

func TestDiagnosticFalsePositiveRedactionCoverage(t *testing.T) {
	t.Parallel()

	set, err := config.NewSet(config.Int("workers", 1, "workers"))
	if err != nil {
		t.Fatalf("NewSet: %v", err)
	}
	_, sourceErr := config.NewEnvSnapshot(set, mapLookup(map[string]string{"WORKERS": "ordinary"}), []config.EnvBinding{config.BindEnv("workers", "WORKERS")})
	diag, ok := config.InspectDiagnostic(sourceErr)
	if !ok {
		t.Fatal("InspectDiagnostic returned ok=false")
	}
	if diag.Redacted() {
		t.Fatal("Diagnostic.Redacted() = true, want false for non-sensitive key")
	}
	var rendered bytes.Buffer
	if err := config.WriteDiagnostic(&rendered, sourceErr); err != nil {
		t.Fatalf("WriteDiagnostic: %v", err)
	}
	if strings.Contains(rendered.String(), "ordinary") {
		t.Fatalf("rendered diagnostic leaked raw non-sensitive value: %s", rendered.String())
	}
}

func ExampleSnapshot_SourceReport() {
	set, _ := config.NewSet(
		config.String("log-level", "info", "log level"),
		config.Define("token", config.KindString, "token", config.Sensitive()),
	)
	explicit, _ := config.NewExplicitSnapshot(set, config.Assignment{Key: "token", Value: "dib_fake_token_value"})
	snapshot := config.Resolve(set, explicit, config.Snapshot{}, config.Snapshot{}, config.Snapshot{})

	for _, entry := range snapshot.SourceReport() {
		fmt.Printf("%s %s set=%v redacted=%v\n", entry.Key(), entry.SourceLabel(), entry.IsSet(), entry.Redacted())
	}

	// Output:
	// log-level default set=true redacted=false
	// token explicit setter set=true redacted=true
}

type sourceReportRow struct {
	key        string
	kind       config.Kind
	set        bool
	source     string
	envName    string
	jsonPath   string
	jsonReader string
	redacted   bool
}

func reportRows(entries []config.SourceReportEntry) []sourceReportRow {
	rows := make([]sourceReportRow, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, sourceReportRow{
			key:        entry.Key(),
			kind:       entry.Kind(),
			set:        entry.IsSet(),
			source:     entry.SourceLabel(),
			envName:    entry.EnvName(),
			jsonPath:   entry.JSONPath(),
			jsonReader: entry.JSONReaderLabel(),
			redacted:   entry.Redacted(),
		})
	}
	return rows
}

func assertReportLabelsClosed(t *testing.T, entries []config.SourceReportEntry) {
	t.Helper()
	allowed := map[string]bool{
		config.SourceDefault:     true,
		config.SourceExplicit:    true,
		config.SourceFlagBinding: true,
		config.SourceEnv:         true,
		config.SourceJSON:        true,
	}
	for _, entry := range entries {
		if entry.SourceLabel() == "" {
			if entry.IsSet() {
				t.Fatalf("set entry %q has empty source label", entry.Key())
			}
			continue
		}
		if !allowed[entry.SourceLabel()] {
			t.Fatalf("entry %q has non-canonical source label %q", entry.Key(), entry.SourceLabel())
		}
	}
}

func assertNoCorpus(t *testing.T, rendered string) {
	t.Helper()
	for _, forbidden := range []string{"dib_fake_secret_value", "dib_fake_password_value", "dib_fake_token_value"} {
		if strings.Contains(rendered, forbidden) {
			t.Fatalf("rendered text leaked %q: %s", forbidden, rendered)
		}
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

var errControlledWrite = errors.New("controlled write failure")

var _ io.Writer = failingWriter{}
