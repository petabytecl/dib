---
baseline_commit: e4547f1
created: "2026-06-12T00:54:14-04:00"
---

# Story 3.2: Resolve Aliases And Unknown Commands Predictably

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Go CLI developer,
I want command aliases and command failures to resolve predictably,
so that users can use shortcuts while tests still assert canonical command behavior.

## Requirements Trace

- FR1: define root and nested command trees with stable names, descriptions, aliases, and usage metadata.
- FR20: command behavior must be covered by executable tests and traceable behavior-matrix rows.
- NFR1: runtime packages, tests, examples, and tooling must remain standard-library-only unless architecture changes.
- NFR2: command routing APIs must use explicit command instances and caller-supplied args; no package-global command registry or default root command.
- NFR3: public command failures needed by callers must be inspectable without string matching.
- NFR4: routing snapshots and diagnostics must be deterministic enough for stable assertions.
- NFR5: library routing paths must not call `os.Exit`, read `os.Args`, or write to process stdout/stderr.
- NFR6: routing must be testable with table-driven unit tests and injected argument slices.
- NFR7: aliases may feel familiar to Cobra users, but Dib must not copy Cobra APIs, source layout, tests, examples, fixtures, or compatibility promises.

## Acceptance Criteria

1. Given a command defines aliases, when input uses an alias, then routing resolves to the intended command, and the route snapshot preserves the canonical command name and the raw alias token.
2. Given aliases can collide with command names or other aliases, when a command tree is built or derived, then collisions fail during setup with typed deterministic diagnostics, and alias-command collisions, alias-alias collisions, and alias cycles are covered by tests.
3. Given unknown command input is supplied near valid commands or aliases, when routing fails, then the typed unknown-command diagnostic identifies the failing token and matched parent path, and diagnostics remain deterministic without string-only assertions.
4. Given alias support must not introduce hidden state, when verification runs, then repeated and concurrent route tests observe stable results from the same reusable command definitions, and no root facade, global command registry, or `/cmd` scaffold is introduced.

## Tasks / Subtasks

- [x] Confirm tracker, artifacts, and source state (AC: 1-4)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Story 3.2 `ready-for-dev`, Story 3.1 `done`, and Epic 3 `in-progress`.
  - [x] Check for Story 3.2 ATDD/test artifacts under `_bmad-output/test-artifacts/`; none existed at story creation, but use them if generated before implementation.
  - [x] Verify `go.mod` still contains only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or generated `go.sum`.
  - [x] Read the current command source before editing: `command/definition.go`, `command/route.go`, `command/result.go`, `command/errors.go`, `command/definition_test.go`, `command/route_test.go`, `command/result_test.go`, `command/errors_test.go`, `command/contract_test.go`, and `command/doc.go`.

- [x] Add alias-aware route lookup without changing process ownership (AC: 1, 3, 4)
  - [x] Extend routing so each next command token matches a child by canonical `Name()` or by one of that child's aliases.
  - [x] Keep canonical route results canonical: `Result.Path()`, `Result.PathNames()`, and `Result.Command()` must still expose the matched command definitions and canonical names, not aliases.
  - [x] Add route snapshot state for raw matched input tokens, for example `Result.MatchTokens() []string` or a similarly small accessor. It must include the root command name first, then each raw token that matched a child, preserving aliases exactly as supplied by the caller.
  - [x] Ensure `Result.MatchTokens()` or equivalent returns defensive copies and remains unchanged after caller args or returned slices are mutated.
  - [x] Preserve existing Story 3.1 routing boundaries: `--` stops child matching and is omitted from remaining args; when the current command has no children, the next token and rest are remaining args; failed routes return zero-value results.
  - [x] Do not parse long flags, shorthand flags, shorthand groups, values, help requests, command-local flags, or inherited flags. Story 3.3 owns flag composition and must reuse `flags/` parser contracts.

- [x] Add setup-time alias validation with inspectable deterministic errors (AC: 2)
  - [x] Validate alias lookup tokens when aliases are set through `command.Aliases(...)` and `Definition.WithAliases(...)`, and when command trees are built through `command.Children(...)` and `Definition.WithChildren(...)`.
  - [x] Reject sibling lookup collisions within each parent scope: duplicate child names, duplicate aliases across siblings, alias matching another child's canonical name, a child's alias equal to its own canonical name, and cross-alias cycles such as child `apply` aliasing `plan` while child `plan` aliases `apply`.
  - [x] Reject blank or whitespace-only aliases during setup. Use the existing `*command.NameError` only if it remains semantically accurate; otherwise add a narrowly scoped typed validation error in `command/errors.go`.
  - [x] Add a typed collision diagnostic, for example `command.ErrDuplicateCommandToken` plus `*command.TokenConflictError`, that supports `errors.Is` and/or `errors.As` and exposes parent path if available, the conflicting lookup token, the first canonical child, and the colliding canonical child.
  - [x] Keep diagnostic strings deterministic but non-contractual. Tests must assert sentinel/typed error identity and accessors, not exact error wording.
  - [x] Do not add alias-to-alias chaining or recursive alias expansion. One input token resolves directly to one child in the current parent scope; validation rejects ambiguous trees instead of ranking matches.

- [x] Preserve reusable immutable definition behavior (AC: 1, 2, 4)
  - [x] Keep `command.NewDefinition("name", command.Aliases(...))` source-compatible and keep option/derivation APIs value-oriented.
  - [x] Ensure alias validation does not mutate receiver definitions or caller-owned alias/child slices.
  - [x] Ensure derived trees returned from `WithChildren` remain independent from originals after validation failures and successes.
  - [x] If route lookup maps are introduced, keep them unobservable, deterministic, concurrency-safe, and derived from immutable definition state. Do not add package-level registries or hidden global caches.

- [x] Add focused command package tests (AC: 1-4)
  - [x] Add alias success tests for root-level and nested aliases, including a case like `ship apply` resolving to canonical path `["dib", "deploy", "apply"]` while match tokens preserve `["dib", "ship", "apply"]`.
  - [x] Add tests proving alias and canonical paths produce identical canonical `PathNames()` and `Command()` results while preserving distinct raw match tokens.
  - [x] Add setup validation tests for duplicate sibling names, alias-vs-child-name collisions, alias-vs-alias collisions, self-alias, blank aliases, and cross-alias cycles. Assert typed diagnostics with `errors.Is`/`errors.As` and accessors.
  - [x] Add unknown-command tests near aliases: typo of an alias at root, typo under an alias-matched parent, and a token that is an alias in a different sibling scope. Assert `command.ErrUnknownCommand`, `*command.UnknownCommandError.Token()`, and `ParentPath()` without string matching.
  - [x] Extend defensive snapshot tests to cover match tokens, caller arg mutation after alias routing, returned slice mutation, repeated route calls, and concurrent alias route calls.
  - [x] Extend process-state tests similar to `command/contract_test.go` so alias routing still ignores misleading `os.Args`, env vars, stdout, and stderr.

- [x] Update docs only where executable evidence exists (AC: 1-4)
  - [x] Update `docs/behavior-matrices.md` with Story 3.2 command alias and alias-validation rows that point to exact command test function names.
  - [x] Update `docs/diagnostics-and-errors.md` with the alias collision diagnostic category only after the implementation exposes the actual typed/sentinel contract.
  - [x] Update `docs/provenance-log.md` only if implementation or docs were influenced by external material. Do not copy source, tests, fixtures, examples, names, or structure from Cobra, pflag, Viper, Go `flag`, or other CLI projects.

- [x] Preserve Story 3.2 scope boundaries (AC: 1-4)
  - [x] Do not implement local or inherited flags, help/usage rendering, callback invocation, `context.Context` execution, config binding, compatibility adapters, migration examples, shell completion, `/cmd` scaffolding, root facade APIs, or package-global registries.
  - [x] Do not add external runtime, test, assertion, fuzzing, or tooling dependencies.
  - [x] Do not create broad shared helpers or `internal/` packages unless a second concrete call site proves the need and architecture boundaries still hold.
  - [x] Do not modify `flags/` parser internals for Story 3.2; alias routing is command-token lookup only.

- [x] Verify the story implementation (AC: 1-4)
  - [x] Run focused command tests, for example `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./command -count=1`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`.
  - [x] Run `git diff --check`.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [x] Consider `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./... -count=1` because this story expands reusable command-tree validation and concurrent alias route lookup.
  - [x] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`.
- No UX artifact was discovered; Dib V1 has no browser UI or frontend UX surface.
- No `project-context.md` file was discovered under the repository.
- No Story 3.2 ATDD/test artifact was discovered under `_bmad-output/test-artifacts/`.
- Loaded previous story intelligence from `_bmad-output/implementation-artifacts/3-1-route-nested-commands-with-inspectable-results.md`.
- Loaded current command source and tests listed in the tasks above.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `e4547f1` (`feat(story-3.1): route nested commands`).
- Existing worktree has unrelated BMAD configuration, `.agents/`, `.codex/`, `.idea/`, story-automator, and installer changes. Do not revert or normalize them while implementing this story.
- Root `go.mod` currently contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

- No `go.sum` file exists at story creation.
- `sprint-status.yaml` has Epic 3 `in-progress`, Story 3.1 `done`, and Story 3.2 moved to `ready-for-dev` by this create-story workflow.

### Architecture Guardrails

- `command/` is the public command routing package. It owns command trees, nested routing, aliases, local/inherited flag attachment, help/usage rendering entry points, and command-specific typed errors. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- Callback handling is deferred. Dib does not invoke callbacks unless a later architecture/API decision explicitly adds that surface. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- The module root must not provide a broad public facade, package-global helpers, or default singleton API. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- `command/` may attach or accept `flags/` definitions and snapshots, but shared flag metadata and parsing semantics live in `flags/`, not `command/`. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- `flags/` must remain usable without `command/` or `config/`; `command/` must not depend on `config/`. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- Definitions and snapshots are reusable values. Derived definitions return new values, per-run snapshots do not mutate definitions, and exported APIs must not expose mutable aliases. [Source: `_bmad-output/planning-artifacts/architecture.md#State-Management-Patterns`]
- Public errors must support Go error inspection where callers need programmatic handling; error strings are diagnostics, not programmatic contracts. [Source: `_bmad-output/planning-artifacts/architecture.md#Error-Handling-Patterns`]
- Tests live beside the package under test in `*_test.go`; docs and examples must not claim behavior without executable package evidence. [Source: `_bmad-output/planning-artifacts/architecture.md#Structure-Patterns`]
- Dependency enforcement is owned by `tools/depgate/`; do not create alternate dependency gates. [Source: `_bmad-output/planning-artifacts/architecture.md#Development-Workflow-Integration`]

### Current Code Context

- `command/definition.go` defines immutable `Definition` values with unexported fields: `name`, `description`, `aliases`, `usage`, and `children`.
- `command.NewDefinition(name string, options ...Option)` validates blank command names with `*command.NameError`, applies options, and returns a value. Preserve this source-compatible constructor.
- `command.Aliases(...)` and `Definition.WithAliases(...)` currently record alias metadata without blank-alias, self-alias, or tree-collision validation. Story 3.2 owns setup-time alias validation.
- `command.Children(...)` and `Definition.WithChildren(...)` currently validate only blank child names and clone child definitions. Extend these setup boundaries rather than adding validation during route calls.
- `command/route.go` currently routes only by canonical child name through `childByName`. Story 3.2 should replace or extend this lookup to consider aliases in the current parent scope.
- `command/result.go` currently stores canonical path definitions and remaining args. Add raw matched-token snapshot state without changing existing accessors' canonical behavior.
- `command/errors.go` currently exposes `command.ErrUnknownCommand` and `*command.UnknownCommandError` with `Token()` and `ParentPath()`. Preserve this contract and add narrowly scoped setup diagnostics for alias collisions.
- `command/route_test.go` includes a Story 3.1 assertion that alias metadata is not routed. Update that test now that Story 3.2 deliberately routes aliases.
- `docs/behavior-matrices.md` currently marks command definitions, route snapshots, unknown diagnostics, and process boundaries as current for Story 3.1. Add Story 3.2 rows only after exact tests exist.
- `docs/diagnostics-and-errors.md` currently documents Story 3.1 unknown-command diagnostics. Extend it only for the actual alias-validation error contract implemented.

### Previous Story Intelligence

- Story 3.1 implemented immutable command definition metadata, child command composition, explicit `Definition.Route(args)`, canonical route snapshots, `--` boundary behavior, and inspectable unknown-command diagnostics.
- Story 3.1 deliberately stored alias metadata but did not route by aliases or validate alias collisions. Story 3.2 is the intended continuation point for that behavior.
- Story 3.1 route failures return zero-value results plus typed errors; preserve this for alias validation/routing failures.
- Story 3.1 review fixed artifact hygiene issues. Keep the Dev Agent Record file list exact and include `_bmad-output/implementation-artifacts/tests/test-summary.md` if QA automation updates it.
- Epic 2 retrospective and Story 3.1 both warned that command routing must not duplicate `flags/` parser syntax. Story 3.2 should inspect command tokens only; Story 3.3 owns flag composition.

### Git Intelligence

- Recent commits:
  - `e4547f1 feat(story-3.1): route nested commands`
  - `4c1c0fb docs: add epic 2 retrospective`
  - `7813af2 feat(story-2.8): harden parser fuzz evidence`
  - `a920b40 feat(story-2.7): preserve parse boundaries`
  - `760ea1c feat(story-2.6): accumulate repeated custom values`
- Story 3.1 touched `command/definition.go`, `command/route.go`, `command/result.go`, `command/errors.go`, command package tests, `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, `sprint-status.yaml`, story artifacts, and the test summary artifact.
- Existing pattern: add focused package-local tests first, keep public contracts inspectable through accessors and `errors.Is`/`errors.As`, document evidence rows after tests exist, and run `go test`, `go vet`, dependency gate, and `git diff --check`.

### Technical Research Notes

- Official Go downloads list `go1.26.4` as the current stable Go 1.26 release on 2026-06-12; keep the module target at `go 1.26` unless a separate story updates version policy. [Source: `https://go.dev/dl/`]
- Use the standard `errors` package inspection model for public routing and setup errors; callers should be able to use `errors.Is` and/or `errors.As` rather than string matching. [Source: `https://pkg.go.dev/errors`]
- Use the standard `testing` package and table-driven tests. No third-party assertion, mocking, or test framework is needed. [Source: `https://pkg.go.dev/testing`]

### Testing Standards

- Treat package tests as executable truth; docs must point to tests that actually exist after implementation.
- Use table-driven tests for alias routing success, alias collision validation, unknown-command failures, immutability, defensive copies, and process-state isolation.
- Assert typed diagnostics with `errors.Is` and/or `errors.As` whenever caller inspection is part of the contract.
- For deterministic snapshots, compare canonical path names, raw match tokens, and remaining args exactly; mutate caller-provided args and returned slices after routing to prove snapshot independence.
- Prefer focused updates in `command/route_test.go`, `command/result_test.go`, `command/errors_test.go`, and `command/definition_test.go` over broad integration tests.
- Keep any runnable examples standard-library-only and only add examples if they test real Story 3.2 behavior.

### Security And Quality Checks

- Do not hardcode secrets, credentials, tokens, private URLs, or host-specific paths.
- Do not copy source, tests, comments, fixtures, examples, internal names, command API shape, or file organization from Cobra, pflag, Viper, Go `flag`, or other CLI projects.
- Do not add external imports to runtime, tests, examples, fuzzing, or tooling.
- Do not add `os.Exit`, `os.Args`, hidden stdout/stderr use, `flag.CommandLine`, package-global mutable registries, or default singleton APIs.
- New docs must not claim command flags, help rendering, execution callbacks, config binding, compatibility parity, release readiness, or future API stability.

### Project Structure Notes

- Expected Story 3.2 source files are likely:
  - UPDATE `command/definition.go`
  - UPDATE `command/definition_test.go`
  - UPDATE `command/route.go`
  - UPDATE `command/route_test.go`
  - UPDATE `command/result.go`
  - UPDATE `command/result_test.go`
  - UPDATE `command/errors.go`
  - UPDATE `command/errors_test.go`
  - UPDATE `command/contract_test.go`
  - UPDATE `command/doc.go` only if alias routing changes package docs materially
  - UPDATE `docs/behavior-matrices.md`
  - UPDATE `docs/diagnostics-and-errors.md`
  - UPDATE `docs/provenance-log.md` only if external influence is recorded
- Do not create `examples/`, `internal/`, `/cmd`, `config/` implementation files, compatibility docs, or migration docs for this story.
- No structure conflict detected: the architecture already reserves command routing, alias handling, typed command errors, and command behavior-matrix evidence for Epic 3.

### Files To Read Before Editing

- `command/definition.go`: current state includes aliases as metadata and children as immutable values; extend setup validation here.
- `command/route.go`: current route lookup uses canonical child names only; extend lookup to direct alias matches.
- `command/result.go`: current route snapshot exposes canonical path and remaining args; add raw match-token state defensively.
- `command/errors.go`: current unknown-command error contract is established; add alias validation diagnostics without breaking it.
- `command/route_test.go`: current routing tests include alias-not-routed Story 3.1 expectation; update for Story 3.2 behavior.
- `command/result_test.go` and `command/errors_test.go`: extend defensive-copy coverage for match tokens and new diagnostic accessors.
- `command/contract_test.go`: extend process-state isolation to alias routing.
- `docs/behavior-matrices.md`: add Story 3.2 command alias evidence rows after exact test names exist.
- `docs/diagnostics-and-errors.md`: add Story 3.2 alias collision category after its public inspection API is implemented.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md#Story-32-Resolve-Aliases-And-Unknown-Commands-Predictably`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-1-Define-Command-Trees`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-20-Provide-Behavior-Test-Matrices`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#Non-Functional-Requirements`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Project-Structure-And-Boundaries`]
- [Source: `_bmad-output/implementation-artifacts/3-1-route-nested-commands-with-inspectable-results.md#Dev-Notes`]
- [Source: `command/definition.go`]
- [Source: `command/route.go`]
- [Source: `command/result.go`]
- [Source: `command/errors.go`]
- [Source: `docs/behavior-matrices.md`]
- [Source: `docs/diagnostics-and-errors.md`]
- [Source: `https://go.dev/dl/`]
- [Source: `https://pkg.go.dev/errors`]
- [Source: `https://pkg.go.dev/testing`]

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./command -count=1` -> PASS
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...` -> PASS
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...` -> PASS
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate` -> PASS
- `git diff --check` -> PASS
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./... -count=1` -> PASS
- `go.mod` remains module plus `go 1.26`; no `go.sum` exists.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Implemented alias-aware command routing with canonical path snapshots and raw `Result.MatchTokens()` snapshots.
- Added setup-time alias validation for blank/self aliases and ambiguous lookup tokens with `ErrInvalidCommandAlias`, `*AliasError`, `ErrDuplicateCommandToken`, and `*TokenConflictError`.
- Preserved explicit, reusable command definitions with no route-time global state, no process ownership, no flag parsing, and no new dependencies.
- Added command package tests for alias routing, alias diagnostics, unknown commands near aliases, defensive snapshots, repeated/concurrent route calls, and process-state isolation.
- Updated behavior matrix and diagnostics docs with executable Story 3.2 evidence.

### File List

- `_bmad-output/implementation-artifacts/3-2-resolve-aliases-and-unknown-commands-predictably.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/implementation-artifacts/tests/test-summary.md`
- `command/alias_workflow_test.go`
- `command/contract_test.go`
- `command/definition.go`
- `command/definition_test.go`
- `command/errors.go`
- `command/errors_test.go`
- `command/result.go`
- `command/result_test.go`
- `command/route.go`
- `command/route_test.go`
- `docs/behavior-matrices.md`
- `docs/diagnostics-and-errors.md`

### Change Log

- 2026-06-12: Implemented Story 3.2 alias routing, setup validation diagnostics, executable command tests, and evidence docs.
- 2026-06-12: Senior review fixed File List omissions for generated Story 3.2 workflow tests and test summary; all verification gates pass.

### Senior Developer Review (AI)

Reviewer: Coto on 2026-06-12

Outcome: Approved after auto-fix.

Findings fixed:

- [x] [AI-Review][Medium] `command/alias_workflow_test.go` was changed but missing from the Dev Agent Record File List.
- [x] [AI-Review][Medium] `_bmad-output/implementation-artifacts/tests/test-summary.md` was changed but missing from the Dev Agent Record File List.

Verification:

- [x] `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./command -count=1` - PASS
- [x] `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...` - PASS
- [x] `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...` - PASS
- [x] `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate` - PASS
- [x] `git diff --check` - PASS
- [x] `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./... -count=1` - PASS
