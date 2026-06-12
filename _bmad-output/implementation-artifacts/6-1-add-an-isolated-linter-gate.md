---
baseline_commit: 7cfdfca
created: "2026-06-12"
---

# Story 6.1: Add an Isolated Linter Gate

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a release reviewer,
I want a deterministic linter gate that is isolated from Dib runtime imports,
so that CI catches maintainability issues without weakening the standard-library dependency promise.

## Requirements Trace

- FR21: `tools/depgate` remains authoritative for proving runtime, test, example, and tool packages do not import external modules unless PRD and architecture are updated.
- FR23: release candidates must run an automated lint gate in CI; the linter must be pinned, reproducible, and isolated as development or CI tooling.
- NFR1: Dib runtime packages must import only the Go standard library.
- NFR9: Dib V1 requires Go 1.26 or newer; do not add a root `toolchain` directive unless architecture changes.
- NFR11: lint and coverage gates must be deterministic enough to run locally and in CI with the same documented commands.
- SM6: CI fails when lint or package-aware coverage validation fails; this story delivers the lint half.

## Acceptance Criteria

1. Given Dib must remain standard-library-only at runtime, when the linter approach is selected, then the selected tool, version or pinning mechanism, invocation command, and isolation model are documented, and no external linter package enters Dib library, test, example, or approved tool imports unless the PRD and architecture are updated.
2. Given maintainers need local and CI parity, when linting is configured, then there is a deterministic local command for running the lint gate, and `.github/workflows/ci.yml` runs the same effective lint gate on push and pull request events.
3. Given dependency evidence remains authoritative, when release checks run, then `go run ./tools/depgate` still passes, and `docs/release-checklist.md` records the lint command, result, and isolation evidence.

## Tasks / Subtasks

- [x] Confirm preconditions and read UPDATE files before editing (AC: 1-3)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 6 `in-progress` and Story 6.1 `ready-for-dev`.
  - [x] Confirm current `HEAD` is `7cfdfcaf62b9f344eb4258eb03fc95f6e4783ac6` (`docs(bmad): add epic 6 release hardening plan`) or account for newer user changes.
  - [x] Read these UPDATE files completely before editing: `.github/workflows/ci.yml`, `tools/cigate/ci_test.go`, `docs/release-checklist.md`, `docs/release_checklist_test.go`, `docs/behavior-matrices.md`, `docs/behavior_matrices_test.go`, `docs/release-notes-v0.md`, `docs/provenance-log.md`, `tools/depgate/main.go`, `tools/depgate/main_test.go`, and `go.mod`.
  - [x] Check whether `docs/testing.md`, `.golangci.yml`, `.golangci-lint-version`, `.gitattributes`, `go.sum`, `Makefile`, or root `README.md` exist before choosing file locations. At story creation, none of these were present.
  - [x] Do not add runtime APIs, package-global helpers, root facade packages, `/cmd` scaffolding, Docker/Kubernetes files, generated shell completion, generated man pages, binary release automation, or coverage policy work owned by Story 6.2.

- [x] Select and document the lint mechanism and isolation model (AC: 1)
  - [x] Perform the short tooling review required by the sprint-change proposal and record the selected mechanism in `docs/testing.md` or another release-relevant docs location.
  - [x] Document the exact local command maintainers run, the exact CI command or action configuration, the version/pinning mechanism, and where the tool is installed or invoked.
  - [x] Preferred low-risk paths:
    - A repository-local standard-library tool, for example `go run ./tools/lint`, if it provides meaningful lint behavior beyond `go vet` and remains standard-library-only.
    - Or an isolated external CI/development linter such as `golangci-lint`, pinned by exact action version and exact linter version or version file, with no root `go.mod` `require`, `replace`, or `toolchain` additions.
  - [x] If external lint tooling is selected, prove it is development/CI-only: no import from the linter appears under `command/`, `flags/`, `config/`, `docs/`, `examples/`, or `tools/` packages, and root `go.mod` remains dependency-free unless architecture is explicitly updated.
  - [x] Avoid floating versions such as `latest`, `stable`, unpinned shell installers, or unreviewed `go install package@latest`.

- [x] Add the deterministic local lint command (AC: 1, 2)
  - [x] Add any required config file such as `.golangci.yml`, `.golangci-lint-version`, `tools/lint/`, or `docs/testing.md` only after the selected mechanism is documented.
  - [x] The command must be runnable from a clean checkout with a clear prerequisite. If the command depends on an external binary, document the expected version and installation/pinning model.
  - [x] Keep any repository-local lint helper under `tools/` and standard-library-only so `go run ./tools/depgate` still passes.
  - [x] Do not add a root `Makefile` unless it is the simplest project-consistent way to expose the command; there is no Makefile today and prior gates use direct `go ...` commands.

- [x] Wire the same effective lint gate into CI (AC: 2)
  - [x] Update `.github/workflows/ci.yml` so push to `main` and pull requests run the lint gate on `ubuntu-24.04` after Go setup.
  - [x] Keep `permissions: contents: read`, `actions/checkout@v6`, `actions/setup-go@v6`, `go-version-file: go.mod`, and `cache: false` unless a documented lint mechanism requires a scoped change.
  - [x] Update `tools/cigate/ci_test.go` so it mechanically guards the new lint gate, its pinning/version markers, and the same denied CI patterns already enforced there.
  - [x] If using `golangci/golangci-lint-action`, update `tools/cigate/ci_test.go` deliberately: the current test only allows `actions/checkout@v6` and `actions/setup-go@v6`, so the new action must be treated as an explicit Story 6.1-approved exception with pinned version and no untrusted event interpolation.

- [x] Record lint evidence in release docs without claiming final release readiness (AC: 3)
  - [x] Update `docs/release-checklist.md` under CI trust gates or release-candidate gates with the lint command, result, version/pinning evidence, and isolation evidence.
  - [x] Update `docs/release_checklist_test.go` so it requires lint evidence and rejects blank or floating-version lint entries.
  - [x] Update `docs/behavior-matrices.md` and `docs/behavior_matrices_test.go` only if the matrix needs a new current row for lint evidence; keep claims tied to actual executable evidence.
  - [x] Update `docs/release-notes-v0.md` only if it currently lists release gates and would become stale by omitting lint. Keep v0 experimental, clean-room, redaction, dependency, and compatibility boundary language intact.
  - [x] Add `docs/provenance-log.md` entries only if copied, adapted, generated, or reference-derived material is introduced. Independently written command/evidence prose based on local output usually does not need a new external-source entry.

- [x] Verify the lint gate and standard-library isolation (AC: 1-3)
  - [x] Run the selected local lint command and record the exact command/result in `docs/release-checklist.md`.
  - [x] `GOCACHE=/tmp/dib-go-build go test ./tools/cigate`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./docs`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go vet ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
  - [x] `git diff --check`
  - [x] `rg -n "^(require|replace|toolchain)\\b" go.mod` returns no output unless an approved architecture update is part of the same change.
  - [x] `test ! -e go.sum` remains true unless an approved architecture update allows root module dependencies.
  - [x] If an external linter binary/action is selected, verify no Go source imports its module path and no root module dependency files were generated.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`; Story 6.1 was `backlog` and Epic 6 was `backlog` at story creation time. This workflow moved Epic 6 to `in-progress` and Story 6.1 to `ready-for-dev`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`; Story 6.1 is the primary story spec. [Source: `_bmad-output/planning-artifacts/epics.md#Story-6.1-Add-an-Isolated-Linter-Gate`]
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`; FR23 requires CI lint failures to fail CI, pinned/reproducible isolated linter tooling, and continued `tools/depgate` proof. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-23-Run-an-isolated-lint-gate`]
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`; core PR gates now include a pinned lint command selected by Story 6.1, and release evidence must record lint evidence. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- Loaded sprint change proposal: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-06-12.md`; it requires a short tooling review and allows an isolated pinned CI action or downloaded binary if no external import enters the root module graph. [Source: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-06-12.md#Conflict-Analysis`]
- Loaded implementation readiness update: `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-12.md`; Epic 6 sequence starts with Story 6.1 and does not require runtime API changes.
- No UX artifact discovered; Dib V1 has no browser UI or frontend surface.
- No `project-context.md` file exists under the project root.

### Current Repository State

- Baseline commit at story creation: `7cfdfcaf62b9f344eb4258eb03fc95f6e4783ac6` (`docs(bmad): add epic 6 release hardening plan`).
- Root `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`; no root `go.sum` was discovered.
- `.github/workflows/ci.yml` currently runs on `ubuntu-24.04` with `actions/checkout@v6`, `actions/setup-go@v6`, `go-version-file: go.mod`, `cache: false`, `go version`, `go test ./...`, `go vet ./...`, and `go run ./tools/depgate`. It has no lint step today.
- There is no `.golangci.yml`, `.golangci-lint-version`, `Makefile`, `docs/testing.md`, or root `README.md` at story creation time.
- `docs/release-checklist.md` currently records test, vet, dependency-gate, race, fuzz, docs/examples, provenance, compatibility, migration, runner/action, Go version, and dependency evidence. It does not record lint evidence.
- `docs/release-notes-v0.md` currently lists `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `go test -race ./...` as release gates; it will become stale if lint is added to release-gate docs but not reflected there.

### Current UPDATE File Intelligence

- `.github/workflows/ci.yml`: add the lint gate here; preserve push/main and pull_request triggers.
- `tools/cigate/ci_test.go`: `TestCIWorkflowRunsTrustGates` must expect the new lint gate. `TestCIWorkflowUsesOnlyTrustedStaticSteps` currently denies all actions except `actions/checkout@v6` and `actions/setup-go@v6`; update this deliberately if an external lint action is selected. `TestReleaseChecklistCapturesRequiredEvidence` must expect lint evidence after the checklist is updated.
- `docs/release-checklist.md`: add lint command, result, version/pinning, and isolation evidence. Do not claim final tag approval.
- `docs/release_checklist_test.go`: add lint evidence checks alongside the existing required passing gate checks. Keep the no Docker/Kubernetes/binary deployment/shell completion/man page assumptions.
- `docs/behavior-matrices.md`: dependency gate evidence currently says CI runs test, vet, and depgate. Update only where lint evidence becomes a current trust gate.
- `docs/behavior_matrices_test.go`: add or adjust markers if `docs/behavior-matrices.md` gets a lint row or changed dependency-gate wording.
- `docs/provenance-log.md`: add entries only for external reference-derived content. A config file written from project decisions may not need a provenance entry; copied/adapted external config snippets do.
- `tools/depgate/main.go`: already runs `go list -deps -test -e -json -buildvcs=false ./...` with `GOWORK=off`, reports non-standard imports deterministically, exits `1` for dependency violations, and exits `2` for execution failures. Do not replace this tool.
- `tools/depgate/main_test.go`: proves stdlib-only fixtures pass, external runtime/test imports fail, workspace mode is disabled, dependency errors are reported, and execution failures are separated from dependency violations.

### Git Intelligence

- Recent commits:
  - `7cfdfca docs(bmad): add epic 6 release hardening plan`
  - `ee6f2b8 chore(bmad): add automator and creative skills`
  - `561dd78 docs: add epic 1 retrospective`
  - `6dba659 docs: add epic 5 retrospective`
  - `d5ce41 docs(story-5.4): record release readiness evidence`
- Relevant recent patterns:
  - Epic 5 established docs/test evidence as first-class release artifacts; extend these tests rather than relying only on prose.
  - Story 5.4 review removed Git-dependent test behavior from docs tests. Do not add tests that require mutable `HEAD`, network access, or a live Git checkout during `go test ./docs`.
  - Existing CI tests prefer static string checks over YAML parsing. Match that pattern unless a small parser is clearly justified.

### Architecture Guardrails

- Dib releases are Go module tags, not binary deployments. Do not add Docker, Kubernetes, generated shell completion, generated man pages, or app scaffolding. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- Lint tooling must be pinned, reproducible, and isolated as development or CI tooling. It must not enter Dib runtime imports or the root module's checked package imports without an approved architecture update. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- `go run ./tools/depgate` is the required dependency gate. Do not invent a second dependency scanner or replace it with ad hoc `go list` shell pipelines. [Source: `_bmad-output/planning-artifacts/architecture.md#Implementation-Handoff`]
- `docs/testing.md` owns local verification, lint, coverage, fuzz, race, dependency-gate, and release-candidate validation guidance. Create it in Story 6.1 if it is the clearest place to document the local lint command. [Source: `_bmad-output/planning-artifacts/architecture.md#Documentation-Organization`]
- CI failures block tagging. Waivers require owner, reason, expiry, and impact; open-ended waivers block release readiness. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- Dib must not claim source compatibility with Go `flag`, pflag, Cobra, or Viper. Preserve native-API and not-drop-in wording if release docs are touched. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#API-Contracts-Versioning-And-Dependency-Policy`]

### Library / Framework Requirements

- Go baseline is `go 1.26`; the official Go downloads page listed `go1.26.4` as a stable release on 2026-06-12. Keep root `go.mod` at `go 1.26`; do not add a root `toolchain` directive unless architecture changes. Source: https://go.dev/dl/
- Web research on 2026-06-12 found `golangci-lint` documented as a fast Go linter runner and `golangci/golangci-lint-action` documented as the official GitHub Action from its authors. The action README example uses `golangci/golangci-lint-action@v9` and a pinned `version` input, and its options include `version` and `version-file`. Sources: https://raw.githubusercontent.com/golangci/golangci-lint/main/README.md and https://raw.githubusercontent.com/golangci/golangci-lint-action/main/README.md
- The `golangci-lint` changelog listed `v2.12.2` released on 2026-05-06 during story creation research. If selecting `golangci-lint`, verify compatibility locally and pin an exact version or version-file rather than floating to latest. Source: https://raw.githubusercontent.com/golangci/golangci-lint/main/CHANGELOG.md
- Using `golangci-lint` is optional, not mandatory. A standard-library-only repository-local lint tool is lower supply-chain risk if it meaningfully satisfies FR23 and NFR11. If an external action/tool is selected, document why the added CI/development dependency is acceptable and isolated.

### Project Structure Notes

- Primary UPDATE targets: `.github/workflows/ci.yml`, `tools/cigate/ci_test.go`, `docs/release-checklist.md`, and `docs/release_checklist_test.go`.
- Likely NEW target: `docs/testing.md`, because architecture says it owns local verification and lint guidance and the file does not exist yet.
- Possible NEW targets depending on lint choice: `.golangci.yml`, `.golangci-lint-version`, `.gitattributes`, or `tools/lint/`.
- Possible UPDATE targets after evidence reconciliation: `docs/behavior-matrices.md`, `docs/behavior_matrices_test.go`, `docs/release-notes-v0.md`, `docs/provenance-log.md`.
- Avoid changes under `command/`, `flags/`, `config/`, or `examples/` unless lint exposes a real issue. If lint forces code changes, keep them mechanical and covered by existing tests.
- Do not add root `require`, `replace`, `toolchain`, or `go.sum` unless there is an explicit architecture/PRD update in the same story. The expected outcome is still a dependency-free root module.

### Anti-Patterns To Avoid

- Do not use `latest`, `stable`, unpinned curl installers, or unpinned GitHub actions as the lint pinning model.
- Do not add the linter as a Go import in package code, tests, examples, or repository tools unless architecture changes.
- Do not add a second dependency gate or weaken `tools/depgate`.
- Do not record lint PASS evidence unless the exact lint command actually ran.
- Do not mark the release ready to tag just because lint passes; Story 6.2 coverage, Story 6.3 public docs, and Story 6.4 reconciliation remain open.
- Do not fold coverage validation into this story beyond avoiding conflicts. Coverage threshold policy and package-aware coverage command are Story 6.2.

### Validation Checklist Applied

- Story includes exact story ID/key, ready-for-dev status, role/action/benefit, BDD-derived acceptance criteria, and task mapping to ACs.
- Story identifies every likely UPDATE file and summarizes current behavior to preserve.
- Story includes previous work and git intelligence from the completed release-evidence story and Epic 6 planning commit.
- Story includes latest technical research relevant to current linter/tooling choices and cites source URLs.
- Story preserves dependency, clean-room, release-evidence, and CI guardrails.

### Story Completion Status

- Status set to `ready-for-dev`.
- Completion note: Ultimate context engine analysis completed - comprehensive developer guide created.

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- Red phase: `GOCACHE=/tmp/dib-go-build go test ./tools/lint` failed before `tools/lint/main.go` existed.
- Red phase: `GOCACHE=/tmp/dib-go-build go test ./tools/cigate` and `GOCACHE=/tmp/dib-go-build go test ./docs` failed after guard tests were updated and before workflow/docs evidence was added.
- Green/validation: lint, targeted guard packages, full tests, vet, depgate, diff hygiene, and dependency-file checks passed on 2026-06-12.

### Completion Notes List

- Selected a repository-local standard-library lint tool under `tools/lint` to avoid external linter packages, actions, root module dependencies, or floating installer versions.
- Added `go run ./tools/lint` as the deterministic local and CI lint gate; it reports non-`gofmt` Go files with sorted diagnostics.
- Documented the lint tooling review, local command, CI command, pinning model, and isolation evidence in release/testing docs without claiming final release readiness.
- Preserved root module isolation: no root `require`, `replace`, `toolchain`, `go.sum`, external linter action, or external linter Go import was added.

### File List

- `.github/workflows/ci.yml`
- `_bmad-output/implementation-artifacts/6-1-add-an-isolated-linter-gate.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/implementation-artifacts/tests/test-summary.md`
- `_bmad-output/story-automator/orchestration-6-20260612-155533.md`
- `docs/behavior-matrices.md`
- `docs/behavior_matrices_test.go`
- `docs/release-checklist.md`
- `docs/release-notes-v0.md`
- `docs/release_checklist_test.go`
- `docs/testing.md`
- `docs/testing_test.go`
- `tools/cigate/ci_test.go`
- `tools/lint/main.go`
- `tools/lint/main_test.go`

### Senior Developer Review (AI)

Reviewer: GPT-5 Codex
Date: 2026-06-12
Outcome: Approved after automatic fixes

#### Findings Fixed

- [x] [AI-Review][Medium] `tools/lint` walked BMAD and agent metadata directories, so future non-product Go snippets under `.agents`, `.codex`, or `_bmad` could fail the source lint gate. Fixed by excluding those metadata directories and adding regression coverage in `tools/lint/main_test.go`.
- [x] [AI-Review][Medium] `docs/release-checklist.md` recorded new Story 6.1 lint evidence beside an older exact tag commit, which made the evidence scope ambiguous. Fixed by recording that lint evidence was collected from the Story 6.1 working tree based on `7cfdfcaf62b9f344eb4258eb03fc95f6e4783ac6`, with final tag commit reconciliation left to release review.
- [x] [AI-Review][Medium] The Dev Agent File List omitted automation artifacts changed during Story 6.1 execution. Fixed by adding the test summary and orchestration files to the File List.

#### Review Validation

- Acceptance Criteria 1-3 cross-checked against `.github/workflows/ci.yml`, `tools/lint`, `tools/cigate`, `docs/testing.md`, `docs/release-checklist.md`, `docs/behavior-matrices.md`, `docs/release-notes-v0.md`, and root module dependency state.
- Local Go documentation check performed with `go doc go/format.Source`.
- `GOCACHE=/tmp/dib-go-build go run ./tools/lint` - PASS
- `GOCACHE=/tmp/dib-go-build go test ./tools/lint` - PASS
- `GOCACHE=/tmp/dib-go-build go test ./tools/cigate` - PASS
- `GOCACHE=/tmp/dib-go-build go test ./docs` - PASS
- `GOCACHE=/tmp/dib-go-build go test ./...` - PASS
- `GOCACHE=/tmp/dib-go-build go vet ./...` - PASS
- `GOCACHE=/tmp/dib-go-build go run ./tools/depgate` - PASS
- `git diff --check` - PASS
- `rg -n "^(require|replace|toolchain)\\b" go.mod` - PASS; returned no output.
- `test ! -e go.sum` - PASS

### Change Log

- 2026-06-12: Added isolated standard-library lint gate, CI wiring, local testing guidance, release evidence, and static guard coverage for Story 6.1.
- 2026-06-12: Senior developer review fixed lint traversal boundaries, clarified Story 6.1 release-evidence scope, completed the File List, and approved the story.
