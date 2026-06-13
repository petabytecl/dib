package cli_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

func TestCommandRunDispatchesDistributedSubcommand(t *testing.T) {
	root := cli.New("svcctl",
		cli.Description("service control"),
		cli.Config(config.String("target", "scrapd", "service target")),
	)
	registerServiceCommands(root, func(cmd cli.CommandContext) error {
		target, err := cmd.Config().GetString("target")
		if err != nil {
			return err
		}
		if target != "api" {
			t.Fatalf("target = %q, want api", target)
		}
		if cmd.Context() == nil {
			t.Fatal("command context did not expose caller context")
		}
		if got := cmd.Route().PathNames(); len(got) != 3 || got[0] != "svcctl" || got[1] != "service" || got[2] != "start" {
			t.Fatalf("route = %v, want [svcctl service start]", got)
		}
		return nil
	})

	result, err := root.Run(context.Background(), []string{"svcctl", "service", "start", "--target", "api"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := result.Route().PathNames(); len(got) != 3 || got[2] != "start" {
		t.Fatalf("result route = %v, want final start", got)
	}
}

func TestCommandRunArgsUsesRootNameForInvocation(t *testing.T) {
	root := cli.New("svcctl")
	root.Command("status", cli.Handle(func(cmd cli.CommandContext) error {
		if cmd.Invocation().Program() != "svcctl" {
			t.Fatalf("program = %q, want svcctl", cmd.Invocation().Program())
		}
		return nil
	}))

	if _, err := root.RunArgs(context.Background(), []string{"status"}); err != nil {
		t.Fatalf("run args: %v", err)
	}
}

func TestCommandRunReturnsResultWithHandlerError(t *testing.T) {
	handlerErr := errors.New("handler failed")
	root := cli.New("svcctl")
	root.Command("start", cli.Handle(func(cli.CommandContext) error {
		return handlerErr
	}))

	result, err := root.Run(context.Background(), []string{"svcctl", "start"})
	if !errors.Is(err, handlerErr) {
		t.Fatalf("run err = %v, want handler error", err)
	}
	if got := result.Route().PathNames(); len(got) != 2 || got[1] != "start" {
		t.Fatalf("result route = %v, want [svcctl start]", got)
	}
}

func TestCommandRunReturnsNoHandlerErrorForMatchedCommandWithoutHandler(t *testing.T) {
	root := cli.New("svcctl")
	root.Command("start")

	result, err := root.Run(context.Background(), []string{"svcctl", "start"})
	if !errors.Is(err, cli.ErrNoHandler) {
		t.Fatalf("run err = %v, want ErrNoHandler", err)
	}
	var dispatchErr *cli.DispatchError
	if !errors.As(err, &dispatchErr) {
		t.Fatalf("run err %T, want *cli.DispatchError", err)
	}
	if got := dispatchErr.Path(); len(got) != 2 || got[1] != "start" {
		t.Fatalf("dispatch path = %v, want [svcctl start]", got)
	}
	if got := result.Route().PathNames(); len(got) != 2 || got[1] != "start" {
		t.Fatalf("result route = %v, want [svcctl start]", got)
	}
}

func TestCommandRunDoesNotInvokeHandlerWhenResolveFails(t *testing.T) {
	root := cli.New("svcctl")
	root.Command("start", cli.Handle(func(cli.CommandContext) error {
		t.Fatal("handler was called after resolve failure")
		return nil
	}))

	if _, err := root.Run(context.Background(), []string{"svcctl", "unknown"}); err == nil {
		t.Fatal("run unknown command: expected error")
	}
}

func TestCommandPlanBuildsLowLevelPlan(t *testing.T) {
	root := cli.New("svcctl",
		cli.Config(config.String("target", "scrapd", "service target")),
	)
	root.Command("start",
		cli.Flags(flags.String("target", "scrapd", "service target")),
		cli.Bindings(cli.BindFlag("target", "target")),
		cli.Handle(func(cli.CommandContext) error {
			return nil
		}),
	)

	plan, err := root.Plan()
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if plan.Root().Name() != "svcctl" {
		t.Fatalf("plan root = %q, want svcctl", plan.Root().Name())
	}
	if _, ok := plan.ConfigSet().Lookup("target"); !ok {
		t.Fatal("plan config set missing target")
	}
	if got := plan.Bindings(); len(got) != 1 || got[0].FlagName != "target" || got[0].ConfigKey != "target" {
		t.Fatalf("plan bindings = %#v, want target->target", got)
	}
}

func TestCommandDefinitionIncludesBuilderMetadata(t *testing.T) {
	root := cli.New("svcctl",
		cli.Aliases("svc"),
		cli.Usage("svcctl <command>"),
		cli.InheritedFlags(flags.Bool("verbose", false, "enable verbose output")),
		cli.FlagNormalizer(strings.ToLower),
	)
	root.Command("start", cli.Description("start service"))

	definition, err := root.Definition()
	if err != nil {
		t.Fatalf("definition: %v", err)
	}
	if got := definition.Aliases(); len(got) != 1 || got[0] != "svc" {
		t.Fatalf("aliases = %v, want [svc]", got)
	}
	if definition.Usage() != "svcctl <command>" {
		t.Fatalf("usage = %q, want svcctl <command>", definition.Usage())
	}
	if got := definition.InheritedFlags(); len(got) != 1 || got[0].Name() != "verbose" {
		t.Fatalf("inherited flags = %v, want verbose", got)
	}
	if got := definition.Children(); len(got) != 1 || got[0].Name() != "start" || got[0].Description() != "start service" {
		t.Fatalf("children = %v, want start child with description", got)
	}
}

func TestCommandContextExposesResolvedState(t *testing.T) {
	handlerErr := errors.New("handler inspected context")
	root := cli.New("svcctl",
		cli.Config(config.String("target", "scrapd", "service target")),
	)
	root.Command("start",
		cli.Flags(flags.String("target", "scrapd", "service target")),
		cli.Bindings(cli.BindFlag("target", "target")),
		cli.Handle(func(cmd cli.CommandContext) error {
			if cmd.Result().Invocation().Program() != "svcctl" {
				t.Fatalf("result invocation program = %q, want svcctl", cmd.Result().Invocation().Program())
			}
			if cmd.Invocation().Program() != "svcctl" {
				t.Fatalf("invocation program = %q, want svcctl", cmd.Invocation().Program())
			}
			if got := cmd.RemainingArgs(); len(got) != 1 || got[0] != "tail" {
				t.Fatalf("remaining args = %v, want [tail]", got)
			}
			if _, ok := cmd.FlagSnapshot(); !ok {
				t.Fatal("expected flag snapshot")
			}
			target, err := cmd.Config().GetString("target")
			if err != nil {
				t.Fatalf("get target: %v", err)
			}
			if target != "api" {
				t.Fatalf("target = %q, want api", target)
			}
			return handlerErr
		}),
	)

	_, err := root.Run(context.Background(), []string{"svcctl", "start", "--target", "api", "tail"})
	if !errors.Is(err, handlerErr) {
		t.Fatalf("run err = %v, want handler error", err)
	}
}

func TestDispatchErrorAccessorsAndNilReceiver(t *testing.T) {
	var nilDispatch *cli.DispatchError
	if nilDispatch.Error() != "cli: dispatch error" {
		t.Fatalf("nil dispatch error string = %q", nilDispatch.Error())
	}
	if nilDispatch.Unwrap() != nil {
		t.Fatal("nil dispatch unwrap returned non-nil")
	}
	if nilDispatch.Category() != nil {
		t.Fatal("nil dispatch category returned non-nil")
	}
	if nilDispatch.Path() != nil {
		t.Fatal("nil dispatch path returned non-nil")
	}

	root := cli.New("svcctl")
	root.Command("start")

	_, err := root.Run(context.Background(), []string{"svcctl", "start"})
	var dispatchErr *cli.DispatchError
	if !errors.As(err, &dispatchErr) {
		t.Fatalf("run err %T, want *cli.DispatchError", err)
	}
	if dispatchErr.Category() != cli.ErrNoHandler {
		t.Fatalf("category = %v, want ErrNoHandler", dispatchErr.Category())
	}
	if !strings.Contains(dispatchErr.Error(), "svcctl start") {
		t.Fatalf("dispatch error string = %q, want route", dispatchErr.Error())
	}
}

func TestNilCommandBuilderFailsAsInvalidCommandDefinition(t *testing.T) {
	var root *cli.Command
	if root.Name() != "" {
		t.Fatalf("nil command name = %q, want empty", root.Name())
	}
	if child := root.Command("start"); child.Name() != "start" {
		t.Fatalf("nil parent child name = %q, want start", child.Name())
	}

	_, err := root.Plan()
	var nameErr *command.NameError
	if !errors.As(err, &nameErr) {
		t.Fatalf("plan err = %v, want command name error", err)
	}
}

func registerServiceCommands(root *cli.Command, start cli.Handler) {
	service := root.Command("service", cli.Description("service commands"))
	service.Command("start",
		cli.Flags(flags.String("target", "scrapd", "service target")),
		cli.Bindings(cli.BindFlag("target", "target")),
		cli.Handle(start),
	)
	service.Command("stop",
		cli.Flags(flags.String("target", "scrapd", "service target")),
		cli.Bindings(cli.BindFlag("target", "target")),
		cli.Handle(func(cli.CommandContext) error {
			return nil
		}),
	)
}
