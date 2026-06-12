---
stepsCompleted: ['step-01-preflight-and-context', 'step-02-generation-mode', 'step-03-test-strategy', 'step-04-generate-tests', 'step-04c-aggregate', 'step-05-validate-and-complete']
lastStep: 'step-05-validate-and-complete'
lastSaved: '2026-06-11T20:12:58-04:00'
workflowType: 'testarch-atdd'
storyId: '2.3'
storyKey: '2-3-reject-invalid-long-flags-with-inspectable-errors'
storyFile: '_bmad-output/implementation-artifacts/2-3-reject-invalid-long-flags-with-inspectable-errors.md'
atddChecklistPath: '_bmad-output/test-artifacts/atdd-checklist-2-3-reject-invalid-long-flags-with-inspectable-errors.md'
generatedTestFiles:
  - 'flags/parse_long_atdd_test.go'
inputDocuments:
  - '_bmad/tea/config.yaml'
  - '_bmad-output/implementation-artifacts/2-3-reject-invalid-long-flags-with-inspectable-errors.md'
  - '_bmad-output/implementation-artifacts/sprint-status.yaml'
  - 'go.mod'
  - 'flags/set.go'
  - 'flags/snapshot.go'
  - 'flags/errors.go'
  - 'flags/definition.go'
  - 'flags/parser.go'
  - 'flags/normalize.go'
  - 'flags/kind.go'
  - 'flags/atdd_contract_test.go'
  - 'flags/normalize_atdd_test.go'
  - 'flags/set_atdd_test.go'
  - 'flags/state_atdd_test.go'
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
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/fixture-architecture.md'
  - '.agents/skills/bmad-testarch-atdd/resources/knowledge/network-first.md'
---

# ATDD Checklist - Epic 2, Story 2.3: Reject Invalid Long Flags With Inspectable Errors

**Date:** 2026-06-11
**Author:** Coto
**Primary Test Level:** Backend package acceptance tests

## Story Summary

Story 2.3 adds explicit long-flag parsing to the existing immutable `flags.Set` and snapshot foundation. The red-phase scaffolds exercise the expected public `flags` package from temporary consumer modules and stay skipped until the dev workflow activates one test at a time.

**As a** Go CLI developer
**I want** long flags to parse familiar forms and reject invalid input with typed diagnostics
**So that** scripts and tests can handle parser failures without scraping error text.

## Acceptance Criteria

1. Known value long flags parse `--name=value` and `--name value`; parsed snapshots record explicit value, source spelling, canonical definition identity, and remaining positional args in order.
2. Known boolean long flags parse `--name`, `--name=true`, and `--name=false`; invalid boolean text returns typed conversion diagnostics.
3. Unknown long flags before `--` return typed unknown-flag diagnostics with flag token and normalized lookup context where applicable.
4. Missing required values return typed missing-value diagnostics and failed parses do not mutate reusable flag sets.
5. Table-driven tests cover attached values, separate values, booleans, unknown flags, missing values, invalid conversions, duplicate single-value flags, exact/normalized names, and diagnostics through typed errors and snapshot state.

## Step 1: Preflight And Context

- Detected stack: backend Go module (`go.mod`).
- Story file: `_bmad-output/implementation-artifacts/2-3-reject-invalid-long-flags-with-inspectable-errors.md`.
- Story status: `ready-for-dev`.
- Existing test framework: standard Go `testing` package with package tests under `flags/`, `command/`, and tools.
- No `project-context.md` was present.
- No existing Story 2.3 ATDD artifact was present before this run.
- Story 2.1 and Story 2.2 ATDD scaffolds were used as local patterns, especially `runConsumerContract` in `flags/atdd_contract_test.go`.

## Step 2: Generation Mode

- Selected mode: AI generation.
- Execution mode: sequential backend package scaffold generation.
- Reason: Story 2.3 is pure Go package behavior. It has no HTTP endpoint, Pact provider, browser journey, UI selector, service boundary, database, or external fixture.

## Step 3: Test Strategy

| Scenario | AC | Level | Priority | Red-phase behavior |
| --- | --- | --- | --- | --- |
| Long value flags preserve source spelling and remaining args | AC1, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails because `Set.Parse`, `Snapshot.RemainingArgs`, and occurrence metadata do not exist |
| Boolean long flags parse presence and explicit values | AC2, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails until `Set.Parse` supports boolean presence, explicit bool values, parse context wrapping, and conversion preservation |
| Unknown long flags expose lookup context | AC3, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails until `ErrUnknownFlag` and `*flags.ParseError` exist |
| Missing required long values are inspectable and reusable-set safe | AC4, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails until required-value consumption and `ErrMissingValue` diagnostics exist |
| Duplicate single-value long flags are inspectable | AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails until parse-time duplicate detection reuses `ErrDuplicateValue` |
| Exact and normalized long names parse safely | AC1, AC3, AC5 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails until parse uses `Set.Lookup` and preserves Story 2.2 shorthand-alias protections |
| `--no-*` long names are ordinary names | AC2, AC3, AC5 | Backend package acceptance / unit | P1 | Removing `t.Skip` fails until parser treats registered `no-*` names normally and unregistered negation as unknown |
| Sensitive conversion errors do not leak attached values | AC2, AC5, NFR8 | Backend package acceptance / unit | P0 | Removing `t.Skip` fails until parse context preserves conversion inspection without echoing sensitive raw values |

Integration, API/contract, E2E, and component tests are not appropriate for this story because there is no service boundary, database, HTTP endpoint, browser journey, or UI component.

## Red-Phase Test Scaffolds Created

### Backend Package Tests (8 tests)

**File:** `flags/parse_long_atdd_test.go` (494 lines)

Tests:

- `TestATDDLongFlagValuesPreserveSourceAndRemainingArgs`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: `Set.Parse`, attached values, separate values, normalized raw spelling, canonical definition identity, explicit state, occurrence metadata, remaining args ordering, and defensive copies.
- `TestATDDBooleanLongFlagsParsePresenceAndExplicitValues`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: boolean presence parses true, explicit false parses false, invalid bool text preserves `ErrConversion`, `*flags.ParseError`, and `*flags.ValueError`.
- `TestATDDUnknownLongFlagsExposeLookupContext`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: unknown long flags return `ErrUnknownFlag`, expose token/name/normalized lookup key, and expose no canonical definition.
- `TestATDDMissingRequiredLongValuesAreInspectableAndLeaveSetReusable`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: omitted values, next long-flag token, and `--` terminator return `ErrMissingValue`; default snapshots and later successful parses prove the set remains reusable.
- `TestATDDDuplicateSingleValueLongFlagsAreInspectable`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: non-repeatable duplicate long flags return `ErrDuplicateValue` and typed parse context.
- `TestATDDExactAndNormalizedLongNamesParseSafely`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: exact sets parse `log-level`, `log_level`, and `log.level` distinctly; normalized sets parse equivalent raw names to canonical definitions; shorthand-only spellings stay unknown.
- `TestATDDNoPrefixedLongNamesAreOrdinaryNames`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: registered `no-color` parses as an ordinary long name and unregistered `--no-color` remains unknown.
- `TestATDDSensitiveConversionErrorsDoNotLeakAttachedValues`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: sensitive attached values do not appear in error strings or parse tokens, while `ErrConversion` and typed parse context remain inspectable.

### API Tests

N/A. This story does not expose HTTP or service API endpoints. The workflow's API worker output is represented by backend package acceptance scaffolds.

### E2E Tests

N/A. This story has no browser or user-interface journey.

### Component Tests

N/A. This story has no UI components.

## Data Factories And Fixtures

N/A. The package acceptance tests construct all data in memory with standard Go values. No external service, database, browser, authentication fixture, data-testid, mock endpoint, or cleanup infrastructure is required.

## Temporary Generation Artifacts

- `_bmad-output/test-artifacts/tmp/tea-atdd-api-tests-2026-06-11T20-12-58-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-e2e-tests-2026-06-11T20-12-58-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-summary-2026-06-11T20-12-58-0400.json`

## Implementation Checklist

### Test: `TestATDDLongFlagValuesPreserveSourceAndRemainingArgs`

**File:** `flags/parse_long_atdd_test.go`

- [ ] Add `func (s Set) Parse(args []string) (Snapshot, error)`.
- [ ] Parse `--name=value` and `--name value` for required-value long flags.
- [ ] Resolve raw long-name spelling through `Set.Lookup` so normalization and shorthand protections are reused.
- [ ] Add `Snapshot.RemainingArgs() []string` with defensive copies.
- [ ] Add `ValueState.Occurrences() []ValueOccurrence` or equivalent public source metadata with source spelling and canonical `Definition`.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDBooleanLongFlagsParsePresenceAndExplicitValues`

**File:** `flags/parse_long_atdd_test.go`

- [ ] Parse boolean presence `--verbose` as true.
- [ ] Parse explicit boolean values through `Definition.Parse`.
- [ ] Wrap invalid boolean conversion with parse context while preserving `errors.Is(err, ErrConversion)` and `errors.As(err, *ValueError)`.
- [ ] Ensure parse context exposes flag spelling/name without requiring string matching.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDUnknownLongFlagsExposeLookupContext`

**File:** `flags/parse_long_atdd_test.go`

- [ ] Add `flags.ErrUnknownFlag`.
- [ ] Add `*flags.ParseError` or equivalent typed context with `Token`, `Name`, `NormalizedName`, and optional `Definition` accessors.
- [ ] Return typed unknown diagnostics for unknown long names before `--`.
- [ ] Include normalized lookup context for configured normalizers.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDMissingRequiredLongValuesAreInspectableAndLeaveSetReusable`

**File:** `flags/parse_long_atdd_test.go`

- [ ] Add `flags.ErrMissingValue`.
- [ ] Reject required-value long flags when the value is omitted, the next token is another long flag, or the next token is `--`.
- [ ] Ensure failed parses do not mutate `Set` definitions, lookup indexes, or default snapshot behavior.
- [ ] Confirm parsing succeeds after earlier failures with the same reusable set.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDDuplicateSingleValueLongFlagsAreInspectable`

**File:** `flags/parse_long_atdd_test.go`

- [ ] Track explicit occurrences in one parse run.
- [ ] Return `ErrDuplicateValue` when a non-repeatable flag appears more than once.
- [ ] Attach typed parse context identifying the duplicate token and canonical definition.
- [ ] Do not implement repeated accumulation beyond what this duplicate guard needs; Story 2.6 owns full repeated/custom accumulation.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDExactAndNormalizedLongNamesParseSafely`

**File:** `flags/parse_long_atdd_test.go`

- [ ] Preserve exact-name parsing for sets created with `NewSet`.
- [ ] Preserve canonical snapshot keys for exact and normalized sets.
- [ ] Preserve normalized lookup behavior for `NewNormalizedSet`.
- [ ] Preserve Story 2.2 review fix: registered shorthand spellings must not become hidden long-name aliases.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDNoPrefixedLongNamesAreOrdinaryNames`

**File:** `flags/parse_long_atdd_test.go`

- [ ] Treat registered `no-*` names as ordinary long names.
- [ ] Do not auto-generate boolean negation aliases.
- [ ] Return `ErrUnknownFlag` for unregistered `--no-*` input.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestATDDSensitiveConversionErrorsDoNotLeakAttachedValues`

**File:** `flags/parse_long_atdd_test.go`

- [ ] Preserve sensitive conversion redaction from `Definition.Parse` when parse context wraps errors.
- [ ] Do not include attached raw sensitive values in `ParseError.Token()`, error strings, debug output, source reports, examples, or docs.
- [ ] Preserve `errors.Is(err, ErrConversion)` while hiding sensitive parser causes.
- [ ] Remove only this test's `t.Skip`, confirm RED, then make it pass.

## Running Tests

```bash
# Current red-phase scaffold state; tests are present but skipped.
go test ./flags

# Full repository check while scaffolds remain skipped.
go test ./...

# Activate one scaffold at a time by removing that test's t.Skip, then run:
go test ./flags -run TestATDDLongFlagValuesPreserveSourceAndRemainingArgs -count=1
go test ./flags -run TestATDDBooleanLongFlagsParsePresenceAndExplicitValues -count=1
go test ./flags -run TestATDDUnknownLongFlagsExposeLookupContext -count=1
go test ./flags -run TestATDDMissingRequiredLongValuesAreInspectableAndLeaveSetReusable -count=1
go test ./flags -run TestATDDDuplicateSingleValueLongFlagsAreInspectable -count=1
go test ./flags -run TestATDDExactAndNormalizedLongNamesParseSafely -count=1
go test ./flags -run TestATDDNoPrefixedLongNamesAreOrdinaryNames -count=1
go test ./flags -run TestATDDSensitiveConversionErrorsDoNotLeakAttachedValues -count=1

# Final Story 2.3 verification after implementation:
go test ./...
go vet ./...
go run ./tools/depgate
git diff --check
```

## Red-Green-Refactor Workflow

### RED Phase (Current)

- Acceptance scaffolds exist and are skipped with `t.Skip`.
- Each skipped test runs a temporary consumer module against `github.com/petabytecl/dib/flags`.
- Removing one skip before implementation should fail with compile errors for missing parse APIs and sentinels, or with behavior assertions for incomplete implementation.
- Tests use only Go standard-library packages and preserve the dependency gate.
- The scaffold intentionally defines parse source metadata as source spelling without attached raw values, reducing sensitive-value leak risk.

### GREEN Phase (DEV Workflow)

1. Pick one scaffold, starting with `TestATDDLongFlagValuesPreserveSourceAndRemainingArgs`.
2. Remove only that test's `t.Skip`.
3. Run the narrow `go test ./flags -run ... -count=1` command and confirm RED.
4. Implement the smallest production change to pass the activated test.
5. Repeat for the remaining scaffolds.

### REFACTOR Phase

- Run `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.
- Confirm root `go.mod` still contains no `require`, `replace`, or `toolchain` directives.
- Confirm no `go.sum`, `/cmd`, `config/` integration, short-flag parser, full terminator matrix, parser fuzz seed, or broad root facade was introduced unless the story implementation deliberately updates the architecture/story contract.
- Consider `go test -race ./...` as extra evidence because sets and snapshots are reusable values.

## Knowledge Base References Applied

- `test-levels-framework.md`: selected package acceptance/unit tests instead of E2E for pure backend library behavior.
- `test-priorities-matrix.md`: classified parser snapshot state, parse diagnostics, normalization safety, and sensitive redaction as P0 because later command/config stories depend on them.
- `test-quality.md`: kept tests deterministic, explicit, isolated, and free of waits, external IO, and host state.
- `test-healing-patterns.md`: avoided hard waits, dynamic data assertions, hidden assertions, network dependencies, and order-dependent tests.
- `ci-burn-in.md`: preserved static, standard-library-only verification and command-level handoff.
- `data-factories.md`, `component-tdd.md`, `overview.md`, `api-request.md`, `auth-session.md`, `recurse.md`, `fixture-architecture.md`, and `network-first.md`: loaded for workflow completeness; UI/API/auth-specific patterns were not directly applicable to this package-only Go story.

## Test Execution Evidence

Current scaffold verification:

```bash
go test ./flags
go test ./...
go vet ./...
go run ./tools/depgate
git diff --check
```

All commands pass while scaffolds remain skipped. Activated RED verification is intentionally left to the dev workflow one scaffold at a time by removing the relevant `t.Skip`.

## Step 5: Validation And Completion

- Prerequisites satisfied: Story 2.3 has clear acceptance criteria, backend Go test framework exists, and development environment is available.
- Test file created correctly: `flags/parse_long_atdd_test.go`.
- Red-phase compliance: eight scaffold tests are skipped with `t.Skip`; no placeholder assertions such as `expect(true).toBe(true)` are present.
- Story metadata and handoff paths are captured in this checklist and linked back into the story file.
- CLI/browser session cleanup: N/A, no Playwright CLI, browser, or MCP session was opened.
- Temp artifacts are stored under `_bmad-output/test-artifacts/tmp/`.
- Validation commands passed: `go test ./flags`, `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `git diff --check`, and `test ! -e go.sum`.
- Security scan note: a broad scan surfaced only existing BMad skill examples plus the intended fake redaction corpus `dib_fake_secret_value`; the high-confidence scan for real key formats, private keys, and quoted credential assignments passed with no candidates.

## Notes

- The ATDD scaffolds intentionally define the expected public API contract for Story 2.3. If implementation discovers a better API shape, update the scaffold and checklist deliberately rather than bypassing the tests.
- The helper in `flags/atdd_contract_test.go` uses temporary consumer modules so public API expectations are tested from outside the `flags` package.
- Keep all runtime and test imports standard-library-only.
