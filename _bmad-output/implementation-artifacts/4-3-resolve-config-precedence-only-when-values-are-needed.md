---
baseline_commit: 6655951
created: "2026-06-12"
---

# Story 4.3: Resolve Config Precedence Only When Values Are Needed

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Go CLI developer,
I want config values resolved by the documented precedence only when callers ask for them,
so that flags, env, JSON, explicit setters, and defaults compose without stale copied state.

## Requirements Trace

- FR12: config resolution applies the canonical V1 precedence: explicit setter, parsed flag, environment variable, JSON file, default. The winning source label is available on the resolved value.
- FR13: callers can bind parsed flag values to registered Config keys. `config/` accepts a narrow flag-derived value carrier without importing `flags/`. Only explicitly-set flags (not flag defaults) count as flag binding input.
- FR20: prove precedence and flag binding behavior through standard-library table tests and behavior-matrix evidence.
- NFR1: runtime packages, tests, examples, and tooling remain standard-library-only.
- NFR2/NFR5/NFR6: APIs use explicit instances and caller-supplied inputs; no package globals, live env reads, ambient files, or process exits.
- NFR3/NFR8: precedence resolution errors are inspectable and redact sensitive values.

## Acceptance Criteria

1. Given multiple sources provide a value for the same registered key, when a caller resolves that key, then the V1 precedence order is explicit setter, parsed flag, environment variable, JSON file, default, and the winning value reports the source label that supplied it.
2. Given parsed flags can bind to Config keys, when a bound Flag was explicitly set by CLI input, then its value can outrank lower-precedence sources through `flag binding` provenance, and a flag's configured default does not accidentally override env or JSON values unless explicitly configured.
3. Given config should not depend on command routing or parser internals, when flag binding is implemented, then `config/` accepts exported flag-derived snapshots or a small source interface, and it does not import `command/` or depend on unexported `flags/` implementation details.
4. Given same-source ties are valid, when repeated writes or loads occur within the same source, then last-writer-wins behavior is deterministic (already proven in 4.2 for each source snapshot), and binding collisions fail as setup errors rather than becoming runtime precedence ties.
5. Given precedence ambiguity is high risk, when verification runs, then tests cover every precedence pair, absent/default/explicit-zero distinctions, flag binding explicit-set behavior, same-source ties, binding collisions, provenance labels, redaction, and immutable snapshots.

## Tasks / Subtasks

- [x] Confirm preconditions and read all UPDATE files before editing (AC: 1-5)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 4 `in-progress` and Story 4.3 `ready-for-dev`.
  - [x] Confirm current `HEAD` is `6655951 feat(story-4.2): ingest explicit config sources`.
  - [x] Verify `go.mod` still contains only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or `go.sum`.
  - [x] Read every UPDATE file before editing: `config/doc.go`, `config/snapshot.go`, `config/errors.go`, `config/source.go`, `docs/behavior-matrices.md`, and `docs/diagnostics-and-errors.md`.
  - [x] Read `config/definition.go`, `config/kind.go`, and `config/set.go` to confirm the existing definition/kind/set APIs.
  - [x] Read `flags/snapshot.go` (exported `Snapshot`, `ValueState`, `ValueState.Explicit()`, `ValueState.Values()`) to understand the shape callers must translate before passing to `config/`.
  - [x] Do NOT import `flags/` or `command/` from `config/` at any point.

- [x] Design and implement the flag binding value carrier (AC: 2, 3, 4)
  - [x] Add a `FlagValue` struct (or equivalent small type) to `config/` that carries:
    - `ConfigKey string` — the registered config key this binding targets.
    - `ExplicitlySet bool` — true only when the flag was explicitly provided on the CLI (not its default).
    - `Value any` — the parsed flag value; only consumed when `ExplicitlySet` is true.
  - [x] Add `NewFlagSnapshot(set Set, values []FlagValue) (Snapshot, error)` that:
    - Validates each `FlagValue.ConfigKey` against the set (unknown key → typed `*SourceError` with `ErrUnknownSourceKey`).
    - Validates that no `ConfigKey` appears more than once in `values` (duplicate binding → typed `*SourceError` with a new sentinel `ErrDuplicateBinding` — chosen as distinct from `ErrInvalidSource` to allow callers to distinguish "duplicate key" from "invalid source setup").
    - For each binding where `ExplicitlySet` is true: validates the value kind against the registered definition, stores it with `flag binding` provenance (`SourceFlagBinding = "flag binding"`).
    - For each binding where `ExplicitlySet` is false: stores an absent source value (the flag's default must NOT be stored as a flag-binding value).
    - Returns a source snapshot (same `Snapshot` type as other sources).
  - [x] Add `SourceFlagBinding = "flag binding"` constant to `config/snapshot.go` alongside existing source label constants.
  - [x] Validate Go value kinds for flag binding using the existing `valueMatchesKind` helper (same path as explicit setter).

- [x] Implement lazy precedence resolution (AC: 1, 2, 4, 5)
  - [x] Add a `Resolve` function to a new `config/resolve.go` file.
    - Signature: `func Resolve(set Set, explicit, flag, env, jsonSrc Snapshot) Snapshot`
    - Each parameter is an independent source snapshot (the zero value is acceptable — it resolves as all-absent for that source tier).
    - The function layers sources lowest-to-highest (JSON, env, flag, explicit) over `set.DefaultSnapshot()` — highest wins.
    - If no higher-precedence source has a value, falls back to the set's default.
  - [x] The zero-value `Snapshot` must be usable as a "no-source" argument. Passing `Snapshot{}` for any tier must be equivalent to that tier having no values (not a panic).
  - [x] The returned snapshot is a new self-contained `Snapshot` with resolved values and their winning provenance labels — it holds no references into source snapshots.
  - [x] Expose `DefaultSnapshot` as the final fallback tier: if all higher-precedence snapshots are absent for a key, use the default from the definition set.
  - [x] Preserve immutability: the caller-supplied source snapshots must not be mutated.
  - [x] For same-source ties: each source snapshot is already last-writer-wins from Story 4.2. The `Resolve` function does not need to handle intra-source ordering — it just selects the first non-absent value per precedence tier.

- [x] Handle binding collisions as setup errors (AC: 4)
  - [x] If a caller passes `FlagValue` entries with duplicate `ConfigKey` values to `NewFlagSnapshot`, return a typed `*SourceError` before any values are stored.
  - [x] The duplicate-binding error must expose the colliding key via `(*SourceError).Key()`.
  - [x] Added `ErrDuplicateBinding` sentinel to `config/errors.go` alongside existing sentinels and documented in `docs/diagnostics-and-errors.md`.
  - [x] Do NOT implement cross-source binding collision detection — only intra-`NewFlagSnapshot` duplicates are setup errors.

- [x] Guard against flag default override of lower-precedence sources (AC: 2)
  - [x] Unit tests prove that a `FlagValue{ExplicitlySet: false}` binding does NOT inject any value into the flag snapshot (it becomes an absent source value for that tier).
  - [x] Tests verify that when a flag is not explicitly set, env and JSON values still win over the absent flag tier.
  - [x] Tests verify that when a flag IS explicitly set, its value wins over env, JSON, and default (correct precedence ranking).

- [x] Add typed source diagnostics for flag binding failures (AC: 4, 5)
  - [x] Reuse `*SourceError` for flag binding setup errors (unknown key, type mismatch, duplicate binding).
  - [x] Source label for flag binding errors is `"flag binding"` (i.e., `SourceFlagBinding`).
  - [x] Sensitive keys do not echo raw values in flag binding errors (redacted via `def.sensitive`).
  - [x] `errors.Is(err, ErrUnknownSourceKey)` is true for unknown config key in `NewFlagSnapshot`.
  - [x] `errors.Is(err, ErrSourceConversion)` is true for kind mismatch in flag binding.

- [x] Add zero-value `Snapshot` safety (AC: 1, 5)
  - [x] `Resolve(set, Snapshot{}, Snapshot{}, Snapshot{}, Snapshot{})` returns the same result as `set.DefaultSnapshot()` — only defaults apply.
  - [x] Passing a `Snapshot{}` (zero value) for any tier does not panic; it behaves as if that source has no values.
  - [x] Test `TestResolveZeroValueSnapshotsBehaviorEqualsDefaultSnapshot` proves this boundary condition.

- [x] Add executable tests (AC: 1-5)
  - [x] Added external-package tests using `package config_test` in new file `config/resolve_test.go`.
  - [x] Test every precedence pair in isolation (see `TestResolvePrecedenceAdjacentPairs`).
  - [x] Test absent/zero distinction: `TestResolveAbsentKeyWithNoDefaultReturnsNoValue`, `TestResolveExplicitZeroValueBeatsLowerTierNonZeroValue`, `TestResolveEmptyEnvStringCounts`.
  - [x] Test flag binding: `TestResolveFlagDefaultDoesNotOverrideEnv`, `TestResolveFlagExplicitBeatsEnvAndJSON`, `TestResolveFlagNotInFlagValuesDefersTolowerTiers`.
  - [x] Test binding collisions: `TestNewFlagSnapshotDuplicateBinding`, `TestNewFlagSnapshotNoPartialStateOnError`.
  - [x] Test provenance labels: `TestResolveProvenanceLabelsAreCanonical`, `TestSourceFlagBindingConstant`.
  - [x] Test redaction: `TestNewFlagSnapshotSensitiveRedaction`, `TestNewFlagSnapshotSensitiveBindingErrorDoesNotEchoCorpus`.
  - [x] Test immutability: `TestResolveSourceSnapshotImmutability`, `TestResolveReusableSourceSnapshots`.
  - [x] Test zero-value snapshot boundary: `TestResolveZeroValueSnapshotsBehaviorEqualsDefaultSnapshot`.
  - [x] All typed errors asserted with `errors.Is`/`errors.As` and accessor methods first; strings checked only for redaction.

- [x] Preserve Story 4.3 scope boundaries (AC: 1-5)
  - [x] Did NOT implement typed getters (Story 4.4 scope).
  - [x] Did NOT implement provenance source reports or rendered diagnostics (Story 4.5 scope).
  - [x] Created `docs/config-precedence.md` as the canonical V1 precedence reference (minimal).
  - [x] Did NOT implement compatibility docs, migration examples, struct decoding, live reload, YAML/TOML, root facade APIs, or global config registries.
  - [x] Did NOT import `flags/` or `command/` from `config/`.
  - [x] Did NOT add external runtime, test, assertion, fuzzing, or tooling dependencies.
  - [x] No source, tests, examples, fixtures, or names copied from Viper, pflag, Cobra, or other CLI/config projects.

- [x] Update documentation after implementation evidence exists (AC: 1-5)
  - [x] Updated `config/doc.go` for Story 4.3 flag binding and precedence resolution.
  - [x] Updated `docs/behavior-matrices.md` with config precedence row and exact executable test names.
  - [x] Updated `docs/diagnostics-and-errors.md` with `flag binding` source label, binding collision errors, and flag-default-override prevention semantics.
  - [x] Created `docs/config-precedence.md` as the canonical V1 precedence authority.
  - [x] Did not claim typed getters, source reports, migration support, compatibility completion, or release readiness.

- [x] Verify the story implementation (AC: 1-5)
  - [x] `go test ./config -count=1` → PASS (49 tests, 0 failures).
  - [x] `go test ./...` → PASS (all packages, 0 failures).
  - [x] `go vet ./...` → PASS (no issues).
  - [x] `go run ./tools/depgate` → PASS (standard-library only).
  - [x] `git diff --check` → PASS (no whitespace issues).
  - [x] `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` was created.
  - [x] `go test -race ./config ./... -count=1` → PASS (all tests pass under race detector).

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md` (Story 4.3 ACs are the primary spec).
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded previous story intelligence from `_bmad-output/implementation-artifacts/4-2-read-config-sources-through-explicit-boundaries.md`.
- No UX artifact discovered; Dib V1 has no browser UI or frontend surface.
- No Story 4.3 ATDD/test artifact discovered under `_bmad-output/test-artifacts/`.
- Current config source files read: `config/doc.go`, `config/definition.go`, `config/kind.go`, `config/set.go`, `config/normalize.go`, `config/snapshot.go`, `config/source.go`, `config/errors.go`, and tests.
- Loaded `flags/snapshot.go` to understand `ValueState.Explicit()` and `ValueState.Values()` as the caller-side surface for constructing `FlagValue` entries.

### Current Repository State

- Baseline commit at story creation: `6655951` (`feat(story-4.2): ingest explicit config sources`).
- Root `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`; no `go.sum` exists.
- `sprint-status.yaml` has Epic 4 `in-progress`, Stories 4.1 and 4.2 `done`, and Story 4.3 moved to `ready-for-dev` by this create-story workflow.
- Existing unrelated BMAD installer/config, `.agents/`, `.codex/`, `.idea/`, and story-automator changes exist in the worktree. Do not revert or normalize them.

### Current Code Context

- `config/snapshot.go` defines `Snapshot`, `Value`, `Source`, and the four existing source constants: `SourceDefault`, `SourceExplicit`, `SourceEnv`, `SourceJSON`. Add `SourceFlagBinding = "flag binding"` here.
- `config/source.go` contains `NewExplicitSnapshot`, `NewEnvSnapshot`, `LoadJSON`, `LoadJSONFile`, and all conversion helpers. Story 4.3 adds `NewFlagSnapshot` here (or in a new `config/flag.go` if preferred for cohesion).
- `config/errors.go` defines both `*DefinitionError` (setup) and `*SourceError` (source ingestion). Story 4.3 reuses `*SourceError` for flag binding errors. If adding `ErrDuplicateBinding`, add it here alongside existing sentinels.
- `config/set.go` defines `Set`, `NewSet`, `DefaultSnapshot`. The resolver in Story 4.3 will call `set.DefaultSnapshot()` as the final fallback tier.
- The `newAbsentSourceValue` and `newSourceValue` helpers in `config/snapshot.go` are already the correct primitives for building source snapshots.
- `newSourceSnapshot(set Set)` in `config/source.go` creates an all-absent snapshot from a set — Story 4.3's `NewFlagSnapshot` should use it as the starting point.
- `lookupSourceDefinition(set, key)` in `config/source.go` is the validated key-lookup helper — reuse it in `NewFlagSnapshot`.
- `valueMatchesKind(kind, value)` in `config/definition.go` is the kind-validation helper — reuse it for flag value type checking (same path as `explicitValue`).

### Architecture Guardrails

- V1 precedence order is fixed: explicit setter > parsed flag > environment variable > JSON file > default. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Flow`]
- `config/` accepts exported flag-derived snapshots or a small source interface. A direct `config → flags` import is deferred unless proven necessary; this story must NOT introduce it. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- `flags/` must remain fully usable without `command/` or `config/`. `command/` must not depend on `config/`. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- Definitions and snapshots are reusable values. Derived definitions return new values; per-run snapshots must not write back to definitions. Caller-observably immutable means no mutable aliases. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Architecture`]
- Source labels for V1 are fixed and closed: `default`, `explicit setter`, `flag binding`, `env`, and `JSON`. [Source: `docs/diagnostics-and-errors.md#Source-Labels`]
- Sensitive values must be redacted in errors, debug strings, diagnostics, and source reports. The fixed fake corpus is `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Fixture-Patterns`]

### Previous Story Intelligence

- Story 4.2 established the three independent source snapshot APIs and all conversion helpers. Use `newSourceSnapshot`, `lookupSourceDefinition`, `explicitValue`, and `convertStringValue` patterns for `NewFlagSnapshot` rather than introducing parallel helpers.
- Story 4.2 established external-package tests with `package config_test`, table-driven tests, `errors.Is`/`errors.As` assertions first, and docs updated only after executable evidence exists.
- Story 4.2's `*SourceError` pattern (category, key, source label, redaction, cause) is the correct diagnostic model for flag binding errors.
- Story 4.2 confirmed: do not store partial state when setup validation fails. `NewFlagSnapshot` should return a zero `Snapshot` and typed error on any setup failure.
- Story 4.1 note: config definition set uses index-based position in `set.definitions`. The `Snapshot.values` slice is ordered by the same index. When implementing `Resolve`, iterate the set definitions in registration order for determinism.
- Story 4.1 review: no open defects. Do not refactor its API shape.

### Git Intelligence

- Recent commits:
  - `6655951 feat(story-4.2): ingest explicit config sources`
  - `0e039b5 feat(story-4.1): register config definitions`
  - `ad1af8a docs: add epic 3 retrospective`
  - `9112e13 feat(story-3.5): preserve execution boundaries`
  - `510b733 feat(story-3.4): render command help`
- Story 4.2 added `config/source.go`, `config/env_test.go`, `config/json_test.go`, `config/source_test.go`, `config/qa_e2e_test.go`, JSON fixtures, and updated `config/snapshot.go`, `config/errors.go`, `config/doc.go`, and docs.
- Existing pattern: focused package tests first (RED), minimal implementation (GREEN), docs updated with exact test names, repository-wide test/vet/depgate/diff checks before completing.

### Flag Binding API Design

The caller-side flow (for `flags.Snapshot` → `config.NewFlagSnapshot`) is:

```go
// Caller (not inside config/) translates flags.Snapshot to []config.FlagValue:
flagSnap, _ := flagSet.Parse(os.Args[1:])
bindings := []config.FlagValue{
    {ConfigKey: "log-level",  ExplicitlySet: vs.Explicit(), Value: vs.Values()[0]},  // if vs, ok := flagSnap.Lookup("log-level"); ok
    {ConfigKey: "output-dir", ExplicitlySet: false},                                  // flag not set
}
flagSrc, err := config.NewFlagSnapshot(configSet, bindings)
```

`FlagValue` does not include the flag name — only the config key matters for binding. The caller controls the mapping. This keeps `config/` free of flag-specific concepts.

### Resolver API Design

Recommended function signature for `config/resolve.go`:

```go
// Resolve returns a snapshot with each registered key resolved by the V1 precedence:
// explicit setter > flag binding > env > JSON > default.
// Passing a zero-value Snapshot{} for any tier is safe and equivalent to that tier having no values.
func Resolve(set Set, explicit, flag, env, jsonSrc Snapshot) Snapshot
```

This is preferred over a variadic `...Snapshot` (which would lose source-tier identity) or a `Resolver` builder (which adds indirection without benefit at V1 scope). The parameter names directly map to the documented precedence levels.

Implementation sketch for `Resolve`:
1. Start with `set.DefaultSnapshot()` as the base (provides default values and provenance).
2. Iterate `set.definitions` in registration order.
3. For each definition key: check JSON snapshot, env snapshot, flag snapshot, explicit snapshot — in reverse precedence order (lowest to highest), overwriting the base value when a higher-tier source has a non-absent value.
4. Return the resulting snapshot.

Or equivalently in forward precedence order: start with default base, layer JSON on top, then env, then flag, then explicit — highest-precedence source wins.

Either direction is correct; choose the implementation that is clearest and most testable.

### Latest Technical Information

- Go 1.26.4 is current stable as of 2026-06-12. Keep `go 1.26` directive; no `toolchain` directive.
- `encoding/json` and `strconv` are used in `config/source.go`; no new standard-library imports are expected for Story 4.3 beyond what is already present.
- No external test or assertion library; use only `testing` package and `errors.Is`/`errors.As`.

### Testing Standards

- All tests in `package config_test` (external package style), consistent with Stories 4.1 and 4.2.
- Table-driven tests preferred; test names describe observable behavior (e.g., `TestResolveExplicitBeatsFlag`, `TestResolveFlagDefaultDoesNotOverrideEnv`).
- Assert typed errors with `errors.Is`/`errors.As` and accessor methods first; only check error strings for redaction or human-facing wording.
- Use `config.FlagValue` entries built from literal values, not from a real `flags.Snapshot`, to keep tests within `config_test` and avoid importing `flags/`.
- Keep fixtures deterministic and clean-room. Use only the fake sensitive corpus.
- If a test verifies concurrent snapshot reuse, use `sync.WaitGroup` and goroutines (standard library only).

### Security And Quality Checks

- Use the architecture-owned fake sensitive-value corpus exactly: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.
- Do not echo sensitive raw values in errors, debug strings, or diagnostics for flag binding failures.
- Do not add global source registries, default resolvers, ambient env reads, or process exits.
- Keep `config/` independently usable without importing `flags/` or `command/`.

### Project Structure Notes

Expected Story 4.3 source files:
- ADD `config/resolve.go` — `Resolve` function, precedence logic
- UPDATE `config/snapshot.go` — add `SourceFlagBinding = "flag binding"` constant
- UPDATE or ADD `config/source.go` or `config/flag.go` — `FlagValue` type, `NewFlagSnapshot`
- UPDATE `config/errors.go` — add `ErrDuplicateBinding` sentinel if chosen; otherwise reuse `ErrInvalidSource` for binding collisions (document the decision)
- ADD `config/resolve_test.go` — precedence, flag binding, zero-value boundary tests
- ADD or UPDATE `config/flag_test.go` or `config/source_test.go` — `NewFlagSnapshot` specific tests (may be consolidated into `resolve_test.go`)
- UPDATE `config/doc.go` — document flag binding and resolution
- UPDATE `docs/behavior-matrices.md` — add config precedence rows
- UPDATE `docs/diagnostics-and-errors.md` — add `flag binding` source label and binding collision errors
- ADD `docs/config-precedence.md` — canonical V1 precedence authority (minimal, if it does not already exist)
- UPDATE `_bmad-output/implementation-artifacts/sprint-status.yaml` — Story 4.3 → `done`

Do not create: `examples/`, `docs/compatibility.md`, `docs/migration/`, typed getter files, source report files, or any `command/` or `flags/` changes.

### Files To Read Before Editing

- `config/doc.go` — current package docs and source-boundary claims.
- `config/definition.go` — kind validation, `valueMatchesKind`, sensitive metadata.
- `config/kind.go` — supported kind vocabulary.
- `config/set.go` — immutable definition set, `DefaultSnapshot`.
- `config/normalize.go` — config normalizer contract.
- `config/snapshot.go` — `Snapshot`, `Value`, `Source`, existing source constants, `newAbsentSourceValue`, `newSourceValue`.
- `config/source.go` — `newSourceSnapshot`, `lookupSourceDefinition`, `explicitValue`, `valueMatchesKind` usage, `*_test.go` for style reference.
- `config/errors.go` — `*SourceError` pattern; where to add new sentinels.
- `config/*_test.go` — external-package test style and redaction corpus assertion patterns.
- `flags/snapshot.go` — read-only for understanding `ValueState.Explicit()` and `Values()`; do not import.
- `docs/behavior-matrices.md` — add precedence evidence rows only after tests exist.
- `docs/diagnostics-and-errors.md` — add Story 4.3 source-label and binding-collision entries.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md#Story-43-Resolve-Config-Precedence-Only-When-Values-Are-Needed`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-12-Resolve-values-by-precedence`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-13-Bind-flags-to-Config-keys`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Flow`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Architecture`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Format-Patterns`]
- [Source: `_bmad-output/implementation-artifacts/4-2-read-config-sources-through-explicit-boundaries.md`]
- [Source: `config/snapshot.go`]
- [Source: `config/source.go`]
- [Source: `config/errors.go`]
- [Source: `config/set.go`]
- [Source: `flags/snapshot.go`]
- [Source: `docs/behavior-matrices.md`]
- [Source: `docs/diagnostics-and-errors.md`]

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

- 2026-06-12: Create-story workflow executed. No activation prepend/append steps configured.
- 2026-06-12: Loaded sprint status, epics, architecture, and Story 4.2 intelligence in parallel.
- 2026-06-12: Read current config source files: `snapshot.go`, `source.go`, `errors.go`, `set.go`, `definition.go`, `kind.go`.
- 2026-06-12: Read `flags/snapshot.go` to understand caller-side flag value shape without importing `flags/`.
- 2026-06-12: Confirmed baseline commit `6655951`, no `go.sum`, module-only `go.mod`, and Story 4.3 `backlog` status in sprint tracker.
- 2026-06-12: Story file created. Sprint status updated to `ready-for-dev`.

### Debug Log

- 2026-06-12: Executed Story 4.3 dev workflow. Read all required source files before editing.
- 2026-06-12: Confirmed HEAD=6655951, go.mod module-only, no go.sum, sprint status Epic 4 in-progress, Story 4.3 ready-for-dev.
- 2026-06-12: RED phase — wrote `config/resolve_test.go` referencing undefined types. Build failed as expected.
- 2026-06-12: GREEN phase — implemented `SourceFlagBinding` constant, `ErrDuplicateBinding` sentinel, `FlagValue` type + `NewFlagSnapshot` in `config/flag.go`, `Resolve` function in `config/resolve.go`.
- 2026-06-12: All tests pass: `go test ./config -count=1` (49 tests PASS), `go test ./...` (all packages PASS), `go vet ./...` (no issues), `go run ./tools/depgate` (PASS), `git diff --check` (clean), `go test -race ./... -count=1` (PASS).
- 2026-06-12: Documentation updated: `config/doc.go`, `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, new `docs/config-precedence.md`.

### Completion Notes List

- Create-story workflow context: Verified preconditions, source files, module state before implementation.
- Implemented `FlagValue` type and `NewFlagSnapshot` in `config/flag.go`. Flag defaults (ExplicitlySet=false) do NOT enter the flag binding tier — only explicitly-set flag values count.
- Implemented `Resolve` in `config/resolve.go` using index-aligned layering over `set.DefaultSnapshot()`. Sources applied lowest-to-highest: JSON, env, flag, explicit. Zero-value Snapshot{} for any tier is safe.
- Added `ErrDuplicateBinding` as a distinct sentinel (not `ErrInvalidSource`) to allow callers to distinguish duplicate flag bindings from invalid source setup.
- Duplicate detection uses canonical `def.key` as the dedup key, so normalized-key aliases (e.g. `log-level` and `log_level`) are correctly treated as duplicates.
- 49 config tests pass including 20 new Story 4.3 tests. No regressions in any package.
- `config/` does not import `flags/` or `command/`. All imports remain standard library only.

### File List

- `_bmad-output/implementation-artifacts/4-3-resolve-config-precedence-only-when-values-are-needed.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `config/snapshot.go` — added `SourceFlagBinding = "flag binding"` constant
- `config/errors.go` — added `ErrDuplicateBinding` sentinel
- `config/flag.go` — new file: `FlagValue` type, `NewFlagSnapshot`
- `config/resolve.go` — new file: `Resolve` function
- `config/resolve_test.go` — new file: 20 Story 4.3 tests
- `config/qa_e2e_test.go` — added `TestQAConfigResolutionWorkflowCoversFlagBindingAndFullPrecedence` and `TestQAConfigFlagBindingDiagnosticsCoverErrorCategories`
- `config/doc.go` — updated for flag binding and precedence resolution
- `docs/behavior-matrices.md` — added config precedence resolution row
- `docs/diagnostics-and-errors.md` — added Story 4.3 source-label and binding-collision entries
- `docs/config-precedence.md` — new file: canonical V1 precedence authority

### Change Log

- 2026-06-12: Story 4.3 — implement config flag binding (`FlagValue`, `NewFlagSnapshot`, `SourceFlagBinding`, `ErrDuplicateBinding`) and lazy precedence resolution (`Resolve`) with V1 precedence order: explicit setter > flag binding > env > JSON > default. 20 new tests added; all 49 config tests pass.
- 2026-06-12: Story 4.3 review (AI) — fixed 3 medium + 1 low issues: (1) added `config/qa_e2e_test.go` to File List; (2) added `TestQAConfigResolutionWorkflowCoversFlagBindingAndFullPrecedence` and `TestQAConfigFlagBindingDiagnosticsCoverErrorCategories` to behavior-matrices.md precedence row; (3) updated `docs/diagnostics-and-errors.md` Current Scope stale text; (4) corrected typo `...EchoCoprus` → `...EchoCorpus` in test name, story, and matrix. Status → done.
