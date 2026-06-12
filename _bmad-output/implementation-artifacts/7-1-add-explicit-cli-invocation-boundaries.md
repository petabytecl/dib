---
baseline_commit: 8094bd1
created: "2026-06-12"
---

# Story 7.1: Add Explicit CLI Invocation Boundaries

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Go CLI developer,
I want a `cli.Invocation` value that carries the program name and user arguments from caller-supplied argv,
so that I do not repeat `os.Args[1:]` slicing or lose testability at the process boundary.

## Requirements Trace

- FR26: Compose CLI invocation, command routing, flag parsing, and config resolution through an optional `cli` package.
- FR20: Provide behavior test matrices and package tests for public behavior.
- NFR2: Primary APIs operate on explicit instances and caller-supplied inputs/outputs; optional `cli` helpers are allowed only with caller-supplied process inputs and no hidden singleton state.
- NFR5: Library APIs must not call `os.Exit`, mutate process-wide streams, or read `os.Args` unless the caller chooses a documented convenience path that still passes inputs explicitly.
- NFR6: Core behavior must be testable with table-driven unit tests and injected args/readers/writers/env lookup.
- Architecture: `cli/` is an optional public composition package that owns explicit invocation values and the golden-path handoff between `command`, `flags`, and `config`. [Source: `_bmad-output/planning-artifacts/architecture.md`]
- Correct Course: #45 approved Epic 7 and `cli` as the optional composition package name. [Source: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-06-12-cli-composition.md`]
- GitHub tracker: Story 7.1 is #47.

## Acceptance Criteria

1. Given a caller passes full argv, when `cli.FromOSArgs(argv)` is called, then the result exposes `Program()` as argv[0] and `Args()` as argv[1:], and the package never reads `os.Args` itself.

2. Given a caller already has stripped args, when `cli.FromArgs(program, args)` is called, then the result exposes the caller-supplied program and args, and all slices are defensively copied.

3. Given invalid full argv is supplied, when the invocation cannot be constructed, then a typed error is returned, and no partial mutable state is exposed.

4. Given invocation values are reusable, when callers mutate the original argv or returned args slice, then the invocation's observable state does not change.

## Tasks / Subtasks

- [x] Confirm preconditions and package boundary before implementation (AC: 1-4)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 7 `in-progress` and Story 7.1 `ready-for-dev`.
  - [x] Confirm `cli/` does not already exist before creating it.
  - [x] Read existing public value-object patterns before editing: `command/boundary.go`, `command/result.go`, `command/errors.go`, `flags/snapshot.go`, `flags/errors.go`, `config/flag.go`, `config/snapshot.go`, and `config/errors.go`.
  - [x] Do not edit `command/`, `flags/`, `config/`, `README.md`, release docs, examples, coverage thresholds, lint tooling, or dependency-gate behavior in this story.

- [x] Create the `cli` package invocation API (AC: 1-4)
  - [x] Add `cli/doc.go` documenting `cli` as optional composition support, not a root facade or process-owning framework.
  - [x] Add `cli/invocation.go` with an immutable `Invocation` value that stores unexported `program string` and `args []string` fields.
  - [x] Implement `FromOSArgs(argv []string) (Invocation, error)`:
    - returns a typed error for empty full argv (`len(argv) == 0`);
    - treats `argv[0]` as the program name and `argv[1:]` as user arguments;
    - defensively copies user args;
    - does not import `os` or read `os.Args`.
  - [x] Implement `FromArgs(program string, args []string) Invocation`:
    - preserves caller-supplied program exactly, including an empty string if the caller chooses that;
    - defensively copies args;
    - does not validate stripped args because this path is already past the full-process argv boundary.
  - [x] Implement `Program() string` and `Args() []string` accessors; `Args()` must always return a defensive copy.

- [x] Add typed invocation errors (AC: 3)
  - [x] Add `cli/errors.go` with a sentinel such as `ErrInvalidInvocation = errors.New("invalid cli invocation")`.
  - [x] Add an inspectable typed error such as `InvocationError` with `Unwrap() error`, `Category() error`, and safe context accessors.
  - [x] Do not include raw argv contents in the error string. CLI args can contain secrets; diagnostics may mention shape such as "missing program" but not echo input values.
  - [x] Ensure callers can use `errors.Is(err, cli.ErrInvalidInvocation)` and `errors.As(err, *cli.InvocationError)`.

- [x] Add invocation tests (AC: 1-4)
  - [x] Add `cli/invocation_test.go` using `package cli_test`.
  - [x] Cover `FromOSArgs([]string{"dib", "deploy", "--verbose"})`: `Program()` returns `dib`, `Args()` returns `[]string{"deploy", "--verbose"}`.
  - [x] Cover `FromArgs("dib", []string{"deploy"})` and an explicit empty program case to prove caller ownership.
  - [x] Prove constructor defensive copies: mutate the original argv/args after construction and verify `Invocation` does not change.
  - [x] Prove accessor defensive copies: mutate the slice returned from `Args()` and verify a later `Args()` call returns the original values.
  - [x] Prove empty full argv returns a typed error and zero-value `Invocation`; assert both `errors.Is` and `errors.As`.
  - [x] Prove the package does not read process globals by setting `os.Args` in the test to a different value, passing an explicit argv slice to `FromOSArgs`, and asserting the explicit slice wins. Keep `os` imports test-only.

- [x] Preserve architecture and dependency guardrails (AC: 1-4)
  - [x] Keep `cli/` standard-library-only.
  - [x] Do not add root package exports, package-global default invocations, callbacks, `os.Exit`, stream mutation, env reads, JSON/file loading, or command/flag/config composition in this story.
  - [x] Do not make `command/`, `flags/`, or `config/` import `cli/`.
  - [x] Keep story 7.2 and 7.3 scope out of this story: no flag-to-config binding translation and no `cli.Resolve` yet.

- [x] Verify (AC: 1-4)
  - [x] `go test ./cli ./...`
  - [x] `go vet ./...`
  - [x] `go run ./tools/lint`
  - [x] `go run ./tools/coverage`
  - [x] `go run ./tools/depgate`
  - [x] `git diff --check`

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`; Epic 7 was added from the approved Correct Course and Story 7.1 is the first backlog story for the epic.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`; Story 7.1 acceptance criteria are BDD-formatted under `## Epic 7: CLI Composition Ergonomics`.
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`; FR-26 defines `cli` composition and NFR-2/NFR-5 preserve explicit caller-supplied inputs.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`; `cli/` is optional public composition, may depend on `command`, `flags`, and `config`, and must not own process lifecycle.
- Loaded Correct Course proposal: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-06-12-cli-composition.md`; GitHub tracker issues #46 through #51 were created and #45 was closed after application.
- No UX artifact; Dib V1 has no browser UI.

### Current Repository State

- Current baseline commit before Story 7.1 implementation: `8094bd1`.
- `cli/` does not exist yet. This story should create the package.
- Existing package docs establish the no-process-global pattern:
  - `command/doc.go`: command APIs do not read process args/env/stdout/stderr or default commands.
  - `flags/doc.go`: flag APIs parse caller-owned inputs and do not read process args or package-level default flag sets.
  - `config/doc.go`: config APIs do not read process args, live env, stdout/stderr, hidden caches, ambient config files, or package-level defaults.
- `README.md` has unrelated local changes in the working tree at story creation time. Do not edit or revert those changes in Story 7.1.

### Existing Code Patterns To Reuse

- Defensive copies use `append([]string(nil), values...)` for string slices:
  - `command/boundary.go` copies caller args in `NewBoundary` and returns copies from `Boundary.Args()`.
  - `command/result.go` copies `matchTokens` and `remaining`, and returns copies from `MatchTokens()` and `RemainingArgs()`.
  - `flags/snapshot.go` returns a copy from `Snapshot.RemainingArgs()`.
- Immutable public values use unexported fields plus accessor methods:
  - `command.Result`, `command.Boundary`, `flags.Snapshot`, and `config.Snapshot`.
- Typed error style:
  - Packages expose sentinel category errors, typed structs with safe context, and `Unwrap` methods for `errors.Is` / `errors.As`.
  - Examples: `command.ErrUnknownCommand`, `command.UnknownCommandError`, `flags.ErrInvalidDefinition`, `flags.ParseError`, `config.ErrUnknownSourceKey`, `config.SourceError`.
- Config flag binding already exists:
  - `config.FlagValue` carries `ConfigKey`, `ExplicitlySet`, and `Value`.
  - `config.NewFlagSnapshot` ignores `ExplicitlySet=false` values and uses `SourceFlagBinding`.
  - This story must not build translation from `flags.Snapshot` to `config.FlagValue`; that is Story 7.2.

### Architecture Guardrails

- `cli/` is a public package but not a root facade. Keep the module root unchanged.
- `cli/` may import only the standard library in Story 7.1. It should not need `command`, `flags`, or `config` until later stories.
- `cli/` must not call `os.Args`; `FromOSArgs` receives argv as an argument so callers make the process-global read explicitly.
- `cli/` must not call `os.Exit`, write stdout/stderr, read env implicitly, load files implicitly, execute callbacks, or hide errors behind rendered text.
- Empty full argv is the invalid boundary case for `FromOSArgs`. `FromArgs` is a lower-level explicit constructor and should not reject an empty program unless a later architecture decision changes the contract.
- Raw argv values can contain sensitive material. Error messages and typed error accessors must avoid echoing argv tokens.

### Prior-Art / Research Notes

- GitHub code search for `FromOSArgs` plus `Invocation` found only app-specific `os.Args` parsing examples, not a reusable pattern suitable for Dib.
- Exa search surfaced Go's standard `flag` documentation and source, which document package-level parsing from `os.Args[1:]` and default `CommandLine` behavior. That is useful as an anti-pattern for this story: Dib should require callers to pass argv explicitly and should not introduce a default global parser. [Source: https://pkg.go.dev/flag@go1.26.3]
- Exa also surfaced a small external Go CLI framework (`mz.attahri.com/code/argv`) that owns a broader program/mux/callback model. Do not adopt or copy that approach; it conflicts with Dib's zero-runtime-dependency and non-framework constraints. [Source: http://github.com/mzattahri/cli]
- No package-registry dependency is appropriate for Story 7.1; implement the small value object with the standard library.

### Git Intelligence

- Recent commits:
  - `8094bd1 feat(story-6.4): Reconcile Release Evidence And Tracker State`
  - `00b8388 feat(story-6.3): Publish Public Usage Documentation`
  - `c8c720b feat(story-6.2): Add Coverage Validation`
  - `9209b31 feat(story-6.1): Add an Isolated Linter Gate`
  - `7cfdfca docs(bmad): add epic 6 release hardening plan`
- Recent stories consistently update sprint status only as part of story workflow, record exact verification commands, and keep file lists accurate.

### File Structure Requirements

- **NEW targets**:
  - `cli/doc.go`
  - `cli/invocation.go`
  - `cli/errors.go`
  - `cli/invocation_test.go`
- **UPDATE targets during implementation**:
  - `_bmad-output/implementation-artifacts/7-1-add-explicit-cli-invocation-boundaries.md` for Dev Agent Record updates.
  - `_bmad-output/implementation-artifacts/sprint-status.yaml` for normal dev-story status transitions.
- **Do not touch in Story 7.1**:
  - `README.md`
  - `docs/behavior-matrices.md`
  - `docs/release-notes-v0.md`
  - `docs/release-checklist.md`
  - `examples/`
  - `command/`
  - `flags/`
  - `config/`
  - `tools/`

### Anti-Patterns To Avoid

- Do not name the package root or create a broad facade API.
- Do not add a package-global `DefaultInvocation`, `CommandLine`, or singleton.
- Do not import `os` from non-test `cli` files.
- Do not expose mutable fields or return the internal args slice.
- Do not echo raw argv tokens in `Error()` output.
- Do not use reflection, generics, or helper abstractions for this simple value object.
- Do not add callbacks, execution behavior, config source loading, flag parsing, or command routing here.
- Do not change coverage thresholds just because `cli/` starts with a small API; `go run ./tools/coverage` currently covers only existing runtime packages until Story 7.4 updates evidence for `cli`.

### Validation Checklist Applied

- Story includes exact story ID/key (`7-1-add-explicit-cli-invocation-boundaries`), ready-for-dev status, role/action/benefit, BDD-derived acceptance criteria, and task mapping to ACs.
- Story identifies every expected new file and explicitly bans unrelated runtime/docs/tooling changes.
- Story includes architecture guardrails for no process globals, no exits, no stream mutation, no implicit env/file reads, and no callback execution.
- Story includes existing code patterns for defensive copies, immutable public values, and typed errors.
- Story captures prior-art research and rejects dependency/framework adoption.
- Story records GitHub tracker issue #47 and current baseline commit.

### Story Completion Status

- Status set to `ready-for-dev`.
- Completion note: Ultimate context engine analysis completed - comprehensive developer guide created.

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

2026-06-12:
- Red phase: `GOCACHE=/tmp/dib-go-cache go test ./cli` failed because `cli` had no non-test Go files after adding `cli/invocation_test.go`.
- Green phase: `GOCACHE=/tmp/dib-go-cache go test ./cli` passed after adding the invocation implementation and typed errors.

### Completion Notes List

- Added the optional `cli` package with immutable `Invocation` values for explicit program and user-args boundaries.
- Added `FromOSArgs` for caller-supplied full argv and `FromArgs` for already stripped args; both preserve caller ownership and copy args defensively.
- Added typed invocation diagnostics with `ErrInvalidInvocation`, `InvocationError`, `Unwrap`, `Category`, and `Problem` accessors without echoing argv contents.
- Added external package tests proving explicit argv wins over ambient `os.Args`, constructor/accessor defensive copies, empty-program ownership for `FromArgs`, and typed error inspection.
- Added QA automation coverage for public invocation reuse, explicit empty full argv diagnostics, and value-free fake-sensitive diagnostic regression.
- Verification passed with `GOCACHE=/tmp/dib-go-cache`: `go test ./cli ./...`, `go vet ./...`, `go run ./tools/lint`, `go run ./tools/coverage`, `go run ./tools/depgate`, and `git diff --check`.

### File List

- `_bmad-output/implementation-artifacts/7-1-add-explicit-cli-invocation-boundaries.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/implementation-artifacts/tests/test-summary.md`
- `cli/doc.go`
- `cli/errors.go`
- `cli/invocation.go`
- `cli/invocation_qa_test.go`
- `cli/invocation_test.go`

### Change Log

- 2026-06-12: Implemented Story 7.1 explicit CLI invocation boundaries and marked the story ready for review.
- 2026-06-12: Added Story 7.1 QA automation tests and test summary evidence.
- 2026-06-12: Senior developer review approved Story 7.1 and marked it done.

## Senior Developer Review (AI)

### Review Summary

- Reviewer: GPT-5 Codex
- Review date: 2026-06-12
- Outcome: Approve
- Auto-fix result: No source fixes required; no verified critical, high, or medium issues remained.

### Findings

- No critical issues found.
- No high-severity acceptance criteria gaps found.
- No medium-severity implementation or documentation discrepancies found.

### Acceptance Criteria Verification

- AC1: Passed. `cli.FromOSArgs(argv)` uses caller-supplied full argv, exposes `Program()` from `argv[0]`, exposes `Args()` from `argv[1:]`, and non-test `cli` code imports only `errors` and `fmt`, with no `os.Args` read.
- AC2: Passed. `cli.FromArgs(program, args)` preserves the caller-supplied program and defensively copies args.
- AC3: Passed. Empty full argv returns zero-value `Invocation` plus typed `*cli.InvocationError`; `errors.Is(err, cli.ErrInvalidInvocation)` and `errors.As` are covered.
- AC4: Passed. Constructor and accessor defensive-copy tests prove mutations to original argv/args and returned args do not change observable invocation state.

### Task Audit

- All checked tasks are supported by implementation or verification evidence.
- Story File List matches the application/source files reviewed plus workflow artifacts.
- No prohibited Story 7.1 scope was found in `command/`, `flags/`, `config/`, README, release docs, examples, coverage thresholds, lint tooling, or dependency-gate behavior.

### Files Reviewed

- `cli/doc.go`
- `cli/errors.go`
- `cli/invocation.go`
- `cli/invocation_test.go`
- `cli/invocation_qa_test.go`
- `_bmad-output/implementation-artifacts/tests/test-summary.md`

### Validation

- `GOCACHE=/tmp/dib-go-cache go test ./cli ./...` - PASS
- `GOCACHE=/tmp/dib-go-cache go vet ./...` - PASS
- `GOCACHE=/tmp/dib-go-cache go run ./tools/lint` - PASS
- `GOCACHE=/tmp/dib-go-cache go run ./tools/coverage` - PASS
- `GOCACHE=/tmp/dib-go-cache go run ./tools/depgate` - PASS
- `git diff --check` - PASS
- `GOCACHE=/tmp/dib-go-cache go list -f '{{.ImportPath}} {{join .Imports ","}}' ./cli` confirmed runtime imports are standard library only: `errors,fmt`.
- Local Go docs for `errors.Is` and `errors.As` were checked with `go doc` to validate typed error inspection behavior.
