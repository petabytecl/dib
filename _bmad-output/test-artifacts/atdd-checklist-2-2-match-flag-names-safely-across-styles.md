---
stepsCompleted: ['step-01-preflight-and-context', 'step-02-generation-mode', 'step-03-test-strategy', 'step-04-generate-tests', 'step-04c-aggregate', 'step-05-validate-and-complete']
lastStep: 'step-05-validate-and-complete'
lastSaved: '2026-06-11T18:14:04-04:00'
workflowType: 'testarch-atdd'
storyId: '2.2'
storyKey: '2-2-match-flag-names-safely-across-styles'
storyFile: '_bmad-output/implementation-artifacts/2-2-match-flag-names-safely-across-styles.md'
atddChecklistPath: '_bmad-output/test-artifacts/atdd-checklist-2-2-match-flag-names-safely-across-styles.md'
generatedTestFiles:
  - 'flags/normalize_atdd_test.go'
inputDocuments:
  - '_bmad/tea/config.yaml'
  - '_bmad-output/implementation-artifacts/2-2-match-flag-names-safely-across-styles.md'
  - '_bmad-output/implementation-artifacts/sprint-status.yaml'
  - 'go.mod'
  - 'flags/set.go'
  - 'flags/definition.go'
  - 'flags/errors.go'
  - 'flags/snapshot.go'
  - 'flags/set_atdd_test.go'
  - 'flags/state_atdd_test.go'
  - 'flags/atdd_contract_test.go'
  - 'docs/behavior-matrices.md'
  - 'docs/diagnostics-and-errors.md'
  - '.agents/skills/bmad-testarch-atdd/resources/tea-index.csv'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/data-factories.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/component-tdd.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/test-quality.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/test-healing-patterns.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/test-levels-framework.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/test-priorities-matrix.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/ci-burn-in.md'
---

# ATDD Checklist - Epic 2, Story 2.2: Match Flag Names Safely Across Styles

**Date:** 2026-06-11
**Author:** Coto
**Primary Test Level:** Backend package acceptance tests

## Story Summary

Story 2.2 adds caller-configured long-name normalization to the existing immutable `flags.Set` foundation while preserving exact matching by default. The acceptance scaffolds exercise the expected public `flags` package from a temporary consumer module and keep all scenarios skipped for TDD red-phase activation.

**As a** Go CLI developer
**I want** flag names to match exactly by default and normalize only when I opt in
**So that** familiar naming styles do not create silent collisions or surprising parse behavior.

## Acceptance Criteria

1. Exact names are used when no normalizer is configured; `log-level`, `log_level`, and `log.level` remain distinct.
2. A configured normalizer can resolve equivalent names to the same canonical definition, and snapshot state can be found by canonical definition name.
3. Normalization collisions fail setup with typed deterministic errors that expose both colliding flag names without string matching.
4. Shorthand names remain one-character identities; long-name normalization never creates hidden shorthand aliases.
5. Table-driven tests cover exact matching, configured normalization, collisions, shorthand uniqueness, diagnostic context, and standard-library-only verification.

## Step 1: Preflight And Context

- Detected stack: backend Go module (`go.mod`).
- Story file: `_bmad-output/implementation-artifacts/2-2-match-flag-names-safely-across-styles.md`.
- Story status: `ready-for-dev`.
- Existing test framework: standard Go `testing` package with package tests under `flags/`, `command/`, and tooling tests under `tools/`.
- No `project-context.md` was present.
- No existing Story 2.2 ATDD artifact was present before this run.
- Story 2.1 ATDD scaffolds and implementation tests were used as the local pattern.

## Step 2: Generation Mode

- Selected mode: AI generation.
- Execution mode: sequential.
- Reason: backend Go package work has no UI recording target, HTTP endpoint, service contract, browser journey, or Pact provider. The workflow is adapted to standard-library Go acceptance scaffolds in `flags/`.

## Step 3: Test Strategy

| Scenario | AC | Level | Priority | Red-phase behavior |
| --- | --- | --- | --- | --- |
| Exact matching remains default | AC1, AC5 | Backend package acceptance / regression guard | P1 | Scaffold is skipped; the existing exact-name behavior may already satisfy this guard, but it protects the default while normalization APIs are added |
| Configured normalizer resolves canonical definitions | AC2, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails because `flags.NewNormalizedSet` and `flags.NameNormalizer` do not exist |
| Normalization collisions are typed and inspectable | AC3, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails because `ErrDuplicateNormalizedName` and collision accessors do not exist |
| Normalized derivation preserves immutability | AC2, AC3, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails because `WithNormalizer` and normalized `With` behavior do not exist |
| Long-name normalization does not create shorthand aliases | AC4, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails until long-name and shorthand indexes remain independently validated under normalization |

Integration, API/contract, E2E, and component tests are not appropriate for this story because there is no service boundary, database, HTTP endpoint, browser journey, or UI component.

## Red-Phase Test Scaffolds Created

### Backend Package Tests (5 tests)

**File:** `flags/normalize_atdd_test.go`

Tests:

- `TestATDDExactFlagNamesRemainDistinctByDefault`
  - Status: RED/regression scaffold, skipped with `t.Skip`.
  - Verifies: exact default matching and canonical snapshot lookup for `log-level`, `log_level`, and `log.level`.
- `TestATDDConfiguredNormalizerResolvesCanonicalDefinitions`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: configured normalization maps `log-level`, `log_level`, and `log.level` to the canonical `log-level` definition.
- `TestATDDNormalizationCollisionsAreInspectable`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: normalized name collisions return `ErrDuplicateNormalizedName`, expose `*flags.DefinitionError`, and provide programmatic `Name`, `CollidingName`, and `NormalizedName` context.
- `TestATDDNormalizedDerivationDoesNotMutateOriginalSets`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: `WithNormalizer` returns a new set, original exact lookup remains unchanged, derived normalized sets preserve normalization, and normalized collisions in `With` remain typed.
- `TestATDDLongNameNormalizationDoesNotCreateShorthandAliases`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: `Lookup("l")` does not resolve shorthand `-l`, normalized long-name lookup still works, and duplicate shorthand validation remains `ErrDuplicateShorthand` rather than a normalized long-name collision.

### API Tests

N/A. This story does not expose HTTP or service API endpoints. The workflow's API worker output is represented by backend package acceptance scaffolds.

### E2E Tests

N/A. This story has no browser or user-interface journey.

### Component Tests

N/A. This story has no UI components.

## Data Factories And Fixtures

N/A. The package acceptance tests construct all data in memory with standard Go values. No external service, database, browser, authentication fixture, data-testid, or mock infrastructure is required.

## Temporary Generation Artifacts

- `_bmad-output/test-artifacts/tmp/tea-atdd-api-tests-2026-06-11T18-14-04-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-e2e-tests-2026-06-11T18-14-04-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-summary-2026-06-11T18-14-04-0400.json`

## Implementation Checklist

### Test: `TestATDDExactFlagNamesRemainDistinctByDefault`

**File:** `flags/normalize_atdd_test.go`

- [ ] Preserve `flags.NewSet(defs ...Definition)` as exact-name-by-default behavior.
- [ ] Confirm `log-level`, `log_level`, and `log.level` can coexist and resolve distinctly.
- [ ] Keep default snapshot lookup keyed by canonical registered definition names.
- [ ] Remove only this test's `t.Skip`, confirm current behavior, then keep it green while adding normalization.

### Test: `TestATDDConfiguredNormalizerResolvesCanonicalDefinitions`

**File:** `flags/normalize_atdd_test.go`

- [ ] Add `flags.NameNormalizer` and an explicit normalized-set construction path such as `flags.NewNormalizedSet`.
- [ ] Apply the normalizer to long-name lookup without changing `Definition.Name()`.
- [ ] Resolve raw spellings `log-level`, `log_level`, and `log.level` to the canonical registered definition.
- [ ] Keep snapshot state addressable by canonical definition name.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDNormalizationCollisionsAreInspectable`

**File:** `flags/normalize_atdd_test.go`

- [ ] Add `flags.ErrDuplicateNormalizedName`.
- [ ] Detect normalized long-name collisions during set construction.
- [ ] Extend `*flags.DefinitionError` or equivalent typed context with both colliding long flag names and the normalized key.
- [ ] Assert diagnostics through `errors.Is` / `errors.As`, not string matching.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDNormalizedDerivationDoesNotMutateOriginalSets`

**File:** `flags/normalize_atdd_test.go`

- [ ] Add immutable normalizer derivation such as `Set.WithNormalizer`.
- [ ] Preserve the original exact set's observable behavior.
- [ ] Preserve the configured normalizer when deriving with `Set.With`.
- [ ] Validate normalized collisions when deriving with new definitions.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDLongNameNormalizationDoesNotCreateShorthandAliases`

**File:** `flags/normalize_atdd_test.go`

- [ ] Keep long-name lookup and shorthand identity separate.
- [ ] Ensure `Lookup("l")` does not resolve a shorthand-only alias.
- [ ] Keep duplicate shorthand validation under `ErrDuplicateShorthand`.
- [ ] Prevent duplicate shorthand failures from being reported as normalized long-name collisions.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

## Running Tests

```bash
# Current red-phase scaffold state; tests are present but skipped.
go test ./flags

# Full repository check while scaffolds remain skipped.
go test ./...

# Activate one scaffold at a time by removing that test's t.Skip, then run:
go test ./flags -run TestATDDExactFlagNamesRemainDistinctByDefault -count=1
go test ./flags -run TestATDDConfiguredNormalizerResolvesCanonicalDefinitions -count=1
go test ./flags -run TestATDDNormalizationCollisionsAreInspectable -count=1
go test ./flags -run TestATDDNormalizedDerivationDoesNotMutateOriginalSets -count=1
go test ./flags -run TestATDDLongNameNormalizationDoesNotCreateShorthandAliases -count=1

# Final Story 2.2 verification after implementation:
go test ./...
go vet ./...
go run ./tools/depgate
git diff --check
```

## Red-Green-Refactor Workflow

### RED Phase (Current)

- Acceptance scaffolds exist and are skipped with `t.Skip`.
- Each skipped test runs a temporary consumer module against `github.com/petabytecl/dib/flags`.
- Removing one skip before implementation should fail with compile errors for missing normalization APIs or with behavior assertions for incomplete implementation.
- `TestATDDExactFlagNamesRemainDistinctByDefault` is primarily a regression guard for existing exact-name behavior; keep it green while adding opt-in normalization.
- Tests use only Go standard-library packages and preserve the dependency gate.

### GREEN Phase (DEV Workflow)

1. Pick one scaffold, starting with `TestATDDConfiguredNormalizerResolvesCanonicalDefinitions` after confirming the exact-name guard.
2. Remove only that test's `t.Skip`.
3. Run the narrow `go test ./flags -run ... -count=1` command and confirm RED.
4. Implement the smallest production change to pass the activated test.
5. Repeat for the remaining scaffolds.

### REFACTOR Phase

- Run `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.
- Confirm root `go.mod` still contains no `require`, `replace`, or `toolchain` directives.
- Confirm no `go.sum`, `/cmd`, `config/`, parser fuzz seed, full long-flag parser, shorthand parser, or broad root facade was introduced unless implementation proves it necessary and the architecture/story is deliberately updated.

## Knowledge Base References Applied

- `test-levels-framework.md`: selected package acceptance/unit tests instead of E2E for pure backend library behavior.
- `test-priorities-matrix.md`: classified normalization API, collision diagnostics, derivation, and shorthand independence as P0 because later Epic 2 parser/config stories depend on stable canonical names.
- `test-quality.md`: kept tests deterministic, explicit, isolated, and under 300 lines.
- `test-healing-patterns.md`: avoided hard waits, dynamic data, hidden assertions, external services, and network dependencies.
- `ci-burn-in.md`: preserved static, standard-library-only verification and command-level handoff.
- `data-factories.md` and `component-tdd.md`: loaded for workflow completeness; UI/factory-specific patterns were not directly applicable to this package-only Go story.

## Test Execution Evidence

Current scaffold verification:

```bash
go test ./flags
go test ./...
```

Both commands pass while scaffolds remain skipped. Activated RED verification is intentionally left to the dev workflow one scaffold at a time by removing the relevant `t.Skip`.

## Step 5: Validation And Completion

- Prerequisites satisfied: Story 2.2 has clear acceptance criteria, backend Go test framework exists, and development environment is available.
- Test file created correctly: `flags/normalize_atdd_test.go`.
- Red-phase compliance: five scaffold tests are skipped with `t.Skip`; no placeholder assertions such as `expect(true).toBe(true)` are present.
- Story metadata and handoff paths are captured in this checklist and linked back into the story file.
- CLI/browser session cleanup: N/A, no Playwright CLI, browser, or MCP session was opened.
- Temp artifacts are stored under `_bmad-output/test-artifacts/tmp/`.
- Validation commands passed: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.

## Notes

- The ATDD scaffolds intentionally define the expected public API contract for Story 2.2. If implementation discovers a better API shape, update the scaffold and checklist deliberately rather than bypassing the tests.
- The helper in `flags/atdd_contract_test.go` uses temporary consumer modules so public API expectations are tested from outside the `flags` package.
- Keep all runtime and test imports standard-library-only.
