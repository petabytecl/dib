package migration_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

func Example_configPrecedenceMigration() {
	set := mustMigrationConfigSet()

	parsedFlags := mustMigrationFlagSnapshot([]string{"--mode=from-flag", "--workers", "4"})
	flagBindings := mustConfigFlagSnapshot(set, parsedFlags)
	env := mustEnvSnapshot(set, map[string]string{
		"DIB_MODE":     "from-env",
		"DIB_WORKERS":  "3",
		"DIB_ENDPOINT": "from-env",
	})
	jsonSource := mustJSONSnapshot(set, `{
		"mode": "from-json",
		"workers": 2,
		"endpoint": "from-json",
		"format": "json",
		"token": "dib_fake_token_value",
		"password": "dib_fake_password_value"
	}`)
	explicit := mustExplicitSnapshot(set, config.Assignment{Key: "mode", Value: "from-explicit"})

	resolved := config.Resolve(set, explicit, flagBindings, env, jsonSource)
	mode, _ := resolved.GetString("mode")
	workers, _ := resolved.GetInt("workers")
	endpoint, _ := resolved.GetString("endpoint")
	format, _ := resolved.GetString("format")
	retries, _ := resolved.GetInt("retries")

	var report bytes.Buffer
	if err := resolved.WriteSourceReport(&report); err != nil {
		panic(err)
	}

	fmt.Println(mode, workers, endpoint, format, retries)
	fmt.Println(sourceByKey(resolved.SourceReport(), "mode"), sourceByKey(resolved.SourceReport(), "workers"))
	fmt.Println(strings.Contains(report.String(), "dib_fake_token_value"))
	// Output:
	// from-explicit 4 from-env json 1
	// explicit setter flag binding
	// false
}

func TestConfigPrecedenceMigrationResolvesCanonicalOrder(t *testing.T) {
	set := mustMigrationConfigSet()
	parsedFlags := mustMigrationFlagSnapshot([]string{"--mode=from-flag", "--workers", "4"})
	flagBindings := mustConfigFlagSnapshot(set, parsedFlags)
	env := mustEnvSnapshot(set, map[string]string{
		"DIB_MODE":     "from-env",
		"DIB_WORKERS":  "3",
		"DIB_ENDPOINT": "from-env",
		"DIB_COLOR":    "from-env-color",
	})
	jsonSource := mustJSONSnapshot(set, `{
		"mode": "from-json",
		"workers": 2,
		"endpoint": "from-json",
		"format": "json",
		"color": "from-json-color",
		"token": "dib_fake_token_value",
		"password": "dib_fake_password_value"
	}`)
	explicit := mustExplicitSnapshot(set, config.Assignment{Key: "mode", Value: "from-explicit"})

	resolved := config.Resolve(set, explicit, flagBindings, env, jsonSource)
	want := map[string]struct {
		value  any
		source string
	}{
		"mode":     {value: "from-explicit", source: config.SourceExplicit},
		"workers":  {value: 4, source: config.SourceFlagBinding},
		"endpoint": {value: "from-env", source: config.SourceEnv},
		"format":   {value: "json", source: config.SourceJSON},
		"retries":  {value: 1, source: config.SourceDefault},
		"color":    {value: "from-env-color", source: config.SourceEnv},
	}
	for key, expectation := range want {
		value, ok := resolved.Lookup(key)
		if !ok {
			t.Fatalf("Lookup(%q) returned ok=false", key)
		}
		got, hasValue := value.Value()
		if !hasValue || !reflect.DeepEqual(got, expectation.value) {
			t.Fatalf("%s Value() = %#v, %t; want %#v, true", key, got, hasValue, expectation.value)
		}
		if got := value.Provenance(); got != expectation.source {
			t.Fatalf("%s Provenance() = %q, want %q", key, got, expectation.source)
		}
		if !resolved.IsSet(key) {
			t.Fatalf("IsSet(%q) = false, want true", key)
		}
	}

	if mode, err := resolved.GetString("mode"); err != nil || mode != "from-explicit" {
		t.Fatalf("GetString(mode) = %q, %v", mode, err)
	}
	if workers, err := resolved.GetInt("workers"); err != nil || workers != 4 {
		t.Fatalf("GetInt(workers) = %d, %v", workers, err)
	}
}

func TestConfigPrecedenceMigrationReportsAreValueFreeAndRedacted(t *testing.T) {
	set := mustMigrationConfigSet()
	env := mustEnvSnapshot(set, map[string]string{
		"DIB_SECRET": "dib_fake_secret_value",
	})
	jsonSource := mustJSONSnapshot(set, `{
		"token": "dib_fake_token_value",
		"password": "dib_fake_password_value"
	}`)
	resolved := config.Resolve(set, config.Snapshot{}, config.Snapshot{}, env, jsonSource)

	report := resolved.SourceReport()
	if !sourceEntryByKey(report, "secret").Redacted() ||
		!sourceEntryByKey(report, "token").Redacted() ||
		!sourceEntryByKey(report, "password").Redacted() {
		t.Fatalf("sensitive report entries are not redacted: %#v", report)
	}

	var rendered bytes.Buffer
	if err := resolved.WriteSourceReport(&rendered); err != nil {
		t.Fatalf("WriteSourceReport: %v", err)
	}
	assertNoSensitiveCorpus(t, rendered.String())

	_, err := config.NewEnvSnapshot(set, mapEnv(map[string]string{"DIB_BAD_SECRET": "dib_fake_secret_value"}), []config.EnvBinding{
		config.BindEnv("bad-secret", "DIB_BAD_SECRET"),
	})
	if err == nil {
		t.Fatal("NewEnvSnapshot returned nil error for sensitive conversion failure")
	}
	if !errors.Is(err, config.ErrSourceConversion) {
		t.Fatalf("error %v does not satisfy ErrSourceConversion", err)
	}
	var sourceErr *config.SourceError
	if !errors.As(err, &sourceErr) {
		t.Fatalf("error does not expose *config.SourceError: %T", err)
	}
	if got := sourceErr.Source(); got != config.SourceEnv {
		t.Fatalf("SourceError.Source() = %q, want %q", got, config.SourceEnv)
	}
	if got := sourceErr.Key(); got != "bad-secret" {
		t.Fatalf("SourceError.Key() = %q, want bad-secret", got)
	}
	if got := sourceErr.EnvName(); got != "DIB_BAD_SECRET" {
		t.Fatalf("SourceError.EnvName() = %q, want DIB_BAD_SECRET", got)
	}
	if got := sourceErr.Kind(); got != config.KindInt {
		t.Fatalf("SourceError.Kind() = %v, want %v", got, config.KindInt)
	}
	if !sourceErr.Redacted() {
		t.Fatal("SourceError.Redacted() = false, want true")
	}
	var diagnostic bytes.Buffer
	if writeErr := config.WriteDiagnostic(&diagnostic, err); writeErr != nil {
		t.Fatalf("WriteDiagnostic: %v", writeErr)
	}
	assertNoSensitiveCorpus(t, err.Error())
	assertNoSensitiveCorpus(t, diagnostic.String())
}

func mustMigrationConfigSet() config.Set {
	set, err := config.NewSet(
		config.String("mode", "from-default", "mode"),
		config.Int("workers", 1, "workers"),
		config.String("endpoint", "from-default", "endpoint"),
		config.String("format", "text", "format"),
		config.Int("retries", 1, "retries"),
		config.String("color", "auto", "color"),
		config.Define("secret", config.KindString, "secret", config.Sensitive()),
		config.Define("token", config.KindString, "token", config.Sensitive()),
		config.Define("password", config.KindString, "password", config.Sensitive()),
		config.Define("bad-secret", config.KindInt, "bad secret", config.Sensitive()),
	)
	if err != nil {
		panic(err)
	}
	return set
}

func mustMigrationFlagSnapshot(args []string) flags.Snapshot {
	set, err := flags.NewSet(
		flags.String("mode", "from-default", "mode"),
		flags.Int("workers", 1, "workers"),
		flags.String("color", "auto", "color"),
	)
	if err != nil {
		panic(err)
	}
	snapshot, err := set.Parse(args)
	if err != nil {
		panic(err)
	}
	return snapshot
}

func mustConfigFlagSnapshot(set config.Set, snapshot flags.Snapshot) config.Snapshot {
	mode, _ := snapshot.Lookup("mode")
	workers, _ := snapshot.Lookup("workers")
	color, _ := snapshot.Lookup("color")
	flagSource, err := config.NewFlagSnapshot(set, []config.FlagValue{
		{ConfigKey: "mode", ExplicitlySet: mode.Explicit(), Value: firstFlagValue(mode)},
		{ConfigKey: "workers", ExplicitlySet: workers.Explicit(), Value: firstFlagValue(workers)},
		{ConfigKey: "color", ExplicitlySet: color.Explicit(), Value: firstFlagValue(color)},
	})
	if err != nil {
		panic(err)
	}
	return flagSource
}

func firstFlagValue(state flags.ValueState) any {
	values := state.Values()
	if len(values) == 0 {
		return nil
	}
	return values[0]
}

func mustExplicitSnapshot(set config.Set, assignments ...config.Assignment) config.Snapshot {
	snapshot, err := config.NewExplicitSnapshot(set, assignments...)
	if err != nil {
		panic(err)
	}
	return snapshot
}

func mustEnvSnapshot(set config.Set, values map[string]string) config.Snapshot {
	snapshot, err := config.NewEnvSnapshot(set, mapEnv(values), []config.EnvBinding{
		config.BindEnv("mode", "DIB_MODE"),
		config.BindEnv("workers", "DIB_WORKERS"),
		config.BindEnv("endpoint", "DIB_ENDPOINT"),
		config.BindEnv("color", "DIB_COLOR"),
		config.BindEnv("secret", "DIB_SECRET"),
	})
	if err != nil {
		panic(err)
	}
	return snapshot
}

func mustJSONSnapshot(set config.Set, body string) config.Snapshot {
	snapshot, err := config.LoadJSON(set, strings.NewReader(body), config.JSONReaderLabel("inline migration JSON"))
	if err != nil {
		panic(err)
	}
	return snapshot
}

func mapEnv(values map[string]string) config.EnvLookup {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}

func sourceByKey(report []config.SourceReportEntry, key string) string {
	return sourceEntryByKey(report, key).SourceLabel()
}

func sourceEntryByKey(report []config.SourceReportEntry, key string) config.SourceReportEntry {
	for _, entry := range report {
		if entry.Key() == key {
			return entry
		}
	}
	return config.SourceReportEntry{}
}

func assertNoSensitiveCorpus(t *testing.T, text string) {
	t.Helper()
	for _, forbidden := range []string{"dib_fake_secret_value", "dib_fake_password_value", "dib_fake_token_value"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("text leaked sensitive corpus value %q:\n%s", forbidden, text)
		}
	}
}
