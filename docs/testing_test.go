package docs

import (
	"os"
	"strings"
	"testing"
)

func TestTestingGuideDocumentsCoverageGate(t *testing.T) {
	content, err := os.ReadFile("testing.md")
	if err != nil {
		t.Fatalf("read testing guide: %v", err)
	}
	text := string(content)
	lower := strings.ToLower(text)

	for _, phrase := range []string{
		"## Coverage Gate",
		"GOCACHE=/tmp/dib-go-build go run ./tools/coverage",
		"go run ./tools/coverage",
		"tools/coverage",
	} {
		if !strings.Contains(text, phrase) {
			t.Fatalf("testing guide missing coverage gate phrase %q", phrase)
		}
	}
	for _, phrase := range []string{
		"command",
		"config",
		"flags",
		"cli",
		"threshold",
		"tools/depgate",
		"tools/coverage",
		"tooling package",
		"exception granted",
		"testcoveragepassespackagesmeetingthreshold",
		"testcoveragefailspackagesbelowthreshold",
		"testcoveragecommandruns",
	} {
		if !strings.Contains(lower, phrase) {
			t.Fatalf("testing guide missing coverage gate phrase %q", phrase)
		}
	}
}

func TestTestingGuideCLIPackageListedAtCorrectThreshold(t *testing.T) {
	content, err := os.ReadFile("testing.md")
	if err != nil {
		t.Fatalf("read testing guide: %v", err)
	}
	text := string(content)

	// Story 7.4 extended the coverage gate to include cli as the fourth public
	// runtime package at the same 85% threshold as command, config, and flags.
	for _, phrase := range []string{
		"`cli`: 85%",
		"Story 7.4 extended",
		"fourth",
	} {
		if !strings.Contains(text, phrase) {
			t.Fatalf("testing guide missing Story 7.4 CLI coverage phrase %q", phrase)
		}
	}
}

func TestTestingGuideDocumentsLintGateIsolationAndPinning(t *testing.T) {
	content, err := os.ReadFile("testing.md")
	if err != nil {
		t.Fatalf("read testing guide: %v", err)
	}
	text := string(content)
	lower := strings.ToLower(text)

	for _, phrase := range []string{
		"# Testing",
		"## Lint Gate",
		"golangci-lint",
		"`.golangci.yml`",
		"golangci/golangci-lint-action@v7",
		"v2.10.1",
		"go.mod",
		"root `require`",
		"root `replace`",
		"`toolchain`",
		"go sum file",
		"Rejected alternatives:",
	} {
		if !strings.Contains(text, phrase) {
			t.Fatalf("testing guide missing lint gate phrase %q", phrase)
		}
	}
	for _, phrase := range []string{
		"golangci-lint run",
		"depguard",
		"dependency-free",
		"external",
		"pinned",
		"effective pin",
		"third-party imports",
		"rejected because",
	} {
		if !strings.Contains(lower, phrase) {
			t.Fatalf("testing guide missing lint gate phrase %q", phrase)
		}
	}

	for _, prohibited := range []string{
		"golangci-lint@latest",
		"golangci-lint@stable",
		"version: latest",
		"version: stable",
		"@latest",
		"@stable",
	} {
		if strings.Contains(lower, prohibited) {
			t.Fatalf("testing guide contains floating or installer-based lint guidance %q", prohibited)
		}
	}
}
