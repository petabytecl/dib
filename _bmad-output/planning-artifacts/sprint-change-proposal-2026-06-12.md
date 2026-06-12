# Sprint Change Proposal: Release Hardening Gates and Usage Documentation

Status: Approved and applied
Date: 2026-06-12
Project: Dib
Prepared by: BMAD Correct Course

## 1. Change Trigger and Context

The V1 sprint implementation is locally marked complete in `_bmad-output/implementation-artifacts/sprint-status.yaml`: Epics 1-5, all stories, and all retrospectives are `done`.

The release-readiness pass surfaced three missing release-hardening requirements:

- A linter gate beyond `go vet`.
- Test coverage validation, not just test execution.
- Public usage documentation for adopters of the library.

Current evidence:

- `.github/workflows/ci.yml` runs `go test ./...`, `go vet ./...`, and `go run ./tools/depgate`.
- `docs/release-checklist.md` records test, vet, dependency gate, race, fuzz, docs, provenance, compatibility, and migration evidence, but does not record lint or coverage threshold evidence.
- The repository has `CONTRIBUTING.md` and domain docs, but no root `README.md` and no dedicated public usage guide.
- Baseline coverage measured on 2026-06-12 with `go test ./... -coverprofile=/tmp/dib-cover.out`:
  - `command`: 85.2% statements
  - `config`: 89.6% statements
  - `flags`: 85.0% statements
  - `tools/depgate`: 53.2% statements
  - total: 84.7% statements

## 2. Impact Assessment

### Epic Impact

Existing Epics 1-5 remain valid. No completed implementation behavior needs rollback or redesign.

This change extends the release criteria and adoption surface. It should not be folded silently into completed Epic 5 because Epic 5 is already framed as migration, compatibility, and release evidence for the existing gates. The recommended path is to add a new follow-up epic:

> Epic 6: Release Hardening and Public Usage Onboarding

Epic 6 should contain focused stories for linting, coverage validation, usage documentation, and release-evidence reconciliation.

### Artifact Impact

The change affects these planning artifacts:

- PRD: Add release quality gates and public usage docs as explicit functional/release requirements.
- Architecture: Update core PR gates, release-candidate gates, development-tool isolation policy, and docs ownership.
- Epics: Add Epic 6 and its stories.
- Sprint status: Regenerate after Epic 6 is accepted; do not hand-edit the current completed sprint status as part of this draft.
- GitHub issues: The board and labels are stale relative to local `sprint-status.yaml`; tracker reconciliation should be part of the Epic 6 planning pass.

## 3. Conflict Analysis

### Standard-Library Runtime Contract

The PRD already defines a development dependency as a tool, generator, linter, or test-only package that does not become a runtime dependency of Dib packages. A linter can be added without violating the runtime contract if it is isolated from library, test, example, and tool package imports inspected by `tools/depgate`.

Implementation stories must decide the linter mechanism after a short tooling review. Acceptable approaches include a pinned CI action or downloaded binary, provided no external import enters the root module graph without an approved architecture update.

### Coverage Gate Scope

A single aggregate threshold would be misleading because the public runtime packages already sit near or above 85%, while `tools/depgate` is lower and has different risk characteristics. Coverage validation should start with package-aware thresholds:

- Public runtime packages (`command`, `config`, `flags`) must meet a release threshold.
- Tooling packages must either meet a separate threshold or carry documented exceptions with targeted tests for critical paths.
- Docs and example packages with no statements should not affect the threshold.

### Public Documentation Scope

Current docs are strong for release evidence and compatibility boundaries, but they are not an adopter-facing entry point. Public usage documentation should explain import, basic command construction, flag parsing, config precedence, diagnostics, and release-gate expectations without copying third-party README structure or implying source compatibility with other libraries.

## 4. Recommended Path Forward

Recommended path: Direct adjustment with one new release-hardening epic.

Reasoning:

- The missing items are release-quality and adoption gaps, not contradictions in the implemented product behavior.
- Existing stories and evidence remain useful.
- A new epic keeps the completed V1 story history intact while making the added release criteria visible.
- The work has moderate planning impact and limited implementation risk if dev tooling remains isolated.

This proposal was approved on 2026-06-12 and applied to the PRD, architecture, epics, and sprint status artifacts.

## 5. Proposed Artifact Changes

### PRD Additions

Add two requirements under Functional Requirements or Release Evidence:

- FR-23: Dib release candidates must run an automated lint gate in CI. The linter must be isolated as development tooling and must not add runtime dependencies to Dib packages.
- FR-24: Dib release candidates must validate test coverage with a documented threshold policy. The coverage gate must report package-level evidence and explain any package-specific exceptions.
- FR-25: Dib must provide public usage documentation for library adopters, including install/import guidance, command construction, flag parsing, config source precedence, diagnostics, and clean-room compatibility boundaries.

Add success metrics:

- SM-6: CI fails when lint or coverage validation fails.
- SM-7: A new adopter can follow the README and usage docs to build a minimal multi-command CLI using Dib without reading implementation internals.

### Architecture Updates

Update core PR gates from:

```sh
go test ./...
go vet ./...
go run ./tools/depgate
```

To:

```sh
go test ./...
go vet ./...
go run ./tools/depgate
<lint command>
<coverage validation command>
```

Add architecture guidance:

- The linter tool must be pinned and reproducible.
- External linter tooling is allowed only as isolated development/CI tooling.
- The coverage validator should use standard Go coverage output where possible.
- `README.md` and `docs/testing.md` become release-relevant docs.
- Release evidence must record exact lint and coverage commands and results.

### Epic 6 Draft

#### Story 6.1: Add an Isolated Linter Gate

Acceptance criteria:

- Select and document the linter mechanism and isolation model.
- Add a deterministic local command for linting.
- Add the lint command to CI.
- Prove `go run ./tools/depgate` still passes.
- Record lint evidence in the release checklist.

#### Story 6.2: Add Coverage Validation

Acceptance criteria:

- Generate coverage with a reproducible command.
- Enforce package-aware thresholds for `command`, `config`, and `flags`.
- Define tooling-package coverage expectations or documented exceptions.
- Add coverage validation to CI.
- Record package-level coverage evidence in the release checklist.

#### Story 6.3: Publish Public Usage Documentation

Acceptance criteria:

- Add a root `README.md` with install/import, quickstart, package overview, and status.
- Add public usage documentation for commands, flags, config precedence, diagnostics, and release gates.
- Keep compatibility language behavior-scoped and clean-room compliant.
- Add or update runnable examples so `go test ./...` verifies documented usage where practical.
- Link existing docs without requiring users to read planning artifacts.

#### Story 6.4: Reconcile Release Evidence and Tracker State

Acceptance criteria:

- Update `docs/release-checklist.md` with lint, coverage, and public-docs evidence sections.
- Update release notes or release docs to mention new quality gates.
- Regenerate BMAD sprint status after Epic 6 is accepted.
- Bring GitHub board/issues/labels back in sync with the local sprint status and new Epic 6 scope.
- Leave an explicit note for any accepted gate waiver with owner, reason, expiry, and impact.

## 6. Risks and Mitigations

- Risk: Linter tooling introduces external dependencies contrary to the zero-runtime-dependency promise.
  - Mitigation: Keep the linter outside runtime/test/example imports; verify with `tools/depgate`; update architecture before any exception.
- Risk: Coverage threshold becomes a vanity metric.
  - Mitigation: Use package-aware thresholds and require package-level evidence, not only an aggregate number.
- Risk: Public docs overstate compatibility with familiar CLI libraries.
  - Mitigation: Reuse existing compatibility boundaries and keep claims behavior-scoped.
- Risk: GitHub issue labels remain stale.
  - Mitigation: Make tracker reconciliation an explicit story acceptance criterion after Epic 6 is approved.

## 7. Validation Performed During Correction

- Loaded BMAD correct-course checklist and applied the direct-adjustment workflow.
- Confirmed local sprint status marks Epics 1-5 and all retrospectives `done`.
- Inspected CI gates in `.github/workflows/ci.yml`.
- Inspected release checklist evidence.
- Verified no root `README.md` or `docs/testing.md` exists.
- Measured current coverage with `go test ./... -coverprofile=/tmp/dib-cover.out` and `go tool cover -func=/tmp/dib-cover.out`.
- Checked GitHub issue state and found the board issue/labels are stale relative to local sprint status.

## 8. Approval and Handoff

Approval received on 2026-06-12:

- Approve adding Epic 6: Release Hardening and Public Usage Onboarding.
- Approve updating PRD, architecture, epics, sprint status, and GitHub tracker state to include the new release-hardening scope.

Applied results:

- PRD updated with FR-23 through FR-25, NFR-11 through NFR-12, success metrics, risks, and release-quality gate decisions.
- Architecture updated with lint isolation, package-aware coverage validation, README/docs ownership, and release-evidence scope.
- Epics updated with Epic 6 and Stories 6.1 through 6.4.
- Sprint status updated with Epic 6 backlog entries.
- Implementation readiness update created at `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-12.md`.
- GitHub tracker updated with Epic 6 issues #39 through #44 and board issue #1.

Next BMAD workflow:

1. Run `bmad-create-story` for Story 6.1: Add an Isolated Linter Gate.
