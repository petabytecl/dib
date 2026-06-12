---
baseline_commit: 760ea1c
created: "2026-06-11T00:00:00-04:00"
---

# Story 2.7: Preserve Parse Boundaries And Remaining Args

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Go CLI developer,
I want parser boundaries and remaining arguments preserved exactly,
so that my application can safely compose flag parsing with positional commands, passthrough arguments, and tests.

## Requirements Trace

- FR9: Parser must treat `--` as a hard boundary that stops flag processing; positionals before `--` accumulate in remaining-args; help requests return typed diagnostics that callers control; all parse failures remain typed, deterministic, and inspectable.
- FR20: Boundary behavior is covered by deterministic table-driven package tests and, for high-ambiguity boundary tokens, by targeted fuzz/property tests with a clean-room seed corpus.
- NFR1, NFR2, NFR3, NFR4, NFR6, NFR7: runtime and tests remain standard-library-only, explicit-instance based, typed-error inspectable, deterministic, table-testable, compatibility-clear, and no `os.Exit` in any runtime path.

## Acceptance Criteria

1. Given positional arguments appear before `--`, when flags are interspersed with positionals, then flags before `--` remain parseable, and positional arguments keep relative order in the remaining-args result.
2. Given input contains the `--` terminator, when parsing reaches the terminator, then flag parsing stops, and every subsequent argument remains untouched even if it looks like a flag.
3. Given a help request is encountered, when parsing identifies it, then Dib returns a typed help-request result or error for caller-controlled rendering and exit policy, and no runtime path calls `os.Exit`.
4. Given parse failures happen before `--`, when unknown flags, missing values, invalid values, or invalid groups are encountered, then the returned diagnostic is typed and deterministic, and the remaining-args behavior is covered by tests for both successful and failed parses.
5. Given parse boundaries feed command routing and config binding, when verification runs, then table-driven tests cover interspersed positionals, `--`, passthrough args, help requests, failed parses, and deterministic snapshot state, and targeted fuzz/property tests cover boundary tokens and remaining-arg preservation.

## Tasks / Subtasks

- [x] Confirm current tracker, artifact, and source state (AC: 1-5)
  - [x] Verify `sprint-status.yaml` marks Story 2.6 `done` and Story 2.7 `ready-for-dev` before moving implementation to `in-progress`.
  - [x] Check for Story 2.7 ATDD artifacts under `_bmad-output/test-artifacts/`; none existed at story creation, but use them if generated before implementation starts.
  - [x] Verify root `go.mod` still declares only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or generated `go.sum`.
  - [x] Reuse existing `flags.Set.Parse`, `flags.Snapshot.RemainingArgs`, `flags.ParseError`, `flags.ErrUnknownFlag`, `flags.ErrMissingValue`, `flags.ErrConversion`, `flags.ErrInvalidGroup`, `parseLong`, `parseShort`, `parseShortGroup`, and `parseResolvedFlagWithOccurrence` foundations. Do not introduce a second parser entrypoint or result type.

- [x] Add `ErrHelpRequest` sentinel and typed help-request detection (AC: 3)
  - [x] Add `ErrHelpRequest = errors.New("flag help request")` to `flags/errors.go` beside the other sentinel errors.
  - [x] In `flags/parse.go`, detect `--help` in `parseLong` before the unknown-flag path: if the resolved name is `help` and no `help` definition is registered in the set, return `newParseError(ErrHelpRequest, token, name, ...)` instead of `ErrUnknownFlag`.
  - [x] In `flags/parse.go`, detect `-h` in `parseShort` before the unknown-flag path: if the shorthand is `h` and no `h` shorthand is registered in the set, return `newParseError(ErrHelpRequest, "--help", "help", ...)` instead of `ErrUnknownFlag`.
  - [x] Do NOT add `os.Exit`, stdout writes, or rendering in any detection path; the error must be returned raw to the caller.
  - [x] Do NOT treat `--help` or `-h` differently when they ARE registered by the caller; in that case they must parse through the normal flag path without triggering `ErrHelpRequest`.
  - [x] Confirm `errors.Is(err, flags.ErrHelpRequest)` works and `errors.As(err, &(*flags.ParseError)(nil))` exposes the token, name, and normalized name.

- [x] Build the full parse boundary test matrix (AC: 1-2-4)
  - [x] Create `flags/parse_boundary_test.go` as the primary test file for this story; prefer this over scattering boundary cases into existing files.
  - [x] **Interspersed positionals** — cover: positional-only args, flags only, interleaved positionals and flags, multiple positionals between flags; assert `RemainingArgs()` preserves relative order and contains no flag tokens.
  - [x] **`--` terminator** — cover: `--` alone, `--` followed by flag-like tokens (`--verbose`, `-v`), `--` followed by ordinary strings, `--` as the first arg, `--` after parsed flags; assert everything after `--` is in `RemainingArgs()` untouched.
  - [x] **Passthrough args** — cover: args after `--` that would be unknown flags, missing-value flags, or otherwise invalid; assert they appear verbatim in `RemainingArgs()`.
  - [x] **Failed parses** — cover: unknown long flag, unknown shorthand, missing value, invalid value, invalid group, duplicate single-value flag; for each assert the error is typed, and explicitly verify `errors.Is(err, <category>)` and `errors.As(err, &(*flags.ParseError)(nil))`; also verify that `Snapshot{}` (zero value) is returned on failure so callers cannot accidentally use partial state.
  - [x] **Deterministic snapshot state** — cover: a parse with flags, positionals, and `--` together; assert snapshot flag values, explicit state, occurrences, and `RemainingArgs()` simultaneously.
  - [x] Do NOT assert only error strings; every boundary assertion must use typed error inspection or snapshot observable state.

- [x] Build help-request test coverage (AC: 3)
  - [x] In `flags/parse_boundary_test.go` or a dedicated `flags/parse_help_test.go`: cover `--help` when unregistered (expect `ErrHelpRequest`), `-h` when unregistered (expect `ErrHelpRequest`), `--help` when a `help` bool flag IS registered (expect normal parse, no error), `-h` when an `h` shorthand IS registered (expect normal parse, no error), `--help` appearing after `--` (must appear in `RemainingArgs()`, not trigger error).
  - [x] Assert `errors.Is(err, flags.ErrHelpRequest)` for help-request cases.
  - [x] Assert `errors.As(err, &pe)` where `pe.Token()`, `pe.Name()`, and `pe.Category()` expose the correct information.

- [x] Add fuzz/property tests for boundary tokens (AC: 5)
  - [x] Create `flags/fuzz_test.go` with `func FuzzParseBoundary(f *testing.F)`. The function must use only the standard `testing` package.
  - [x] Register clean-room deterministic seed inputs with `f.Add(...)` covering: empty string slice (encoded as string), `--` alone, `--` followed by flag-like args, interspersed positionals and flags, `--help` alone, flag followed by missing value, flag with attached value.
  - [x] The fuzz body must call `set.Parse(args)` (where args is derived from the fuzz input) and verify: no panic occurs, `RemainingArgs()` never returns nil when args are non-empty after `--`, parse succeeds when all args are positionals, parse never mutates the reusable Set definition.
  - [x] Create `flags/testdata/fuzz/FuzzParseBoundary/` directory and place at least three clean-room seed corpus files as plain text (one per seed scenario matching `f.Add` entries).
  - [x] Do NOT import any external fuzzing library; use only `testing.F` from the Go standard library.
  - [x] Fuzz targets must be runnable with `go test -fuzz=FuzzParseBoundary -fuzztime=5s ./flags` without errors; regular `go test ./flags` must skip fuzz bodies (using `f.Fuzz` which is only entered during fuzz runs).

- [x] Update adoption-facing docs for boundary behavior (AC: 5)
  - [x] Update `docs/behavior-matrices.md` to add a **Parse boundaries** row documenting interspersed positionals, `--` terminator, passthrough args, help request detection, and failed parse remaining-args behavior; leave later hooks for full compatibility tables and release evidence.
  - [x] Update `docs/diagnostics-and-errors.md` only if `ErrHelpRequest` changes or clarifies the public parse diagnostic contract beyond what is already documented; it must not appear as an "unknown flag" category.
  - [x] Do not add migration examples, command/config docs, compatibility adapters, release evidence, root facade APIs, package-global helpers, or third-party dependencies.

- [x] Preserve Story 2.7 scope boundaries (AC: 1-5)
  - [x] Do not implement command routing, config binding, config precedence, help rendering (content/format), or usage text generation; Story 3.x and Epic 4 own those.
  - [x] Do not expand parser fuzz coverage into the full matrix proof; Story 2.8 owns the full matrix and complete fuzz evidence.
  - [x] Do not add compatibility examples, release evidence, `/cmd` scaffolding, or shared `internal/` packages.
  - [x] Do not copy pflag, Cobra, Viper, Go `flag`, or other source, tests, comments, fixtures, or file organization.

- [x] Verify the story implementation (AC: 1-5)
  - [x] Run a focused boundary test: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./flags -run 'TestParseBoundary|TestParseHelp|TestParseTerminator|TestParseInterspersed' -count=1`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`.
  - [x] Run `git diff --check`.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [x] Run a short fuzz cycle to confirm no panic: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -fuzz=FuzzParseBoundary -fuzztime=5s ./flags`.
  - [x] Consider `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./...` as extra evidence because sets and snapshots are reusable values.
  - [x] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded previous story: `_bmad-output/implementation-artifacts/2-6-accumulate-repeated-and-custom-values.md`.
- Loaded current source: `flags/parse.go`, `flags/errors.go`, `flags/snapshot.go`, `flags/set.go`, `flags/definition.go`, `flags/kind.go`, `flags/parser.go`, `flags/normalize.go`, `flags/parse_long_test.go`, `flags/parse_shorthand_test.go`, `flags/parse_group_test.go`, `flags/repeated_test.go`, `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, and `go.mod`.
- No `project-context.md`, UX artifact, `examples/` directory, or Story 2.7 ATDD artifact was discovered at story creation.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `760ea1c` (`feat(story-2.6): accumulate repeated custom values`).
- Story 2.6 is `done`; Story 2.7 is being created from `backlog`.
- `go.mod` contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

- No `go.sum` file exists at story creation.
- The worktree contains unrelated BMAD configuration, skill, and `.codex/` changes. Do not revert them while implementing this story.

### Architecture Guardrails

- `flags/` owns explicit flag sets, long/shorthand parsing, shorthand groups, repeated/custom values, no-option defaults, normalization, parse diagnostics, usage metadata, and flag-specific typed errors. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- `flags/` must remain fully usable without `command/` or `config/`; callers compose surfaces explicitly. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- The module root must not provide a broad public facade, package-global helpers, or default singleton API. [Source: `_bmad-output/planning-artifacts/architecture.md#API-Boundaries`]
- Definitions, flag sets, and snapshots are reusable values; derived values return new values and must avoid exported mutable internals or shallow-copy aliasing. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Architecture`]
- Public APIs use explicit inputs and returned values/errors and must not depend on package globals, hidden process IO, implicit environment reads, or root singletons — including no `os.Exit` in any runtime path. [Source: `_bmad-output/planning-artifacts/architecture.md#API-And-Communication-Patterns`]
- Public errors must support `errors.Is` / `errors.As` compatible inspection; error strings are diagnostics, not programmatic contracts. [Source: `_bmad-output/planning-artifacts/architecture.md#API-And-Communication-Patterns`]
- Standard Go fuzzing support (`testing.F`) is the only approved fuzzing mechanism; no external fuzzing library may be introduced. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Organization`]
- Fuzz corpus data stays under the relevant package `testdata/fuzz/` directory; seed corpus files must be clean-room and deterministic. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Boundaries`]
- Package tests are the executable contract and must live beside package code. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Organization`]

### Requirements Notes

- FR9 requires callers to control how flag parsing treats non-flag arguments, the `--` terminator, and parse errors; help requests must produce a typed help-request diagnostic for caller-controlled rendering and exit policy. [Source: `_bmad-output/planning-artifacts/epics.md#Story-27-Preserve-Parse-Boundaries-And-Remaining-Args`]
- FR20 requires table-driven tests and targeted fuzz/property tests proving parse-boundary and remaining-arg behavior. [Source: `_bmad-output/planning-artifacts/epics.md#Story-27-Preserve-Parse-Boundaries-And-Remaining-Args`]
- NFR5 explicitly forbids `os.Exit` in runtime paths unless the caller chooses a documented convenience path; the help-request detection path must return an error to the caller, not exit. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#NFR5`]
- The architecture fuzz spec says: `func FuzzXxx(*testing.F)` with seed inputs via `F.Add`; no third-party fuzzing dependency is needed. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Organization`]

### Current Code Context

- `flags.Set.Parse(args []string) (Snapshot, error)` is the single public parse entrypoint. [Source: `flags/parse.go:6`]
- **`--` terminator is already implemented**: `if arg == "--" { snapshot.remaining = append(snapshot.remaining, args[i+1:]...); break }`. Story 2.7 must NOT change this logic — only add tests proving it. [Source: `flags/parse.go:11-14`]
- **Interspersed positionals are already implemented**: non-flag, non-short-flag tokens fall through to `snapshot.remaining = append(snapshot.remaining, arg)`. Story 2.7 must NOT change this logic — only add tests proving it. [Source: `flags/parse.go:26-27`]
- `parseLong` returns `newParseError(ErrUnknownFlag, token, name, normalizedName, Definition{}, false, nil)` for unregistered long flags. The help-request detection for `--help` must happen BEFORE this return by checking if `name == "help"` and the set has no `help` definition. [Source: `flags/parse.go:56-66`]
- `parseShort` returns `newParseError(ErrUnknownFlag, token, shorthand, "", Definition{}, false, nil)` for unregistered shorthands. The help-request detection for `-h` must happen BEFORE this return by checking if `shorthand == "h"` and the set has no `h` shorthand. [Source: `flags/parse.go:42-54`]
- `s.Lookup(name)` uses the long-name index; `s.lookupShorthand(shorthand)` uses the shorthand index. Both return `(Definition, bool)`. Use the `bool` return to distinguish "registered" from "unregistered" before routing to help-request detection. [Source: `flags/set.go:75-90`, `flags/parse.go:201-206`]
- `ParseError` carries `category`, `token`, `name`, `normalizedName`, `definition`, and `hasDefinition`. For `ErrHelpRequest` there is no matching definition, so `hasDefinition = false` and `definition = Definition{}`. [Source: `flags/errors.go:157-175`]
- `newParseError` is the correct constructor for building help-request errors consistently with other parse errors. [Source: `flags/errors.go:167-177`]
- `ErrHelpRequest` must be placed in `flags/errors.go` alongside `ErrUnknownFlag`, `ErrMissingValue`, etc. It is a peer sentinel, not a wrapper. [Source: `flags/errors.go:8-20`]
- `TestParseLongFlagsCoversContractMatrix` already exercises `--` terminator with `[]string{"...", "--", "--unknown", "pos-b"}` and asserts `RemainingArgs() = ["pos-a", "--unknown", "pos-b"]`. Story 2.7 boundary tests must not duplicate this matrix row but CAN add more targeted edge cases. [Source: `flags/parse_long_test.go:12-56`]
- No `testdata/fuzz/` directory exists in `flags/` yet. Story 2.7 creates the first fuzz harness. [Source: filesystem scan at story creation]

### Previous Story Intelligence

- Story 2.6 explicitly deferred full boundary implementation: "Do not implement complete interspersed/terminator boundary behavior beyond preserving the current parser behavior; Story 2.7 owns the full boundary matrix."
- Story 2.6 confirmed no production parser changes were needed for its scope; the boundary machinery (`--` handling, positional accumulation) was already in place.
- Story 2.5 implemented grouped shorthand parsing and explicitly left full parse boundary matrices and fuzz proofs to later stories.
- Story 2.5 established fuzz/property test scaffolding for grouped inputs as targeted property checks. Story 2.7 expands into real `testing.F` fuzz targets for boundary tokens.
- Story 2.4 established shorthand occurrence metadata, duplicate handling across long/short spellings, and redaction-safe conversion errors.
- Story 2.3 established long-flag parsing, parse snapshots, remaining args, source occurrences, typed parse diagnostics, and the `--` terminator's initial minimal coverage.
- Story 2.1 established the `Snapshot.RemainingArgs()` public contract.
- Current verification baseline from previous completed stories: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `git diff --check`, and optional `go test -race ./...`.

### Git Intelligence

- Recent commits: `760ea1c feat(story-2.6): accumulate repeated custom values`, `1b9952e feat(story-2.5): parse short flag groups`, `a84276f feat(story-2.4): parse short boolean flags`, `27e455a feat(story-2.3): reject invalid long flags`, `f8da984 test: add story 2.3 ATDD scaffolds`.
- Story 2.6 touched: `flags/repeated_test.go`, `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, `sprint-status.yaml`, `tests/test-summary.md`. No production parser changes.
- Story 2.5 touched: `flags/parse.go`, `flags/errors.go`, `flags/parse_group_test.go`, `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`. Pattern: keep parser changes inside `flags/parse.go`, tests beside `flags/`, cross-surface docs limited to the behavior just implemented.
- Implementation pattern from recent commits: minimal targeted edits to `flags/parse.go` for new sentinel/detection, new `flags/parse_boundary_test.go` for boundary tests, new `flags/fuzz_test.go` for fuzz harness.

### Technical Research Notes

- Standard Go fuzzing documentation confirms: `func FuzzXxx(*testing.F)` with seed inputs via `f.Add`; `go test -fuzz=FuzzXxx -fuzztime=Xs` runs the fuzzer; regular `go test` executes only seed inputs as unit tests. No third-party library is needed. [Source: https://pkg.go.dev/testing#hdr-Fuzzing]
- Go fuzzing seed corpus: placing plain-text files under `testdata/fuzz/FuzzXxx/` makes `go test` load them as additional seeds without `f.Add`; each file maps to one `f.Add` call equivalent. Files must be deterministic and clean-room. [Source: https://go.dev/doc/security/fuzz/]
- Go `flag.ErrHelp` documents the convention: `-h` or `--help` with an unregistered name returns this error so callers can render usage and exit on their own schedule. Dib's equivalent is `ErrHelpRequest`. [Source: https://pkg.go.dev/flag#ErrHelp] (inspiration-only per clean-room policy — must not copy source, tests, or examples from `flag` package)
- Standard Go `errors.Is` / `errors.As` work with `ParseError.Unwrap() []error` which already returns `[]error{e.category, e.cause}`. `errors.Is(err, ErrHelpRequest)` will traverse through `*ParseError.Unwrap()` correctly. [Source: https://pkg.go.dev/errors]
- Go 1.26 is the current stable release; no version upgrade is needed for this story. [Source: https://go.dev/dl/]

### Testing Standards

- Follow red-green-refactor. If Story 2.7 ATDD scaffolds exist by implementation time, activate one skipped test at a time and confirm RED before production changes.
- Package tests are the executable contract; keep tests beside `flags/` code.
- Tests must use only the Go standard library and local module imports.
- Assert through observable public APIs: `Snapshot.RemainingArgs()`, `Snapshot.Lookup(...)`, `ValueState.Values()`, `ValueState.Explicit()`, `ValueState.Occurrences()`, `errors.Is`, `errors.As`, `*flags.ParseError`.
- Avoid assertions that depend only on exact error strings. String checks are acceptable only for deliberate human-facing diagnostic wording or redaction.
- For fuzz test seeds, create clean-room plain-text files under `flags/testdata/fuzz/FuzzParseBoundary/`; do not copy seed files from Go standard library source trees.
- Use the architecture-defined fake sensitive corpus when testing redaction (not directly needed for this story, but if a boundary test touches redaction, use `dib_fake_secret_value`, `dib_fake_password_value`, `dib_fake_token_value`).

### Security And Quality Checks

- Do not hardcode secrets, credentials, tokens, or private URLs.
- Do not add `os.Exit`, `os.Args`, hidden stdout/stderr use, `flag.CommandLine`, or package-global mutable registries.
- Help-request detection must return an error to the caller — never render usage text or exit the process.
- Keep public APIs explicit and immutable from the caller's perspective.
- No runtime or test import may come from outside the Go standard library except local module package imports already used by existing tests.
- The `ErrHelpRequest` sentinel and its detection must not silently fall through for registered `help`/`h` flags; the only behavioral change is for unregistered names.

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

All verification commands passed on first attempt. No debugging required.

### Completion Notes List

- Added `ErrHelpRequest = errors.New("flag help request")` to `flags/errors.go` as a peer sentinel alongside `ErrUnknownFlag` etc.
- Added `--help` detection in `parseLong` (flags/parse.go): when `name == "help"` and no `help` definition is registered, returns `ErrHelpRequest` before the `ErrUnknownFlag` path.
- Added `-h` detection in `parseShort` (flags/parse.go): when `shorthand == "h"` and no `h` shorthand is registered, returns `ErrHelpRequest` with normalized token `"--help"` and name `"help"`.
- Both detections guard on `!ok` (not registered), so registered `--help`/`-h` parse normally without triggering `ErrHelpRequest`.
- Created `flags/parse_boundary_test.go` with 16 test functions covering: interspersed positionals (6 subtests), `--` terminator (6 subtests), passthrough args (4 subtests), failed parses with typed error + zero-value snapshot assertions (8 subtests), deterministic snapshot, 5 help-request cases, and 1 test pinning the designed scope boundary that `-h` inside a shorthand group (`-vh`) returns `ErrUnknownFlag` not `ErrHelpRequest`.
- Created `flags/fuzz_test.go` with `FuzzParseBoundary` using 7 seed inputs and 3 invariant checks.
- Created `flags/testdata/fuzz/FuzzParseBoundary/` with 3 clean-room seed corpus files (seed1–seed3). The fuzz engine also generated seed4–seed7 during the `-fuzztime=5s` verification run (coverage-extending inputs including `-h`).
- Updated `docs/behavior-matrices.md`: added Parse boundaries row.
- Updated `docs/diagnostics-and-errors.md`: documented `ErrHelpRequest` contract.
- Review fix (AI): added `TestParseHelpShortInGroupDoesNotTriggerHelpRequest` to document that `-h` inside a group triggers `ErrUnknownFlag`, not `ErrHelpRequest` — this pins the intended scope boundary from the story design.
- Verification results:
  - `go test ./flags -run 'TestParseBoundary|TestParseHelp...' -count=1`: PASS (14 tests match pattern; `TestSetReusableAfterHelpRequest` runs in full suite)
  - `go test ./... -count=1`: PASS (all packages)
  - `go vet ./...`: clean
  - `go run ./tools/depgate`: clean
  - `git diff --check`: no whitespace errors
  - `go.mod`: no require/replace/toolchain; no go.sum created
  - `go test -fuzz=FuzzParseBoundary -fuzztime=5s ./flags`: PASS (~8M execs, 0 failures)
  - `go test -race ./... -count=1`: PASS (clean)

### File List

- flags/errors.go
- flags/parse.go
- flags/parse_boundary_test.go
- flags/fuzz_test.go
- flags/testdata/fuzz/FuzzParseBoundary/seed1
- flags/testdata/fuzz/FuzzParseBoundary/seed2
- flags/testdata/fuzz/FuzzParseBoundary/seed3
- flags/testdata/fuzz/FuzzParseBoundary/seed4
- flags/testdata/fuzz/FuzzParseBoundary/seed5
- flags/testdata/fuzz/FuzzParseBoundary/seed6
- flags/testdata/fuzz/FuzzParseBoundary/seed7
- docs/behavior-matrices.md
- docs/diagnostics-and-errors.md
- _bmad-output/implementation-artifacts/sprint-status.yaml
- _bmad-output/implementation-artifacts/2-7-preserve-parse-boundaries-and-remaining-args.md
