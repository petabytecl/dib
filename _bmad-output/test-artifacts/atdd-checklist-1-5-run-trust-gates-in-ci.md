---
stepsCompleted: ['step-01-preflight-and-context', 'step-02-generation-mode', 'step-03-test-strategy', 'step-04-generate-tests', 'step-04c-aggregate', 'step-05-validate-and-complete']
lastStep: 'step-05-validate-and-complete'
lastSaved: '2026-06-11T15:59:33-04:00'
workflowType: 'testarch-atdd'
storyId: '1.5'
storyKey: '1-5-run-trust-gates-in-ci'
storyFile: '_bmad-output/implementation-artifacts/1-5-run-trust-gates-in-ci.md'
atddChecklistPath: '_bmad-output/test-artifacts/atdd-checklist-1-5-run-trust-gates-in-ci.md'
generatedTestFiles:
  - 'tools/cigate/ci_test.go'
inputDocuments:
  - '_bmad/tea/config.yaml'
  - '_bmad-output/implementation-artifacts/1-5-run-trust-gates-in-ci.md'
  - '_bmad-output/implementation-artifacts/sprint-status.yaml'
  - 'go.mod'
  - 'CONTRIBUTING.md'
  - 'tools/depgate/main_test.go'
  - 'tools/depgate/main.go'
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

# ATDD Checklist - Epic 1, Story 1.5: Run Trust Gates in CI

**Date:** 2026-06-11
**Author:** Coto
**Primary Test Level:** Backend repository acceptance tests

## Story Summary

Story 1.5 creates GitHub Actions CI and release evidence documentation for Dib's trust gates. The acceptance scaffolds are skipped Go tests that inspect repository files after implementation without adding project dependencies or YAML parsing libraries.

**As a** technical reviewer
**I want** Dib's core trust gates to run automatically in CI
**So that** standard-library-only dependency enforcement, tests, vet, and clean-room evidence do not rely on manual discipline.

## Acceptance Criteria

1. `.github/workflows/ci.yml` runs on pull requests and pushes to `main`, and uses an explicit GitHub-hosted runner image such as `ubuntu-24.04`.
2. CI installs Go from `go.mod`, and release guidance treats Go version drift across `go.mod`, docs, CI, and release guidance as blocking.
3. CI executes `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` as blocking gates.
4. `docs/release-checklist.md` records `go test -race ./...` as a release-candidate gate and includes placeholders for exact commit, test, vet, dependency-gate, race-test, docs/examples, provenance, compatibility, migration, runner image, action versions, owner, date, waivers, reason, and expiry.
5. Local verification uses the same commands as CI and does not add non-standard-library project dependencies or generated scaffolding that weakens the dependency claim.

## Step 1: Preflight And Context

- Detected stack: backend Go module (`go.mod`).
- Story file: `_bmad-output/implementation-artifacts/1-5-run-trust-gates-in-ci.md`.
- Story status: `ready-for-dev`.
- Existing test framework: standard Go `testing` package.
- No `project-context.md` was present.
- `.github/workflows/ci.yml` and `docs/release-checklist.md` were absent before ATDD generation.

## Step 2: Generation Mode

- Selected mode: AI generation.
- Execution mode: sequential.
- Reason: backend repository CI tooling has no UI recording target and no HTTP API contract. The workflow's API/E2E worker shape was adapted to static Go acceptance scaffolds.

## Step 3: Test Strategy

| Scenario | AC | Level | Priority | Red-phase behavior |
| --- | --- | --- | --- | --- |
| CI workflow triggers and runs trust gates | AC1, AC2, AC3 | Backend repository acceptance | P0 | Removing `t.Skip` before implementation fails because `.github/workflows/ci.yml` is missing |
| CI workflow uses only trusted static steps | AC1, AC3, AC5 | Backend repository acceptance | P0 | Removing `t.Skip` before implementation fails until workflow hardening is implemented |
| Release checklist captures required evidence | AC2, AC4 | Backend repository acceptance | P0 | Removing `t.Skip` before implementation fails because `docs/release-checklist.md` is missing |
| Repository still proves zero project dependencies | AC5 | Backend repository acceptance | P0 | Removing `t.Skip` before implementation fails until workflow exists and root `go.mod` remains unchanged |

## Red-Phase Test Scaffolds Created

### Backend Repository Tests (4 tests)

**File:** `tools/cigate/ci_test.go`

- `TestCIWorkflowRunsTrustGates`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: workflow name, triggers, runner, official checkout/setup-go actions, Go version source, disabled cache, and trust gate commands.
- `TestCIWorkflowUsesOnlyTrustedStaticSteps`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: no floating runner label, untrusted GitHub/input interpolation, remote shell installer, dependency fetch, module rewrite, or unsupported third-party actions.
- `TestReleaseChecklistCapturesRequiredEvidence`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: release checklist placeholders and blocking policy for all required evidence.
- `TestStory15DoesNotWeakenStandardLibraryClaim`
  - Status: RED scaffold, skipped with `t.Skip`.
  - Verifies: root `go.mod` remains the zero-dependency Go 1.26 baseline while workflow evidence exists.

### API Tests

N/A. Story 1.5 does not expose HTTP or service API endpoints.

### E2E Tests

N/A. Story 1.5 has no browser or user-interface journey.

### Component Tests

N/A. Story 1.5 has no UI components.

## Fixtures Created

N/A. Static repository files are sufficient; no external service, database, browser, or authentication fixture is required.

## Mock Requirements

N/A. No external service is mocked.

## Required UI Test Identifiers

N/A. This story has no UI.

## Implementation Checklist

### Test: `TestCIWorkflowRunsTrustGates`

**File:** `tools/cigate/ci_test.go`

- [ ] Create `.github/workflows/ci.yml`.
- [ ] Configure `pull_request` and `push` to `main`.
- [ ] Use `runs-on: ubuntu-24.04`.
- [ ] Use `actions/checkout@v6` and `actions/setup-go@v6`.
- [ ] Configure `go-version-file: go.mod` and `cache: false`.
- [ ] Add distinct steps for `go version`, `go test ./...`, `go vet ./...`, and `go run ./tools/depgate`.
- [ ] Remove this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestCIWorkflowUsesOnlyTrustedStaticSteps`

**File:** `tools/cigate/ci_test.go`

- [ ] Avoid `ubuntu-latest`.
- [ ] Avoid `${{ github.event.* }}`, `${{ github.head_ref }}`, and `${{ inputs.* }}` interpolation in `run:` blocks.
- [ ] Avoid `curl`, `go get`, and `go mod tidy` in the workflow.
- [ ] Keep workflow actions limited to official `actions/checkout@v6` and `actions/setup-go@v6`.
- [ ] Remove this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestReleaseChecklistCapturesRequiredEvidence`

**File:** `tools/cigate/ci_test.go`

- [ ] Create `docs/release-checklist.md`.
- [ ] Document Go version alignment across `go.mod`, CI, docs, and release guidance as release-blocking.
- [ ] Include placeholders for exact commit, test, vet, dependency-gate, race-test, docs/examples, provenance, compatibility, migration, runner image, action versions, owner, date, waiver reason, and waiver expiry.
- [ ] State that CI failures block tagging.
- [ ] Frame release evidence for Go module tags.
- [ ] Remove this test's `t.Skip`, confirm RED, then make it pass.

### Test: `TestStory15DoesNotWeakenStandardLibraryClaim`

**File:** `tools/cigate/ci_test.go`

- [ ] Confirm root `go.mod` still contains only `module github.com/petabytecl/dib` and `go 1.26`.
- [ ] Do not add `require`, `replace`, or `toolchain` directives.
- [ ] Do not add root `go.sum` or dependency-fetch commands.
- [ ] Remove this test's `t.Skip`, confirm RED, then make it pass.

## Running Tests

```bash
# Current red-phase scaffold state; tests are present but skipped.
go test ./tools/cigate

# Full repository check while scaffolds remain skipped.
go test ./...

# Activate one scaffold at a time by removing that test's t.Skip, then run:
go test ./tools/cigate -run TestCIWorkflowRunsTrustGates
go test ./tools/cigate -run TestCIWorkflowUsesOnlyTrustedStaticSteps
go test ./tools/cigate -run TestReleaseChecklistCapturesRequiredEvidence
go test ./tools/cigate -run TestStory15DoesNotWeakenStandardLibraryClaim

# Final Story 1.5 verification after implementation:
go test ./...
go vet ./...
go run ./tools/depgate
git diff --check
```

## Red-Green-Refactor Workflow

### RED Phase (Current)

- Acceptance scaffolds exist and are skipped with `t.Skip`.
- Removing a skip before implementation should fail because `.github/workflows/ci.yml` or `docs/release-checklist.md` does not exist yet.
- Tests use only Go standard-library packages and do not parse YAML with external dependencies.

### GREEN Phase (DEV Workflow)

1. Pick one scaffold, starting with `TestCIWorkflowRunsTrustGates`.
2. Remove only that test's `t.Skip`.
3. Run the narrow `go test ./tools/cigate -run ...` command and confirm RED.
4. Implement the smallest production change to pass the activated test.
5. Repeat for the remaining scaffolds.

### REFACTOR Phase

- Run `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.
- Confirm root `go.mod` still contains no dependency directives.
- Confirm no extra workflow, release, binary, Docker, Kubernetes, service, or deployment files were created.

## Knowledge Base References Applied

- `test-levels-framework.md`: selected repository acceptance tests instead of E2E for backend tooling.
- `test-priorities-matrix.md`: classified all CI trust-gate evidence checks as P0 because failure weakens release trust.
- `test-quality.md`: kept assertions explicit and deterministic without hard waits or external dependencies.
- `ci-burn-in.md`: applied static GitHub Actions hardening around explicit runners, static commands, and untrusted interpolation.
- `data-factories.md`, `component-tdd.md`, `test-healing-patterns.md`, `overview.md`, `api-request.md`, `auth-session.md`, `recurse.md`: loaded for workflow completeness; UI/API/auth-specific patterns were not directly applicable.

## Test Execution Evidence

Current scaffold verification should report skipped tests:

```bash
go test ./tools/cigate
go test ./...
```

Activated tests are expected to fail before implementation and pass only after the workflow and release checklist exist.

## Notes

- `tools/cigate` is test-only acceptance scaffolding for Story 1.5; it is not runtime code and does not create a CLI.
- Do not add YAML parser dependencies. The acceptance checks intentionally use string assertions to preserve the standard-library dependency claim.
- The dev workflow should update this checklist only if implementation changes the scaffold path or command set.
