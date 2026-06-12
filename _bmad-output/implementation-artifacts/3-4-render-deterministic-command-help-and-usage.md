---
baseline_commit: 8fa4448
created: "2026-06-12T01:53:37-04:00"
---

# Story 3.4: Render Deterministic Command Help And Usage

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a CLI author,
I want deterministic help and usage output generated from definitions,
so that user-facing text is stable enough for review, examples, and tests.

## Requirements Trace

- FR4: command help and usage text must render from command definitions to caller-supplied writers.
- FR20: command help/usage behavior must be covered by executable tests and behavior-matrix evidence.
- NFR1: runtime packages, tests, examples, and tooling must remain standard-library-only unless architecture changes.
- NFR2: primary APIs must operate on explicit instances and caller-supplied inputs/outputs; no package-global command registry, default root, or ambient process state.
- NFR3: setup/rendering failures needed by callers must be inspectable without string matching.
- NFR4: help, usage, diagnostics, and route snapshots must be deterministic enough for stable golden or snapshot tests.
- NFR5: rendering must not call `os.Exit`, read `os.Args`, write to process-global stdout/stderr, or invoke callbacks.
- NFR6: behavior must be testable with table-driven tests and injected writers.
- NFR7: behavior may be familiar to Cobra/pflag users, but Dib must not copy APIs, source, tests, examples, fixtures, or compatibility promises.
- NFR8: help/usage and diagnostics must not leak raw sensitive values for sensitive flags.

## Acceptance Criteria

1. Given command definitions include names, aliases, descriptions, argument or usage metadata, and visible flags, when help or usage is rendered to a caller-supplied writer, then output includes those elements in deterministic order, and hidden flags do not appear while remaining parseable when their definitions allow it.
2. Given flags may be deprecated, when help or usage includes deprecated visible flags, then a deterministic deprecation note is rendered, and the note does not leak sensitive default values.
3. Given rendering is human-facing but still contractual, when tests assert help and usage output, then golden tests may verify formatting, and structured behavior is still asserted through definitions, snapshots, and typed diagnostics.
4. Given rendering must not execute commands, when help or usage is requested, then no callback invocation occurs, and no process-global stdout/stderr or `os.Exit` behavior is used.
5. Given documentation and examples depend on stable text, when verification runs, then command help/usage tests cover ordering, aliases, hidden flags, deprecated flags, inherited/local flags, and redaction-safe diagnostics, and all examples remain standard-library-only.

## Tasks / Subtasks

- [x] Confirm tracker, artifacts, and source state (AC: 1-5)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 3 `in-progress`, Stories 3.1-3.3 `done`, and Story 3.4 `ready-for-dev` before implementation starts.
  - [x] Check for Story 3.4 ATDD/test artifacts under `_bmad-output/test-artifacts/`; none existed at story creation, but use them if generated before implementation.
  - [x] Verify `go.mod` still contains only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or generated `go.sum`.
  - [x] Read current command and flags source before editing: `command/definition.go`, `command/flags.go`, `command/route.go`, `command/result.go`, `command/errors.go`, command package tests, `flags/definition.go`, `flags/set.go`, `flags/snapshot.go`, `flags/errors.go`, and `docs/diagnostics-and-errors.md`.

- [x] Add explicit writer-based command rendering APIs (AC: 1, 3, 4)
  - [x] Add help/usage rendering entry points under `command/`, likely in new `command/help.go` with tests in new `command/help_test.go`.
  - [x] Keep rendering explicit and caller-owned: accept `io.Writer`, return `error`, and do not write to `os.Stdout`, `os.Stderr`, buffers hidden in package globals, or process-owned streams.
  - [x] Preserve `Definition.Route(args)` behavior: unregistered `--help` and `-h` still return `flags.ErrHelpRequest` from the parser when flags are in scope; routing must not automatically render help, exit, or mutate process state.
  - [x] Expose help/usage for a specific `Definition` and, if needed, for a matched `Result` so callers can render the final matched command after routing without recomputing available flags.
  - [x] Use existing `Definition.Usage()` as the current argument/usage metadata unless implementation proves a narrower public metadata option is required. Do not introduce a broad argument-schema API for this story without tests and docs tying it directly to FR4.
  - [x] If a writer fails, return the writer error and a zero/unchanged observable state. Do not wrap writer failures in command routing errors unless a typed rendering error is intentionally added and documented.

- [x] Render command metadata deterministically (AC: 1, 3, 5)
  - [x] Include the command's canonical name, aliases, description, usage metadata, child commands, and visible flags in a stable order.
  - [x] Preserve definition registration order for aliases, children, inherited flags, and local flags. Do not sort unless the implementation documents and tests that ordering as the public contract.
  - [x] Render aliases deterministically and make clear that the canonical command name remains the programmatic route identity.
  - [x] Render child commands from `Definition.Children()` with names and descriptions only; do not route or parse while rendering.
  - [x] Keep all formatting deterministic across Go versions by avoiding map iteration, clock values, terminal width detection, environment variables, locale-sensitive output, or host-specific paths.
  - [x] Add exact-output tests for representative help and usage text. Golden fixtures are acceptable only if they are small, local, clean-room, and paired with structured assertions.

- [x] Render composed flags without duplicating parser semantics (AC: 1, 2, 5)
  - [x] Reuse Story 3.3 flag composition behavior and exported `flags` accessors: `flags.Set.Definitions()`, `flags.Definition.Name()`, `Shorthand()`, `Usage()`, `Default()`, `Hidden()`, `Deprecated()`, `Sensitive()`, `NoOptionDefault()`, `RepeatPolicy()`, and `Arity()`.
  - [x] For command-path help, compose inherited flags root-to-leaf and the final command's local flags using the same `composeFlags` path as routing. Do not create a second flag-composition implementation.
  - [x] Hidden flags must be omitted from help/usage output but remain parseable through `Route` when their definitions are otherwise available in scope.
  - [x] Visible deprecated flags must render a deterministic deprecation note. If the deprecation message is caller-provided, render it exactly or through a documented deterministic normalization; never append raw sensitive defaults.
  - [x] Sensitive flag defaults, no-option defaults, and parse values must not appear in help, usage, writer errors, rendered diagnostics, examples, or golden fixtures. Prefer omitting sensitive defaults entirely or rendering a fixed redaction marker if the package chooses to show default metadata.
  - [x] Non-sensitive defaults may be rendered only if the story tests exact formatting and kind handling. Keep default rendering small and deterministic; do not invent complex type formatting beyond values already exposed by `flags.Definition`.

- [x] Preserve routing, parser, and process boundaries (AC: 3, 4)
  - [x] Do not add callback invocation, execution APIs, context propagation, shell completion, generated man pages, config binding, compatibility adapters, `/cmd` scaffolding, root facade APIs, package-global registries, or default singleton APIs.
  - [x] Do not reinterpret flag tokens for rendering. `flags.Set.Parse` owns help request detection, unknown flags, missing values, duplicate values, conversions, `--`, and shorthand group behavior.
  - [x] Do not add package-global help templates, mutable global format settings, terminal width detection, color detection, environment lookup, or ambient process inspection.
  - [x] Ensure failed rendering does not mutate command definitions, route results, flag sets, flag snapshots, or returned slices.
  - [x] Preserve zero-value safety: zero-value `Definition` rendering should return an inspectable name/setup error rather than panic.

- [x] Add focused command help/usage tests (AC: 1-5)
  - [x] Add tests that render a root command and nested command with aliases, descriptions, usage metadata, children, inherited flags, and local flags.
  - [x] Add tests that assert deterministic ordering of command metadata, child commands, inherited flags, and final-command local flags.
  - [x] Add tests proving hidden flags are omitted from output while still parseable through `Route`.
  - [x] Add tests proving deprecated visible flags render a deterministic note.
  - [x] Add tests proving sensitive fake values `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value` never appear in help/usage output or rendering errors.
  - [x] Add tests proving rendering uses supplied writers only and ignores misleading `os.Args`, environment variables, `os.Stdout`, and `os.Stderr`.
  - [x] Add tests for writer failure propagation using a small test writer that returns a deterministic error.
  - [x] Add defensive/reuse tests: render the same definition repeatedly and concurrently, mutate returned slices after rendering, and confirm output remains stable.
  - [x] Add structured assertions alongside exact text checks, for example verifying `Route` still exposes hidden/deprecated flag definitions and snapshots through `Result.Flags()` / `Result.FlagSnapshot()`.

- [x] Update documentation only after executable evidence exists (AC: 3, 5)
  - [x] Update `docs/behavior-matrices.md` with Story 3.4 command help/usage rows after exact test names exist.
  - [x] Update `docs/diagnostics-and-errors.md` only if Story 3.4 adds a public rendering diagnostic or materially changes help-request guidance.
  - [x] Update package docs or examples only for implemented behavior. Any examples must be runnable Go example tests and standard-library-only.
  - [x] Update `docs/provenance-log.md` only if implementation or docs were influenced by external material. Do not copy source, tests, fixtures, examples, names, or structure from Cobra, pflag, Viper, Go `flag`, or other CLI projects.

- [x] Preserve Story 3.4 scope boundaries (AC: 1-5)
  - [x] Do not implement command execution, callback invocation, config resolution, config binding, migration examples, shell completion, man pages, generated assets, terminal color, localization, width-aware wrapping, or compatibility adapters.
  - [x] Do not modify `flags/` parser internals unless a focused command help/usage test exposes a proven exported-contract gap; prefer consuming exported `flags` metadata and parse contracts.
  - [x] Do not add external runtime, test, assertion, fuzzing, or tooling dependencies.
  - [x] Do not create broad shared helpers or `internal/` packages unless a second concrete call site proves the need and architecture boundaries still hold.

- [x] Verify the story implementation (AC: 1-5)
  - [x] Run focused command tests, for example `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./command -count=1`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`.
  - [x] Run `git diff --check`.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [x] Consider `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./... -count=1` because this story adds repeated/concurrent rendering of reusable command definitions.
  - [x] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD and addendum material from `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/`.
- No UX artifact was discovered; Dib V1 has no browser UI or frontend UX surface.
- No `project-context.md` file was discovered under the repository.
- No Story 3.4 ATDD/test artifact was discovered under `_bmad-output/test-artifacts/`.
- Loaded previous story intelligence from `_bmad-output/implementation-artifacts/3-3-apply-local-and-inherited-flags-predictably-during-routing.md`.
- Loaded current command and flags source listed in the tasks above, plus `docs/behavior-matrices.md` and `docs/diagnostics-and-errors.md`.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `8fa4448` (`feat(story-3.3): compose command flags`).
- Existing worktree has unrelated BMAD configuration, `.agents/`, `.codex/`, `.idea/`, story-automator, and installer changes. Do not revert or normalize them while implementing this story.
- Root `go.mod` currently contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

- No `go.sum` file exists at story creation.
- `sprint-status.yaml` has Epic 3 `in-progress`, Stories 3.1-3.3 `done`, and Story 3.4 moved to `ready-for-dev` by this create-story workflow.

### Architecture Guardrails

- `command/` owns command trees, nested routing, aliases, local/inherited flag attachment, help/usage rendering entry points, and command-specific typed errors. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- `command/` may attach or accept `flags/` definitions and snapshots for command-local and inherited flags. Shared flag metadata and parsing semantics live in `flags/`, not `command/`. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- `flags/` must remain fully usable without `command/` or `config/`; `command/` must not depend on `config/`. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- Definitions and snapshots are reusable values. Derived definitions return new values, per-run snapshots do not mutate definitions, and exported APIs must not expose mutable aliases. [Source: `_bmad-output/planning-artifacts/architecture.md#State-Management-Patterns`]
- Prefer returned values and errors over stdout/stderr side effects. Rendered strings are human-facing diagnostics but still product contracts. [Source: `_bmad-output/planning-artifacts/architecture.md#Format-Patterns`]
- Public errors must support Go error inspection where callers need programmatic handling; error strings are diagnostics, not programmatic contracts. [Source: `_bmad-output/planning-artifacts/architecture.md#Error-Handling-Patterns`]
- Tests live beside the package under test in `*_test.go`; runnable examples live in Go example tests where practical. [Source: `_bmad-output/planning-artifacts/architecture.md#Structure-Patterns`]
- `io.Writer` is the architecture-approved integration point for help, usage, and diagnostics. [Source: `_bmad-output/planning-artifacts/architecture.md#Integration-Points`]
- Callback handling remains deferred. Dib must not invoke callbacks unless a later architecture/API decision explicitly adds that surface. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]

### Current Code Context

- `command/definition.go` defines immutable `Definition` values with unexported `name`, `description`, `aliases`, `usage`, `children`, `localFlags`, `inheritedFlags`, and `flagNormalizer` fields. Add rendering behavior around this type rather than creating a parallel command type.
- `command.NewDefinition(name string, options ...Option)` validates blank names, applies options, validates aliases, and validates the flag-composition tree. Preserve this constructor and variadic option style.
- `Definition.Description()`, `Aliases()`, `Usage()`, `Children()`, `LocalFlags()`, and `InheritedFlags()` already expose the metadata Story 3.4 needs through defensive accessors.
- `command/flags.go` owns flag composition through `composeFlags(path []Definition)` and `newFlagSet(normalizer, definitions)`. Reuse this path for rendering available inherited/local flags so help output matches routing behavior.
- `command/route.go` uses `flags.Set.Parse` for flag-aware routing and keeps `flags.ErrHelpRequest` as a typed parser diagnostic. Rendering must not be coupled to route failure handling.
- `command/result.go` exposes `Result.Flags() (flags.Set, bool)` and `Result.FlagSnapshot() (flags.Snapshot, bool)`. If rendering from a route result is useful, consume these snapshots instead of recomposing private state.
- `command/errors.go` currently exposes `ErrUnknownCommand`, alias/token setup diagnostics, and `ErrFlagComposition`. Add a rendering error only if callers need structured rendering failure context beyond ordinary writer errors.
- `flags.Definition` already exposes `Name`, `Kind`, `Default`, `Usage`, `Shorthand`, `Hidden`, `Sensitive`, `Deprecated`, `NoOptionDefault`, `RepeatPolicy`, and `Arity`. Do not duplicate these metadata fields in `command/`.
- `flags.Set.Definitions()` returns definitions in deterministic registration order. Use that order when rendering available flags unless tests explicitly define a different stable ordering.
- `docs/behavior-matrices.md` currently has command rows through Story 3.3. Story 3.4 should add command help/usage evidence after tests exist.
- `docs/diagnostics-and-errors.md` already states that `flags.ErrHelpRequest` does not render help, call `os.Exit`, read `os.Args`, or write to stdout/stderr; preserve that boundary.

### Previous Story Intelligence

- Story 3.3 implemented `command/flags.go`, `command.LocalFlags`, `command.InheritedFlags`, `command.FlagNormalizer`, `Result.Flags`, and `Result.FlagSnapshot`.
- Story 3.3 established that inherited flags compose root-to-leaf, final-command local flags compose last, sibling local flags and ancestor local flags remain isolated, and conflicts are setup-time `command.ErrFlagComposition` errors wrapping underlying `flags` definition diagnostics.
- Story 3.3 explicitly did not implement help/usage rendering, hidden/deprecated flag rendering, callback invocation, execution APIs, config binding, compatibility adapters, migration examples, shell completion, `/cmd` scaffolding, root facade APIs, or package-global registries.
- Epic 2 completed exported parser behavior. The recurring warning from Stories 3.1-3.3 and Epic 2 is: command routing and rendering must not duplicate `flags/` parser syntax.
- Recent story reviews corrected artifact/file-list drift. Keep the Dev Agent Record file list exact and include `_bmad-output/implementation-artifacts/tests/test-summary.md` if QA automation updates it.

### Git Intelligence

- Recent commits:
  - `8fa4448 feat(story-3.3): compose command flags`
  - `1e7d1af feat(story-3.2): resolve command aliases`
  - `e4547f1 feat(story-3.1): route nested commands`
  - `4c1c0fb docs: add epic 2 retrospective`
  - `7813af2 feat(story-2.8): harden parser fuzz evidence`
- Story 3.3 touched `command/definition.go`, `command/flags.go`, `command/route.go`, `command/result.go`, `command/errors.go`, command package tests, behavior/diagnostics docs, sprint status, and test summary artifacts.
- Existing pattern: add focused package-local tests first, keep public contracts inspectable through accessors and `errors.Is`/`errors.As`, document evidence rows after tests exist, and run `go test`, `go vet`, dependency gate, race tests when state reuse/concurrency risk increases, and `git diff --check`.

### Technical Research Notes

- Official Go downloads list `go1.26.4` as the current stable Go 1.26 release on 2026-06-12; keep the module target at `go 1.26` unless a separate story updates version policy. [Source: `https://go.dev/dl/`]
- Use the standard `io.Writer` contract for caller-supplied rendering destinations. Return writer errors; do not hide them behind process exits or global streams. [Source: `https://pkg.go.dev/io#Writer`]
- Use the standard `errors` package inspection model for public routing, setup, parse, and any new rendering errors. [Source: `https://pkg.go.dev/errors`]
- Use the standard `testing` package and table-driven tests. No third-party assertion, mocking, golden, or test framework is needed. [Source: `https://pkg.go.dev/testing`]

### Testing Standards

- Treat package tests as executable truth; docs must point to tests that actually exist after implementation.
- Use table-driven tests for help/usage rendering across names, aliases, descriptions, usage metadata, children, inherited/local flags, hidden flags, deprecated flags, sensitive defaults, writer failure, explicit process boundaries, and repeated/concurrent rendering.
- Exact output tests should be small and deliberate. Golden files may verify human-facing formatting, but public behavior should also be asserted through definitions, route results, flag sets, snapshots, and typed diagnostics.
- Assert typed diagnostics with `errors.Is` and/or `errors.As` whenever caller inspection is part of the contract.
- For deterministic output, compare rendered bytes exactly and also compare structured sources such as `Definition.Children()`, `Result.Flags().Definitions()`, `Result.FlagSnapshot()`, and `Definition.Aliases()`.
- Keep fixtures local to `command/` if any are added. Do not depend on live env, current working directory, terminal width, wall clock, stdin/stdout, or host files.
- Prefer focused updates in `command/help_test.go`, `command/contract_test.go`, `command/result_test.go`, and `docs/behavior-matrices.md`. Touch `flags/` tests only if exported metadata behavior needs coverage for Story 3.4.

### Security And Quality Checks

- Use the architecture-owned fake sensitive-value corpus exactly: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.
- Do not hardcode real secrets, credentials, tokens, private URLs, or host-specific paths.
- Do not copy source, tests, comments, fixtures, examples, internal names, command API shape, or file organization from Cobra, pflag, Viper, Go `flag`, or other CLI projects.
- Do not add external imports to runtime, tests, examples, fuzzing, or tooling.
- Do not add `os.Exit`, `os.Args`, hidden stdout/stderr use, `flag.CommandLine`, package-global mutable registries, default templates, or default singleton APIs.
- New docs must not claim execution callbacks, config binding, compatibility parity, release readiness, or future API stability.

### Project Structure Notes

- Expected Story 3.4 source files are likely:
  - ADD `command/help.go`
  - ADD `command/help_test.go`
  - UPDATE `command/contract_test.go`
  - UPDATE `command/result.go` only if result-based rendering is added
  - UPDATE `command/result_test.go` only if result-based rendering is added
  - UPDATE `command/errors.go` and `command/errors_test.go` only if a public rendering diagnostic is added
  - UPDATE `command/doc.go` only if package docs need to mention implemented help/usage rendering
  - UPDATE `docs/behavior-matrices.md`
  - UPDATE `docs/diagnostics-and-errors.md` only if diagnostic/help-request guidance changes
  - UPDATE `docs/provenance-log.md` only if external influence is recorded
  - UPDATE `_bmad-output/implementation-artifacts/tests/test-summary.md` if QA automation updates it
- Do not create `examples/`, `internal/`, `/cmd`, `config/` implementation files, compatibility docs, migration docs, shell-completion assets, generated man pages, or root facade files for this story.
- No structure conflict detected: the architecture already reserves `command/help.go` and `command/help_test.go` for help/usage rendering.

### Files To Read Before Editing

- `command/definition.go`: current command definition metadata, options, derivation APIs, validation, flag fields, and clone behavior.
- `command/flags.go`: current inherited/local flag composition path to reuse for rendering.
- `command/route.go`: current route and parser-boundary behavior, especially `flags.ErrHelpRequest`.
- `command/result.go`: current route snapshot and flag accessor behavior.
- `command/errors.go`: existing command diagnostics and accessor style.
- `command/*_test.go`: established command tests for metadata, routing, aliases, flags, errors, snapshots, and process boundaries.
- `flags/definition.go`: exported flag metadata needed for rendering hidden, deprecated, sensitive, default, shorthand, kind, repeat, and arity details.
- `flags/set.go`: deterministic definition ordering and lookup behavior.
- `flags/snapshot.go`: snapshot/value accessors for structured assertions.
- `flags/errors.go`: parser/setup diagnostics that rendering must not replace.
- `docs/behavior-matrices.md`: add Story 3.4 evidence rows after exact tests exist.
- `docs/diagnostics-and-errors.md`: preserve help-request and redaction guidance; update only for implemented diagnostics.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md#Story-34-Render-Deterministic-Command-Help-And-Usage`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-4-Generate-deterministic-help-and-usage-text`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#State-Management-Patterns`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Format-Patterns`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Integration-Points`]
- [Source: `_bmad-output/implementation-artifacts/3-3-apply-local-and-inherited-flags-predictably-during-routing.md#Dev-Notes`]
- [Source: `command/definition.go`]
- [Source: `command/flags.go`]
- [Source: `command/route.go`]
- [Source: `command/result.go`]
- [Source: `command/errors.go`]
- [Source: `flags/definition.go`]
- [Source: `flags/set.go`]
- [Source: `flags/snapshot.go`]
- [Source: `docs/behavior-matrices.md`]
- [Source: `docs/diagnostics-and-errors.md`]

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./command -count=1` — failed before implementation because `WriteHelp` / `WriteUsage` were undefined; passed after implementation.
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...` — passed.
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...` — passed.
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate` — passed.
- `git diff --check` — passed.
- `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./... -count=1` — passed.
- `go.mod` remains `module github.com/petabytecl/dib` plus `go 1.26`; no `go.sum` exists.

### Completion Notes List

- Added writer-based `Definition.WriteUsage`, `Definition.WriteHelp`, `Result.WriteUsage`, and `Result.WriteHelp` APIs with direct writer-error propagation and zero-value name validation.
- Rendered deterministic usage/help text from command metadata and composed flag definitions without routing side effects, process-global streams, callbacks, terminal/environment inspection, or default value rendering.
- Omitted hidden flags from help while preserving parseability through route results; rendered visible deprecated flags with deterministic notes; kept sensitive defaults, no-option defaults, and parsed values out of rendered output and writer errors.
- Added exact-output and structured command tests for aliases, children, usage metadata, inherited/local flags, hidden/deprecated/sensitive flags, writer failures, process boundaries, repeat/concurrent rendering, and unchanged `flags.ErrHelpRequest` routing behavior.
- Updated `docs/behavior-matrices.md` with Story 3.4 help/usage rendering evidence after tests existed.

### Senior Developer Review (AI)

Reviewer: GPT-5 Codex on 2026-06-12

Outcome: Approved after automatic artifact fixes.

Findings fixed automatically:

- [x] [AI-Review][Medium] Story file list omitted `command/help_qa_test.go`, even though the QA workflow added executable Story 3.4 tests there.
- [x] [AI-Review][Medium] `docs/behavior-matrices.md` did not cite the QA tests for direct definition-local rendering and `WriteUsage` writer/invalid-target behavior.
- [x] [AI-Review][Low] Completion notes contained an unrelated "Ultimate context engine analysis" entry not tied to Story 3.4 implementation evidence.

Review notes:

- Acceptance Criteria 1-5 are implemented by `command/help.go`, `command/help_test.go`, and `command/help_qa_test.go`.
- The renderer uses caller-supplied `io.Writer` values, composed flag metadata, deterministic definition order, and canonical route paths. No callback execution, `os.Exit`, `os.Args`, stdout/stderr writes, terminal width lookup, or environment-dependent rendering was found.
- Hidden flags remain in the composed route flag set while being omitted from help text. Sensitive defaults, no-option defaults, and parsed values are not rendered.
- Story context and architecture guidance were loaded from the story file and `_bmad-output/planning-artifacts/architecture.md`. No `project-context.md` file was present. MCP documentation search was not applicable for this standard-library-only Go review; local architecture and package documentation were used.

### File List

- `_bmad-output/implementation-artifacts/3-4-render-deterministic-command-help-and-usage.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/implementation-artifacts/tests/test-summary.md`
- `command/help.go`
- `command/help_qa_test.go`
- `command/help_test.go`
- `docs/behavior-matrices.md`

### Change Log

- 2026-06-12: Implemented deterministic writer-based command help/usage rendering and Story 3.4 executable evidence.
- 2026-06-12: Senior developer review approved implementation after fixing artifact and evidence documentation drift.
