package cli_test

import (
	"reflect"
	"testing"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/flags"
)

func TestResultInvocationAccessor(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	set := mustResolveConfigSet(t)
	inv := cli.FromArgs("dib", []string{"arg1"})
	result, err := cli.Resolve(inv, cli.NewPlan(root, set))
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	got := result.Invocation()
	if got.Program() != "dib" {
		t.Fatalf("Invocation().Program() = %q, want dib", got.Program())
	}
	if args := got.Args(); !reflect.DeepEqual(args, []string{"arg1"}) {
		t.Fatalf("Invocation().Args() = %#v, want [arg1]", args)
	}
}

func TestResultRouteAccessor(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	set := mustResolveConfigSet(t)
	result, err := cli.Resolve(cli.FromArgs("dib", nil), cli.NewPlan(root, set))
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	cmd, ok := result.Route().Command()
	if !ok {
		t.Fatal("Route().Command() returned ok=false")
	}
	if cmd.Name() != "dib" {
		t.Fatalf("Route().Command().Name() = %q, want dib", cmd.Name())
	}
}

func TestResultFlagSnapshotPresentAccessor(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib", command.LocalFlags(
		flags.String("mode", "default", "mode"),
	))
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	set := mustResolveConfigSet(t)
	result, err := cli.Resolve(
		cli.FromArgs("dib", []string{"--mode=explicit"}),
		cli.NewPlan(root, set).WithBindings([]cli.FlagBinding{cli.BindFlag("mode", "mode")}),
	)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	snap, ok := result.FlagSnapshot()
	if !ok {
		t.Fatal("FlagSnapshot() returned ok=false, want true")
	}
	state, exists := snap.Lookup("mode")
	if !exists {
		t.Fatal("FlagSnapshot().Lookup(mode) returned ok=false")
	}
	if !state.Explicit() {
		t.Fatal("FlagSnapshot().Lookup(mode).Explicit() = false, want true")
	}
}

func TestResultFlagSnapshotAbsentAccessor(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	set := mustResolveConfigSet(t)
	result, err := cli.Resolve(cli.FromArgs("dib", nil), cli.NewPlan(root, set))
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	snap, ok := result.FlagSnapshot()
	if ok {
		t.Fatal("FlagSnapshot() returned ok=true, want false")
	}
	// Zero-value snapshot has no entries; verify no lookup succeeds.
	if _, exists := snap.Lookup("mode"); exists {
		t.Fatal("zero-value FlagSnapshot() Lookup(mode) returned ok=true, want false")
	}
}

func TestResultConfigAccessor(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	set := mustResolveConfigSet(t)
	jsonSnap := mustBindingJSON(t, set, `{"mode":"from-json"}`)
	result, err := cli.Resolve(
		cli.FromArgs("dib", nil),
		cli.NewPlan(root, set).WithJSON(jsonSnap),
	)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	got, err := result.Config().GetString("mode")
	if err != nil {
		t.Fatalf("Config().GetString(mode) returned unexpected error: %v", err)
	}
	if got != "from-json" {
		t.Fatalf("Config().GetString(mode) = %q, want from-json", got)
	}
}

func TestResultRemainingArgsDefensiveCopy(t *testing.T) {
	t.Parallel()

	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	set := mustResolveConfigSet(t)
	result, err := cli.Resolve(
		cli.FromArgs("dib", []string{"first", "second"}),
		cli.NewPlan(root, set),
	)
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}

	first := result.RemainingArgs()
	if len(first) == 0 {
		t.Fatal("RemainingArgs() returned empty, want non-empty")
	}
	original := first[0]
	first[0] = "mutated"

	second := result.RemainingArgs()
	if len(second) == 0 || second[0] == "mutated" {
		t.Fatalf("RemainingArgs() second call = %#v; want %q in position 0", second, original)
	}
	if !reflect.DeepEqual(second, []string{"first", "second"}) {
		t.Fatalf("RemainingArgs() = %#v, want [first second]", second)
	}
}
