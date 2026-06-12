# Sprint Change Proposal: CLI Composition Package

Status: Approved and applied
Date: 2026-06-12
Project: Dib
Prepared by: BMAD Correct Course
Mode: Batch, assumed because no question tool is available in this mode
GitHub tracking issue: https://github.com/petabytecl/dib/issues/45
Approved by: Coto on 2026-06-12

Applied GitHub tracker:
- Epic 7: https://github.com/petabytecl/dib/issues/46
- Story 7.1: https://github.com/petabytecl/dib/issues/47
- Story 7.2: https://github.com/petabytecl/dib/issues/48
- Story 7.3: https://github.com/petabytecl/dib/issues/49
- Story 7.4: https://github.com/petabytecl/dib/issues/50
- Epic 7 retrospective: https://github.com/petabytecl/dib/issues/51

## 1. Issue Summary

The completed V1 library exposes three strong independent packages:

- `command` for command routing and route snapshots.
- `flags` for explicit flag sets and parse snapshots.
- `config` for config definitions, source snapshots, precedence, typed getters, provenance, and diagnostics.

The current public model is correct but still has an adoption-friction gap. A caller who wants to use all three packages together has to repeat the same choreography:

1. Slice the process invocation manually, usually `os.Args[1:]`.
2. Route commands with explicit args.
3. Extract `command.Result.RemainingArgs()` and `command.Result.FlagSnapshot()`.
4. Manually translate `flags.Snapshot` values into `[]config.FlagValue`.
5. Build `config.NewFlagSnapshot`.
6. Resolve config sources in the correct order.
7. Keep command route, remaining args, flag values, and resolved config together for application code.

The party-mode review converged on the same conclusion: this is not a need for a framework or root facade. It is a need for a small explicit composition package that names the invocation boundary and packages the handoff between `command`, `flags`, and `config`.

Trigger: stakeholder request after Epic 6 completion:

> Create a new `cli` package with composition.

Concrete evidence from the codebase:

- `command.Result` already exposes `RemainingArgs()` and `FlagSnapshot()`.
- `config.NewFlagSnapshot` currently expects caller-built `[]config.FlagValue`.
- `examples/migration/config_precedence_migration_test.go` contains helper code that manually maps `flags.Snapshot` values to `config.FlagValue` entries.
- Planning artifacts currently say Dib has three public capability packages and no root facade.

## 2. Change Analysis Checklist

### 2.1 Trigger And Context

- [x] 1.1 Triggering story identified: post-Epic-6 public onboarding and party-mode ergonomics review, after Story 6.3 exposed the public quickstart flow and Story 6.4 reconciled release evidence.
- [x] 1.2 Core problem defined: new stakeholder requirement to reduce explicit composition ceremony without hiding process state or turning Dib into a framework.
- [x] 1.3 Supporting evidence collected: current README package split, architecture package-boundary rules, `command.Result` API, `config.NewFlagSnapshot`, and migration example glue.

### 2.2 Epic Impact Assessment

- [x] 2.1 Current epic assessment: Epics 1-6 are complete in local sprint status and remain valid.
- [x] 2.2 Epic-level change needed: add a new Epic 7 for CLI composition ergonomics.
- [x] 2.3 Remaining epics reviewed: no future epics exist in current planning artifacts.
- [x] 2.4 Future epic invalidation: no existing epic is invalidated.
- [x] 2.5 Epic order: add Epic 7 after Epic 6; no resequencing.

### 2.3 Artifact Conflict And Impact Analysis

- [x] 3.1 PRD impact: add a functional requirement for explicit CLI composition and update package-boundary language.
- [x] 3.2 Architecture impact: update public package boundaries to allow a fourth optional `cli/` composition package. Keep `command`, `flags`, and `config` independent.
- [N/A] 3.3 UI/UX impact: no UX artifact exists and this is an importable Go library.
- [x] 3.4 Other artifacts: update README, behavior matrix, testing docs, release notes/checklist, examples, sprint status, and GitHub issue tracker after approval.

### 2.4 Path Forward Evaluation

- [x] 4.1 Direct adjustment: viable. Add a new post-release-hardening epic with focused stories.
- [x] 4.2 Potential rollback: not viable. Completed work remains valid and should not be reverted.
- [x] 4.3 PRD MVP review: viable only as a small expansion. This changes V1 ergonomics, not core parser/config behavior.
- [x] 4.4 Recommended path: direct adjustment with new Epic 7.

### 2.5 Proposal And Handoff

- [x] 5.1 Issue summary created.
- [x] 5.2 Epic and artifact impacts documented.
- [x] 5.3 Recommended path documented.
- [x] 5.4 PRD MVP impact and action plan defined.
- [x] 5.5 Agent handoff plan defined.
- [x] 6.1 Checklist completion reviewed.
- [x] 6.2 Proposal accuracy reviewed against loaded artifacts.
- [x] 6.3 User approval received on 2026-06-12.
- [x] 6.4 Sprint status updates applied after approval.
- [x] 6.5 Next steps and handoff plan defined.

## 3. Impact Analysis

### Epic Impact

Existing Epics 1-6 remain valid and complete. This change should not be backfilled into Epic 3 or Epic 4 because it cuts across command routing, flag parsing, and config resolution. It is also not only documentation, because it introduces a new public package.

Recommended new epic:

> Epic 7: CLI Composition Ergonomics

Epic 7 should deliver a small `cli` package that composes the existing packages while keeping the existing surfaces independently usable.

### Story Impact

New stories are required:

1. Add explicit invocation boundary helpers in `cli`.
2. Compose command routing and config flag bindings in `cli`.
3. Document and test the golden-path three-package flow.
4. Reconcile planning docs, release evidence, sprint status, and GitHub tracker state.

No completed story requires rollback.

### PRD Impact

The PRD currently states that runtime packages are organized around command routing, flag parsing, and configuration resolution. It also forbids root facades and package-global helpers. The change should preserve those constraints while adding a fourth optional composition package.

The PRD should add one functional requirement:

> FR-26: Compose CLI invocation, command routing, flag parsing, and config resolution through an explicit `cli` package.

The PRD should add one success metric:

> SM-8: A new adopter can build a minimal multi-command CLI that uses `command`, `flags`, and `config` together through `cli` without manually slicing `os.Args[1:]` or manually converting `flags.Snapshot` values into `config.FlagValue` entries.

### Architecture Impact

Architecture must be updated before implementation because it currently names exactly three public capability packages and explicitly says no root facade package is introduced.

The update should be narrow:

- Add `cli/` as a fourth optional public package.
- `cli/` may depend on `command`, `flags`, and `config`.
- `command`, `flags`, and `config` must not depend on `cli`.
- `command` must still not depend on `config`.
- `flags` must still work without `command` or `config`.
- `config` should not need to import `command`; a direct `config -> flags` import remains unnecessary if the bridge lives in `cli`.
- `cli` must not execute callbacks, call `os.Exit`, read process globals, mutate streams, or own application lifecycle.
- `cli.FromOSArgs(os.Args)` is allowed because the caller passes `os.Args` explicitly. The package must not call `os.Args` itself.

### Technical Impact

The likely implementation surface:

```go
package cli

type Invocation struct {
    // unexported fields
}

func FromOSArgs(argv []string) (Invocation, error)
func FromArgs(program string, args []string) Invocation

func (i Invocation) Program() string
func (i Invocation) Args() []string
```

Composition sketch:

```go
type Plan struct {
    Root command.Definition
    Config config.Set
    FlagBindings []FlagBinding
    Explicit config.Snapshot
    Env config.Snapshot
    JSON config.Snapshot
}

type FlagBinding struct {
    FlagName string
    ConfigKey string
}

func Resolve(inv Invocation, plan Plan) (Result, error)
```

Result sketch:

```go
type Result struct {
    // unexported fields
}

func (r Result) Invocation() Invocation
func (r Result) Route() command.Result
func (r Result) Config() config.Snapshot
func (r Result) RemainingArgs() []string
func (r Result) FlagSnapshot() (flags.Snapshot, bool)
```

These names are proposed direction, not final API freeze.

### Documentation Impact

Public docs should teach one vocabulary:

- `os.Args`: full process invocation.
- `os.Args[0]`: program name.
- `os.Args[1:]`: user arguments.
- `cli.Invocation`: Dib's explicit value for program name plus user arguments.

README should add a "Using the packages together" quickstart. Compatibility docs must continue to say Dib is not a Cobra, pflag, Viper, or Go `flag` compatibility layer.

### Tracker Impact

GitHub is already stale relative to local sprint status for older completed epics and stories. This change should not be blocked by that cleanup, but the new Epic 7 tracker work should include:

- Create or update GitHub labels for `epic:7`.
- Create GitHub issues for Epic 7 and its stories after proposal approval.
- Add a comment to the sprint board that local `sprint-status.yaml` remains authoritative until the stale older issue set is reconciled.

## 4. Recommended Approach

Recommended path: Direct adjustment with one new Epic 7.

Rationale:

- The completed implementation remains correct.
- The change improves adoption ergonomics and aligns with the public-docs feedback loop from Epic 6.
- A `cli` package is a visible, auditable composition path without weakening package independence.
- The risk is manageable if the architecture explicitly rejects callbacks, root facade behavior, package globals, and ambient process reads.

Scope classification: Moderate.

This is more than a single direct implementation story because it requires PRD, architecture, epics, sprint status, docs, tests, examples, and tracker updates. It does not require fundamental replanning or rollback.

Estimated effort:

- Planning artifact updates: low to medium.
- Implementation: medium.
- Test/docs/release evidence: medium.

Primary risks:

- `cli` becomes an application framework.
- `cli` hides process state or starts reading `os.Args` directly.
- `cli` duplicates command/flag/config behavior instead of composing exported APIs.
- The docs imply source compatibility or adapter behavior.

Mitigations:

- Make `cli` an optional composition package, not the root module package.
- Make all process inputs caller-supplied.
- Keep result accessors transparent and typed.
- Use tests to prove input slices are copied and no process globals are read.
- Keep compatibility wording behavior-scoped.

## 5. Detailed Change Proposals

### 5.1 PRD Changes

Section: 5 Features

OLD:

```text
This PRD defines the first-version product requirements for Dib, a Go standard-library-only toolkit for command routing, flag parsing, and configuration resolution.
```

NEW:

```text
This PRD defines the first-version product requirements for Dib, a Go standard-library-only toolkit for command routing, flag parsing, configuration resolution, and explicit composition of those surfaces through an optional CLI invocation package.
```

Rationale: The product now includes a public composition path while preserving the independent surfaces.

Section: 5.5 Validation And Release Evidence

OLD:

```text
#### FR-25: Provide public usage documentation
```

NEW:

```text
#### FR-25: Provide public usage documentation

...

#### FR-26: Compose CLI invocation, command routing, flag parsing, and config resolution

Developers can use an optional `cli` package to carry a full process invocation, route commands, translate explicitly-set flags into config bindings, and return typed command, flag, config, and remaining-argument results without handing Dib process lifecycle control.

Consequences (testable):
- Callers pass full argv explicitly, for example `cli.FromOSArgs(os.Args)`, and Dib never reads `os.Args` itself.
- `cli.Invocation` exposes program name and user arguments through defensive accessors.
- `cli.Resolve` or equivalent routes command input, composes exported flag/config snapshots, and returns a result containing route, flag, config, and remaining-argument state.
- The composition package does not execute callbacks, call `os.Exit`, mutate streams, read env implicitly, or hide errors behind rendered text.
- The original `command`, `flags`, and `config` packages remain independently usable.
```

Rationale: Adds the new ergonomic requirement as an explicit product behavior.

Section: 6 Cross-Cutting Non-Functional Requirements

OLD:

```text
NFR-2 Explicit-instance API: Primary APIs must operate on explicit instances and caller-supplied inputs/outputs. V1 does not include package-level global command, flag, or config helpers.
```

NEW:

```text
NFR-2 Explicit-instance API: Primary APIs must operate on explicit instances and caller-supplied inputs/outputs. V1 does not include package-level global command, flag, or config helpers. The optional `cli` package may provide explicit composition helpers only when all process inputs are caller-supplied and no hidden singleton state is introduced.
```

Rationale: Allows composition without weakening explicit-instance constraints.

Section: 8 Non-Goals

OLD:

```text
A global singleton as the default configuration or command pattern.
```

NEW:

```text
A global singleton as the default configuration or command pattern.
A process-owning CLI framework, callback runner, source-compatible adapter, or root module facade.
```

Rationale: Clarifies what the new `cli` package must not become.

Section: 10 Success Metrics

OLD:

```text
SM-7: Public onboarding works without planning artifacts. Target: a new adopter can follow the README and usage docs to build a minimal multi-command CLI using Dib. Validates FR-25 and NFR-12.
```

NEW:

```text
SM-7: Public onboarding works without planning artifacts. Target: a new adopter can follow the README and usage docs to build a minimal multi-command CLI using Dib. Validates FR-25 and NFR-12.
SM-8: Three-package composition is ergonomic without hidden behavior. Target: a new adopter can use `cli` to compose `command`, `flags`, and `config` without manually slicing `os.Args[1:]` or manually converting `flags.Snapshot` values into `config.FlagValue` entries. Validates FR-26, NFR-2, NFR-5, and NFR-6.
```

Rationale: Makes the ergonomics goal measurable.

### 5.2 Architecture Changes

Section: API Boundaries

OLD:

```text
The module root does not provide a broad public facade, package-global helpers, or default singleton API.
```

NEW:

```text
The module root does not provide a broad public facade, package-global helpers, or default singleton API.

`cli/` is an optional public composition package. It owns explicit invocation values and the golden-path handoff between `command`, `flags`, and `config`. It may depend on `command`, `flags`, and `config`; those packages must not depend on `cli`.
```

Rationale: Adds the approved composition package without creating a root facade.

Section: Component Boundaries

OLD:

```text
`config/` accepts explicit flag binding inputs from callers. A direct `config -> flags` package import is deferred unless API design proves it necessary; if introduced, it must depend only on exported snapshot/value contracts.
```

NEW:

```text
`config/` accepts explicit flag binding inputs from callers. A direct `config -> flags` package import remains unnecessary for the CLI composition path because `cli/` owns translation from exported `flags.Snapshot` values into `config.FlagValue` entries.
```

Rationale: Places bridge logic in the new package and preserves existing package independence.

Section: Project Directory Structure

OLD:

```text
├── command/
├── flags/
├── config/
```

NEW:

```text
├── cli/
│   ├── doc.go
│   ├── invocation.go
│   ├── resolve.go
│   ├── result.go
│   ├── bindings.go
│   ├── errors.go
│   ├── invocation_test.go
│   ├── resolve_test.go
│   └── result_test.go
├── command/
├── flags/
├── config/
```

Rationale: Establishes the new package as first-class but narrow.

Section: Pattern Enforcement

ADD:

```text
`cli/` must not call `os.Args`, `os.Exit`, mutate process streams, execute callbacks, read env implicitly, or load files implicitly. It may accept full argv, env-derived snapshots, JSON-derived snapshots, readers, writers, and contexts only when callers pass them explicitly.
```

Rationale: Keeps the ergonomic package aligned with existing runtime-boundary rules.

### 5.3 Epics Changes

Section: Epic List

OLD:

```text
### Epic 6: Release Hardening And Public Usage Onboarding

Developers and reviewers can trust Dib's final release gates and start from public usage documentation without reading planning artifacts.
```

NEW:

```text
### Epic 6: Release Hardening And Public Usage Onboarding

Developers and reviewers can trust Dib's final release gates and start from public usage documentation without reading planning artifacts.

### Epic 7: CLI Composition Ergonomics

Developers can use `command`, `flags`, and `config` together through an explicit `cli` composition package that removes repetitive invocation and flag-binding glue without making Dib a process-owning framework.

FRs covered: FR26, FR20, FR21, FR25
```

Rationale: Adds the new scope after completed release hardening.

ADD:

```text
## Epic 7: CLI Composition Ergonomics

Developers can use `command`, `flags`, and `config` together through an explicit `cli` composition package that removes repetitive invocation and flag-binding glue without making Dib a process-owning framework.

### Story 7.1: Add Explicit CLI Invocation Boundaries

Requirements: FR26, FR20

As a Go CLI developer,
I want a `cli.Invocation` value that carries the program name and user arguments from caller-supplied argv,
So that I do not repeat `os.Args[1:]` slicing or lose testability at the process boundary.

Acceptance Criteria:

Given a caller passes full argv
When `cli.FromOSArgs(argv)` is called
Then the result exposes `Program()` as argv[0] and `Args()` as argv[1:]
And the package never reads `os.Args` itself.

Given a caller already has stripped args
When `cli.FromArgs(program, args)` is called
Then the result exposes the caller-supplied program and args
And all slices are defensively copied.

Given invalid full argv is supplied
When the invocation cannot be constructed
Then a typed error is returned
And no partial mutable state is exposed.

Given invocation values are reusable
When callers mutate the original argv or returned args slice
Then the invocation's observable state does not change.

Verification:
go test ./cli ./...
go vet ./...
go run ./tools/lint
go run ./tools/coverage
go run ./tools/depgate
```

```text
### Story 7.2: Compose Command Routing With Config Flag Bindings

Requirements: FR26, FR13, FR20

As a Go CLI developer,
I want `cli` to translate explicitly-set route flags into config flag bindings,
So that command flags and config precedence work together without manual `config.FlagValue` glue.

Acceptance Criteria:

Given a command route result has a flag snapshot
When `cli` applies explicit flag bindings
Then only explicitly-set flags enter the config `flag binding` tier
And default flag values do not override env, JSON, or defaults.

Given a binding maps a flag name to a config key
When the flag is absent or not explicitly set
Then the resulting config flag source leaves that key absent for the flag tier.

Given a binding references an unknown flag or config key
When composition runs
Then a typed `cli` error preserves enough context for `errors.Is` or `errors.As` inspection.

Given the bridge composes exported package contracts
When implementation is reviewed
Then `cli/` may import `command`, `flags`, and `config`
And `command/`, `flags/`, and `config/` do not import `cli`.

Verification:
go test ./cli ./config ./command ./flags
go vet ./...
go run ./tools/lint
go run ./tools/coverage
go run ./tools/depgate
```

```text
### Story 7.3: Resolve A CLI Composition Plan Without Owning Execution

Requirements: FR26, FR1, FR12, FR16, FR20

As a Go CLI developer,
I want a `cli.Resolve` or equivalent composition call that returns route, flags, config, and remaining args,
So that application code can make execution decisions without Dib invoking callbacks or exiting the process.

Acceptance Criteria:

Given a plan contains a root command, config set, source snapshots, and flag bindings
When `cli.Resolve(invocation, plan)` succeeds
Then the result exposes the route result, flag snapshot when present, resolved config snapshot, invocation, and remaining args.

Given routing fails
When `cli.Resolve` returns an error
Then the error remains inspectable as the underlying command or flag error where applicable
And no config resolution occurs from partial route state.

Given config source construction fails
When a binding or source error occurs
Then a typed config or cli error is returned
And sensitive values remain redacted.

Given Dib does not own application execution
When `cli.Resolve` succeeds or fails
Then it does not invoke callbacks, call `os.Exit`, write to stdout/stderr, read env, or load files implicitly.

Verification:
go test ./cli ./...
go vet ./...
go run ./tools/lint
go run ./tools/coverage
go run ./tools/depgate
```

```text
### Story 7.4: Document And Reconcile CLI Composition Evidence

Requirements: FR25, FR26, FR20, FR21

As a new adopter,
I want docs and examples that show the three packages working together through `cli`,
So that I can start from public documentation without learning internal glue.

Acceptance Criteria:

Given `cli` is added as a public package
When README is updated
Then it includes a "Using command, flags, and config together" quickstart
And it names the invocation boundary as `os.Args`, `os.Args[0]`, and `os.Args[1:]`.

Given examples are executable evidence
When a composition example is added
Then it compiles through `go test ./...`
And it does not imply Cobra, pflag, Viper, or Go `flag` source compatibility.

Given release evidence must stay accurate
When docs are updated
Then `docs/behavior-matrices.md`, `docs/release-notes-v0.md`, `docs/release-checklist.md`, and `docs/testing.md` record the `cli` package scope and gate evidence.

Given BMAD and GitHub tracking must align
When Epic 7 is approved
Then sprint status and GitHub issues are updated for Epic 7 and its stories
And stale older GitHub issue state is either reconciled or explicitly annotated.

Verification:
go test ./docs ./examples/migration ./...
go vet ./...
go run ./tools/lint
go run ./tools/coverage
go run ./tools/depgate
git diff --check
```

### 5.4 Sprint Status Changes

After proposal approval, add these entries to `sprint-status.yaml` through the normal BMAD sprint-planning path:

```yaml
  epic-7: backlog
  7-1-add-explicit-cli-invocation-boundaries: backlog
  7-2-compose-command-routing-with-config-flag-bindings: backlog
  7-3-resolve-a-cli-composition-plan-without-owning-execution: backlog
  7-4-document-and-reconcile-cli-composition-evidence: backlog
  epic-7-retrospective: optional
```

Do not hand-edit sprint status before the planning artifacts are approved.

### 5.5 GitHub Tracker Changes

After proposal approval:

- Create label `epic:7` if missing.
- Create an Epic 7 issue.
- Create Story 7.1 through 7.4 issues.
- Create an Epic 7 retrospective issue.
- Add a board comment that local `sprint-status.yaml` is currently authoritative because older completed issue labels are stale.

## 6. Implementation Handoff

Scope classification: Moderate.

Recommended route:

1. Product Owner / Developer: approve this proposal and update PRD, architecture, epics, and sprint status.
2. Developer: create Story 7.1 with `bmad-create-story`.
3. Developer: implement stories with `bmad-story-automator` or normal story flow.
4. QA/Test Architect: ensure composition tests cover invocation slicing, flag binding, source precedence, typed errors, redaction, and no hidden process reads.
5. Technical Writer: update README and examples so the new package is easy to adopt without hiding the model.

Success criteria:

- `cli` exists as a new optional public package.
- `cli` composes exported `command`, `flags`, and `config` contracts.
- `command`, `flags`, and `config` remain independently usable.
- No package reads `os.Args` implicitly.
- No package calls `os.Exit` or owns application lifecycle.
- No callbacks are invoked by `cli`.
- Public docs explain `os.Args`, `os.Args[0]`, `os.Args[1:]`, and `cli.Invocation`.
- CI and local gates pass: `go test ./...`, `go vet ./...`, `go run ./tools/lint`, `go run ./tools/coverage`, and `go run ./tools/depgate`.

## 7. Approval Status

This proposal is approved and applied to the local planning artifacts.

Applied local artifact updates:

- Added FR-26 and CLI composition guardrails to the PRD.
- Updated architecture package boundaries for optional `cli/` composition.
- Added Epic 7 and Stories 7.1 through 7.4 to the epics backlog.
- Updated `sprint-status.yaml` with Epic 7 backlog entries.
- Created GitHub tracker issues #46 through #51 and linked them from Correct Course issue #45.
- Created Story 7.1 context at `_bmad-output/implementation-artifacts/7-1-add-explicit-cli-invocation-boundaries.md` and moved it to `ready-for-dev`.

Approved scope:

- Epic 7: CLI Composition Ergonomics.
- PRD, architecture, epics, sprint status, README/docs, release evidence, and GitHub tracker updates.
- `cli` as the package name for the optional composition package.

Suggested next BMAD step after approval:

```text
$bmad-create-epics-and-stories
```

or, if you want to apply this proposal directly without regenerating the full backlog:

```text
$bmad-create-story 7.1
```

The approved route is to update epics and sprint status first, then create Story 7.1.
