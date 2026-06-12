---
baseline_commit: d6473cf
created: "2026-06-12"
---

# Story 5.4: Prove Release Readiness With Dependency And Provenance Evidence

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a release reviewer,
I want a complete release-readiness evidence package,
so that Dib's v0 module tag is backed by tests, dependency checks, compatibility documentation, migration guidance, and clean-room provenance.

## Requirements Trace

- FR18: compatibility boundaries must remain adopter-facing and must not claim source compatibility with Go `flag`, pflag, Cobra, or Viper.
- FR19: migration examples must remain executable and standard-library-only.
- FR20: behavior matrices must remain aligned with package tests, docs tests, examples, diagnostics, redaction, and config precedence.
- FR21: the release checklist must include dependency-check evidence and the runtime dependency rule must pass.
- FR22: parser fuzz targets are the parser hardening evidence when parser behavior changes or when release-candidate fuzz evidence is required.
- NFR1: runtime packages, tests, examples, and tools must remain standard-library-only unless the architecture changes.
- NFR4: release evidence, help/usage/diagnostics/source reports, and docs tests must remain deterministic enough for repeatable review.
- NFR7: compatibility language must stay limitation-framed, not source-compatible or drop-in.
- NFR8: docs, examples, diagnostics, source reports, and release evidence must not expose raw sensitive values.
- NFR9: Dib V1 requires Go 1.26 or newer; release evidence must keep `go.mod`, CI, release guidance, and docs aligned.

## Acceptance Criteria

1. Given Dib is preparing a v0 module release, when `docs/release-checklist.md` is completed, then it records exact commit, test, vet, dependency-gate, race-test, docs/examples, provenance, compatibility, and migration evidence, and unresolved checklist items block release readiness.
2. Given dependency evidence is central to adoption, when release checks run, then `go run ./tools/depgate` proves zero external imports for library, test, example, and tool packages unless the architecture has been updated, and dependency-gate output is recorded as release evidence.
3. Given v0 may include future breaking changes, when release notes are prepared, then they state the v0 experimental API status, and they still preserve correctness, redaction, clean-room, dependency, and release-gate expectations.
4. Given compatibility and migration docs are adopter-facing contracts, when release readiness is reviewed, then examples, compatibility boundaries, behavior matrices, provenance log, diagnostics/errors docs, and config precedence docs align with implemented behavior, and any waivers include owner, reason, expiry, and impact.
5. Given release evidence must be reproducible, when a reviewer reruns the documented commands, then `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `go test -race ./...` pass for the tagged commit, and no release process assumes binary deployment, Docker, Kubernetes, generated shell completion, or generated man pages.

## Tasks / Subtasks

- [x] Confirm preconditions and read UPDATE files before editing (AC: 1-5)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 5 `in-progress`, Stories 5.1 through 5.3 `done`, and Story 5.4 `ready-for-dev`.
  - [x] Confirm current `HEAD` is `d6473cf` (`docs(story-5.3): consolidate adoption evidence`) or intentionally account for newer user changes.
  - [x] Read these UPDATE files completely before editing: `docs/release-checklist.md`, `docs/release_checklist_test.go`, `docs/behavior-matrices.md`, `docs/behavior_matrices_test.go`, `docs/compatibility.md`, `docs/compatibility_test.go`, `docs/provenance-log.md`, `docs/config-precedence.md`, `docs/diagnostics-and-errors.md`, `.github/workflows/ci.yml`, `go.mod`, `tools/depgate/main.go`, and `tools/depgate/main_test.go`.
  - [x] Inspect current release-relevant tests and examples with `rg -n "^func (Test|Fuzz|Example)" flags command config docs examples/migration tools/depgate`.
  - [x] Do not add runtime APIs, compatibility adapters, package-global helpers, `/cmd` scaffolding, generated assets, Docker/Kubernetes files, generated shell completion, generated man pages, or process-owning release automation.

- [x] Complete `docs/release-checklist.md` as reproducible v0 release evidence (AC: 1, 2, 5)
  - [x] Fill release identity with exact Go module tag candidate, exact commit hash, owner, date, and reviewer.
  - [x] Record Go version alignment: `go.mod` version, CI `go-version-file: go.mod`, release guidance version, docs version references, and drift review result.
  - [x] Run and record exact outcomes for `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `go test -race ./...`.
  - [x] Record runner/action evidence from `.github/workflows/ci.yml`: `ubuntu-24.04`, `actions/checkout@v6`, `actions/setup-go@v6`, and `go-version-file: go.mod`.
  - [x] Record dependency evidence: root `go.mod` has no `require`, `replace`, or `toolchain` directives; root `go.sum` is absent; `go run ./tools/depgate` passed; fixture-local external modules are isolated under `tools/depgate/testdata/` and are intentional negative fixtures.
  - [x] Keep unresolved checklist items blocking. Do not write "ready to tag", "approved", or equivalent unless every required field has actual evidence and any waivers are complete.

- [x] Record release-candidate parser hardening evidence without overstating scope (AC: 1, 5)
  - [x] If parser behavior changed in or after Story 5.3, run and record `go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags`, `go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags`, and `go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags`.
  - [x] If parser behavior did not change, record the release decision explicitly: corpus-backed fuzz targets exist and normal `go test ./...` runs seed corpora, while mutation fuzzing is either run for this release candidate or waived with owner, reason, expiry, and impact.
  - [x] Do not imply long-running fuzz campaigns occurred unless the commands were actually run and their outcomes were captured.

- [x] Reconcile adopter-facing docs against implementation evidence (AC: 3, 4)
  - [x] Review `docs/behavior-matrices.md` against current package tests, docs tests, examples, and dependency-gate files. Any current row that lacks evidence must be fixed, downgraded to deferred, or backed by a focused test before release readiness is claimed.
  - [x] Review `docs/compatibility.md` for source-compatibility drift. It must keep Dib framed as a clean-room native Go API, not a source-compatible clone, drop-in replacement, or framework compatibility layer.
  - [x] Review `examples/migration/` by running `go test ./examples/migration` and confirming examples still demonstrate Dib-native APIs with explicit instances, injected inputs, typed errors, deterministic rendering, and redaction-safe reports.
  - [x] Review `docs/config-precedence.md` for the exact precedence order `explicit setter > flag binding > env > JSON > default`.
  - [x] Review `docs/diagnostics-and-errors.md` for exact source-label vocabulary: `default`, `explicit setter`, `flag binding`, `env`, and `JSON`; confirm release evidence does not introduce lowercase `json` or synonyms as canonical labels.
  - [x] Review `docs/provenance-log.md` and `docs/clean-room-policy.md` before adding or updating any reference-derived content.

- [x] Add or update docs tests to prevent release-evidence drift (AC: 1-5)
  - [x] Update `docs/release_checklist_test.go` so it now expects completed Story 5.4 fields that are actually filled, exact required commands, Go version alignment, dependency evidence, v0 experimental status, waiver shape, and no binary-deployment assumptions.
  - [x] Preserve or replace the previous "Story 5.4 outcomes unfilled" guard so the new guard prevents partial release readiness claims instead of blocking completion forever.
  - [x] Add coverage that final review cannot claim readiness when any required evidence field remains blank or any waiver lacks owner, reason, expiry, and impact.
  - [x] Extend `docs/compatibility_test.go` or `docs/behavior_matrices_test.go` only where release-readiness wording changes need mechanical protection.

- [x] Prepare v0 release notes or release guidance in the correct location (AC: 3)
  - [x] If no release-notes file exists, add a minimal docs artifact such as `docs/release-notes-v0.md` only if it is needed to satisfy the release-notes AC; keep it focused on v0 experimental API status, correctness/redaction/dependency/clean-room expectations, compatibility boundaries, and migration guidance pointers.
  - [x] Do not create generated changelog tooling or binary packaging. Dib releases are Go module tags.
  - [x] Link release guidance from `docs/release-checklist.md` if the checklist records release guidance version/status.

- [x] Verify and record final release readiness evidence (AC: 1-5)
  - [x] `GOCACHE=/tmp/dib-go-build go test ./docs`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./examples/migration`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go vet ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
  - [x] `GOCACHE=/tmp/dib-go-build go test -race ./...`
  - [x] Fuzz commands listed above, or a complete waiver entry with owner, reason, expiry, and impact.
  - [x] `git diff --check`
  - [x] `rg -n "(?i)drop-in|source-compatible|clone API|framework compatibility layer|compatible replacement|release readiness is complete|ready to tag|tag readiness complete" docs examples/migration`
  - [x] Confirm the search above finds only explicit limitation or boundary language, not positive compatibility or premature release-readiness claims.
  - [x] `rg -n "^(require|replace|toolchain)\\b" go.mod` returns no output.
  - [x] `test ! -e go.sum`

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`; Story 5.4 was `backlog` at story creation time and Epic 5 was already `in-progress`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`; Story 5.4 is the primary story spec. [Source: `_bmad-output/planning-artifacts/epics.md#Story-5.4-Prove-Release-Readiness-With-Dependency-And-Provenance-Evidence`]
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`; release evidence must record test, vet, dependency-gate, race-test, docs/examples, runner/action version, provenance, compatibility, and migration status. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- Loaded PRD shards from `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/`; FR18 through FR21 are the controlling requirements, with FR22 relevant to fuzz evidence. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#5.5-Validation-And-Release-Evidence`]
- No UX artifact discovered; Dib V1 has no browser UI or frontend surface. [Source: `_bmad-output/planning-artifacts/architecture.md#Frontend-Architecture`]
- No `project-context.md` file exists under the project root.
- Loaded previous story intelligence from `_bmad-output/implementation-artifacts/5-3-consolidate-behavior-matrices-into-adoption-evidence.md`.
- Web research completed 2026-06-12 against the official Go downloads page. The page lists `go1.26.4` under stable versions. Keep root `go.mod` at `go 1.26`; do not add a `toolchain` directive unless architecture changes. Source: https://go.dev/dl/

### Current Repository State

- Baseline commit at story creation: `d6473cf` (`docs(story-5.3): consolidate adoption evidence`).
- Root `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`; no root `go.sum` was discovered.
- `.github/workflows/ci.yml` runs on `ubuntu-24.04` and uses `actions/checkout@v6`, `actions/setup-go@v6` with `go-version-file: go.mod`, then `go test ./...`, `go vet ./...`, and `go run ./tools/depgate`.
- Existing dirty/untracked BMAD installer/config, `.agents/`, `.codex/`, `.idea/`, and story-automator files exist in the worktree. Do not revert or normalize them.
- `docs/release-checklist.md` is still a release-candidate checklist. Story 5.3 intentionally left exact command outcomes and tagging readiness blank for Story 5.4.
- `docs/release_checklist_test.go` currently enforces that Story 5.4 outcome fields are unfilled. Story 5.4 must update that test when it fills real release evidence.

### Current UPDATE File Intelligence

- `docs/release-checklist.md` currently has empty release identity, Go version alignment, command outcome, race/fuzz, dependency evidence, and final review fields. It already lists the required gates and Story 5.3 evidence inputs.
- `docs/release_checklist_test.go` currently checks release-checklist headings, required Story 5.3 evidence input phrases, and blank Story 5.4 outcome fields. It must be rewritten or extended after the checklist is completed so it guards completed evidence rather than forbidding it.
- `docs/behavior-matrices.md` marks "Release-candidate readiness" as Story 5.4 and `deferred`. After Story 5.4, update only if release readiness is actually supported by captured evidence; otherwise keep unresolved pieces blocked or waived.
- `docs/behavior_matrices_test.go` checks the release-readiness row stays Story 5.4/deferred and rejects premature "ready to tag" style language. Update this only if the evidence package legitimately changes the matrix status.
- `docs/compatibility.md` explicitly says it does not claim release readiness and that Story 5.4 owns release-readiness evidence. Preserve source-compatibility boundary language.
- `docs/compatibility_test.go` guards compatibility boundary phrases, evidence links, migration example links, and ambiguous compatibility positioning.
- `docs/provenance-log.md` contains inspiration-only entries for Go fuzzing/testing docs and Story 5.1/5.2 compatibility/example references. Add entries before acceptance if Story 5.4 uses any new external reference material for release notes, compatibility docs, examples, fixtures, generated content, or evidence prose.
- `docs/config-precedence.md` is the canonical precedence authority: `explicit setter`, `flag binding`, `env`, `JSON`, `default`.
- `docs/diagnostics-and-errors.md` owns programmatic error, diagnostic, source-label, and redaction vocabulary. It fixes source labels to `default`, `explicit setter`, `flag binding`, `env`, and `JSON`.
- `.github/workflows/ci.yml` is the CI evidence source. It has no race or fuzz step today; those are release-candidate local gates unless CI is intentionally expanded.
- `tools/depgate/main.go` runs `go list -deps -test -e -json -buildvcs=false ./...` with `GOWORK=off`, reports non-standard imports deterministically, exits `1` for dependency violations, and exits `2` for execution failures.
- `tools/depgate/main_test.go` proves stdlib-only fixtures pass, external runtime/test imports fail, workspace mode is disabled, dependency errors are reported, and execution failures are separate from dependency violations.

### Previous Story Intelligence

- Story 5.3 completed `docs/behavior-matrices.md` as the consolidated adoption-evidence matrix and added `docs/behavior_matrices_test.go`.
- Story 5.3 updated `docs/release-checklist.md` with Story 5.3 evidence inputs while intentionally leaving exact release-candidate outcomes for Story 5.4.
- Story 5.3 added `docs/release_checklist_test.go` specifically to prevent Story 5.3 from pre-filling Story 5.4 outcomes. Story 5.4 must now replace that guard with completed-evidence validation.
- Story 5.3 did not use new external reference material for matrix content; `docs/provenance-log.md` was not changed.
- Story 5.3 verification passed: `go test ./docs`, `go test ./examples/migration`, `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `git diff --check`, compatibility boundary search, module directive scan, and root `go.sum` absence check.
- The Senior Developer Review for Story 5.3 auto-fixed a missing File List entry for `docs/release_checklist_test.go`; keep the Story 5.4 File List exact.

### Git Intelligence

- Recent commits:
  - `d6473cf docs(story-5.3): consolidate adoption evidence`
  - `8f69699 docs(story-5.2): add migration examples`
  - `e6eccf4 docs(story-5.1): publish compatibility boundaries`
  - `a718d1d docs: add epic 4 retrospective`
  - `ca11b27 feat(story-4.5): report config provenance`
- Relevant recent patterns:
  - Epic 5 changes are docs/tests/examples centered; avoid runtime changes unless evidence reconciliation exposes a real implementation gap.
  - Story 5.1 established compatibility-boundary docs/tests; preserve limitation framing.
  - Story 5.2 established executable migration examples under `examples/migration/`; cite and test them rather than rewriting them.
  - Story 5.3 established docs drift tests for behavior matrices and release-checklist placeholders; update those tests in place where the release evidence changes.

### Architecture Guardrails

- Dib releases are Go module releases, not binary deployments. Do not add Docker, Kubernetes, binary packaging, generated shell completion, generated man pages, or app scaffolding. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- Release evidence must be tied to the exact commit and must include test, vet, dependency-gate, race-test, docs/examples, runner/action version, provenance, compatibility, and migration status. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- CI failures block tagging. Waivers require owner, reason, and expiry; Story 5.4 AC also requires impact. [Source: `_bmad-output/planning-artifacts/architecture.md#Infrastructure-Deployment`]
- `go run ./tools/depgate` is the required dependency gate once `tools/depgate/` exists; the temporary `go list` command is no longer acceptable as release-candidate evidence. [Source: `_bmad-output/planning-artifacts/architecture.md#Implementation-Handoff`]
- The dependency gate must inspect all non-tool packages included by `go test ./...`, including package tests and examples; tool packages must also remain standard-library-only unless architecture changes. [Source: `_bmad-output/planning-artifacts/architecture.md#Development-Workflow-Integration`]
- Record provenance for copied/generated/reference-derived artifacts before acceptance. Prefer independently written release evidence based on local command outputs and existing Dib docs. [Source: `_bmad-output/planning-artifacts/architecture.md#Clean-Room-Provenance-Enforcement`]
- Dib must not claim source compatibility with Go `flag`, pflag, Cobra, or Viper. It offers a native Dib API with familiar concepts and documented differences. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#API-Contracts-Versioning-And-Dependency-Policy`]
- Redaction docs and examples must use only the architecture-owned fake sensitive corpus: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Fixture-Patterns`]

### Library / Framework Requirements

- Go baseline is `go 1.26`; official Go downloads listed `go1.26.4` as stable on 2026-06-12. Record actual local `go version` output in the release checklist if useful, but do not pin patch-level tooling through a root `toolchain` directive unless architecture changes.
- Use only the Go standard library for runtime packages, tests, examples, and repo tooling unless architecture explicitly changes.
- Keep `GOWORK=off` behavior in dependency-gate evidence; the current gate already disables workspace mode internally.
- Prefer docs tests using standard `testing`, `os.ReadFile`, `regexp`, `strings`, and filesystem checks, matching `docs/compatibility_test.go`, `docs/behavior_matrices_test.go`, and `docs/release_checklist_test.go`.

### Project Structure Notes

- Primary UPDATE target: `docs/release-checklist.md`.
- Primary test UPDATE target: `docs/release_checklist_test.go`.
- Possible UPDATE targets after reconciliation: `docs/behavior-matrices.md`, `docs/behavior_matrices_test.go`, `docs/compatibility.md`, `docs/compatibility_test.go`, `docs/provenance-log.md`.
- Possible NEW target: `docs/release-notes-v0.md` or similar, only if needed to satisfy v0 release-notes evidence.
- Avoid changes under `command/`, `flags/`, `config/`, `examples/migration/`, or `tools/depgate/` unless release reconciliation finds a genuine unsupported claim that is best fixed with focused executable evidence.
- No UX files exist and no UI work applies.

### Anti-Patterns To Avoid

- Do not invent a second dependency scanner, shell script, Makefile gate, or CI-only dependency check. Use `go run ./tools/depgate`.
- Do not record command outcomes unless the commands actually ran in this story. If a command is skipped, record it as a blocking gap or waiver with owner, reason, expiry, and impact.
- Do not claim "release readiness complete", "ready to tag", or "approved" while any checklist item is blank or any waiver is incomplete.
- Do not add positive compatibility claims such as "compatible replacement", "drop-in", "source-compatible", "clone API", or "framework compatibility layer" except when framed as explicit limitations.
- Do not treat Story 5.3 adoption evidence as final release-candidate evidence. Story 5.3 supplied inputs; Story 5.4 must record exact outcomes.

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- 2026-06-12: Confirmed initial sprint preconditions, current `HEAD` `d6473cf68a59e5abcf1baf58c0515d7eeeb34626`, local `go version go1.26.4 linux/amd64`, root `go.mod` without `require`/`replace`/`toolchain`, and absent root `go.sum`.
- 2026-06-12: Red phase confirmed with `GOCACHE=/tmp/dib-go-build go test ./docs` failing against unfilled Story 5.4 release evidence and missing `docs/release-notes-v0.md`.
- 2026-06-12: Final validation passed: `go test ./docs`, `go test ./examples/migration`, `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `go test -race ./...`, all with `GOCACHE=/tmp/dib-go-build`.
- 2026-06-12: Parser fuzz validation passed for `FuzzParse`, `FuzzParseBoundary`, and `FuzzParseShortGroups` with `-fuzztime=5s`.
- 2026-06-12: Hygiene checks passed: `git diff --check`, compatibility-boundary search reviewed as limitation/test-guard language only, module directive scan returned no output, and `test ! -e go.sum` exited 0.

### Completion Notes List

- Completed the v0 release evidence package for tag candidate `v0.1.0` at exact commit `d5ce41ce693413b88df95e644eb4358702ae205e`.
- Replaced Story 5.4 placeholder tests with completed-evidence guards for identity, commands, Go version alignment, dependency evidence, v0 status, waiver shape, and release-scope boundaries.
- Added `docs/release-notes-v0.md` with experimental v0 status, Go 1.26+ guidance, release gates, compatibility boundaries, and migration pointers.
- Reconciled the behavior matrix and compatibility evidence link so release-candidate evidence is current without claiming tag approval.
- No new external reference material was used for Story 5.4 release evidence or release notes; no provenance-log update was required.
- Senior developer review auto-fixed the release-checklist docs test so it no longer shells out to `git` or requires the recorded release commit to match mutable repository `HEAD`.

### File List

- `_bmad-output/implementation-artifacts/5-4-prove-release-readiness-with-dependency-and-provenance-evidence.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `docs/behavior-matrices.md`
- `docs/behavior_matrices_test.go`
- `docs/compatibility.md`
- `docs/release-checklist.md`
- `docs/release_checklist_test.go`
- `docs/release-notes-v0.md`

### Change Log

- 2026-06-12: Completed Story 5.4 release-readiness evidence package, added v0 release notes, updated docs drift tests, ran all required validation gates, and moved story to review.
- 2026-06-12: Senior developer review fixed deterministic docs-test issue, removed unrelated completion note, reran validation gates, and marked story done.

## Senior Developer Review (AI)

Reviewer: GPT-5 Codex on 2026-06-12

Outcome: Approve after automatic fixes.

Findings fixed:

- MEDIUM: `docs/release_checklist_test.go` invoked `git rev-parse HEAD` during `go test ./docs`, which made the docs test depend on a live Git checkout and on `HEAD` not changing after the story is committed. Fixed by keeping the exact-commit format guard while removing the Git subprocess and mutable-HEAD equality.
- LOW: The completion notes included an unrelated "Ultimate context engine" entry that was not supported by this story's tasks or file list. Removed the note from the Dev Agent Record.

Validation:

- `GOCACHE=/tmp/dib-go-build go test ./docs`
- `GOCACHE=/tmp/dib-go-build go test ./examples/migration`
- `GOCACHE=/tmp/dib-go-build go test ./...`
- `GOCACHE=/tmp/dib-go-build go vet ./...`
- `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
- `GOCACHE=/tmp/dib-go-build go test -race ./...`
- `GOCACHE=/tmp/dib-go-build go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags`
- `GOCACHE=/tmp/dib-go-build go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags`
- `GOCACHE=/tmp/dib-go-build go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags`
- `git diff --check`
- Compatibility boundary search reviewed; hits are limitation or test-guard language only.
- `rg -n "^(require|replace|toolchain)\b" go.mod` returned no output.
- `test ! -e go.sum`
