---
baseline_commit: a718d1d
created: "2026-06-12"
---

# Story 5.1: Publish Compatibility Boundaries For Familiar CLI Concepts

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a developer evaluating Dib,
I want a clear compatibility table for Go `flag`, pflag, Cobra, and Viper concepts,
so that I can decide whether Dib fits without assuming source compatibility.

## Requirements Trace

- FR17: compatibility prose must respect the clean-room source policy and provenance requirements.
- FR18: developers can see which Go `flag`, pflag, Cobra, and Viper behaviors Dib supports, narrows, omits, or intentionally changes.
- FR20: compatibility claims must trace to implemented behavior matrices, tests, examples, or documented exclusions where practical.
- FR21: adoption evidence must preserve the standard-library dependency claim through the dependency gate.
- NFR1: runtime packages, tests, examples, and tooling remain standard-library-only unless architecture changes.
- NFR7: familiar behavior must be documented as supported, narrowed, omitted, or intentionally different.

## Acceptance Criteria

1. Given Dib supports familiar CLI and config concepts, when `docs/compatibility.md` is written, then it documents supported, narrowed, omitted, and intentionally different V1 behavior for Go `flag`, pflag, Cobra, and Viper, and it never describes Dib as source-compatible or a drop-in replacement.
2. Given intentional differences can affect migration, when a difference is documented, then the docs include the user-facing reason and link or trace to behavior matrices, examples, or tests where practical.
3. Given clean-room boundaries apply to compatibility prose, when compatibility docs reference inspiration projects, then references are classified in `docs/provenance-log.md` where required, and copied examples, fixtures, internal names, or file organization are not introduced.
4. Given compatibility docs are adoption evidence, when verification runs, then docs are checked for stale claims against implemented behavior where practical, and `go test ./...` and `go run ./tools/depgate` pass.

## Tasks / Subtasks

- [x] Confirm preconditions and read UPDATE files before editing (AC: 1-4)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 5 `in-progress` and Story 5.1 `ready-for-dev`.
  - [x] Confirm current `HEAD` is `a718d1d` (`docs: add epic 4 retrospective`) or intentionally account for newer user changes.
  - [x] Verify `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or `go.sum`.
  - [x] Read `docs/clean-room-policy.md` and `docs/provenance-log.md` completely before writing compatibility prose.
  - [x] Read `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, `docs/config-precedence.md`, and `docs/release-checklist.md` completely before creating compatibility claims.
  - [x] Read current package docs for `command/`, `flags/`, and `config/` to confirm public-surface terminology.
  - [x] Do NOT add source-compatible adapters, root facade APIs, package-global helpers, `/cmd` scaffolding, generated completion/manpage assumptions, new config formats, migration examples, or runtime package code.

- [x] Create `docs/compatibility.md` as the adopter-facing compatibility boundary document (AC: 1, 2)
  - [x] State up front that Dib is a clean-room native Go API with familiar concepts, not a source-compatible clone, drop-in replacement, or framework compatibility layer.
  - [x] Include one table with columns for inspiration surface, supported familiar concepts, narrowed behavior, omitted behavior, intentionally different behavior, user-facing reason, and Dib evidence.
  - [x] Cover Go `flag`: explicit flag sets, typed values, defaults, custom parsing concepts, parse errors, and testable argument parsing; make clear Dib omits `flag.*` source compatibility and package-global `CommandLine` helpers.
  - [x] Cover pflag: long flags, one-character shorthands, shorthand groups, no-option defaults, repeated/custom values, `--`, interspersed args, and opt-in normalization; make clear Dib omits source-compatible pflag APIs, broader extension surface, and automatic `--no-*` aliases.
  - [x] Cover Cobra: nested command routing, aliases, local/inherited flags, deterministic help/usage, and caller-controlled errors; make clear Dib omits `*cobra.Command` compatibility, shell completion, manpage generation, scaffolding, suggestions, and process lifecycle ownership.
  - [x] Cover Viper: defaults, explicit setters, flag bindings, env bindings, JSON readers/paths, precedence, typed getters, and source reports; make clear Dib omits package-global singleton config, non-JSON formats, remote stores, live reload, aliases, and reflection-heavy struct decoding.
  - [x] Include a concise "How to read this table" section defining `supported`, `narrowed`, `omitted`, and `intentionally different`.
  - [x] Keep the document about behavior boundaries only. Do not include migration walkthroughs; Story 5.2 owns examples.

- [x] Trace every compatibility claim to current Dib evidence (AC: 2, 4)
  - [x] Link or reference `docs/behavior-matrices.md` rows for flags, commands, config, diagnostics, provenance, redaction, determinism, and dependency evidence.
  - [x] Link or reference `docs/config-precedence.md` for the exact source precedence: `explicit setter > flag binding > env > JSON > default`.
  - [x] Link or reference `docs/diagnostics-and-errors.md` for typed error, diagnostic, source-label, and redaction vocabulary.
  - [x] Where evidence is intentionally deferred, state that it is deferred rather than implying support. Current deferred areas include compatibility tables before this story, migration examples before Story 5.2, consolidated release evidence before Stories 5.3 and 5.4, callback invocation, shell completion, manpages, and scaffolding.
  - [x] Do not cite rendered diagnostic strings as the only programmatic contract; cite structured tests, typed errors, snapshots, source reports, and behavior matrix rows.

- [x] Preserve clean-room provenance and wording boundaries (AC: 3)
  - [x] Update `docs/provenance-log.md` with Story 5.1 inspiration-only entries if current public Go `flag`, pflag, Cobra, or Viper documentation is used while writing `docs/compatibility.md`.
  - [x] If reusing the existing Initial Entries is enough, explicitly confirm they cover every external reference used by this story and update affected artifacts or notes if needed.
  - [x] Do not copy examples, fixtures, prose structure, README layout, internal names, source-derived names, or test cases from inspiration projects.
  - [x] Use Dib-owned wording and Dib-owned examples only. Prefer links and source names over copied public-doc excerpts.
  - [x] Avoid prohibited positioning terms in compatibility prose: `drop-in replacement`, `source-compatible`, `clone API`, and `framework compatibility layer`, except when saying Dib is not those things.

- [x] Add documentation freshness checks where practical (AC: 4)
  - [x] Search `docs/compatibility.md` for prohibited positive claims such as "drop-in", "source-compatible", and "compatible replacement"; allow only negative boundary statements.
  - [x] Search `docs/compatibility.md` for stale or unsupported claims against current `docs/behavior-matrices.md` rows.
  - [x] Confirm `docs/compatibility.md` does not claim Story 5.2 migration examples, Story 5.3 matrix consolidation, or Story 5.4 release readiness are complete.
  - [x] If adding a mechanical docs test is low-risk, place it where existing tooling can run with `go test ./...`; otherwise document the exact manual `rg` checks in the completion notes.

- [x] Verify the story implementation (AC: 1-4)
  - [x] `GOCACHE=/tmp/dib-go-build go test ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go vet ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
  - [x] `git diff --check`
  - [x] `rg -n "(?i)drop-in|source-compatible|clone API|framework compatibility layer" docs/compatibility.md docs/clean-room-policy.md`
  - [x] Confirm the search above finds only explicit boundary language, not positive compatibility claims.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` exists.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`; Story 5.1 is the primary spec.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD shards from `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/`.
- No UX artifact discovered; Dib V1 has no browser UI or frontend surface.
- No `project-context.md` file exists under the project root.
- Loaded previous story intelligence from `_bmad-output/implementation-artifacts/4-5-report-provenance-and-redact-sensitive-diagnostics.md`.
- Web research completed 2026-06-12 against public docs for Go downloads, Go `flag`, pflag, Cobra flags, and Viper. Treat these as inspiration-only behavior references unless copied/adapted content is introduced.

### Current Repository State

- Baseline commit at story creation: `a718d1d` (`docs: add epic 4 retrospective`).
- Root `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`; no `go.sum` exists in the discovered tree.
- `docs/compatibility.md` does not exist yet.
- `docs/provenance-log.md` already has Initial Entries for Go `flag`, pflag, Cobra flags documentation, and Viper package documentation, all classified as `inspiration-only` with 2026-06-11 access dates.
- Existing unrelated BMAD installer/config, `.agents/`, `.codex/`, `.idea/`, and story-automator changes exist in the worktree. Do not revert or normalize them.

### Current Implementation Evidence To Cite

- `docs/behavior-matrices.md` is the strongest local evidence source. It already marks Epic 2 flag parser behavior, Epic 3 command routing/rendering/boundary behavior, and Epic 4 config resolution/provenance behavior as `current`.
- Flag evidence includes long flags, shorthand flags, shorthand groups, repeated/custom values, no-option defaults, parse boundaries, help requests, typed parse diagnostics, redaction, and fuzz/property hardening. [Source: `docs/behavior-matrices.md#Flag-Parser-Evidence-Map`]
- Command evidence includes immutable definitions, nested routing, aliases, unknown diagnostics, local/inherited flag composition, parser boundary preservation, deterministic help/usage rendering, and caller-controlled boundary metadata. [Source: `docs/behavior-matrices.md#Flag-Parser-Evidence-Map`]
- Config evidence includes registered definitions, default snapshots, explicit/env/JSON source ingestion, flag binding, precedence resolution, typed getters, source reports, rendered diagnostics, redaction, and deterministic reports. [Source: `docs/behavior-matrices.md#Flag-Parser-Evidence-Map`]
- `docs/config-precedence.md` is the canonical precedence authority. The order is `explicit setter`, `flag binding`, `env`, `JSON`, `default`. [Source: `docs/config-precedence.md#Precedence-Order`]
- `docs/diagnostics-and-errors.md` is the diagnostic/source-label/redaction vocabulary authority. Config source labels are exactly `default`, `explicit setter`, `flag binding`, `env`, and `JSON`. [Source: `docs/diagnostics-and-errors.md#Source-Labels`]
- `docs/release-checklist.md` exists but is not completed release evidence. Story 5.1 must not imply release readiness is complete. [Source: `docs/release-checklist.md#Release-Checklist`]

### Compatibility Content Boundaries

- This story creates adopter-facing compatibility boundaries only. Story 5.2 owns executable migration examples; Story 5.3 owns final behavior matrix consolidation as adoption evidence; Story 5.4 owns release-readiness evidence.
- Do not add runtime package code unless a docs freshness test requires a new test-only helper. The expected implementation is documentation plus provenance update.
- Keep Dib's positioning precise: behavior-scoped familiarity, native Dib API, standard-library-only runtime packages, explicit instances, typed errors, deterministic output, clean-room evidence, and caller-owned process lifecycle.
- Avoid feature-parity framing. The goal is compatibility clarity, not maximizing parity with Go `flag`, pflag, Cobra, or Viper.

### Current Public Docs Research Snapshot

- Go downloads listed Go 1.26.4 as a stable version on 2026-06-12. `go.mod` already uses `go 1.26`; do not add a `toolchain` directive.
- Go `flag` public package docs are for standard-library package `flag` at Go 1.26.4 and list package/global helpers plus `FlagSet` APIs. Use this only for behavior boundary context.
- pflag pkg.go.dev showed module `github.com/spf13/pflag` version `v1.0.10` published 2025-09-02 and documents shorthand, no-option defaults, normalization, deprecation/hidden flags, and Go flag support concepts. Use this only for behavior boundary context.
- Cobra public docs describe command flags as local or persistent and include guidance for flag behavior; Story 5.1 should discuss Dib's command routing and inherited/local flags without importing Cobra's app framework shape.
- Viper pkg.go.dev showed module `github.com/spf13/viper` version `v1.21.0` published 2025-09-08 and documents defaults, config files, env, flags, remote key/value stores, watching, aliases, and getters. Dib supports only the V1 subset in the PRD and implemented behavior.

### Files Expected To Be Created Or Updated

- NEW: `docs/compatibility.md`
- UPDATE: `docs/provenance-log.md` if Story 5.1 uses current public reference docs while writing compatibility prose or needs to expand affected artifacts for existing entries.
- Optional UPDATE: `docs/behavior-matrices.md` only if the dev agent decides the Story 5.1 compatibility row should move from `later` to `current` in the same story. If updated, keep it narrowly scoped and cite `docs/compatibility.md`.
- Avoid changes under `command/`, `flags/`, `config/`, `examples/`, or `tools/` unless required by a narrowly scoped docs freshness test.

### Previous Story Intelligence

- Story 4.5 added `config/report.go`, `config/report_test.go`, `ExampleSnapshot_SourceReport`, and QA coverage for provenance reports and diagnostics.
- Story 4.5 updated `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, and `docs/config-precedence.md` after executable evidence existed.
- Story 4.5 completed with `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`; use the same verification rhythm.
- Story 4.5 explicitly deferred compatibility tables, migration examples, and release readiness to Epic 5. Do not blur that boundary.

### Git Intelligence

- Recent commits:
  - `a718d1d docs: add epic 4 retrospective`
  - `ca11b27 feat(story-4.5): report config provenance`
  - `b974995 feat(story-4.4): add typed config getters`
  - `c231957 feat(story-4.3): resolve config precedence`
  - `6655951 feat(story-4.2): ingest explicit config sources`
- Established implementation rhythm: read all update files first, make the smallest docs/API change needed, update evidence docs only after claims are settled, then run repository gates.

### Architecture Guardrails

- Runtime packages must import only the Go standard library. [Source: `_bmad-output/planning-artifacts/architecture.md#Technical-Constraints-Dependencies`]
- Dib must not claim source compatibility with Go `flag`, pflag, Cobra, or Viper. It offers a native Dib API with familiar concepts and documented differences. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#API-Contracts-Versioning-And-Dependency-Policy`]
- Compatibility with Go `flag`, pflag, Cobra, and Viper is semantic familiarity only; Dib must not promise source compatibility, drop-in replacement behavior, naming parity, legacy mental-model preservation, or reproduced edge-case bugs. [Source: `_bmad-output/planning-artifacts/architecture.md#Non-Functional-Requirements`]
- Do not add compatibility aliases or migration-oriented terminology from existing CLI/config libraries merely because users may recognize them. [Source: `_bmad-output/planning-artifacts/architecture.md#Naming-Patterns`]
- Record provenance for copied, generated, adapted, or inspiration-only reference-derived artifacts. [Source: `_bmad-output/planning-artifacts/architecture.md#Clean-Room-Provenance-Enforcement`]
- Docs/examples must reflect package boundaries, snapshot contracts, and public error behavior. [Source: `_bmad-output/planning-artifacts/architecture.md#API-Communication-Patterns`]

### External Reference Links For Provenance

- Go downloads: https://go.dev/dl/
- Go `flag` package docs: https://pkg.go.dev/flag
- pflag package docs: https://pkg.go.dev/github.com/spf13/pflag
- Cobra flags docs: https://cobra.dev/docs/how-to-guides/working-with-flags/
- Viper package docs: https://pkg.go.dev/github.com/spf13/viper

## Project Structure Notes

- `docs/compatibility.md` is the architecture-owned home for FR18. Create it under `docs/`; do not create a new docs subtree.
- `docs/provenance-log.md` owns source, access date, license/terms, affected artifact, and classification. Update it before acceptance if Story 5.1 uses external docs beyond already recorded provenance.
- `docs/behavior-matrices.md` remains the executable-evidence index; package tests remain the executable source of truth.
- No UX files exist and no UI work applies.

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- Red phase: `GOCACHE=/tmp/dib-go-build go test ./docs` failed before `docs/compatibility.md` existed.
- Green/refactor: `GOCACHE=/tmp/dib-go-build go test ./docs` passed after adding compatibility prose and freshness checks.
- Verification: `GOCACHE=/tmp/dib-go-build go test ./...`, `GOCACHE=/tmp/dib-go-build go vet ./...`, `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`, and `git diff --check` all passed.
- Manual wording gate: `rg -n "(?i)drop-in|source-compatible|clone API|framework compatibility layer" docs/compatibility.md docs/clean-room-policy.md` returned only explicit boundary or omission language.
- Dependency check: root `go.mod` still contains only module plus `go 1.26`; `go.sum` is absent.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Created adopter-facing compatibility boundaries for Go `flag`, pflag, Cobra, and Viper concepts without source-compatibility positioning or migration walkthroughs.
- Added a standard-library docs test to keep compatibility wording tied to evidence links, deferred-story boundaries, and prohibited positive compatibility claims.
- Updated provenance with Story 5.1 inspiration-only entries for public Go `flag`, pflag, Cobra, and Viper documentation used while writing the compatibility prose.
- Updated the behavior matrix compatibility row from deferred to current and kept migration, consolidated adoption evidence, and release readiness deferred to Stories 5.2, 5.3, and 5.4.
- Senior review auto-fixed ambiguous compatibility-positioning wording in linked evidence and expanded docs freshness coverage for that evidence.

### Change Log

- 2026-06-12: Added compatibility boundary documentation, Story 5.1 provenance entries, behavior matrix evidence update, and docs freshness test.
- 2026-06-12: Senior review auto-fixed ambiguous compatibility evidence wording, expanded clean-room positioning test coverage, and marked story done.

### File List

- `_bmad-output/implementation-artifacts/5-1-publish-compatibility-boundaries-for-familiar-cli-concepts.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/implementation-artifacts/tests/test-summary.md`
- `docs/behavior-matrices.md`
- `docs/compatibility.md`
- `docs/compatibility_test.go`
- `docs/provenance-log.md`

## Senior Developer Review (AI)

Reviewer: GPT-5 Codex on 2026-06-12

Outcome: Approved after automatic fixes. No critical issues remain.

Findings fixed:

- [x] MEDIUM: `docs/behavior-matrices.md` used ambiguous positive `source-compatible` wording in the command evidence row linked by Story 5.1 evidence. Reworded it as preservation of the stable Story 1.1 constructor call shape.
- [x] MEDIUM: `docs/compatibility_test.go` only guarded compatibility prose against prohibited positive positioning, not linked evidence docs. Added `TestCompatibilityEvidenceDocsAvoidAmbiguousPositioning`.
- [x] MEDIUM: `_bmad-output/implementation-artifacts/tests/test-summary.md` had Story 5.1 changes but was absent from the story File List. Added it to the File List.

Validation:

- [x] `GOCACHE=/tmp/dib-go-build go test ./docs`
- [x] `GOCACHE=/tmp/dib-go-build go test ./...`
- [x] `GOCACHE=/tmp/dib-go-build go vet ./...`
- [x] `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
- [x] `git diff --check`
- [x] Official web fallback checked for Go `flag`, pflag, Cobra flags, and Viper documentation referenced by provenance.
