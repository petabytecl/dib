package docs

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestReleaseChecklistRecordsReleaseCandidateEvidence(t *testing.T) {
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
		"go module tag: `v0.1.0`",
		"exact commit: `d5ce41ce693413b88df95e644eb4358702ae205e`",
		"owner: coto",
		"date: 2026-06-12",
		"reviewer: release reviewer",
		"story 6.1 evidence scope:",
		"`go.mod` version: `go 1.26`",
		"ci `actions/setup-go` source: `go-version-file: go.mod`",
		"release guidance version: `docs/release-notes-v0.md` states go 1.26+",
		"documentation version references:",
		"drift review result:",
		"`go test ./...`:",
		"`go run ./tools/lint`:",
		"`go vet ./...`:",
		"`go run ./tools/depgate`:",
		"`go test -race ./...`:",
		"`go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags`:",
		"`go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags`:",
		"`go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags`:",
		"`docs/behavior-matrices.md` records story 5.4 release-candidate evidence",
		"`docs/provenance-log.md`",
		"`docs/compatibility.md`",
		"`examples/migration/`",
		"root `go.mod` contains no `require`, `replace`, or `toolchain` directives:",
		"root `go.sum` absent:",
		"dependency gate reviewed:",
		"fixture-local external modules are isolated under `tools/depgate/testdata/`",
		"v0 experimental api status",
	}
	for _, phrase := range required {
		if !strings.Contains(lower, strings.ToLower(phrase)) {
			t.Fatalf("release checklist missing release evidence phrase %q", phrase)
		}
	}

	for _, field := range []string{
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
		"Lint gate reviewed",
		"Lint isolation evidence",
		"Lint pinning evidence",
		"Any fixture-local dependency exceptions",
		"All required evidence captured",
		"All waivers approved with expiry",
		"Tagging decision",
	} {
		assertChecklistFieldFilled(t, text, field)
	}
}

func TestReleaseChecklistGuardsReleaseReadinessClaims(t *testing.T) {
	content, err := os.ReadFile("release-checklist.md")
	if err != nil {
		t.Fatalf("read release checklist: %v", err)
	}
	lower := strings.ToLower(string(content))

	for _, prohibited := range []string{
		"ready to tag",
		"tag readiness complete",
		"release readiness is complete",
		"docker",
		"kubernetes",
		"binary deployment",
		"generated shell completion",
		"generated man pages",
	} {
		if strings.Contains(lower, prohibited) {
			t.Fatalf("release checklist contains prohibited release-readiness or deployment claim %q", prohibited)
		}
	}
}

func TestReleaseChecklistEvidenceMatchesRepositoryState(t *testing.T) {
	content, err := os.ReadFile("release-checklist.md")
	if err != nil {
		t.Fatalf("read release checklist: %v", err)
	}
	text := string(content)

	requiredChecklistMatch(t, text, "exact commit", `(?m)^- Exact commit: `+"`"+`([0-9a-f]{40})`+"`"+`$`)

	goMod, err := os.ReadFile("../go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	goVersion := requiredChecklistMatch(t, string(goMod), "go.mod go directive", `(?m)^go ([0-9]+(?:\.[0-9]+)*)$`)
	recordedGoVersion := requiredChecklistMatch(t, text, "recorded go.mod version", `(?m)^- `+"`"+`go\.mod`+"`"+` version: `+"`"+`(go [^`+"`"+`]+)`+"`"+`$`)
	if recordedGoVersion != "go "+goVersion {
		t.Fatalf("release checklist go.mod version = %q, want %q from go.mod", recordedGoVersion, "go "+goVersion)
	}

	for _, directive := range []string{"require", "replace", "toolchain"} {
		if regexp.MustCompile(`(?m)^` + directive + `\b`).Match(goMod) {
			t.Fatalf("go.mod contains %s directive while release checklist records zero external dependency baseline", directive)
		}
	}
	if _, err := os.Stat("../go.sum"); err == nil {
		t.Fatal("go.sum exists while release checklist records it absent")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat go.sum: %v", err)
	}
}

func TestReleaseChecklistRecordsPassingRequiredGates(t *testing.T) {
	content, err := os.ReadFile("release-checklist.md")
	if err != nil {
		t.Fatalf("read release checklist: %v", err)
	}
	text := string(content)

	for _, command := range []string{
		"go test ./...",
		"go run ./tools/lint",
		"go vet ./...",
		"go run ./tools/depgate",
		"go test -race ./...",
		"go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags",
		"go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags",
		"go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags",
	} {
		pattern := regexp.MustCompile(`(?m)^\s*-\s*` + "`" + regexp.QuoteMeta(command) + "`" + `: PASS on \d{4}-\d{2}-\d{2} with ` + "`" + `GOCACHE=/tmp/dib-go-build ` + regexp.QuoteMeta(command) + "`")
		if !pattern.MatchString(text) {
			t.Fatalf("release checklist must record passing gate outcome for %q", command)
		}
	}
}

func TestReleaseChecklistRejectsFloatingLintEvidence(t *testing.T) {
	content, err := os.ReadFile("release-checklist.md")
	if err != nil {
		t.Fatalf("read release checklist: %v", err)
	}
	lower := strings.ToLower(string(content))

	for _, prohibited := range []string{
		"golangci-lint@latest",
		"golangci-lint@stable",
		"version: latest",
		"version: stable",
		"go install ",
		"curl ",
	} {
		if strings.Contains(lower, prohibited) {
			t.Fatalf("release checklist contains floating or installer-based lint evidence %q", prohibited)
		}
	}
}

func TestReleaseChecklistRequiresCompleteWaiverShape(t *testing.T) {
	content, err := os.ReadFile("release-checklist.md")
	if err != nil {
		t.Fatalf("read release checklist: %v", err)
	}
	text := string(content)

	if !strings.Contains(text, "| Item | Owner | Reason | Expiry | Impact |") {
		t.Fatal("waiver table must include owner, reason, expiry, and impact columns")
	}

	rowPattern := regexp.MustCompile(`(?m)^\| ([^|]+) \| ([^|]+) \| ([^|]+) \| ([^|]+) \| ([^|]+) \|$`)
	for _, match := range rowPattern.FindAllStringSubmatch(text, -1) {
		item := strings.TrimSpace(match[1])
		if item == "Item" || item == "---" {
			continue
		}
		for i, value := range match[2:] {
			if strings.TrimSpace(value) == "" {
				t.Fatalf("waiver row %q has blank required column %d", match[0], i+2)
			}
		}
	}
}

func TestReleaseNotesV0ExistAndPreserveBoundaries(t *testing.T) {
	content, err := os.ReadFile("release-notes-v0.md")
	if err != nil {
		t.Fatalf("read v0 release notes: %v", err)
	}
	lower := strings.ToLower(string(content))

	for _, phrase := range []string{
		"v0 experimental api status",
		"go 1.26 or newer",
		"standard-library-only",
		"clean-room",
		"redaction",
		"not a source-compatible clone",
		"not a drop-in replacement",
		"migration examples",
		"`go test ./...`",
		"`go run ./tools/lint`",
		"`go vet ./...`",
		"`go run ./tools/depgate`",
		"`go test -race ./...`",
	} {
		if !strings.Contains(lower, phrase) {
			t.Fatalf("release notes missing boundary phrase %q", phrase)
		}
	}
}

func TestReleaseNotesV0AvoidReleaseScopeAssumptions(t *testing.T) {
	content, err := os.ReadFile("release-notes-v0.md")
	if err != nil {
		t.Fatalf("read v0 release notes: %v", err)
	}
	lower := strings.ToLower(string(content))

	for _, prohibited := range []string{
		"ready to tag",
		"release readiness is complete",
		"tag readiness complete",
		"docker",
		"kubernetes",
		"binary deployment",
		"generated shell completion",
		"generated man pages",
	} {
		if strings.Contains(lower, prohibited) {
			t.Fatalf("release notes contain prohibited release-scope claim %q", prohibited)
		}
	}
}

func assertChecklistFieldFilled(t *testing.T, text, field string) {
	t.Helper()

	pattern := regexp.MustCompile(`(?m)^\s*- ` + regexp.QuoteMeta(field) + `:\s+\S`)
	if !pattern.MatchString(text) {
		t.Fatalf("release checklist field %q must exist and be filled", field)
	}
}

func requiredChecklistMatch(t *testing.T, text, label, expression string) string {
	t.Helper()

	pattern := regexp.MustCompile(expression)
	match := pattern.FindStringSubmatch(text)
	if match == nil {
		t.Fatalf("missing %s matching %s", label, expression)
	}
	return strings.TrimSpace(match[1])
}
