---
baseline_commit: b4bcea91f5d747b764a08abfe06c775a806bd571
created: "2026-06-11T15:51:57-04:00"
---

# Story 1.5: Run Trust Gates in CI

Status: review

## Story

As a technical reviewer,
I want Dib's core trust gates to run automatically in CI,
so that standard-library-only dependency enforcement, tests, vet, and clean-room evidence do not rely on manual discipline.

## Requirements Trace

- FR20: Maintainers must be able to validate Dib behavior and trust evidence through repeatable tests and matrices.
- FR21: Maintainers must be able to verify the standard-library-only dependency rule with repository checks and release evidence.
- NFR1, NFR6, NFR9: runtime packages, tests, examples, and tools stay standard-library-only; checks remain table-driven and use Go 1.26+.

## Acceptance Criteria

1. Given Dib uses GitHub Actions for repository verification, when `.github/workflows/ci.yml` is created, then it runs on pull requests and pushes to the default development branch, and it uses an explicit GitHub-hosted runner image such as `ubuntu-24.04`.
2. Given the project targets Go 1.26, when CI config installs Go, then the configured Go version matches `go.mod`, docs, and release guidance, and version drift is called out as a release-blocking issue.
3. Given core PR gates are defined by architecture, when CI runs, then it executes `go test ./...`, `go vet ./...`, and `go run ./tools/depgate`, and failures from any gate block the check.
4. Given release-candidate evidence will be consolidated later, when CI and docs are updated, then `docs/release-checklist.md` records that `go test -race ./...` is a release-candidate gate, and the checklist has placeholders for exact commit, test, vet, dependency-gate, race-test, docs/examples, provenance, compatibility, and migration evidence.
5. Given CI is part of Dib's adopter trust story, when verification runs locally, then the same commands used in CI pass locally, and the workflow avoids non-standard-library project dependencies or generated scaffolding that would weaken the dependency claim.

## Tasks / Subtasks

- [x] Confirm current tracker and repository state (AC: 1-5)
  - [x] Verify Story 1.4 is `done` in `sprint-status.yaml` and GitHub issue #10 is closed.
  - [x] Verify `.github/workflows/ci.yml` does not already exist. If it exists, update it in place rather than replacing user work.
  - [x] Verify `docs/release-checklist.md` does not already exist. If it exists, update it in place.
  - [x] Verify root `go.mod` still declares `module github.com/petabytecl/dib` and `go 1.26` with no root dependency directives.

- [x] Add the GitHub Actions trust-gate workflow (AC: 1, 2, 3, 5)
  - [x] Create `.github/workflows/ci.yml`.
  - [x] Trigger on pull requests and pushes to `main`, the current default development branch.
  - [x] Use explicit `runs-on: ubuntu-24.04`; do not use `ubuntu-latest`.
  - [x] Use official GitHub actions only: `actions/checkout` and `actions/setup-go`.
  - [x] Configure `actions/setup-go` from `go.mod` using `go-version-file: go.mod` so CI tracks the repository Go version policy.
  - [x] Keep `check-latest` unset or `false` for stable repeatable CI unless architecture is updated.
  - [x] Set `cache: false` unless a root `go.sum` exists; the repository currently has no third-party dependencies to cache.
  - [x] Run `go version` before gates so CI output records the exact installed Go patch version.
  - [x] Run `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` as distinct named steps.
  - [x] Do not add third-party action dependencies, generated scaffolding, Docker/Kubernetes config, or a `/cmd` binary scaffold.

- [x] Add release-candidate checklist documentation (AC: 2, 4, 5)
  - [x] Create `docs/release-checklist.md`.
  - [x] Record the Go version alignment policy: `go.mod`, CI, docs, and release guidance must agree, and drift blocks release.
  - [x] Include placeholders for exact commit, `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `go test -race ./...`, docs/examples, provenance, compatibility, migration evidence, runner image, action versions, owner, date, and waivers.
  - [x] State that CI failures block tagging and that waivers require owner, reason, and expiry.
  - [x] Keep release evidence framed for Go module tags, not binary deployments.

- [x] Verify the story output (AC: 1-5)
  - [x] Run `go test ./...`.
  - [x] Run `go vet ./...`.
  - [x] Run `go run ./tools/depgate`.
  - [x] Confirm the workflow syntax is static YAML and has no dynamic shell interpolation from untrusted inputs.
  - [x] Confirm no runtime package imports `.github/`, `docs/`, or `tools/depgate`.
  - [x] Confirm root `go.mod` still has no `require`, `replace`, or `toolchain` directives.
  - [x] Confirm no extra workflow, release, binary, Docker, Kubernetes, service, or deployment files were created.
  - [x] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### ATDD Artifacts

- Checklist: `_bmad-output/test-artifacts/atdd-checklist-1-5-run-trust-gates-in-ci.md`
- Backend repository tests: `tools/cigate/ci_test.go`
- Temp API/back-end generation summary: `_bmad-output/test-artifacts/tmp/tea-atdd-api-tests-2026-06-11T15-59-33-0400.json`
- Temp E2E generation summary: `_bmad-output/test-artifacts/tmp/tea-atdd-e2e-tests-2026-06-11T15-59-33-0400.json`
- Temp aggregate summary: `_bmad-output/test-artifacts/tmp/tea-atdd-summary-2026-06-11T15-59-33-0400.json`
- Dev workflow handoff: remove one `t.Skip` in `tools/cigate/ci_test.go` at a time, confirm RED with the narrow `go test ./tools/cigate -run ...` command, then implement the smallest change to pass.

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded readiness report: `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-11.md`.
- Loaded previous story: `_bmad-output/implementation-artifacts/1-4-enforce-the-standard-library-dependency-gate.md`.
- Loaded current source/docs: `go.mod`, `CONTRIBUTING.md`, and the current absence of `.github/`.
- No UX document, `project-context.md`, `CLAUDE.md`, or local `MEMORY.md` was discovered in the repo.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `b4bcea91f5d747b764a08abfe06c775a806bd571` (`fix: harden depgate review coverage`).
- `main` is ahead of `origin/main` by four commits at story creation; do not assume Story 1.1-1.4 work has been pushed.
- GitHub story issues #7, #8, #9, and #10 are closed. Issue #11 is the active Story 1.5 tracker and is still open.
- `sprint-status.yaml` marks Story 1.1 through Story 1.4 as `done`; Story 1.5 is the first backlog story before this file is created.
- Root `go.mod` currently contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

- Local verification guidance in `CONTRIBUTING.md` already requires:

```text
go test ./...
go vet ./...
go run ./tools/depgate
```

### Architecture Guardrails

- GitHub Actions plus Go module release gates are the approved CI/release path; Dib is not a binary deployment product. [Source: `_bmad-output/planning-artifacts/architecture.md:239`, `_bmad-output/planning-artifacts/architecture.md:259`]
- Use an explicit runner image such as `ubuntu-24.04`, official actions, and Go version alignment across `go.mod`, docs, CI, and release notes. [Source: `_bmad-output/planning-artifacts/architecture.md:241`]
- Core PR gates are `go test ./...`, `go vet ./...`, and the dependency gate. [Source: `_bmad-output/planning-artifacts/architecture.md:246`]
- `go test -race ./...` is a release-candidate gate, not necessarily a required PR gate for this story unless explicitly added later. [Source: `_bmad-output/planning-artifacts/architecture.md:256`, `_bmad-output/planning-artifacts/architecture.md:672`]
- Release evidence must record test, vet, dependency-gate, race-test, docs/examples, runner/action version, provenance, compatibility/migration status, and waivers with owner, reason, and expiry. [Source: `_bmad-output/planning-artifacts/architecture.md:261`, `_bmad-output/planning-artifacts/architecture.md:677`]
- `.github/workflows/ci.yml` owns CI execution. No `.env`, Docker, Kubernetes, or app deployment config belongs in V1. [Source: `_bmad-output/planning-artifacts/architecture.md:632`]
- `go run ./tools/depgate` is now the required dependency-gate entry point. Do not resurrect the temporary `go list` evidence path. [Source: `_bmad-output/planning-artifacts/architecture.md:670`, `_bmad-output/planning-artifacts/architecture.md:822`]
- Runtime code must remain under `command/`, `flags/`, and `config/`; this story should not touch those packages except to confirm gates pass. [Source: `_bmad-output/planning-artifacts/epics.md:91`, `_bmad-output/planning-artifacts/architecture.md:690`]

### GitHub Actions Research Notes

- GitHub's Go CI documentation recommends `actions/setup-go` for consistent Go setup across runners and shows using normal local Go commands in workflow steps. Source: https://docs.github.com/actions/automating-builds-and-tests/building-and-testing-go
- `actions/setup-go` supports `go-version-file: go.mod`, default `check-latest: false`, and built-in Go caching. Source: https://github.com/actions/setup-go
- Current upstream releases observed at story creation: `actions/setup-go@v6.4.0` and `actions/checkout@v6.0.3`. It is acceptable to use major tags `actions/setup-go@v6` and `actions/checkout@v6` because final SHA-pinning strategy remains a deferred architecture decision.
- GitHub runner images document `ubuntu-24.04` as an explicit supported YAML label. Source: https://github.com/actions/runner-images

### Previous Story Intelligence

- Story 1.4 implemented `tools/depgate` and made it the repository dependency authority.
- Story 1.4 code review found that workspace mode could hide external modules as main modules; `tools/depgate` now sets `GOWORK=off`. CI should not override this by setting `GOWORK` globally.
- Story 1.4 added regression fixtures with local `replace` directives under `tools/depgate/testdata/`; these are intentional fixture-local dependencies and must not be copied into the root module.
- `tools/depgate` currently uses only standard-library imports and must remain isolated tooling. `.github/workflows/ci.yml` should call it, not import it.
- The Dev Agent Record for Story 1.4 includes the final passing gates after review fixes: `go test -count=1 ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.

### Expected File Changes

Expected new files:

```text
.github/workflows/ci.yml
docs/release-checklist.md
```

Expected updates:

```text
_bmad-output/implementation-artifacts/1-5-run-trust-gates-in-ci.md
_bmad-output/implementation-artifacts/sprint-status.yaml
```

Do not create these in this story unless a failing verification gate proves they are strictly necessary:

```text
go.sum
cmd/
internal/
Dockerfile
docker-compose.yml
kubernetes/
.env
.github/workflows/release.yml
```

### Implementation Guidance

- Keep the workflow small and deterministic. A single job named `ci` is enough unless implementation evidence proves a split is needed.
- Use a static workflow shape similar to:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  ci:
    runs-on: ubuntu-24.04
    steps:
      - name: Checkout
        uses: actions/checkout@v6
      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: false
      - name: Show Go version
        run: go version
      - name: Test
        run: go test ./...
      - name: Vet
        run: go vet ./...
      - name: Dependency gate
        run: go run ./tools/depgate
```

- If implementation chooses exact action tags instead of major tags, use the current observed releases and record them in `docs/release-checklist.md`; do not add SHA pinning unless the architecture is updated.
- Avoid `go get`, `go mod tidy`, generated workflow templates, and dynamic shell fragments. There is no root `go.sum`; adding one would be a sign that a dependency was introduced.
- `docs/release-checklist.md` should be a template with unchecked placeholders, not a claim that a release is ready.

### Testing Standards

- CI syntax cannot be fully executed locally without GitHub Actions, so local verification is the command equivalence contract.
- Required local verification for implementation: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.
- If an available local YAML parser exists in the environment without adding project dependencies, it may be used as extra evidence. Do not add a YAML parser dependency to the repository.
- Keep release-checklist wording audit-friendly and avoid recording stale command output.

### Security And Quality Checks

- Do not hardcode secrets, credentials, tokens, or repository-specific private URLs in workflow or docs.
- Do not request broad workflow permissions. The default read-only token permissions are enough for checkout and Go setup; if a `permissions:` block is added, prefer `contents: read`.
- Do not use third-party actions, curl-piped installers, remote shell scripts, or generated scaffolding.
- Do not add untrusted input interpolation to shell commands. The required gate commands are static.
- Error output from CI should show command failures but must not include secret values; this story should not introduce secret handling.

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `go test ./tools/cigate -run TestCIWorkflowRunsTrustGates` failed RED before `.github/workflows/ci.yml` existed, then passed after workflow creation.
- `go test ./tools/cigate -run TestReleaseChecklistCapturesRequiredEvidence` failed RED before `docs/release-checklist.md` existed, then passed after checklist creation.
- `go test ./tools/cigate -run TestCIWorkflowUsesOnlyTrustedStaticSteps` passed after the static workflow was created.
- `go test ./tools/cigate -run TestStory15DoesNotWeakenStandardLibraryClaim` passed after workflow/docs implementation with root `go.mod` unchanged and no `go.sum`.
- Final verification passed: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `git diff --check`.
- Unsafe workflow/docs interpolation scan returned no matches for `${{ github.event`, `${{ github.head_ref`, `${{ inputs.`, `ubuntu-latest`, `go get`, `go mod tidy`, or `curl ` outside the scaffold deny-list.
- Strict secret-token/key pattern scan returned no matches.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Confirmed Story 1.4 is done locally and GitHub issue #10 is closed.
- Activated Story 1.5 ATDD scaffolds in `tools/cigate/ci_test.go` and implemented static repository checks without third-party dependencies.
- Added `.github/workflows/ci.yml` with a single `ci` job on `ubuntu-24.04`, official checkout/setup-go actions, Go version from `go.mod`, `cache: false`, and distinct `go version`, `go test ./...`, `go vet ./...`, and dependency-gate steps.
- Added `docs/release-checklist.md` as a Go module tag evidence template with Go version alignment, CI gates, race gate, docs/examples, provenance, compatibility, migration, runner/action, owner/date, and waiver placeholders.
- Preserved the root zero-dependency module baseline: no `require`, `replace`, `toolchain`, root `go.sum`, generated scaffolding, Docker/Kubernetes config, binary scaffold, or release workflow was added.

### File List

- `.github/workflows/ci.yml`
- `_bmad-output/implementation-artifacts/1-5-run-trust-gates-in-ci.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/test-artifacts/atdd-checklist-1-5-run-trust-gates-in-ci.md`
- `_bmad-output/test-artifacts/tmp/tea-atdd-api-tests-2026-06-11T15-59-33-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-e2e-tests-2026-06-11T15-59-33-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-summary-2026-06-11T15-59-33-0400.json`
- `docs/release-checklist.md`
- `tools/cigate/ci_test.go`

### Change Log

- 2026-06-11: Added Story 1.5 ATDD scaffold and checklist artifacts.
- 2026-06-11: Implemented CI trust-gate workflow, release checklist template, activated repository acceptance tests, and moved story to review.
