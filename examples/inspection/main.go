// Package main demonstrates inspect-only Dib resolution.
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

const defaultAuditLimit = 10

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}

func run(out, errOut io.Writer, argv []string) int {
	plan, err := auditPlan()
	if err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}

	inv, err := cli.FromOSArgs(argv)
	if err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}
	result, err := cli.Resolve(inv, plan)
	if err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}

	tenant, err := result.Config().GetString("tenant")
	if err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}
	limit, err := result.Config().GetInt("limit")
	if err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}

	if writeErr := writeInspectionResult(out, result, tenant, limit); writeErr != nil {
		_, _ = fmt.Fprintln(errOut, writeErr)
		return 1
	}
	return 0
}

func writeInspectionResult(out io.Writer, result cli.Result, tenant string, limit int) error {
	lines := []string{
		"path=" + strings.Join(result.Route().PathNames(), " "),
		"remaining=" + strings.Join(result.RemainingArgs(), ","),
		"tenant=" + tenant + " limit=" + strconv.Itoa(limit),
	}
	return writeLines(out, lines)
}

func writeLines(out io.Writer, lines []string) error {
	for _, line := range lines {
		//nolint:gosec // CLI examples write plain text to caller-owned stdout, not HTML.
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

func auditPlan() (cli.Plan, error) {
	list, err := command.NewDefinition("list",
		command.Description("list audit events"),
		command.LocalFlags(flags.Int("limit", defaultAuditLimit, "maximum events")),
	)
	if err != nil {
		return cli.Plan{}, err
	}
	events, err := command.NewDefinition("events",
		command.Description("event commands"),
		command.Children(list),
	)
	if err != nil {
		return cli.Plan{}, err
	}
	root, err := command.NewDefinition("auditctl",
		command.Description("audit control"),
		command.InheritedFlags(flags.String("tenant", "default", "tenant id")),
		command.Children(events),
	)
	if err != nil {
		return cli.Plan{}, err
	}
	set, err := config.NewSet(
		config.String("tenant", "default", "tenant id"),
		config.Int("limit", defaultAuditLimit, "maximum events"),
	)
	if err != nil {
		return cli.Plan{}, err
	}
	return cli.NewPlan(root, set).
		WithBindings([]cli.FlagBinding{
			cli.BindFlag("tenant", "tenant"),
			cli.BindFlag("limit", "limit"),
		}), nil
}
