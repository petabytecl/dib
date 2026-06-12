package migration_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/flags"
)

func Example_nestedCommandMigration() {
	root := mustMigrationCommandTree()

	result, err := root.Route([]string{
		"--profile", "prod",
		"deploy",
		"push",
		"--dry-run",
		"manifest.json",
	})
	if err != nil {
		panic(err)
	}

	snapshot, _ := result.FlagSnapshot()
	composedFlags, _ := result.Flags()
	profile, _ := snapshot.Lookup("profile")
	dryRun, _ := snapshot.Lookup("dry-run")

	var help bytes.Buffer
	if err := result.WriteHelp(&help); err != nil {
		panic(err)
	}

	fmt.Println(result.PathNames())
	fmt.Println(result.MatchTokens())
	fmt.Println(result.RemainingArgs())
	fmt.Println(profile.Values()[0], dryRun.Values()[0], composedFlags.Len(), strings.Contains(help.String(), "Flags:"))
	// Output:
	// [dib deploy apply]
	// [dib deploy push]
	// [manifest.json]
	// prod true 3 true
}

func TestNestedCommandMigrationRoutesAliasesFlagsAndHelp(t *testing.T) {
	root := mustMigrationCommandTree()

	result, err := root.Route([]string{
		"--profile=prod",
		"ship",
		"push",
		"--dry-run",
		"manifest.json",
	})
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("PathNames() = %#v, want deploy apply", got)
	}
	if got := result.MatchTokens(); !reflect.DeepEqual(got, []string{"dib", "ship", "push"}) {
		t.Fatalf("MatchTokens() = %#v, want alias tokens", got)
	}
	if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"manifest.json"}) {
		t.Fatalf("RemainingArgs() = %#v, want manifest", got)
	}

	snapshot, ok := result.FlagSnapshot()
	if !ok {
		t.Fatal("FlagSnapshot() returned ok=false")
	}
	assertFlagValues(t, snapshot, "profile", []any{"prod"}, true)
	assertFlagValues(t, snapshot, "dry-run", []any{true}, true)

	composedFlags, ok := result.Flags()
	if !ok {
		t.Fatal("Flags() returned ok=false")
	}
	if got := composedFlags.Len(); got != 3 {
		t.Fatalf("Flags().Len() = %d, want inherited and local flags", got)
	}
	for _, name := range []string{"profile", "region", "dry-run"} {
		if _, ok := composedFlags.Lookup(name); !ok {
			t.Fatalf("Flags().Lookup(%q) returned ok=false", name)
		}
	}

	var routedHelp bytes.Buffer
	if err := result.WriteHelp(&routedHelp); err != nil {
		t.Fatalf("Result.WriteHelp: %v", err)
	}
	for _, phrase := range []string{"Usage:", "dib deploy apply", "--profile", "--dry-run"} {
		if !strings.Contains(routedHelp.String(), phrase) {
			t.Fatalf("routed help missing %q:\n%s", phrase, routedHelp.String())
		}
	}

	var rootHelp bytes.Buffer
	if err := root.WriteHelp(&rootHelp); err != nil {
		t.Fatalf("Definition.WriteHelp: %v", err)
	}
	if !strings.Contains(rootHelp.String(), "deploy") || !strings.Contains(rootHelp.String(), "status") {
		t.Fatalf("root help omitted child commands:\n%s", rootHelp.String())
	}
}

func TestNestedCommandMigrationKeepsBoundaryCallerOwned(t *testing.T) {
	root := mustMigrationCommandTree()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	args := []string{"deploy", "apply", "--dry-run", "manifest.json"}

	boundary, err := root.RouteBoundary(context.Background(), args, &stdout, &stderr)
	if err != nil {
		t.Fatalf("RouteBoundary: %v", err)
	}
	args[0] = "mutated"

	if got := boundary.Args(); !reflect.DeepEqual(got, []string{"deploy", "apply", "--dry-run", "manifest.json"}) {
		t.Fatalf("Args() = %#v, want original explicit args", got)
	}
	result, ok := boundary.Result()
	if !ok {
		t.Fatal("Result() returned ok=false")
	}
	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("PathNames() = %#v, want routed command", got)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("RouteBoundary wrote to caller writers: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestNestedCommandMigrationTypedErrors(t *testing.T) {
	root := mustMigrationCommandTree()

	result, err := root.Route([]string{"deploy", "missing"})
	if err == nil {
		t.Fatal("Route returned nil error")
	}
	if !errors.Is(err, command.ErrUnknownCommand) {
		t.Fatalf("error %v does not satisfy ErrUnknownCommand", err)
	}
	var unknown *command.UnknownCommandError
	if !errors.As(err, &unknown) {
		t.Fatalf("error does not expose *command.UnknownCommandError: %T", err)
	}
	if got := unknown.ParentPath(); !reflect.DeepEqual(got, []string{"dib", "deploy"}) {
		t.Fatalf("ParentPath() = %#v, want deploy parent", got)
	}
	if got := result.PathNames(); len(got) != 0 {
		t.Fatalf("failed route returned path %#v", got)
	}
}

func mustMigrationCommandTree() command.Definition {
	apply := mustCommandDefinition(
		"apply",
		command.Aliases("push"),
		command.Usage("[flags] manifest"),
		command.Description("apply a manifest"),
		command.LocalFlags(flags.Bool("dry-run", false, "preview changes")),
	)
	plan := mustCommandDefinition(
		"plan",
		command.Aliases("preview"),
		command.Description("show a plan"),
	)
	deploy := mustCommandDefinition(
		"deploy",
		command.Aliases("ship"),
		command.Description("deployment commands"),
		command.Children(apply, plan),
		command.InheritedFlags(flags.String("region", "us", "target region")),
	)
	status := mustCommandDefinition("status", command.Description("show status"))
	return mustCommandDefinition(
		"dib",
		command.Usage("<command>"),
		command.Description("migration example root"),
		command.Children(deploy, status),
		command.InheritedFlags(flags.String("profile", "dev", "profile name")),
	)
}

func mustCommandDefinition(name string, options ...command.Option) command.Definition {
	definition, err := command.NewDefinition(name, options...)
	if err != nil {
		panic(err)
	}
	return definition
}
