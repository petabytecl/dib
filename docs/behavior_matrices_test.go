package docs

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestBehaviorMatricesCoverAdoptionEvidenceRows(t *testing.T) {
	content, err := os.ReadFile("behavior-matrices.md")
	if err != nil {
		t.Fatalf("read behavior matrix: %v", err)
	}
	text := string(content)
	lower := strings.ToLower(text)

	for _, heading := range []string{
		"# Behavior Matrices",
		"## How To Audit This Matrix",
		"## Consolidated Adoption Evidence",
		"## Rendering And Diagnostic Evidence",
		"## Dependency And Release Evidence",
		"## Current Scope Boundaries",
	} {
		if !strings.Contains(text, heading) {
			t.Fatalf("behavior matrix missing heading %q", heading)
		}
	}

	requiredRows := map[string][]string{
		"shared immutability, snapshots, and explicit instances": {"story 2.1", "story 3.1", "story 4.1", "fr20", "current"},
		"flag definitions and parsing":                           {"story 2.1", "story 2.8", "fr5", "fr6", "fr22", "current"},
		"parser boundaries":                                      {"story 2.7", "fr9", "errhelprequest", "current"},
		"fuzz and property hardening":                            {"story 2.8", "fuzzparse", "fuzzparseshortgroups", "current"},
		"command routing and aliases":                            {"story 3.1", "story 3.2", "fr1", "current"},
		"local and inherited command flags":                      {"story 3.3", "fr3", "current"},
		"help and usage rendering":                               {"story 3.4", "fr4", "current"},
		"execution-boundary metadata":                            {"story 3.5", "fr2", "current"},
		"config definitions and source ingestion":                {"story 4.1", "story 4.2", "fr11", "fr12", "current"},
		"config precedence and typed getters":                    {"story 4.3", "story 4.4", "fr13", "fr16", "current"},
		"source reports, diagnostics, and redaction":             {"story 4.5", "nfr8", "current"},
		"compatibility boundaries":                               {"story 5.1", "nfr7", "supported", "narrowed", "omitted", "intentionally different"},
		"migration examples":                                     {"story 5.2", "fr18", "fr19", "current"},
		"dependency gate evidence":                               {"story 5.3", "story 6.1", "story 6.2", "fr21", "fr23", "fr24", "go run ./tools/depgate", "golangci-lint run", "go run ./tools/coverage", "current"},
		"release-candidate evidence":                             {"story 5.4", "docs/release-notes-v0.md", "current"},
		"public usage documentation":                             {"story 6.3", "fr25", "nfr12", "readme.md", "current"},
		"release hardening reconciliation":                       {"story 6.4", "fr23", "fr24", "fr25", "nfr11", "docs/release-notes-v0.md", "docs/release-checklist.md", "current"},
		"cli composition ergonomics":                             {"story 7.1", "story 7.3", "issue #52", "fr26", "cli/resolve_test.go", "testcommandrundispatchesdistributedsubcommand", "example_lowleveldispatch", "current"},
	}

	for row, phrases := range requiredRows {
		if !strings.Contains(lower, row) {
			t.Fatalf("behavior matrix missing row %q", row)
		}
		for _, phrase := range phrases {
			if !strings.Contains(lower, phrase) {
				t.Fatalf("behavior matrix row set missing phrase %q required by %q", phrase, row)
			}
		}
	}
}

func TestBehaviorMatrixEvidenceReferencesResolve(t *testing.T) {
	content, err := os.ReadFile("behavior-matrices.md")
	if err != nil {
		t.Fatalf("read behavior matrix: %v", err)
	}
	text := string(content)
	root := filepath.Clean("..")

	for _, path := range referencedLocalPaths(text) {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("behavior matrix references missing local path %q: %v", path, err)
		}
	}

	knownFunctions := repositoryTestFunctions(t, root)
	for _, name := range referencedGoTestFunctions(text) {
		if !knownFunctions[name] {
			t.Fatalf("behavior matrix references missing Go test/example/fuzz function %q", name)
		}
	}
}

func TestBehaviorMatrixKeepsDeferredAndBoundaryLanguage(t *testing.T) {
	content, err := os.ReadFile("behavior-matrices.md")
	if err != nil {
		t.Fatalf("read behavior matrix: %v", err)
	}
	text := string(content)
	lower := strings.ToLower(text)

	for _, prohibited := range []string{
		"story 5.4 release readiness complete",
		"release readiness is complete",
		"tag readiness complete",
		"ready to tag",
	} {
		if strings.Contains(lower, prohibited) {
			t.Fatalf("behavior matrix overstates release readiness with %q", prohibited)
		}
	}
	if !strings.Contains(lower, "story 5.4") || !strings.Contains(lower, "human release review") {
		t.Fatal("behavior matrix must keep Story 5.4 release evidence scoped to human release review")
	}

	boundaryTerms := regexp.MustCompile(`(?i)drop-in|source-compatible|source compatibility|clone API|framework compatibility layer|compatible replacement`)
	limitationFrame := regexp.MustCompile(`(?i)\b(not|never|no|omits?|without|avoid|does not|do not)\b`)
	for _, line := range strings.Split(text, "\n") {
		if boundaryTerms.MatchString(line) && !limitationFrame.MatchString(line) {
			t.Fatalf("behavior matrix contains ambiguous compatibility positioning: %q", line)
		}
	}
}

func TestBehaviorMatrixSensitiveCorpusOnlyNamesRedactionCorpus(t *testing.T) {
	content, err := os.ReadFile("behavior-matrices.md")
	if err != nil {
		t.Fatalf("read behavior matrix: %v", err)
	}

	corpus := []string{"dib_fake_secret_value", "dib_fake_password_value", "dib_fake_token_value"}
	for lineNumber, line := range strings.Split(string(content), "\n") {
		for _, value := range corpus {
			if !strings.Contains(line, value) {
				continue
			}
			lower := strings.ToLower(line)
			if !strings.Contains(lower, "fake sensitive corpus") && !strings.Contains(lower, "redaction corpus") {
				t.Fatalf("sensitive corpus value %q appears outside corpus naming context on line %d: %q", value, lineNumber+1, line)
			}
		}
	}
}

func referencedLocalPaths(text string) []string {
	candidates := make(map[string]bool)
	for _, pattern := range []*regexp.Regexp{
		regexp.MustCompile(`\]\(([^)#]+)(?:#[^)]+)?\)`),
		regexp.MustCompile("`([A-Za-z0-9_./-]+\\.(?:md|go|yml|yaml|mod|sum))`"),
		regexp.MustCompile("`([A-Za-z0-9_./-]+/)`"),
	} {
		for _, match := range pattern.FindAllStringSubmatch(text, -1) {
			path := strings.TrimPrefix(match[1], "../")
			if path == "" || strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
				continue
			}
			if strings.HasPrefix(path, "#") {
				continue
			}
			candidates[path] = true
		}
	}

	paths := make([]string, 0, len(candidates))
	for path := range candidates {
		paths = append(paths, path)
	}
	return paths
}

func referencedGoTestFunctions(text string) []string {
	pattern := regexp.MustCompile("`((?:Test|Fuzz|Example)[A-Za-z0-9_]*)`")
	matches := pattern.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var names []string
	for _, match := range matches {
		name := match[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}

func repositoryTestFunctions(t *testing.T, root string) map[string]bool {
	t.Helper()

	functions := make(map[string]bool)
	pattern := regexp.MustCompile(`func\s+((?:Test|Fuzz|Example)[A-Za-z0-9_]*)\s*\(`)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".idea", ".codex", "_bmad-output":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range pattern.FindAllStringSubmatch(string(content), -1) {
			functions[match[1]] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repository test files: %v", err)
	}
	return functions
}
