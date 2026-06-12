package docs

import (
	"os"
	"strings"
	"testing"
)

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
		"repository-local standard-library lint tool",
		"GOCACHE=/tmp/dib-go-build go run ./tools/lint",
		"go run ./tools/lint",
		"tools/lint",
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
		"checked-out dib",
		"commit plus the go version selected from `go.mod`",
		"no external linter",
		"binary version to record",
		"skipping repository metadata",
		"bmad/agent artifact directories",
		"does not add",
		"third-party imports",
		"go install",
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
