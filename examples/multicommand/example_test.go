package multicommand_test

import (
	"context"
	"fmt"
	"strings"
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

// Example_dispatchStartStop demonstrates the application-owned dispatch layer
// that sits after cli.Resolve. Dib resolves argv into route, flags, and config;
// the caller decides which handler to run.
func Example_dispatchStartStop() {
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
	fmt.Println(strings.Join([]string{startEvent, stopEvent}, ", "))

	// Output:
	// 0
	// 0
	// start scrapd, stop scrapd
}

func TestDispatchStartStopCallsApplicationHandlers(t *testing.T) {
	app := serviceApp{}

	code, event, err := runServiceCLI(context.Background(), []string{"svcctl", "start", "--target", "scrapd"}, app)
	if err != nil {
		t.Fatalf("run start: %v", err)
	}
	if code != 0 {
		t.Fatalf("start exit code = %d, want 0", code)
	}
	if event != "start scrapd" {
		t.Fatalf("start event = %q, want start scrapd", event)
	}

	code, event, err = runServiceCLI(context.Background(), []string{"svcctl", "stop", "--target", "scrapd"}, app)
	if err != nil {
		t.Fatalf("run stop: %v", err)
	}
	if code != 0 {
		t.Fatalf("stop exit code = %d, want 0", code)
	}
	if event != "stop scrapd" {
		t.Fatalf("stop event = %q, want stop scrapd", event)
	}
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
		return 2, "", fmt.Errorf("missing command")
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
