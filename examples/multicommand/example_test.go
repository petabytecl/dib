package multicommand_test

import (
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
