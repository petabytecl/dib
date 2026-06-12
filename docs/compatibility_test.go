package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

func TestCompatibilityDocumentBoundaries(t *testing.T) {
	content, err := os.ReadFile("compatibility.md")
	if err != nil {
		t.Fatalf("read compatibility document: %v", err)
	}
	text := string(content)
	lower := strings.ToLower(text)

	required := []string{
		"# compatibility boundaries",
		"clean-room native go api",
		"not a source-compatible clone",
		"not a drop-in replacement",
		"not a framework compatibility layer",
		"| inspiration surface | supported familiar concepts | narrowed behavior | omitted behavior | intentionally different behavior | user-facing reason | dib evidence |",
		"go `flag`",
		"pflag",
		"cobra",
		"viper",
		"explicit setter > flag binding > env > json > default",
		"examples/migration",
		"story 5.2",
		"story 5.3",
		"story 5.4",
		"consolidated adoption evidence",
		"does not claim tag approval",
	}
	for _, phrase := range required {
		if !strings.Contains(lower, phrase) {
			t.Fatalf("compatibility document missing required phrase %q", phrase)
		}
	}

	for _, link := range []string{
		"behavior-matrices.md",
		"behavior-matrices.md#consolidated-adoption-evidence",
		"config-precedence.md",
		"diagnostics-and-errors.md",
		"release-checklist.md",
		"provenance-log.md",
	} {
		if !strings.Contains(text, link) {
			t.Fatalf("compatibility document missing evidence link to %s", link)
		}
	}
}

func TestCompatibilityDocumentDoesNotMakePositiveCompatibilityClaims(t *testing.T) {
	content, err := os.ReadFile("compatibility.md")
	if err != nil {
		t.Fatalf("read compatibility document: %v", err)
	}
	text := string(content)

	boundaryTerms := regexp.MustCompile(`(?i)drop-in|source-compatible|clone API|framework compatibility layer`)
	for _, line := range strings.Split(text, "\n") {
		if boundaryTerms.MatchString(line) && !regexp.MustCompile(`(?i)\b(not|never|no|omits?|without)\b`).MatchString(line) {
			t.Fatalf("compatibility boundary term is not framed as a limitation: %q", line)
		}
	}

	prohibitedClaims := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bcompatible replacement\b`),
		regexp.MustCompile(`(?i)\brelease readiness is complete\b`),
	}
	for _, pattern := range prohibitedClaims {
		if match := pattern.FindString(text); match != "" {
			t.Fatalf("compatibility document contains prohibited positive claim %q", match)
		}
	}
}

func TestCompatibilityEvidenceDocsAvoidAmbiguousPositioning(t *testing.T) {
	for _, path := range []string{
		"behavior-matrices.md",
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read compatibility evidence document %s: %v", path, err)
		}

		boundaryTerms := regexp.MustCompile(`(?i)drop-in|source-compatible|source compatibility|clone API|framework compatibility layer|compatible replacement`)
		limitationFrame := regexp.MustCompile(`(?i)\b(not|never|no|omits?|without|avoid|does not|do not)\b`)
		for _, line := range strings.Split(string(content), "\n") {
			if boundaryTerms.MatchString(line) && !limitationFrame.MatchString(line) {
				t.Fatalf("%s contains ambiguous compatibility positioning: %q", path, line)
			}
		}
	}
}

func TestCompatibilityTableCoversStorySurfaces(t *testing.T) {
	content, err := os.ReadFile("compatibility.md")
	if err != nil {
		t.Fatalf("read compatibility document: %v", err)
	}

	rows := compatibilityRows(t, string(content))
	checks := map[string][]string{
		"go flag": {
			"explicit flag sets",
			"typed flag values",
			"default snapshots",
			"custom parsing",
			"parse errors",
			"package-global `commandline` helpers",
			"implicit reads of `os.args`",
			"typed `*flags.parseerror`",
		},
		"pflag": {
			"long flags",
			"one-character shorthands",
			"shorthand groups",
			"no-option defaults",
			"repeated values",
			"custom values",
			"`--` boundaries",
			"interspersed args",
			"automatic `--no-*` aliases",
		},
		"cobra": {
			"nested command routing",
			"aliases",
			"local and inherited flags",
			"deterministic help and usage rendering",
			"caller-controlled execution-boundary metadata",
			"shell completion",
			"manpage generation",
			"scaffolding",
			"process lifecycle ownership",
		},
		"viper": {
			"defaults",
			"explicit setters",
			"flag bindings",
			"env bindings",
			"json readers and paths",
			"explicit setter > flag binding > env > json > default",
			"typed getters",
			"source reports",
			"package-global singleton config",
			"non-json formats",
			"remote stores",
			"live reload",
			"reflection-heavy struct decoding",
		},
	}

	if len(rows) != len(checks) {
		t.Fatalf("compatibility table has %d data rows, want %d: %#v", len(rows), len(checks), rows)
	}
	for surface, phrases := range checks {
		row, ok := rows[surface]
		if !ok {
			t.Fatalf("compatibility table missing row for %q; got rows %#v", surface, rows)
		}
		for _, phrase := range phrases {
			if !strings.Contains(row, phrase) {
				t.Fatalf("compatibility table row %q missing phrase %q in %q", surface, phrase, row)
			}
		}
	}
}

func TestCompatibilityEvidenceLinksResolve(t *testing.T) {
	content, err := os.ReadFile("compatibility.md")
	if err != nil {
		t.Fatalf("read compatibility document: %v", err)
	}

	linkPattern := regexp.MustCompile(`\[[^]]+\]\(([^)]+)\)`)
	matches := linkPattern.FindAllStringSubmatch(string(content), -1)
	if len(matches) == 0 {
		t.Fatal("compatibility document has no markdown evidence links")
	}

	for _, match := range matches {
		link := match[1]
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			continue
		}
		target, anchor, _ := strings.Cut(link, "#")
		if target == "" {
			t.Fatalf("compatibility document contains anchor-only link %q", link)
		}

		targetPath := filepath.Clean(target)
		if filepath.IsAbs(targetPath) || (strings.HasPrefix(targetPath, "..") && !strings.HasPrefix(targetPath, "../examples/")) {
			t.Fatalf("compatibility document contains non-local evidence link %q", link)
		}
		targetContent, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("compatibility evidence link %q does not resolve: %v", link, err)
		}
		if anchor == "" {
			continue
		}
		anchors := markdownAnchors(string(targetContent))
		if !anchors[anchor] {
			t.Fatalf("compatibility evidence link %q has missing anchor; available anchors %#v", link, anchors)
		}
	}
}

func TestCompatibilityMigrationExampleEvidence(t *testing.T) {
	content, err := os.ReadFile("compatibility.md")
	if err != nil {
		t.Fatalf("read compatibility document: %v", err)
	}
	text := string(content)

	for _, path := range []string{
		"../examples/migration/standard_flag_concepts_test.go",
		"../examples/migration/shorthand_flag_migration_test.go",
		"../examples/migration/nested_command_migration_test.go",
		"../examples/migration/config_precedence_migration_test.go",
	} {
		if !strings.Contains(text, path) {
			t.Fatalf("compatibility document missing migration evidence link %s", path)
		}
		example, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration evidence %s: %v", path, err)
		}
		if !strings.Contains(string(example), "Example_") {
			t.Fatalf("migration evidence %s does not contain runnable Example function", path)
		}
	}
}

func TestCompatibilityCleanRoomProvenanceBoundary(t *testing.T) {
	content, err := os.ReadFile("compatibility.md")
	if err != nil {
		t.Fatalf("read compatibility document: %v", err)
	}
	lower := normalizeWhitespace(strings.ToLower(string(content)))

	for _, phrase := range []string{
		"provenance boundary",
		"inspiration-only",
		"does not copy external examples",
		"fixtures",
		"internal names",
		"document layout",
	} {
		if !strings.Contains(lower, phrase) {
			t.Fatalf("compatibility document missing clean-room phrase %q", phrase)
		}
	}
}

func compatibilityRows(t *testing.T, text string) map[string]string {
	t.Helper()

	rows := make(map[string]string)
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, "| ---") {
			continue
		}
		cells := splitMarkdownTableRow(trimmed)
		if len(cells) != 7 {
			t.Fatalf("compatibility table row has %d cells, want 7: %q", len(cells), line)
		}
		if strings.EqualFold(cells[0], "Inspiration surface") {
			continue
		}

		key := normalizeSurface(cells[0])
		rows[key] = strings.ToLower(strings.Join(cells[1:], " | "))
	}
	return rows
}

func splitMarkdownTableRow(line string) []string {
	trimmed := strings.Trim(line, "|")
	rawCells := strings.Split(trimmed, "|")
	cells := make([]string, 0, len(rawCells))
	for _, cell := range rawCells {
		cells = append(cells, strings.TrimSpace(cell))
	}
	return cells
}

func normalizeSurface(surface string) string {
	replacer := strings.NewReplacer("`", "", "*", "", "\\", "")
	return strings.ToLower(replacer.Replace(surface))
}

func normalizeWhitespace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func markdownAnchors(text string) map[string]bool {
	anchors := make(map[string]bool)
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		heading := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		if heading == "" {
			continue
		}
		anchors[githubMarkdownAnchor(heading)] = true
	}
	return anchors
}

func githubMarkdownAnchor(heading string) string {
	var builder strings.Builder
	previousHyphen := false
	for _, r := range strings.ToLower(heading) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			previousHyphen = false
		case unicode.IsSpace(r) || r == '-':
			if !previousHyphen && builder.Len() > 0 {
				builder.WriteByte('-')
				previousHyphen = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}
