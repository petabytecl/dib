---
baseline_commit: 18253ee
created: "2026-06-12"
---

# Story 7.2: Compose Command Routing With Config Flag Bindings

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Go CLI developer,
I want `cli` to translate explicitly-set route flags into config flag bindings,
so that command flags and config precedence work together without manual `config.FlagValue` glue.

## Requirements Trace

- FR26: Compose CLI invocation, command routing, flag parsing, and config resolution through an optional `cli` package.
- FR13: Bound parsed flag values override lower-precedence config sources only when the flag was explicitly set.
- FR20: Provide behavior test matrices and package tests for public behavior.
- NFR1: Runtime packages must import only the Go standard library.
- NFR2: `cli` helpers are allowed only with caller-supplied inputs and no hidden singleton state.
- NFR3: Public error cases needed by callers must be inspectable without string matching.
- NFR5: Library APIs must not call `os.Exit`, mutate process streams, or read `os.Args` implicitly.
- NFR8: Diagnostics must identify bad keys, flags, and sources without dumping sensitive values.
- Architecture: `cli/` owns translation from exported `flags.Snapshot` values into `config.FlagValue`; `command/`, `flags/`, and `config/` remain independent. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural Boundaries`]
- Correct Course: Story 7.2 is GitHub issue #48 and exists to remove manual `flags.Snapshot` to `config.FlagValue` glue. [Source: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-06-12-cli-composition.md#Story 7.2`]

## Acceptance Criteria

1. Given a command route result has a flag snapshot, when `cli` applies explicit flag bindings, then only explicitly-set flags enter the config `flag binding` tier, and default flag values do not override env, JSON, or config defaults.

2. Given a binding maps a flag name to a config key, when the flag is absent from CLI input or present only as a definition default, then the resulting config flag source leaves that key absent for the flag tier.

3. Given a binding references an unknown flag or config key, when composition runs, then a typed `cli` error preserves the flag name, config key, category, and wrapped cause needed for `errors.Is` or `errors.As` inspection without echoing raw flag values.

4. Given the bridge composes exported package contracts, when implementation is reviewed, then `cli/` may import `command`, `flags`, and `config`, and `command/`, `flags/`, and `config/` do not import `cli`.

## Tasks / Subtasks

- [x] Confirm preconditions and current API surface (AC: 1-4)
  - [x] Verify sprint status marks Epic 7 `in-progress`, Story 7.1 `done`, and Story 7.2 `ready-for-dev`.
  - [x] Read current `cli/` files before editing: `cli/doc.go`, `cli/errors.go`, `cli/invocation.go`, `cli/invocation_test.go`, and `cli/invocation_qa_test.go`.
  - [x] Read route and snapshot APIs before editing: `command/result.go`, `command/route.go`, `command/flags.go`, `flags/snapshot.go`, `flags/definition.go`, `flags/set.go`, `config/flag.go`, `config/source.go`, `config/resolve.go`, and `config/errors.go`.
  - [x] Do not edit `command/`, `flags/`, or `config/` for this story unless an existing exported contract is proven insufficient; the expected implementation belongs in `cli/`.

- [x] Add the public flag-binding bridge in `cli` (AC: 1-4)
  - [x] Add `cli/binding.go` with a small value such as:
    - `type FlagBinding struct { FlagName string; ConfigKey string }`
    - optional constructor `BindFlag(flagName, configKey string) FlagBinding` if it improves call-site clarity.
  - [x] Add a public function such as `NewFlagSnapshot(set config.Set, route command.Result, bindings []FlagBinding) (config.Snapshot, error)`.
  - [x] Use only exported contracts:
    - `route.FlagSnapshot()` to get the parsed `flags.Snapshot`;
    - `snapshot.Lookup(binding.FlagName)` to get `flags.ValueState`;
    - `state.Explicit()` to decide whether a value enters the flag tier;
    - `state.Values()` for effective parsed values;
    - `config.NewFlagSnapshot(set, []config.FlagValue{...})` to create the config source snapshot.
  - [x] Preserve binding order when constructing `[]config.FlagValue`; let `config.NewFlagSnapshot` enforce duplicate config-key collisions.
  - [x] If `len(bindings) == 0`, return an empty flag-tier snapshot from `config.NewFlagSnapshot(set, nil)` without requiring a route flag snapshot.
  - [x] If a binding exists but the route has no flag snapshot, return a typed `cli` binding error rather than silently ignoring the binding.
  - [x] If `snapshot.Lookup(binding.FlagName)` returns false, return a typed `cli` binding error for an unknown route flag.

- [x] Translate flag values without changing precedence semantics (AC: 1-2)
  - [x] For `state.Explicit() == false`, pass `ExplicitlySet: false` and do not let the value override lower-precedence sources.
  - [x] For scalar values, use the effective parsed value from `state.Values()`. Current parser behavior stores one value for non-repeatable scalar flags and accumulated values for repeatable flags.
  - [x] For `flags.KindStringList` / `config.KindStringList`, convert `[]any{"a", "b"}` from `ValueState.Values()` into `[]string{"a", "b"}` before calling `config.NewFlagSnapshot`; do not pass `[]any` to config.
  - [x] Keep defensive-copy behavior for slice values. Do not expose slices returned by `Values()` or config snapshots as mutable shared state.
  - [x] Do not create a new precedence implementation in `cli`; resolution still belongs to `config.Resolve`.

- [x] Add typed binding errors in `cli` (AC: 3)
  - [x] Add `cli/binding_errors.go` or extend `cli/errors.go` with sentinels such as `ErrInvalidBinding` and `ErrUnknownFlagBinding`.
  - [x] Add an inspectable `BindingError` type with safe accessors such as `FlagName()`, `ConfigKey()`, `Category()`, and `Cause()`.
  - [x] Implement `Is(target error) bool` for `cli` binding categories and `Unwrap() error` or `Cause() error` for wrapped config causes, so callers can use `errors.Is(err, cli.ErrUnknownFlagBinding)` and still inspect config causes such as `config.ErrUnknownSourceKey`, `config.ErrDuplicateBinding`, or `config.ErrSourceConversion`.
  - [x] Error strings must not include raw parsed flag values. They may include flag names, config keys, categories, and value-free source labels.

- [x] Add bridge tests in `cli` (AC: 1-4)
  - [x] Add `cli/binding_test.go` using `package cli_test`.
  - [x] Cover a routed command with an explicit flag, env, JSON, and default config value: flag binding wins over env/JSON/default only when the route flag is explicitly set.
  - [x] Cover a routed command whose flag definition has a default but no CLI occurrence: the flag-tier snapshot leaves the config key absent, and `config.Resolve` lets env, JSON, or config default win.
  - [x] Cover an absent-but-known binding by including a flag definition on the route and omitting it from args; assert the config flag source has no value for that key.
  - [x] Cover unknown flag binding: route has a flag snapshot, binding references a missing flag name, `errors.Is(err, cli.ErrUnknownFlagBinding)` passes, and `errors.As(err, *cli.BindingError)` exposes flag/config context.
  - [x] Cover unknown config key: binding references a valid explicit route flag but a missing config key, the returned error wraps the config source failure and supports inspection without string matching.
  - [x] Cover `KindStringList` translation from repeated or string-list flag values into a config `[]string`.
  - [x] Cover sensitive values with the fake sensitive corpus (`dib_fake_secret_value`, `dib_fake_password_value`, `dib_fake_token_value`) and assert error strings/diagnostics do not leak raw values.
  - [x] Cover import boundaries with `go list` evidence or a focused test/script assertion: `cli` may import `command`, `flags`, and `config`; those packages must not import `cli`.

- [x] Preserve Story 7.2 scope boundaries (AC: 1-4)
  - [x] Do not add `cli.Resolve`, `Plan`, or a full route/config composition result in this story; Story 7.3 owns that.
  - [x] Do not execute callbacks, call `os.Exit`, read `os.Args`, read env, load JSON/files, mutate stdout/stderr, or hide errors behind rendered text.
  - [x] Do not add a root package facade, package-global default command/config/flag values, or singleton composition state.
  - [x] Do not duplicate command routing, flag parsing, or config resolution logic that already exists in `command`, `flags`, and `config`.
  - [x] Do not update README, examples, release docs, behavior matrices, coverage threshold policy, or tracker reconciliation in this story unless tests or package docs need a narrow API comment update; Story 7.4 owns public docs and evidence reconciliation.

- [x] Verify (AC: 1-4)
  - [x] `GOCACHE=/tmp/dib-go-cache go test ./cli ./config ./command ./flags`
  - [x] `GOCACHE=/tmp/dib-go-cache go vet ./...`
  - [x] `GOCACHE=/tmp/dib-go-cache go run ./tools/lint`
  - [x] `GOCACHE=/tmp/dib-go-cache go run ./tools/coverage`
  - [x] `GOCACHE=/tmp/dib-go-cache go run ./tools/depgate`
  - [x] `git diff --check`

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`; Epic 7 is `in-progress`, Story 7.1 is `done`, and Story 7.2 was `backlog` at story creation.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`; Story 7.2 acceptance criteria are defined under `## Epic 7: CLI Composition Ergonomics`.
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`; FR26 requires optional `cli` composition without process lifecycle ownership, and NFR2/NFR5 preserve caller-supplied inputs.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`; `cli/` may depend on `command`, `flags`, and `config`, but those packages must not depend on `cli`.
- Loaded Correct Course proposal: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-06-12-cli-composition.md`; Story 7.2 is GitHub issue #48.
- No UX artifact was discovered; Dib V1 has no browser UI.
- No `project-context.md` persistent fact file was found.

### Previous Story Intelligence

- Story 7.1 created the optional `cli` package and finished with status `done`.
- Current `cli` files:
  - `cli/doc.go`: documents `cli` as optional composition support, not a root facade or process-owning framework.
  - `cli/invocation.go`: `Invocation`, `FromOSArgs`, `FromArgs`, `Program`, and `Args`; args are defensively copied.
  - `cli/errors.go`: `ErrInvalidInvocation` plus inspectable `InvocationError` with value-free diagnostics.
  - `cli/invocation_test.go` and `cli/invocation_qa_test.go`: prove explicit argv ownership, defensive copies, typed errors, and no ambient `os.Args` reads.
- Story 7.1 explicitly deferred flag-to-config binding translation and `cli.Resolve`; do not collapse Story 7.2 and Story 7.3.
- Story 7.1 verification used `GOCACHE=/tmp/dib-go-cache` because the default Go cache may not be writable in this environment.

### Existing Code Patterns To Reuse

- Immutable public values use unexported fields plus accessors: `command.Result`, `command.Boundary`, `flags.Snapshot`, `config.Snapshot`, and `cli.Invocation`.
- Defensive copies use `append([]T(nil), values...)` or typed clone helpers before returning slice data.
- Typed error style:
  - Packages expose sentinel category errors and typed structs with safe context accessors.
  - `config.SourceError` has an `Is` method so `errors.Is(err, config.ErrUnknownSourceKey)` works even when there is a wrapped cause.
  - `cli.InvocationError` uses `Unwrap` for category inspection. `BindingError` needs both category inspection and wrapped config-cause inspection, so prefer an `Is` method for `cli` categories plus `Unwrap`/`Cause` for underlying config errors.
- Config flag binding already exists:
  - `config.FlagValue{ConfigKey, ExplicitlySet, Value}` is the bridge input.
  - `config.NewFlagSnapshot` ignores `ExplicitlySet=false` values and uses provenance label `config.SourceFlagBinding`.
  - `config.NewFlagSnapshot` validates unknown config keys, duplicate config-key bindings, and kind/value mismatches.
- The manual glue to replace lives in `examples/migration/config_precedence_migration_test.go`, where helper `mustConfigFlagSnapshot` calls `snapshot.Lookup`, `ValueState.Explicit`, `ValueState.Values`, and `config.NewFlagSnapshot`.

### Current API Details

- `command.Result.FlagSnapshot() (flags.Snapshot, bool)` returns the parsed flag snapshot for routes with command flags.
- `command.Result.Flags() (flags.Set, bool)` returns the composed flag definitions available to the matched command.
- `flags.Snapshot.Lookup(name string) (flags.ValueState, bool)` looks up by long flag name.
- `flags.ValueState.Explicit() bool` reports whether CLI input explicitly set the flag.
- `flags.ValueState.Values() []any` returns effective values:
  - scalar flags are represented as one value;
  - repeatable scalar flags accumulate multiple `any` values;
  - string-list flags are flattened into one `[]any` of strings.
- `flags.ValueState.Occurrences()` exposes source spellings for explicit occurrences. Story 7.2 does not need spellings for value translation, but tests may use them to confirm explicit-set behavior.
- `flags.ValueOccurrence.Definition()` returns the matched `flags.Definition`; for absent/default values, prefer `route.Flags().Lookup(binding.FlagName)` when kind metadata is needed.
- `config.Set.Lookup(key string) (config.Definition, bool)` exposes config key kind and sensitivity metadata.
- `config.NewFlagSnapshot(set, values)` returns a `config.Snapshot` for the flag-binding tier and must remain the only config ingestion boundary used by this story.
- `config.Resolve(set, explicit, flag, env, jsonSrc)` owns precedence. The order is explicit setter > flag binding > env > JSON > default.

### Suggested Implementation Shape

The story does not require these exact exported names if the implementation uses a clearer local convention, but the public surface should remain small and Story 7.3-ready:

```go
package cli

type FlagBinding struct {
	FlagName  string
	ConfigKey string
}

func BindFlag(flagName, configKey string) FlagBinding

func NewFlagSnapshot(set config.Set, route command.Result, bindings []FlagBinding) (config.Snapshot, error)
```

Implementation outline:

1. Return `config.NewFlagSnapshot(set, nil)` for zero bindings.
2. Read `route.FlagSnapshot()` and fail with a typed `BindingError` when bindings require a missing route flag snapshot.
3. For each binding, look up the route flag in the snapshot. Missing lookup is an unknown flag binding.
4. Translate `flags.ValueState` to `config.FlagValue`.
5. Call `config.NewFlagSnapshot` once with all translated values.
6. If config returns an error, wrap it in `BindingError` while preserving `errors.Is` / `errors.As` access to the config error.

Use a small helper to translate values:

- If `Explicit() == false`, the raw value is irrelevant because config ignores it.
- If `Explicit() == true` and the bound config key kind is `KindStringList`, convert all `state.Values()` entries to strings and pass `[]string`.
- If `Explicit() == true` and there are multiple values for a non-`KindStringList` config key, let `config.NewFlagSnapshot` reject incompatible shapes unless an existing `flags` definition proves last-value semantics are required. Current non-repeatable scalar flags store one value; repeatable scalar flags can store several and should not be silently collapsed in this bridge.

### File Structure Requirements

- **NEW targets**:
  - `cli/binding.go`
  - `cli/binding_test.go`
  - `cli/binding_errors.go` if keeping binding errors separate from invocation errors improves readability.
- **UPDATE targets during implementation**:
  - `cli/doc.go` only if package docs need one narrow sentence naming flag-binding composition.
  - `_bmad-output/implementation-artifacts/7-2-compose-command-routing-with-config-flag-bindings.md` for Dev Agent Record updates.
  - `_bmad-output/implementation-artifacts/sprint-status.yaml` for normal dev-story status transitions.
- **Do not touch in Story 7.2 unless a blocking issue is found**:
  - `command/`
  - `flags/`
  - `config/`
  - `README.md`
  - `docs/behavior-matrices.md`
  - `docs/release-notes-v0.md`
  - `docs/release-checklist.md`
  - `examples/`
  - coverage thresholds and release evidence docs

### Architecture Guardrails

- `cli/` is optional composition support, not a root facade.
- `cli/` may import `command`, `flags`, and `config`; reverse imports are prohibited.
- Runtime imports must remain standard-library-only plus local module packages.
- All process inputs are caller-supplied. Do not read ambient args/env or files.
- Do not execute callbacks or own application lifecycle.
- Do not render diagnostics as the only error surface; return typed errors.
- Do not leak raw sensitive flag/config values in error strings or diagnostic helpers.
- Keep definitions reusable and results per-run snapshots.

### Git Intelligence

- Recent commits:
  - `18253ee feat(story-7.1): Add Explicit CLI Invocation Boundaries`
  - `401aa92 docs(bmad): add epic 7 cli composition plan`
  - `8094bd1 feat(story-6.4): Reconcile Release Evidence And Tracker State`
  - `00b8388 feat(story-6.3): Publish Public Usage Documentation`
  - `c8c720b feat(story-6.2): Add Coverage Validation`
- Story 7.1 changed only story workflow artifacts plus `cli/doc.go`, `cli/errors.go`, `cli/invocation.go`, `cli/invocation_qa_test.go`, and `cli/invocation_test.go`.
- `go list` import check with `GOCACHE=/tmp/dib-go-cache` currently reports:
  - `cli`: `errors,fmt`
  - `command`: standard library plus `github.com/petabytecl/dib/flags`
  - `flags`: standard library only
  - `config`: standard library only

### Anti-Patterns To Avoid

- Do not manually resolve config precedence in `cli`.
- Do not convert every default flag value into a config flag binding.
- Do not silently ignore unknown flag names in bindings.
- Do not silently drop duplicate bindings; let typed errors report them.
- Do not collapse repeated scalar values to "first" or "last" unless a documented config contract requires it.
- Do not pass `[]any` where `config.KindStringList` requires `[]string`.
- Do not add an API that accepts raw argv and routes/configures in one call; that is Story 7.3 or later.
- Do not broaden `cli` into a compatibility layer for Cobra, pflag, Viper, or Go `flag`.

### Validation Checklist Applied

- Story includes exact story ID/key (`7-2-compose-command-routing-with-config-flag-bindings`), ready-for-dev status, role/action/benefit, acceptance criteria, and task mapping to ACs.
- Story identifies expected new and updated files and bans unrelated package/docs/tooling changes.
- Story captures previous Story 7.1 learnings and preserves Story 7.3 scope.
- Story names exact existing APIs to reuse and the manual glue being replaced.
- Story includes kind-shape guidance for `KindStringList`, typed error requirements, import-boundary requirements, and sensitive-value non-leakage tests.
- Story records GitHub tracker issue #48 and baseline commit `18253ee`.

### Story Completion Status

- Status set to `ready-for-dev`.
- Completion note: Ultimate context engine analysis completed - comprehensive developer guide created.

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `GOCACHE=/tmp/dib-go-cache go test ./cli ./config ./command ./flags` - pass
- `GOCACHE=/tmp/dib-go-cache go vet ./...` - pass
- `GOCACHE=/tmp/dib-go-cache go run ./tools/lint` - pass
- `GOCACHE=/tmp/dib-go-cache go run ./tools/coverage` - pass (`command` 85.2%, `config` 89.6%, `flags` 85.0%)
- `GOCACHE=/tmp/dib-go-cache go run ./tools/depgate` - pass
- `git diff --check` - pass
- `GOCACHE=/tmp/dib-go-cache go test ./...` - pass

### Implementation Plan

- Add a minimal `cli` flag-binding API that translates routed `flags.Snapshot` values into `config.FlagValue` inputs without introducing config precedence logic.
- Add typed `cli` binding errors that preserve route flag/config key context and wrap config source causes for `errors.Is` / `errors.As`.
- Prove precedence, absent defaults, string-list conversion, sensitive diagnostics, and import boundaries with `cli` package tests.

### Completion Notes List

- Implemented `FlagBinding`, `BindFlag`, and `NewFlagSnapshot` in `cli`, using only exported `command`, `flags`, and `config` package contracts.
- Added inspectable `BindingError` support with `ErrInvalidBinding` and `ErrUnknownFlagBinding`; config failures remain inspectable through the wrapped `config.SourceError`.
- Added `cli` tests for explicit-only config flag binding precedence, absent/default behavior, missing route snapshots, unknown flags, unknown config keys, string-list conversion with defensive copies, sensitive value redaction, and package import boundaries.
- Preserved Story 7.2 scope: no `cli.Resolve`, no process ownership, no command/flags/config edits, and no public documentation expansion.
- Review fix: `BindingError.Error()` now reports wrapped config failures with value-free category/source context so non-sensitive conversion failures from flag values do not echo raw CLI input through the `cli` diagnostic.
- Review fix: duplicate config binding failures now report the duplicate binding's flag/config context while preserving the wrapped `config.ErrDuplicateBinding` source error.

### Senior Developer Review (AI)

Reviewer: GPT-5 Codex on 2026-06-12

Outcome: Approved after automatic fixes.

Findings fixed:

- HIGH: `cli.BindingError.Error()` appended the wrapped `config.SourceError` string directly, which could echo raw non-sensitive parsed flag values on source conversion failures. Fixed by rendering wrapped causes through value-free category/source context while preserving `errors.Is` / `errors.As` and `Cause()` inspection.
- MEDIUM: duplicate config binding errors reported the first matching binding instead of the duplicate binding that triggered the collision. Fixed by deriving duplicate binding context in binding order from the config set's normalized key lookup.

Validation:

- `GOCACHE=/tmp/dib-go-cache go test ./cli ./config ./command ./flags` - pass
- `GOCACHE=/tmp/dib-go-cache go test ./...` - pass
- `GOCACHE=/tmp/dib-go-cache go vet ./...` - pass
- `GOCACHE=/tmp/dib-go-cache go run ./tools/lint` - pass
- `GOCACHE=/tmp/dib-go-cache go run ./tools/coverage` - pass (`command` 85.2%, `config` 89.6%, `flags` 85.0%)
- `GOCACHE=/tmp/dib-go-cache go run ./tools/depgate` - pass
- `git diff --check` - pass

### File List

- `cli/binding.go`
- `cli/binding_errors.go`
- `cli/binding_test.go`
- `_bmad-output/implementation-artifacts/7-2-compose-command-routing-with-config-flag-bindings.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`

### Change Log

- 2026-06-12: Added CLI flag-binding bridge, typed binding errors, and bridge behavior tests; verified all Story 7.2 gates and moved story to review.
- 2026-06-12: Senior developer review fixed value-free binding diagnostics and duplicate binding context; verified all Story 7.2 gates and moved story to done.
