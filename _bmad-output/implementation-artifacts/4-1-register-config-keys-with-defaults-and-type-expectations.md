---
baseline_commit: ad1af8a
created: "2026-06-12T02:49:45-04:00"
---

# Story 4.1: Register Config Keys With Defaults And Type Expectations

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Go CLI developer,
I want reusable Config key definitions with defaults, types, and sensitivity metadata,
so that config resolution starts from explicit, inspectable contracts rather than ad hoc map lookups.

## Requirements Trace

- FR11: register Config keys with default values, type expectations, optional documentation metadata, exact matching by default, optional normalization, and deterministic setup errors.
- FR16: expose enough retrieval foundation to distinguish default/not-found behavior and mark sensitive values for future diagnostics redaction.
- FR20: prove config definition behavior through executable table-driven tests and behavior-matrix evidence.
- NFR1: runtime packages, tests, examples, and tooling remain standard-library-only unless architecture changes.
- NFR2: primary APIs use explicit instances and caller-supplied inputs; no package-level global config registry or default singleton.
- NFR3: public error cases needed by callers are inspectable without string matching.
- NFR6: behavior is testable with table-driven unit tests and no ambient process state.
- NFR8: sensitive metadata prevents fake sensitive values from leaking in diagnostics when config errors or debug strings exist.

## Acceptance Criteria

1. Given a caller registers a Config key, when the key definition is created, then it captures the stable key name, optional default value, type expectation, documentation metadata, and sensitivity classification, and exact key matching is used by default.
2. Given a caller opts into key normalization, when normalized keys collide, then setup fails with a typed deterministic error, and collisions are detected before resolution.
3. Given no higher-precedence source sets a registered key, when the key is resolved, then the default value can be returned with `default` provenance, and missing unregistered keys return a documented not-found result rather than panicking.
4. Given redaction must be defined before source diagnostics expand, when sensitive metadata is configured, then diagnostics may identify the key and source label, and they must not echo `dib_fake_secret_value`, `dib_fake_password_value`, or `dib_fake_token_value`.
5. Given config definitions must be reusable, when verification runs, then tests cover defaults, not-found results, type expectations, exact matching, normalization collisions, sensitivity metadata, immutable definition reuse, and redaction corpus false positives/false negatives.

## Tasks / Subtasks

- [x] Confirm tracker, artifacts, and source state (AC: 1-5)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 4 `in-progress` and Story 4.1 `ready-for-dev` before implementation starts.
  - [x] Check for Story 4.1 ATDD/test artifacts under `_bmad-output/test-artifacts/`; none existed at story creation, but use them if generated before implementation.
  - [x] Verify `go.mod` still contains only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or generated `go.sum`.
  - [x] Read current config, flags, command, and docs source before editing: `config/doc.go`, `flags/definition.go`, `flags/set.go`, `flags/normalize.go`, `flags/snapshot.go`, `flags/errors.go`, `flags/*_test.go` patterns relevant to definitions/normalization/redaction, `docs/behavior-matrices.md`, and `docs/diagnostics-and-errors.md`.

- [x] Define config key type vocabulary and metadata (AC: 1, 4, 5)
  - [x] Add a public config type expectation model, likely `config.Kind`, aligned with existing flag value kinds where sensible: string, bool, int, int64, uint, uint64, float64, duration, and string list.
  - [x] Prefer independent config names such as `config.KindString` rather than importing `flags` solely for kind constants. `config/` must remain independently usable and should not depend on `command/`.
  - [x] Add reusable key definition constructors or a single constructor with typed helpers. The public API should make the key name, default value, type expectation, documentation/usage text, and sensitivity metadata inspectable through accessors.
  - [x] Preserve the `flags` package pattern: unexported fields, exported accessors, constructor-time validation, defensive copies for mutable defaults, and derived values that do not mutate originals.
  - [x] Treat default values as optional. A registered key may have no default; absence must be distinguishable from a default zero value.
  - [x] If `[]string` defaults are supported, return defensive copies from accessors and snapshots. Do not expose caller-owned slices.
  - [x] Do not add source-compatible Viper names, struct decoding APIs, compatibility aliases, reflection-heavy mapping, or package-global registries.

- [x] Implement definition set construction and lookup (AC: 1, 2, 5)
  - [x] Add an immutable config definition set, likely `config.Set`, with `NewSet(defs ...Definition) (Set, error)` for exact lookup by default.
  - [x] Add opt-in key normalization, likely `config.NameNormalizer`, `NewNormalizedSet(normalizer, defs...)`, and `Set.WithNormalizer(normalizer)`, mirroring the `flags` semantics without sharing unneeded implementation.
  - [x] Exact matching is the default. Keys such as `log-level`, `log_level`, and `log.level` must remain distinct unless a normalizer is supplied.
  - [x] Detect duplicate exact keys and duplicate normalized keys at setup time, before any resolution.
  - [x] Validate key names before and after normalization. Follow existing flag invalid-name behavior unless architecture requires a different config key rule: reject empty, whitespace-containing, and leading-hyphen keys.
  - [x] Ensure `Set.Definitions()` returns deterministic registration order and defensive values.
  - [x] Ensure `Set.With(...)` and `Set.WithNormalizer(...)` return new sets without mutating the receiver.

- [x] Add config typed setup errors (AC: 2, 3, 4, 5)
  - [x] Add config-specific sentinels and typed errors in `config/errors.go`, following `flags.DefinitionError` style.
  - [x] Expected setup categories include invalid definition, duplicate key, duplicate normalized key, and invalid default/type mismatch.
  - [x] Add a documented not-found result for unregistered keys. Prefer a structured return shape such as `(Value, bool)` where `bool=false` indicates not found, or an inspectable `ErrNotFound` only if the API needs an error path. The story requires no panic.
  - [x] Public errors must support `errors.Is` and/or `errors.As` and expose the key, colliding key, normalized key, expected kind, and provenance context where relevant.
  - [x] Error strings are diagnostics only. Tests must assert typed contracts first and only check strings for redaction or deliberate human-facing text.
  - [x] Sensitive failures must not echo default or rejected raw values when the key is marked sensitive.

- [x] Provide the minimal default-resolution snapshot (AC: 3, 5)
  - [x] Add just enough resolution API for this story to return registered defaults with `default` provenance and to report unregistered keys without panic.
  - [x] Use the exact source label `default`, matching `docs/diagnostics-and-errors.md`.
  - [x] Keep this scope intentionally narrow: do not implement explicit setters, flag bindings, env bindings, JSON loading, full precedence, source reports, or typed retrieval beyond what is required to prove default and not-found behavior.
  - [x] The default snapshot must be self-contained and reusable. It must not depend on live environment variables, filesystem state, readers, command routes, flag parses, wall clock, or process args after creation.
  - [x] A no-default registered key must remain distinguishable from a key with a default zero value such as `""`, `0`, `false`, or an empty `[]string`.
  - [x] Preserve immutable definitions after resolution; resolving defaults must not mutate definitions or sets.

- [x] Add tests and external consumer contract coverage (AC: 1-5)
  - [x] Add package-local table-driven tests under `config/`, likely `definition_test.go`, `set_test.go`, `normalize_test.go`, `errors_test.go`, and `snapshot_test.go` or similarly focused files.
  - [x] Add external-package tests using `package config_test` so only exported API is exercised.
  - [x] If the existing consumer-contract helper pattern is useful, add a config-local equivalent rather than importing test helpers across packages.
  - [x] Test key metadata for every supported kind, default presence, documentation text, sensitivity metadata, and defensive copy behavior.
  - [x] Test exact matching, configured normalization, normalized collision diagnostics, invalid normalized keys, and `With`/`WithNormalizer` immutability.
  - [x] Test default resolution with `default` provenance, not-found behavior for unregistered keys, no-default registered keys, explicit zero defaults, and repeated/concurrent reuse if the API stores reusable state.
  - [x] Test sensitive redaction with the exact corpus values: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.
  - [x] Test false positives/false negatives: non-sensitive keys may expose ordinary non-secret invalid values where appropriate, while sensitive keys must redact the fake corpus.

- [x] Update documentation only after executable evidence exists (AC: 1-5)
  - [x] Update `docs/behavior-matrices.md` shared contracts and add the first config evidence rows with exact test function names.
  - [x] Update `docs/diagnostics-and-errors.md` with Story 4.1 config setup/default/not-found diagnostic categories only after tests prove the public contracts.
  - [x] Keep `docs/config-precedence.md` out of this story unless the implementation creates a real default-resolution contract that needs a minimal canonical doc. Do not write speculative precedence docs for sources not implemented yet.
  - [x] Update `config/doc.go` only for implemented behavior: registered keys, defaults, type expectations, sensitivity metadata, exact/normalized lookup, and default provenance.
  - [x] Update `docs/provenance-log.md` only if implementation or docs were influenced by external material beyond project-owned artifacts and official Go reference facts.

- [x] Preserve Story 4.1 scope boundaries (AC: 1-5)
  - [x] Do not implement explicit setters, env lookup, JSON readers/paths, flag binding, full precedence, source reports, config file fixtures, migration examples, compatibility docs, struct decoding, live reload, remote config, dotenv/YAML/TOML/HCL/INI/properties, shell completion, `/cmd` scaffolding, root facade APIs, or package-global registries.
  - [x] Do not import `command/` from `config/`. If a `config -> flags` import is considered, it must be justified by exported flag snapshot/value contracts; for Story 4.1 it should not be needed.
  - [x] Do not add external runtime, test, assertion, fuzzing, or tooling dependencies.
  - [x] Do not copy source, tests, examples, fixtures, internal names, or file organization from Viper, pflag, Cobra, Go `flag`, or other CLI/config projects.

- [x] Verify the story implementation (AC: 1-5)
  - [x] Run focused config tests, for example `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./config -count=1`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`.
  - [x] Run `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`.
  - [x] Run `git diff --check`.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [x] Consider `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./config ./... -count=1` if reusable sets, snapshots, or concurrent tests are added.
  - [x] Record exact commands and outcomes in the Dev Agent Record.

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD and addendum material from `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/`.
- No UX artifact was discovered; Dib V1 has no browser UI or frontend UX surface.
- No `project-context.md` file was discovered under the repository.
- No Story 4.1 ATDD/test artifact was discovered under `_bmad-output/test-artifacts/`.
- Loaded previous-story and previous-epic intelligence from `_bmad-output/implementation-artifacts/3-5-preserve-caller-controlled-execution-boundaries.md` and `_bmad-output/implementation-artifacts/epic-3-retro-2026-06-12.md`.
- Loaded current config stub, established flags definition/set/normalization/snapshot/error patterns, command contract tests, and docs behavior/diagnostic guidance.

### Current Repository State

- Baseline commit at story creation: `ad1af8a` (`docs: add epic 3 retrospective`).
- Existing worktree has unrelated BMAD installer/config, `.agents/`, `.codex/`, `.idea/`, and story-automator changes. Do not revert or normalize them while implementing this story.
- Root `go.mod` currently contains only:

```text
module github.com/petabytecl/dib

go 1.26
```

- No `go.sum` file exists at story creation.
- `config/` currently contains only `doc.go`; Story 4.1 is the first real config implementation story.
- `sprint-status.yaml` has Epic 4 moved to `in-progress` and Story 4.1 moved to `ready-for-dev` by this create-story workflow.

### Architecture Guardrails

- Dib V1 is organized around `command/`, `flags/`, and `config/`; each surface must remain independently usable before callers compose them. [Source: `_bmad-output/planning-artifacts/architecture.md#Context`]
- Runtime packages must import only the Go standard library; tests and examples must also remain dependency-free unless the architecture changes. [Source: `_bmad-output/planning-artifacts/architecture.md#Technical-Summary`]
- Primary APIs must use explicit instances and returned values/errors. No package-level global command, flag, or config helpers are allowed. [Source: `_bmad-output/planning-artifacts/architecture.md#Technical-Summary`]
- Definitions and snapshots are reusable values. Derived definitions return new values, per-run snapshots do not mutate definitions, and exported APIs must not expose mutable aliases. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Architecture`]
- `config/` owns caller-owned reusable key definitions, defaults, explicit setters, flag binding inputs, env bindings, JSON readers/files, precedence, typed getters, provenance, redaction, and config-specific typed errors. Story 4.1 owns only the first slice of that surface. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- `command/` must not depend on `config/`; `flags/` must remain usable without `command/` or `config/`; `config/` may later accept explicit flag binding inputs without depending on command internals. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- Config data enters through explicit setters, parsed flag snapshots or binding inputs, injected env lookup, JSON paths/readers, and defaults. Story 4.1 should implement only defaults and registered-key contracts. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Boundaries`]
- Config provenance/source-label vocabulary is closed for V1: `default`, `explicit setter`, `flag binding`, `env`, and `JSON`. Story 4.1 uses only `default`. [Source: `docs/diagnostics-and-errors.md#Source-Labels`]
- Sensitive values must be redacted across errors, debug strings, diagnostics, source reports, and examples once those surfaces exist. The fixed fake corpus is `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Fixture-Patterns`]

### Current Code Context

- `config/doc.go` currently documents explicit config keys/sources/resolution contracts but has no implementation. This story will update `config/`, not `command/` or `flags/`.
- `flags/definition.go` is the closest local pattern for immutable public definitions:
  - public `Definition` values hold unexported fields;
  - constructors such as `String`, `Bool`, `Int`, `Duration`, and `StringList` clone mutable defaults;
  - options such as `Sensitive()` set metadata;
  - accessors such as `Name()`, `Kind()`, `Default()`, `Usage()`, and `Sensitive()` expose state defensively.
- `flags/set.go` is the closest local pattern for definition sets:
  - `NewSet` exact-matches by default;
  - `NewNormalizedSet` opts into normalization;
  - duplicate exact names and duplicate normalized names fail at setup;
  - `Definitions()`, `With`, and `WithNormalizer` preserve deterministic order and do not mutate prior sets.
- `flags/normalize.go` defines a deterministic, side-effect-free normalizer contract. Reuse the concept for config keys, but do not create shared internal helpers unless a second concrete need proves it.
- `flags/errors.go` shows the typed setup error pattern: sentinels, an inspectable `DefinitionError`, accessors for name/collision/normalized-name context, and `Unwrap` support for `errors.Is`.
- `flags/snapshot.go` shows snapshot/value-state defensive accessors and default state capture. Config default-resolution snapshots should follow this style without depending on flag parser internals.
- `command/contract_test.go` and `flags/*_test.go` show the current testing style: external-package tests, table-driven cases, explicit process-boundary assertions, typed error inspection, and no third-party assertions.
- `docs/behavior-matrices.md` already has later hooks for config definitions, snapshots, source labels, typed errors, and redaction. Story 4.1 should replace the relevant later hooks with current evidence after tests exist.
- `docs/diagnostics-and-errors.md` currently says later stories own config provenance and remaining config error categories. Story 4.1 should add only setup/default/not-found guidance that is implemented and tested.

### Previous Story Intelligence

- Epic 3 completed without changing Epic 4 direction. Config can now rely on exported `flags` snapshots later, but Story 4.1 should not bind flags yet.
- Epic 3 retrospective action items for Epic 4:
  - carry command evidence patterns into config stories;
  - keep `config/` independent from command internals;
  - add behavior-matrix evidence tied to concrete tests for provenance, typed errors, redaction, immutable snapshots, and explicit boundaries;
  - verify docs against code before closing the epic.
- Story 3.5 preserved caller-owned execution boundaries and confirmed that command routing does not own process lifecycle. Config should follow the same explicit-boundary posture: no ambient env, filesystem, stdout/stderr, or process args in Story 4.1.
- Recent story reviews corrected artifact/file-list drift. Keep the Dev Agent Record file list exact and include `_bmad-output/implementation-artifacts/tests/test-summary.md` if QA automation updates it.

### Git Intelligence

- Recent commits:
  - `ad1af8a docs: add epic 3 retrospective`
  - `9112e13 feat(story-3.5): preserve execution boundaries`
  - `510b733 feat(story-3.4): render command help`
  - `8fa4448 feat(story-3.3): compose command flags`
  - `1e7d1af feat(story-3.2): resolve command aliases`
- Story 3.5 touched `command/boundary.go`, boundary tests, `command/doc.go`, `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, sprint status, story artifacts, and test summary artifacts.
- Existing pattern: add focused package-local tests first, keep public contracts inspectable through accessors and `errors.Is`/`errors.As`, document evidence rows after tests exist, and run focused package tests plus repository-wide test/vet/depgate/diff checks.

### Latest Technical Information

- Official Go downloads list `go1.26.4` as the current Go 1.26 stable release on 2026-06-12. Keep the module directive at `go 1.26`; do not add a `toolchain` directive unless a separate version-policy story approves it. [Source: `https://go.dev/dl/`]
- Use only the Go standard library for config implementation and tests. Relevant standard packages are likely `errors`, `fmt`, `reflect`, `strings`, `sync`, `testing`, and `time`; do not add assertion, mocking, or config parsing dependencies.
- JSON loading is intentionally later. Do not import `encoding/json` in Story 4.1 unless a minimal default-value validation path genuinely requires it, which should be unlikely.

### Testing Standards

- Treat package tests as executable truth; docs must cite tests that actually exist after implementation.
- Use table-driven tests for key metadata, exact lookup, normalized lookup, setup errors, default resolution, not-found results, no-default vs zero-default distinction, sensitivity metadata, redaction, immutable sets, and reusable snapshots.
- Assert typed config diagnostics with `errors.Is` and/or `errors.As` whenever caller inspection is part of the contract.
- Assert structured state rather than rendered strings for provenance and not-found behavior.
- Keep fixtures local to `config/`; Story 4.1 should not need JSON fixtures.
- Avoid live env, current working directory, process args, stdin/stdout/stderr, wall clock, host files, and third-party dependencies.
- If concurrency/reuse is part of the observable contract, add repeated and concurrent tests without data races.

### Security And Quality Checks

- Use the architecture-owned fake sensitive-value corpus exactly: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.
- Do not hardcode real secrets, credentials, tokens, private URLs, or host-specific paths.
- Do not copy source, tests, comments, fixtures, examples, internal names, command/config API shape, or file organization from Viper, pflag, Cobra, Go `flag`, or other CLI projects.
- Do not add external imports to runtime, tests, examples, fuzzing, or tooling.
- Do not add package-global config registries, default config singletons, implicit env readers, `os.Getenv`, file reads, `os.Args`, stdout/stderr writes, `os.Exit`, `/cmd` scaffolding, migration adapters, or broad root facade APIs.
- New docs must not claim explicit setters, env, JSON, flag binding, full precedence, source reports, typed getters, migration support, release readiness, or future API stability until those stories implement them.

### Project Structure Notes

- Expected Story 4.1 source files are likely:
  - UPDATE `config/doc.go`
  - ADD `config/kind.go`
  - ADD `config/definition.go`
  - ADD `config/set.go`
  - ADD `config/normalize.go`
  - ADD `config/snapshot.go` or `config/defaults.go`
  - ADD `config/errors.go`
  - ADD `config/definition_test.go`
  - ADD `config/set_test.go`
  - ADD `config/normalize_test.go`
  - ADD `config/errors_test.go`
  - ADD `config/snapshot_test.go` or equivalent default-resolution tests
  - UPDATE `docs/behavior-matrices.md`
  - UPDATE `docs/diagnostics-and-errors.md`
  - UPDATE `_bmad-output/implementation-artifacts/tests/test-summary.md` if QA automation updates it
- Do not create `config/json.go`, `config/binding_env.go`, `config/binding_flag.go`, `config/testdata/json/`, `docs/config-precedence.md`, `examples/`, compatibility docs, migration docs, shell-completion assets, generated man pages, or root facade files for this story unless an implementation necessity is documented and still fits AC 1-5.
- No structure conflict detected: the architecture reserves `config/` for this surface, and the package currently has only `doc.go`.

### Files To Read Before Editing

- `config/doc.go`: current config package contract language and package docs.
- `flags/definition.go`: local definition constructor/accessor/option style and defensive default copying pattern.
- `flags/set.go`: local immutable set, exact lookup, normalized lookup, `With`, and default snapshot pattern.
- `flags/normalize.go`: local normalizer contract and nil-normalizer behavior.
- `flags/snapshot.go`: local snapshot/value-state defensive accessors.
- `flags/errors.go`: local sentinel and typed setup/value error style.
- `flags/set_test.go`, `flags/normalize_test.go`, `flags/set_atdd_test.go`: local behavior tests for definition metadata, exact lookup, normalized collisions, and consumer contracts.
- `command/contract_test.go`: explicit-instance and process-boundary testing style.
- `docs/behavior-matrices.md`: add current config evidence rows after exact test names exist.
- `docs/diagnostics-and-errors.md`: add config setup/default/not-found guidance only after implementation exists.
- `docs/clean-room-policy.md` and `docs/provenance-log.md`: preserve clean-room restrictions and record external influence only if needed.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md#Story-41-Register-Config-Keys-With-Defaults-And-Type-Expectations`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-11-Register-Config-keys-and-defaults`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-16-Retrieve-typed-Config-values`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-20-Provide-behavior-test-matrices`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#Config-Semantics-Table`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural-Boundaries`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Flow`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Format-Patterns`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Runtime-Boundary-Patterns`]
- [Source: `_bmad-output/implementation-artifacts/epic-3-retro-2026-06-12.md#Next-Epic-Preview`]
- [Source: `config/doc.go`]
- [Source: `flags/definition.go`]
- [Source: `flags/set.go`]
- [Source: `flags/normalize.go`]
- [Source: `flags/snapshot.go`]
- [Source: `flags/errors.go`]
- [Source: `docs/behavior-matrices.md`]
- [Source: `docs/diagnostics-and-errors.md`]
- [Source: `https://go.dev/dl/`]

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- 2026-06-12: Resolved `bmad-dev-story` workflow customization; no activation prepend/append steps configured.
- 2026-06-12: Confirmed no `project-context.md` exists, no Story 4.1 ATDD artifact exists under `_bmad-output/test-artifacts/`, `go.mod` is minimal, and no `go.sum` exists.
- 2026-06-12: Red phase: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./config -count=1` failed because the new `config` API was undefined.
- 2026-06-12: Green/refactor phase: implemented config definitions, sets, normalization, typed errors, default snapshots, tests, and docs.
- 2026-06-12: Validation PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./config -count=1`.
- 2026-06-12: Validation PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`.
- 2026-06-12: Validation PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`.
- 2026-06-12: Validation PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`.
- 2026-06-12: Validation PASS: `git diff --check`.
- 2026-06-12: Validation PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./config ./... -count=1`.
- 2026-06-12: Confirmed `go.mod` still contains only `module github.com/petabytecl/dib` and `go 1.26`; no `go.sum` was created.
- 2026-06-12: QA automation PASS: generated Story 4.1 public workflow and setup-error tests in `config/qa_e2e_test.go`.
- 2026-06-12: QA automation PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./config -count=1`.
- 2026-06-12: QA automation PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`.
- 2026-06-12: QA automation PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`.
- 2026-06-12: QA automation PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`.
- 2026-06-12: QA automation PASS: `git diff --check`.

### Implementation Plan

- Add an independent `config.Kind` vocabulary and immutable `Definition` constructors/accessors without importing `flags` or `command`.
- Add immutable `Set` construction with exact lookup by default, opt-in normalization, duplicate validation, and non-mutating derivation.
- Add typed config setup errors with inspectable key, collision, normalized key, kind, and default provenance context.
- Add minimal default snapshot resolution with `default` provenance, no-default distinction, not-found boolean results, and defensive value accessors.
- Back the public API with external-package table tests, then update docs with exact executable evidence.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Implemented Story 4.1 config definitions with optional defaults, type expectations, usage text, sensitivity metadata, and defensive `[]string` handling.
- Implemented immutable config sets with exact lookup by default, opt-in key normalization, setup-time duplicate/normalized collision validation, and non-mutating derivation.
- Implemented typed config setup errors and redaction-safe sensitive default diagnostics.
- Implemented minimal default snapshots with `default` provenance, not-found lookup results, no-default distinction, and reusable/concurrent-safe defensive accessors.
- Added external-package config tests covering metadata, kind vocabulary, exact/normalized lookup, setup errors, default resolution, immutability, redaction corpus behavior, and race-safe reuse.
- Updated config package documentation and behavior/diagnostic docs with Story 4.1 evidence only.

### Senior Developer Review (AI)

Reviewer: Coto on 2026-06-12

Outcome: Approved. No verified implementation defects remained after review.

Checklist validation:
- Story file loaded and confirmed reviewable (`Status: review` before this review update).
- Epic/story resolved as 4.1.
- Architecture guidance loaded from `_bmad-output/planning-artifacts/architecture.md`; no `project-context.md` or Story 4.1 ATDD artifact was present in the repository.
- Acceptance Criteria 1-5 cross-checked against `config/` implementation, tests, and docs.
- File List was checked against relevant source/doc/artifact changes. The implementation files listed for Story 4.1 match the reviewed config source, config tests, docs, story, sprint status, and test summary artifacts.
- Code quality, security/redaction, standard-library-only dependency posture, and test quality were reviewed for the Story 4.1 source surface.

Reviewed evidence:
- `config/definition.go`, `config/kind.go`, `config/set.go`, `config/normalize.go`, `config/errors.go`, and `config/snapshot.go` implement reusable definitions, optional defaults, type expectations, sensitivity metadata, exact matching by default, opt-in normalization, typed setup errors, and default snapshots.
- `config/*_test.go` and `config/qa_e2e_test.go` cover metadata, every supported kind, defensive string-list defaults, exact lookup, normalized lookup/collisions, invalid normalized keys, immutable set derivation, default provenance, no-default versus zero defaults, not-found lookup, typed setup errors, redaction corpus behavior, and concurrent snapshot reuse.
- `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, and `config/doc.go` document only implemented Story 4.1 behavior and preserve later config-source scope boundaries.

Issues found:
- Critical: 0
- High: 0
- Medium: 0
- Low: 0

Validation run during review:
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./config -count=1`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test ./...`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go vet ./...`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go run ./tools/depgate`
- PASS: `git diff --check`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go test -race ./config ./... -count=1`
- PASS: `GOCACHE=/tmp/dib-go-build GOPATH=/tmp/dib-gopath go list -deps -f '{{if and (not .Standard) (not .Module.Main)}}{{.ImportPath}}{{end}}' ./config | sed '/^$/d'` produced no output.

### File List

- `_bmad-output/implementation-artifacts/4-1-register-config-keys-with-defaults-and-type-expectations.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `config/definition.go`
- `config/definition_test.go`
- `config/doc.go`
- `config/errors.go`
- `config/errors_test.go`
- `config/kind.go`
- `config/normalize.go`
- `config/qa_e2e_test.go`
- `config/set.go`
- `config/set_test.go`
- `config/snapshot.go`
- `config/snapshot_test.go`
- `docs/behavior-matrices.md`
- `docs/diagnostics-and-errors.md`
- `_bmad-output/implementation-artifacts/tests/test-summary.md`

### Change Log

- 2026-06-12: Added Story 4.1 config definition, set, normalization, typed setup error, and default snapshot APIs with external-package tests.
- 2026-06-12: Updated config package documentation, behavior matrix evidence, diagnostics guidance, story status, and sprint status for review.
- 2026-06-12: Added QA-generated Story 4.1 config public workflow and setup-error tests, then updated the test automation summary.
- 2026-06-12: Completed senior developer review; no code fixes required; story and sprint status moved to done.
