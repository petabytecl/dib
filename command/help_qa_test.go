package command_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/flags"
)

func TestWriteHelpRendersDefinitionLocalFlagsWithoutRoute(t *testing.T) {
	definition := mustDefinition(
		t,
		"serve",
		command.Usage("[flags]"),
		command.LocalFlags(
			flags.Bool("watch", false, "Watch files.", flags.Shorthand("w")),
			flags.String("token", "dib_fake_token_value", "Access token.", flags.Sensitive()),
			flags.String("hidden-token", "dib_fake_secret_value", "Hidden token.", flags.Hidden(), flags.Sensitive()),
			flags.Bool("legacy", false, "Legacy mode.", flags.Deprecated("use --watch")),
		),
	)

	var out bytes.Buffer
	if err := definition.WriteHelp(&out); err != nil {
		t.Fatalf("WriteHelp returned unexpected error: %v", err)
	}

	const want = `Usage:
  serve [flags]

Flags:
  --watch, -w  Watch files.
  --token <string>  Access token.
  --legacy  Legacy mode. (deprecated: use --watch)
`
	if got := out.String(); got != want {
		t.Fatalf("WriteHelp output:\n%s\nwant:\n%s", got, want)
	}

	for _, forbidden := range []string{
		"hidden-token",
		"dib_fake_secret_value",
		"dib_fake_token_value",
	} {
		if strings.Contains(out.String(), forbidden) {
			t.Fatalf("WriteHelp output leaked %q:\n%s", forbidden, out.String())
		}
	}
}

func TestWriteUsagePropagatesWriterFailuresAndRejectsInvalidTargets(t *testing.T) {
	writerErr := errors.New("deterministic usage writer failure")
	root := mustHelpTree(t)

	if err := root.WriteUsage(failingWriter{err: writerErr}); !errors.Is(err, writerErr) {
		t.Fatalf("Definition.WriteUsage error = %v, want writer error", err)
	}

	result, err := root.Route([]string{"deploy", "apply"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}
	if err := result.WriteUsage(failingWriter{err: writerErr}); !errors.Is(err, writerErr) {
		t.Fatalf("Result.WriteUsage error = %v, want writer error", err)
	}

	var zeroDefinition command.Definition
	if err := zeroDefinition.WriteUsage(&bytes.Buffer{}); !isNameError(err) {
		t.Fatalf("zero Definition.WriteUsage error = %T, want *command.NameError", err)
	}

	var zeroResult command.Result
	if err := zeroResult.WriteUsage(&bytes.Buffer{}); !isNameError(err) {
		t.Fatalf("zero Result.WriteUsage error = %T, want *command.NameError", err)
	}
	if err := zeroResult.WriteHelp(&bytes.Buffer{}); !isNameError(err) {
		t.Fatalf("zero Result.WriteHelp error = %T, want *command.NameError", err)
	}
}

func isNameError(err error) bool {
	var nameErr *command.NameError
	return errors.As(err, &nameErr)
}
