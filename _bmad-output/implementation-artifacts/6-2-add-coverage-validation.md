---
baseline_commit: 9209b31
created: "2026-06-12"
---

# Story 6.2: Add Coverage Validation

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a release reviewer,
I want package-aware coverage validation in CI,
so that Dib release candidates prove meaningful test coverage instead of only proving that tests execute.

## Requirements Trace

- FR24: Maintainers can validate release-candidate coverage with a documented, package-aware threshold policy.
- FR20: Behavior test matrices include coverage validation evidence.
- NFR1: Runtime packages must import only the Go standard library; the coverage tool itself must remain stdlib-only.
- NFR11: Coverage gates must be deterministic enough to run locally and in CI with the same documented commands.
- Architecture: "Coverage validation must use standard Go coverage output where practical and apply package-aware thresholds." [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- Architecture: "Public runtime packages (`command`, `config`, and `flags`) are release-surface packages and must report threshold evidence separately from tooling packages." [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- Architecture: "Exact coverage threshold values and any tooling-package exception policy must be finalized by Story 6.2." [Source: `_bmad-output/planning-artifacts/architecture.md#Gap-Analysis-Results`]

## Acceptance Criteria

1. Given public runtime packages are the release surface, when coverage validation runs, then `command`, `config`, and `flags` report package-level coverage, and each public runtime package fails below the approved threshold.
2. Given tooling packages have different risk profiles, when coverage policy is documented, then tooling packages either have a separate threshold or a documented exception, and any exception names the critical-path tests that preserve confidence.
3. Given maintainers need reproducible evidence, when CI runs, then coverage is generated from standard Go coverage output where practical, and `docs/release-checklist.md` records the coverage command, package results, thresholds, and any accepted exception.

## Tasks / Subtasks

- [x] Confirm preconditions and read UPDATE files before editing (AC: 1-3)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 6 `in-progress` and Story 6.2 `ready-for-dev`.
  - [x] Confirm current `HEAD` is `9209b31` (`feat(story-6.1): Add an Isolated Linter Gate`) or account for newer user changes.
  - [x] **OBSERVE CURRENT COVERAGE BEFORE WRITING CODE** — run `GOCACHE=/tmp/dib-go-build go test -cover ./command ./config ./flags` and record the per-package percentages. This informs the threshold policy for step 2.
  - [x] Read these UPDATE files completely before editing: `.github/workflows/ci.yml`, `tools/cigate/ci_test.go`, `docs/release-checklist.md`, `docs/release_checklist_test.go`, `docs/testing.md`, `docs/testing_test.go`, `docs/behavior-matrices.md`, `docs/behavior_matrices_test.go`, `docs/release-notes-v0.md`, `go.mod`, `tools/lint/main.go` (as implementation reference).
  - [x] Do not add runtime APIs, package-global helpers, root facade packages, lint gate changes owned by Story 6.1, public usage documentation owned by Story 6.3, or tracker reconciliation owned by Story 6.4.

- [x] Select and document the coverage mechanism and threshold policy (AC: 1, 2)
  - [x] From the observed per-package coverage in the precondition check, determine per-package thresholds:
    - Set each threshold to the observed value floored to the nearest 5%, minimum 80%.
    - Example: observed 92.3% → threshold 90%; observed 84.7% → threshold 80%.
  - [x] Document the coverage gate in `docs/testing.md` under a new `## Coverage Gate` section:
    - Explain the approach (stdlib-only `tools/coverage` tool, consistent with `tools/lint` precedent).
    - Record the local command: `GOCACHE=/tmp/dib-go-build go run ./tools/coverage`.
    - Record the CI command: `go run ./tools/coverage`.
    - State why the stdlib-only path was preferred (no external tool, no root dependency, consistent with project isolation policy).
  - [x] Document the tooling package exception in `docs/testing.md`:
    - `tools/depgate`: exception granted; critical-path tests `TestDepgateFixtures`, `TestDepgateReportsEveryViolationDeterministically`, and `TestDepgateDisablesWorkspaceMode` in `tools/depgate/main_test.go` preserve confidence.
    - `tools/lint`: exception granted; critical-path tests `TestLintPassesCleanFormattedGoFiles`, `TestLintReportsUnformattedFilesDeterministically`, and `TestLintCommandRunsFromRepositoryRoot` in `tools/lint/main_test.go` preserve confidence.
    - `tools/coverage`: exception granted; critical-path tests `TestCoveragePassesPackagesMeetingThreshold`, `TestCoverageFailsPackagesBelowThreshold`, and `TestCoverageCommandRunsFromRepositoryRoot` in `tools/coverage/main_test.go` preserve confidence.
  - [x] Add tests to `docs/testing_test.go` that guard the new `## Coverage Gate` section and coverage-specific guidance, following the same pattern as `TestTestingGuideDocumentsLintGateIsolationAndPinning`.

- [x] Implement `tools/coverage` as a stdlib-only program (AC: 1, 2, 3)
  - [x] Create `tools/coverage/main.go` using ONLY these standard library packages: `fmt`, `io`, `os`, `os/exec`, `regexp`, `strconv`, `strings`.
  - [x] Entry point for testability: `func run(stdout, stderr io.Writer) int` — called by `main()` with `os.Stdout` and `os.Stderr`.
  - [x] Per-package thresholds are defined as a package-level slice of structs (set from the values determined in step 2):
    ```
    type threshold struct { pkg string; minPct float64 }
    ```
  - [x] The tool runs `go test -cover ./command ./config ./flags` (all three packages in one invocation) from the repository root, parses coverage lines of the form `ok  <pkg>  <time>  coverage: <X.X>% of statements`, and checks each against its threshold.
  - [x] Exit codes: 0 = all packages pass, 1 = one or more packages below threshold, 2 = execution failure (tool could not run `go test` or parse its output).
  - [x] Output format (to stdout) for each package: `coverage: <pkg>: <X.X>% (threshold <min>%) PASS` or `FAIL`.
  - [x] Execution errors go to stderr: `coverage execution error: <detail>`.
  - [x] Create `tools/coverage/main_test.go` covering:
    - `TestCoveragePassesPackagesMeetingThreshold`: all packages at or above threshold returns exit 0.
    - `TestCoverageFailsPackagesBelowThreshold`: a package below threshold returns exit 1.
    - `TestCoverageSeparatesExecutionFailuresFromThresholdViolations`: invocation failure returns exit 2 with stderr message; no stdout written.
    - `TestCoverageReportsDeterministicOutput`: output lines appear in threshold slice order.
    - `TestCoverageCommandRunsFromRepositoryRoot`: integration test that runs `exec.Command("go", "run", "./tools/coverage")` from repo root with a GOCACHE temp dir; asserts exit 0 and verifies the three public packages appear in output.
  - [x] The tool relies on coverage output emitted to stdout by `go test`. When running all three packages in one command, `go test -cover ./command ./config ./flags` emits one `ok ...` line per package with coverage percentage. Use `regexp.MustCompile(`coverage: (\d+\.\d+)%`)` to extract percentages.
  - [x] Verify `go run ./tools/depgate` still passes after adding `tools/coverage`. The `os/exec` import is stdlib; no external imports are permitted.

- [x] Wire the coverage gate into CI (AC: 2, 3)
  - [x] Update `.github/workflows/ci.yml` to add a `Coverage` step **after `Vet`, before `Dependency gate`**:
    ```yaml
    - name: Coverage
      run: go run ./tools/coverage
    ```
  - [x] Keep all existing steps: Checkout, Set up Go, Show Go version, Lint, Test, Vet, **Coverage** (new), Dependency gate.
  - [x] Update `tools/cigate/ci_test.go`:
    - Add `"coverage gate": "run: go run ./tools/coverage"` to `TestCIWorkflowRunsTrustGates`.
    - Add a new `TestCIWorkflowRunsCoverageAfterVetAndBeforeDependencyGate` test using `assertOrderedMarkers` to enforce the step order: `"run: go vet ./..."`, `"run: go run ./tools/coverage"`, `"run: go run ./tools/depgate"`.
    - Add `"coverage evidence": "go run ./tools/coverage"` to `TestReleaseChecklistCapturesRequiredEvidence`.
    - Do NOT change `TestCIWorkflowUsesOnlyTrustedStaticSteps` — the coverage step uses `go run ./tools/coverage`, not a new action, so no action allowlist update is needed.

- [x] Record coverage evidence in release docs without claiming final release readiness (AC: 3)
  - [x] Update `docs/release-checklist.md`:
    - Under **CI Trust Gates**, add: `- \`go run ./tools/coverage\`: PASS on 2026-06-12 with \`GOCACHE=/tmp/dib-go-build go run ./tools/coverage\`; per-package results recorded below.`
    - Under **CI Trust Gates** or a new **Coverage Validation Evidence** subsection, record each package result:
      - `command`: observed 85.2%, threshold 85%
      - `config`: observed 89.6%, threshold 85%
      - `flags`: observed 85.0%, threshold 85%
    - Under **Standard-Library Dependency Evidence**, add:
      - `Coverage gate reviewed:` confirming `tools/coverage` imports only the Go standard library.
      - `Coverage isolation evidence:` confirming `tools/coverage` uses `os/exec` (stdlib) to invoke `go test -cover`, no external coverage tool package enters the module graph.
    - Update `Story 6.1 evidence scope:` line to add a `Story 6.2 evidence scope:` line: coverage evidence was collected from the Story 6.2 working tree; final tag commit reconciliation remains a later release-review step.
    - Do NOT claim release readiness or approve the tag.
  - [x] Update `docs/release_checklist_test.go`:
    - In `TestReleaseChecklistRecordsReleaseCandidateEvidence`, add to the required phrases: `"story 6.2 evidence scope:"`, `"go run ./tools/coverage"`.
    - In `TestReleaseChecklistRecordsReleaseCandidateEvidence`, add fields: `"Coverage gate reviewed"`, `"Coverage isolation evidence"`.
    - In `TestReleaseChecklistRecordsPassingRequiredGates`, add `"go run ./tools/coverage"` to the commands list.
  - [x] Update `docs/behavior-matrices.md`:
    - In the **Dependency gate evidence** row, add `Story 6.2` to story coverage, `FR24` to FR/NFR trace, and `go run ./tools/coverage` to the executable evidence column.
    - The row should end with: `... ; \`go run ./tools/depgate\`; \`go run ./tools/lint\`; \`go run ./tools/coverage\`; \`docs/release-checklist.md\` | current`
  - [x] Update `docs/behavior_matrices_test.go` in `TestBehaviorMatricesCoverAdoptionEvidenceRows`:
    - In the `"dependency gate evidence"` entry, add `"story 6.2"`, `"fr24"`, and `"go run ./tools/coverage"` to the required phrases list.
  - [x] Update `docs/release-notes-v0.md`:
    - In the Release Gates list, add `- \`go run ./tools/coverage\`` after `go run ./tools/lint`.
  - [x] Update `docs/testing_test.go` to add a new test `TestTestingGuideDocumentsCoverageGate` that checks for the `## Coverage Gate` section, the local command `GOCACHE=/tmp/dib-go-build go run ./tools/coverage`, the tool path `tools/coverage`, mentions of per-package thresholds for `command`, `config`, and `flags`, and the tooling exception list.

- [x] Verify the coverage gate and standard-library isolation (AC: 1-3)
  - [x] `GOCACHE=/tmp/dib-go-build go run ./tools/coverage` — record the exact per-package results in `docs/release-checklist.md`.
  - [x] `GOCACHE=/tmp/dib-go-build go test ./tools/coverage`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./tools/cigate`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./docs`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go vet ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
  - [x] `git diff --check`
  - [x] `rg -n "^(require|replace|toolchain)\\b" go.mod` returns no output.
  - [x] `test ! -e go.sum` remains true.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`; Story 6.2 is `backlog`, Epic 6 is `in-progress`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`; Story 6.2 spec at `##-Story-6.2:-Add-Coverage-Validation`. [Source: `_bmad-output/planning-artifacts/epics.md`]
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`; coverage gate is a core CI gate owned by Story 6.2 (see `Infrastructure-Deployment` and `Gap-Analysis-Results` sections). [Source: `_bmad-output/planning-artifacts/architecture.md`]
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`; FR24 requires package-aware coverage validation with a documented threshold policy. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`]
- Loaded Story 6.1: `_bmad-output/implementation-artifacts/6-1-add-an-isolated-linter-gate.md`; Story 6.1 established `tools/lint` as the reference pattern for stdlib-only CI gate tools. Key learnings: `run(ctx, dir, stdout, stderr)` testable entry point, exit 1/2 separation, `TestLintCommandRunsFromRepositoryRoot` integration test. [Source: `_bmad-output/implementation-artifacts/6-1-add-an-isolated-linter-gate.md`]
- No UX artifact; Dib V1 has no browser UI.

### Current Repository State (as of baseline commit 9209b31)

- Root `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`; no root `go.sum`.
- `.github/workflows/ci.yml` has steps: Checkout, Set up Go, Show Go version, Lint (`go run ./tools/lint`), Test (`go test ./...`), Vet (`go vet ./...`), Dependency gate (`go run ./tools/depgate`). No coverage step today.
- `docs/testing.md` has `## Lint Gate` and `## Release Candidate Gates` sections but no `## Coverage Gate` section.
- `docs/release-checklist.md` records lint, test, vet, dependency-gate, race, fuzz, and provenance evidence. It does not record coverage evidence.
- `docs/release-notes-v0.md` lists `go test ./...`, `go run ./tools/lint`, `go vet ./...`, `go run ./tools/depgate`, and `go test -race ./...` as release gates; adding `go run ./tools/coverage` prevents it becoming stale.
- `tools/cigate/ci_test.go:TestCIWorkflowUsesOnlyTrustedStaticSteps` currently denies all actions except `actions/checkout@v6` and `actions/setup-go@v6`. The coverage step uses `go run ./tools/coverage` (not a new action), so this test does NOT need updating.
- `docs/behavior_matrices_test.go:TestBehaviorMatricesCoverAdoptionEvidenceRows` expects `"story 6.1"` and `"fr23"` in the dependency gate evidence row; Story 6.2 must add `"story 6.2"`, `"fr24"`, and `"go run ./tools/coverage"` to that same row.
- `tools/lint/main.go` is the reference implementation: stdlib-only, `run(ctx, dir, stdout, stderr)` entry point, `lintViolationExit=1`, `executionFailureExit=2`.

### Current UPDATE File Intelligence

- `.github/workflows/ci.yml`: add Coverage step after Vet, before Dependency gate. Keep all existing steps intact.
- `tools/cigate/ci_test.go`: add coverage gate to `TestCIWorkflowRunsTrustGates`; add new ordering test; add coverage marker to `TestReleaseChecklistCapturesRequiredEvidence`. No action allowlist changes needed.
- `docs/testing.md`: add `## Coverage Gate` section after `## Lint Gate` documenting the approach, local command, CI command, and tooling exception list.
- `docs/testing_test.go`: add `TestTestingGuideDocumentsCoverageGate` to guard the new coverage section.
- `docs/release-checklist.md`: add coverage command result, per-package results, thresholds, and isolation evidence. Do not approve the tag.
- `docs/release_checklist_test.go`: require coverage command result, story 6.2 evidence scope marker, coverage gate reviewed, and coverage isolation evidence fields.
- `docs/behavior-matrices.md`: update the `dependency gate evidence` row to include `story 6.2`, `FR24`, and `go run ./tools/coverage`.
- `docs/behavior_matrices_test.go`: update `"dependency gate evidence"` entry in `TestBehaviorMatricesCoverAdoptionEvidenceRows` to require `"story 6.2"`, `"fr24"`, and `"go run ./tools/coverage"`.
- `docs/release-notes-v0.md`: add `go run ./tools/coverage` to the Release Gates list.

### Git Intelligence

- Recent commits:
  - `9209b31 feat(story-6.1): Add an Isolated Linter Gate`
  - `7cfdfca docs(bmad): add epic 6 release hardening plan`
  - `ee6f2b8 chore(bmad): add automator and creative skills`
  - `561dd78 docs: add epic 1 retrospective`
  - `6dba659 docs: add epic 5 retrospective`
- Relevant patterns:
  - Story 6.1 introduced `tools/lint` as a stdlib-only tool with `run(ctx, dir, stdout, stderr)` entry point. Mirror this pattern exactly for `tools/coverage`.
  - Story 6.1 added lint evidence to `docs/release-checklist.md` by extending existing sections rather than restructuring the file. Follow the same extension pattern for coverage evidence.
  - Epic 5 established docs/test evidence as first-class release artifacts; extend tests rather than relying on prose.
  - Existing CI tests prefer static string checks over YAML parsing. Match that pattern.

### Architecture Guardrails

- Coverage tool must be stdlib-only. `os/exec` is part of the Go standard library; it may be used in `tools/coverage/main.go`. Do NOT import external coverage or reporting packages. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- `go run ./tools/depgate` remains the required dependency gate. Do not invent a second dependency scanner. [Source: `_bmad-output/planning-artifacts/architecture.md#Implementation-Handoff`]
- Story 6.2 owns threshold policy. Architecture defers exact threshold values to this story. Once set, thresholds must be recorded in `docs/release-checklist.md` as evidence. [Source: `_bmad-output/planning-artifacts/architecture.md#Gap-Analysis-Results`]
- CI failures block tagging. Waivers require owner, reason, expiry, and impact. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- Dib must not claim source compatibility with Go `flag`, pflag, Cobra, or Viper. Preserve native-API language if release docs are touched. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`]
- `docs/testing.md` owns local verification, lint, coverage, fuzz, race, dependency-gate, and release-candidate validation guidance. [Source: `_bmad-output/planning-artifacts/architecture.md#Documentation-Organization`]
- Do not mark the release ready to tag. Stories 6.3 (public docs) and 6.4 (reconciliation) remain open.

### Library / Framework Requirements

- Go baseline is `go 1.26`. Keep root `go.mod` at `go 1.26`; do not add a `toolchain` directive.
- `os/exec.Command("go", "test", "-cover", "./command", "./config", "./flags")` — this is the canonical stdlib invocation.
- Coverage output format from `go test -cover`: `ok  github.com/petabytecl/dib/command  0.002s  coverage: 85.2% of statements`
- Parse coverage percentage with: `regexp.MustCompile(`coverage: (\d+\.\d+)%`)`
- When a package has no tests or only test-build errors, `go test -cover` may output `FAIL` instead of a coverage line. The tool must treat the inability to extract a coverage percentage as an execution failure (exit 2), not a threshold violation (exit 1).

### Project Structure Notes

- NEW targets: `tools/coverage/main.go`, `tools/coverage/main_test.go`.
- UPDATE targets: `.github/workflows/ci.yml`, `tools/cigate/ci_test.go`, `docs/testing.md`, `docs/testing_test.go`, `docs/release-checklist.md`, `docs/release_checklist_test.go`, `docs/behavior-matrices.md`, `docs/behavior_matrices_test.go`, `docs/release-notes-v0.md`.
- The `tools/coverage` program must be placed under `tools/coverage/` (consistent with `tools/lint/` and `tools/depgate/` layout).
- Do NOT touch `command/`, `flags/`, `config/`, or `examples/` unless coverage reveals a real defect. Coverage changes are instrumentation-only.
- Do NOT add root `require`, `replace`, `toolchain`, or `go.sum`.

### Anti-Patterns To Avoid

- Do not use an external coverage tool (`gocov`, `gocover-cobertura`, etc.) as a Go import or binary dependency.
- Do not add a second dependency gate or weaken `tools/depgate`.
- Do not record coverage PASS evidence unless the exact coverage command actually ran.
- Do not set thresholds without first running `go test -cover ./command ./config ./flags` to observe current values.
- Do not set thresholds below 80% for public runtime packages.
- Do not fold lint work back into this story (owned by 6.1), public docs (owned by 6.3), or reconciliation (owned by 6.4).
- Do not mark the release ready to tag just because coverage passes.
- Do not change `TestCIWorkflowUsesOnlyTrustedStaticSteps` — no new action is added; only a `go run` step.
- Do not restructure `docs/release-checklist.md` heading hierarchy — extend existing sections.

### Validation Checklist Applied

- Story includes exact story ID/key, ready-for-dev status, role/action/benefit, BDD-derived acceptance criteria, and task mapping to ACs.
- Story identifies every likely UPDATE file and summarizes current behavior to preserve.
- Story includes previous story intelligence from Story 6.1 (tools/lint reference pattern).
- Story includes git intelligence from the most recent commits.
- Story preserves dependency, clean-room, release-evidence, and CI guardrails.
- Story gives concrete guidance on threshold policy (observe first, floor to nearest 5%, minimum 80%).

### Story Completion Status

- Status set to `ready-for-dev`.
- Completion note: Ultimate context engine analysis completed — comprehensive developer guide created.

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

None.

### Completion Notes List

- Observed pre-implementation coverage: command 85.2%, config 89.6%, flags 85.0%. All thresholds set to 85% (floor to nearest 5%, minimum 80%).
- Created `tools/coverage/main.go` — stdlib-only (fmt, io, os, os/exec, regexp, strconv, strings). Internal `check()` function decouples parsing logic from subprocess execution for unit testability.
- Created `tools/coverage/main_test.go` — 5 tests: 4 unit tests using fake go test output via `check()`, 1 integration test running real `go run ./tools/coverage`.
- Added `## Coverage Gate` section to `docs/testing.md` documenting approach, local/CI commands, per-package thresholds, and tooling exception list for depgate, lint, and coverage.
- Added `TestTestingGuideDocumentsCoverageGate` to `docs/testing_test.go`.
- Updated `.github/workflows/ci.yml`: Coverage step inserted after Vet, before Dependency gate.
- Updated `tools/cigate/ci_test.go`: added coverage gate to trust gates test, new ordering test, coverage evidence to checklist test.
- Updated `docs/release-checklist.md`: coverage PASS evidence, per-package results, story 6.2 evidence scope line, Coverage gate reviewed and isolation evidence fields.
- Updated `docs/release_checklist_test.go`: added story 6.2 scope, coverage command, Coverage gate/isolation fields, and coverage to passing gates test.
- Updated `docs/behavior-matrices.md`: dependency gate evidence row extended with Story 6.2, FR24, `go run ./tools/coverage`.
- Updated `docs/behavior_matrices_test.go`: added story 6.2, fr24, go run coverage to required phrases.
- Updated `docs/release-notes-v0.md`: added `go run ./tools/coverage` to Release Gates list.
- All tests pass: `go test ./...` — 9 packages OK. No regressions. Depgate confirms stdlib-only. Lint confirms formatting. go.sum absent.

### File List

- tools/coverage/main.go (new)
- tools/coverage/main_test.go (new)
- .github/workflows/ci.yml (modified)
- tools/cigate/ci_test.go (modified)
- docs/testing.md (modified)
- docs/testing_test.go (modified)
- docs/release-checklist.md (modified)
- docs/release_checklist_test.go (modified)
- docs/behavior-matrices.md (modified)
- docs/behavior_matrices_test.go (modified)
- docs/release-notes-v0.md (modified)
- _bmad-output/implementation-artifacts/sprint-status.yaml (modified)
- _bmad-output/implementation-artifacts/tests/test-summary.md (modified)
- _bmad-output/implementation-artifacts/6-2-add-coverage-validation.md (modified)

## Change Log

- 2026-06-12: Implemented Story 6.2 — added stdlib-only `tools/coverage` program with per-package thresholds (85% for command, config, flags), wired Coverage step into CI after Vet, recorded coverage evidence in release-checklist.md, extended behavior-matrices.md and release-notes-v0.md. All tests pass, depgate confirms zero external imports.
- 2026-06-12: Code review (AI) — 0 CRITICAL, 0 HIGH, 2 MEDIUM, 1 LOW issues found and auto-fixed. Fixed: behavior-matrices.md "Dependency And Release Evidence" prose stale (missing coverage); testing.md "Release Candidate Gates" missing coverage command; story File List missing test-summary.md. Status → done.
