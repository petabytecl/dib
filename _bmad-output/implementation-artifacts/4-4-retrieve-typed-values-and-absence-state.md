---
baseline_commit: c231957
created: "2026-06-12"
---

# Story 4.4: Retrieve Typed Values And Absence State

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Go CLI developer,
I want typed Config retrieval to distinguish absent values, zero values, and conversion failures,
so that config-dependent code can handle decisions explicitly.

## Requirements Trace

- FR11: config definitions are already registered with kinds and defaults (Story 4.1). Story 4.4 exposes typed getters that rely on the registered `Kind` to type-assert resolved `any` values from a `Snapshot`.
- FR16: callers can retrieve Config values through typed getters (`GetString`, `GetBool`, etc.) and existence checks (`IsSet`). This is the primary FR delivered by this story.
- FR20: prove typed getter behavior through standard-library table tests with typed-error inspection via `errors.Is`/`errors.As`, and add a row to `docs/behavior-matrices.md`.
- NFR1: runtime packages, tests, examples, and tooling remain standard-library-only.
- NFR2/NFR5/NFR6: typed getters accept explicit `Snapshot` instances; no package globals, live env reads, ambient process state, or `os.Exit`.
- NFR3/NFR8: getter errors are inspectable with `errors.Is`/`errors.As`; sensitive key identity is preserved but raw values are redacted in all error output.

## Acceptance Criteria

1. Given a registered key has a resolved value, when a caller retrieves it through a typed getter, then the getter returns the typed value and provenance, and conversion failures return typed errors without panics.
2. Given a key is absent, when a caller checks presence, then `IsSet` or equivalent distinguishes absent values from zero values, and missing unregistered keys return the documented not-found result.
3. Given a source provides an explicit zero value or empty env value, when the value is resolved, then the value counts as set, and tests distinguish explicit zero, empty string, default zero, and absent.
4. Given values may be sensitive, when typed retrieval fails or reports diagnostics, then diagnostics identify the key, source label, and failure category, and sensitive raw values remain redacted.
5. Given typed retrieval is a public caller contract, when verification runs, then tests cover string, bool, numeric, duration, and string-list typed values, conversion failures, absent/default/zero states, source labels, and `errors.Is`/`errors.As` inspection.

## Tasks / Subtasks

- [x] Confirm preconditions and read all UPDATE files before editing (AC: 1-5)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 4 `in-progress` and Story 4.4 `ready-for-dev`.
  - [x] Confirm HEAD is `c231957` (`feat(story-4.3): resolve config precedence`).
  - [x] Verify `go.mod` still contains only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or `go.sum`.
  - [x] Read `config/snapshot.go` — `Snapshot`, `Value`, source constants, `Lookup` method, `newDefaultValue`, `newAbsentSourceValue`, `newSourceValue`, `clonePublicValue`.
  - [x] Read `config/errors.go` — existing `DefinitionError`, `SourceError` types and all sentinels.
  - [x] Read `config/kind.go` — `Kind` vocabulary (`KindString` through `KindStringList`).
  - [x] Read `config/definition.go` — `Definition` struct, `valueMatchesKind` helper, `clonePublicValue`, `cloneStringSlice`.
  - [x] Read `config/set.go` — `Set.DefaultSnapshot`, `Set.Lookup`.
  - [x] Read `config/resolve.go` — `Resolve` function for context on what a resolved snapshot contains.
  - [x] Read `config/*_test.go` files to confirm external-package test style (`package config_test`).
  - [x] Do NOT implement provenance source reports or rendered config diagnostics (Story 4.5 scope).

- [x] Add new getter error sentinels and `*GetError` type to `config/errors.go` (AC: 1, 2, 4, 5)
  - [x] Add three new sentinels alongside existing sentinels:
    - `ErrKeyNotFound    = errors.New("config key not found")` — requested key is not registered (or invalid)
    - `ErrKeyAbsent      = errors.New("config key has no value")` — registered key with no resolved value from any source and no default
    - `ErrGetConversion  = errors.New("config value type mismatch")` — registered key resolved but the caller requested a different kind
  - [x] Add `GetError` struct to `config/errors.go`:
    ```
    key         string   — the requested config key
    kind        Kind     — the actual registered kind (use KindString as zero for not-found)
    wantKind    Kind     — the kind the caller's getter requested (e.g. KindBool for GetBool)
    sourceLabel string   — provenance label from the resolved value (empty when absent or not found)
    redacted    bool     — from the definition (false when key not found)
    category    error    — one of ErrKeyNotFound, ErrKeyAbsent, ErrGetConversion
    ```
  - [x] Implement `GetError.Error() string`.
  - [x] Implement `GetError.Unwrap() error` returning `category`.
  - [x] Implement `GetError.Is(target error) bool` returning `target == category`.
  - [x] Add exported accessors: `Key() string`, `Kind() Kind`, `WantKind() Kind`, `SourceLabel() string`, `Redacted() bool`, `Category() error`.
  - [x] Add a package-level `newGetError` constructor following the same defensive-nil pattern as `newSourceError` and `newDefinitionError`.

- [x] Implement `IsSet` on `Snapshot` (AC: 2, 3)
  - [x] Add `IsSet(key string) bool` as a method on `Snapshot` in `config/getter.go`.
  - [x] Logic: Lookup → false if unregistered; false if !hasValue; true otherwise.
  - [x] `IsSet` returns `true` even for explicit zero values or empty string defaults.
  - [x] Do NOT add `MustGet`, panic helpers, or global getter functions.

- [x] Implement typed getters on `Snapshot` in `config/getter.go` (AC: 1, 3, 4, 5)
  - [x] Create `config/getter.go` (new file) with all typed getters as methods on `Snapshot`.
  - [x] Implemented generic `getTyped[T any]` helper with lookup, absence check, and type assertion.
  - [x] `GetStringList` returns a defensive copy: `append([]string(nil), slice...)`.
  - [x] Include `import "time"` in `config/getter.go`.

- [x] Add executable tests in `config/getter_test.go` (AC: 1-5)
  - [x] Create `config/getter_test.go` with `package config_test`.
  - [x] All 28 unit tests pass covering all required behaviors and test names.
  - [x] Added `TestQAConfigTypedGettersWorkflowCoversAllKinds`, `TestQAConfigTypedGettersResolvedPresenceStates`, and `TestQAConfigGetterDiagnosticsCoverAbsenceAndKindMismatch` to `config/qa_e2e_test.go`.

- [x] Preserve Story 4.4 scope boundaries (AC: 1-5)
  - [x] No provenance source reports, String() rendering, struct tag decoding, live reload, YAML/TOML, root facade APIs, or global state.
  - [x] No `flags/` or `command/` imports from `config/`.
  - [x] No external dependencies added.
  - [x] No source copied from Viper, pflag, Cobra, or other projects.

- [x] Update documentation after implementation evidence exists (AC: 1-5)
  - [x] Updated `config/doc.go` — documented `IsSet`, typed getters, new error sentinels, and `*GetError`.
  - [x] Updated `docs/behavior-matrices.md` — added "Typed config retrieval" row with exact executable test names.
  - [x] Updated `docs/diagnostics-and-errors.md` — added `ErrKeyNotFound`, `ErrKeyAbsent`, `ErrGetConversion`, and `*GetError` entries.

- [x] Verify the story implementation (AC: 1-5)
  - [x] `go test ./config -count=1` → PASS
  - [x] `go test ./...` → PASS (all packages, 0 failures)
  - [x] `go vet ./...` → PASS (no issues)
  - [x] `go run ./tools/depgate` → PASS (standard-library only)
  - [x] `git diff --check` → PASS (no whitespace issues)
  - [x] `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` exists
  - [x] `go test -race ./config ./... -count=1` → PASS

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md` (Story 4.4 ACs are the primary spec).
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded previous story intelligence from `_bmad-output/implementation-artifacts/4-3-resolve-config-precedence-only-when-values-are-needed.md`.
- No UX artifact discovered; Dib V1 has no browser UI or frontend surface.
- Current config source files read: `config/snapshot.go`, `config/errors.go`, `config/kind.go`, `config/definition.go`, `config/set.go`, `config/resolve.go`, `config/source.go`, `config/flag.go`.

### Current Repository State

- Baseline commit at story creation: `c231957` (`feat(story-4.3): resolve config precedence`).
- Root `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`; no `go.sum` exists.
- `sprint-status.yaml` has Epic 4 `in-progress`, Stories 4.1–4.3 `done`, and Story 4.4 moved to `ready-for-dev` by this create-story workflow.
- Story 4.3 completion notes: explicitly called out "Did NOT implement typed getters (Story 4.4 scope)."

### Current Code Context

**`config/snapshot.go`** — core Snapshot/Value model:
- `Snapshot` holds `values []Value`, `byKey map[string]int`, `normalizer NameNormalizer`.
- `Snapshot.Lookup(key string) (Value, bool)` — returns `(Value{}, false)` for unregistered/invalid keys; returns `(Value, true)` for registered keys regardless of whether `hasValue` is set.
- `Value.Value() (any, bool)` — returns raw typed `any` and `hasValue` flag. The `any` is a Go-native typed value (e.g. `string`, `bool`, `int`, `time.Duration`), NOT a string.
- `Value.Provenance() string` — returns source label or `""` for absent.
- `Value.Definition() (Definition, bool)` — exposes the registered definition (kind, sensitive, usage, etc.).
- `Value.Source() Source` — returns provenance metadata.
- Values stored in snapshot are already strongly typed from ingestion — no string-to-type conversion needed at getter time.

**`config/errors.go`** — current sentinels and error types:
- Existing sentinels: `ErrInvalidDefinition`, `ErrDuplicateKey`, `ErrDuplicateNormalizedKey`, `ErrInvalidDefault`, `ErrInvalidSource`, `ErrUnknownSourceKey`, `ErrDuplicateBinding`, `ErrSourceRead`, `ErrJSONDecode`, `ErrSourceConversion`.
- ADD Story 4.4 sentinels here: `ErrKeyNotFound`, `ErrKeyAbsent`, `ErrGetConversion`.
- ADD `*GetError` type and constructor here following the same pattern as `*SourceError`.

**`config/kind.go`** — Kind values to cover in typed getters:
```
KindString   → string
KindBool     → bool
KindInt      → int
KindInt64    → int64
KindUint     → uint
KindUint64   → uint64
KindFloat64  → float64
KindDuration → time.Duration
KindStringList → []string
```

**`config/definition.go`** — `valueMatchesKind` is internal; do not reuse it in getters. Each getter just type-asserts with `ok` idiom. `cloneStringSlice` helper is available (unexported) for defensive `[]string` copies.

**`config/source.go`** — pattern reference: `newSourceSnapshot`, `lookupSourceDefinition`, `explicitValue`. The getter implementation is simpler — no string conversion, just type-assertion.

**`config/flag.go`** (added in 4.3) — defines `FlagValue` and `NewFlagSnapshot`. Pattern reference for the `ExplicitlySet` / `ExplicitlyNotSet` distinction.

**`config/resolve.go`** — `Resolve(set, explicit, flag, env, jsonSrc Snapshot) Snapshot`. The returned snapshot is what callers will pass typed getters on.

### Architecture Guardrails

- V1 typed getters must NOT call `os.Exit`, write to stdout/stderr, read from process env, or use package-global state. [Source: `_bmad-output/planning-artifacts/architecture.md#API-Communication-Patterns`]
- Public errors must support `errors.Is`/`errors.As`. Error strings are diagnostics, not the programmatic contract. [Source: `_bmad-output/planning-artifacts/architecture.md#Error-Handling-Patterns`]
- Sensitive values must be redacted in all error output. The fake corpus: `dib_fake_secret_value`, `dib_fake_password_value`, `dib_fake_token_value`. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Fixture-Patterns`]
- `config/` must NOT import `flags/` or `command/`. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- Definitions and snapshots are reusable, immutable values. Typed getters must return defensive copies of slice values. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Architecture`]

### Previous Story Intelligence

- Story 4.3 explicitly deferred typed getters as out of scope. `config/resolve.go` and `config/flag.go` provide the resolved snapshot that 4.4 adds getters on top of.
- Story 4.3 established the pattern for `*GetError`: use `newGetError` constructor, nil-safe methods, `Unwrap` returns category, custom `Is` checks category and unwrap chain.
- Story 4.3 confirmed: do not return partial results on failure — all error paths return zero value + typed error.
- Story 4.2 established: tests in `package config_test`, external-package style. Import `config` as the package under test.
- Story 4.2 confirmed: `cloneStringSlice` is unexported; call it within the `config` package, not from `config_test`. In tests, verify defensive copy behavior by modifying the returned slice and calling the getter again.
- Story 4.1 note: `Snapshot.values` is index-aligned with `Set.definitions`. The `Lookup` method already normalizes keys. Do not bypass `Lookup` in getter implementations.
- Story 4.1 review: no open defects. Do not refactor its API shape.

### Git Intelligence

- Recent commits:
  - `c231957 feat(story-4.3): resolve config precedence`
  - `6655951 feat(story-4.2): ingest explicit config sources`
  - `0e039b5 feat(story-4.1): register config definitions`
  - `ad1af8a docs: add epic 3 retrospective`
  - `9112e13 feat(story-3.5): preserve execution boundaries`
- Story 4.3 added: `config/flag.go`, `config/resolve.go`, `config/resolve_test.go`, updated `config/snapshot.go`, `config/errors.go`, `config/doc.go`, `config/qa_e2e_test.go`, and docs.
- Established pattern: RED phase (write tests referencing undefined types) → GREEN phase (implement) → docs updated with exact test names.
- Tests added last so doc names are based on final implemented test names.

### Typed Getter API Design

```go
// In config/getter.go:

// IsSet reports whether key has a resolved value, including defaults.
// Returns false for unregistered keys and registered keys with no value from any source.
func (s Snapshot) IsSet(key string) bool

// GetString returns the resolved string value for key.
// Returns ("", *GetError) if key is unregistered (ErrKeyNotFound),
// has no value (ErrKeyAbsent), or is not KindString (ErrGetConversion).
func (s Snapshot) GetString(key string) (string, error)

// GetBool returns the resolved bool value for key.
func (s Snapshot) GetBool(key string) (bool, error)

// GetInt returns the resolved int value for key.
func (s Snapshot) GetInt(key string) (int, error)

// GetInt64 returns the resolved int64 value for key.
func (s Snapshot) GetInt64(key string) (int64, error)

// GetUint returns the resolved uint value for key.
func (s Snapshot) GetUint(key string) (uint, error)

// GetUint64 returns the resolved uint64 value for key.
func (s Snapshot) GetUint64(key string) (uint64, error)

// GetFloat64 returns the resolved float64 value for key.
func (s Snapshot) GetFloat64(key string) (float64, error)

// GetDuration returns the resolved time.Duration value for key.
func (s Snapshot) GetDuration(key string) (time.Duration, error)

// GetStringList returns a defensive copy of the resolved []string value for key.
func (s Snapshot) GetStringList(key string) ([]string, error)
```

**Internal getter helper** (recommended for DRY implementation):
```go
// getTyped is an unexported helper shared by all typed getters.
// It handles the lookup, absence check, and type assertion.
// wantKind is the Kind the caller's getter expects.
// convert is a function that type-asserts raw any to T.
func getTyped[T any](s Snapshot, key string, wantKind Kind, convert func(any) (T, bool)) (T, error) {
    var zero T
    v, ok := s.Lookup(key)
    if !ok {
        return zero, &GetError{key: key, kind: KindString, wantKind: wantKind, category: ErrKeyNotFound}
    }
    def, _ := v.Definition()
    raw, hasValue := v.Value()
    if !hasValue {
        return zero, &GetError{key: key, kind: def.kind, wantKind: wantKind, redacted: def.sensitive, category: ErrKeyAbsent}
    }
    val, ok := convert(raw)
    if !ok {
        return zero, &GetError{key: key, kind: def.kind, wantKind: wantKind, sourceLabel: v.Provenance(), redacted: def.sensitive, category: ErrGetConversion}
    }
    return val, nil
}
```
Note: Go 1.26 supports type parameters (generics). This is standard-library-only and appropriate here. If preferred, a non-generic approach with per-type helper functions is equally correct.

### GetError Design

```go
// In config/errors.go (additions):

var (
    ErrKeyNotFound   = errors.New("config key not found")
    ErrKeyAbsent     = errors.New("config key has no value")
    ErrGetConversion = errors.New("config value type mismatch")
)

type GetError struct {
    key         string
    kind        Kind    // actual registered kind
    wantKind    Kind    // kind the getter requested
    sourceLabel string  // provenance label when value was present but wrong kind
    redacted    bool    // from definition.sensitive
    category    error   // ErrKeyNotFound | ErrKeyAbsent | ErrGetConversion
}

// Error() format examples:
//   "config: config key not found for \"log-level\""
//   "config: config key has no value for \"timeout\" as duration"
//   "config: config value type mismatch for \"verbose\" as bool from explicit setter: want string"
//   "config: config key has no value for \"api-key\" value redacted"
```

### IsSet Presence Semantics

| State | `hasValue` | `IsSet` result |
|---|---|---|
| Unregistered key | n/a (Lookup returns false) | `false` |
| Registered, no default, no source | `false` | `false` |
| Registered, has zero default (e.g. `Int("x", 0, "")`) | `true` | `true` |
| Registered, explicitly set to zero (e.g. `SetInt("x", 0)`) | `true` | `true` |
| Registered, set to explicit empty string | `true` | `true` |
| Registered, env lookup returns `("", true)` | `true` | `true` |
| Registered, set to any non-zero value | `true` | `true` |

The key insight: `hasValue` is set during ingestion based on **whether a source contributed a value**, NOT on whether the value is zero. `IsSet` reads this flag directly.

### Testing Standards

- All tests in `package config_test` (external package style), consistent with Stories 4.1–4.3.
- Use `config.NewSet(...)`, `config.NewExplicitSnapshot(...)`, `config.NewEnvSnapshot(...)`, `config.Resolve(...)`, etc. from the external test package.
- For `FlagValue` usage in tests, construct `config.FlagValue{ConfigKey: "...", ExplicitlySet: true, Value: ...}` directly — no `flags/` import needed.
- Assert typed errors with `errors.Is`/`errors.As` and accessor methods first.
- Only check error strings for redaction: `strings.Contains(err.Error(), "dib_fake_secret_value")` must be false.
- Use `append([]string(nil), got...)` in tests to make defensive copies for mutation tests.
- Keep tests focused and deterministic — no reliance on env variables, filesystem, or clock.

### Sensitive Value Redaction in Getters

For `ErrKeyAbsent` and `ErrGetConversion` on sensitive keys:
- `GetError.Redacted()` returns `true`.
- `GetError.Error()` appends `" value redacted"` and does NOT append the source label (the source label might appear, but the raw value never does — since getters don't echo values in error strings, only the metadata matters here).
- Tests must verify that none of `dib_fake_secret_value`, `dib_fake_password_value`, or `dib_fake_token_value` appear in `err.Error()` when the key is sensitive.

For `ErrKeyNotFound`, `redacted` is `false` because the key is not registered and no sensitivity metadata is available.

### Latest Technical Information

- Go 1.26.4 is current stable as of 2026-06-12. Go 1.18+ generics are available and standard-library-safe; use them if the generic `getTyped` helper reduces code without compromising readability.
- No new standard-library packages are expected — `time` is already used in `config/definition.go`.
- `errors.Is` + `errors.As` is the standard inspection contract used throughout the codebase.

### Project Structure Notes

Expected Story 4.4 source files:
- ADD `config/getter.go` — `IsSet`, typed getters (`GetString` through `GetStringList`)
- UPDATE `config/errors.go` — add `ErrKeyNotFound`, `ErrKeyAbsent`, `ErrGetConversion` sentinels; add `*GetError` type, constructor, and accessors
- ADD `config/getter_test.go` — all typed getter and `IsSet` tests
- UPDATE `config/qa_e2e_test.go` — add `TestQAConfigTypedGettersWorkflowCoversAllKinds` and `TestQAConfigGetterDiagnosticsCoverAbsenceAndKindMismatch`
- UPDATE `config/doc.go` — document `IsSet`, typed getters, new error sentinels
- UPDATE `docs/behavior-matrices.md` — add "Typed config retrieval" row with exact test names
- UPDATE `docs/diagnostics-and-errors.md` — add `ErrKeyNotFound`, `ErrKeyAbsent`, `ErrGetConversion`, `*GetError` entries
- UPDATE `_bmad-output/implementation-artifacts/sprint-status.yaml` — Story 4.4 → `done`

Do NOT create: `examples/`, `docs/compatibility.md`, `docs/migration/`, provenance source report files, rendered diagnostic files, struct-decoding files, or any `command/` or `flags/` changes.

### Files To Read Before Editing

- `config/snapshot.go` — `Snapshot`, `Value.Value()`, `Value.Definition()`, `Value.Provenance()`, `clonePublicValue`.
- `config/errors.go` — where to add new sentinels and `*GetError` type.
- `config/kind.go` — `Kind` vocabulary and `String()`.
- `config/definition.go` — `Definition.Sensitive()`, `cloneStringSlice` (for defensive copy in `GetStringList`).
- `config/set.go` — `DefaultSnapshot` (used in tests).
- `config/resolve.go` — `Resolve` function (used in QA tests).
- `config/source.go` — `NewExplicitSnapshot`, `NewEnvSnapshot`, `NewFlagSnapshot` (used in tests for source setup).
- `config/*_test.go` — external-package test style and redaction corpus assertion patterns.
- `docs/behavior-matrices.md` — add typed getter row only after tests exist.
- `docs/diagnostics-and-errors.md` — add Story 4.4 getter error entries.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md#Story-44-Retrieve-Typed-Values-And-Absence-State`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-16-Retrieve-typed-Config-values`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-11-Register-Config-keys`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#API-Communication-Patterns`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Error-Handling-Patterns`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Architecture`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Fixture-Patterns`]
- [Source: `_bmad-output/implementation-artifacts/4-3-resolve-config-precedence-only-when-values-are-needed.md`]
- [Source: `config/snapshot.go`]
- [Source: `config/errors.go`]
- [Source: `config/kind.go`]
- [Source: `config/definition.go`]
- [Source: `config/set.go`]
- [Source: `config/resolve.go`]
- [Source: `docs/behavior-matrices.md`]
- [Source: `docs/diagnostics-and-errors.md`]
- [Source: `docs/config-precedence.md`]

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

- 2026-06-12: Create-story workflow executed. No activation prepend/append steps configured.
- 2026-06-12: Loaded sprint status, epics, architecture, Story 4.3 intelligence in parallel.
- 2026-06-12: Read all current config source files: snapshot.go, errors.go, kind.go, definition.go, set.go, resolve.go, source.go, flag.go, doc.go.
- 2026-06-12: Confirmed baseline commit c231957, no go.sum, module-only go.mod, Story 4.4 backlog status in sprint tracker.
- 2026-06-12: Story file created. Sprint status updated to ready-for-dev.

### Completion Notes List

- Implemented `config/getter.go` with `IsSet` and 9 typed getters (`GetString` through `GetStringList`) using a package-level generic `getTyped[T any]` helper that handles lookup, absence, and type-assertion in one place.
- Added `ErrKeyNotFound`, `ErrKeyAbsent`, `ErrGetConversion` sentinels and `*GetError` type with full `errors.Is`/`errors.As` support to `config/errors.go`. `GetError.Error()` omits kind for `ErrKeyNotFound` (placeholder zero) and appends redaction/kind-mismatch context only when applicable.
- `GetStringList` returns a defensive copy using `append([]string(nil), val...)`. `IsSet` delegates to `Value.hasValue` directly via `v.Value()` — no new fields needed.
- 28 unit tests in `config/getter_test.go` and 3 QA e2e tests in `config/qa_e2e_test.go` cover all ACs: all 9 kinds, defensive copy, all 3 error sentinels, `errors.Is`/`errors.As` inspection, all 5 source labels, zero/empty value semantics, sensitive key redaction.
- All gates passed: `go test ./...`, `go vet ./...`, `go run ./tools/depgate`, `go test -race ./config ./... -count=1`.

### File List

- `config/getter.go` (added)
- `config/errors.go` (updated — added `ErrKeyNotFound`, `ErrKeyAbsent`, `ErrGetConversion`, `*GetError`)
- `config/getter_test.go` (added)
- `config/qa_e2e_test.go` (updated — added 2 QA tests)
- `config/doc.go` (updated — documented `IsSet`, typed getters, new sentinels)
- `docs/behavior-matrices.md` (updated — added "Typed config retrieval" row)
- `docs/diagnostics-and-errors.md` (updated — added Story 4.4 getter error entries)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (updated — Story 4.4 → done)

### Senior Developer Review (AI)

Reviewer: Codex on 2026-06-12

Outcome: Approved after automatic fixes.

Findings addressed:
- [Fixed][Medium] `getTyped` constructed `*GetError` values directly instead of using the required `newGetError` constructor path in `config/getter.go`.
- [Fixed][Medium] `getTyped` trusted the raw Go value assertion before enforcing the registered `Kind` contract. Getter mismatches now fail based on `Definition.kind` before returning a value.
- [Fixed][Low] Story completion notes undercounted QA coverage: `config/qa_e2e_test.go` added 3 typed getter QA tests, not 2, for 31 new tests total.

Validation:
- `GOCACHE=/tmp/dib-go-build go test ./config -count=1` → PASS
- `GOCACHE=/tmp/dib-go-build go test ./...` → PASS
- `GOCACHE=/tmp/dib-go-build go vet ./...` → PASS
- `GOCACHE=/tmp/dib-go-build go run ./tools/depgate` → PASS
- `git diff --check` → PASS
- `GOCACHE=/tmp/dib-go-build go test -race ./config ./... -count=1` → PASS

### Change Log

- 2026-06-12: Story 4.4 implemented. Added typed config getters (`GetString` through `GetStringList`), `IsSet`, and `*GetError` with three new sentinel categories. All 31 new tests pass; no regressions.
- 2026-06-12: Senior developer review completed. Tightened getter kind-contract enforcement, routed getter diagnostics through `newGetError`, corrected story QA test count, and marked story done after all review gates passed.
