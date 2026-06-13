package cli_test

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

func TestNewFlagSnapshotExplicitFlagBindingWinsOverLowerPrecedence(t *testing.T) {
	t.Parallel()

	set := mustBindingConfigSet(t)
	route := mustBindingRoute(t, []string{"--mode=from-flag"}, flags.String("mode", "from-flag-default", "mode"))
	env := mustBindingEnv(t, set, map[string]string{"DIB_MODE": "from-env"})
	jsonSource := mustBindingJSON(t, set, `{"mode":"from-json"}`)

	flagSource, err := cli.NewFlagSnapshot(set, route, []cli.FlagBinding{
		cli.BindFlag("mode", "mode"),
	})
	if err != nil {
		t.Fatalf("NewFlagSnapshot returned unexpected error: %v", err)
	}

	resolved := config.Resolve(set, config.Snapshot{}, flagSource, env, jsonSource)
	assertConfigValue(t, resolved, "mode", "from-flag", config.SourceFlagBinding)
}

func TestNewFlagSnapshotDefaultFlagDoesNotOverrideLowerPrecedence(t *testing.T) {
	t.Parallel()

	set := mustBindingConfigSet(t)
	route := mustBindingRoute(t, nil, flags.String("mode", "from-flag-default", "mode"))

	flagSource, err := cli.NewFlagSnapshot(set, route, []cli.FlagBinding{
		cli.BindFlag("mode", "mode"),
	})
	if err != nil {
		t.Fatalf("NewFlagSnapshot returned unexpected error: %v", err)
	}

	flagValue, ok := flagSource.Lookup("mode")
	if !ok {
		t.Fatal("flag source Lookup(mode) returned ok=false")
	}
	if got, hasValue := flagValue.Value(); hasValue {
		t.Fatalf("flag source Value() = %#v, true; want absent", got)
	}

	t.Run("env wins over JSON and default", func(t *testing.T) {
		t.Parallel()

		env := mustBindingEnv(t, set, map[string]string{"DIB_MODE": "from-env"})
		jsonSource := mustBindingJSON(t, set, `{"mode":"from-json"}`)
		resolved := config.Resolve(set, config.Snapshot{}, flagSource, env, jsonSource)
		assertConfigValue(t, resolved, "mode", "from-env", config.SourceEnv)
	})

	t.Run("JSON wins over default when env is absent", func(t *testing.T) {
		t.Parallel()

		jsonSource := mustBindingJSON(t, set, `{"mode":"from-json"}`)
		resolved := config.Resolve(set, config.Snapshot{}, flagSource, config.Snapshot{}, jsonSource)
		assertConfigValue(t, resolved, "mode", "from-json", config.SourceJSON)
	})

	t.Run("config default wins when lower sources are absent", func(t *testing.T) {
		t.Parallel()

		resolved := config.Resolve(set, config.Snapshot{}, flagSource, config.Snapshot{}, config.Snapshot{})
		assertConfigValue(t, resolved, "mode", "from-config-default", config.SourceDefault)
	})
}

func TestNewFlagSnapshotAbsentKnownBindingLeavesFlagTierAbsent(t *testing.T) {
	t.Parallel()

	set := mustBindingConfigSet(t)
	route := mustBindingRoute(t, []string{"--workers=4"}, flags.String("mode", "from-flag-default", "mode"), flags.Int("workers", 1, "workers"))

	flagSource, err := cli.NewFlagSnapshot(set, route, []cli.FlagBinding{
		cli.BindFlag("mode", "mode"),
	})
	if err != nil {
		t.Fatalf("NewFlagSnapshot returned unexpected error: %v", err)
	}

	value, ok := flagSource.Lookup("mode")
	if !ok {
		t.Fatal("Lookup(mode) returned ok=false")
	}
	if got, hasValue := value.Value(); hasValue {
		t.Fatalf("flag source Value() = %#v, true; want absent", got)
	}
}

func TestNewFlagSnapshotZeroBindingsDoesNotRequireRouteFlagSnapshot(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}
	route, err := root.Route(nil)
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}

	snapshot, err := cli.NewFlagSnapshot(mustBindingConfigSet(t), route, nil)
	if err != nil {
		t.Fatalf("NewFlagSnapshot returned unexpected error: %v", err)
	}
	if snapshot.IsSet("mode") {
		t.Fatal("empty flag-tier snapshot unexpectedly set mode")
	}
}

func TestNewFlagSnapshotBindingRequiresRouteFlagSnapshot(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}
	route, err := root.Route(nil)
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}

	_, err = cli.NewFlagSnapshot(mustBindingConfigSet(t), route, []cli.FlagBinding{
		cli.BindFlag("mode", "mode"),
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error")
	}
	assertBindingError(t, err, cli.ErrInvalidBinding, "", "", nil)
}

func TestNewFlagSnapshotUnknownFlagBindingIsInspectable(t *testing.T) {
	t.Parallel()

	route := mustBindingRoute(t, []string{"--mode=from-flag"}, flags.String("mode", "from-flag-default", "mode"))

	_, err := cli.NewFlagSnapshot(mustBindingConfigSet(t), route, []cli.FlagBinding{
		cli.BindFlag("missing", "mode"),
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error")
	}
	assertBindingError(t, err, cli.ErrUnknownFlagBinding, "missing", "mode", nil)
}

func TestNewFlagSnapshotUnknownConfigKeyWrapsConfigCause(t *testing.T) {
	t.Parallel()

	route := mustBindingRoute(t, []string{"--mode=from-flag"}, flags.String("mode", "from-flag-default", "mode"))

	_, err := cli.NewFlagSnapshot(mustBindingConfigSet(t), route, []cli.FlagBinding{
		cli.BindFlag("mode", "missing-config"),
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error")
	}
	assertBindingError(t, err, cli.ErrInvalidBinding, "mode", "missing-config", config.ErrUnknownSourceKey)

	var sourceErr *config.SourceError
	if !errors.As(err, &sourceErr) {
		t.Fatalf("error does not expose *config.SourceError: %T", err)
	}
	if got := sourceErr.Key(); got != "missing-config" {
		t.Fatalf("SourceError.Key() = %q, want missing-config", got)
	}
	if got := sourceErr.Source(); got != config.SourceFlagBinding {
		t.Fatalf("SourceError.Source() = %q, want %q", got, config.SourceFlagBinding)
	}
}

func TestNewFlagSnapshotDuplicateConfigBindingIsInspectable(t *testing.T) {
	t.Parallel()

	route := mustBindingRoute(t, []string{"--mode=from-flag", "--workers=4"}, flags.String("mode", "", "mode"), flags.Int("workers", 1, "workers"))

	_, err := cli.NewFlagSnapshot(mustBindingConfigSet(t), route, []cli.FlagBinding{
		cli.BindFlag("mode", "mode"),
		cli.BindFlag("workers", "mode"),
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error")
	}
	assertBindingError(t, err, cli.ErrInvalidBinding, "workers", "mode", config.ErrDuplicateBinding)

	var sourceErr *config.SourceError
	if !errors.As(err, &sourceErr) {
		t.Fatalf("error does not expose *config.SourceError: %T", err)
	}
	if got := sourceErr.Key(); got != "mode" {
		t.Fatalf("SourceError.Key() = %q, want mode", got)
	}
	if got := sourceErr.Source(); got != config.SourceFlagBinding {
		t.Fatalf("SourceError.Source() = %q, want %q", got, config.SourceFlagBinding)
	}
}

func TestNewFlagSnapshotStringListTranslationAndDefensiveCopies(t *testing.T) {
	t.Parallel()

	set := mustBindingConfigSet(t)
	route := mustBindingRoute(t, []string{"--tag=alpha", "--tag=beta"}, flags.StringList("tag", nil, "tags", flags.Repeatable()))

	flagSource, err := cli.NewFlagSnapshot(set, route, []cli.FlagBinding{
		cli.BindFlag("tag", "tag"),
	})
	if err != nil {
		t.Fatalf("NewFlagSnapshot returned unexpected error: %v", err)
	}

	first, err := flagSource.GetStringList("tag")
	if err != nil {
		t.Fatalf("GetStringList returned unexpected error: %v", err)
	}
	if want := []string{"alpha", "beta"}; !reflect.DeepEqual(first, want) {
		t.Fatalf("GetStringList = %#v, want %#v", first, want)
	}
	first[0] = "mutated"
	second, err := flagSource.GetStringList("tag")
	if err != nil {
		t.Fatalf("GetStringList after mutation returned unexpected error: %v", err)
	}
	if want := []string{"alpha", "beta"}; !reflect.DeepEqual(second, want) {
		t.Fatalf("GetStringList after mutation = %#v, want %#v", second, want)
	}
}

func TestNewFlagSnapshotSensitiveBindingErrorsAreValueFree(t *testing.T) {
	t.Parallel()

	const secret = "dib_fake_secret_value"
	const password = "dib_fake_password_value"
	const token = "dib_fake_token_value"
	set := mustBindingConfigSet(t)
	route := mustBindingRoute(t, []string{"--secret=" + secret}, flags.String("secret", "", "secret", flags.Sensitive()))

	_, err := cli.NewFlagSnapshot(set, route, []cli.FlagBinding{
		cli.BindFlag("secret", "secret"),
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error")
	}
	if !errors.Is(err, config.ErrSourceConversion) {
		t.Fatalf("error does not satisfy ErrSourceConversion: %v", err)
	}
	for _, forbidden := range []string{secret, password, token} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("binding error leaked sensitive value %q: %v", forbidden, err)
		}
	}
	var sourceErr *config.SourceError
	if !errors.As(err, &sourceErr) {
		t.Fatalf("error does not expose *config.SourceError: %T", err)
	}
	if !sourceErr.Redacted() {
		t.Fatal("SourceError.Redacted() = false, want true")
	}
}

func TestNewFlagSnapshotConversionBindingErrorDoesNotEchoFlagValue(t *testing.T) {
	t.Parallel()

	const raw = "dib_non_sensitive_invalid_int"
	set := mustBindingConfigSet(t)
	route := mustBindingRoute(t, []string{"--mode=" + raw}, flags.String("mode", "", "mode"))

	_, err := cli.NewFlagSnapshot(set, route, []cli.FlagBinding{
		cli.BindFlag("mode", "workers"),
	})
	if err == nil {
		t.Fatal("NewFlagSnapshot returned nil error")
	}
	if !errors.Is(err, config.ErrSourceConversion) {
		t.Fatalf("error does not satisfy ErrSourceConversion: %v", err)
	}
	if strings.Contains(err.Error(), raw) {
		t.Fatalf("binding error leaked raw flag value %q: %v", raw, err)
	}
	if !strings.Contains(err.Error(), config.SourceFlagBinding) {
		t.Fatalf("binding error = %q, want source context %q", err.Error(), config.SourceFlagBinding)
	}
	var sourceErr *config.SourceError
	if !errors.As(err, &sourceErr) {
		t.Fatalf("error does not expose *config.SourceError: %T", err)
	}
	if !strings.Contains(sourceErr.Error(), raw) {
		t.Fatalf("wrapped non-sensitive source error = %q, want raw value for config diagnostics", sourceErr.Error())
	}
}

func TestPackageImportBoundariesForCLIComposition(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		pkg            string
		mayImportCLI   bool
		wantCLIImports []string
	}{
		{
			pkg:          "./cli",
			mayImportCLI: true,
			wantCLIImports: []string{
				"github.com/petabytecl/dib/command",
				"github.com/petabytecl/dib/config",
			},
		},
		{pkg: "./command"},
		{pkg: "./flags"},
		{pkg: "./config"},
	} {
		t.Run(tc.pkg, func(t *testing.T) {
			t.Parallel()

			out := goListImports(t, tc.pkg)
			imports := strings.Fields(out)
			if !tc.mayImportCLI {
				for _, imp := range imports {
					if imp == "github.com/petabytecl/dib/cli" {
						t.Fatalf("%s imports cli: %v", tc.pkg, imports)
					}
				}
				return
			}

			for _, want := range tc.wantCLIImports {
				if !containsString(imports, want) {
					t.Fatalf("%s imports = %v, want %s", tc.pkg, imports, want)
				}
			}
		})
	}
}

func mustBindingConfigSet(t *testing.T) config.Set {
	t.Helper()
	set, err := config.NewSet(
		config.String("mode", "from-config-default", "mode"),
		config.Int("workers", 1, "workers"),
		config.StringList("tag", []string{"from-config-default"}, "tags"),
		config.Define("secret", config.KindInt, "secret", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	return set
}

func mustBindingRoute(t *testing.T, args []string, defs ...flags.Definition) command.Result {
	t.Helper()
	root, err := command.NewDefinition("dib", command.LocalFlags(defs...))
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}
	route, err := root.Route(args)
	if err != nil {
		t.Fatalf("Route(%v) returned unexpected error: %v", args, err)
	}
	return route
}

func mustBindingEnv(t *testing.T, set config.Set, values map[string]string) config.Snapshot {
	t.Helper()
	snapshot, err := config.NewEnvSnapshot(set, func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}, []config.EnvBinding{
		config.BindEnv("mode", "DIB_MODE"),
	})
	if err != nil {
		t.Fatalf("NewEnvSnapshot returned unexpected error: %v", err)
	}
	return snapshot
}

func mustBindingJSON(t *testing.T, set config.Set, body string) config.Snapshot {
	t.Helper()
	snapshot, err := config.LoadJSON(set, strings.NewReader(body), config.JSONReaderLabel("inline cli binding test"))
	if err != nil {
		t.Fatalf("LoadJSON returned unexpected error: %v", err)
	}
	return snapshot
}

func assertConfigValue(t *testing.T, snapshot config.Snapshot, key string, wantValue any, wantSource string) {
	t.Helper()
	value, ok := snapshot.Lookup(key)
	if !ok {
		t.Fatalf("Lookup(%q) returned ok=false", key)
	}
	got, hasValue := value.Value()
	if !hasValue || !reflect.DeepEqual(got, wantValue) {
		t.Fatalf("%s Value() = %#v, %t; want %#v, true", key, got, hasValue, wantValue)
	}
	if got := value.Provenance(); got != wantSource {
		t.Fatalf("%s Provenance() = %q, want %q", key, got, wantSource)
	}
}

func assertBindingError(t *testing.T, err, category error, flagName, configKey string, cause error) {
	t.Helper()
	if !errors.Is(err, category) {
		t.Fatalf("error does not satisfy %v: %v", category, err)
	}
	if cause != nil && !errors.Is(err, cause) {
		t.Fatalf("error does not satisfy cause %v: %v", cause, err)
	}
	var bindingErr *cli.BindingError
	if !errors.As(err, &bindingErr) {
		t.Fatalf("error does not expose *cli.BindingError: %T", err)
	}
	if got := bindingErr.FlagName(); got != flagName {
		t.Fatalf("BindingError.FlagName() = %q, want %q", got, flagName)
	}
	if got := bindingErr.ConfigKey(); got != configKey {
		t.Fatalf("BindingError.ConfigKey() = %q, want %q", got, configKey)
	}
	if got := bindingErr.Category(); !errors.Is(got, category) {
		t.Fatalf("BindingError.Category() = %v, want %v", got, category)
	}
	if cause != nil && bindingErr.Cause() == nil {
		t.Fatal("BindingError.Cause() = nil, want wrapped cause")
	}
}

func goListImports(t *testing.T, pkg string) string {
	t.Helper()
	cmd := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", pkg)
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/dib-go-cache")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list imports for %s failed: %v\n%s", pkg, err, out)
	}
	return string(out)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
