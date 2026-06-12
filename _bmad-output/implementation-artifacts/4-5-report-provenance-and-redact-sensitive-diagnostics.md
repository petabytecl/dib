---
baseline_commit: b974995
created: "2026-06-12"
---

# Story 4.5: Report Provenance And Redact Sensitive Diagnostics

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a security-sensitive CLI developer,
I want config diagnostics and source reports to explain why a value won without exposing secrets,
so that audited tools remain debuggable and safe.

## Requirements Trace

- FR16: callers can inspect resolved config value provenance through source reports and typed retrieval diagnostics without relying on rendered strings.
- FR20: source reports, rendered diagnostics, failure categories, redaction, and deterministic ordering must be proven with standard-library tests.
- NFR1: runtime packages, tests, examples, and tooling remain standard-library-only.
- NFR3: public diagnostic behavior remains inspectable through typed errors and structured state; rendered text is human-facing only.
- NFR4: source reports and rendered diagnostics are deterministic enough for stable tests.
- NFR8: sensitive raw values never appear in errors, debug strings, rendered diagnostics, source reports, examples, or validation failures.

## Acceptance Criteria

1. Given a config value resolves successfully, when a caller asks for provenance or a source report, then the report identifies the key, winning source label, and relevant source metadata, and source labels use only `default`, `explicit setter`, `flag binding`, `env`, and `JSON`.
2. Given resolution attempts fail, when diagnostics are rendered or inspected, then diagnostics distinguish attempted source label from failure category where both apply, and conversion failure and source read failure are distinguishable.
3. Given a key is marked sensitive, when errors, debug strings, diagnostics, source reports, examples, or validation failures are produced, then fake sensitive values from the architecture corpus never appear, and the key and source remain identifiable enough for debugging.
4. Given provenance output must be deterministic, when tests compare reports, then ordering of keys, attempted sources, failures, and diagnostics is stable, and golden tests are limited to human-facing rendering while structured state is asserted separately.
5. Given config provenance is adoption evidence, when verification runs, then tests cover success reports, failure reports, redaction false positives/false negatives, deterministic rendering, typed errors, and standard-library-only examples.


- [x] Confirm preconditions and read all UPDATE files before editing (AC: 1-5)
  - [x] Verify `_bmad-output/implementation-artifacts/sprint-status.yaml` marks Epic 4 `in-progress` and Story 4.5 `ready-for-dev`.
  - [x] Confirm current `HEAD` is `b974995` (`feat(story-4.4): add typed config getters`) or intentionally account for newer user changes.
  - [x] Verify `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`, with no `require`, `replace`, `toolchain`, or `go.sum`.
  - [x] Read `config/snapshot.go` completely: `Snapshot`, `Value`, `Source`, `Value.Source()`, canonical source constants, and clone behavior.
  - [x] Read `config/errors.go` completely: `*DefinitionError`, `*SourceError`, `*GetError`, sentinels, accessors, redaction behavior, and current unwrap semantics.
  - [x] Read `config/source.go`, `config/flag.go`, `config/resolve.go`, and `config/getter.go` completely to understand all success and failure paths that reports must describe.
  - [x] Read `config/*_test.go`, especially `qa_e2e_test.go`, for external-package style and redaction assertions.
  - [x] Read `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, and `docs/config-precedence.md` before updating docs.
  - [x] Do NOT implement compatibility docs, migration examples, root facade APIs, global config registries, new config formats, struct decoding, live reload, or any `command/` or `flags/` changes.

- [x] Add structured source report support for resolved snapshots (AC: 1, 3, 4)
  - [x] Add a focused `config/report.go` file for provenance/source-report APIs.
  - [x] Reuse existing `Snapshot` and `Value.Source()` data; do not create a parallel source model or reread env/JSON/files.
  - [x] Expose a structured report entry type with accessor methods, not exported mutable fields. It should include:
    - key: canonical registered config key
    - kind: registered `Kind`
    - set/present state: whether the snapshot has a resolved value
    - source label: winning provenance label or empty when absent
    - source metadata: env name, JSON path, JSON reader label where applicable
    - redaction flag: whether the key is sensitive
  - [x] Add `Snapshot.SourceReport() []SourceReportEntry` or an equivalent method that returns entries in definition/snapshot order.
  - [x] Include registered absent keys in the report with empty source label and `set=false` so "why no value won" is also deterministic.
  - [x] Return defensive copies for slices or internal data; callers must not mutate snapshot state through the report.
  - [x] Source report entries must never expose raw values, including non-sensitive values. Value retrieval remains the typed getter surface.
  - [x] Validate every report source label against the closed vocabulary (`default`, `explicit setter`, `flag binding`, `env`, `JSON`), with empty label allowed only for absent registered keys.

- [x] Add rendered source report output on top of structured state (AC: 1, 3, 4)
  - [x] Add `Snapshot.WriteSourceReport(w io.Writer) error` or an equivalent explicit-writer API.
  - [x] Rendering must use only caller-supplied writers; no stdout/stderr, no process state, no terminal-width detection.
  - [x] Rendering order must match `SourceReport()` order.
  - [x] Render enough metadata for debugging: key, source label, kind, redaction status, env name, JSON path, or JSON reader label when present.
  - [x] Do not render raw values. For sensitive keys, render redaction status, not the raw value. For non-sensitive keys, still keep source reports value-free.
  - [x] Propagate writer errors directly.

- [x] Add structured config diagnostic inspection for existing errors (AC: 2, 3, 4)
  - [x] Add a small structured diagnostic type or function in `config/report.go` or `config/diagnostic.go`, for example `InspectDiagnostic(err error) (Diagnostic, bool)`.
  - [x] It must recognize existing `*DefinitionError`, `*SourceError`, and `*GetError` through `errors.As`.
  - [x] It must expose category, key, kind, wanted kind when applicable, attempted source label when applicable, env name, JSON path, JSON reader label, redaction flag, and whether a safe cause exists.
  - [x] It must preserve the distinction between `ErrSourceConversion`, `ErrSourceRead`, `ErrJSONDecode`, `ErrUnknownSourceKey`, `ErrDuplicateBinding`, `ErrKeyAbsent`, `ErrKeyNotFound`, and `ErrGetConversion`.
  - [x] For `*SourceError`, source label and category are both required in the diagnostic when available. This is the main "attempted source label vs failure category" requirement.
  - [x] Do not change existing sentinel names or break `errors.Is` / `errors.As` behavior.
  - [x] Do not expose raw sensitive values or unsafe causes for sensitive failures.

- [x] Add rendered diagnostic output on top of structured diagnostics (AC: 2, 3, 4)
  - [x] Add `WriteDiagnostic(w io.Writer, err error) error` or an equivalent explicit-writer API.
  - [x] Unknown/non-config errors may either return a clear unsupported diagnostic result from `InspectDiagnostic` or render a minimal generic message without pretending to classify it; choose the smaller API that fits existing style.
  - [x] Rendering must be deterministic and based on structured diagnostic fields.
  - [x] Golden/string tests may assert human-facing formatting, but programmatic tests must assert structured diagnostic fields and typed errors first.
  - [x] Propagate writer errors directly.

- [x] Add executable tests for source reports (AC: 1, 3, 4, 5)
  - [x] Add `config/report_test.go` using `package config_test`.
  - [x] Cover all winning source labels in one resolved snapshot: default, explicit setter, flag binding, env, and JSON.
  - [x] Cover absent registered keys with stable empty source label and `set=false`.
  - [x] Cover env metadata (`EnvName`) and JSON metadata (`JSONPath`, `JSONReaderLabel`).
  - [x] Cover deterministic report order from registration/snapshot order.
  - [x] Cover defensive report values if any report field can carry mutable data.
  - [x] Cover sensitive keys with all fake corpus values absent from structured reports, rendered source reports, and any `fmt`/debug-style output the public types expose.
  - [x] Add at least one package-local runnable Go example test, such as `ExampleSnapshot_SourceReport` or `ExampleWriteDiagnostic`, using only the standard library. This satisfies the story's example evidence without creating Epic 5 migration examples or an `examples/` folder.

- [x] Add executable tests for diagnostic reports and rendering (AC: 2, 3, 4, 5)
  - [x] Assert `InspectDiagnostic` fields for:
    - source read failure from `LoadJSONFile` with `ErrSourceRead`, `SourceJSON`, and path metadata
    - JSON decode failure with `ErrJSONDecode`
    - source conversion failure with `ErrSourceConversion` and attempted source label
    - duplicate flag binding with `ErrDuplicateBinding` and `SourceFlagBinding`
    - getter absent/not-found/kind-mismatch with `ErrKeyAbsent`, `ErrKeyNotFound`, and `ErrGetConversion`
    - definition/default validation failure with `ErrInvalidDefault` and `SourceDefault` provenance where applicable
  - [x] Assert `errors.Is`/`errors.As` on the original errors still work after diagnostics are inspected or rendered.
  - [x] Assert rendered diagnostics are deterministic and use caller-supplied writers only.
  - [x] Assert writer errors are returned directly.
  - [x] Assert none of `dib_fake_secret_value`, `dib_fake_password_value`, or `dib_fake_token_value` appear in rendered diagnostics, report output, error strings, `fmt.Sprint(reportEntry)`, `fmt.Sprintf("%#v", reportEntry)`, or `fmt.Sprint(diagnostic)`.
  - [x] Include false-positive redaction coverage: ordinary non-sensitive values should not be marked redacted, but the report still does not print raw values.

- [x] Add QA/e2e coverage for the full config provenance workflow (AC: 1-5)
  - [x] Add `TestQAConfigProvenanceReportsExplainWinningSourcesWithoutValues` to `config/qa_e2e_test.go`.
  - [x] Add `TestQAConfigDiagnosticsDistinguishSourceAndCategory` to `config/qa_e2e_test.go`.
  - [x] Add `TestQAConfigProvenanceRenderingRedactsSensitiveCorpus` to `config/qa_e2e_test.go`.
  - [x] Keep tests deterministic, package-local, and standard-library-only.

- [x] Update package docs and project docs after executable evidence exists (AC: 1-5)
  - [x] Update `config/doc.go` to document structured source reports, rendered source reports, diagnostic inspection/rendering, and the fact that reports do not include raw values.
  - [x] Update `docs/behavior-matrices.md` with a "Config provenance and diagnostics" row naming the exact final test functions.
  - [x] Update `docs/diagnostics-and-errors.md` with Story 4.5 report and diagnostic APIs, redaction rules, and current scope.
  - [x] Update `docs/config-precedence.md` to remove "source reports deferred" language and point to the implemented provenance/report surface.
  - [x] Do not claim Epic 5 compatibility docs, migration examples, or release readiness.

- [x] Verify the story implementation (AC: 1-5)
  - [x] `GOCACHE=/tmp/dib-go-build go test ./config -count=1`
  - [x] `GOCACHE=/tmp/dib-go-build go test ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go vet ./...`
  - [x] `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
  - [x] `git diff --check`
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives and no `go.sum` exists.
  - [x] Run `GOCACHE=/tmp/dib-go-build go test -race ./config ./... -count=1` if time permits before review.


### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md` (Story 4.5 ACs are the primary spec).
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded PRD shards from `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/`.
- No UX artifact discovered; Dib V1 has no browser UI or frontend surface.
- No `project-context.md` file exists under the project root.
- Loaded previous story intelligence from `_bmad-output/implementation-artifacts/4-4-retrieve-typed-values-and-absence-state.md`.
- Current config source files read: `config/doc.go`, `config/snapshot.go`, `config/errors.go`, `config/definition.go`, `config/source.go`, `config/flag.go`, `config/resolve.go`, `config/getter.go`, `config/qa_e2e_test.go`, and related docs.

### Current Repository State

- Baseline commit at story creation: `b974995` (`feat(story-4.4): add typed config getters`).
- Root `go.mod` contains only `module github.com/petabytecl/dib` and `go 1.26`; no `go.sum` exists.
- `sprint-status.yaml` has Epic 4 `in-progress`, Stories 4.1-4.4 `done`, and Story 4.5 moved to `ready-for-dev` by this create-story workflow.
- Existing unrelated BMAD installer/config, `.agents/`, `.codex/`, `.idea/`, and story-automator changes exist in the worktree. Do not revert or normalize them.

### Current Code Context

**`config/snapshot.go`** — current provenance primitives:
- Defines the closed source label constants: `SourceDefault`, `SourceExplicit`, `SourceFlagBinding`, `SourceEnv`, `SourceJSON`.
- `Snapshot.Lookup(key)` returns a cloned `Value` for registered keys and false for unregistered/invalid keys.
- `Value.Value()` returns a cloned raw Go value and `hasValue`. Source reports should not expose this raw value.
- `Value.Provenance()` returns the winning source label or empty string for absent values.
- `Value.Source()` returns safe metadata: key, label, env name, JSON path, JSON reader label, and redaction flag.
- `Source` already has accessors: `Key()`, `Label()`, `EnvName()`, `JSONPath()`, `JSONReaderLabel()`, `Redacted()`.

**`config/errors.go`** — existing diagnostic model:
- `*DefinitionError` covers setup-time definition failures and exposes key, collision, normalized key, kind, default provenance, and redaction via accessors.
- `*SourceError` covers explicit/env/JSON/flag-binding source ingestion failures and exposes key, source label, env name, JSON path, JSON reader label, kind, redaction, category, and safe cause.
- `*GetError` covers typed retrieval failures and exposes key, actual kind, wanted kind, source label, redaction, and category.
- `SourceError.Unwrap()` returns the underlying cause; `SourceError.Is()` checks both category and cause. Preserve this behavior.
- `GetError.Unwrap()` returns its category and `GetError.Is()` matches only its category. Preserve this behavior.

**`config/source.go`** — source metadata already captured:
- Env values created by `NewEnvSnapshot` set `Source{label: SourceEnv, envName: envName}`.
- JSON reader/path values set `Source{label: SourceJSON, jsonPath: path, jsonReaderLabel: options.readerLabel}`.
- JSON object keys are sorted before ingestion, so multi-key diagnostics are deterministic.
- Sensitive raw values are already redacted in `*SourceError` construction.

**`config/flag.go`** — flag provenance:
- `NewFlagSnapshot` uses `SourceFlagBinding`.
- `ExplicitlySet: false` leaves the flag tier absent for that key; do not report it as a winning source.
- Duplicate normalized config keys in flag values return `ErrDuplicateBinding`.

**`config/resolve.go`** — report order:
- `Resolve` starts from `set.DefaultSnapshot()` and overlays JSON, env, flag, and explicit source snapshots by index.
- `Snapshot.values` order follows the set definition order. Use this order for deterministic source reports.

**`config/getter.go`** — retrieval diagnostics:
- `IsSet` depends on `Value.Value()`'s `hasValue`.
- Typed getters use `newGetError`, enforce registered `Kind` before returning values, and include source label for kind mismatches.
- Do not duplicate typed getter logic in source-report code.

### Architecture Guardrails

- Public errors must support `errors.Is` / `errors.As`. Error strings are diagnostics, not the programmatic contract. [Source: `_bmad-output/planning-artifacts/architecture.md#Error-Handling-Patterns`]
- Config source labels are fixed to `default`, `explicit setter`, `flag binding`, `env`, and `JSON`. [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Flow`]
- Sensitive values must be redacted in errors, `String` output, debug strings, rendered diagnostics, source reports, examples, and validation failures. [Source: `_bmad-output/planning-artifacts/architecture.md#Authentication-Security`]
- The fixed sensitive-value corpus is `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`. [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Fixture-Patterns`]
- `config/` must not import `command/` or depend on unexported `flags/` implementation details. [Source: `_bmad-output/planning-artifacts/architecture.md#Component-Boundaries`]
- Source reports and diagnostics must be deterministic and testable without stdout/stderr scraping. [Source: `_bmad-output/planning-artifacts/architecture.md#Contract-Assertion-Patterns`]

### Previous Story Intelligence

- Story 4.4 explicitly deferred provenance source reports and rendered config diagnostics to Story 4.5.
- Story 4.4 review fixed direct construction of `*GetError`; new code should use constructor paths and keep current error accessors stable.
- Story 4.4 review fixed getter behavior to check registered `Kind` before trusting the raw Go type. Story 4.5 diagnostics should report registered kind/source metadata, not infer behavior from raw values.
- Story 4.3 established `Resolve` and `NewFlagSnapshot`; do not refactor precedence or flag binding while adding reporting.
- Story 4.2 established deterministic JSON key ordering and source diagnostics. Do not reintroduce map-order nondeterminism in failure reporting.
- Story 4.1 review had no open defects. Do not refactor config definition APIs.

### Git Intelligence

- Recent commits:
  - `b974995 feat(story-4.4): add typed config getters`
  - `c231957 feat(story-4.3): resolve config precedence`
  - `6655951 feat(story-4.2): ingest explicit config sources`
  - `0e039b5 feat(story-4.1): register config definitions`
  - `ad1af8a docs: add epic 3 retrospective`
- Story 4.4 added `config/getter.go`, `config/getter_test.go`, updated `config/errors.go`, `config/doc.go`, `config/qa_e2e_test.go`, and docs.
- Established implementation rhythm: read all update files first, add focused tests, implement the smallest API surface, update docs after test names settle, then run repository gates.

### Recommended API Shape

Use structured state as the primary contract and rendering as a thin wrapper:

```go
// SourceReportEntry is a value-safe provenance row for one registered key.
type SourceReportEntry struct { /* unexported fields */ }

func (e SourceReportEntry) Key() string
func (e SourceReportEntry) Kind() Kind
func (e SourceReportEntry) IsSet() bool
func (e SourceReportEntry) SourceLabel() string
func (e SourceReportEntry) EnvName() string
func (e SourceReportEntry) JSONPath() string
func (e SourceReportEntry) JSONReaderLabel() string
func (e SourceReportEntry) Redacted() bool

func (s Snapshot) SourceReport() []SourceReportEntry
func (s Snapshot) WriteSourceReport(w io.Writer) error
```

For failure reporting, prefer an inspectable adapter over a new error hierarchy:

```go
type Diagnostic struct { /* unexported fields */ }

func InspectDiagnostic(err error) (Diagnostic, bool)
func WriteDiagnostic(w io.Writer, err error) error
```

`InspectDiagnostic` should read the existing typed errors with `errors.As`; it should not wrap, replace, or mutate the original error. If this exact naming feels awkward during implementation, keep the same contract: structured fields first, deterministic rendering second, no raw values.

### Redaction Rules For This Story

- Source reports must never include raw config values, even when a key is not sensitive.
- Rendered source reports and diagnostics must never include raw config values.
- Sensitive key diagnostics must preserve key name, source label, failure category, redaction status, and safe metadata. They must not expose unsafe causes or raw values.
- Public report/diagnostic types should avoid exported fields so `fmt.Sprintf("%#v", value)` does not dump accidental internals.
- Tests must check all three corpus strings against:
  - original error strings from current error types
  - structured source report entries via `fmt.Sprint` and `%#v`
  - rendered source reports
  - structured diagnostics via `fmt.Sprint` and `%#v`
  - rendered diagnostics

### Latest Technical Information

- Official Go downloads list Go 1.26.4 as the current featured stable release on 2026-06-12. The local module remains `go 1.26`; do not add a `toolchain` directive.
- Use the standard library only. `errors.Is`, `errors.As`, `fmt`, `io`, `strings`, `bytes`, and `testing` are sufficient for this story.
- No third-party assertion, golden-test, rendering, or redaction library is needed.

### Project Structure Notes

Expected Story 4.5 source files:
- ADD `config/report.go` or `config/diagnostic.go` — structured source report, diagnostic inspection, and rendering APIs.
- ADD `config/report_test.go` — source-report and diagnostic tests.
- UPDATE `config/qa_e2e_test.go` — full provenance workflow tests.
- UPDATE `config/doc.go` — document report/diagnostic APIs and value-free reports.
- UPDATE `docs/behavior-matrices.md` — add "Config provenance and diagnostics" row with exact test names.
- UPDATE `docs/diagnostics-and-errors.md` — add Story 4.5 diagnostic/report contract and current scope.
- UPDATE `docs/config-precedence.md` — mark source reports as implemented and link them to precedence.
- UPDATE `_bmad-output/implementation-artifacts/sprint-status.yaml` — Story 4.5 → `done` when implementation/review is complete.

Do NOT create: compatibility docs, migration examples, release-readiness docs beyond current docs, provenance log entries for copied/adapted material, an `examples/` folder, `/cmd`, new config format files, or any `command/` or `flags/` changes. A package-local `Example...` function in `config/report_test.go` is allowed and expected for AC5.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md#Story-45-Report-Provenance-And-Redact-Sensitive-Diagnostics`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-16-Retrieve-typed-Config-values`]
- [Source: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md#FR-20-Provide-behavior-test-matrices`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Data-Flow`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Error-Handling-Patterns`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Test-Fixture-Patterns`]
- [Source: `_bmad-output/implementation-artifacts/4-4-retrieve-typed-values-and-absence-state.md`]
- [Source: `config/snapshot.go`]
- [Source: `config/errors.go`]
- [Source: `config/source.go`]
- [Source: `config/flag.go`]
- [Source: `config/resolve.go`]
- [Source: `config/getter.go`]
- [Source: `docs/behavior-matrices.md`]
- [Source: `docs/diagnostics-and-errors.md`]
- [Source: `docs/config-precedence.md`]
- [Source: `https://go.dev/dl/`]
- [Source: `https://pkg.go.dev/errors`]
- [Source: `https://pkg.go.dev/testing`]

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- 2026-06-12: Create-story workflow executed. No activation prepend/append steps configured.
- 2026-06-12: Persistent fact glob `**/project-context.md` resolved to no files.
- 2026-06-12: Loaded sprint status, epics, architecture, PRD shards, Story 4.4 intelligence, config source files, tests, and docs.
- 2026-06-12: Confirmed baseline commit `b974995`, module-only `go.mod`, no `go.sum`, and Story 4.5 backlog status in sprint tracker before creation.
- 2026-06-12: Verified official Go sources for Go 1.26.4, `errors`, and `testing` references.
- 2026-06-12: Story file created and sprint status updated to ready-for-dev.
- 2026-06-12: Dev-story workflow activated with no prepend/append steps; `project-context.md` glob resolved to no files.
- 2026-06-12: Confirmed Story 4.5 sprint status ready-for-dev, HEAD `b974995`, module-only `go.mod`, and no `go.sum`; sprint status moved to in-progress.
- 2026-06-12: Read required config source, config tests, QA/e2e tests, and docs before implementation.
- 2026-06-12: Red phase confirmed with `GOCACHE=/tmp/dib-go-build go test ./config -count=1` failing on missing source report and diagnostic APIs.
- 2026-06-12: Validation passed: `GOCACHE=/tmp/dib-go-build go test ./config -count=1`; `GOCACHE=/tmp/dib-go-build go test ./...`; `GOCACHE=/tmp/dib-go-build go vet ./...`; `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`; `git diff --check`; module/no-go.sum check; `GOCACHE=/tmp/dib-go-build go test -race ./config ./... -count=1`.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Story scope is restricted to config provenance/source reports and rendered/structured diagnostics.
- Story explicitly prevents raw value exposure in source reports, including non-sensitive values.
- Story carries forward previous Epic 4 implementation lessons: no precedence refactor, no flag/command imports, no raw value diagnostics, and structured assertions before rendered-string tests.
- Added `Snapshot.SourceReport` and `Snapshot.WriteSourceReport` with deterministic value-free entries for resolved and absent registered keys.
- Added `Diagnostic`, `InspectDiagnostic`, and `WriteDiagnostic` for existing config error types while preserving sentinel names and `errors.Is` / `errors.As` behavior.
- Added read-only `DefinitionError.Redacted` and `DefinitionError.Category` accessors to support structured diagnostics without changing unwrap behavior.
- Added standard-library tests and QA/e2e coverage for winning source labels, metadata, absent keys, deterministic rendering, writer errors, typed diagnostics, source-vs-category distinction, redaction false positives, and sensitive corpus absence.
- Updated package and project docs for implemented provenance reports and diagnostics without adding Epic 5 compatibility or migration claims.

### File List

- `_bmad-output/implementation-artifacts/4-5-report-provenance-and-redact-sensitive-diagnostics.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `config/doc.go`
- `config/errors.go`
- `config/qa_e2e_test.go`
- `config/report.go`
- `config/report_test.go`
- `docs/behavior-matrices.md`
- `docs/config-precedence.md`
- `docs/diagnostics-and-errors.md`

### Senior Developer Review (AI)

Reviewer: GPT-5 Codex on 2026-06-12

Outcome: Approved. No critical, high, or medium implementation issues remained after adversarial review, so no code fixes were required.

Review coverage:
- Verified all acceptance criteria against `config/report.go`, `config/errors.go`, `config/report_test.go`, `config/qa_e2e_test.go`, `config/doc.go`, `docs/behavior-matrices.md`, `docs/config-precedence.md`, and `docs/diagnostics-and-errors.md`.
- Cross-checked the story File List against git changes. Application source/docs changes are documented; unrelated BMad installer, tool, IDE, and workspace metadata changes remain outside application review scope.
- Confirmed source labels remain limited to `default`, `explicit setter`, `flag binding`, `env`, and `JSON` for public source-report/diagnostic surfaces.
- Confirmed source reports and rendered diagnostics are value-free, deterministic, writer-bound, and preserve key/source/category metadata without exposing the sensitive corpus.
- Confirmed `InspectDiagnostic` recognizes `*DefinitionError`, `*SourceError`, and `*GetError` through `errors.As` while preserving existing `errors.Is` behavior on original errors.
- Per checklist doc lookup requirement, consulted official Go package docs for `errors`, `testing`, and `io` as the relevant standard-library references.

Validation:
- `GOCACHE=/tmp/dib-go-build go test ./config -count=1`
- `GOCACHE=/tmp/dib-go-build go test ./...`
- `GOCACHE=/tmp/dib-go-build go vet ./...`
- `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`
- `git diff --check`
- `test ! -e go.sum && ! grep -Eq '^(require|replace|toolchain)\b' go.mod`
- `GOCACHE=/tmp/dib-go-build go test -race ./config ./... -count=1`

### Change Log

- 2026-06-12: Implemented Story 4.5 config provenance source reports, structured diagnostic inspection, rendered diagnostics, tests, and docs; validated all required gates and optional race tests.
- 2026-06-12: Senior developer review approved Story 4.5; no code defects required auto-fix; story and sprint status moved to done.
