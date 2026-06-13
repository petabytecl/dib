// Package main demonstrates typed Dib error handling.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

const usageErrorExitCode = 2

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}

func run(out, errOut io.Writer, argv []string) int {
	app := workerCLI(out)
	if _, err := app.Run(context.Background(), argv); err != nil {
		_ = writeDiagnostic(errOut, err)
		return usageErrorExitCode
	}
	return 0
}

func workerCLI(out io.Writer) *cli.Command {
	if out == nil {
		out = io.Discard
	}

	root := cli.New("workerctl",
		cli.Config(config.Int("workers", 1, "worker count")),
	)
	root.Command("run",
		cli.Description("run workers"),
		cli.Flags(flags.Int("workers", 1, "worker count")),
		cli.Bindings(cli.BindFlag("workers", "workers")),
		cli.Handle(func(cmd cli.CommandContext) error {
			workers, err := cmd.Config().GetInt("workers")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(out, "workers=%d\n", workers)
			return err
		}),
	)
	return root
}

func writeDiagnostic(w io.Writer, err error) error {
	if _, writeErr := fmt.Fprintf(w, "error=%v\n", err); writeErr != nil {
		return writeErr
	}

	var parseErr *flags.ParseError
	if errors.As(err, &parseErr) {
		_, writeErr := fmt.Fprintf(w, "parse_error token=%q name=%q category=%q\n",
			parseErr.Token(),
			parseErr.Name(),
			errorString(parseErr.Category()),
		)
		return writeErr
	}

	var unknownErr *command.UnknownCommandError
	if errors.As(err, &unknownErr) {
		_, writeErr := fmt.Fprintf(w, "unknown_command token=%q parent=%q\n",
			unknownErr.Token(),
			strings.Join(unknownErr.ParentPath(), " "),
		)
		return writeErr
	}

	var dispatchErr *cli.DispatchError
	if errors.As(err, &dispatchErr) {
		_, writeErr := fmt.Fprintf(w, "dispatch_error path=%q category=%q\n",
			strings.Join(dispatchErr.Path(), " "),
			errorString(dispatchErr.Category()),
		)
		return writeErr
	}
	return nil
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
