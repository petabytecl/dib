package cli_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

func TestResolveSuccessPathFlagBindingWinsOverLowerPrecedence(t *testing.T) {
	t.Parallel()

	child, err := command.NewDefinition("run", command.LocalFlags(
		flags.String("mode", "flag-default", "operation mode"),
	))
	if err != nil {
		t.Fatalf("NewDefinition(run) failed: %v", err)
	}
	root, err := command.NewDefinition("dib", command.Children(child))
	if err != nil {
		t.Fatalf("NewDefinition(dib) failed: %v", err)
	}

	set := mustResolveConfigSet(t)
	env := mustResolveEnv(t, set, map[string]string{"DIB_MODE": "from-env"})
	jsonSnap := mustBindingJSON(t, set, `{"mode":"from-json"}`)

	plan := cli.NewPlan(root, set).
		WithEnv(env).
		WithJSON(jsonSnap).
		WithBindings([]cli.FlagBinding{cli.BindFlag("mode", "mode")})

	inv := cli.FromArgs("dib", []string{"run", "--mode=from-flag", "extra-arg"})
	result, err := cli.Resolve(inv, plan)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	cmd, ok := result.Route().Command()
	if !ok {
		t.Fatal("Route().Command() returned ok=false")
	}
	if got := cmd.Name(); got != "run" {
		t.Fatalf("Route().Command().Name() = %q, want run", got)
	}

	snap, hasFlagSnap := result.FlagSnapshot()
	if !hasFlagSnap {
		t.Fatal("FlagSnapshot() returned hasFlagSnap=false, want true")
	}
	state, ok := snap.Lookup("mode")
	if !ok {
		t.Fatal("FlagSnapshot().Lookup(mode) returned ok=false")
	}
	if !state.Explicit() {
		t.Fatal("FlagSnapshot().Lookup(mode).Explicit() = false, want true")
	}

	assertConfigValue(t, result.Config(), "mode", "from-flag", config.SourceFlagBinding)

	remaining := result.RemainingArgs()
	if want := []string{"extra-arg"}; !reflect.DeepEqual(remaining, want) {
		t.Fatalf("RemainingArgs() = %#v, want %#v", remaining, want)
	}

	if got := result.Invocation().Program(); got != "dib" {
		t.Fatalf("Invocation().Program() = %q, want dib", got)
	}
}

func TestResolveSuccessPathNoFlagDefinitions(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	set := mustResolveConfigSet(t)
	env := mustResolveEnv(t, set, map[string]string{"DIB_MODE": "from-env"})
	jsonSnap := mustBindingJSON(t, set, `{"mode":"from-json"}`)

	plan := cli.NewPlan(root, set).
		WithEnv(env).
		WithJSON(jsonSnap)

	inv := cli.FromArgs("dib", nil)
	result, err := cli.Resolve(inv, plan)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	if _, hasFlagSnap := result.FlagSnapshot(); hasFlagSnap {
		t.Fatal("FlagSnapshot() returned hasFlagSnap=true, want false")
	}

	assertConfigValue(t, result.Config(), "mode", "from-env", config.SourceEnv)
}

func TestResolveSuccessPathZeroBindings(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib", command.LocalFlags(
		flags.String("mode", "flag-default", "mode"),
	))
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	set := mustResolveConfigSet(t)
	jsonSnap := mustBindingJSON(t, set, `{"mode":"from-json"}`)

	plan := cli.NewPlan(root, set).WithJSON(jsonSnap)

	inv := cli.FromArgs("dib", []string{"--mode=from-flag"})
	result, err := cli.Resolve(inv, plan)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	// Zero bindings: flag tier is empty, JSON wins over default
	assertConfigValue(t, result.Config(), "mode", "from-json", config.SourceJSON)
}

func TestResolveRoutingFailureIsInspectable(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib", command.Children(
		mustChild(t, "run"),
	))
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	plan := cli.NewPlan(root, mustResolveConfigSet(t))
	inv := cli.FromArgs("dib", []string{"no-such-command"})
	result, err := cli.Resolve(inv, plan)
	if err == nil {
		t.Fatal("Resolve returned nil error, want routing error")
	}

	if !errors.Is(err, command.ErrUnknownCommand) {
		t.Fatalf("error does not satisfy command.ErrUnknownCommand: %v", err)
	}
	if _, ok := result.Route().Command(); ok {
		t.Fatal("expected zero-value Result on routing failure, but Route().Command() is set")
	}
	if result.Invocation().Program() != "" {
		t.Fatal("expected zero-value Result on routing failure, but Invocation().Program() is set")
	}
}

func TestResolveBindingFailureAfterRoutingSuccessIsInspectable(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib", command.LocalFlags(
		flags.String("mode", "default", "mode"),
	))
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	plan := cli.NewPlan(root, mustResolveConfigSet(t)).
		WithBindings([]cli.FlagBinding{cli.BindFlag("no-such-flag", "mode")})

	inv := cli.FromArgs("dib", []string{"--mode=from-flag"})
	result, err := cli.Resolve(inv, plan)
	if err == nil {
		t.Fatal("Resolve returned nil error, want binding error")
	}

	if !errors.Is(err, cli.ErrUnknownFlagBinding) {
		t.Fatalf("error does not satisfy cli.ErrUnknownFlagBinding: %v", err)
	}

	var bindErr *cli.BindingError
	if !errors.As(err, &bindErr) {
		t.Fatalf("error does not expose *cli.BindingError: %T", err)
	}

	if _, ok := result.Route().Command(); ok {
		t.Fatal("expected zero-value Result on binding failure, but Route().Command() is set")
	}
	if result.Invocation().Program() != "" {
		t.Fatal("expected zero-value Result on binding failure, but Invocation().Program() is set")
	}
}

func TestResolveSensitiveValueReachableNotLeakedInErrors(t *testing.T) {
	t.Parallel()

	const fakeSecret = "dib_fake_secret_value"

	set, err := config.NewSet(
		config.String("mode", "default", "mode"),
		config.String("api-key", "", "api key", config.Sensitive()),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}

	envSnap, err := config.NewEnvSnapshot(set, func(name string) (string, bool) {
		switch name {
		case "DIB_API_KEY":
			return fakeSecret, true
		}
		return "", false
	}, []config.EnvBinding{
		config.BindEnv("api-key", "DIB_API_KEY"),
	})
	if err != nil {
		t.Fatalf("NewEnvSnapshot returned unexpected error: %v", err)
	}

	t.Run("success: sensitive value reachable via config getter", func(t *testing.T) {
		t.Parallel()

		root, err := command.NewDefinition("dib")
		if err != nil {
			t.Fatalf("NewDefinition returned unexpected error: %v", err)
		}

		plan := cli.NewPlan(root, set).WithEnv(envSnap)
		result, err := cli.Resolve(cli.FromArgs("dib", nil), plan)
		if err != nil {
			t.Fatalf("Resolve returned unexpected error: %v", err)
		}

		got, err := result.Config().GetString("api-key")
		if err != nil {
			t.Fatalf("Config().GetString(api-key) returned unexpected error: %v", err)
		}
		if got != fakeSecret {
			t.Fatalf("Config().GetString(api-key) = %q, want %q", got, fakeSecret)
		}
	})

	t.Run("binding error does not leak sensitive env value", func(t *testing.T) {
		t.Parallel()

		root, err := command.NewDefinition("dib", command.LocalFlags(
			flags.String("mode", "default", "mode"),
		))
		if err != nil {
			t.Fatalf("NewDefinition returned unexpected error: %v", err)
		}

		plan := cli.NewPlan(root, set).
			WithEnv(envSnap).
			WithBindings([]cli.FlagBinding{cli.BindFlag("no-such-flag", "mode")})

		_, err = cli.Resolve(cli.FromArgs("dib", []string{"--mode=x"}), plan)
		if err == nil {
			t.Fatal("Resolve returned nil error, want binding error")
		}

		if strings.Contains(err.Error(), fakeSecret) {
			t.Fatalf("error leaked sensitive value %q: %v", fakeSecret, err)
		}
	})
}

func TestResolveRemainingArgsDefensiveCopy(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	plan := cli.NewPlan(root, mustResolveConfigSet(t))
	inv := cli.FromArgs("dib", []string{"posarg1", "posarg2"})
	result, err := cli.Resolve(inv, plan)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	first := result.RemainingArgs()
	if len(first) == 0 {
		t.Fatal("RemainingArgs() returned empty slice, want non-empty")
	}
	first[0] = "mutated"

	second := result.RemainingArgs()
	if reflect.DeepEqual(first, second) {
		t.Fatal("RemainingArgs() returned same slice on second call; want defensive copy")
	}
	if second[0] == "mutated" {
		t.Fatalf("RemainingArgs()[0] = %q after mutation, want original value", second[0])
	}
}

func TestCLIImportsFlagsAfterResolve(t *testing.T) {
	t.Parallel()

	out := goListImports(t, "./cli")
	imports := strings.Fields(out)
	if !containsString(imports, "github.com/petabytecl/dib/flags") {
		t.Fatalf("./cli imports = %v, want github.com/petabytecl/dib/flags", imports)
	}
}

func mustResolveConfigSet(t *testing.T) config.Set {
	t.Helper()
	set, err := config.NewSet(
		config.String("mode", "config-default", "mode"),
		config.Int("workers", 1, "workers"),
	)
	if err != nil {
		t.Fatalf("NewSet returned unexpected error: %v", err)
	}
	return set
}

func mustResolveEnv(t *testing.T, set config.Set, values map[string]string) config.Snapshot {
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

func mustChild(t *testing.T, name string) command.Definition {
	t.Helper()
	def, err := command.NewDefinition(name)
	if err != nil {
		t.Fatalf("NewDefinition(%q) returned unexpected error: %v", name, err)
	}
	return def
}
