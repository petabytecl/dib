---
baseline_commit: a920b40
created: "2026-06-12T00:00:00-04:00"
---

# Story 2.8: Prove Flag Parsing Across Matrices And Fuzz Inputs

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a technical reviewer,
I want parser behavior proven across documented matrices and fuzz inputs,
so that flag parsing can be trusted before command routing and config binding depend on it.

## Requirements Trace

- FR5: matrix and fuzz evidence must prove explicit `flags.Set` instances stay reusable and do not rely on package-global mutable state.
- FR6: evidence must cover long flags, shorthand flags, boolean presence, attached values, separate values, and unknown/missing/conversion diagnostics.
- FR7: evidence must cover shorthand groups, final/non-final value behavior, no-option defaults, invalid groups, and grouped diagnostics.
- FR8: evidence must cover repeated accumulation, duplicate rejection, custom parser success/failure, string-list flattening, redaction, and custom-result immutability.
- FR9: evidence must cover interspersed positionals, `--`, passthrough args, help requests, failed parse zero snapshots, and typed deterministic diagnostics.
- FR10: evidence must cover exact long-name matching, configured normalization, normalized collisions, and shorthand independence from long-name normalization.
- FR20: `docs/behavior-matrices.md` must become a reviewer-usable flag parser evidence map with FR and executable test traceability.
- FR22: parser hardening must use standard Go fuzzing/property tests and clean-room deterministic seed corpus only.
- NFR1, NFR2, NFR3, NFR4, NFR6, NFR7, NFR8, NFR9: keep runtime/tests standard-library-only, explicit-instance based, typed-error inspectable, deterministic, table-testable, compatibility-clear, redaction-safe, and aligned to Go 1.26.

## Acceptance Criteria

1. Given Stories 2.1 through 2.7 define parser behavior, when `docs/behavior-matrices.md` or an equivalent parser matrix artifact is updated, then it covers definitions, normalization, long flags, shorthand flags, shorthand groups, repeated values, custom values, no-option defaults, parse boundaries, help requests, and diagnostics, and each matrix row traces back to relevant FRs and tests.
2. Given parser tests are executable evidence, when verification runs, then table-driven tests cover valid, invalid, ambiguous, duplicate, boundary, and remaining-arg cases, and typed diagnostics are asserted with `errors.Is` or `errors.As` where caller inspection is required.
3. Given parser fuzzing must use only standard Go support, when fuzz targets and seed corpus files are added, then they live under the package-specific `testdata/fuzz/` flow, and fuzzing proves parser inputs do not panic, mutate reusable definitions, or produce nondeterministic boundary behavior.
4. Given parser evidence must not weaken dependency claims, when parser examples, fixtures, fuzz seeds, or docs are added, then clean-room provenance is recorded where required, and `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` pass.

## Tasks / Subtasks

- [x] Confirm tracker, artifact, and source state (AC: 1-4)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Story 2.7 `done` and Story 2.8 `ready-for-dev` before implementation starts.
  - [x] Check for Story 2.8 ATDD artifacts under `_bmad-output/test-artifacts/`; none existed at story creation, but use them if generated before implementation.
  - [x] Verify `go.mod` still contains only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or generated `go.sum`.
  - [x] Reuse the existing public parser surface: `flags.Set.Parse`, `flags.Snapshot`, `flags.ValueState`, `flags.ValueOccurrence`, `flags.ParseError`, existing sentinels, and existing `flags/fuzz_test.go`. Do not add a second parser API or result type.

- [x] Convert `docs/behavior-matrices.md` into the full Epic 2 parser evidence map (AC: 1)
  - [x] Add or refactor a dedicated flag parser matrix section covering: definitions, immutable sets, normalization, long flags, shorthand flags, shorthand groups, repeated values, custom values, no-option defaults, parse boundaries, help requests, diagnostics, redaction, and fuzz/property hardening.
  - [x] For each matrix row, include at least: behavior area, relevant FR IDs, expected behavior, executable evidence by exact test/fuzz function name, and current/later status.
  - [x] Trace existing tests rather than inventing paper evidence. Expected evidence includes `TestSet...`, `TestExact...`, `TestParseLong...`, `TestParseShort...`, `TestParseGroup...`, `TestParseRepeatable...`, `TestParseBoundary...`, `TestParseHelp...`, and `FuzzParseBoundary`.
  - [x] Keep the matrix honest: mark command routing, config binding, compatibility tables, migration examples, and release-candidate evidence as later hooks where they are not implemented yet.
  - [x] Do not claim source compatibility with Go `flag`, pflag, Cobra, or Viper. Document behavior-scoped familiarity only.

- [x] Add parser matrix tests only where trace gaps exist (AC: 2)
  - [x] Audit the matrix against the current `flags/` tests and add narrowly scoped tests for any missing valid, invalid, ambiguous, duplicate, boundary, or remaining-arg case.
  - [x] Prefer extending existing focused files over scattering cases: `flags/set_test.go`, `flags/normalize_test.go`, `flags/parse_long_test.go`, `flags/parse_shorthand_test.go`, `flags/parse_group_test.go`, `flags/repeated_test.go`, `flags/parse_boundary_test.go`, and ATDD files where already present.
  - [x] Assert caller-observable behavior through `Snapshot.Lookup`, `RemainingArgs`, `ValueState.Values`, `Explicit`, `Occurrences`, `errors.Is`, `errors.As`, and `*flags.ParseError`; do not rely on error strings as the programmatic contract.
  - [x] Preserve the current Story 2.7 decision that grouped `-h` such as `-vh` is not a help request and returns `ErrUnknownFlag` for the unknown grouped member.

- [x] Extend standard-library fuzz/property coverage for full parser behavior (AC: 3)
  - [x] Keep `flags/fuzz_test.go` standard-library-only; use `testing.F`, `f.Add`, and deterministic testdata corpus files.
  - [x] Either extend `FuzzParseBoundary` only if its name remains accurate, or add a new broad target such as `FuzzParse` for full parser coverage. If a new target is added, place seeds under `flags/testdata/fuzz/FuzzParse/`.
  - [x] Seed coverage must include clean-room inputs for long flags, short flags, shorthand groups, no-option defaults, repeated values, duplicate single-value flags, custom value conversion failures, normalization spellings, positionals, `--`, help requests, and invalid/ambiguous forms.
  - [x] The fuzz target must verify parser invariants, not exact arbitrary-output compatibility: no panic, deterministic success/failure on repeated parses, reusable `Set` definitions are not mutated, successful parses return defensively copied `RemainingArgs`/values/occurrences, failed parses return zero-value snapshots, and typed parse failures remain inspectable.
  - [x] Keep corpus files in Go fuzz corpus format and independently written. Do not copy seeds, examples, or fixtures from Go, pflag, Cobra, Viper, or other parser projects.

- [x] Update clean-room and release evidence docs only as needed (AC: 4)
  - [x] Update `docs/provenance-log.md` if new matrix text, fuzz seeds, examples, or fixtures are influenced by external public documentation; classify such entries as `inspiration-only` unless material is actually copied or adapted and reviewed.
  - [x] Update `docs/release-checklist.md` only if a parser fuzz-evidence slot is needed for future release candidates; do not fill release-candidate results as if this story tags a release.
  - [x] Update `docs/diagnostics-and-errors.md` only if new tests reveal a missing public diagnostic contract. Do not rename existing sentinels or change wrapping behavior without a concrete test-driven need.

- [x] Preserve Story 2.8 scope boundaries (AC: 1-4)
  - [x] Do not implement command routing, config binding, config precedence, help rendering, usage rendering, compatibility adapters, migration examples, shell completion, `/cmd` scaffolding, or root facade APIs.
  - [x] Do not add external runtime, test, fuzzing, assertion, or tooling dependencies.
  - [x] Do not rewrite parser internals unless a missing matrix/fuzz case exposes an actual bug. If parser behavior changes, add the focused regression test first and keep the change inside `flags/`.
  - [x] Do not create broad shared helpers or `internal/` packages for this story.

- [x] Verify the story implementation (AC: 1-4)
  - [x] Run focused matrix tests, for example `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./flags -run 'TestSet|TestDefinition|TestExact|TestParse|TestSensitive|TestNonSensitive' -count=1`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`.
  - [x] Run every parser fuzz target with a short deterministic local cycle, for example `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -fuzz=FuzzParse -fuzztime=5s ./flags` or the actual target names added/kept.
  - [x] Run `git diff --check`.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [x] Consider `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./... -count=1` as extra evidence because the story is about reusable definitions and deterministic snapshots.
  - [x] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD/addendum/rubric/reconcile docs under `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/`.
- Loaded previous story: `_bmad-output/implementation-artifacts/2-7-preserve-parse-boundaries-and-remaining-args.md`.
- Loaded current source/docs: `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, `docs/provenance-log.md`, `docs/release-checklist.md`, `flags/fuzz_test.go`, `flags/testdata/fuzz/FuzzParseBoundary/*`, `flags/parse.go`, `flags/errors.go`, `flags/snapshot.go`, and the `flags/*_test.go` inventory.
- No `project-context.md`, UX artifact, Story 2.8 ATDD artifact, `examples/` directory, or `flags/testdata/fuzz/FuzzParse/` directory was discovered at story creation.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `a920b40` (`feat(story-2.7): preserve parse boundaries`).
- Story 2.7 is `done`; Story 2.8 is being created from `backlog`.
- The worktree contains unrelated BMAD configuration, skill, `.codex/`, and story-automator changes. Do not revert or normalize them while implementing this story.
- Root `go.mod` currently contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

- No `go.sum` file exists at story creation.

### Architecture Guardrails

- `flags/` owns explicit flag sets, long/shorthand parsing, shorthand groups, repeated/custom values, no-option defaults, normalization, parse diagnostics, usage metadata, and flag-specific typed errors. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- `flags/` must remain fully usable without `command/` or `config/`; callers compose surfaces explicitly. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- The module root must not provide a broad public facade, package-global helpers, or default singleton API. [Source: `_bmad-output/planning-artifacts/architecture.md#API-Boundaries`]
- Definitions, flag sets, and snapshots are reusable values; derived values return new values and must avoid exported mutable internals or shallow-copy aliasing. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Architecture`]
- Public APIs use explicit inputs and returned values/errors and must not depend on package globals, hidden process IO, implicit environment reads, or root singletons. [Source: `_bmad-output/planning-artifacts/architecture.md#API-And-Communication-Patterns`]
- Public errors must support `errors.Is` / `errors.As` compatible inspection; error strings are diagnostics, not programmatic contracts. [Source: `_bmad-output/planning-artifacts/architecture.md#API-And-Communication-Patterns`]
- Tests live beside the package under test; fixtures stay under package-specific `testdata/`; fuzz corpus data belongs under the relevant `testdata/fuzz/` directory. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Organization`]
- Standard Go fuzzing support is the only approved fuzzing mechanism; seed corpus files must be clean-room and deterministic. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Boundaries`]
- Dependency enforcement is owned by `tools/depgate/`; do not create alternate dependency gates. [Source: `_bmad-output/planning-artifacts/architecture.md#Development-Workflow-Integration`]

### Current Code Context

- `flags.Set.Parse(args []string) (Snapshot, error)` is the single public parse entrypoint. It returns `Snapshot{}` on any parse error and never reads `os.Args`. [Source: `flags/parse.go`]
- `parseLong`, `parseShort`, `parseShortGroup`, `parseResolvedFlagWithOccurrence`, `applyNoOptionParsedValue`, and `applyParsedValue` already implement the behavior created in Stories 2.3-2.7. Treat them as the implementation under proof, not as surfaces to redesign. [Source: `flags/parse.go`]
- Public parse sentinels currently include `ErrUnknownFlag`, `ErrMissingValue`, `ErrDuplicateValue`, `ErrConversion`, `ErrInvalidGroup`, and `ErrHelpRequest`. [Source: `flags/errors.go`]
- `ParseError` exposes `Category`, `Token`, `Name`, `NormalizedName`, and `Definition`, and unwraps category plus optional cause for Go error inspection. [Source: `flags/errors.go`]
- `Snapshot.RemainingArgs`, `ValueState.Values`, and `ValueState.Occurrences` return defensive copies. Matrix/fuzz assertions should preserve that contract. [Source: `flags/snapshot.go`]
- Existing fuzz coverage is `FuzzParseBoundary` with string-encoded arg lists and seed corpus under `flags/testdata/fuzz/FuzzParseBoundary/`. It checks no panic, non-nil remaining args after `--`, all-positional success, and repeated-parse outcome stability. [Source: `flags/fuzz_test.go`]
- Existing seed corpus files are `seed1` through `seed7` under `flags/testdata/fuzz/FuzzParseBoundary/`, including empty input, `--`, passthrough flag-like args, `-h`, interspersed positionals, missing value, and attached value. [Source: `flags/testdata/fuzz/FuzzParseBoundary/*`]
- `docs/behavior-matrices.md` currently has shared rows for immutable definitions, no mutable aliases, per-run snapshots, explicit instances, normalization, long parsing, short parsing, grouped shorthand parsing, repeated/custom values, parse boundaries, public error inspection, diagnostics, source labels, and redaction corpus. Story 2.8 should add traceability depth without making false release or compatibility claims. [Source: `docs/behavior-matrices.md`]

### Previous Story Intelligence

- Story 2.7 added `ErrHelpRequest`, help detection for unregistered `--help`/`-h`, `flags/parse_boundary_test.go`, `FuzzParseBoundary`, boundary fuzz seeds, and docs updates.
- Story 2.7 intentionally deferred the full parser fuzz matrix proof to Story 2.8.
- Story 2.7 review fixed a subtle scope boundary: `-h` inside a shorthand group, such as `-vh`, must not trigger help-request behavior. Preserve this behavior in the matrix and fuzz invariants.
- Story 2.6 established repeated/custom value tests and documented custom-parser wrapping/redaction. Story 2.8 should trace those tests rather than reimplement custom parser behavior.
- Story 2.5 established grouped shorthand behavior and targeted property coverage. Story 2.8 should include grouped cases in the full matrix/fuzz proof.
- Current verification baseline from previous completed stories: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `git diff --check`, short parser fuzz runs, and optional `go test -race ./...`.

### Git Intelligence

- Recent commits: `a920b40 feat(story-2.7): preserve parse boundaries`, `760ea1c feat(story-2.6): accumulate repeated custom values`, `1b9952e feat(story-2.5): parse short flag groups`, `a84276f feat(story-2.4): parse short boolean flags`, `27e455a feat(story-2.3): reject invalid long flags`.
- Recent implementation pattern: add focused tests beside `flags/`, keep parser changes inside `flags/parse.go` when needed, document cross-surface contracts in `docs/behavior-matrices.md` and `docs/diagnostics-and-errors.md`, and keep dependency proof through `tools/depgate`.
- Story 2.8 likely touches `docs/behavior-matrices.md`, `flags/fuzz_test.go`, `flags/testdata/fuzz/...`, possibly focused `flags/*_test.go`, and possibly `docs/provenance-log.md` or `docs/release-checklist.md`.

### Technical Research Notes

- Go fuzz tests must be named `FuzzXxx`, accept only `*testing.F`, live in `*_test.go`, and contain exactly one `f.Fuzz` target. Seed corpus argument types must match `f.Add` and target argument types. [Source: https://go.dev/doc/security/fuzz/]
- Go fuzz targets should be fast and deterministic; target state must not persist past each call and must not depend on global state because fuzz invocations can run in parallel and nondeterministic order. [Source: https://go.dev/doc/security/fuzz/]
- Regular `go test` runs seed corpus entries as unit tests; `go test -fuzz=FuzzName` enables fuzzing. Seed corpus files belong under `testdata/fuzz/FuzzName/` and use Go fuzz corpus format. [Source: https://go.dev/doc/security/fuzz/]
- The `testing` package documents `testing.F` and native fuzzing support; no third-party fuzzing library is needed. [Source: https://pkg.go.dev/testing#hdr-Fuzzing]

### Testing Standards

- Treat package tests as executable truth; docs must point to tests that actually exist after implementation.
- Use table-driven tests for matrix gaps and fuzz/property tests for broad parser hardening.
- Keep fuzz input decoding simple, deterministic, and bounded. Avoid fixtures that depend on live env, current working directory, wall clock, stdin/stdout, host files, or global process state.
- Assert typed diagnostics with `errors.Is` and `errors.As` whenever caller inspection is part of the contract.
- Use the architecture-defined sensitive corpus when testing redaction: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.
- Regular package verification must pass under `go test ./...`; fuzz-specific verification should run with short `-fuzztime` during implementation and record exact results.

### Security And Quality Checks

- Do not hardcode secrets, credentials, tokens, or private URLs.
- Do not copy source, tests, comments, fixtures, examples, internal names, file organization, or README structure from Go `flag`, pflag, Cobra, Viper, or other parser projects.
- Do not add external imports to runtime, tests, examples, fuzzing, or tooling.
- Do not add `os.Exit`, `os.Args`, hidden stdout/stderr use, `flag.CommandLine`, package-global mutable registries, or default singleton APIs.
- New docs must not claim unsupported parser behavior, compatibility parity, command/config integration, release readiness, or future API stability.

### Project Structure Notes

- Align with the existing architecture-owned structure: parser implementation and tests stay in `flags/`; parser fuzz corpus stays under `flags/testdata/fuzz/`; cross-surface matrix docs stay under `docs/behavior-matrices.md`.
- If a broad parser fuzz target is added, prefer `flags/testdata/fuzz/FuzzParse/` because the architecture names `flags/testdata/fuzz/FuzzParse/basic.txt` as the intended parser hardening location. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Organization`]
- Do not create `examples/`, `internal/`, `/cmd`, or compatibility/migration docs for this story unless a later architecture/story explicitly owns that surface.
- No structure conflict detected: the current repo already contains `flags/fuzz_test.go` and `flags/testdata/fuzz/FuzzParseBoundary/`; Story 2.8 can add `FuzzParse` alongside it or carefully broaden the existing target.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md#Story-28-Prove-Flag-Parsing-Across-Matrices-And-Fuzz-Inputs`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-20-Provide-Behavior-Test-Matrices`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-22-Support-Fuzz-Or-Property-Style-Parser-Hardening`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#Resolved-Parser-Behavior-Matrix`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Project-Structure-And-Boundaries`]
- [Source: `_bmad-output/implementation-artifacts/2-7-preserve-parse-boundaries-and-remaining-args.md#Previous-Story-Intelligence`]
- [Source: `docs/behavior-matrices.md`]
- [Source: `flags/fuzz_test.go`]
- [Source: `flags/testdata/fuzz/FuzzParseBoundary/`]
- [Source: `https://go.dev/doc/security/fuzz/`]
- [Source: `https://pkg.go.dev/testing#hdr-Fuzzing`]

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

- Discovered `stopBeforeLong` polarity is inverted from what Dev Notes implied: long flags use `stopBeforeLong=true` (never consumes a following long-flag token as value) while short flags use `stopBeforeLong=false`. Required correcting the no-option default test for short flags and adding a third case for long flags.
- Initial `FuzzParse` invariant checked sensitive-value non-leakage against all error output; fuzzer immediately generated `--dib_fake_secret_value` as a flag *name* (not a value), which correctly appears in unknown-flag error tokens. Invariant narrowed to the story-specified structural properties only; sensitive-value leakage for actual flag values is covered by dedicated focused tests.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Added `TestParseLongNoOptionDefault` to `flags/parse_long_test.go` — covers end-of-args, before `--`, and before another long-flag token (stopBeforeLong=true for long flags).
- Added `TestParseShorthandNoOptionDefault` to `flags/parse_shorthand_test.go` — covers end-of-args and before `--` (stopBeforeLong=false for short flags).
- Added `FuzzParse` broad fuzz target to `flags/fuzz_test.go` with 23 clean-room seeds covering all major parser behavior areas. Verified anchored 5s fuzz run, no failures.
- Created 23 seed corpus files under `flags/testdata/fuzz/FuzzParse/`.
- Updated `docs/behavior-matrices.md` with full "Flag Parser Evidence Map" section: 14 rows with FR IDs, expected behavior, exact test/fuzz function names, and current/later status markers.
- Updated `docs/provenance-log.md` with inspiration-only entries for Go fuzzing documentation and `testing` package docs.
- `docs/diagnostics-and-errors.md` and `docs/release-checklist.md` unchanged — no new diagnostic contracts discovered and no release-candidate slot needed.
- All gates: `go test ./...` ✓, `go vet ./...` ✓, `go run ./tools/depgate` ✓, `git diff --check` ✓, `go test -race ./...` ✓, `FuzzParse -fuzztime=5s` ✓, `FuzzParseBoundary -fuzztime=5s` ✓, `go.mod` unchanged, no `go.sum`.
- QA automation pass added missing fuzz invariants for reusable set definitions, defensive `Values()`/`Occurrences()` copies, and sensitive conversion redaction in `FuzzParse`.
- QA automation pass narrowed `FuzzParseShortGroups` to deterministic typed diagnostics after fuzzing proved sensitive-looking unknown flag names are not sensitive values.
- Created Story 2.8 test automation summary at `_bmad-output/implementation-artifacts/tests/test-summary.md`.
- Senior review pass fixed a missing normalization-spelling seed in `FuzzParse`, added `seed23-normalized-long`, and corrected fuzz verification to use anchored `-fuzz` regexes because unanchored `FuzzParse` matches multiple fuzz targets.

### File List

- `flags/parse_long_test.go`
- `flags/parse_shorthand_test.go`
- `flags/fuzz_test.go`
- `flags/parse_group_test.go`
- `flags/testdata/fuzz/FuzzParse/seed01-empty`
- `flags/testdata/fuzz/FuzzParse/seed02-boolean-long`
- `flags/testdata/fuzz/FuzzParse/seed03-attached-string`
- `flags/testdata/fuzz/FuzzParse/seed04-separate-string`
- `flags/testdata/fuzz/FuzzParse/seed05-attached-int`
- `flags/testdata/fuzz/FuzzParse/seed06-no-option-default-long`
- `flags/testdata/fuzz/FuzzParse/seed07-repeatable-long`
- `flags/testdata/fuzz/FuzzParse/seed08-boolean-short`
- `flags/testdata/fuzz/FuzzParse/seed09-attached-short`
- `flags/testdata/fuzz/FuzzParse/seed10-separate-short`
- `flags/testdata/fuzz/FuzzParse/seed11-no-option-default-short`
- `flags/testdata/fuzz/FuzzParse/seed12-repeatable-short`
- `flags/testdata/fuzz/FuzzParse/seed13-group-boolean-attached`
- `flags/testdata/fuzz/FuzzParse/seed14-group-boolean-separate`
- `flags/testdata/fuzz/FuzzParse/seed15-interspersed-positionals`
- `flags/testdata/fuzz/FuzzParse/seed16-terminator-passthrough`
- `flags/testdata/fuzz/FuzzParse/seed17-help-long`
- `flags/testdata/fuzz/FuzzParse/seed18-help-short`
- `flags/testdata/fuzz/FuzzParse/seed19-unknown-long`
- `flags/testdata/fuzz/FuzzParse/seed20-conversion-failure`
- `flags/testdata/fuzz/FuzzParse/seed21-duplicate-single-value`
- `flags/testdata/fuzz/FuzzParse/seed22-sensitive-value`
- `flags/testdata/fuzz/FuzzParse/seed23-normalized-long`
- `docs/behavior-matrices.md`
- `docs/provenance-log.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/implementation-artifacts/2-8-prove-flag-parsing-across-matrices-and-fuzz-inputs.md`
- `_bmad-output/implementation-artifacts/tests/test-summary.md`

## Change Log

- 2026-06-11: Story 2.8 implementation complete. Added `TestParseLongNoOptionDefault` and `TestParseShorthandNoOptionDefault` gap tests. Added `FuzzParse` broad parser fuzz target with 22 clean-room corpus seeds. Updated `docs/behavior-matrices.md` with full Epic 2 parser evidence map. Updated `docs/provenance-log.md` with Go fuzzing documentation entries.
- 2026-06-12: Senior Developer Review (AI) complete. Fixed `FuzzParse` normalization-spelling seed coverage, added `seed23-normalized-long`, corrected fuzz verification commands to anchored regexes, and marked story done after gates passed.

## Senior Developer Review (AI)

Reviewer: Codex on 2026-06-12

### Findings

- HIGH: `FuzzParse` was claimed to cover normalization spellings, but the broad fuzz set used `flags.NewSet` with no `NameNormalizer` and had no normalized-name seed. Fixed by switching `FuzzParse` to `flags.NewNormalizedSet`, adding `log-level`, adding the `--log_level=debug` seed, and adding `flags/testdata/fuzz/FuzzParse/seed23-normalized-long`.
- MEDIUM: The recorded fuzz command `go test -fuzz=FuzzParse -fuzztime=5s ./flags` is ambiguous in this package because it matches `FuzzParse`, `FuzzParseBoundary`, and `FuzzParseShortGroups`. Fixed the reproducible verification record to use anchored `-fuzz='^FuzzParse$'`, `-fuzz='^FuzzParseBoundary$'`, and `-fuzz='^FuzzParseShortGroups$'`.
- MEDIUM: The broad fuzz target asserted typed parse errors with `t.Errorf` and then continued using the `*flags.ParseError` variable. Fixed by making that assertion fatal so a non-typed error cannot cascade into a nil dereference during fuzzing.

### Verification

- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./flags -run 'TestParseLongNoOptionDefault|TestParseShorthandNoOptionDefault|TestParseShortGroupKeepsShorthandIndependentFromLongNormalization|TestExactAndNormalizedLongNamesParseSafely' -count=1`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./flags -run '^FuzzParse$' -count=1`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`
- PASS: `git diff --check`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./... -count=1`
- PASS: `go.mod` remains only `module github.com/petabytecl/dib` plus `go 1.26`; no `go.sum` exists.

### GitHub Sync

- Attempted to comment on GitHub issue #19 through the GitHub connector twice. The host cancelled both MCP tool calls.
- Attempted fallback `gh issue comment 19 --repo petabytecl/dib`; it failed because the sandbox could not connect to `api.github.com`.
- GitHub issue reflection is therefore blocked by external connector/network availability; this local story artifact records the review sync.
- 2026-06-12: QA automation pass complete. Strengthened `FuzzParse` invariants, fixed grouped-shorthand fuzz scope, updated the sensitive corpus seed, wrote test summary, and re-ran parser fuzz plus verification gates.
