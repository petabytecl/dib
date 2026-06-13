// Package main demonstrates low-level Dib config precedence.
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

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args, os.LookupEnv))
}

func run(out, errOut io.Writer, argv []string, lookup config.EnvLookup) int {
	if lookup == nil {
		lookup = func(string) (string, bool) { return "", false }
	}

	plan, err := deployPlan(lookup)
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

	region, err := result.Config().GetString("region")
	if err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}
	workers, err := result.Config().GetInt("workers")
	if err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}
	format, err := result.Config().GetString("format")
	if err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}

	if writeErr := writeConfigResult(out, result, region, workers, format); writeErr != nil {
		_, _ = fmt.Fprintln(errOut, writeErr)
		return 1
	}
	return 0
}

func writeConfigResult(out io.Writer, result cli.Result, region string, workers int, format string) error {
	lines := []string{
		"route=" + strings.Join(result.Route().PathNames(), " "),
		"region=" + region + " workers=" + strconv.Itoa(workers) + " format=" + format,
		"sources region=" + sourceLabel(result.Config(), "region") +
			" workers=" + sourceLabel(result.Config(), "workers") +
			" format=" + sourceLabel(result.Config(), "format"),
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

func deployPlan(lookup config.EnvLookup) (cli.Plan, error) {
	deploy, err := command.NewDefinition("deploy",
		command.Description("deploy an application"),
		command.LocalFlags(
			flags.String("region", "", "deployment region"),
			flags.Int("workers", 1, "worker count"),
		),
	)
	if err != nil {
		return cli.Plan{}, err
	}
	root, err := command.NewDefinition("deployctl",
		command.Description("deployment control"),
		command.Children(deploy),
	)
	if err != nil {
		return cli.Plan{}, err
	}

	set, err := config.NewSet(
		config.String("region", "us-east", "deployment region"),
		config.Int("workers", 1, "worker count"),
		config.String("format", "text", "output format"),
	)
	if err != nil {
		return cli.Plan{}, err
	}

	env, err := config.NewEnvSnapshot(set, lookup, []config.EnvBinding{
		config.BindEnv("region", "DIB_REGION"),
		config.BindEnv("workers", "DIB_WORKERS"),
	})
	if err != nil {
		return cli.Plan{}, err
	}
	jsonSource, err := config.LoadJSON(set, strings.NewReader(`{
		"region": "json-region",
		"workers": 2,
		"format": "json"
	}`), config.JSONReaderLabel("embedded example config"))
	if err != nil {
		return cli.Plan{}, err
	}

	return cli.NewPlan(root, set).
		WithEnv(env).
		WithJSON(jsonSource).
		WithBindings([]cli.FlagBinding{
			cli.BindFlag("region", "region"),
			cli.BindFlag("workers", "workers"),
		}), nil
}

func sourceLabel(snapshot config.Snapshot, key string) string {
	for _, entry := range snapshot.SourceReport() {
		if entry.Key() == key {
			return entry.SourceLabel()
		}
	}
	return ""
}
