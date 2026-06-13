package docs

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestREADMEExistsAndCoversAdoptionOnboarding(t *testing.T) {
	content, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	text := string(content)
	lower := strings.ToLower(text)

	for _, phrase := range []string{
		"github.com/petabytecl/dib/flags",
		"github.com/petabytecl/dib/command",
		"github.com/petabytecl/dib/config",
		"github.com/petabytecl/dib/cli",
		"## Status",
		"## Packages",
		"## Install",
		"## Quickstart",
		"## Compatibility",
		"## Documentation",
		"go get github.com/petabytecl/dib",
		"flags.NewSet",
		"set.Parse",
		"command.NewDefinition",
		"config.NewSet",
		"config.Resolve",
		"Using command, flags, and config together",
		"Dispatching application commands",
	} {
		if !strings.Contains(text, phrase) {
			t.Fatalf("README.md missing required phrase %q", phrase)
		}
	}

	for _, phrase := range []string{
		"v0",
		"go 1.26",
		"docs/compatibility.md",
		"docs/behavior-matrices.md",
		"docs/testing.md",
		"docs/config-precedence.md",
		"docs/diagnostics-and-errors.md",
		"docs/release-checklist.md",
		"contributing.md",
		"examples/migration/",
		"examples/multicommand/",
	} {
		if !strings.Contains(lower, phrase) {
			t.Fatalf("README.md missing required phrase (case-insensitive) %q", phrase)
		}
	}
}

func TestREADMEQuickstartUsesRealAPI(t *testing.T) {
	content, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	text := string(content)

	// Verify the quickstart snippets use the real Dib API function signatures,
	// not stubs or placeholder names. Each phrase corresponds to an actual
	// exported identifier verified against the source files during Story 6.3.
	for _, phrase := range []string{
		"result.PathNames()",
		"config.NewEnvSnapshot",
		"config.BindEnv",
		"state.Values()",
		".GetString(",
		"cli.FromOSArgs",
		"cli.Resolve",
		"result.Route().PathNames()",
		"case \"start\"",
		"case \"stop\"",
	} {
		if !strings.Contains(text, phrase) {
			t.Fatalf("README.md quickstart missing real API phrase %q", phrase)
		}
	}
}

func TestREADMEDoesNotImplySourceCompatibility(t *testing.T) {
	content, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	text := string(content)

	boundaryTerms := regexp.MustCompile(`(?i)drop-in|source-compatible clone|clone API|framework compatibility layer`)
	limitationFrame := regexp.MustCompile(`(?i)\b(not|never|no|omits?|without|avoid|does not|do not)\b`)
	for _, line := range strings.Split(text, "\n") {
		if boundaryTerms.MatchString(line) && !limitationFrame.MatchString(line) {
			t.Fatalf("README.md contains compatibility term without limitation framing: %q", line)
		}
	}

	for _, prohibited := range []string{
		"compatible replacement",
		"clone api",
	} {
		if strings.Contains(strings.ToLower(text), prohibited) {
			t.Fatalf("README.md contains prohibited positive compatibility claim %q", prohibited)
		}
	}
}
