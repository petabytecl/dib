package command_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/petabytecl/dib/command"
)

func TestRouteRootAndNestedCommands(t *testing.T) {
	root := mustRoutingTree(t)

	tests := []struct {
		name      string
		args      []string
		wantPath  []string
		wantRem   []string
		wantChild string
	}{
		{
			name:     "empty args match root",
			args:     nil,
			wantPath: []string{"dib"},
		},
		{
			name:      "nested deploy apply",
			args:      []string{"deploy", "apply"},
			wantPath:  []string{"dib", "deploy", "apply"},
			wantChild: "apply",
		},
		{
			name:      "leaf preserves positionals",
			args:      []string{"deploy", "apply", "manifest.yaml", "prod"},
			wantPath:  []string{"dib", "deploy", "apply"},
			wantRem:   []string{"manifest.yaml", "prod"},
			wantChild: "apply",
		},
		{
			name:      "leaf preserves flag-like remaining args",
			args:      []string{"deploy", "apply", "--dry-run", "-v"},
			wantPath:  []string{"dib", "deploy", "apply"},
			wantRem:   []string{"--dry-run", "-v"},
			wantChild: "apply",
		},
		{
			name:      "terminator stops child matching",
			args:      []string{"deploy", "--", "apply", "--dry-run"},
			wantPath:  []string{"dib", "deploy"},
			wantRem:   []string{"apply", "--dry-run"},
			wantChild: "deploy",
		},
		{
			name:      "terminator after leaf is omitted",
			args:      []string{"deploy", "apply", "--", "--dry-run", "prod"},
			wantPath:  []string{"dib", "deploy", "apply"},
			wantRem:   []string{"--dry-run", "prod"},
			wantChild: "apply",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := root.Route(tt.args)
			if err != nil {
				t.Fatalf("Route returned unexpected error: %v", err)
			}
			if got := result.PathNames(); !reflect.DeepEqual(got, tt.wantPath) {
				t.Fatalf("PathNames() = %q, want %q", got, tt.wantPath)
			}
			if got := result.RemainingArgs(); !reflect.DeepEqual(got, tt.wantRem) {
				t.Fatalf("RemainingArgs() = %q, want %q", got, tt.wantRem)
			}
			if tt.wantChild != "" {
				got, ok := result.Command()
				if !ok {
					t.Fatal("Command() returned ok=false")
				}
				if got.Name() != tt.wantChild {
					t.Fatalf("Command().Name() = %q, want %q", got.Name(), tt.wantChild)
				}
			}
		})
	}
}

func TestRouteRootWithoutChildrenPreservesRemainingArgs(t *testing.T) {
	root, err := command.NewDefinition("dib")
	if err != nil {
		t.Fatalf("NewDefinition returned unexpected error: %v", err)
	}

	result, err := root.Route([]string{"status", "--verbose"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib"}) {
		t.Fatalf("PathNames() = %q, want %q", got, []string{"dib"})
	}
	if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"status", "--verbose"}) {
		t.Fatalf("RemainingArgs() = %q, want %q", got, []string{"status", "--verbose"})
	}
}

func TestRouteUnknownCommandErrorsAreInspectable(t *testing.T) {
	root := mustRoutingTree(t)

	tests := []struct {
		name       string
		args       []string
		wantToken  string
		wantParent []string
	}{
		{
			name:       "unknown at root",
			args:       []string{"destroy"},
			wantToken:  "destroy",
			wantParent: []string{"dib"},
		},
		{
			name:       "unknown under matched parent",
			args:       []string{"deploy", "destroy"},
			wantToken:  "destroy",
			wantParent: []string{"dib", "deploy"},
		},
		{
			name:       "flag-like token before leaf is still an unknown command",
			args:       []string{"deploy", "--dry-run"},
			wantToken:  "--dry-run",
			wantParent: []string{"dib", "deploy"},
		},
		{
			name:       "alias metadata is not routed in story 3.1",
			args:       []string{"ship"},
			wantToken:  "ship",
			wantParent: []string{"dib"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := root.Route(tt.args)
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
			if got := unknown.Token(); got != tt.wantToken {
				t.Fatalf("Token() = %q, want %q", got, tt.wantToken)
			}
			if got := unknown.ParentPath(); !reflect.DeepEqual(got, tt.wantParent) {
				t.Fatalf("ParentPath() = %q, want %q", got, tt.wantParent)
			}
			if got := result.PathNames(); len(got) != 0 {
				t.Fatalf("failed Route returned non-zero path: %q", got)
			}
			if got := result.RemainingArgs(); len(got) != 0 {
				t.Fatalf("failed Route returned non-zero remaining args: %q", got)
			}
		})
	}
}

func TestRouteRejectsInvalidRootDefinition(t *testing.T) {
	var root command.Definition

	result, err := root.Route(nil)
	if err == nil {
		t.Fatal("Route returned nil error")
	}
	var nameErr *command.NameError
	if !errors.As(err, &nameErr) {
		t.Fatalf("error does not expose NameError: %T", err)
	}
	if got := result.PathNames(); len(got) != 0 {
		t.Fatalf("failed Route returned non-zero path: %q", got)
	}
	if got := result.RemainingArgs(); len(got) != 0 {
		t.Fatalf("failed Route returned non-zero remaining args: %q", got)
	}
}

func TestRouteSnapshotsAreDefensiveAndDeterministic(t *testing.T) {
	root := mustRoutingTree(t)
	args := []string{"deploy", "apply", "manifest.yaml"}

	result, err := root.Route(args)
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	args[0] = "destroy"

	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("PathNames() changed after caller args mutation: %q", got)
	}
	if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"manifest.yaml"}) {
		t.Fatalf("RemainingArgs() changed after caller args mutation: %q", got)
	}

	names := result.PathNames()
	names[0] = "mutated"
	if got := result.PathNames(); !reflect.DeepEqual(got, []string{"dib", "deploy", "apply"}) {
		t.Fatalf("PathNames() leaked mutable slice: %q", got)
	}

	remaining := result.RemainingArgs()
	remaining[0] = "mutated"
	if got := result.RemainingArgs(); !reflect.DeepEqual(got, []string{"manifest.yaml"}) {
		t.Fatalf("RemainingArgs() leaked mutable slice: %q", got)
	}

	const runs = 32
	var wg sync.WaitGroup
	errs := make(chan string, runs)
	for i := 0; i < runs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := root.Route([]string{"deploy", "apply", "manifest.yaml"})
			if err != nil {
				errs <- err.Error()
				return
			}
			if !reflect.DeepEqual(got.PathNames(), []string{"dib", "deploy", "apply"}) {
				errs <- "unexpected path"
				return
			}
			if !reflect.DeepEqual(got.RemainingArgs(), []string{"manifest.yaml"}) {
				errs <- "unexpected remaining args"
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func mustRoutingTree(t *testing.T) command.Definition {
	t.Helper()

	apply, err := command.NewDefinition("apply", command.Description("apply a deployment"))
	if err != nil {
		t.Fatalf("NewDefinition(apply) returned unexpected error: %v", err)
	}
	plan, err := command.NewDefinition("plan")
	if err != nil {
		t.Fatalf("NewDefinition(plan) returned unexpected error: %v", err)
	}
	deploy, err := command.NewDefinition(
		"deploy",
		command.Description("manage deployments"),
		command.Aliases("ship"),
		command.Usage("deploy <command>"),
		command.Children(apply, plan),
	)
	if err != nil {
		t.Fatalf("NewDefinition(deploy) returned unexpected error: %v", err)
	}
	status, err := command.NewDefinition("status")
	if err != nil {
		t.Fatalf("NewDefinition(status) returned unexpected error: %v", err)
	}
	root, err := command.NewDefinition("dib", command.Children(deploy, status))
	if err != nil {
		t.Fatalf("NewDefinition(dib) returned unexpected error: %v", err)
	}
	return root
}
