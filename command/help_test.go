package command_test

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/petabytecl/dib/command"
	"github.com/petabytecl/dib/flags"
)

func TestWriteHelpRendersDefinitionMetadataDeterministically(t *testing.T) {
	root := mustHelpTree(t)

	var out bytes.Buffer
	if err := root.WriteHelp(&out); err != nil {
		t.Fatalf("WriteHelp returned unexpected error: %v", err)
	}

	const want = `Usage:
  dib <command> [flags]

Aliases:
  db, devbox

Description:
  Developer workspace commands.

Commands:
  deploy  Manage deployments.
  status  Show workspace status.

Flags:
  --verbose, -v  Enable verbose output.
  --profile <string>  Select profile.
`
	if got := out.String(); got != want {
		t.Fatalf("WriteHelp output:\n%s\nwant:\n%s", got, want)
	}

	if got := childNames(root.Children()); !reflect.DeepEqual(got, []string{"deploy", "status"}) {
		t.Fatalf("Children order = %q, want %q", got, []string{"deploy", "status"})
	}
	if got := root.Aliases(); !reflect.DeepEqual(got, []string{"db", "devbox"}) {
		t.Fatalf("Aliases() = %q, want %q", got, []string{"db", "devbox"})
	}
}

func TestWriteHelpRendersRoutedCommandFlagsAndUsage(t *testing.T) {
	root := mustHelpTree(t)

	result, err := root.Route([]string{"ship", "apply", "--cluster", "prod", "manifest.yaml"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}

	var help bytes.Buffer
	if err := result.WriteHelp(&help); err != nil {
		t.Fatalf("Result.WriteHelp returned unexpected error: %v", err)
	}

	const wantHelp = `Usage:
  dib deploy apply <manifest> [flags]

Aliases:
  push

Description:
  Apply a deployment manifest.

Flags:
  --verbose, -v  Enable verbose output.
  --profile <string>  Select profile.
  --cluster <string>  Target cluster.
  --retries <int>  Retry count.
  --tag <string-list>  Deployment tag.
  --token <string>  API token.
  --dry-run  Print planned operations. (deprecated: use --preview)
`
	if got := help.String(); got != wantHelp {
		t.Fatalf("Result.WriteHelp output:\n%s\nwant:\n%s", got, wantHelp)
	}

	var usage bytes.Buffer
	if err := result.WriteUsage(&usage); err != nil {
		t.Fatalf("Result.WriteUsage returned unexpected error: %v", err)
	}
	if got, want := usage.String(), "Usage:\n  dib deploy apply <manifest> [flags]\n"; got != want {
		t.Fatalf("Result.WriteUsage output = %q, want %q", got, want)
	}

	set, ok := result.Flags()
	if !ok {
		t.Fatal("Flags() returned ok=false")
	}
	if got := flagNames(set.Definitions()); !reflect.DeepEqual(got, []string{"verbose", "profile", "cluster", "retries", "tag", "hidden-token", "token", "dry-run"}) {
		t.Fatalf("Flags().Definitions() names = %q, want routed hidden/deprecated definitions preserved", got)
	}
	if _, ok := set.Lookup("hidden-token"); !ok {
		t.Fatal("hidden-token was not parseable through routed flag set")
	}
}

func TestWriteHelpOmitsHiddenAndSensitiveValues(t *testing.T) {
	root := mustHelpTree(t)

	result, err := root.Route([]string{"deploy", "apply", "--hidden-token", "dib_fake_secret_value", "--token", "dib_fake_token_value"})
	if err != nil {
		t.Fatalf("Route returned unexpected error: %v", err)
	}

	var out bytes.Buffer
	if err := result.WriteHelp(&out); err != nil {
		t.Fatalf("WriteHelp returned unexpected error: %v", err)
	}

	got := out.String()
	for _, forbidden := range []string{
		"hidden-token",
		"dib_fake_secret_value",
		"dib_fake_password_value",
		"dib_fake_token_value",
	} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("help output leaked %q:\n%s", forbidden, got)
		}
	}
	if !strings.Contains(got, "--token <string>  API token.") {
		t.Fatalf("help output omitted visible sensitive flag metadata:\n%s", got)
	}
}

func TestWriteHelpUsesSuppliedWriterOnly(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	t.Cleanup(func() {
		os.Args = originalArgs
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	})

	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout-*")
	if err != nil {
		t.Fatalf("CreateTemp stdout returned unexpected error: %v", err)
	}
	stderrFile, err := os.CreateTemp(t.TempDir(), "stderr-*")
	if err != nil {
		t.Fatalf("CreateTemp stderr returned unexpected error: %v", err)
	}
	os.Stdout = stdoutFile
	os.Stderr = stderrFile
	os.Args = []string{"ambient", "--help"}
	t.Setenv("DIB_WIDTH", "10")

	var out bytes.Buffer
	if err := mustHelpTree(t).WriteHelp(&out); err != nil {
		t.Fatalf("WriteHelp returned unexpected error: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("WriteHelp wrote no output to supplied writer")
	}
	if err := stdoutFile.Sync(); err != nil {
		t.Fatalf("stdout Sync returned unexpected error: %v", err)
	}
	if err := stderrFile.Sync(); err != nil {
		t.Fatalf("stderr Sync returned unexpected error: %v", err)
	}
	assertEmptyFile(t, stdoutFile)
	assertEmptyFile(t, stderrFile)
}

func TestWriteHelpPropagatesWriterFailuresAndRejectsInvalidDefinitions(t *testing.T) {
	writerErr := errors.New("deterministic writer failure")
	err := mustHelpTree(t).WriteHelp(failingWriter{err: writerErr})
	if !errors.Is(err, writerErr) {
		t.Fatalf("WriteHelp error = %v, want writer error", err)
	}

	var zero command.Definition
	err = zero.WriteHelp(&bytes.Buffer{})
	var nameErr *command.NameError
	if !errors.As(err, &nameErr) {
		t.Fatalf("zero-value WriteHelp error = %T, want *command.NameError", err)
	}
}

func TestWriteHelpIsRepeatableConcurrentAndDefensive(t *testing.T) {
	root := mustHelpTree(t)
	var want bytes.Buffer
	if err := root.WriteHelp(&want); err != nil {
		t.Fatalf("initial WriteHelp returned unexpected error: %v", err)
	}

	children := root.Children()
	children[0] = mustDefinition(t, "mutated")
	aliases := root.Aliases()
	aliases[0] = "mutated"

	var again bytes.Buffer
	if err := root.WriteHelp(&again); err != nil {
		t.Fatalf("second WriteHelp returned unexpected error: %v", err)
	}
	if again.String() != want.String() {
		t.Fatalf("WriteHelp changed after caller slice mutation:\n%s\nwant:\n%s", again.String(), want.String())
	}

	const runs = 32
	errs := make(chan string, runs)
	var wg sync.WaitGroup
	for range runs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var out bytes.Buffer
			if err := root.WriteHelp(&out); err != nil {
				errs <- err.Error()
				return
			}
			if out.String() != want.String() {
				errs <- "non-deterministic help output"
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func TestWriteHelpDoesNotChangeRouteHelpRequestBehavior(t *testing.T) {
	root := mustHelpTree(t)

	result, err := root.Route([]string{"deploy", "apply", "--help"})
	if err == nil {
		t.Fatal("Route returned nil error")
	}
	if !errors.Is(err, flags.ErrHelpRequest) {
		t.Fatalf("Route error does not satisfy flags.ErrHelpRequest: %v", err)
	}
	if got := result.PathNames(); len(got) != 0 {
		t.Fatalf("help-request Route returned non-zero path: %q", got)
	}
}

func mustHelpTree(t *testing.T) command.Definition {
	t.Helper()

	apply := mustDefinition(
		t,
		"apply",
		command.Description("Apply a deployment manifest."),
		command.Aliases("push"),
		command.Usage("<manifest> [flags]"),
		command.LocalFlags(
			flags.String("hidden-token", "dib_fake_secret_value", "Hidden token.", flags.Hidden(), flags.Sensitive()),
			flags.String("token", "dib_fake_token_value", "API token.", flags.Sensitive(), flags.NoOptionDefault("dib_fake_password_value")),
			flags.Bool("dry-run", false, "Print planned operations.", flags.Deprecated("use --preview")),
		),
	)
	plan := mustDefinition(t, "plan", command.Description("Preview deployment."))
	deploy := mustDefinition(
		t,
		"deploy",
		command.Description("Manage deployments."),
		command.Aliases("ship"),
		command.Usage("<action>"),
		command.InheritedFlags(
			flags.String("cluster", "", "Target cluster."),
			flags.Int("retries", 0, "Retry count."),
			flags.StringList("tag", nil, "Deployment tag.", flags.Repeatable()),
		),
		command.Children(apply, plan),
	)
	status := mustDefinition(t, "status", command.Description("Show workspace status."))
	return mustDefinition(
		t,
		"dib",
		command.Description("Developer workspace commands."),
		command.Aliases("db", "devbox"),
		command.Usage("<command> [flags]"),
		command.InheritedFlags(
			flags.Bool("verbose", false, "Enable verbose output.", flags.Shorthand("v")),
			flags.String("profile", "", "Select profile."),
		),
		command.Children(deploy, status),
	)
}

func childNames(definitions []command.Definition) []string {
	names := make([]string, len(definitions))
	for i, definition := range definitions {
		names[i] = definition.Name()
	}
	return names
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}
