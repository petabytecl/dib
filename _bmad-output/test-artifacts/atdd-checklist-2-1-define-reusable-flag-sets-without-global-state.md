---
stepsCompleted: ['step-01-preflight-and-context', 'step-02-generation-mode', 'step-03-test-strategy', 'step-04-generate-tests', 'step-04c-aggregate', 'step-05-validate-and-complete']
lastStep: 'step-05-validate-and-complete'
lastSaved: '2026-06-11T16:36:04-04:00'
workflowType: 'testarch-atdd'
storyId: '2.1'
storyKey: '2-1-define-reusable-flag-sets-without-global-state'
storyFile: '_bmad-output/implementation-artifacts/2-1-define-reusable-flag-sets-without-global-state.md'
atddChecklistPath: '_bmad-output/test-artifacts/atdd-checklist-2-1-define-reusable-flag-sets-without-global-state.md'
generatedTestFiles:
  - 'flags/atdd_contract_test.go'
  - 'flags/set_atdd_test.go'
  - 'flags/state_atdd_test.go'
inputDocuments:
  - '_bmad/tea/config.yaml'
  - '_bmad-output/implementation-artifacts/2-1-define-reusable-flag-sets-without-global-state.md'
  - '_bmad-output/implementation-artifacts/sprint-status.yaml'
  - 'go.mod'
  - 'flags/doc.go'
  - 'command/definition_test.go'
  - 'command/contract_test.go'
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
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/overview.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/api-request.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/auth-session.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/recurse.md'
---

# ATDD Checklist - Epic 2, Story 2.1: Define Reusable Flag Sets Without Global State

**Date:** 2026-06-11
**Author:** Coto
**Primary Test Level:** Backend package acceptance tests

## Story Summary

Story 2.1 creates the foundation for reusable `flags` definitions, independent Flag sets, machine-readable value state, typed diagnostics, and caller-observable immutability. The ATDD scaffolds are skipped Go tests that exercise the expected public `flags` package from a temporary consumer module.

**As a** Go CLI developer
**I want** reusable Flag sets with explicit definitions and value metadata
**So that** I can parse CLI input without package-global mutable state or hidden process dependencies.

## Acceptance Criteria

1. Definitions capture long name, optional shorthand, default value, usage text, value parser, repeat policy, hidden/deprecated metadata, sensitivity metadata, and no-option default where applicable; built-ins include string, bool, int, int64, uint, uint64, float64, duration, and string list.
2. Two independent Flag sets can use the same names while definitions and snapshots remain independent; no package-level registry, default Flag set, or ambient `os.Args` dependency is introduced.
3. Value arity, default handling, explicit-set tracking, duplicate detection, diagnostic categories, and public error inspection are machine readable through `errors.Is` or `errors.As`.
4. Derived or extended Flag sets leave original behavior unchanged and do not leak caller-owned slice/map aliases across definitions or snapshots.
5. Table-driven tests cover validation, duplicate names, duplicate shorthands, explicit-set tracking, default values, and reusable definitions; `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` pass.

## Step 1: Preflight And Context

- Detected stack: backend Go module (`go.mod`).
- Story file: `_bmad-output/implementation-artifacts/2-1-define-reusable-flag-sets-without-global-state.md`.
- Story status: `ready-for-dev`.
- Existing test framework: standard Go `testing` package with package tests under `command/` and tooling tests under `tools/`.
- No `project-context.md` was present.
- No existing Story 2.1 ATDD artifact was present before this run.
- `flags/` contained package documentation only before ATDD generation.

## Step 2: Generation Mode

- Selected mode: AI generation.
- Execution mode: sequential.
- Reason: backend Go package work has no UI recording target and no HTTP API contract. The workflow was adapted to standard-library Go acceptance scaffolds in `flags/`.

## Step 3: Test Strategy

| Scenario | AC | Level | Priority | Red-phase behavior |
| --- | --- | --- | --- | --- |
| Definition metadata and built-in value kinds | AC1 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails because expected `flags` constructors and metadata APIs do not exist |
| Setup validation and inspectable errors | AC3, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails because typed validation errors do not exist |
| Independent sets and explicit snapshots | AC2 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails because explicit Flag set construction and snapshot APIs do not exist |
| Immutable derivation and no alias leaks | AC4, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails because immutable derivation and defensive-copy APIs do not exist |
| Value arity, explicit tracking, conversion wrapping, and redaction | AC1, AC3, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails because value-state and diagnostic APIs do not exist |

Integration, API/contract, E2E, and component tests are not appropriate for this story because there is no service boundary, database, HTTP endpoint, browser journey, or UI component.

## Red-Phase Test Scaffolds Created

### Backend Package Tests (5 tests)

**Files:**

- `flags/atdd_contract_test.go` (49 lines)
- `flags/set_atdd_test.go` (191 lines)
- `flags/state_atdd_test.go` (192 lines)

Tests:

- `TestATDDFlagDefinitionsExposeMetadata`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: public constructors, long/shorthand names, built-in kinds, default values, usage text, hidden/deprecated/sensitive metadata, no-option defaults, repeat policy, and deterministic inspection.
- `TestATDDFlagSetValidationErrorsAreInspectable`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: invalid names, duplicate names, duplicate shorthands, invalid shorthands, invalid no-option defaults, `errors.Is`, and `errors.As` over `*flags.DefinitionError`.
- `TestATDDIndependentFlagSetsIgnoreAmbientProcessState`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: two Flag sets with identical names retain separate defaults and do not read `os.Args` or environment variables through default snapshots.
- `TestATDDDerivedFlagSetsDoNotMutateOriginalsOrLeakAliases`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: `With` returns a derived set, originals stay unchanged, caller-owned defaults are copied, returned definitions are copies, and snapshot values do not alias mutable storage.
- `TestATDDValueAndDiagnosticFoundationIsMachineReadable`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: custom parser wrapping, `ErrConversion`, `*flags.ValueError`, value arity, sensitivity metadata, explicit-set tracking, and sensitive-value redaction.

### API Tests

N/A. This story does not expose HTTP or service API endpoints. The workflow's API worker output is represented by backend package acceptance scaffolds.

### E2E Tests

N/A. This story has no browser or user-interface journey.

### Component Tests

N/A. This story has no UI components.

## Data Factories And Fixtures

N/A. The package acceptance tests construct all data in memory with standard Go values. No external service, database, browser, authentication fixture, data-testid, or mock infrastructure is required.

## Temporary Generation Artifacts

- `_bmad-output/test-artifacts/tmp/tea-atdd-api-tests-2026-06-11T16-30-15-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-e2e-tests-2026-06-11T16-30-15-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-summary-2026-06-11T16-30-15-0400.json`

## Implementation Checklist

### Test: `TestATDDFlagDefinitionsExposeMetadata`

**File:** `flags/set_atdd_test.go`

- [ ] Add public `flags.NewSet` and built-in definition constructors for string, bool, int, int64, uint, uint64, float64, duration, and string list.
- [ ] Add option helpers for shorthand, repeatable values, hidden, deprecated, sensitive, and no-option default metadata.
- [ ] Add deterministic inspection methods for set length, definition lookup, definition list, kind, default, usage, shorthand, repeat policy, no-option default, hidden, deprecated, and sensitive metadata.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDFlagSetValidationErrorsAreInspectable`

**File:** `flags/set_atdd_test.go`

- [ ] Validate empty/invalid long names.
- [ ] Validate duplicate long names within one Flag set.
- [ ] Validate duplicate and invalid shorthand values.
- [ ] Validate no-option default compatibility with the value kind.
- [ ] Expose sentinel errors and `*flags.DefinitionError` for caller inspection.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDIndependentFlagSetsIgnoreAmbientProcessState`

**File:** `flags/state_atdd_test.go`

- [ ] Add default snapshot/value-state APIs for definition defaults.
- [ ] Keep snapshots independent between Flag sets with identical names.
- [ ] Do not read `os.Args`, env vars, package globals, default Flag sets, stdout, stderr, or hidden registries.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDDerivedFlagSetsDoNotMutateOriginalsOrLeakAliases`

**File:** `flags/state_atdd_test.go`

- [ ] Add immutable derivation with `Set.With`.
- [ ] Defensively copy caller-owned slice defaults.
- [ ] Return copied definition slices from inspection APIs.
- [ ] Return copied snapshot values so callers cannot mutate stored state through returned slices.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDValueAndDiagnosticFoundationIsMachineReadable`

**File:** `flags/state_atdd_test.go`

- [ ] Add custom parser support through a small function/interface type.
- [ ] Expose arity and explicit-set state in definitions and default snapshots.
- [ ] Wrap conversion/parser failures so `errors.Is` reaches both Dib sentinels and caller parser errors.
- [ ] Expose `*flags.ValueError` with name and kind context.
- [ ] Redact sensitive raw values from public error strings.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

## Running Tests

```bash
# Current red-phase scaffold state; tests are present but skipped.
go test ./flags

# Full repository check while scaffolds remain skipped.
go test ./...

# Activate one scaffold at a time by removing that test's t.Skip, then run:
go test ./flags -run TestATDDFlagDefinitionsExposeMetadata
go test ./flags -run TestATDDFlagSetValidationErrorsAreInspectable
go test ./flags -run TestATDDIndependentFlagSetsIgnoreAmbientProcessState
go test ./flags -run TestATDDDerivedFlagSetsDoNotMutateOriginalsOrLeakAliases
go test ./flags -run TestATDDValueAndDiagnosticFoundationIsMachineReadable

# Final Story 2.1 verification after implementation:
go test ./...
go vet ./...
go run ./tools/depgate
git diff --check
```

## Red-Green-Refactor Workflow

### RED Phase (Current)

- Acceptance scaffolds exist and are skipped with `t.Skip`.
- Each skipped test runs a temporary consumer module against `github.com/petabytecl/dib/flags`.
- Removing one skip before implementation should fail with compile errors for the missing public `flags` API or with behavior assertions for incomplete implementation.
- Tests use only Go standard-library packages and preserve the dependency gate.

### GREEN Phase (DEV Workflow)

1. Pick one scaffold, starting with `TestATDDFlagDefinitionsExposeMetadata`.
2. Remove only that test's `t.Skip`.
3. Run the narrow `go test ./flags -run ...` command and confirm RED.
4. Implement the smallest production change to pass the activated test.
5. Repeat for the remaining scaffolds.

### REFACTOR Phase

- Run `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.
- Confirm root `go.mod` still contains no `require`, `replace`, or `toolchain` directives.
- Confirm no `go.sum`, `/cmd`, `internal/`, parser fuzz seed, normalization parser, or broad root facade was introduced unless the story implementation proves it necessary and architecture is updated.

## Knowledge Base References Applied

- `test-levels-framework.md`: selected package acceptance/unit tests instead of E2E for pure backend library behavior.
- `test-priorities-matrix.md`: classified all foundation contract scenarios as P0 because later Epic 2 parser stories depend on this API shape.
- `test-quality.md`: kept tests deterministic, explicit, isolated, and under 300 lines per file.
- `test-healing-patterns.md`: avoided hard waits, dynamic data, hidden assertions, and external network dependency patterns.
- `ci-burn-in.md`: preserved static, standard-library-only verification and command-level handoff.
- `data-factories.md`, `component-tdd.md`, `overview.md`, `api-request.md`, `auth-session.md`, `recurse.md`: loaded for workflow completeness; UI/API/auth-specific patterns were not directly applicable to this package-only Go story.

## Test Execution Evidence

Current scaffold verification:

```bash
go test ./flags
go test ./...
```

Both commands pass while scaffolds remain skipped. Activated RED verification is intentionally left to the dev workflow one scaffold at a time by removing the relevant `t.Skip`.

## Step 5: Validation And Completion

- Prerequisites satisfied: Story 2.1 has clear acceptance criteria, backend Go test framework exists, and development environment is available.
- Test files created correctly: `flags/atdd_contract_test.go`, `flags/set_atdd_test.go`, and `flags/state_atdd_test.go`.
- Red-phase compliance: five scaffold tests are skipped with `t.Skip`; no placeholder assertions such as `expect(true).toBe(true)` are present.
- Story metadata and handoff paths are captured in this checklist and linked back into the story file.
- CLI/browser session cleanup: N/A, no Playwright CLI, browser, or MCP session was opened.
- Temp artifacts are stored under `_bmad-output/test-artifacts/tmp/`.
- Validation commands passed: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.
- Security scan note: only the intended fake redaction corpus `dib_fake_secret_value` and story guidance text matched secret-like terms.

## Notes

- The ATDD scaffolds intentionally define the expected public API contract for Story 2.1. If implementation discovers a better API shape, update the scaffold and checklist deliberately rather than bypassing the tests.
- The helper in `flags/atdd_contract_test.go` uses temporary consumer modules so public API expectations are tested from outside the `flags` package.
- Keep all runtime and test imports standard-library-only.
