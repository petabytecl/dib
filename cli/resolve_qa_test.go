package cli_test

import (
	"testing"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

// TestQAResolveFullPrecedenceChainWithAllSourceTiers proves that when all five
// source tiers are supplied — explicit, flag binding, env, JSON, and default —
// config.Resolve selects the highest-precedence value (explicit) without calling
// os.Exit, writing to stdout/stderr, or reading os.Args implicitly.
func TestQAResolveFullPrecedenceChainWithAllSourceTiers(t *testing.T) {
	t.Parallel()

	child, err := command.NewDefinition("deploy", command.LocalFlags(
		flags.String("mode", "flag-default", "deploy mode"),
	))
	if err != nil {
		t.Fatalf("NewDefinition(deploy): %v", err)
	}
	root, err := command.NewDefinition("dib", command.Children(child))
	if err != nil {
		t.Fatalf("NewDefinition(dib): %v", err)
	}

	set, err := config.NewSet(config.String("mode", "config-default", "deploy mode"))
	if err != nil {
		t.Fatalf("config.NewSet: %v", err)
	}

	// All five tiers supply different values for "mode" so precedence is observable.
	explicit, err := config.NewExplicitSnapshot(set, config.Assignment{Key: "mode", Value: "from-explicit"})
	if err != nil {
		t.Fatalf("config.NewExplicitSnapshot: %v", err)
	}
	envSnap, err := config.NewEnvSnapshot(set,
		func(name string) (string, bool) {
			if name == "DIB_MODE" {
				return "from-env", true
			}
			return "", false
		},
		[]config.EnvBinding{config.BindEnv("mode", "DIB_MODE")},
	)
	if err != nil {
		t.Fatalf("config.NewEnvSnapshot: %v", err)
	}
	jsonSnap := mustBindingJSON(t, set, `{"mode":"from-json"}`)

	// WithExplicit is the explicit-setter tier; it must beat the flag binding tier.
	plan := cli.NewPlan(root, set).
		WithExplicit(explicit).
		WithEnv(envSnap).
		WithJSON(jsonSnap).
		WithBindings([]cli.FlagBinding{cli.BindFlag("mode", "mode")})

	// The flag is set on the command line but explicit must still win.
	result, err := cli.Resolve(
		cli.FromArgs("dib", []string{"deploy", "--mode=from-flag", "extra"}),
		plan,
	)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	if got := result.Invocation().Program(); got != "dib" {
		t.Fatalf("Invocation().Program() = %q, want dib", got)
	}
	cmd, ok := result.Route().Command()
	if !ok {
		t.Fatal("Route().Command() returned ok=false")
	}
	if cmd.Name() != "deploy" {
		t.Fatalf("Route().Command().Name() = %q, want deploy", cmd.Name())
	}
	remaining := result.RemainingArgs()
	if len(remaining) != 1 || remaining[0] != "extra" {
		t.Fatalf("RemainingArgs() = %#v, want [extra]", remaining)
	}

	// Explicit setter beats flag binding, env, JSON, and config default.
	assertConfigValue(t, result.Config(), "mode", "from-explicit", config.SourceExplicit)
}

// TestQAResolvePlanWithExplicitSnapshotAloneCoversRequiredAC1Path verifies AC1:
// when only WithExplicit is supplied with no flag bindings, the resolved config
// returns the explicit value and no flag snapshot is present.
func TestQAResolvePlanWithExplicitSnapshotAloneCoversRequiredAC1Path(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition: %v", err)
	}
	set, err := config.NewSet(config.String("mode", "default-value", "mode"))
	if err != nil {
		t.Fatalf("config.NewSet: %v", err)
	}
	explicit, err := config.NewExplicitSnapshot(set, config.Assignment{Key: "mode", Value: "caller-set"})
	if err != nil {
		t.Fatalf("config.NewExplicitSnapshot: %v", err)
	}

	plan := cli.NewPlan(root, set).WithExplicit(explicit)
	result, err := cli.Resolve(cli.FromArgs("dib", nil), plan)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	// No flag definitions on the root command, so FlagSnapshot() must be absent.
	if _, ok := result.FlagSnapshot(); ok {
		t.Fatal("FlagSnapshot() = _, true; want false (no flag definitions on command)")
	}
	assertConfigValue(t, result.Config(), "mode", "caller-set", config.SourceExplicit)
}

// TestQAResolvePlanExplicitSnapshotWinsOverFlagBinding verifies the precedence
// order at the boundary between the two highest-priority tiers: explicit setter
// overrides an explicit flag binding value when both are supplied.
func TestQAResolvePlanExplicitSnapshotWinsOverFlagBinding(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib", command.LocalFlags(
		flags.String("mode", "flag-default", "mode"),
	))
	if err != nil {
		t.Fatalf("NewDefinition: %v", err)
	}
	set, err := config.NewSet(config.String("mode", "config-default", "mode"))
	if err != nil {
		t.Fatalf("config.NewSet: %v", err)
	}
	explicit, err := config.NewExplicitSnapshot(set, config.Assignment{Key: "mode", Value: "from-explicit"})
	if err != nil {
		t.Fatalf("config.NewExplicitSnapshot: %v", err)
	}

	plan := cli.NewPlan(root, set).
		WithExplicit(explicit).
		WithBindings([]cli.FlagBinding{cli.BindFlag("mode", "mode")})

	// Flag is set to a different value; explicit must still win.
	result, err := cli.Resolve(
		cli.FromArgs("dib", []string{"--mode=from-flag"}),
		plan,
	)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}
	assertConfigValue(t, result.Config(), "mode", "from-explicit", config.SourceExplicit)
}

// TestQAResolvePlanBindingsIsDefensivelyCopied verifies that mutating the slice
// returned by Plan.Bindings() does not affect subsequent calls or the behaviour
// of Resolve — consistent with the immutable-value requirement for Plan.
func TestQAResolvePlanBindingsIsDefensivelyCopied(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib", command.LocalFlags(
		flags.String("mode", "default", "mode"),
	))
	if err != nil {
		t.Fatalf("NewDefinition: %v", err)
	}

	plan := cli.NewPlan(root, mustResolveConfigSet(t)).
		WithBindings([]cli.FlagBinding{cli.BindFlag("mode", "mode")})

	first := plan.Bindings()
	if len(first) == 0 {
		t.Fatal("Bindings() returned empty slice, want non-empty")
	}
	original := first[0].FlagName
	first[0] = cli.BindFlag("mutated", "mode")

	second := plan.Bindings()
	if len(second) == 0 {
		t.Fatal("Bindings() second call returned empty slice")
	}
	if second[0].FlagName != original {
		t.Fatalf("Bindings()[0].FlagName = %q after mutation of first return, want %q; Plan.Bindings() is not a defensive copy", second[0].FlagName, original)
	}

	// Resolve still uses the original binding, not the mutated one.
	inv := cli.FromArgs("dib", []string{"--mode=live"})
	if _, err := cli.Resolve(inv, plan); err != nil {
		t.Fatalf("Resolve with original bindings returned unexpected error: %v", err)
	}
}

// TestQAResolvePlanCallerSuppliedInputsAreNotMutated verifies the immutable-plan
// contract: WithBindings defensively copies the caller's slice so that mutations
// to the original slice after plan construction do not affect the plan.
func TestQAResolvePlanCallerSuppliedInputsAreNotMutated(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib", command.LocalFlags(
		flags.String("mode", "default", "mode"),
	))
	if err != nil {
		t.Fatalf("NewDefinition: %v", err)
	}

	callerBindings := []cli.FlagBinding{cli.BindFlag("mode", "mode")}
	plan := cli.NewPlan(root, mustResolveConfigSet(t)).WithBindings(callerBindings)

	// Mutate the caller's original slice after plan creation.
	callerBindings[0] = cli.BindFlag("mutated", "mode")

	// Resolve must still succeed using the defensively copied original.
	inv := cli.FromArgs("dib", []string{"--mode=live"})
	if _, err := cli.Resolve(inv, plan); err != nil {
		t.Fatalf("Resolve failed after caller mutation: plan did not copy the input slice: %v", err)
	}
}
