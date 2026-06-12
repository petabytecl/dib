---
baseline_commit: 2dd6a3cfc39acb88cbd784af5ee02387d481744f
created: "2026-06-11T19:46:09-04:00"
---

# Story 2.3: Reject Invalid Long Flags With Inspectable Errors

Status: done

## Story

As a Go CLI developer,
I want long flags to parse familiar forms and reject invalid input with typed diagnostics,
so that scripts and tests can handle parser failures without scraping error text.

## Requirements Trace

- FR6: Long flags parse `--name=value`, `--name value`, and boolean presence/explicit boolean values where the flag definition permits them.
- FR9: Parsing preserves caller-controlled boundaries, remaining positional args, and typed diagnostics for unknown flags, missing values, invalid values, and duplicates.
- FR10: Long-flag parsing reuses caller-configured normalization and reports canonical definition identity separately from raw CLI spelling.
- FR20: Parser behavior is covered by table-driven tests and diagnostics are asserted through typed errors and snapshot state, not only strings.
- NFR1, NFR2, NFR3, NFR4, NFR6, NFR7, NFR8: runtime and tests remain standard-library-only, explicit-instance based, typed-error inspectable, deterministic, table-testable, compatibility-clear, and redaction-safe.

## Acceptance Criteria

1. Given a known long flag accepts a value, when input uses `--name=value` or `--name value`, then the parsed snapshot records the explicit value, source spelling, and canonical flag definition, and remaining positional arguments preserve their relative order.
2. Given a known boolean long flag, when input uses `--name`, `--name=true`, or `--name=false`, then the parsed snapshot records the expected boolean value, and invalid boolean text returns a typed conversion parse error.
3. Given an unknown long flag is provided before `--`, when parsing runs, then parsing returns a typed unknown-flag diagnostic, and the diagnostic includes the flag token and normalized/canonical lookup context where applicable.
4. Given a long flag requires a value, when input omits the value or the next argument cannot be consumed as a value, then parsing returns a typed missing-value diagnostic, and the failed parse does not mutate the reusable Flag set.
5. Given long flag behavior is a public contract, when verification runs, then table-driven tests cover attached values, separate values, booleans, unknown flags, missing values, invalid conversions, duplicate single-value flags, and exact/normalized names, and diagnostics are asserted through typed errors and snapshot state, not only strings.

## Tasks / Subtasks

- [x] Confirm current tracker, artifact, and source state (AC: 1-5)
  - [x] Verify `sprint-status.yaml` marks Story 2.2 `done` and Story 2.3 `ready-for-dev`.
  - [x] Check for Story 2.3 ATDD artifacts under `_bmad-output/test-artifacts/`; none existed at story creation, but use them if generated before implementation starts.
  - [x] Verify root `go.mod` still declares `module github.com/petabytecl/dib` and `go 1.26` with no `require`, `replace`, or `toolchain` directives.
  - [x] Reuse the existing `flags.Set`, `Definition`, `Snapshot`, `ValueState`, `NameNormalizer`, `Definition.Parse`, and typed error foundation instead of replacing them.

- [x] Add the explicit long-flag parse entrypoint (AC: 1-5)
  - [x] Add caller-explicit parse API `func (s Set) Parse(args []string) (Snapshot, error)`.
  - [x] Accept caller-supplied args only; do not read `os.Args`, `flag.CommandLine`, stdin/stdout/stderr, package globals, or process state.
  - [x] Build parsing from `s.DefaultSnapshot()` or equivalent per-run state so every parse produces an independent snapshot.
  - [x] Keep `flags/parser.go` value-parser semantics intact; if CLI token parsing needs its own file, use `flags/parse.go` or `flags/parse_long.go` rather than overloading value parsing concepts.
  - [x] Return a self-contained snapshot on success with defensive-copy accessors for values, remaining args, and source metadata.

- [x] Extend snapshot state for parsed values and remaining args (AC: 1, 2, 5)
  - [x] Keep snapshot lookup keyed by canonical `Definition.Name()` values; raw CLI spelling must not become the state key.
  - [x] Add `Snapshot.RemainingArgs() []string`, returning a defensive copy.
  - [x] Add a small public source/occurrence model so tests can inspect source spelling and canonical definition identity without internals or error strings. Target API: `ValueState.Occurrences() []ValueOccurrence`, with defensive copies and accessors for the raw token/spelling plus the canonical `Definition` or canonical definition name.
  - [x] Preserve existing `ValueState.Default()`, `Values()`, `Explicit()`, and `Arity()` behavior for default snapshots and parsed snapshots.
  - [x] Do not expose mutable internal slices, maps, or caller-owned `args` storage through the snapshot.

- [x] Implement long flag token forms (AC: 1, 2, 4)
  - [x] Parse `--name=value` by treating the substring before `=` as the raw long-name spelling and the substring after `=` as the raw value; `--name=` is an attached empty value and should be passed to the definition parser.
  - [x] Parse `--name value` for definitions with required values when the next token is available and is not the `--` terminator or another long-flag token.
  - [x] Parse boolean long flags as `--name` -> true by default, plus `--name=true` and `--name=false` through `Definition.Parse` so invalid boolean text is reported as conversion failure.
  - [x] Treat `--no-flag` as an ordinary long name only if the caller registered `no-flag`; do not generate automatic negation aliases.
  - [x] Preserve non-flag positional args in relative order in the success snapshot.
  - [x] Implement only the minimal `--` stop behavior needed for this story's "unknown before `--`" contract; Story 2.7 owns the complete terminator/interspersed matrix.

- [x] Reuse normalized lookup and canonical identity correctly (AC: 1, 3, 5)
  - [x] Resolve long names through `Set.Lookup` so exact-name sets stay exact and normalized sets reuse `NameNormalizer`.
  - [x] Do not bypass the raw lookup validation added in Story 2.2; invalid raw names and shorthand-only names must not become long-name aliases through normalization.
  - [x] Record the raw source spelling from the token separately from the canonical definition name.
  - [x] For normalized matches, expose enough context for tests to assert raw spelling, normalized lookup key where available, and canonical definition identity.
  - [x] Keep normalizer callbacks deterministic in tests; do not add caching or mutation that changes lookup behavior across parses.

- [x] Add typed parse diagnostics (AC: 2, 3, 4, 5)
  - [x] Add sentinel categories for parse errors not already represented, including unknown long flag and missing value. Reuse existing `ErrConversion` and `ErrDuplicateValue`.
  - [x] Add a typed parse-context error, such as `*flags.ParseError`, with inspectable accessors for category, token, raw long name, normalized lookup key where available, and canonical definition where applicable.
  - [x] Preserve `errors.Is(err, flags.ErrConversion)` and `errors.As(err, *flags.ValueError)` for conversion failures from `Definition.Parse`; wrapping with parse context is acceptable only if both inspections still work.
  - [x] Return typed unknown-flag errors for unknown long names before the terminator; include the original flag token and lookup context without relying on rendered text.
  - [x] Return typed missing-value errors when a required value is omitted or the next token is `--` or another long-flag token.
  - [x] Return typed duplicate-value errors when a non-repeatable flag appears more than once in one parse run.
  - [x] Keep sensitive raw values out of error strings and debug output. Error context may identify the flag name/token but must not echo sensitive values.

- [x] Add table-driven package tests and ATDD activation (AC: 1-5)
  - [x] If `$bmad-testarch-atdd` has generated Story 2.3 scaffolds, activate one skipped ATDD test at a time, confirm RED with a narrow `go test ./flags -run ... -count=1`, then implement the smallest production change that passes.
  - [x] Add focused package tests, likely `flags/parse_long_test.go`, for attached values, separate values, positionals, boolean presence, explicit booleans, invalid booleans, unknown long flags, missing values, duplicate single-value flags, exact matching, normalized matching, and `--no-*` as an ordinary registered/unregistered long name.
  - [x] Assert parsed values through `Snapshot.Lookup(...).Values()`, explicit state, remaining args accessors, source metadata, and canonical definition identity.
  - [x] Assert diagnostics through `errors.Is`, `errors.As`, and typed accessors; do not make exact error strings the only contract.
  - [x] Add redaction-focused tests if sensitive values are present in parse errors.
  - [x] Keep tests deterministic and standard-library-only.

- [x] Update adoption-facing docs only for new public behavior (AC: 5)
  - [x] Update `docs/behavior-matrices.md` if the new parse entrypoint, remaining-args snapshot, or long-flag behavior needs adoption evidence.
  - [x] Update `docs/diagnostics-and-errors.md` if new parse sentinels or typed parse errors become part of the public contract.
  - [x] Do not add compatibility examples, fuzz seeds, command/config integration docs, or release evidence; later stories own those surfaces.

- [x] Verify the story implementation (AC: 1-5)
  - [x] Run `go test ./...`.
  - [x] Run `go vet ./...`.
  - [x] Run `go run ./tools/depgate`.
  - [x] Run `git diff --check`.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [x] Confirm no package imports outside the Go standard library and local module were added.
  - [x] Consider `go test -race ./...` as extra evidence because snapshots and sets are reusable values.
  - [x] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### ATDD Artifacts

- Checklist: `_bmad-output/test-artifacts/atdd-checklist-2-3-reject-invalid-long-flags-with-inspectable-errors.md`
- Backend package acceptance scaffold:
  - `flags/parse_long_atdd_test.go`
- Temp API/back-end generation summary: `_bmad-output/test-artifacts/tmp/tea-atdd-api-tests-2026-06-11T20-12-58-0400.json`
- Temp E2E generation summary: `_bmad-output/test-artifacts/tmp/tea-atdd-e2e-tests-2026-06-11T20-12-58-0400.json`
- Temp aggregate summary: `_bmad-output/test-artifacts/tmp/tea-atdd-summary-2026-06-11T20-12-58-0400.json`
- Dev workflow handoff: remove one `t.Skip` in `flags/parse_long_atdd_test.go` at a time, confirm RED with the narrow `go test ./flags -run ... -count=1` command, then implement the smallest change to pass.

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`.
- Loaded previous story: `_bmad-output/implementation-artifacts/2-2-match-flag-names-safely-across-styles.md`.
- Loaded current source/docs: `flags/set.go`, `flags/snapshot.go`, `flags/errors.go`, `flags/definition.go`, `flags/parser.go`, `flags/normalize.go`, `flags/kind.go`, `flags/set_test.go`, `flags/value_test.go`, `flags/errors_test.go`, `flags/state_atdd_test.go`, `flags/set_atdd_test.go`, `flags/normalize_test.go`, `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, and `docs/provenance-log.md`.
- No `project-context.md`, `CLAUDE.md`, local `MEMORY.md`, UX artifact, or Story 2.3 ATDD artifact was discovered in the repo at story creation.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `2dd6a3cfc39acb88cbd784af5ee02387d481744f` (`fix: address flag normalization review findings`).
- `main` is aligned with `origin/main` at story creation.
- Story 2.2 is `done`; it implemented exact-by-default lookup, opt-in name normalization, deterministic normalized collision diagnostics, and review fixes around raw lookup validation and shorthand alias prevention.
- Root `go.mod` contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

### Architecture Guardrails

- `flags/` owns explicit flag sets, long/shorthand parsing, repeated/custom values, no-option defaults, normalization, parse diagnostics, usage metadata, and flag-specific typed errors. [Source: `_bmad-output/planning-artifacts/architecture.md:562`]
- Shared flag metadata and parsing semantics live in `flags/`, and `flags/` must remain fully usable without `command/` or `config/`. [Source: `_bmad-output/planning-artifacts/architecture.md:567`, `_bmad-output/planning-artifacts/architecture.md:570`]
- The module root must not provide a broad public facade, package-global helpers, or default singleton API. [Source: `_bmad-output/planning-artifacts/architecture.md:564`]
- Definitions, flag sets, and snapshots are reusable values; derived values return new values and must avoid exported mutable internals or shallow-copy aliasing. [Source: `_bmad-output/planning-artifacts/architecture.md:205`, `_bmad-output/planning-artifacts/architecture.md:207`]
- Command routing, flag parsing, and config resolution return self-contained snapshots containing selected command, set flags, remaining args, provenance, diagnostics, and typed errors. [Source: `_bmad-output/planning-artifacts/architecture.md:209`, `_bmad-output/planning-artifacts/architecture.md:211`]
- Public APIs use explicit inputs and returned values/errors and must not depend on package globals, hidden process IO, implicit environment reads, or root singletons. [Source: `_bmad-output/planning-artifacts/architecture.md:229`]
- Public errors must support `errors.Is` / `errors.As` compatible inspection; error strings are diagnostics, not programmatic contracts. [Source: `_bmad-output/planning-artifacts/architecture.md:231`]
- CLI args, custom parsers, readers, env lookup functions, and config inputs are untrusted boundary data. [Source: `_bmad-output/planning-artifacts/architecture.md:214`]
- Raw sensitive values must never appear in errors, `String` output, debug strings, rendered diagnostics, source reports, or example output. [Source: `_bmad-output/planning-artifacts/architecture.md:219`]
- Package tests are the executable contract and must live beside package code. [Source: `_bmad-output/planning-artifacts/architecture.md:642`, `_bmad-output/planning-artifacts/architecture.md:654`]

### Requirements Notes

- Epic 2 covers inspectable flag parsing without package-global state, including explicit flag sets, long/shorthand parsing, repeated/custom values, terminators, and typed parse errors. [Source: `_bmad-output/planning-artifacts/epics.md:177`, `_bmad-output/planning-artifacts/epics.md:181`]
- Story 2.3 requires `--name=value`, `--name value`, boolean long flags, unknown long-flag diagnostics, missing-value diagnostics, duplicate single-value coverage, exact/normalized names, and snapshot state assertions. [Source: `_bmad-output/planning-artifacts/epics.md:443`, `_bmad-output/planning-artifacts/epics.md:475`]
- FR6 requires long flags in attached and separate-value forms, boolean long flags in present and explicit false forms, and one-character shorthand later. This story owns long flags only. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:138`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:143`]
- FR9 requires `--` boundary handling, interspersed positional args, preserved positional order, typed parse errors, and no sensitive value leaks. This story implements only the long-flag subset plus minimal boundary behavior needed by its acceptance criteria. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:166`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:173`]
- FR10 requires caller-configured normalization, definition-time normalization collision detection, and exact matching when no normalizer is configured. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:176`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:183`]
- The parser matrix records long-flag attached/separate values, boolean presence/explicit values, `--no-*` as an ordinary long name, repeated flag duplicate errors, positionals, terminator behavior, and help-request boundaries. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:410`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:421`]
- Public parse, config source, command lookup, and conversion failures must expose stable typed or sentinel boundary errors. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:442`]

### Current Code Context

- `flags.Set` currently stores immutable definitions plus `byExactName`, normalized `byName`, shorthand `byShort`, and the configured `NameNormalizer`. [Source: `flags/set.go:3`, `flags/set.go:10`]
- `NewSet` and `NewNormalizedSet` share `newSet`, validate definitions, build deterministic indexes, reject exact duplicates, reject shorthand duplicates, and reject invalid/colliding normalized keys. [Source: `flags/set.go:12`, `flags/set.go:62`]
- `Set.Lookup` validates raw long-name lookup input before normalization and blocks shorthand-only names from resolving as long-name aliases. Story 2.3 parsing must use this path, not a direct map lookup. [Source: `flags/set.go:74`, `flags/set.go:90`]
- `Set.With` and `Set.WithNormalizer` already derive new sets without mutating the original. [Source: `flags/set.go:92`, `flags/set.go:102`]
- `DefaultSnapshot` currently keys values by canonical definition names and returns default `ValueState` entries with defensive copies for mutable defaults. [Source: `flags/set.go:104`, `flags/set.go:110`]
- `Snapshot.Lookup` returns cloned `ValueState`; new remaining/source metadata should preserve this defensive-copy behavior. [Source: `flags/snapshot.go:8`, `flags/snapshot.go:15`]
- `ValueState` currently records default value, effective values, explicit state, and arity. Story 2.3 should extend this rather than introduce a parallel parse result that cannot compose with future config binding. [Source: `flags/snapshot.go:17`, `flags/snapshot.go:78`]
- `Definition.Parse` converts one raw value through the definition parser, validates the returned kind, returns defensive copies, and redacts sensitive parser causes by returning `ValueError` without the raw cause. [Source: `flags/definition.go:229`, `flags/definition.go:246`]
- `ErrDuplicateValue` and `ErrConversion` already exist; parse diagnostics should reuse them instead of creating duplicates. [Source: `flags/errors.go:8`, `flags/errors.go:17`]
- `ValueError` already supports `errors.Is(err, ErrConversion)` and exposes `Name()` and `Kind()` through `errors.As`. Preserve this behavior for parse-time conversion failures. [Source: `flags/errors.go:109`, `flags/errors.go:151`]
- `NameNormalizer` is documented as deterministic and side-effect free; parsing tests should use pure normalizers. [Source: `flags/normalize.go:3`, `flags/normalize.go:11`]
- `ArityRequired` and `ArityOptional` already model required vs optional CLI values. [Source: `flags/kind.go:43`, `flags/kind.go:50`]

### Scope Boundaries

Likely implementation targets:

```text
flags/parse.go
flags/snapshot.go
flags/errors.go
flags/set.go
flags/parse_long_test.go
flags/*_atdd_test.go if Story 2.3 ATDD is generated
docs/behavior-matrices.md
docs/diagnostics-and-errors.md
```

Use existing files when the change is cohesive; create a new small file when it keeps `flags/` focused. Do not introduce broad shared `internal/` packages for one-package parsing helpers.

Out of scope for this story:

- Short flag parsing and boolean shorthand presence from Story 2.4.
- Short flag grouping and optional shorthand values from Story 2.5.
- Full repeated/custom CLI accumulation behavior from Story 2.6 beyond detecting duplicate non-repeatable values and preserving existing custom conversion behavior.
- The complete `--` terminator and interspersed-args matrix from Story 2.7, except minimal behavior needed by Story 2.3 acceptance criteria.
- Parser fuzzing and full behavior-matrix proof from Story 2.8.
- Command routing, config binding, config precedence, examples, compatibility adapters, root facade APIs, or release evidence.
- Any dependency on `flag.CommandLine`, pflag, Cobra, Viper, or other third-party modules.

### Technical Research Notes

- GitHub repository search for `golang cli flag parser long flags` returned no candidate to adopt directly in this repo.
- GitHub code search for `ErrUnknownFlag ParseFlagValue language:go` returned no candidate to adopt directly.
- GitHub code search for `func parseLong flag language:go` returned broad unrelated/prior-art results only; no source, tests, examples, fixtures, comments, internal names, or file organization should be copied.
- Exa research surfaced Go's `flag` package docs/source as relevant prior art for explicit `FlagSet.Parse(arguments []string)` and independent flag sets; Dib must not copy source or inherit package-global/error-handling behavior. Existing `docs/provenance-log.md` already classifies Go `flag`, pflag, Cobra, and Viper references as inspiration-only.
- No new runtime, test, or tooling dependency is justified for this story.

### Previous Story Intelligence

- Story 2.2 added `flags.NameNormalizer`, `flags.NewNormalizedSet`, and `Set.WithNormalizer`, preserving exact matching as the default.
- Story 2.2 added `ErrDuplicateNormalizedName`, `DefinitionError.CollidingName()`, and `DefinitionError.NormalizedName()` for typed normalization diagnostics.
- Story 2.2 review fixed lookup bypass risks: raw long-name lookup input is validated before normalization, and registered shorthand spellings cannot become hidden long-name aliases. Story 2.3 parsing must not regress either fix.
- Story 2.2 review also required deterministic normalizer tests and explicit normalizer documentation. Keep Story 2.3 parse tests table-driven and deterministic; do not iterate maps when expected order matters.
- Story 2.1 established reusable sets, exact lookup, deterministic definitions, immutable derivation, default snapshots, built-in value kinds, custom parser support, and inspectable definition/value errors.
- Story 2.1 review fixed sensitive parser error redaction, custom default/result aliasing, typed nil parser validation, custom kind/default/parser mismatch validation, and nil option validation. Parse work must preserve those behaviors.
- Current verification baseline from previous completed stories: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `git diff --check`, and optional `go test -race ./...`.

### Testing Standards

- Follow red-green-refactor. If Story 2.3 ATDD scaffolds exist, activate one skipped test at a time and confirm RED before production changes.
- Package tests are the executable contract; keep tests beside `flags/` code.
- Tests must use only the Go standard library and local module imports.
- Assert through observable public APIs: parsed snapshot state, remaining args, source metadata, canonical definitions, and typed error inspection.
- Avoid assertions that depend only on exact error strings. String checks are acceptable only for deliberate human-facing diagnostic wording or redaction.
- Use the architecture-defined fake sensitive corpus when testing redaction:
  - `dib_fake_secret_value`
  - `dib_fake_password_value`
  - `dib_fake_token_value`

### Security And Quality Checks

- Do not hardcode secrets, credentials, tokens, or private URLs.
- Treat all CLI tokens and custom parser results as untrusted boundary data.
- Do not leak sensitive raw values in errors, debug strings, rendered diagnostics, source reports, examples, or tests.
- Do not add package-global mutable registries, `os.Args`, hidden stdout/stderr use, `flag.CommandLine`, or `os.Exit`.
- Keep public APIs explicit and immutable from the caller's perspective.
- Keep functions small and files focused; prefer `flags/parse.go` plus focused helper functions over a large parser block inside `set.go`.
- No runtime or test import may come from outside the Go standard library except local module package imports already used by external-package tests.

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- 2026-06-11T21:26:53-04:00: Loaded `sprint-status.yaml`; confirmed Story 2.2 `done` and Story 2.3 `ready-for-dev`, then moved Story 2.3 to `in-progress`.
- 2026-06-11T21:26:53-04:00: Confirmed Story 2.3 ATDD artifacts existed, including `flags/parse_long_atdd_test.go` and `_bmad-output/test-artifacts/atdd-checklist-2-3-reject-invalid-long-flags-with-inspectable-errors.md`.
- 2026-06-11T21:26:53-04:00: Confirmed `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`.
- RED: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./flags -run TestATDDLongFlagValuesPreserveSourceAndRemainingArgs -count=1` failed because `flags.Set` had no `Parse` method.
- GREEN: Activated each Story 2.3 ATDD test in `flags/parse_long_atdd_test.go` and verified each narrow `go test ./flags -run ... -count=1` command passed after implementation.
- Validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./flags -count=1` passed.
- Validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...` passed.
- Validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...` passed.
- Validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate` passed.
- Validation: `git diff --check` passed.
- Extra validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./...` passed.
- Validation: confirmed no `go.sum` exists; `go.mod` still has no `require`, `replace`, or `toolchain` directives.
- Validation: `go list` imports showed only standard-library imports and local module packages.
- Review fix: moved duplicate single-value detection ahead of second-value parsing so duplicate occurrences cannot be misreported as conversion failures or invoke caller parsers unnecessarily.
- Review fix: preserved raw long-name spelling in duplicate parse diagnostics while still exposing normalized lookup key and canonical definition identity.
- Review validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./flags -run 'TestParseLongDuplicate' -count=1` passed.
- Review validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./flags -count=1` passed.
- Review validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...` passed.
- Review validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...` passed.
- Review validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate` passed.
- Review validation: `git diff --check` passed.
- Review validation: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./...` passed.
- GitHub issue sync: no issue or PR number was present in the story, branch, commits, or local artifacts; `gh issue list --repo petabytecl/dib --search '"Story 2.3" OR "Reject Invalid Long Flags"' --limit 20` failed because `api.github.com` was unreachable from this workspace, so no GitHub issue comment could be posted.

### Completion Notes List

- Added explicit `flags.Set.Parse(args []string) (Snapshot, error)` for caller-supplied long-flag parsing without process globals or IO.
- Added snapshot remaining-argument and value-occurrence metadata with defensive-copy accessors.
- Added inspectable parse diagnostics through `flags.ErrUnknownFlag`, `flags.ErrMissingValue`, and `*flags.ParseError`, while preserving `ErrConversion`, `ErrDuplicateValue`, and `*flags.ValueError` inspection.
- Implemented long attached/separate values, boolean presence and explicit boolean values, exact and normalized lookup, ordinary `--no-*` names, duplicate single-value detection, minimal `--` stop handling, and redaction-safe parse error text.
- Activated Story 2.3 ATDD consumer contracts and added focused package tests for edge cases and defensive accessors.
- Updated adoption-facing behavior and diagnostics docs for the new public parse behavior.
- Senior review fixed duplicate error precedence and raw/normalized duplicate diagnostic context.

### File List

- `_bmad-output/implementation-artifacts/2-3-reject-invalid-long-flags-with-inspectable-errors.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `docs/behavior-matrices.md`
- `docs/diagnostics-and-errors.md`
- `flags/errors.go`
- `flags/parse.go`
- `flags/parse_long_atdd_test.go`
- `flags/parse_long_test.go`
- `flags/snapshot.go`

### Change Log

- 2026-06-11: Story created and marked ready for development.
- 2026-06-11: Implemented inspectable long-flag parsing, activated ATDD coverage, updated docs, and marked story ready for review.
- 2026-06-11: Senior developer review fixed duplicate parse diagnostics, reran validation gates, and marked story done.

## Senior Developer Review (AI)

Reviewer: GPT-5 Codex
Date: 2026-06-11T21:43:42-04:00
Outcome: Approved after automatic fixes

### Findings Fixed

- HIGH: Duplicate non-repeatable long flags were detected after parsing the second value, so `--workers=2 --workers=not-an-int` returned `ErrConversion` instead of the required `ErrDuplicateValue`. Fixed by checking duplicate state immediately after definition lookup and before value parsing.
- MEDIUM: Duplicate diagnostics for normalized spellings reported the canonical name as `ParseError.Name()` instead of the raw source long name. Fixed duplicate error construction to preserve raw name, normalized lookup key, and canonical definition separately.

### Review Notes

- Acceptance criteria 1-5 were cross-checked against implementation and tests.
- Story File List covers the story source and documentation files. `flags/parse.go` and `flags/parse_long_test.go` are currently untracked in git status, so `git diff --name-only` omits them even though the Go toolchain compiled and tested them.
- GitHub issue reflection was attempted, but no issue or PR reference was discoverable locally and the GitHub CLI could not reach `api.github.com` from this workspace.

### Validation Checklist

- [x] Story file loaded from `_bmad-output/implementation-artifacts/2-3-reject-invalid-long-flags-with-inspectable-errors.md`
- [x] Story Status verified as reviewable (`review`)
- [x] Epic and Story IDs resolved (`2.3`)
- [x] Story Context warning recorded: no separate story context file discovered
- [x] Epic Tech Spec located via `_bmad-output/planning-artifacts/epics.md`
- [x] Architecture/standards docs loaded from `_bmad-output/planning-artifacts/architecture.md`
- [x] Tech stack detected: Go 1.26 module, standard library runtime
- [x] MCP doc search/web fallback not required for local code review; no new external technical claims introduced
- [x] Acceptance Criteria cross-checked against implementation
- [x] File List reviewed and validated for completeness
- [x] Tests identified and mapped to ACs; duplicate diagnostics coverage added
- [x] Code quality review performed on changed files
- [x] Security review performed on changed files and dependencies
- [x] Outcome decided: approve after fixes
- [x] Review notes appended under "Senior Developer Review (AI)"
- [x] Change Log updated with review entry
- [x] Status updated to `done`
- [x] Sprint status synced
- [x] Story saved successfully
