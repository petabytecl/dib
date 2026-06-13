package multicommand_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

// Example_composedCLI demonstrates routing, flag parsing, and config resolution
// through cli.Resolve with caller-supplied inputs.
func Example_composedCLI() {
	// Define the serve sub-command with a --host flag.
	serve, _ := command.NewDefinition("serve",
		command.Description("start the server"),
		command.LocalFlags(flags.String("host", "", "server hostname")),
	)
	// Define the root command with serve as a child.
	root, _ := command.NewDefinition("app",
		command.Description("my application"),
		command.Children(serve),
	)

	// Define config keys.
	set, _ := config.NewSet(config.String("host", "localhost", "server hostname"))

	// Build a plan that binds the --host flag to the host config key.
	plan := cli.NewPlan(root, set).
		WithBindings([]cli.FlagBinding{cli.BindFlag("host", "host")})

	// Build an invocation from caller-supplied argv (not os.Args directly).
	inv, _ := cli.FromOSArgs([]string{"app", "serve", "--host", "example.com"})

	// Resolve: routes the command, parses flags, resolves config by precedence.
	result, err := cli.Resolve(inv, plan)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	host, _ := result.Config().GetString("host")
	fmt.Println(result.Route().PathNames())
	fmt.Println(host)
	// Output:
	// [app serve]
	// example.com
}

// Example_dispatchStartStop demonstrates distributed command registration and
// handler dispatch through the high-level cli.Command builder.
func Example_dispatchStartStop() {
	root := serviceCommands(
		func(cmd cli.CommandContext) error {
			target, err := cmd.Config().GetString("target")
			if err != nil {
				return err
			}
			fmt.Println("start", target)
			return nil
		},
		func(cmd cli.CommandContext) error {
			target, err := cmd.Config().GetString("target")
			if err != nil {
				return err
			}
			fmt.Println("stop", target)
			return nil
		},
	)

	for _, argv := range [][]string{
		{"svcctl", "service", "start", "--target", "scrapd"},
		{"svcctl", "service", "stop", "--target", "scrapd"},
	} {
		if _, err := root.Run(context.Background(), argv); err != nil {
			fmt.Println("error:", err)
			return
		}
	}

	// Output:
	// start scrapd
	// stop scrapd
}

func TestDispatchStartStopCallsApplicationHandlers(t *testing.T) {
	startErr := errors.New("start handler called")
	stopErr := errors.New("stop handler called")
	root := serviceCommands(
		func(cli.CommandContext) error {
			return startErr
		},
		func(cli.CommandContext) error {
			return stopErr
		},
	)

	result, err := root.Run(context.Background(), []string{"svcctl", "service", "start", "--target", "scrapd"})
	if !errors.Is(err, startErr) {
		t.Fatalf("start err = %v, want start handler error", err)
	}
	if got := result.Route().PathNames(); len(got) != 3 || got[2] != "start" {
		t.Fatalf("start route = %v, want final start", got)
	}

	result, err = root.Run(context.Background(), []string{"svcctl", "service", "stop", "--target", "scrapd"})
	if !errors.Is(err, stopErr) {
		t.Fatalf("stop err = %v, want stop handler error", err)
	}
	if got := result.Route().PathNames(); len(got) != 3 || got[2] != "stop" {
		t.Fatalf("stop route = %v, want final stop", got)
	}
}

func serviceCommands(start, stop cli.Handler) *cli.Command {
	root := cli.New("svcctl",
		cli.Description("service control"),
		cli.Config(config.String("target", "scrapd", "service target")),
	)
	registerServiceCommands(root, start, stop)
	return root
}

func registerServiceCommands(root *cli.Command, start, stop cli.Handler) {
	service := root.Command("service", cli.Description("service commands"))
	service.Command("start",
		cli.Flags(flags.String("target", "scrapd", "service target")),
		cli.Bindings(cli.BindFlag("target", "target")),
		cli.Handle(start),
	)
	service.Command("stop",
		cli.Flags(flags.String("target", "scrapd", "service target")),
		cli.Bindings(cli.BindFlag("target", "target")),
		cli.Handle(stop),
	)
}

func TestCommandBuilderPlanRemainsInspectable(t *testing.T) {
	root := serviceCommands(
		func(cli.CommandContext) error {
			return nil
		},
		func(cli.CommandContext) error {
			return nil
		},
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
}

func TestCommandBuilderRunArgsUsesRootName(t *testing.T) {
	wantErr := errors.New("status handler called")
	root := cli.New("svcctl")
	root.Command("status", cli.Handle(func(cmd cli.CommandContext) error {
		if cmd.Invocation().Program() != "svcctl" {
			t.Fatalf("program = %q, want svcctl", cmd.Invocation().Program())
		}
		return wantErr
	}))

	_, err := root.RunArgs(context.Background(), []string{"status"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("run args err = %v, want status handler error", err)
	}
}

func TestCommandBuilderMissingHandlerIsInspectable(t *testing.T) {
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
		t.Fatalf("route = %v, want [svcctl start]", got)
	}
}

func TestCommandBuilderDoesNotCallHandlerAfterResolveFailure(t *testing.T) {
	root := cli.New("svcctl")
	root.Command("start", cli.Handle(func(cli.CommandContext) error {
		t.Fatal("handler was called after resolve failure")
		return nil
	}))

	if _, err := root.Run(context.Background(), []string{"svcctl", "unknown"}); err == nil {
		t.Fatal("run unknown: expected error")
	}
}

func Example_lowLevelDispatch() {
	app := serviceApp{}

	startCode, startEvent, err := runServiceCLI(context.Background(), []string{"svcctl", "start", "--target", "scrapd"}, app)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	stopCode, stopEvent, err := runServiceCLI(context.Background(), []string{"svcctl", "stop", "--target", "scrapd"}, app)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(startCode)
	fmt.Println(stopCode)
	fmt.Println(startEvent + ", " + stopEvent)

	// Output:
	// 0
	// 0
	// start scrapd, stop scrapd
}

type serviceApp struct{}

func (a serviceApp) Start(_ context.Context, target string) (string, error) {
	return "start " + target, nil
}

func (a serviceApp) Stop(_ context.Context, target string) (string, error) {
	return "stop " + target, nil
}

func runServiceCLI(ctx context.Context, argv []string, app serviceApp) (int, string, error) {
	inv, err := cli.FromOSArgs(argv)
	if err != nil {
		return 2, "", err
	}

	plan, err := servicePlan()
	if err != nil {
		return 2, "", err
	}

	result, err := cli.Resolve(inv, plan)
	if err != nil {
		return 2, "", err
	}

	target, err := result.Config().GetString("target")
	if err != nil {
		return 2, "", err
	}

	route := result.Route().PathNames()
	if len(route) < 2 {
		return 2, "", errors.New("missing command")
	}

	switch route[1] {
	case "start":
		event, err := app.Start(ctx, target)
		if err != nil {
			return 1, "", err
		}
		return 0, event, nil
	case "stop":
		event, err := app.Stop(ctx, target)
		if err != nil {
			return 1, "", err
		}
		return 0, event, nil
	default:
		return 2, "", fmt.Errorf("unhandled command %q", route[1])
	}
}

func servicePlan() (cli.Plan, error) {
	start, err := command.NewDefinition("start",
		command.Description("start the service"),
		command.LocalFlags(flags.String("target", "scrapd", "service target")),
	)
	if err != nil {
		return cli.Plan{}, err
	}
	stop, err := command.NewDefinition("stop",
		command.Description("stop the service"),
		command.LocalFlags(flags.String("target", "scrapd", "service target")),
	)
	if err != nil {
		return cli.Plan{}, err
	}
	root, err := command.NewDefinition("svcctl",
		command.Description("service control"),
		command.Children(start, stop),
	)
	if err != nil {
		return cli.Plan{}, err
	}
	set, err := config.NewSet(config.String("target", "scrapd", "service target"))
	if err != nil {
		return cli.Plan{}, err
	}
	return cli.NewPlan(root, set).
		WithBindings([]cli.FlagBinding{cli.BindFlag("target", "target")}), nil
}

// TestComposedCLIResolvesBehavior asserts the composed resolution behavior
// without relying on stdout output scraping.
func TestComposedCLIResolvesBehavior(t *testing.T) {
	hostDef := flags.String("host", "", "server hostname")
	serve, err := command.NewDefinition("serve",
		command.Description("start the server"),
		command.LocalFlags(hostDef),
	)
	if err != nil {
		t.Fatalf("define serve: %v", err)
	}
	root, err := command.NewDefinition("app",
		command.Description("my application"),
		command.Children(serve),
	)
	if err != nil {
		t.Fatalf("define root: %v", err)
	}

	set, err := config.NewSet(config.String("host", "localhost", "server hostname"))
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	plan := cli.NewPlan(root, set).
		WithBindings([]cli.FlagBinding{cli.BindFlag("host", "host")})

	inv, err := cli.FromOSArgs([]string{"app", "serve", "--host", "example.com"})
	if err != nil {
		t.Fatalf("invocation: %v", err)
	}

	result, err := cli.Resolve(inv, plan)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// Verify routing result.
	names := result.Route().PathNames()
	if len(names) != 2 || names[0] != "app" || names[1] != "serve" {
		t.Fatalf("expected path [app serve], got %v", names)
	}

	// Verify the flag-bound config value takes precedence over the default.
	host, err := result.Config().GetString("host")
	if err != nil {
		t.Fatalf("get host: %v", err)
	}
	if host != "example.com" {
		t.Fatalf("expected host=example.com, got %q", host)
	}

	// Verify the invocation is preserved in the result.
	if result.Invocation().Program() != "app" {
		t.Fatalf("expected program=app, got %q", result.Invocation().Program())
	}
}
