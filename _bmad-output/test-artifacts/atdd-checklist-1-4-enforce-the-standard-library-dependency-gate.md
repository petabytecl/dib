---
stepsCompleted: ['step-01-preflight-and-context', 'step-02-generation-mode', 'step-03-test-strategy', 'step-04-generate-tests', 'step-04c-aggregate', 'step-05-validate-and-complete']
lastStep: 'step-05-validate-and-complete'
lastSaved: '2026-06-11T14:54:34-04:00'
workflowType: 'testarch-atdd'
storyId: '1.4'
storyKey: '1-4-enforce-the-standard-library-dependency-gate'
storyFile: '_bmad-output/implementation-artifacts/1-4-enforce-the-standard-library-dependency-gate.md'
atddChecklistPath: '_bmad-output/test-artifacts/atdd-checklist-1-4-enforce-the-standard-library-dependency-gate.md'
generatedTestFiles:
  - 'tools/depgate/main_test.go'
  - 'tools/depgate/testdata/stdlib-only/go.mod'
  - 'tools/depgate/testdata/stdlib-only/pkg/pkg.go'
  - 'tools/depgate/testdata/stdlib-only/pkg/pkg_test.go'
  - 'tools/depgate/testdata/non-stdlib-test-import/go.mod'
  - 'tools/depgate/testdata/non-stdlib-test-import/pkg/pkg.go'
  - 'tools/depgate/testdata/non-stdlib-test-import/pkg/pkg_test.go'
inputDocuments:
  - '_bmad/tea/config.yaml'
  - '_bmad-output/implementation-artifacts/1-4-enforce-the-standard-library-dependency-gate.md'
  - '_bmad-output/implementation-artifacts/sprint-status.yaml'
  - 'go.mod'
  - 'command/definition_test.go'
  - 'command/contract_test.go'
  - 'CONTRIBUTING.md'
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

# ATDD Checklist - Epic 1, Story 1.4: Enforce the Standard-Library Dependency Gate

**Date:** 2026-06-11
**Author:** Coto
**Primary Test Level:** Unit-style backend command acceptance tests

## Story Summary

Story 1.4 creates isolated repository tooling that rejects non-standard-library imports across runtime packages, tests, examples, and tool packages. The acceptance scaffolds are Go tests in the future `tools/depgate` command package, with fixture modules that model both an all-stdlib project and a package-test-only external import violation.

**As a** dependency-policy reviewer
**I want** a repeatable repository gate that fails on non-standard-library imports
**So that** Dib's zero-runtime-dependency claim is enforced before feature implementation scales.

## Acceptance Criteria

1. `go run ./tools/depgate` inspects packages included by `go test ./...`, including package tests and future `examples/` packages, and fails on non-standard-library imports.
2. `tools/depgate/` stays isolated repository tooling and does not create runtime imports into the tool package.
3. Dependency failures identify the package and offending import path, exit non-zero, and do not hide other violations.
4. Documentation and CI guidance make `go run ./tools/depgate` the required dependency check once the tool exists.
5. Verification runs `go test ./...`, `go vet ./...`, and `go run ./tools/depgate`; tests cover passing stdlib-only and failing non-stdlib fixtures.

## Step 1: Preflight And Context

- Detected stack: backend Go module (`go.mod`).
- Story file: `_bmad-output/implementation-artifacts/1-4-enforce-the-standard-library-dependency-gate.md`.
- Story status: `ready-for-dev`.
- Existing test framework: standard Go `testing` package via `command/definition_test.go` and `command/contract_test.go`.
- No `project-context.md` was present.
- `tools/depgate/` did not exist before ATDD generation.

## Step 2: Generation Mode

- Selected mode: AI generation.
- Execution mode: sequential.
- Reason: backend-only Go tooling has no UI recording target, no API endpoint contract, and no explicit user authorization to launch background subagents. The workflow's API/E2E worker structure was adapted to backend command acceptance scaffolds.

## Step 3: Test Strategy

| Scenario | AC | Level | Priority | Red-phase behavior |
| --- | --- | --- | --- | --- |
| Stdlib-only fixture exits successfully without dependency violations | AC1, AC5 | Backend command acceptance | P0 | Removing `t.Skip` before implementation fails because `tools/depgate` has no command implementation yet |
| Non-stdlib test imports fail with actionable package/import diagnostics | AC1, AC3, AC5 | Backend command acceptance | P0 | Removing `t.Skip` before implementation fails before the command exists |
| Multiple violations are all reported with deterministic output | AC3 | Backend command acceptance | P0 | Removing `t.Skip` before implementation fails before the command exists |

AC4 is documented as an implementation task rather than an ATDD scaffold because it is a contributor documentation update. The dev workflow should update `CONTRIBUTING.md` when implementing the tool.

## Red-Phase Test Scaffolds Created

### Backend Command Tests (3 tests)

**File:** `tools/depgate/main_test.go`

- `TestDepgateAllowsStdlibOnlyFixture`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: a fixture module with only standard-library imports is accepted.
- `TestDepgateReportsNonStandardTestImports`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: a package test importing fake external packages fails and names both package/import fields.
- `TestDepgateReportsEveryViolationDeterministically`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: multiple violations are emitted on every run with stable output.

### API Tests

N/A. Story 1.4 does not expose HTTP or service API endpoints.

### E2E Tests

N/A. Story 1.4 has no browser or user-interface journey.

### Component Tests

N/A. Story 1.4 has no UI components.

## Fixtures Created

### Stdlib-Only Fixture

**Directory:** `tools/depgate/testdata/stdlib-only/`

- `go.mod`
- `pkg/pkg.go`
- `pkg/pkg_test.go`

This fixture represents a module whose package and package tests use only the Go standard library.

### Non-Stdlib Test Import Fixture

**Directory:** `tools/depgate/testdata/non-stdlib-test-import/`

- `go.mod`
- `pkg/pkg.go`
- `pkg/pkg_test.go`

This fixture uses package tests to import `example.com/external/notstdlib` and `example.com/external/also-notstdlib`. The fake imports are deliberately unresolved and must not be fetched with `go get`.

## Mock Requirements

N/A. No external service, network, browser, database, or authentication fixture is required.

## Required UI Test Identifiers

N/A. This story has no UI.

## Implementation Checklist

### Test: `TestDepgateAllowsStdlibOnlyFixture`

**File:** `tools/depgate/main_test.go`

- [ ] Create `tools/depgate/main.go` as an isolated `package main` command.
- [ ] Make the command run `go list -deps -test -e -json -buildvcs=false ./...` from its current working directory.
- [ ] Allow decoded packages when `Standard == true` or `Module != nil && Module.Main == true`.
- [ ] Remove the test's `t.Skip` and verify it fails before implementation if the command is missing or incorrect.
- [ ] Implement the minimum behavior until the test passes.

### Test: `TestDepgateReportsNonStandardTestImports`

**File:** `tools/depgate/main_test.go`

- [ ] Decode the `go list` JSON stream with `encoding/json.Decoder`.
- [ ] Treat unresolved external test imports surfaced by `-e` as dependency violations.
- [ ] Emit line-oriented diagnostics containing `package=` and `import=`.
- [ ] Remove the test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestDepgateReportsEveryViolationDeterministically`

**File:** `tools/depgate/main_test.go`

- [ ] Collect all violations before returning.
- [ ] Sort diagnostics deterministically before printing.
- [ ] Keep dependency-policy failures distinct from execution and JSON decode failures.
- [ ] Remove the test's `t.Skip`, confirm RED, then make it pass.

### Documentation Task For AC4

- [ ] Update `CONTRIBUTING.md` so local verification requires `go run ./tools/depgate` after `go test ./...` and `go vet ./...`.
- [ ] Remove the old conditional wording that allowed a temporary ad hoc dependency check.

## Running Tests

```bash
# Current red-phase scaffold state; tests are present but skipped.
go test ./tools/depgate

# Full repository check while scaffolds remain skipped.
go test ./...

# Activate one scaffold at a time by removing that test's t.Skip, then run:
go test ./tools/depgate -run TestDepgateAllowsStdlibOnlyFixture
go test ./tools/depgate -run TestDepgateReportsNonStandardTestImports
go test ./tools/depgate -run TestDepgateReportsEveryViolationDeterministically

# Final Story 1.4 verification after implementation:
go test ./...
go vet ./...
go run ./tools/depgate
```

## Red-Green-Refactor Workflow

### RED Phase (Current)

- Acceptance scaffolds exist and are skipped with `t.Skip`.
- Removing a skip before implementation should fail because `tools/depgate/main.go` does not exist yet or behavior is incomplete.
- Fixture modules are static and local; they do not modify root `go.mod`.

### GREEN Phase (DEV Workflow)

1. Pick one scaffold, starting with `TestDepgateAllowsStdlibOnlyFixture`.
2. Remove only that test's `t.Skip`.
3. Run the narrow `go test ./tools/depgate -run ...` command and confirm RED.
4. Implement the smallest production change to pass the activated test.
5. Repeat for the remaining scaffolds.

### REFACTOR Phase

- Run `go test ./...`, `go vet ./...`, and `go run ./tools/depgate`.
- Confirm root `go.mod` still contains no dependency directives.
- Confirm runtime packages do not import `tools/depgate`.

## Validation

- ATDD output file created at `_bmad-output/test-artifacts/atdd-checklist-1-4-enforce-the-standard-library-dependency-gate.md`.
- Red-phase Go scaffolds are marked with `t.Skip`, the Go equivalent of the workflow's `test.skip()` requirement.
- No active failing tests were added before implementation.
- Worker and summary artifacts were stored under `_bmad-output/test-artifacts/tmp/`.
- No browser sessions were opened.

## Knowledge Base References Applied

- `test-levels-framework.md`: selected backend command/unit-style acceptance tests instead of E2E.
- `test-priorities-matrix.md`: classified dependency-policy failures as P0 because they protect repository trust and release gates.
- `test-quality.md`: kept scaffolds deterministic, focused, standard-library-only, and explicit.
- `data-factories.md`: used local fixture modules with explicit intent instead of shared mutable state.
- `ci-burn-in.md`: carried forward guidance that Story 1.5 should wire fixed commands and avoid shell injection from dynamic inputs.
- `overview.md`, `api-request.md`, `auth-session.md`, and `recurse.md`: loaded per TEA config; not applied because this story has no Playwright API/auth/polling surface.

## Test Execution Evidence

The scaffold verification command is expected to pass while all red-phase tests remain skipped:

```text
go test ./tools/depgate
```

Final command output is recorded in the workflow completion response rather than pasted here to avoid stale evidence.
