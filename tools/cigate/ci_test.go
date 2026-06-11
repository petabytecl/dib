package cigate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIWorkflowRunsTrustGates(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/ci.yml")

	requireAll(t, workflow, map[string]string{
		"workflow name":        "name: CI",
		"pull request trigger": "pull_request:",
		"push trigger":         "push:",
		"explicit runner":      "runs-on: ubuntu-24.04",
		"checkout action":      "uses: actions/checkout@v6",
		"setup-go action":      "uses: actions/setup-go@v6",
		"go.mod version file":  "go-version-file: go.mod",
		"cache disabled":       "cache: false",
		"go version evidence":  "run: go version",
		"go test gate":         "run: go test ./...",
		"go vet gate":          "run: go vet ./...",
		"dependency gate":      "run: go run ./tools/depgate",
	})

	if !triggersMainBranch(workflow) {
		t.Fatalf("CI workflow must run push checks for the main branch:\n%s", workflow)
	}
}

func TestCIWorkflowUsesOnlyTrustedStaticSteps(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/ci.yml")

	denyAll(t, workflow, map[string]string{
		"floating runner":                  "ubuntu-latest",
		"untrusted event interpolation":    "${{ github.event",
		"untrusted head ref interpolation": "${{ github.head_ref",
		"untrusted workflow input":         "${{ inputs.",
		"remote shell installer":           "curl ",
		"dependency fetch":                 "go get ",
		"module rewrite":                   "go mod tidy",
	})

	for _, line := range strings.Split(workflow, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "uses:") {
			continue
		}
		action := strings.TrimSpace(strings.TrimPrefix(trimmed, "uses:"))
		if action != "actions/checkout@v6" && action != "actions/setup-go@v6" {
			t.Fatalf("CI workflow uses unsupported action %q; only official checkout/setup-go actions are allowed", action)
		}
	}
}

func TestReleaseChecklistCapturesRequiredEvidence(t *testing.T) {
	checklist := readRepoFile(t, "docs/release-checklist.md")
	lowerChecklist := strings.ToLower(checklist)

	requireAll(t, lowerChecklist, map[string]string{
		"title":                  "release checklist",
		"commit evidence":        "exact commit",
		"go version policy":      "go version alignment",
		"go.mod reference":       "go.mod",
		"ci reference":           ".github/workflows/ci.yml",
		"go test gate":           "go test ./...",
		"go vet gate":            "go vet ./...",
		"dependency gate":        "go run ./tools/depgate",
		"race gate":              "go test -race ./...",
		"docs examples":          "docs/examples",
		"provenance evidence":    "provenance",
		"compatibility evidence": "compatibility",
		"migration evidence":     "migration",
		"runner image":           "runner image",
		"checkout version":       "actions/checkout",
		"setup-go version":       "actions/setup-go",
		"owner":                  "owner",
		"date":                   "date",
		"waiver":                 "waiver",
		"expiry":                 "expiry",
		"tagging block":          "ci failures block tagging",
		"go module release":      "go module tag",
	})
}

func TestStory15DoesNotWeakenStandardLibraryClaim(t *testing.T) {
	goMod := readRepoFile(t, "go.mod")
	workflow := readRepoFile(t, ".github/workflows/ci.yml")

	if got, want := strings.TrimSpace(goMod), "module github.com/petabytecl/dib\n\ngo 1.26"; got != want {
		t.Fatalf("go.mod drifted from the zero-dependency baseline:\n%s", goMod)
	}

	denyAll(t, goMod, map[string]string{
		"require directive":   "\nrequire ",
		"replace directive":   "\nreplace ",
		"toolchain directive": "\ntoolchain ",
	})
	denyAll(t, workflow, map[string]string{
		"generated go.sum dependency cache": "cache: true",
	})

	if _, err := os.Stat(repoPath("go.sum")); err == nil {
		t.Fatal("go.sum exists at the repository root; Story 1.5 must not introduce project dependencies")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat go.sum: %v", err)
	}
}

func readRepoFile(t *testing.T, relativePath string) string {
	t.Helper()

	data, err := os.ReadFile(repoPath(relativePath))
	if err != nil {
		t.Fatalf("read %s: %v", relativePath, err)
	}
	return string(data)
}

func repoPath(relativePath string) string {
	return filepath.Join("..", "..", relativePath)
}

func requireAll(t *testing.T, content string, required map[string]string) {
	t.Helper()

	for label, needle := range required {
		if !strings.Contains(content, needle) {
			t.Fatalf("missing %s marker %q in content:\n%s", label, needle, content)
		}
	}
}

func denyAll(t *testing.T, content string, denied map[string]string) {
	t.Helper()

	for label, needle := range denied {
		if strings.Contains(content, needle) {
			t.Fatalf("found denied %s marker %q in content:\n%s", label, needle, content)
		}
	}
}

func triggersMainBranch(workflow string) bool {
	return strings.Contains(workflow, "branches: [main]") ||
		strings.Contains(workflow, "branches:\n      - main") ||
		strings.Contains(workflow, "branches:\n        - main")
}
