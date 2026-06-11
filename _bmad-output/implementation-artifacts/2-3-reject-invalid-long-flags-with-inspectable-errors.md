---
baseline_commit: 2dd6a3cfc39acb88cbd784af5ee02387d481744f
created: "2026-06-11T19:46:09-04:00"
---

# Story 2.3: Reject Invalid Long Flags With Inspectable Errors

Status: ready-for-dev

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

- [ ] Confirm current tracker, artifact, and source state (AC: 1-5)
  - [ ] Verify `sprint-status.yaml` marks Story 2.2 `done` and Story 2.3 `ready-for-dev`.
  - [ ] Check for Story 2.3 ATDD artifacts under `_bmad-output/test-artifacts/`; none existed at story creation, but use them if generated before implementation starts.
  - [ ] Verify root `go.mod` still declares `module github.com/petabytecl/dib` and `go 1.26` with no `require`, `replace`, or `toolchain` directives.
  - [ ] Reuse the existing `flags.Set`, `Definition`, `Snapshot`, `ValueState`, `NameNormalizer`, `Definition.Parse`, and typed error foundation instead of replacing them.

- [ ] Add the explicit long-flag parse entrypoint (AC: 1-5)
  - [ ] Add caller-explicit parse API `func (s Set) Parse(args []string) (Snapshot, error)`.
  - [ ] Accept caller-supplied args only; do not read `os.Args`, `flag.CommandLine`, stdin/stdout/stderr, package globals, or process state.
  - [ ] Build parsing from `s.DefaultSnapshot()` or equivalent per-run state so every parse produces an independent snapshot.
  - [ ] Keep `flags/parser.go` value-parser semantics intact; if CLI token parsing needs its own file, use `flags/parse.go` or `flags/parse_long.go` rather than overloading value parsing concepts.
  - [ ] Return a self-contained snapshot on success with defensive-copy accessors for values, remaining args, and source metadata.

- [ ] Extend snapshot state for parsed values and remaining args (AC: 1, 2, 5)
  - [ ] Keep snapshot lookup keyed by canonical `Definition.Name()` values; raw CLI spelling must not become the state key.
  - [ ] Add `Snapshot.RemainingArgs() []string`, returning a defensive copy.
  - [ ] Add a small public source/occurrence model so tests can inspect source spelling and canonical definition identity without internals or error strings. Target API: `ValueState.Occurrences() []ValueOccurrence`, with defensive copies and accessors for the raw token/spelling plus the canonical `Definition` or canonical definition name.
  - [ ] Preserve existing `ValueState.Default()`, `Values()`, `Explicit()`, and `Arity()` behavior for default snapshots and parsed snapshots.
  - [ ] Do not expose mutable internal slices, maps, or caller-owned `args` storage through the snapshot.

- [ ] Implement long flag token forms (AC: 1, 2, 4)
  - [ ] Parse `--name=value` by treating the substring before `=` as the raw long-name spelling and the substring after `=` as the raw value; `--name=` is an attached empty value and should be passed to the definition parser.
  - [ ] Parse `--name value` for definitions with required values when the next token is available and is not the `--` terminator or another long-flag token.
  - [ ] Parse boolean long flags as `--name` -> true by default, plus `--name=true` and `--name=false` through `Definition.Parse` so invalid boolean text is reported as conversion failure.
  - [ ] Treat `--no-flag` as an ordinary long name only if the caller registered `no-flag`; do not generate automatic negation aliases.
  - [ ] Preserve non-flag positional args in relative order in the success snapshot.
  - [ ] Implement only the minimal `--` stop behavior needed for this story's "unknown before `--`" contract; Story 2.7 owns the complete terminator/interspersed matrix.

- [ ] Reuse normalized lookup and canonical identity correctly (AC: 1, 3, 5)
  - [ ] Resolve long names through `Set.Lookup` so exact-name sets stay exact and normalized sets reuse `NameNormalizer`.
  - [ ] Do not bypass the raw lookup validation added in Story 2.2; invalid raw names and shorthand-only names must not become long-name aliases through normalization.
  - [ ] Record the raw source spelling from the token separately from the canonical definition name.
  - [ ] For normalized matches, expose enough context for tests to assert raw spelling, normalized lookup key where available, and canonical definition identity.
  - [ ] Keep normalizer callbacks deterministic in tests; do not add caching or mutation that changes lookup behavior across parses.

- [ ] Add typed parse diagnostics (AC: 2, 3, 4, 5)
  - [ ] Add sentinel categories for parse errors not already represented, including unknown long flag and missing value. Reuse existing `ErrConversion` and `ErrDuplicateValue`.
  - [ ] Add a typed parse-context error, such as `*flags.ParseError`, with inspectable accessors for category, token, raw long name, normalized lookup key where available, and canonical definition where applicable.
  - [ ] Preserve `errors.Is(err, flags.ErrConversion)` and `errors.As(err, *flags.ValueError)` for conversion failures from `Definition.Parse`; wrapping with parse context is acceptable only if both inspections still work.
  - [ ] Return typed unknown-flag errors for unknown long names before the terminator; include the original flag token and lookup context without relying on rendered text.
  - [ ] Return typed missing-value errors when a required value is omitted or the next token is `--` or another long-flag token.
  - [ ] Return typed duplicate-value errors when a non-repeatable flag appears more than once in one parse run.
  - [ ] Keep sensitive raw values out of error strings and debug output. Error context may identify the flag name/token but must not echo sensitive values.

- [ ] Add table-driven package tests and ATDD activation (AC: 1-5)
  - [ ] If `$bmad-testarch-atdd` has generated Story 2.3 scaffolds, activate one skipped ATDD test at a time, confirm RED with a narrow `go test ./flags -run ... -count=1`, then implement the smallest production change that passes.
  - [ ] Add focused package tests, likely `flags/parse_long_test.go`, for attached values, separate values, positionals, boolean presence, explicit booleans, invalid booleans, unknown long flags, missing values, duplicate single-value flags, exact matching, normalized matching, and `--no-*` as an ordinary registered/unregistered long name.
  - [ ] Assert parsed values through `Snapshot.Lookup(...).Values()`, explicit state, remaining args accessors, source metadata, and canonical definition identity.
  - [ ] Assert diagnostics through `errors.Is`, `errors.As`, and typed accessors; do not make exact error strings the only contract.
  - [ ] Add redaction-focused tests if sensitive values are present in parse errors.
  - [ ] Keep tests deterministic and standard-library-only.

- [ ] Update adoption-facing docs only for new public behavior (AC: 5)
  - [ ] Update `docs/behavior-matrices.md` if the new parse entrypoint, remaining-args snapshot, or long-flag behavior needs adoption evidence.
  - [ ] Update `docs/diagnostics-and-errors.md` if new parse sentinels or typed parse errors become part of the public contract.
  - [ ] Do not add compatibility examples, fuzz seeds, command/config integration docs, or release evidence; later stories own those surfaces.

- [ ] Verify the story implementation (AC: 1-5)
  - [ ] Run `go test ./...`.
  - [ ] Run `go vet ./...`.
  - [ ] Run `go run ./tools/depgate`.
  - [ ] Run `git diff --check`.
  - [ ] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [ ] Confirm no package imports outside the Go standard library and local module were added.
  - [ ] Consider `go test -race ./...` as extra evidence because snapshots and sets are reusable values.
  - [ ] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### ATDD Artifacts

- No Story 2.3 ATDD artifact existed under `_bmad-output/test-artifacts/` at story creation.
- Expected next workflow step: run `$bmad-testarch-atdd _bmad-output/implementation-artifacts/2-3-reject-invalid-long-flags-with-inspectable-errors.md`.
- Dev workflow handoff after ATDD generation: remove one generated `t.Skip` at a time, confirm RED with a narrow `go test ./flags -run ... -count=1`, then implement the smallest production change that passes.

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

### Completion Notes List

### File List

### Change Log

- 2026-06-11: Story created and marked ready for development.
