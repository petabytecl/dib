---
baseline_commit: 02290031d21b6e3938ab8f2011ece39f79233cbc
created: "2026-06-11T16:21:10-04:00"
---

# Story 2.1: Define Reusable Flag Sets Without Global State

Status: ready-for-dev

## Story

As a Go CLI developer,
I want reusable Flag sets with explicit definitions and value metadata,
so that I can parse CLI input without package-global mutable state or hidden process dependencies.

## Requirements Trace

- FR5: Developers can define independent Flag sets without package-level mutable state.
- FR8: Developers can define repeated flags, custom parsers, and built-in value metadata.
- FR9: The initial model must prepare parse-boundary and diagnostic state for later parser stories.
- FR20: Definition behavior must be covered by table-driven tests and documented behavior matrices.
- NFR1, NFR2, NFR3, NFR6, NFR8: runtime code remains standard-library-only, explicit-instance based, inspectable through typed errors, table-testable, and redaction-safe.

## Acceptance Criteria

1. Given a caller defines a Flag set, when flags are registered, then each definition captures name, optional shorthand, default value, usage text, value parser, repeat policy, hidden/deprecated metadata, sensitivity metadata, and no-option default where applicable, and built-in value kinds include string, bool, int, int64, uint, uint64, float64, duration, and string list.
2. Given two independent Flag sets use the same flag names, when each Flag set is parsed or inspected, then their definitions and parse snapshots remain independent, and no package-level global registry, default Flag set, or ambient `os.Args` dependency is introduced.
3. Given later parser behavior depends on typed values and diagnostics, when the initial value and diagnostic model is implemented, then value arity, default handling, explicit-set tracking, duplicate detection, and diagnostic categories are represented in machine-readable state, and public errors are inspectable through `errors.Is` or `errors.As` where callers need programmatic handling.
4. Given definitions are reusable values, when callers derive or extend a Flag set, then the original Flag set keeps unchanged observable behavior, and tests prove no caller-observable mutation or slice/map alias leak across repeated parses.
5. Given this story establishes the parser foundation, when verification runs, then table-driven tests cover definition validation, duplicate long names, duplicate shorthands, explicit-set tracking, default values, and reusable definitions, and `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` pass.

## Tasks / Subtasks

- [ ] Confirm current tracker and repository state (AC: 1-5)
  - [ ] Verify this story is the first Epic 2 implementation story and Story 1.5 is `done` in `sprint-status.yaml`.
  - [ ] Verify `flags/` currently contains only package documentation before adding the first runtime implementation files.
  - [ ] Verify root `go.mod` still declares `module github.com/petabytecl/dib` and `go 1.26` with no root dependency directives.
  - [ ] Check whether ATDD artifacts for Story 2.1 exist; if they do, use them as executable starting points.

- [ ] Add reusable flag definition values (AC: 1, 2, 4)
  - [ ] Add focused `flags/` source files for flag definitions and flag-set definitions, following the architecture's expected package layout without creating a root facade.
  - [ ] Represent long name, optional one-character shorthand, usage text, default value, value kind/parser, repeat policy, no-option default, hidden/deprecated metadata, and sensitivity metadata.
  - [ ] Make definition structs caller-observably immutable: do not expose mutable maps, slices, parser internals, or fields that allow in-place changes to an existing definition.
  - [ ] Provide derivation/extension APIs that return a new Flag set and leave the original value unchanged.
  - [ ] Provide deterministic inspection APIs for definitions so tests and downstream packages do not need to scrape string output.

- [ ] Add value and diagnostic foundation state (AC: 1, 3)
  - [ ] Add built-in value kinds for string, bool, int, int64, uint, uint64, float64, duration, and string list.
  - [ ] Model value arity explicitly enough for later parser stories to distinguish required values, boolean presence, no-option defaults, and repeated values.
  - [ ] Add minimal custom parser interfaces or function types that preserve caller errors through wrapping and `errors.As`.
  - [ ] Add machine-readable diagnostic/error categories for invalid definitions, duplicate long names, duplicate shorthands, invalid shorthands, invalid no-option defaults, duplicate values, and conversion failures needed by this foundation.
  - [ ] Keep sensitive raw values out of `Error`, `String`, debug, and diagnostic text paths.

- [ ] Implement setup-time validation (AC: 1, 3, 5)
  - [ ] Reject empty or invalid long names before definitions can be used.
  - [ ] Reject duplicate long names within the same Flag set.
  - [ ] Reject duplicate shorthands within the same Flag set.
  - [ ] Reject invalid shorthand definitions that are empty, multi-rune, or otherwise unusable.
  - [ ] Reject no-option defaults where the value kind or arity cannot consume them.
  - [ ] Return typed or sentinel errors that callers can inspect with `errors.Is` or `errors.As`; error strings are diagnostics only.

- [ ] Add tests and behavior evidence (AC: 2, 4, 5)
  - [ ] Add table-driven tests beside the package under test, likely `flags/set_test.go`, `flags/value_test.go`, and `flags/errors_test.go` if that split keeps files focused.
  - [ ] Cover definition validation, duplicate long names, duplicate shorthands, default values, explicit-set tracking foundation state, independent Flag sets with identical names, and derivation without mutation or slice/map alias leaks.
  - [ ] Add tests proving package APIs do not read `os.Args`, mutate package globals, or share parse/definition state between Flag sets.
  - [ ] Update `docs/behavior-matrices.md` only if the new public behavior needs an adoption-facing matrix row; do not duplicate package-level test detail.
  - [ ] Avoid `flags/testdata/fuzz/FuzzParse/` unless ATDD scaffolding for this story explicitly asks for a parser seed. Parser fuzzing is primarily a later Epic 2 hardening story.

- [ ] Verify the story output (AC: 1-5)
  - [ ] Run `go test ./...`.
  - [ ] Run `go vet ./...`.
  - [ ] Run `go run ./tools/depgate`.
  - [ ] Run `git diff --check`.
  - [ ] Confirm root `go.mod` still has no `require`, `replace`, or `toolchain` directives.
  - [ ] Confirm no package imports outside the Go standard library were added.
  - [ ] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### ATDD Artifacts

- Checklist: `_bmad-output/test-artifacts/atdd-checklist-2-1-define-reusable-flag-sets-without-global-state.md`
- Backend package acceptance scaffolds:
  - `flags/atdd_contract_test.go`
  - `flags/set_atdd_test.go`
  - `flags/state_atdd_test.go`
- Temp API/back-end generation summary: `_bmad-output/test-artifacts/tmp/tea-atdd-api-tests-2026-06-11T16-30-15-0400.json`
- Temp E2E generation summary: `_bmad-output/test-artifacts/tmp/tea-atdd-e2e-tests-2026-06-11T16-30-15-0400.json`
- Temp aggregate summary: `_bmad-output/test-artifacts/tmp/tea-atdd-summary-2026-06-11T16-30-15-0400.json`
- Dev workflow handoff: remove one `t.Skip` in `flags/set_atdd_test.go` or `flags/state_atdd_test.go` at a time, confirm RED with the narrow `go test ./flags -run ...` command, then implement the smallest change to pass.

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`.
- Loaded previous story: `_bmad-output/implementation-artifacts/1-5-run-trust-gates-in-ci.md`.
- Loaded current source/docs: `flags/doc.go`, `go.mod`, `docs/behavior-matrices.md`, and `docs/diagnostics-and-errors.md`.
- No UX document, `project-context.md`, `CLAUDE.md`, or local `MEMORY.md` was discovered in the repo.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `02290031d21b6e3938ab8f2011ece39f79233cbc` (`chore: mark story 1.5 done`).
- `main` is ahead of `origin/main` by eight commits at story creation; do not assume Epic 1 work has been pushed.
- `sprint-status.yaml` marks Epic 1 and Story 1.1 through Story 1.5 as `done`; Epic 2 is moved to `in-progress` by this story creation.
- Story 2.1 is the first implementation story in Epic 2. `flags/doc.go` is currently package documentation only.
- Root `go.mod` currently contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

### Architecture Guardrails

- `flags/` is the public flag parsing package and owns explicit flag sets, long/shorthand parsing, repeated/custom values, no-option defaults, normalization, parse diagnostics, usage metadata, and flag-specific typed errors. [Source: `_bmad-output/planning-artifacts/architecture.md:562`]
- `flags/` must remain fully usable without `command/` or `config/`; shared flag metadata and parsing semantics live in `flags/`, not `command/`. [Source: `_bmad-output/planning-artifacts/architecture.md:567`, `_bmad-output/planning-artifacts/architecture.md:570`]
- The module root does not provide a broad public facade, package-global helpers, or default singleton API. [Source: `_bmad-output/planning-artifacts/architecture.md:564`]
- Definitions and Flag sets are reusable values; derived definitions return new values and must avoid exported mutable internals or shallow-copy aliasing. [Source: `_bmad-output/planning-artifacts/architecture.md:205`]
- Flag parsing returns self-contained snapshots that must not write back to definitions or depend on live process state after creation. This story may add the foundation shape but should avoid implementing broad parser behavior from later stories unless tests require a minimal snapshot. [Source: `_bmad-output/planning-artifacts/architecture.md:207`]
- Setup-time validation must catch invalid definitions and duplicate names where possible. [Source: `_bmad-output/planning-artifacts/architecture.md:217`]
- Public APIs use explicit inputs and returned values/errors and must not depend on package globals, hidden process IO, implicit environment reads, or root singletons. [Source: `_bmad-output/planning-artifacts/architecture.md:229`]
- Source remains organized by capability package; shared code belongs under `internal/` only after multiple concrete call sites prove the need. [Source: `_bmad-output/planning-artifacts/architecture.md:636`, `_bmad-output/planning-artifacts/architecture.md:637`]
- Package tests are the executable contract and must live beside package code. [Source: `_bmad-output/planning-artifacts/architecture.md:642`, `_bmad-output/planning-artifacts/architecture.md:654`]

### Requirements Notes

- FR5 requires independent Flag sets without package-level mutable state; two Flag sets with the same names must parse independently, and explicit instances are the only primary API surface. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:127`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:134`]
- FR8 requires repeated values, custom value parsers, typed/wrapped parser errors, and built-ins for string, bool, int, int64, uint, uint64, float64, duration, and string list. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:156`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:164`]
- FR9 requires typed diagnostics for unknown flags, missing values, invalid values, duplicate shorthands, invalid groups, and help requests, and diagnostics must not leak sensitive values. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:166`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:174`]
- FR20 requires table-driven parser, command, and config behavior tests. For this story, keep coverage focused on the definition and foundation state that exists now. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:283`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:286`]
- NFRs require standard-library-only runtime packages, explicit-instance APIs, typed errors, deterministic diagnostics, no default process control, table-driven tests, compatibility clarity, and redaction-safe sensitive diagnostics. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:310`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:317`]

### Scope Boundaries

Implement in this story:

```text
flags/doc.go
flags/flag.go
flags/set.go
flags/value.go
flags/errors.go
flags/set_test.go
flags/value_test.go
flags/errors_test.go
docs/behavior-matrices.md
```

The exact file split may vary, but keep files focused and small. Do not create these unless a failing test proves they are required now:

```text
flags/parse.go
flags/parse_long_test.go
flags/parse_shorthand_test.go
flags/normalize.go
flags/fuzz_test.go
flags/testdata/fuzz/FuzzParse/
cmd/
internal/
go.sum
```

Out of scope for this story:

- Full long-flag parse forms from Story 2.3.
- Shorthand parsing and shorthand groups from Stories 2.4 and 2.5.
- Name normalization from Story 2.2.
- Repeated/custom CLI accumulation behavior beyond definition metadata and foundation validation from Story 2.6.
- Parse terminator and remaining-args behavior from Story 2.7.
- Parser fuzzing and full behavior-matrix proof from Story 2.8.

### Technical Research Notes

- Go's standard `flag` package provides useful prior art for independent `FlagSet` values, built-in primitive/duration registrations, custom `Value` implementations, and boolean presence semantics. Dib should not mirror the top-level `flag.*` global/default APIs because the project explicitly requires primary explicit instances and no package-global helpers. Source: https://pkg.go.dev/flag
- `strconv` provides the standard-library primitives needed for bool, int, int64, uint, uint64, and float64 conversion and exposes `NumError`, `ErrRange`, and `ErrSyntax` for preserving wrapped conversion context where useful. Source: https://pkg.go.dev/strconv
- `time.ParseDuration` is the standard-library source for duration values and supports the V1 duration built-in. Source: https://pkg.go.dev/time#ParseDuration

### Previous Story Intelligence

- Story 1.5 added `.github/workflows/ci.yml` with `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` as CI trust gates.
- Story 1.4 implemented `tools/depgate`; it is now the dependency authority and should remain isolated tooling.
- Story 1.3 established `docs/behavior-matrices.md` and `docs/diagnostics-and-errors.md`. Use those docs as adoption-facing evidence, but avoid turning them into the executable source of truth.
- Story 1.1 created `flags/doc.go` with explicit-instance/no-global package guidance; this story should build on that package, not replace it.
- Epic 1 work has repeatedly preserved the zero-dependency root module: no `require`, `replace`, `toolchain`, or `go.sum` should appear unless architecture changes.

### Testing Standards

- Follow red-green-refactor if ATDD artifacts are present. Activate one skipped Story 2.1 test at a time, confirm it fails, then implement the smallest production change that passes.
- Add table-driven unit tests before or alongside each behavior. Keep tests standard-library-only.
- Required final verification for implementation: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.
- Use `go test -race ./...` as optional extra evidence if the implementation changes concurrency-sensitive state; the release checklist treats race testing as a release-candidate gate.
- Prefer testing observable immutability by deriving a Flag set, mutating caller-owned slices or values after construction, and proving previous definitions/snapshots remain unchanged.

### Security And Quality Checks

- Do not hardcode secrets, credentials, tokens, or private URLs.
- Treat custom parser inputs and default/no-option values as untrusted boundary data.
- Sensitive raw values must not appear in errors, `String` output, debug text, rendered diagnostics, source reports, examples, or validation failures.
- Do not import non-standard-library modules in runtime, tests, examples, or tools.
- Do not add `os.Args`, `flag.CommandLine`, package-global mutable registries, hidden stdout/stderr use, or `os.Exit`.
- Keep public errors inspectable with `errors.Is` / `errors.As`; do not make string matching the contract.

## Dev Agent Record

### Agent Model Used

TBD

### Debug Log References

TBD

### Completion Notes List

TBD

### File List

TBD

### Change Log

- 2026-06-11: Story created and marked ready for development.
