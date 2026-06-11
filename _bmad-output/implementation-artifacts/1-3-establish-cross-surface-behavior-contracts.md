---
baseline_commit: 9ea64eea900cfbb3533df76250d87138dc27d052
created: "2026-06-11T13:33:07-04:00"
---

# Story 1.3: Establish Cross-Surface Behavior Contracts

Status: done

## Story

As a Dib implementer,
I want the shared behavior contracts for definitions, snapshots, typed errors, provenance, and redaction established early,
so that `command/`, `flags/`, and `config/` do not drift as separate implementations grow.

## Requirements Trace

- FR20: Provide behavior test matrices; this story creates the first cross-surface contract matrix and one executable package-level contract test.
- NFR1, NFR2, NFR3, NFR5, NFR6, NFR8, NFR9: standard-library runtime, explicit-instance APIs, typed errors, no process control, testability, redaction, and Go 1.26+ remain in force.

## Acceptance Criteria

1. Given the architecture requires caller-observably immutable definitions, when the shared behavior contract is documented or represented in initial package-level tests, then derived definitions must return new values rather than mutating existing values, and callers must not observe mutable aliases to slices, maps, readers, env lookups, or config data after construction or resolution.
2. Given the architecture requires per-run snapshots, when a package parses, routes, or resolves input in later stories, then the expected contract states that snapshots never write back to definitions, and snapshots do not depend on live process state, environment variables, readers, or lookup functions after creation.
3. Given public errors are a cross-package contract, when initial error contract guidance or test scaffolding is created, then errors must be inspectable with `errors.Is` or `errors.As` where callers need programmatic handling, and string output is documented as diagnostics rather than the programmatic contract.
4. Given config provenance and redaction rules affect diagnostics across packages, when shared test guidance or fixtures are created, then the source-label vocabulary is fixed to `default`, `explicit setter`, `flag binding`, `env`, and `JSON`, and the fake sensitive-value corpus is fixed to `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.
5. Given behavior contracts are not just prose, when verification runs, then at least one initial standard-library-only test or example demonstrates an explicit-instance, no-ambient-state contract, and the story leaves clear acceptance hooks for later package stories to add immutability, snapshot, typed-error, provenance, and redaction assertions.

## Tasks / Subtasks

- [x] Confirm live repository state before editing (AC: 1-5)
  - [x] Preserve the uncommitted Story 1.1 and Story 1.2 baseline files: `go.mod`, `command/`, `flags/`, `config/`, `CONTRIBUTING.md`, `docs/clean-room-policy.md`, `docs/provenance-log.md`, both completed story files, and `sprint-status.yaml`.
  - [x] Verify `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, and `command/contract_test.go` do not already exist before creating them. If any exists, update it in place instead of overwriting user work.
  - [x] Do not reset, remove, or rewrite previous story implementation work. This story is additive contract documentation plus one narrow test.

- [x] Create the initial cross-surface behavior matrix in `docs/behavior-matrices.md` (AC: 1, 2, 5)
  - [x] Add a clearly marked initial section for shared contracts, not a final V1 behavior matrix.
  - [x] Document definition immutability: derived definitions return new values, exported APIs do not expose mutable internals, and constructors or derivation methods must defensively handle caller-owned slices, maps, readers, env lookups, and config data.
  - [x] Document snapshot isolation: route, parse, and config resolution snapshots never write back to definitions and must not depend on live process state, env lookup functions, readers, or caller-owned mutable inputs after creation.
  - [x] Document explicit-instance and no-ambient-state expectations across `command/`, `flags/`, and `config/`.
  - [x] Add acceptance hooks for later stories to fill in executable assertions for `command/`, `flags/`, and `config/` without claiming behavior that is not implemented yet.

- [x] Create diagnostic and error contract guidance in `docs/diagnostics-and-errors.md` (AC: 3, 4)
  - [x] State that public errors requiring caller handling must be inspectable with `errors.Is` or `errors.As`; diagnostic strings are not programmatic contracts.
  - [x] Define the diagnostic vocabulary shape at a contract level: package surface, command/flag/key name, source label where relevant, typed category, and redaction status.
  - [x] Fix the config source-label vocabulary exactly as `default`, `explicit setter`, `flag binding`, `env`, and `JSON`.
  - [x] Fix the fake sensitive-value corpus exactly as `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.
  - [x] State that raw sensitive values must not appear in errors, `String` output, debug strings, rendered diagnostics, source reports, examples, or validation failures.
  - [x] Keep the guidance clean-room and implementation-neutral; do not copy examples from Go `errors`, Go `testing`, pflag, Cobra, Viper, or other projects.

- [x] Add one standard-library-only executable contract test in `command/` (AC: 5)
  - [x] Add `command/contract_test.go` using `package command_test` so the test observes only exported behavior.
  - [x] Demonstrate explicit-instance construction through `command.NewDefinition` and confirm the resulting `Definition` name is driven by the caller-provided argument, not by `os.Args` or environment variables.
  - [x] If the test mutates process-wide state such as `os.Args` or uses `t.Setenv`, restore it with `t.Cleanup` and do not mark the test or parent test as parallel.
  - [x] Keep the test narrow. Do not add new exported runtime APIs, no root facade, no callbacks, no flag parsing, no config binding, no stdout/stderr assertions, and no package-global helpers.
  - [x] Do not modify `command/definition.go` unless the new test exposes a real Story 1.1 regression.

- [x] Verify the story output (AC: 5)
  - [x] Run `go test ./...`.
  - [x] Run `go vet ./...`.
  - [x] Confirm the story did not create `tools/depgate/`, CI, a root facade package, `/cmd`, examples, config precedence docs, compatibility docs, release checklist, parser fuzz seeds, or full package behavior matrices beyond the initial shared contract hooks.
  - [x] Confirm no runtime, test, documentation-example, or tool dependency was added.
  - [x] Record verification commands and outcomes in the Dev Agent Record.

### Review Findings

- [x] [Review][Patch] Clarify that provenance classifications do not approve disallowed inspiration-project copying [`docs/provenance-log.md:19`]
- [x] [Review][Patch] Clarify provenance URL and access-date limits for copied or adapted material [`docs/provenance-log.md:40`]
- [x] [Review][Patch] Clean stale Dev Agent Record placeholders and unrelated completion notes [`_bmad-output/implementation-artifacts/1-1-adopt-an-auditable-go-module-baseline.md:154`]

## Dev Notes

### Source Discovery

- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded PRD workspace: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`, `addendum.md`, `reconcile-brief.md`, and `review-rubric.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded readiness report: `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-11.md`.
- Loaded product brief: `_bmad-output/planning-artifacts/briefs/brief-dib-2026-06-10/brief.md` and `addendum.md`.
- Loaded previous stories: `_bmad-output/implementation-artifacts/1-1-adopt-an-auditable-go-module-baseline.md` and `_bmad-output/implementation-artifacts/1-2-publish-the-clean-room-contribution-contract.md`.
- Loaded current clean-room docs: `CONTRIBUTING.md`, `docs/clean-room-policy.md`, and `docs/provenance-log.md`.
- Loaded current source files: `go.mod`, `command/doc.go`, `command/definition.go`, `command/definition_test.go`, `flags/doc.go`, and `config/doc.go`.
- No UX document, `project-context.md`, `CLAUDE.md`, or local `MEMORY.md` was discovered in the repo.

### Current Repository State

- Branch at story creation: `main`.
- `sprint-status.yaml` marks Story 1.1 and Story 1.2 done; Story 1.3 is the first backlog story.
- Story 1.1 and Story 1.2 changes are still uncommitted in the working tree. Treat them as owned baseline work, not disposable local noise.
- Existing runtime package files:
  - `go.mod` declares module `github.com/petabytecl/dib` with `go 1.26`.
  - `command/doc.go`, `flags/doc.go`, and `config/doc.go` document explicit instances, caller-owned inputs, immutable definitions, per-run snapshots, typed errors, and no ambient process state.
  - `command/definition.go` exposes `Definition`, `NewDefinition`, `Name`, and `NameError`; use `NewDefinition` for validated definitions because the zero value is not validated.
  - `command/definition_test.go` already uses `package command_test`, table-driven validation, `errors.As`, and a runnable `ExampleNewDefinition`.
- Existing docs:
  - `CONTRIBUTING.md` links to clean-room/provenance docs and requires independent tests/docs/examples/fixtures/matrices/fuzz seeds.
  - `docs/clean-room-policy.md` allows public docs only for behavior understanding and factual metadata; it disallows copied wording, examples, names, layout, and structure.
  - `docs/provenance-log.md` has the initial inspiration-only public source entries and approval fields.

### Architecture Guardrails

- This story establishes shared contracts; it is not the full implementation of command routing, flag parsing, config resolution, dependency tooling, CI, examples, compatibility docs, or release evidence.
- Do not create a root public facade package, package-global command/flag/config helpers, default singleton APIs, `/cmd` scaffolding, generated assets, shell completion, man pages, Docker/Kubernetes files, or deployment config.
- Runtime packages and tests must remain standard-library-only unless the architecture is updated.
- `flags/` must remain independently usable without `command/` or `config/`; `command/` must not depend on `config/`; callers compose surfaces explicitly.
- `internal/` remains provisional. Do not introduce shared internal packages in this story; there are not yet two concrete call sites proving the abstraction.
- `tools/depgate/` belongs to Story 1.4. Do not create it here and do not use the temporary Story 1.1 `go list` command as release-candidate evidence.
- Callback invocation remains deferred. Do not add callback execution behavior or callback-oriented public APIs in this story.

### Expected File Changes

Expected new files for this story:

```text
docs/behavior-matrices.md
docs/diagnostics-and-errors.md
command/contract_test.go
```

Expected BMAD tracking updates during implementation:

```text
_bmad-output/implementation-artifacts/1-3-establish-cross-surface-behavior-contracts.md
_bmad-output/implementation-artifacts/sprint-status.yaml
```

Do not create these in this story unless a verification issue proves they are strictly necessary:

```text
tools/depgate/
.github/workflows/ci.yml
docs/config-precedence.md
docs/compatibility.md
docs/testing.md
docs/release-checklist.md
examples/
flags/testdata/fuzz/
internal/
```

### Behavior Matrix Content Requirements

`docs/behavior-matrices.md` should be an initial contract matrix, not a final support table. Include:

- Purpose: package tests remain executable truth; this doc records shared hooks to prevent package drift.
- A table for shared contracts with columns similar to: contract, applies to, required behavior, initial evidence, later story hook.
- Rows for immutable definitions, no mutable aliases, per-run snapshots, no ambient process state, typed errors, diagnostics-vs-programmatic contract, source labels, and redaction corpus.
- Explicit "not yet implemented" or "future story hook" language where package behavior does not exist yet.
- Trace each row to Story 1.3 or later owner stories where practical.

### Diagnostics And Error Content Requirements

`docs/diagnostics-and-errors.md` should define:

- Public caller handling uses typed or sentinel errors inspectable through Go error inspection. The story requires `errors.Is` or `errors.As`; do not require the newer `errors.AsType` API unless a later API decision chooses it.
- Error strings, help text, source reports, and rendered diagnostics are human-facing diagnostics, not the only programmatic contract.
- Diagnostic vocabulary should identify the package surface, command/flag/key where applicable, source label where applicable, typed category, and redaction status.
- Canonical source labels: `default`, `explicit setter`, `flag binding`, `env`, `JSON`.
- Sensitive corpus: `dib_fake_secret_value`, `dib_fake_password_value`, `dib_fake_token_value`.
- Redaction scope: errors, `String` output, debug strings, rendered diagnostics, source reports, examples, and validation failures.

### Initial Test Guidance

Use a minimal external-caller test in `command/contract_test.go`. One acceptable shape:

- Save and restore `os.Args` with `t.Cleanup`.
- Use `t.Setenv` only in a non-parallel test.
- Set unrelated process state to misleading values.
- Call `command.NewDefinition("serve")` and `command.NewDefinition("status")`.
- Assert returned `Definition.Name()` values match explicit inputs and remain independent.

Do not overstate this test: it demonstrates the first explicit-instance/no-ambient-state hook. Later package stories must add deeper immutability, snapshot, provenance, redaction, and typed-error assertions when those APIs exist.

### Previous Story Intelligence

- Story 1.1 established the module, package boundaries, and `command.NewDefinition` behavior proof. Preserve the narrow API; do not add a broad command builder or derived-definition API just to satisfy Story 1.3 prose.
- Story 1.1 code review removed a false `Unwrap`/`errors.Is` claim. Do not document wrapping behavior unless the error type actually implements it. `errors.As` for `*command.NameError` is the current proven inspection behavior.
- Story 1.2 tightened clean-room policy and provenance language. Do not copy snippets, tests, examples, tables, names, file layout, or source-derived structure from Go `flag`, pflag, Cobra, Viper, or Go docs.
- Story 1.2 intentionally did not create runtime files or tools. Story 1.3 may add one command test and two docs, but must not introduce dependency tooling or broader runtime behavior.

### Latest Technical Context

- Go `errors` documentation confirms `errors.Is` and `errors.As` inspect an error tree and support wrapping contracts; use them only where the type or sentinel behavior actually exists.
- Go 1.26 adds `errors.AsType`, but Story 1.3 acceptance criteria and existing tests call for `errors.Is` or `errors.As`. Do not churn existing Story 1.1 tests to `AsType` unless a later API decision approves that convention.
- Go `testing` documentation supports black-box tests through a separate `_test` package, which is the right pattern for Story 1.3's external-caller contract test.
- Go example tests are executable documentation when they include an `Output:` comment. This story does not require a new example because Story 1.1 already has `ExampleNewDefinition`; add a new example only if it clarifies the contract without broadening scope.
- `testing.T.Setenv` affects process environment, so no test using it should be parallel or have parallel ancestors.

### Testing Standards

- Use only the standard `testing` package and standard library imports.
- Prefer external-caller tests (`package command_test`) when asserting public contracts.
- Tests must inspect returned values/errors, not stdout/stderr or diagnostic strings.
- Keep process-state mutation in tests explicit, restored, and non-parallel.
- Required verification for this story: `go test ./...` and `go vet ./...`.
- If `tools/depgate/` does not exist, do not create it here. If it already exists by the time implementation runs, use its own documentation.

### Security And Quality Checks

- No secrets, tokens, generated credentials, real private paths, network calls, or host-specific files belong in docs or tests.
- The fake sensitive values are documentation/test corpus only: `dib_fake_secret_value`, `dib_fake_password_value`, and `dib_fake_token_value`.
- Do not dump fake sensitive values in rendered diagnostics, source reports, examples, or validation failures once redaction behavior exists.
- Preserve immutability expectations: do not expose mutable internals, do not mutate caller-owned inputs, and prefer returning new values over modifying existing values in place.
- Keep docs concise enough for reviewers to audit. Use project-owned wording and record provenance if any reference influence becomes more than inspiration-only.

### References

- `_bmad-output/planning-artifacts/epics.md:264` - Story 1.3 source story.
- `_bmad-output/planning-artifacts/epics.md:274` - Immutable definition acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:279` - Snapshot isolation acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:284` - Public error contract acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:289` - Provenance and redaction vocabulary acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:294` - Executable contract hook acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:94` - Definitions, snapshots, validation, boundary inputs, typed errors, and source-label rules.
- `_bmad-output/planning-artifacts/epics.md:108` - Clean-room/provenance docs and behavior matrix requirements.
- `_bmad-output/planning-artifacts/architecture.md:265` - Implementation sequence places immutable definition/result snapshot contracts and typed error/redaction contracts before public capability surfaces.
- `_bmad-output/planning-artifacts/architecture.md:274` - Cross-component dependency notes for snapshots and error taxonomy.
- `_bmad-output/planning-artifacts/architecture.md:296` - Critical conflict points include terminology, errors, config provenance, redaction, and contract assertion drift.
- `_bmad-output/planning-artifacts/architecture.md:556` - Public package boundaries and callback deferral.
- `_bmad-output/planning-artifacts/architecture.md:598` - Cross-cutting mapping for typed errors, determinism, redaction, clean-room evidence, and runtime dependency enforcement.
- `_bmad-output/planning-artifacts/architecture.md:605` - Internal communication and explicit values/snapshots/typed errors.
- `_bmad-output/planning-artifacts/architecture.md:641` - Test organization and sensitive-value corpus.
- `_bmad-output/planning-artifacts/architecture.md:650` - Documentation ownership for behavior matrices and diagnostics/errors docs.
- `_bmad-output/planning-artifacts/architecture.md:667` - Build verification commands and dependency gate ownership.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:281` - FR20 behavior test matrix requirement.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:310` - Cross-cutting NFRs.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:398` - V1 compatibility boundary table.
- `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-11.md:283` - FR20 testable behavior matrix scope.
- `_bmad-output/implementation-artifacts/1-1-adopt-an-auditable-go-module-baseline.md:56` - Existing Story 1.1 command behavior proof scope.
- `_bmad-output/implementation-artifacts/1-1-adopt-an-auditable-go-module-baseline.md:83` - Story 1.1 package boundary and API guardrails.
- `_bmad-output/implementation-artifacts/1-2-publish-the-clean-room-contribution-contract.md:69` - Story 1.2 review findings resolved around clean-room/provenance clarity.
- `_bmad-output/implementation-artifacts/1-2-publish-the-clean-room-contribution-contract.md:213` - Story 1.2 verification and no-runtime-change record.
- Go `errors` package documentation: https://pkg.go.dev/errors
- Go `testing` package documentation: https://pkg.go.dev/testing
- Go testable examples blog: https://go.dev/blog/examples

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `gh search repos "Go CLI behavior matrix clean room" --limit 5` - no reusable candidates returned.
- `gh search code "behavior matrix CLI Go clean-room" --language Markdown --limit 5` - no reusable candidates returned.
- `test -f command/contract_test.go` - failed before implementation, confirming the required executable contract artifact was absent.
- Code review patches applied: clarified provenance classification limits, added URL/access-date guidance for copied or adapted entries, and cleaned stale Dev Agent Record metadata.
- Post-review verification passed: `go test ./...`, `go vet ./...`, `git diff --check HEAD`, the temporary dependency gate, placeholder scan, and secret scan.

### Completion Notes List

- Added the initial cross-surface behavior matrix with explicit future-story hooks and current-scope boundaries.
- Added diagnostics and error contract guidance covering `errors.Is`/`errors.As`, exact source labels, exact fake sensitive-value corpus, and redaction scope.
- Added a black-box command contract test proving `command.NewDefinition` uses explicit caller input rather than ambient `os.Args` or environment state.
- No production runtime code changes were required; the existing Story 1.1 API already satisfied the new executable contract.
- Verification found no forbidden paths, no new dependencies, and no dependency-tooling or CI scope creep.
- Resolved all three code-review patch findings and moved Story 1.3 to done.

### File List

- `_bmad-output/implementation-artifacts/1-3-establish-cross-surface-behavior-contracts.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `command/contract_test.go`
- `docs/behavior-matrices.md`
- `docs/diagnostics-and-errors.md`

### Verification

- `go test ./command` - passed.
- `go test ./...` - passed.
- `go vet ./...` - passed.
- `git diff --check HEAD` - passed.
- Forbidden path check for `tools/depgate/`, `.github/workflows`, `examples`, `internal`, `docs/config-precedence.md`, `docs/compatibility.md`, `docs/testing.md`, `docs/release-checklist.md`, `flags/testdata/fuzz`, and `cmd` - passed.
- `go list -m all` - returned only `github.com/petabytecl/dib`.
- Post-review `go test ./...`, `go vet ./...`, `git diff --check HEAD`, and temporary dependency gate - passed.
