package docs

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestReleaseChecklistRecordsStory53EvidenceInputs(t *testing.T) {
	content, err := os.ReadFile("release-checklist.md")
	if err != nil {
		t.Fatalf("read release checklist: %v", err)
	}
	text := string(content)
	lower := strings.ToLower(text)

	for _, heading := range []string{
		"# Release Checklist",
		"## Release Identity",
		"## Go Version Alignment",
		"## CI Trust Gates",
		"## Release-Candidate Gates",
		"## Standard-Library Dependency Evidence",
		"## Waivers",
		"## Final Review",
	} {
		if !strings.Contains(text, heading) {
			t.Fatalf("release checklist missing heading %q", heading)
		}
	}

	required := []string{
		"`go test ./...`:",
		"`go vet ./...`:",
		"`go run ./tools/depgate`:",
		"`go test -race ./...`:",
		"`go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags`:",
		"`go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags`:",
		"`go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags`:",
		"`docs/behavior-matrices.md` consolidates story",
		"5.3 adoption evidence",
		"`docs/provenance-log.md`",
		"`docs/compatibility.md`",
		"`examples/migration/`",
		"story 5.4 records final release-candidate command outcomes",
		"root `go.mod` contains no `require`, `replace`, or `toolchain` directives:",
		"root `go.sum` absent:",
		"dependency gate reviewed:",
	}
	for _, phrase := range required {
		if !strings.Contains(lower, strings.ToLower(phrase)) {
			t.Fatalf("release checklist missing Story 5.3 evidence phrase %q", phrase)
		}
	}
}

func TestReleaseChecklistLeavesStory54OutcomesUnfilled(t *testing.T) {
	content, err := os.ReadFile("release-checklist.md")
	if err != nil {
		t.Fatalf("read release checklist: %v", err)
	}
	text := string(content)
	lower := strings.ToLower(text)

	emptyFields := []string{
		"Go module tag",
		"Exact commit",
		"Owner",
		"Date",
		"Reviewer",
		"`go.mod` version",
		"Release guidance version",
		"Documentation version references",
		"Drift review result",
		"`go test ./...`",
		"`go vet ./...`",
		"`go run ./tools/depgate`",
		"`go test -race ./...`",
		"`go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags`",
		"`go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags`",
		"`go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags`",
		"Root `go.mod` contains no `require`, `replace`, or `toolchain` directives",
		"Root `go.sum` absent",
		"Dependency gate reviewed",
		"Any fixture-local dependency exceptions",
		"All required evidence captured",
		"All waivers approved with expiry",
		"Tagging decision",
	}
	for _, field := range emptyFields {
		assertChecklistFieldEmpty(t, text, field)
	}

	for _, prohibited := range []string{
		"ready to tag",
		"tag readiness complete",
		"release readiness is complete",
		"all required evidence captured: yes",
		"tagging decision: yes",
		"tagging decision: approved",
		"`go test ./...`: pass",
		"`go vet ./...`: pass",
		"`go run ./tools/depgate`: pass",
	} {
		if strings.Contains(lower, prohibited) {
			t.Fatalf("release checklist pre-fills Story 5.4 outcome with %q", prohibited)
		}
	}
}

func assertChecklistFieldEmpty(t *testing.T, text, field string) {
	t.Helper()

	pattern := regexp.MustCompile(`(?m)^\s*- ` + regexp.QuoteMeta(field) + `:\s*$`)
	if !pattern.MatchString(text) {
		t.Fatalf("release checklist field %q must exist and remain empty for Story 5.4", field)
	}
}
