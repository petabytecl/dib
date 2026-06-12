package command_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/petabytecl/dib/command"
)

func TestAliasRoutingPublicAPIWorkflow(t *testing.T) {
	apply := mustDefinition(t, "apply", command.Aliases("push"))
	plan := mustDefinition(t, "plan", command.Aliases("preview"))
	deploy := mustDefinition(t, "deploy", command.Aliases("ship"), command.Children(apply, plan))
	status := mustDefinition(t, "status")
	root := mustDefinition(t, "dib", command.Children(deploy, status))

	result, err := root.Route([]string{"ship", "push", "manifest.yaml", "--env", "prod"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("PathNames() = %q, want %q", got, []string{"dib", "deploy", "apply"})
	}
	if got := result.MatchTokens(); !reflect.DeepEqual(got, []string{"dib", "ship", "push"}) {
		t.Fatalf("MatchTokens() = %q, want %q", got, []string{"dib", "ship", "push"})
	}
	if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"manifest.yaml", "--env", "prod"}) {
		t.Fatalf("RemainingArgs() = %q, want %q", got, []string{"manifest.yaml", "--env", "prod"})
	}

	canonical, err := root.Route([]string{"deploy", "apply", "manifest.yaml", "--env", "prod"})
	if err != nil {
		t.Fatalf("canonical Route returned unexpected error: %v", err)
	}
	if !reflect.DeepEqual(canonical.PathNames(), result.PathNames()) {
		t.Fatalf("canonical PathNames() = %q, want %q", canonical.PathNames(), result.PathNames())
	}
	if reflect.DeepEqual(canonical.MatchTokens(), result.MatchTokens()) {
		t.Fatalf("canonical and alias MatchTokens() both = %q, want raw tokens to differ", canonical.MatchTokens())
	}

	nested, err := root.Route([]string{"ship", "preview", "extra"})
	if err != nil {
		t.Fatalf("Route through sibling alias returned unexpected error: %v", err)
	}
	if got := nested.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "plan"}) {
		t.Fatalf("PathNames() = %q, want %q", got, []string{"dib", "deploy", "plan"})
	}
	if got := nested.MatchTokens(); !reflect.DeepEqual(got, []string{"dib", "ship", "preview"}) {
		t.Fatalf("MatchTokens() = %q, want %q", got, []string{"dib", "ship", "preview"})
	}
}

func TestAliasRoutingPublicAPIWorkflowTypedFailures(t *testing.T) {
	apply := mustDefinition(t, "apply", command.Aliases("push"))
	deploy := mustDefinition(t, "deploy", command.Aliases("ship"), command.Children(apply))
	root := mustDefinition(t, "dib", command.Children(deploy))

	result, err := root.Route([]string{"ship", "pus"})
	if err == nil {
		t.Fatal("Route returned nil error")
	}
	if !errors.Is(err, command.ErrUnknownCommand) {
		t.Fatalf("error does not satisfy ErrUnknownCommand: %v", err)
	}
	var unknown *command.UnknownCommandError
	if !errors.As(err, &unknown) {
		t.Fatalf("error does not expose UnknownCommandError: %T", err)
	}
	if got := unknown.Token(); got != "pus" {
		t.Fatalf("Token() = %q, want %q", got, "pus")
	}
	if got := unknown.ParentPath(); !reflect.DeepEqual(got, []string{"dib", "deploy"}) {
		t.Fatalf("ParentPath() = %q, want %q", got, []string{"dib", "deploy"})
	}
	if got := result.PathNames(); len(got) != 0 {
		t.Fatalf("failed Route returned non-zero path: %q", got)
	}
	if got := result.MatchTokens(); len(got) != 0 {
		t.Fatalf("failed Route returned non-zero match tokens: %q", got)
	}
	if got := result.RemainingArgs(); len(got) != 0 {
		t.Fatalf("failed Route returned non-zero remaining args: %q", got)
	}
}

func TestAliasSetupValidationThroughConstructorOptions(t *testing.T) {
	deploy := mustDefinition(t, "deploy", command.Aliases("run"))
	status := mustDefinition(t, "status", command.Aliases("run"))

	_, err := command.NewDefinition("dib", command.Children(deploy, status))
	if err == nil {
		t.Fatal("NewDefinition returned nil error")
	}
	if !errors.Is(err, command.ErrDuplicateCommandToken) {
		t.Fatalf("error does not satisfy ErrDuplicateCommandToken: %v", err)
	}
	var conflict *command.TokenConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("error does not expose TokenConflictError: %T", err)
	}
	if got := conflict.ParentPath(); !reflect.DeepEqual(got, []string{"dib"}) {
		t.Fatalf("ParentPath() = %q, want %q", got, []string{"dib"})
	}
	if got := conflict.Token(); got != "run" {
		t.Fatalf("Token() = %q, want %q", got, "run")
	}
	if got := conflict.FirstCommand(); got != "deploy" {
		t.Fatalf("FirstCommand() = %q, want %q", got, "deploy")
	}
	if got := conflict.CollidingCommand(); got != "status" {
		t.Fatalf("CollidingCommand() = %q, want %q", got, "status")
	}
}
