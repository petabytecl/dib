---
baseline_commit: d45dd20d268ecb5014429477119388a2ab850235
created: "2026-06-11T17:41:09-04:00"
---

# Story 2.2: Match Flag Names Safely Across Styles

Status: ready-for-dev

## Story

As a Go CLI developer,
I want flag names to match exactly by default and normalize only when I opt in,
so that familiar naming styles do not create silent collisions or surprising parse behavior.

## Requirements Trace

- FR10: Name normalization is caller-configured, detects definition-time collisions, and exact matching remains the default.
- FR20: Normalization behavior must be covered by table-driven tests and adoption-facing behavior evidence where useful.
- NFR1, NFR2, NFR3, NFR4, NFR6, NFR7, NFR8: runtime and tests remain standard-library-only, explicit-instance based, typed-error inspectable, deterministic, table-testable, compatibility-clear, and redaction-safe.

## Acceptance Criteria

1. Given no normalizer is configured, when flags named `log-level`, `log_level`, and `log.level` are registered or parsed, then exact names are used and no implicit equivalence is applied.
2. Given a caller configures a name normalizer, when equivalent names such as `log-level`, `log_level`, and `log.level` normalize to the same key, then the Flag set can resolve those names consistently and the parse snapshot reports the canonical definition rather than the raw spelling alone.
3. Given normalization creates a definition collision, when the Flag set is built or derived, then setup fails with a typed deterministic error and the diagnostic identifies the colliding flag names without relying on string-only matching.
4. Given shorthand names are one-character identities, when name normalization is configured, then shorthand uniqueness remains independently enforced and long-name normalization never creates hidden shorthand aliases.
5. Given later command and config stories depend on stable flag names, when verification runs, then table-driven tests cover exact matching, configured normalization, normalization collisions, shorthand uniqueness, and diagnostic context, and no runtime or test dependency outside the Go standard library is introduced.

## Tasks / Subtasks

- [ ] Confirm current tracker, artifact, and source state (AC: 1-5)
  - [ ] Verify `sprint-status.yaml` marks Story 2.1 `done` and Story 2.2 `ready-for-dev`.
  - [ ] Check for Story 2.2 ATDD artifacts under `_bmad-output/test-artifacts/`; none existed at story creation, but use them if they are generated before implementation starts.
  - [ ] Verify root `go.mod` still declares `module github.com/petabytecl/dib` and `go 1.26` with no `require`, `replace`, or `toolchain` directives.
  - [ ] Reuse the existing `flags.Set`, `Definition`, `Snapshot`, and typed error foundation instead of replacing it.

- [ ] Add an explicit name-normalization API (AC: 1, 2, 5)
  - [ ] Preserve `NewSet(defs ...Definition) (Set, error)` as exact-name-by-default behavior.
  - [ ] Add a focused caller opt-in surface for long-name normalization, such as a named constructor and/or immutable derivation method, without package globals or process state.
  - [ ] Define the normalizer shape with Go standard-library types only, for example a function from raw long name to normalized key.
  - [ ] Treat nil or absent normalizers deterministically; do not panic or silently enable equivalence.
  - [ ] Keep canonical `Definition.Name()` as the registered definition name, not the raw spelling used during lookup.

- [ ] Build normalized long-name resolution without mutating existing sets (AC: 1, 2, 4)
  - [ ] Add normalized lookup indexes inside `flags.Set` while preserving deterministic definition order.
  - [ ] Make exact sets resolve `log-level`, `log_level`, and `log.level` as three distinct long names.
  - [ ] Make normalized sets resolve supported raw spellings to the same canonical definition when the configured normalizer maps them to the same key.
  - [ ] Ensure `Set.With` and any new derivation API return a new set, preserve the original set's observable behavior, and validate normalization for the combined definitions.
  - [ ] Keep shorthand indexes independent from long-name normalization; a long-name normalizer must never create or accept hidden shorthand aliases.

- [ ] Add typed deterministic normalization-collision diagnostics (AC: 3, 5)
  - [ ] Add a sentinel or typed category for normalization collisions, for example `ErrDuplicateNormalizedName` or `ErrNameNormalizationCollision`.
  - [ ] Extend `DefinitionError` or add a small typed error so callers can inspect both colliding long flag names programmatically with `errors.As`; do not make tests depend on error strings.
  - [ ] Preserve existing `DefinitionError.Name()` and `DefinitionError.Shorthand()` behavior for Story 2.1 errors.
  - [ ] Reject invalid normalized definition keys deterministically if a configured normalizer produces an empty or otherwise unusable long-name key.
  - [ ] Keep raw sensitive values out of collision and validation diagnostics.

- [ ] Connect normalization to snapshot/parse foundations only as far as this story needs (AC: 2, 5)
  - [ ] Provide enough name-resolution state for later long-flag parsing to record canonical definition identity separately from raw CLI spelling.
  - [ ] If modifying `Snapshot`, keep it self-contained, keyed by canonical definition names, and free of caller-owned mutable aliases.
  - [ ] Do not implement full `--name=value`, `--name value`, unknown-flag, missing-value, or shorthand parse behavior from Stories 2.3 and 2.4 unless an ATDD scaffold for Story 2.2 explicitly requires a minimal hook.

- [ ] Add tests and documentation evidence (AC: 1-5)
  - [ ] Add table-driven package tests, likely in a new `flags/normalize_test.go` plus focused additions to `flags/set_test.go` or `flags/errors_test.go`.
  - [ ] Cover exact default matching for `log-level`, `log_level`, and `log.level`.
  - [ ] Cover configured normalization with a pure test normalizer that maps `-`, `_`, and `.` consistently.
  - [ ] Cover normalized collision errors through `errors.Is` and/or `errors.As`, including machine-readable access to both colliding names.
  - [ ] Cover derived normalized sets, original-set immutability, shorthand uniqueness, and the absence of hidden shorthand aliases.
  - [ ] Update `docs/behavior-matrices.md` only if the new public behavior needs an adoption-facing row; keep package tests as the executable contract.

- [ ] Verify the story output (AC: 1-5)
  - [ ] Run `go test ./...`.
  - [ ] Run `go vet ./...`.
  - [ ] Run `go run ./tools/depgate`.
  - [ ] Run `git diff --check`.
  - [ ] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [ ] Confirm no package imports outside the Go standard library were added.
  - [ ] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### ATDD Artifacts

- Checklist: `_bmad-output/test-artifacts/atdd-checklist-2-2-match-flag-names-safely-across-styles.md`
- Backend package acceptance scaffold:
  - `flags/normalize_atdd_test.go`
- Temp API/back-end generation summary: `_bmad-output/test-artifacts/tmp/tea-atdd-api-tests-2026-06-11T18-14-04-0400.json`
- Temp E2E generation summary: `_bmad-output/test-artifacts/tmp/tea-atdd-e2e-tests-2026-06-11T18-14-04-0400.json`
- Temp aggregate summary: `_bmad-output/test-artifacts/tmp/tea-atdd-summary-2026-06-11T18-14-04-0400.json`
- Dev workflow handoff: remove one `t.Skip` in `flags/normalize_atdd_test.go` at a time, confirm RED with the narrow `go test ./flags -run ... -count=1` command, then implement the smallest change to pass.

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`.
- Loaded previous story: `_bmad-output/implementation-artifacts/2-1-define-reusable-flag-sets-without-global-state.md`.
- Loaded current source/docs: `flags/set.go`, `flags/definition.go`, `flags/errors.go`, `flags/snapshot.go`, `flags/set_test.go`, `docs/behavior-matrices.md`, and `docs/diagnostics-and-errors.md`.
- No UX document, `project-context.md`, `CLAUDE.md`, local `MEMORY.md`, or Story 2.2 ATDD artifact was discovered in the repo at story creation. Story 2.2 ATDD artifacts were generated after story creation and are linked above.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `d45dd20d268ecb5014429477119388a2ab850235` (`fix: address flag set review findings`).
- `main` is aligned with `origin/main` at story creation.
- Story 2.1 is `done`; it implemented the reusable `flags.Set`, value-state, parser, and typed-error foundation this story must extend.
- Root `go.mod` contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

### Architecture Guardrails

- `flags/` owns explicit flag sets, long/shorthand parsing, repeated/custom values, no-option defaults, normalization, parse diagnostics, usage metadata, and flag-specific typed errors. [Source: `_bmad-output/planning-artifacts/architecture.md:562`]
- Shared flag metadata and parsing semantics live in `flags/`, and `flags/` must remain fully usable without `command/` or `config/`. [Source: `_bmad-output/planning-artifacts/architecture.md:567`, `_bmad-output/planning-artifacts/architecture.md:570`]
- The module root must not provide a broad public facade, package-global helpers, or default singleton API. [Source: `_bmad-output/planning-artifacts/architecture.md:564`]
- Definitions, Flag sets, and snapshots are reusable values; derived definitions return new values and must avoid exported mutable internals or shallow-copy aliasing. [Source: `_bmad-output/planning-artifacts/architecture.md:205`, `_bmad-output/planning-artifacts/architecture.md:207`]
- Setup-time validation must catch invalid definitions, duplicate names, and normalization collisions where possible. [Source: `_bmad-output/planning-artifacts/architecture.md:217`]
- Public APIs use explicit inputs and returned values/errors and must not depend on package globals, hidden process IO, implicit environment reads, or root singletons. [Source: `_bmad-output/planning-artifacts/architecture.md:229`]
- Public errors must support `errors.Is` / `errors.As` compatible inspection; error strings are diagnostics, not programmatic contracts. [Source: `_bmad-output/planning-artifacts/architecture.md:231`]
- Public source remains organized by capability package; shared code belongs under `internal/` only after multiple concrete call sites prove the need. [Source: `_bmad-output/planning-artifacts/architecture.md:636`, `_bmad-output/planning-artifacts/architecture.md:637`]
- Package tests are the executable contract and must live beside package code. [Source: `_bmad-output/planning-artifacts/architecture.md:642`, `_bmad-output/planning-artifacts/architecture.md:654`]

### Requirements Notes

- Epic 2 covers inspectable flag parsing without package-global state, including explicit flag sets, long/shorthand parsing, repeated/custom values, terminators, and typed parse errors. [Source: `_bmad-output/planning-artifacts/epics.md:177`, `_bmad-output/planning-artifacts/epics.md:181`]
- Story 2.2 requires exact matching by default, caller-configured normalization, deterministic collision errors, independently enforced one-character shorthands, and standard-library-only tests. [Source: `_bmad-output/planning-artifacts/epics.md:408`, `_bmad-output/planning-artifacts/epics.md:441`]
- FR10 requires a configured normalization function to map equivalent names, detect normalization collisions at definition time, and leave exact names in place when no normalization is configured. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:176`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:183`]
- FR20 requires table-driven behavior tests for parser behavior; this story should keep coverage focused on definition-time and name-resolution behavior that exists now. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:281`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:288`]
- NFRs require standard-library-only runtime packages, explicit-instance APIs, typed errors, deterministic output, no default process control, table-driven tests, compatibility clarity, and redaction-safe sensitive diagnostics. [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:310`, `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:318`]

### Current Code Context

- `flags.Set` currently stores immutable definitions plus exact `byName` and `byShort` indexes; `NewSet` validates definitions, rejects duplicate exact long names, and rejects duplicate shorthands. [Source: `flags/set.go:3`, `flags/set.go:38`]
- `Set.Lookup` currently performs exact long-name lookup only, and `Set.With` currently derives through `NewSet(next...)`. [Source: `flags/set.go:51`, `flags/set.go:65`]
- `DefaultSnapshot` currently keys value states by the registered definition name. Keep snapshot keys canonical if this story adds name-resolution metadata. [Source: `flags/set.go:67`, `flags/set.go:73`]
- `Definition.Name()` already returns the registered long flag name; use it as the canonical name exposed by normalized resolution. [Source: `flags/definition.go:163`, `flags/definition.go:166`]
- `flags/errors.go` already has sentinel setup errors and `DefinitionError` accessors for name and shorthand. Extend this pattern instead of introducing string-only diagnostics. [Source: `flags/errors.go:8`, `flags/errors.go:65`]
- `docs/behavior-matrices.md` already records immutable definitions, no mutable aliases, per-run snapshots, explicit instances, public error inspection, diagnostic vocabulary, and redaction corpus. Story 2.2 should add only concise normalization evidence if needed. [Source: `docs/behavior-matrices.md:9`, `docs/behavior-matrices.md:18`]
- `docs/diagnostics-and-errors.md` states that programmatic errors must be inspectable through Go error inspection and that diagnostic strings are not the programmatic contract. [Source: `docs/diagnostics-and-errors.md:7`, `docs/diagnostics-and-errors.md:20`]

### Scope Boundaries

Likely implementation targets:

```text
flags/set.go
flags/normalize.go
flags/errors.go
flags/snapshot.go
flags/normalize_test.go
flags/set_test.go
flags/errors_test.go
docs/behavior-matrices.md
docs/diagnostics-and-errors.md
```

The exact file split may vary, but keep files focused. Do not create or modify these unless a failing test or ATDD scaffold proves they are required now:

```text
flags/parse.go
flags/parse_long_test.go
flags/parse_shorthand_test.go
flags/fuzz_test.go
flags/testdata/fuzz/FuzzParse/
cmd/
config/
internal/
go.sum
```

Out of scope for this story:

- Full long-flag parse forms from Story 2.3, including `--name=value`, `--name value`, unknown flags, and missing values.
- Shorthand parsing and shorthand groups from Stories 2.4 and 2.5.
- Repeated/custom CLI accumulation behavior from Story 2.6 beyond preserving existing metadata.
- Parse terminator and remaining-args behavior from Story 2.7.
- Parser fuzzing and full behavior-matrix proof from Story 2.8.
- Compatibility adapters or source-compatible clones for `flag`, pflag, Cobra, or Viper.

### Technical Research Notes

- GitHub code and repository searches for pflag-style normalization returned no local code to copy into Dib. Keep the implementation clean-room and test-driven.
- pflag provides useful prior art for opt-in name normalization and examples where `-`, `_`, and `.` can compare the same, but Dib must not import pflag or inherit mutable/global behavior. Source: https://pkg.go.dev/github.com/spf13/pflag
- pflag issue history shows normalization can create surprising alias/deprecation side effects if canonical identity and aliases are not modeled carefully. Dib should keep canonical registered definitions separate from raw lookup spelling. Source: https://github.com/spf13/pflag/issues/366
- `go list -m -versions github.com/spf13/pflag` showed public versions through `v1.0.10`; this is prior-art context only. The project dependency rule still forbids adding pflag or any other third-party runtime/test dependency.
- Use standard-library helpers such as `strings.NewReplacer` in tests when useful; do not add package dependencies for simple normalization.

### Previous Story Intelligence

- Story 2.1 implemented reusable `flags.Set` construction, exact lookup, deterministic definition inspection, immutable derivation with `With`, default snapshots, built-in kinds, custom parser support, and inspectable definition/value errors.
- Story 2.1 review fixed five issues that must not regress: sensitive parser errors do not expose parser causes; custom defaults/results do not leak aliases; typed nil parsers are rejected; custom kind/default/parser mismatches are rejected; nil options are setup validation errors.
- Story 2.1 established the current verification baseline: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `git diff --check`, and optional `go test -race ./...`.
- Story 1.4 implemented `tools/depgate`; it remains the dependency authority and should not be imported by runtime packages.
- Story 1.3 established `docs/behavior-matrices.md` and `docs/diagnostics-and-errors.md`; update them only for public behavior evidence, not as a substitute for package tests.

### Testing Standards

- Follow red-green-refactor. If Story 2.2 ATDD scaffolds are present, activate one skipped test at a time, confirm RED with a narrow `go test ./flags -run ... -count=1`, then implement the smallest production change that passes.
- Add table-driven package tests before or alongside each behavior. Keep tests standard-library-only.
- Required final verification for implementation: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, and `git diff --check`.
- Use `go test -race ./...` as optional extra evidence if the implementation touches concurrency-sensitive state or shared aliases.
- Prefer observable assertions over internals: exact lookup behavior, normalized lookup behavior, inspectable collision errors, original/derived set independence, and canonical definition names returned from lookups.

### Security And Quality Checks

- Do not hardcode secrets, credentials, tokens, or private URLs.
- Treat CLI spellings and normalizer output as untrusted boundary data.
- Sensitive raw values must not appear in errors, `String` output, debug text, rendered diagnostics, source reports, examples, or validation failures.
- Do not import non-standard-library modules in runtime, tests, examples, or tools.
- Do not add `os.Args`, `flag.CommandLine`, package-global mutable registries, hidden stdout/stderr use, or `os.Exit`.
- Keep public errors inspectable with `errors.Is` / `errors.As`; do not make string matching the contract.
- Normalizer callbacks in tests should be pure and deterministic. If the public API accepts caller functions, document that they must be deterministic because Dib cannot make mutable closures immutable.

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

### Completion Notes List

### File List

### Change Log

- 2026-06-11: Story created and marked ready for development.
