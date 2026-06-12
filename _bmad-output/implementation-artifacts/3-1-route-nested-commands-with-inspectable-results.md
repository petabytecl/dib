---
baseline_commit: 4c1c0fb
created: "2026-06-12T00:27:42-04:00"
---

# Story 3.1: Route Nested Commands With Inspectable Results

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Go CLI developer,
I want nested command input to return inspectable routing results,
so that I can build multi-command CLIs without giving a framework control over process lifecycle.

## Requirements Trace

- FR1: define root and nested command trees with stable names, descriptions, aliases, and usage metadata.
- FR20: command routing behavior must be covered by executable tests and traceable behavior-matrix rows.
- NFR1: runtime packages, tests, examples, and tooling must remain standard-library-only unless architecture changes.
- NFR2: command routing APIs must use explicit command instances and caller-supplied args; no package-global command registry or default root command.
- NFR3: unknown-command failures must be inspectable without string matching.
- NFR4: route snapshots, diagnostics, and docs must be deterministic enough for stable assertions.
- NFR5: library routing paths must not call `os.Exit`, read `os.Args`, or write to process stdout/stderr.
- NFR6: routing must be testable with table-driven unit tests and injected argument slices.
- NFR7: Dib may be familiar to Cobra-style command trees but must not copy Cobra APIs, source layout, fixtures, examples, or compatibility promises.

## Acceptance Criteria

1. Given a root command has nested children, when input such as `deploy apply` is routed, then the route snapshot identifies the canonical matched command path, and remaining args are preserved according to the flag parser boundary rules from Epic 2.
2. Given command definitions include names, descriptions, aliases, and usage metadata, when command definitions are derived or reused, then original definitions keep unchanged observable behavior, and snapshots do not mutate command definitions.
3. Given routing fails to find a command, when unknown command input is routed, then Dib returns a typed unknown-command error, and no runtime path calls `os.Exit`, writes directly to process streams, or reads `os.Args`.
4. Given route results are public contracts, when verification runs, then tests cover root routing, nested routing, remaining args, unknown commands, immutable definition reuse, and deterministic route snapshots, and `command/` consumes `flags/` contracts without reimplementing flag syntax.

## Tasks / Subtasks

- [x] Confirm tracker, artifacts, and source state (AC: 1-4)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Story 3.1 `ready-for-dev` and Epic 2 stories `done` before implementation starts.
  - [x] Check for Story 3.1 ATDD/test artifacts under `_bmad-output/test-artifacts/`; none existed at story creation, but use them if generated before implementation.
  - [x] Verify `go.mod` still contains only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or generated `go.sum`.
  - [x] Read the current `command/` package before editing: `command/definition.go`, `command/definition_test.go`, `command/contract_test.go`, and `command/doc.go`.

- [x] Extend command definitions as immutable tree values (AC: 1, 2)
  - [x] Keep `command.NewDefinition("name")` source-compatible by using variadic options if metadata/options are added.
  - [x] Add description, aliases, usage metadata, and child definitions as caller-observable immutable state. Accessors must return defensive copies for slices or any other mutable data.
  - [x] Add derivation APIs for children and metadata that return new `Definition` values and do not mutate the receiver or caller-owned slices.
  - [x] Validate blank command names through the existing `*command.NameError` contract. Add only narrowly needed setup-time validation for this story; alias collision diagnostics belong to Story 3.2.
  - [x] Preserve existing `NewDefinition` tests and example behavior.

- [x] Add route result and unknown-command error contracts (AC: 1, 3, 4)
  - [x] Add a public route entrypoint on explicit definitions, for example `Definition.Route(args []string) (Result, error)` or an equivalently small API under `command/`.
  - [x] Add a route snapshot/result type that exposes the canonical matched command path, stable path names, and remaining args through defensive-copy accessors.
  - [x] The route snapshot must be self-contained: later mutation or reuse of command definitions, child slices, aliases, or caller args must not change an existing result.
  - [x] Add an inspectable unknown-command error type and/or sentinel in `command/errors.go`. The error must expose the failing token and matched parent path for caller inspection, and must support Go error inspection with `errors.Is` and/or `errors.As`.
  - [x] Unknown command input must return a zero-value route result plus the typed error so callers cannot accidentally use partial routing state.

- [x] Implement canonical nested routing without process ownership (AC: 1, 3)
  - [x] Match child commands by canonical command name only in Story 3.1. Store alias metadata immutably, but do not route by aliases, detect alias collisions, or introduce alias cycles yet; Story 3.2 owns alias resolution.
  - [x] Route root input and nested input deterministically. `deploy apply` should match the root plus `deploy` plus `apply` path when those children exist.
  - [x] Use this routing boundary: if the current matched command has children and the next token is a non-`--` token that does not match a child name, return the typed unknown-command error; if the current matched command has no children, preserve that token and the rest as remaining args.
  - [x] Preserve Epic 2 boundary semantics for `--`: after the route has matched the command path, `--` prevents later tokens from being interpreted as command names; the route result must omit the `--` marker and preserve only post-terminator tokens in caller order, matching `flags.Snapshot.RemainingArgs()` behavior.
  - [x] Do not parse long flags, shorthand flags, shorthand groups, values, help requests, or command-local/inherited flags inside Story 3.1. Story 3.3 owns flag composition. This story must not duplicate `flags/parse.go` syntax logic.
  - [x] Do not invoke callbacks, run command handlers, accept `context.Context`, or model process lifecycle execution. Story 3.5 owns caller-controlled execution boundaries; architecture currently defers callback invocation.

- [x] Add focused command package tests (AC: 1-4)
  - [x] Add root routing tests: empty args or root-only input returns the root path and correct remaining args.
  - [x] Add nested routing tests for at least a two-level path such as `deploy apply`.
  - [x] Add remaining-args tests covering positionals after the matched command, flag-like tokens after the matched command, and `--` boundary behavior.
  - [x] Add unknown-command tests for a failing token at root and under a matched parent; assert typed error inspection and matched parent path without relying on error strings.
  - [x] Add immutable definition reuse tests proving derived definitions do not mutate originals, child/alias slices are defensively copied, caller args are defensively copied into results, and repeated/concurrent route calls observe deterministic results.
  - [x] Add explicit-process-state tests similar to `command/contract_test.go`: route while `os.Args`, env vars, stdout, and stderr are misleading, and prove routing uses only supplied args and returned values/errors.

- [x] Update behavior and diagnostics docs only where evidence exists (AC: 4)
  - [x] Update `docs/behavior-matrices.md` with Story 3.1 command routing rows that point to exact command test function names.
  - [x] Update `docs/diagnostics-and-errors.md` with the unknown-command error category only after the implementation exposes the actual typed/sentinel contract.
  - [x] Update `docs/provenance-log.md` only if implementation or docs were influenced by external material. Do not copy source, tests, fixtures, examples, names, or structure from Cobra, pflag, Viper, Go `flag`, or other CLI projects.

- [x] Preserve Story 3.1 scope boundaries (AC: 1-4)
  - [x] Do not implement alias routing/collision behavior, local or inherited flags, help/usage rendering, callback invocation, context cancellation, config binding, compatibility adapters, migration examples, shell completion, `/cmd` scaffolding, root facade APIs, or package-global registries.
  - [x] Do not add external runtime, test, assertion, fuzzing, or tooling dependencies.
  - [x] Do not create broad shared helpers or `internal/` packages unless a second concrete call site exists and the need is unavoidable.
  - [x] Do not modify `flags/` parser internals for Story 3.1 unless a command test exposes a proven exported-contract gap; prefer consuming exported `flags` contracts in later Story 3.3.

- [x] Verify the story implementation (AC: 1-4)
  - [x] Run focused command tests, for example `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./command -count=1`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`.
  - [x] Run `git diff --check`.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [x] Consider `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./... -count=1` because this story introduces reusable command-tree definitions and immutable route snapshots.
  - [x] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`.
- No UX artifact was discovered; Dib V1 has no browser UI or frontend UX surface.
- No `project-context.md` file was discovered under the repository.
- No Story 3.1 ATDD/test artifact was discovered under `_bmad-output/test-artifacts/`.
- Loaded current command source: `command/definition.go`, `command/definition_test.go`, `command/contract_test.go`, and `command/doc.go`.
- Loaded relevant flag parser contracts from `flags/parse.go`, `flags/snapshot.go`, `flags/errors.go`, `docs/behavior-matrices.md`, and `docs/diagnostics-and-errors.md`.
- Loaded previous implementation intelligence from Story 2.8 and Epic 2 retrospective.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `4c1c0fb` (`docs: add epic 2 retrospective`).
- Existing worktree has unrelated BMAD configuration, `.agents/`, `.codex/`, `.idea/`, story-automator, and installer changes. Do not revert or normalize them while implementing this story.
- Root `go.mod` currently contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

- No `go.sum` file exists at story creation.
- `sprint-status.yaml` has Epic 1 and Epic 2 done, Epic 3 moved to `in-progress`, and Story 3.1 moved to `ready-for-dev` by this create-story workflow.

### Architecture Guardrails

- `command/` is the public command routing package. It owns command trees, nested routing, aliases, local/inherited flag attachment, help/usage rendering entry points, and command-specific typed errors. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- Callback handling is deferred. Dib does not invoke callbacks unless a later architecture/API decision explicitly adds that surface. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- The module root must not provide a broad public facade, package-global helpers, or default singleton API. [Source: `_bmad-output/planning-artifacts/architecture.md#API-Boundaries`]
- `command/` may attach or accept `flags/` definitions and snapshots, but shared flag metadata and parsing semantics live in `flags/`, not `command/`. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- `flags/` must remain usable without `command/` or `config/`; `command/` must not depend on `config/`. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- Definitions and snapshots are reusable values. Derived definitions must return new values and avoid exported mutable internals or shallow-copy aliasing. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Architecture`]
- Public APIs use explicit inputs and returned values/errors, not package globals, hidden process IO, implicit environment reads, or root singletons. [Source: `_bmad-output/planning-artifacts/architecture.md#API-And-Communication-Patterns`]
- Public errors must support Go error inspection where callers need programmatic handling; error strings are diagnostics, not programmatic contracts. [Source: `_bmad-output/planning-artifacts/architecture.md#API-And-Communication-Patterns`]
- Tests live beside the package under test in `*_test.go`; docs and examples must not claim behavior without executable package evidence. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Organization`]
- Dependency enforcement is owned by `tools/depgate/`; do not create alternate dependency gates. [Source: `_bmad-output/planning-artifacts/architecture.md#Development-Workflow-Integration`]

### Current Code Context

- `command/definition.go` currently defines a minimal immutable `Definition` with one unexported `name` field, `NewDefinition(name string) (Definition, error)`, `Name() string`, and `*NameError` for blank names.
- `NewDefinition` rejects empty or whitespace-only names with `*command.NameError`; tests inspect it with `errors.As`.
- `command/contract_test.go` proves command definition construction uses caller-supplied names rather than `os.Args` or environment state. Reuse this pattern for routing tests.
- `command/doc.go` already states that command APIs should return per-run snapshots and avoid process args, environment variables, streams, hidden caches, and package-level default commands.
- No `command/route.go`, `command/result.go`, `command/errors.go`, `command/command.go`, `command/validation.go`, `command/route_test.go`, or `command/result_test.go` exists yet.
- `docs/behavior-matrices.md` currently marks command routing integration as `later`; Story 3.1 should convert only root/nested routing rows to `current` once executable tests exist.
- `docs/diagnostics-and-errors.md` currently has command diagnostic vocabulary but no concrete unknown-command category. Add only the actual category implemented by this story.

### Flag Parser Contracts To Preserve

- `flags.Set.Parse(args []string) (Snapshot, error)` is the single public parser entrypoint and never reads `os.Args`. [Source: `flags/parse.go`]
- `flags.Snapshot.RemainingArgs()` returns a defensive copy of positional args left after parsing flags. [Source: `flags/snapshot.go`]
- `flags.ParseError` exposes category, token, name, normalized name, and canonical definition, and unwraps public sentinels for Go error inspection. [Source: `flags/errors.go`]
- Parse boundary behavior from Epic 2: positionals interspersed with flags keep relative order; `--` stops flag parsing and preserves subsequent tokens; failed parses return zero-value snapshots; help requests return typed errors and never call `os.Exit`. [Source: `docs/behavior-matrices.md#Shared-Contracts`]
- Story 3.1 should not copy private helpers such as `isLongFlagToken`, `isShortFlagToken`, `splitLongFlag`, or `splitShortFlag` from `flags/parse.go`. If command routing needs flag-aware behavior, defer it to Story 3.3 where `command/` can consume exported `flags` contracts deliberately.

### Previous Story Intelligence

- Epic 2 completed the parser foundation: explicit `flags.Set` instances, exact/normalized long names, shorthand flags, shorthand groups, repeated/custom values, parse boundaries, help requests, typed diagnostics, behavior matrices, and fuzz evidence.
- Story 2.8 reinforced the implementation pattern: executable tests are the contract, docs point to exact test/fuzz names, fuzz/property inputs are clean-room and deterministic, and dependency gates remain standard-library-only.
- Epic 2 retrospective calls out the main Epic 3 risk: command routing must reuse `flags/` parser contracts and must not duplicate long, shorthand, group, terminator, help-request, or remaining-args parser logic.
- Recent story reviews corrected artifact/file-list drift. Keep the Dev Agent Record file list exact and include `_bmad-output/implementation-artifacts/tests/test-summary.md` if QA automation updates it.

### Git Intelligence

- Recent commits:
  - `4c1c0fb docs: add epic 2 retrospective`
  - `7813af2 feat(story-2.8): harden parser fuzz evidence`
  - `a920b40 feat(story-2.7): preserve parse boundaries`
  - `760ea1c feat(story-2.6): accumulate repeated custom values`
  - `1b9952e feat(story-2.5): parse short flag groups`
- Story 2.8 touched `docs/behavior-matrices.md`, `docs/provenance-log.md`, `flags/fuzz_test.go`, focused parser tests, fuzz seeds, `sprint-status.yaml`, and the story/test-summary artifacts.
- Story 2.7 touched `flags/errors.go`, `flags/parse.go`, `flags/parse_boundary_test.go`, `flags/fuzz_test.go`, `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, and boundary fuzz seeds.
- Existing pattern: add focused package-local tests first, keep public contracts inspectable through accessors and `errors.Is`/`errors.As`, document evidence rows after tests exist, and run `go test`, `go vet`, dependency gate, and `git diff --check`.

### Technical Research Notes

- Official Go downloads list `go1.26.4` as the current stable Go 1.26 release on 2026-06-12; keep the module target at `go 1.26` unless a separate story updates version policy. [Source: `https://go.dev/dl/`]
- Go module layout guidance supports simple package directories under the module root; do not add an application `/cmd` scaffold for a library story. [Source: `https://go.dev/doc/modules/layout`]
- Use the standard `errors` package inspection model for public routing errors; callers should be able to use `errors.Is` and/or `errors.As` rather than string matching. [Source: `https://pkg.go.dev/errors`]
- Use the standard `testing` package and table-driven tests. No third-party assertion, mocking, or test framework is needed. [Source: `https://pkg.go.dev/testing`]

### Testing Standards

- Treat package tests as executable truth; docs must point to tests that actually exist after implementation.
- Use table-driven tests for routing success, unknown-command failures, remaining args, immutability, defensive copies, and process-state isolation.
- Assert typed diagnostics with `errors.Is` and/or `errors.As` whenever caller inspection is part of the contract.
- For deterministic snapshots, compare returned path names and remaining args exactly; also mutate caller-provided args and source slices after routing to prove snapshot independence.
- Prefer focused tests in new `command/route_test.go`, `command/result_test.go`, and `command/errors_test.go` over broad integration tests.
- Keep any runnable examples standard-library-only and only add examples if they test real Story 3.1 behavior.

### Security And Quality Checks

- Do not hardcode secrets, credentials, tokens, private URLs, or host-specific paths.
- Do not copy source, tests, comments, fixtures, examples, internal names, command API shape, or file organization from Cobra, pflag, Viper, Go `flag`, or other CLI projects.
- Do not add external imports to runtime, tests, examples, fuzzing, or tooling.
- Do not add `os.Exit`, `os.Args`, hidden stdout/stderr use, `flag.CommandLine`, package-global mutable registries, or default singleton APIs.
- New docs must not claim alias routing, command flags, help rendering, execution callbacks, config binding, compatibility parity, release readiness, or future API stability.

### Project Structure Notes

- Expected Story 3.1 source files are likely:
  - UPDATE `command/definition.go`
  - UPDATE `command/definition_test.go`
  - UPDATE `command/contract_test.go`
  - UPDATE `command/doc.go`
  - NEW `command/route.go`
  - NEW `command/result.go`
  - NEW `command/errors.go`
  - NEW `command/route_test.go`
  - NEW `command/result_test.go`
  - NEW `command/errors_test.go`
  - UPDATE `docs/behavior-matrices.md`
  - UPDATE `docs/diagnostics-and-errors.md`
  - UPDATE `docs/provenance-log.md` only if external influence is recorded
- Do not create `examples/`, `internal/`, `/cmd`, `config/` implementation files, compatibility docs, or migration docs for this story.
- No structure conflict detected: the architecture already reserves `command/route.go`, `command/result.go`, `command/errors.go`, route/result tests, and command behavior-matrix evidence for Epic 3.

### Files To Read Before Editing

- `command/definition.go`: current state is a one-field immutable definition and `NameError`; preserve `NewDefinition("serve")` behavior.
- `command/definition_test.go`: current tests verify stable names, blank-name errors, and `ExampleNewDefinition`; update rather than discard.
- `command/contract_test.go`: current process-state test demonstrates explicit inputs; extend the same principle to route behavior.
- `command/doc.go`: current package docs already state snapshot/no-process-state guardrails; update only to reflect implemented routing.
- `docs/behavior-matrices.md`: add Story 3.1 command routing evidence rows after exact test names exist.
- `docs/diagnostics-and-errors.md`: add the unknown-command category after its public inspection API is implemented.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md#Story-31-Route-Nested-Commands-With-Inspectable-Results`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-1-Define-Command-Trees`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-20-Provide-Behavior-Test-Matrices`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#Non-Functional-Requirements`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Project-Structure-And-Boundaries`]
- [Source: `_bmad-output/implementation-artifacts/2-8-prove-flag-parsing-across-matrices-and-fuzz-inputs.md#Dev-Notes`]
- [Source: `_bmad-output/implementation-artifacts/epic-2-retro-2026-06-12.md#Next-Epic-Preview`]
- [Source: `command/definition.go`]
- [Source: `command/contract_test.go`]
- [Source: `docs/behavior-matrices.md`]
- [Source: `docs/diagnostics-and-errors.md`]
- [Source: `https://go.dev/dl/`]
- [Source: `https://go.dev/doc/modules/layout`]
- [Source: `https://pkg.go.dev/errors`]
- [Source: `https://pkg.go.dev/testing`]

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./command -count=1` red phase: failed to compile because Story 3.1 command metadata, child, routing, result, and unknown-command APIs did not exist yet.
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./command -count=1`: PASS.
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`: PASS.
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`: PASS.
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`: PASS.
- `git diff --check`: PASS.
- `test "$(sed -n '1p' go.mod)" = 'module github.com/petabytecl/dib' && test "$(sed -n '3p' go.mod)" = 'go 1.26' && ! rg -n '^(require|replace|toolchain)\b' go.mod && test ! -e go.sum`: PASS.
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./... -count=1`: PASS.

### Completion Notes List

- Implemented immutable command definition metadata, child command composition, and derivation APIs while preserving `command.NewDefinition("name")` compatibility.
- Added explicit `Definition.Route(args)` nested routing with self-contained route results, canonical path accessors, remaining args, and `--` boundary behavior.
- Added inspectable unknown-command diagnostics through `command.ErrUnknownCommand` and `*command.UnknownCommandError`; failed routes return zero-value results.
- Added focused command tests for routing, result immutability, unknown diagnostics, metadata derivation, process-state isolation, and deterministic concurrent route calls.
- Updated behavior matrix and diagnostics docs only for executable Story 3.1 evidence; no provenance-log update was needed because implementation was not influenced by external material.

### Senior Developer Review (AI)

Reviewer: GPT-5 Codex
Date: 2026-06-12
Outcome: Approved after auto-fix. No critical issues remain.

#### Review Scope

- Loaded the complete story file and confirmed status was `review`.
- Loaded `_bmad-output/planning-artifacts/architecture.md` for command, snapshot, typed-error, dependency, and test guardrails.
- Reviewed claimed File List against `git status --porcelain`, `git diff --name-only`, and untracked story source files.
- Reviewed command implementation and tests in `command/definition.go`, `command/route.go`, `command/result.go`, `command/errors.go`, and package tests.
- Reviewed Story 3.1 documentation evidence in `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, and `_bmad-output/implementation-artifacts/tests/test-summary.md`.

#### Findings Fixed

- [Medium] File List omitted `_bmad-output/implementation-artifacts/tests/test-summary.md`, which was modified by Story 3.1 QA automation. Fixed by adding it to the File List.
- [Low] Completion Notes contained an unrelated "Ultimate context engine" line that did not describe Story 3.1 implementation evidence. Fixed by removing the stray note.
- [Low] GitHub issue reflection was required by orchestration, but external sync was unavailable during review. Fixed locally by recording the issue target and failed sync attempts for retry.

#### Acceptance Criteria Validation

- AC1: Implemented. `Definition.Route(args)` returns deterministic root/nested path names and preserves leaf remaining args and `--` post-terminator args.
- AC2: Implemented. Metadata, aliases, usage, children, and route snapshots use immutable value/defensive-copy APIs; derivation tests prove originals are not mutated.
- AC3: Implemented. Unknown commands return zero-value results plus `command.ErrUnknownCommand` / `*command.UnknownCommandError`; routing code has no `os.Exit`, `os.Args`, stdout, or stderr use.
- AC4: Implemented. Tests cover root routing, nested routing, remaining args, unknown diagnostics, immutable definition reuse, deterministic snapshots, and process-state isolation. `command/` does not reimplement flag syntax beyond the Story 3.1 `--` route boundary.

#### GitHub Sync

- Target issue identified from story-automator logs: `petabytecl/dib#20`.
- Review-start and review-complete connector attempts to comment on issue #20 were cancelled by the host.
- Review-start and review-complete fallback `gh issue comment 20 --repo petabytecl/dib` attempts failed because `api.github.com` was unreachable from this workspace.
- Review evidence is recorded locally here so the issue can be retried when GitHub access is available.

### File List

- `_bmad-output/implementation-artifacts/3-1-route-nested-commands-with-inspectable-results.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/implementation-artifacts/tests/test-summary.md`
- `command/contract_test.go`
- `command/definition.go`
- `command/definition_test.go`
- `command/doc.go`
- `command/errors.go`
- `command/errors_test.go`
- `command/result.go`
- `command/result_test.go`
- `command/route.go`
- `command/route_test.go`
- `docs/behavior-matrices.md`
- `docs/diagnostics-and-errors.md`

## Change Log

- 2026-06-12: Implemented Story 3.1 nested command routing, immutable command tree metadata, inspectable unknown-command diagnostics, executable command tests, and evidence docs.
- 2026-06-12: Senior Developer Review approved Story 3.1 after artifact hygiene fixes; recorded GitHub sync blocker for issue #20.
