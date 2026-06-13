---
baseline_commit: e72254e
created: "2026-06-12"
---

# Story 7.3: Resolve A CLI Composition Plan Without Owning Execution

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Go CLI developer,
I want a `cli.Resolve` or equivalent composition call that returns route, flags, config, and remaining args,
so that application code can make execution decisions without Dib invoking callbacks or exiting the process.

## Requirements Trace

- FR26: Compose CLI invocation, command routing, flag parsing, and config resolution through an optional `cli` package without process lifecycle ownership.
- FR1: Command trees and nested routing — `cli.Resolve` delegates routing to `command.Definition.Route`.
- FR12: Resolve config values by documented precedence — `cli.Resolve` calls `config.Resolve` with all five source tiers.
- FR16: Retrieve typed config values — the resolved `config.Snapshot` in `cli.Result` is the caller's typed-retrieval entry point.
- FR20: Provide behavior test matrices and package tests for public behavior.
- NFR1: Runtime packages must import only the Go standard library.
- NFR2: `cli` helpers are allowed only with caller-supplied inputs and no hidden singleton state.
- NFR3: Public error cases needed by callers must be inspectable without string matching.
- NFR5: Library APIs must not call `os.Exit`, mutate process streams, or read `os.Args` implicitly.
- NFR8: Diagnostics must identify bad keys, flags, and sources without dumping sensitive values.
- Architecture: `cli/` may import `command`, `flags`, and `config`; those packages must not import `cli`. [Source: `_bmad-output/planning-artifacts/architecture.md#Architectural Boundaries`]
- Architecture: `cli/resolve.go` and `cli/result.go` are the planned new files for this story. [Source: `_bmad-output/planning-artifacts/architecture.md#Complete Project Directory Structure`]
- Sprint proposal: Story 7.3 is GitHub issue #49. [Source: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-06-12-cli-composition.md`]

## Acceptance Criteria

1. Given a plan contains a root command, config set, source snapshots, and flag bindings, when `cli.Resolve(invocation, plan)` succeeds, then the result exposes the route result, flag snapshot when present, resolved config snapshot, invocation, and remaining args.

2. Given routing fails, when `cli.Resolve` returns an error, then the error remains inspectable as the underlying command or flag error where applicable, and no config resolution occurs from partial route state.

3. Given config source construction fails, when a binding or source error occurs, then a typed config or cli error is returned, and sensitive values remain redacted.

4. Given Dib does not own application execution, when `cli.Resolve` succeeds or fails, then it does not invoke callbacks, call `os.Exit`, write to stdout/stderr, read env, or load files implicitly.

## Tasks / Subtasks

- [x] Confirm preconditions and current API surface (AC: 1-4)
  - [x] Verify sprint status marks Story 7.2 `done` and Story 7.3 `ready-for-dev`.
  - [x] Read ALL current `cli/` files before editing: `cli/doc.go`, `cli/errors.go`, `cli/invocation.go`, `cli/binding.go`, `cli/binding_errors.go`.
  - [x] Read route/result/config APIs: `command/result.go`, `command/route.go`, `command/flags.go`, `flags/snapshot.go`, `config/resolve.go`, `config/set.go`, `config/snapshot.go`, `config/source.go`, `config/flag.go`.
  - [x] Do not edit `command/`, `flags/`, or `config/` for this story unless an existing exported contract is proven insufficient; the expected implementation belongs entirely in `cli/`.

- [x] Add `cli.Plan` composition value in `cli/resolve.go` (AC: 1, 4)
  - [x] Define `type Plan struct` with unexported fields for root command, config set, pre-built source snapshots, and flag bindings.
  - [x] Add constructor `func NewPlan(root command.Definition, set config.Set, bindings []FlagBinding) Plan` as the minimal required constructor; callers add source snapshots through chainable `With*` methods or a `PlanOption` pattern to keep zero values safe.
    - Choose the approach that produces the clearest call-site usage without introducing a mutable builder — either a `PlanOption` variadic or explicit `With*` value methods that return new `Plan` values. Value-method chaining (immutable) is consistent with `flags.Set` and `config.Set` patterns already in the codebase.
  - [x] Expose accessor methods on `Plan` for: `Root() command.Definition`, `ConfigSet() config.Set`, `ExplicitSnapshot() config.Snapshot`, `EnvSnapshot() config.Snapshot`, `JSONSnapshot() config.Snapshot`, `Bindings() []FlagBinding`.
  - [x] Pre-built source snapshots default to zero-value `config.Snapshot{}` (empty) when not supplied — this is the safe default because `config.Resolve` handles empty snapshots correctly.
  - [x] If using `With*` value methods, implement at least `WithExplicit(s config.Snapshot) Plan`, `WithEnv(s config.Snapshot) Plan`, `WithJSON(s config.Snapshot) Plan`, `WithBindings(b []FlagBinding) Plan`.
  - [x] Do not store mutable shared state inside `Plan`. All slices must be defensively copied at construction time.

- [x] Add `cli.Result` composition result in `cli/result.go` (AC: 1)
  - [x] Define `type Result struct` with unexported fields: invocation `Invocation`, route `command.Result`, flagSnapshot `flags.Snapshot`, hasFlagSnapshot `bool`, config `config.Snapshot`, remaining `[]string`.
  - [x] Add unexported constructor `newResult(inv Invocation, route command.Result, cfg config.Snapshot) Result` and `newFlagResult(inv Invocation, route command.Result, flagSnap flags.Snapshot, cfg config.Snapshot) Result`.
  - [x] Expose accessors: `Invocation() Invocation`, `Route() command.Result`, `FlagSnapshot() (flags.Snapshot, bool)`, `Config() config.Snapshot`, `RemainingArgs() []string`.
  - [x] `RemainingArgs()` must return a defensive copy — use `append([]string(nil), r.remaining...)` — consistent with `command.Result.RemainingArgs()` and `cli.Invocation.Args()`.
  - [x] `FlagSnapshot()` returns `(flags.Snapshot, false)` zero value when no flag snapshot is present — same dual-return pattern as `command.Result.FlagSnapshot()`.
  - [x] `Config()` returns the fully resolved `config.Snapshot`; callers use `config.Getter` (or equivalent typed getter) on it to retrieve values.

- [x] Implement `cli.Resolve` in `cli/resolve.go` (AC: 1-4)
  - [x] Implement `func Resolve(inv Invocation, plan Plan) (Result, error)` following this strict sequence:
    1. Call `plan.Root().Route(inv.Args())` to get `command.Result` and routing error.
    2. If routing returns an error, return `Result{}` and the raw routing error — do NOT attempt config resolution. The raw error is already a typed `command` error (`errors.Is(err, command.Err*)`).
    3. Call `cli.NewFlagSnapshot(plan.ConfigSet(), routeResult, plan.Bindings())` to build the flag-tier snapshot. If this returns an error, return `Result{}` and the error (already a typed `*cli.BindingError`).
    4. Call `config.Resolve(plan.ConfigSet(), plan.ExplicitSnapshot(), flagTierSnapshot, plan.EnvSnapshot(), plan.JSONSnapshot())` to get the fully resolved config snapshot.
    5. Build and return `Result`. Use `newFlagResult` if `routeResult.FlagSnapshot()` is present; otherwise use `newResult`.
  - [x] `cli.Resolve` must not call `os.Exit`, write to any writer, read `os.Args`, read env, load files, execute callbacks, or hide errors behind rendered text.
  - [x] `cli.Resolve` returns the raw routing error directly (step 2 above) — do not wrap it in a new `cli` error type, because the AC says "inspectable as the underlying command or flag error where applicable." Callers can use `errors.Is(err, command.ErrUnknownCommand)` directly.
  - [x] `cli.Resolve` returns the raw `*cli.BindingError` directly from `cli.NewFlagSnapshot` (step 3 above) — it already satisfies `errors.Is`/`errors.As`.
  - [x] `config.Resolve` does not return an error; it returns a fully resolved snapshot. No config error handling is needed at step 4.

- [x] Add `cli.ResolveError` only if needed for AC 3 (AC: 3)
  - [x] The current design covers AC 3 without a new error type because routing errors are raw `command` errors and binding errors are raw `*cli.BindingError`. If the implementation reveals a new category of error that cannot be exposed through existing typed errors, add a `ResolveError` then — do not add it preemptively.

- [x] Add `cli/resolve_test.go` tests (AC: 1-4)
  - [x] Use `package cli_test`. Cover the following cases:
  - [x] **Success path**: route a root+child command with explicit flag bindings, env snapshot, and JSON snapshot; assert `result.Route()` identifies the correct command, `result.FlagSnapshot()` returns the flag snapshot, `result.Config()` resolves by precedence (flag over env over JSON over default), `result.RemainingArgs()` matches expected, `result.Invocation().Program()` matches the supplied program name.
  - [x] **Success path — no flags**: route a command with no flag definitions; `result.FlagSnapshot()` returns `(flags.Snapshot{}, false)` and config resolves from env/JSON/default.
  - [x] **Success path — zero bindings**: supply empty `Bindings()` on the plan; `result.Config()` resolves from explicit/env/JSON/default without a flag tier.
  - [x] **Routing failure**: pass args that do not match any command; assert `errors.Is(err, command.ErrUnknownCommand)` (or the appropriate command sentinel), `Result{}` is returned, and no config resolution occurred.
  - [x] **Binding failure after routing success**: route succeeds but a binding references an unknown flag name; assert `errors.Is(err, cli.ErrUnknownFlagBinding)` and `errors.As(err, new(*cli.BindingError))` succeed; `Result{}` is returned.
  - [x] **Sensitive value redaction**: supply a config key marked sensitive with a fake-corpus value (`dib_fake_secret_value`) through the env snapshot; assert the env value is reachable through `result.Config()` typed getter but does NOT appear in any error strings or `BindingError.Error()` output if a binding error also fires.
  - [x] **Defensive copy — RemainingArgs**: mutate the returned slice; assert `result.RemainingArgs()` is unchanged on a second call.
  - [x] **Import boundary**: assert via `go list` or a build-constraint test that `command`, `flags`, and `config` do not import `cli` after this story's additions.
  - [x] **No process control**: the entire test file runs without capturing `os.Stdout`, `os.Stderr`, or `os.Exit` — verified implicitly by the tests not panicking and `go vet ./...` passing.

- [x] Add `cli/result_test.go` tests (AC: 1)
  - [x] Cover `Result` accessor contracts: `Invocation()`, `Route()`, `FlagSnapshot()` present and absent, `Config()`, `RemainingArgs()` defensive copy.
  - [x] Use `package cli_test`.

- [x] Verify (AC: 1-4)
  - [x] `GOCACHE=/tmp/dib-go-cache go test ./cli ./...`
  - [x] `GOCACHE=/tmp/dib-go-cache go vet ./...`
  - [x] `GOCACHE=/tmp/dib-go-cache go run ./tools/lint`
  - [x] `GOCACHE=/tmp/dib-go-cache go run ./tools/coverage`
  - [x] `GOCACHE=/tmp/dib-go-cache go run ./tools/depgate`
  - [x] `git diff --check`

## Dev Notes

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`; Epic 7 is `in-progress`, Story 7.1 is `done`, Story 7.2 is `done`, Story 7.3 was `backlog` at story creation.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`; Story 7.3 acceptance criteria defined under `## Epic 7: CLI Composition Ergonomics`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`; architecture directly lists `cli/resolve.go` and `cli/result.go` as expected files for this story's composition layer.
- Loaded sprint change proposal: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-06-12-cli-composition.md`; Story 7.3 is GitHub issue #49.
- No UX artifact exists; Dib V1 is a Go library.

### Previous Story Intelligence (Story 7.2)

- Story 7.2 added the `cli` flag-binding bridge and finished with status `done` on baseline commit `e72254e`.
- Existing `cli` files after Story 7.2:
  - `cli/doc.go`: package-level doc; update one sentence to name `cli.Resolve` as the composition entry point after this story.
  - `cli/invocation.go`: `Invocation`, `FromOSArgs`, `FromArgs`, `Program()`, `Args()` — all defensively copied.
  - `cli/errors.go`: `ErrInvalidInvocation`, `InvocationError` with `Category()`, `Problem()`, `Unwrap()`.
  - `cli/binding.go`: `FlagBinding`, `BindFlag`, `NewFlagSnapshot` — the bridge this story calls.
  - `cli/binding_errors.go`: `ErrInvalidBinding`, `ErrUnknownFlagBinding`, `BindingError` with `FlagName()`, `ConfigKey()`, `Category()`, `Cause()`, `Is()`, `Unwrap()`.
- Story 7.2 used `GOCACHE=/tmp/dib-go-cache` prefix for all `go` commands because the default Go cache may not be writable in this environment. Use the same prefix throughout this story.
- Story 7.2 explicitly deferred `cli.Resolve`, `Plan`, and a full route/config composition result — Story 7.3 owns those exactly.
- Story 7.2 review fix: `BindingError.Error()` uses `bindingCauseSummary` to avoid echoing raw flag values through the cause chain. Story 7.3 errors that wrap `*cli.BindingError` must preserve this redaction.

### Existing Code Patterns To Reuse

- **Immutable value pattern**: all public values use unexported fields plus accessor methods. Follow `cli.Invocation`, `command.Result`, `flags.Snapshot`, and `config.Snapshot`.
- **Defensive copies**: slices are copied with `append([]T(nil), values...)` before storage and before return. This is non-negotiable; see `command.Result.RemainingArgs()` and `cli.Invocation.Args()`.
- **Dual-return optional**: `(value, bool)` dual return for optional fields. See `command.Result.FlagSnapshot() (flags.Snapshot, bool)` and `command.Result.Flags() (flags.Set, bool)`. `cli.Result.FlagSnapshot()` must use the same shape.
- **Raw error passthrough for routing**: `command.Definition.Route` already returns typed `command` errors. Do not wrap these in a new `cli` error type; return them directly so `errors.Is(err, command.ErrUnknownCommand)` works at the call site without extra unwrapping.
- **Typed error wrapping for binding**: `cli.NewFlagSnapshot` returns `*cli.BindingError` which already has `Is`/`Unwrap`. Return it directly from `cli.Resolve`; do not double-wrap.
- **`With*` value methods for optional fields**: `config.Set.With(defs...)` returns a new `Set`. Use the same value-method chaining pattern for `Plan` optional fields (`WithExplicit`, `WithEnv`, `WithJSON`, `WithBindings`) instead of a mutable builder.
- **Empty source snapshots**: `config.Resolve` accepts any `config.Snapshot`, including zero values. Passing a zero-value `config.Snapshot{}` for a tier that the caller has not supplied is correct and safe.

### Current API Details — Everything `cli.Resolve` Uses

```
command.Definition.Route(args []string) (command.Result, error)
  → routes args against the command tree; returns typed command errors

command.Result.FlagSnapshot() (flags.Snapshot, bool)
  → flag snapshot present only when the matched command has flag definitions

command.Result.RemainingArgs() []string
  → args after the matched command path and flag terminator

cli.NewFlagSnapshot(set config.Set, route command.Result, bindings []FlagBinding) (config.Snapshot, error)
  → Story 7.2 bridge; translates explicit route flags into config.FlagValue entries

config.Resolve(set Set, explicit, flag, env, jsonSrc Snapshot) Snapshot
  → single precedence call; never returns an error; returns fully resolved snapshot
  → precedence: explicit setter > flag binding > env > JSON > default

config.Set.DefaultSnapshot() Snapshot
  → already exists; returns a snapshot with only default values; use this as the
     basis when no other snapshot is needed, but note that config.Resolve handles
     zero-value snapshots correctly without needing DefaultSnapshot explicitly

config.Snapshot (zero value is safe)
  → empty snapshot accepted by config.Resolve; callers omit a tier by passing
     config.Snapshot{} or the zero value
```

### Suggested Implementation Shape

The story does not require these exact exported names if the implementation uses a clearer local convention, but the public surface should match the architecture diagram and be Story 7.4-ready for documentation and examples.

```go
// cli/resolve.go

package cli

import (
    "github.com/petabytecl/dib/command"
    "github.com/petabytecl/dib/config"
)

// Plan carries the caller-supplied composition inputs for cli.Resolve.
type Plan struct {
    root     command.Definition
    set      config.Set
    explicit config.Snapshot
    env      config.Snapshot
    jsonSrc  config.Snapshot
    bindings []FlagBinding
}

// NewPlan returns a Plan with the required root command and config set.
// Optional source snapshots and flag bindings are added via With* methods.
func NewPlan(root command.Definition, set config.Set) Plan {
    return Plan{root: root, set: set}
}

func (p Plan) WithExplicit(s config.Snapshot) Plan { p.explicit = s; return p }
func (p Plan) WithEnv(s config.Snapshot) Plan      { p.env = s; return p }
func (p Plan) WithJSON(s config.Snapshot) Plan      { p.jsonSrc = s; return p }
func (p Plan) WithBindings(b []FlagBinding) Plan {
    p.bindings = append([]FlagBinding(nil), b...)
    return p
}

func (p Plan) Root() command.Definition   { return p.root }
func (p Plan) ConfigSet() config.Set      { return p.set }
func (p Plan) ExplicitSnapshot() config.Snapshot { return p.explicit }
func (p Plan) EnvSnapshot() config.Snapshot      { return p.env }
func (p Plan) JSONSnapshot() config.Snapshot     { return p.jsonSrc }
func (p Plan) Bindings() []FlagBinding            { return append([]FlagBinding(nil), p.bindings...) }

// Resolve routes the invocation, builds the flag-tier config snapshot, resolves
// config by precedence, and returns a Result with all composed outputs.
//
// Resolve does not invoke callbacks, call os.Exit, write to any writer, read
// os.Args or env, or load files. All process inputs are caller-supplied.
func Resolve(inv Invocation, plan Plan) (Result, error) {
    route, err := plan.Root().Route(inv.Args())
    if err != nil {
        return Result{}, err  // raw command error; inspectable with errors.Is
    }

    flagSnap, err := NewFlagSnapshot(plan.ConfigSet(), route, plan.Bindings())
    if err != nil {
        return Result{}, err  // raw *cli.BindingError
    }

    resolved := config.Resolve(
        plan.ConfigSet(),
        plan.ExplicitSnapshot(),
        flagSnap,
        plan.EnvSnapshot(),
        plan.JSONSnapshot(),
    )

    snap, hasFlagSnap := route.FlagSnapshot()
    if hasFlagSnap {
        return newFlagResult(inv, route, snap, resolved), nil
    }
    return newResult(inv, route, resolved), nil
}
```

```go
// cli/result.go

package cli

import (
    "github.com/petabytecl/dib/command"
    "github.com/petabytecl/dib/config"
    "github.com/petabytecl/dib/flags"
)

// Result is an immutable snapshot of a successful cli.Resolve call.
type Result struct {
    invocation   Invocation
    route        command.Result
    flagSnapshot flags.Snapshot
    hasFlagSnap  bool
    config       config.Snapshot
}

func newResult(inv Invocation, route command.Result, cfg config.Snapshot) Result {
    return Result{invocation: inv, route: route, config: cfg}
}

func newFlagResult(inv Invocation, route command.Result, snap flags.Snapshot, cfg config.Snapshot) Result {
    return Result{invocation: inv, route: route, flagSnapshot: snap, hasFlagSnap: true, config: cfg}
}

func (r Result) Invocation() Invocation                 { return r.invocation }
func (r Result) Route() command.Result                  { return r.route }
func (r Result) FlagSnapshot() (flags.Snapshot, bool)   { return r.flagSnapshot, r.hasFlagSnap }
func (r Result) Config() config.Snapshot                { return r.config }
func (r Result) RemainingArgs() []string {
    return append([]string(nil), r.route.RemainingArgs()...)
}
```

### File Structure Requirements

- **NEW targets**:
  - `cli/resolve.go` — `Plan` type, `Plan` constructor and `With*` methods, `Resolve` function.
  - `cli/result.go` — `Result` type, `newResult`, `newFlagResult`, accessor methods.
  - `cli/resolve_test.go` — end-to-end and error-path tests for `Resolve`.
  - `cli/result_test.go` — unit tests for `Result` accessors and defensive copies.

- **UPDATE targets during implementation**:
  - `cli/doc.go` — add one sentence naming `cli.Resolve` as the primary composition entry point.
  - `_bmad-output/implementation-artifacts/7-3-resolve-a-cli-composition-plan-without-owning-execution.md` — Dev Agent Record updates.
  - `_bmad-output/implementation-artifacts/sprint-status.yaml` — normal status transitions.

- **Do not touch in Story 7.3 unless a blocking issue is found**:
  - `command/`, `flags/`, `config/` — all required contracts already exist.
  - `cli/binding.go`, `cli/binding_errors.go` — Story 7.2 work; only touch if a bug blocks this story.
  - `README.md`, `docs/behavior-matrices.md`, `docs/release-notes-v0.md`, `docs/release-checklist.md`, `examples/` — Story 7.4 owns public docs and evidence reconciliation.
  - Coverage thresholds — Story 7.4 owns threshold updates for the new `cli` functions.

### Architecture Guardrails

- `cli/` may import `command`, `flags`, and `config`; reverse imports are strictly prohibited. Verify with `go list` in tests.
- Runtime imports must remain standard-library-only plus local module packages (`github.com/petabytecl/dib/*`).
- All process inputs are caller-supplied. `cli.Resolve` does NOT read `os.Args`, `os.Environ`, filesystem paths, or stdin.
- Do not execute callbacks, write to `os.Stdout`/`os.Stderr`, or call `os.Exit`.
- Do not render diagnostics as the only error surface; always return typed errors.
- Do not leak raw sensitive flag or config values in error strings, `Error()` output, or diagnostic helpers.
- Keep `Plan` and `Result` as caller-observably immutable values — no exported mutable fields, no slice aliases.
- Do not add a root package facade, package-global default `Plan`, or singleton composition state.
- `cli.Resolve` does not own or replicate `config.Resolve` precedence logic; it only delegates.

### Anti-Patterns To Avoid

- **Do not wrap routing errors** in a new `cli.ResolveError` type — the AC says they must remain inspectable as `command` errors.
- **Do not double-wrap binding errors** — `*cli.BindingError` is already the correct return from `cli.NewFlagSnapshot`.
- **Do not call `config.Resolve` on partial state** — if routing fails, return immediately without touching config.
- **Do not expose mutable `Plan` fields** — `With*` methods must return new `Plan` values, not mutate the receiver.
- **Do not omit defensive copies** in `Result.RemainingArgs()` and `Plan.Bindings()`.
- **Do not read `os.Args` inside `cli.Resolve`** — use `inv.Args()` which the caller has already prepared.
- **Do not add a `Plan` that accepts an `io.Reader` for JSON** — JSON loading produces a `config.Snapshot`; callers build that snapshot with `config.LoadJSON`/`config.LoadJSONFile` before creating the plan.
- **Do not duplicate command routing or flag parsing** — route via `command.Definition.Route`, not by re-implementing routing in `cli`.
- **Do not add `examples/`, `README.md`, or behavior matrix updates** in this story — Story 7.4 owns those.

### Git Intelligence

- Recent commits:
  - `e72254e feat(story-7.2): Compose Command Routing With Config Flag Bindings`
  - `18253ee feat(story-7.1): Add Explicit CLI Invocation Boundaries`
  - `401aa92 docs(bmad): add epic 7 cli composition plan`
  - `8094bd1 feat(story-6.4): Reconcile Release Evidence And Tracker State`
  - `00b8388 feat(story-6.3): Publish Public Usage Documentation`
- Story 7.2 added: `cli/binding.go`, `cli/binding_errors.go`, `cli/binding_test.go`.
- Story 7.1 added: `cli/doc.go`, `cli/errors.go`, `cli/invocation.go`, `cli/invocation_test.go`, `cli/invocation_qa_test.go`.
- Verified `go list` imports after Story 7.2: `cli` imports `errors,fmt,github.com/petabytecl/dib/command,github.com/petabytecl/dib/config`; `command`, `flags`, `config` do not import `cli`.
- Story 7.3 will add `flags` to `cli`'s import list (for `flags.Snapshot` in `cli/result.go`). This is explicitly allowed by architecture.
- Coverage after Story 7.2: `command` 85.2%, `config` 89.6%, `flags` 85.0%. Story 7.3 adds `cli` functions that must be covered — see `go run ./tools/coverage` verification step.

### Story Completion Status

- Status set to `ready-for-dev`.
- Completion note: Ultimate context engine analysis completed — comprehensive developer guide created.

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- Implemented `cli.Plan` as an immutable value type with `NewPlan(root, set)` constructor and `WithExplicit`, `WithEnv`, `WithJSON`, `WithBindings` value-method chaining. All slices defensively copied at construction time.
- Implemented `cli.Result` with unexported constructors `newResult`/`newFlagResult` and five accessors. `RemainingArgs()` delegates to `route.RemainingArgs()` (already a defensive copy in `command.Result`).
- Implemented `cli.Resolve` following the strict route → flag-tier → config-resolve sequence; returns raw `command` routing errors and raw `*cli.BindingError` without wrapping, preserving `errors.Is`/`errors.As` inspectability.
- No `cli.ResolveError` type needed; existing typed errors cover all AC 3 cases.
- Updated `cli/doc.go` to name `cli.Resolve` as the primary composition entry point.
- All quality gates passed: `go test ./...`, `go vet ./...`, `tools/lint`, `tools/coverage`, `tools/depgate`, `git diff --check`.

### File List

- cli/resolve.go (new)
- cli/result.go (new)
- cli/resolve_test.go (new)
- cli/resolve_qa_test.go (new)
- cli/result_test.go (new)
- cli/doc.go (updated)
- _bmad-output/implementation-artifacts/7-3-resolve-a-cli-composition-plan-without-owning-execution.md (updated)
- _bmad-output/implementation-artifacts/sprint-status.yaml (updated)

## Senior Developer Review (AI)

**Date:** 2026-06-12
**Reviewer:** Coto (AI review)
**Outcome:** Approved

**Summary:** Implementation is correct and complete. All four ACs are fully implemented. Import boundaries verified (`command`/`flags`/`config` do not import `cli`). New code in `resolve.go` and `result.go` achieves 100% statement coverage. All quality gates pass.

**Issues found:** 1 Medium, 0 High, 0 Critical.

**Fixed:**
- `cli/resolve_qa_test.go` was missing from Dev Agent Record File List — added.

**Verified:**
- `go test ./...` — all packages pass
- `go vet ./...` — clean
- `tools/lint` — clean
- `tools/coverage` — `command` 85.2%, `config` 89.6%, `flags` 85.0% (all pass; `cli` threshold deferred to Story 7.4)
- `tools/depgate` — clean
- `git diff --check` — clean
- Import boundaries: `TestPackageImportBoundariesForCLIComposition` and `TestCLIImportsFlagsAfterResolve` pass

## Change Log

- 2026-06-12: Story 7.3 implemented — added `cli.Plan`, `cli.Result`, and `cli.Resolve` composition layer in `cli/resolve.go` and `cli/result.go`; added comprehensive tests in `cli/resolve_test.go`, `cli/resolve_qa_test.go`, and `cli/result_test.go`; updated `cli/doc.go` package documentation.
- 2026-06-12: Story 7.3 review — approved; fixed File List to include `cli/resolve_qa_test.go`; status set to done.
