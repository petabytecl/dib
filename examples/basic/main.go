// Package main demonstrates high-level Dib handler dispatch.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/petabytecl/dib/cli"
	"github.com/petabytecl/dib/config"
	"github.com/petabytecl/dib/flags"
)

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}

func run(out, errOut io.Writer, argv []string) int {
	app := newGreeter(out)
	if _, err := app.Run(context.Background(), argv); err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}

func newGreeter(out io.Writer) *cli.Command {
	if out == nil {
		out = io.Discard
	}

	root := cli.New("greeter",
		cli.Description("small greeting CLI"),
		cli.Config(
			config.String("name", "world", "person to greet"),
			config.Bool("shout", false, "uppercase the greeting"),
		),
	)

	root.Command("hello",
		cli.Description("print a greeting"),
		cli.Flags(
			flags.String("name", "world", "person to greet", flags.Shorthand("n")),
			flags.Bool("shout", false, "uppercase the greeting"),
		),
		cli.Bindings(
			cli.BindFlag("name", "name"),
			cli.BindFlag("shout", "shout"),
		),
		cli.Handle(func(cmd cli.CommandContext) error {
			name, err := cmd.Config().GetString("name")
			if err != nil {
				return err
			}
			shout, err := cmd.Config().GetBool("shout")
			if err != nil {
				return err
			}

			message := "hello, " + name
			if shout {
				message = strings.ToUpper(message)
			}
			_, err = fmt.Fprintln(out, message)
			return err
		}),
	)

	return root
}
