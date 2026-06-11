---
baseline_commit: 9ea64eea900cfbb3533df76250d87138dc27d052
created: "2026-06-11T12:53:37-04:00"
---

# Story 1.2: Publish the Clean-Room Contribution Contract

Status: done

## Story

As a technical reviewer,
I want Dib's clean-room policy and provenance expectations documented before feature work expands,
so that I can verify the project is not copying source, tests, examples, fixtures, names, or structure from inspiration projects.

## Requirements Trace

- FR17: Publish the clean-room source policy with allowed and disallowed inputs.
- FR21: Preserve the zero-runtime-dependency trust claim while adding documentation.
- NFR1, NFR7, NFR9: standard-library runtime ceiling, compatibility clarity, and Go 1.26+ baseline remain in force.

## Acceptance Criteria

1. Given Dib is inspired by public behavior from Go `flag`, pflag, Cobra, and Viper, when the clean-room policy is written, then `docs/clean-room-policy.md` defines allowed inputs as public documentation and observable behavior, and defines disallowed inputs as copied source, tests, comments, examples, internal names, fixtures, file organization, and source-derived generated content.
2. Given contributors need a short operational contract, when contribution guidance is written, then `CONTRIBUTING.md` links to the clean-room policy, and states that compatibility examples, tests, docs, and fixtures must be independently written or explicitly recorded with provenance.
3. Given future stories may use public documentation or other references, when `docs/provenance-log.md` is created, then it provides a repeatable entry format with source, access date, license or terms, affected artifact, and classification as copied, adapted, or inspiration-only, and includes initial inspiration-only entries for the PRD/architecture-approved public documentation sources where appropriate.
4. Given clean-room compliance is part of the adoption promise, when verification runs, then `go test ./...` still passes, and the documentation states that provenance gaps block acceptance for copied, generated, adapted, or reference-derived artifacts.

## Tasks / Subtasks

- [x] Confirm live repository state before editing (AC: 1-4)
  - [x] Preserve the uncommitted Story 1.1 baseline files: `go.mod`, `command/`, `flags/`, `config/`, Story 1.1, and `sprint-status.yaml`.
  - [x] Verify `docs/` exists but has no files, and `CONTRIBUTING.md` does not exist before implementation.
  - [x] Do not reset, remove, or rewrite Story 1.1 implementation work; this story is additive documentation.

- [x] Create `docs/clean-room-policy.md` (AC: 1, 4)
  - [x] Define allowed inputs: public documentation, package docs, observable behavior, independently written user stories, independently written tests, and project-authored planning artifacts.
  - [x] Define disallowed inputs: copied source, tests, comments, examples, fixtures, internal names, file organization, non-public implementation details, copied README shapes, and source-derived generated content from Go `flag`, pflag, Cobra, Viper, or comparable projects.
  - [x] State that AI-generated content is unacceptable when it is too closely derived from source projects, examples, fixtures, names, or structure.
  - [x] State that provenance gaps block acceptance for copied, generated, adapted, or reference-derived artifacts until the gap is resolved.
  - [x] Keep the policy behavior-scoped: Dib may study public behavior, but must not claim source compatibility or a drop-in API.

- [x] Create `CONTRIBUTING.md` with the short operational contract (AC: 2, 4)
  - [x] Link to `docs/clean-room-policy.md` and `docs/provenance-log.md`.
  - [x] Require contributors to independently write compatibility examples, tests, docs, fixtures, behavior matrices, and fuzz seeds unless provenance explicitly records copied or adapted status.
  - [x] State that compatibility claims must be behavior-scoped, test-backed where practical, and must not describe Dib as source-compatible with Go `flag`, pflag, Cobra, or Viper.
  - [x] Include local verification for this story: `go test ./...`; also mention `go vet ./...` as a normal baseline check, but do not introduce `tools/depgate/` before Story 1.4.
  - [x] State that runtime, test, and documentation examples must not add external dependencies unless the architecture is updated.

- [x] Create `docs/provenance-log.md` (AC: 3, 4)
  - [x] Provide a repeatable entry template with source, URL, access date, license or terms, affected artifact, classification, and notes.
  - [x] Define classifications exactly enough for future use: `copied`, `adapted`, and `inspiration-only`.
  - [x] Add initial `inspiration-only` entries for these architecture-approved public documentation sources:
    - [x] Go `flag` package documentation: `https://pkg.go.dev/flag`, terms/license source `https://go.dev/LICENSE`.
    - [x] pflag package documentation: `https://pkg.go.dev/github.com/spf13/pflag`, license source `https://github.com/spf13/pflag/blob/master/LICENSE`.
    - [x] Cobra flags documentation: `https://cobra.dev/docs/how-to-guides/working-with-flags/`, license source `https://github.com/spf13/cobra/blob/main/LICENSE.txt`.
    - [x] Viper package documentation: `https://pkg.go.dev/github.com/spf13/viper`, license source `https://github.com/spf13/viper/blob/master/LICENSE`.
  - [x] For each initial entry, set affected artifact to the planning or docs artifact actually influenced, not runtime source code.
  - [x] Do not copy examples, code snippets, tables, fixtures, file layout, or internal names from the referenced projects into the log or docs.

- [x] Verify the story output (AC: 4)
  - [x] Run `go test ./...` after adding docs.
  - [x] Run `go vet ./...` if the working tree still has Go packages available.
  - [x] Confirm no runtime package, test, or example import was added.
  - [x] Record verification commands and outcomes in the Dev Agent Record.

### Review Findings

- [x] [Review][Patch] CONTRIBUTING uses path mentions instead of Markdown links required by AC2 [`CONTRIBUTING.md:3`]
- [x] [Review][Patch] Copied/adapted provenance language can be read as permitting disallowed copied tests, examples, fixtures, names, or structure [`CONTRIBUTING.md:11`]
- [x] [Review][Patch] Generated/reference-derived material requires provenance but has no explicit mapping to the allowed provenance classifications [`docs/provenance-log.md:3`]
- [x] [Review][Patch] Copied/adapted entries require reviewer approval but the provenance template has no approval field [`docs/provenance-log.md:10`]
- [x] [Review][Patch] Public documentation and project-authored planning artifacts are allowed too broadly without clarifying clean provenance and no copied examples, wording, names, or structure [`docs/clean-room-policy.md:11`]
- [x] [Review][Patch] Initial provenance notes only deny copying into runtime source, not the docs and story artifacts changed by this story [`docs/provenance-log.md:40`]
- [x] [Review][Patch] CONTRIBUTING has a Story 1.4-specific dependency-gate note that will become stale in an evergreen contributor guide [`CONTRIBUTING.md:33`]

## Dev Notes

### Source Discovery

- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded PRD workspace: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`, `addendum.md`, `reconcile-brief.md`, and `review-rubric.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded readiness report: `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-11.md`.
- Loaded previous story: `_bmad-output/implementation-artifacts/1-1-adopt-an-auditable-go-module-baseline.md`.
- Loaded product brief for source grounding: `_bmad-output/planning-artifacts/briefs/brief-dib-2026-06-10/brief.md`.
- No UX document or `project-context.md` was discovered.

### Current Repository State

- Branch at story creation: `main`.
- `sprint-status.yaml` marks Story 1.1 done and Story 1.2 as the first backlog story.
- Story 1.1 changes are still uncommitted in the working tree. Treat them as owned baseline work, not disposable local noise.
- Current implementation files from Story 1.1:
  - `go.mod` declares module `github.com/petabytecl/dib` with `go 1.26`.
  - `command/doc.go`, `flags/doc.go`, and `config/doc.go` establish package-level direction for explicit instances, caller-owned inputs, immutable definitions, per-run snapshots, typed errors, and no ambient process state.
  - `command/definition.go` contains the minimal validated command definition proof; `Definition` is valid only when created by `NewDefinition`.
  - `command/definition_test.go` uses external-caller tests and a runnable example.
- `docs/` exists but contains no files at story creation. `CONTRIBUTING.md` is absent.

### Architecture Guardrails

- This is documentation/provenance work, not a runtime behavior story.
- Do not add a root facade package, `/cmd` scaffold, binary entry point, CI workflow, dependency gate tool, compatibility matrix, migration examples, behavior matrix, release checklist, parser fuzz seeds, or config precedence docs in this story unless strictly needed to satisfy Story 1.2 acceptance criteria.
- Keep runtime code, tests, examples, and future fixture guidance standard-library-only unless the architecture is updated.
- Public docs and observable behavior are allowed references. Copied source, tests, comments, examples, fixtures, names, file organization, implementation details, copied README shapes, and source-derived generated content are disallowed.
- Any copied, generated, adapted, or reference-derived artifact needs provenance before acceptance. Inspiration-only references must not contribute copied names, examples, fixtures, or structure.
- Compatibility language must say familiar behavior, behavior-scoped, or native Dib API. Do not say source-compatible, drop-in replacement, clone API, or framework compatibility layer.

### File Structure Requirements

Expected new files for this story:

```text
docs/clean-room-policy.md
docs/provenance-log.md
CONTRIBUTING.md
```

Do not update the runtime packages for this story unless a verification issue makes a tiny corrective edit necessary. Do not create `tools/depgate/`; that belongs to Story 1.4.

### Clean-Room Policy Content Requirements

The policy should be concise and operational. Include:

- Purpose and scope: Dib is a clean-room Go library/toolkit, not a source-compatible clone.
- Allowed inputs list.
- Disallowed inputs list.
- AI/generated-content rule for source-derived material.
- Contributor workflow: record provenance before acceptance when copying, adapting, generating, or using references beyond inspiration.
- Acceptance blocker rule: missing provenance blocks copied, generated, adapted, or reference-derived artifacts.
- Relationship to `docs/provenance-log.md` and `CONTRIBUTING.md`.

### Provenance Log Content Requirements

Use a template that future stories can copy without inventing new fields:

```text
## Entry Template

- Source:
- URL:
- Access date:
- License or terms:
- Affected artifact:
- Classification: copied | adapted | inspiration-only
- Notes:
```

Initial entries should be `inspiration-only` and should identify that they informed PRD/architecture/story planning and clean-room policy boundaries. They must not imply runtime source code, tests, examples, or fixtures were copied.

### Previous Story Intelligence

- Story 1.1 established the Go module, public package boundaries, minimal command behavior proof, and local verification.
- Story 1.1 review found and resolved a misleading `errors.Is` comment and clarified that `Definition` must be constructed through `NewDefinition` for validation.
- Reuse Story 1.1's discipline: small focused files, explicit source notes, standard-library-only validation, no ambient process state, and no overbroad API surface.
- Do not treat the temporary `go list` dependency check from Story 1.1 as accepted release-candidate evidence. The architecture limits that temporary command to the initial scaffold path and requires `tools/depgate/` before release-candidate evidence.

### Latest Technical Context

- Go `flag` public docs are acceptable inspiration for behavior concepts such as `FlagSet`, typed flag helpers, custom values, parsing, usage rendering, and `--` termination, but do not copy examples or source-derived prose.
- pflag public docs are acceptable inspiration for POSIX/GNU-style long flags, shorthand flags, shorthand grouping, no-option defaults, and interspersed args, but do not copy names, examples, tests, fixtures, or file organization.
- Cobra public docs are acceptable inspiration for command trees, local/persistent flags, nested subcommands, aliases, help, and optional Viper integration, but Dib must not copy command structures or generated scaffold shapes.
- Viper public docs are acceptable inspiration for defaults, explicit setters, flags, environment variables, config files, external stores, and precedence, but Dib intentionally excludes broad Viper scope such as package-global singleton defaults, non-JSON formats, remote stores, live reload, and source-compatible APIs.
- Current license/terms sources checked during story creation: Go license (`https://go.dev/LICENSE`), pflag license (`https://github.com/spf13/pflag/blob/master/LICENSE`), Cobra Apache-2.0 license (`https://github.com/spf13/cobra/blob/main/LICENSE.txt`), and Viper MIT license (`https://github.com/spf13/viper/blob/master/LICENSE`).

### Testing Standards

- `go test ./...` is required by the story.
- `go vet ./...` is a useful baseline check and should stay clean.
- Documentation-only changes should not add imports, generated code, fixtures, or examples that change the dependency graph.
- If docs include shell commands, keep them copy-pasteable and aligned with the current module state.

### Security And Quality Checks

- No secrets, tokens, generated credentials, or real private paths belong in docs or examples.
- Do not invent fake secrets in this story; the architecture-owned redaction corpus is for later config/redaction tests.
- Do not paste copied third-party examples into docs.
- Keep docs short enough for reviewers to audit. Prefer direct operational language over broad policy prose.

### References

- `_bmad-output/planning-artifacts/epics.md:234` - Story 1.2 source story.
- `_bmad-output/planning-artifacts/epics.md:244` - Clean-room policy acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:249` - CONTRIBUTING acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:254` - Provenance log acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:259` - Verification and provenance-gap acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:108` - Required clean-room and provenance docs list.
- `_bmad-output/planning-artifacts/epics.md:109` - Provenance fields for copied/generated/adapted/inspiration-only artifacts.
- `_bmad-output/planning-artifacts/architecture.md:60` - Clean-room policy as process and architecture constraint.
- `_bmad-output/planning-artifacts/architecture.md:433` - Provenance enforcement rule and inspiration-only boundary.
- `_bmad-output/planning-artifacts/architecture.md:650` - Documentation ownership for clean-room policy, CONTRIBUTING, and provenance log.
- `_bmad-output/planning-artifacts/architecture.md:805` - First implementation priority includes clean-room policy/provenance docs.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:250` - FR17 clean-room policy.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:290` - FR21 runtime dependency rule.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:387` - Clean-room drift mitigation.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/addendum.md:16` - Clean-room guardrails.
- `_bmad-output/implementation-artifacts/1-1-adopt-an-auditable-go-module-baseline.md:156` - Story 1.1 Dev Agent Record and verification context.
- Go `flag` package documentation: https://pkg.go.dev/flag
- pflag package documentation: https://pkg.go.dev/github.com/spf13/pflag
- Cobra flags documentation: https://cobra.dev/docs/how-to-guides/working-with-flags/
- Viper package documentation: https://pkg.go.dev/github.com/spf13/viper
- Go license: https://go.dev/LICENSE
- pflag license: https://github.com/spf13/pflag/blob/master/LICENSE
- Cobra license: https://github.com/spf13/cobra/blob/main/LICENSE.txt
- Viper license: https://github.com/spf13/viper/blob/master/LICENSE

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `gh search repos "clean room policy Go library" --limit 5` - no reusable implementation selected.
- `gh search code "clean-room policy" --language Markdown --limit 5` - research only; no code or prose copied.
- `go test ./...` - passed.
- `go vet ./...` - passed.
- `git diff --check HEAD` - passed.
- `git status --short -- CONTRIBUTING.md docs _bmad-output/implementation-artifacts/1-2-publish-the-clean-room-contribution-contract.md _bmad-output/implementation-artifacts/sprint-status.yaml` - Story 1.2 changed only docs plus BMAD tracking files.
- Code review patches applied: tightened CONTRIBUTING links and provenance language, clarified clean-room allowed inputs, added generated/reference-derived classification mapping, added provenance approval fields, and updated inspiration-only notes to cover docs and story artifacts.

### Completion Notes List

- Added `docs/clean-room-policy.md` with allowed inputs, disallowed inputs, AI/source-derived content rules, provenance blocker language, and behavior-scoped compatibility language.
- Added `CONTRIBUTING.md` with links to clean-room and provenance docs, independent-writing expectations, dependency guardrails, and local verification commands.
- Added `docs/provenance-log.md` with reusable provenance fields, classification definitions, and initial inspiration-only entries for Go `flag`, pflag, Cobra, and Viper public documentation sources.
- This story did not add or modify runtime packages, tests, examples, imports, generated code, fixtures, or external dependencies.
- Unit tests were not added because this is a documentation-only story; the existing regression suite and vet checks passed.
- Resolved all seven code-review patch findings and moved the story to done.

### File List

- `CONTRIBUTING.md`
- `_bmad-output/implementation-artifacts/1-2-publish-the-clean-room-contribution-contract.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `docs/clean-room-policy.md`
- `docs/provenance-log.md`

### Change Log

- 2026-06-11T13:01:21-04:00 - Implemented Story 1.2 clean-room contribution contract documentation and marked the story ready for review.
- 2026-06-11T13:15:31-04:00 - Addressed all code-review patch findings and marked Story 1.2 done.
