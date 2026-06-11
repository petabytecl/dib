# Story 1.1: Adopt an Auditable Go Module Baseline

Status: ready-for-dev

## Story

As a Go library adopter,
I want Dib to start as a minimal, standard-library-only Go module with visible package boundaries,
so that I can inspect the foundation before trusting later command, flag, and config behavior.

## Requirements Trace

- FR20: Provide behavior test matrices; this story starts executable behavior proof with at least one standard-library-only package test or runnable example.
- FR21: Enforce the runtime dependency rule; this story establishes the module and initial packages without non-standard-library imports.
- NFR1, NFR2, NFR5, NFR6, NFR9: standard-library runtime, explicit instances, no ambient process state, testability, and Go 1.26+ baseline.

## Acceptance Criteria

1. Given a fresh checkout without an initialized Go module, when the module baseline is created, then `go.mod` declares module `github.com/petabytecl/dib` with `go 1.26`, and no `toolchain` directive is added unless the architecture is updated.
2. Given the architecture-defined package boundaries, when the initial source tree is created, then `command/`, `flags/`, and `config/` each contain package documentation or a minimal compilable package file, and the repository does not introduce a root facade package, package-global command/flag/config helpers, or a `/cmd` scaffold.
3. Given the baseline module exists, when verification runs, then `go test ./...` and `go vet ./...` pass, and at least one standard-library-only package test or runnable example proves the module is not only empty structure.
4. Given later stories will add behavior to reusable definitions and snapshots, when baseline docs or package comments describe the implementation direction, then they state that Dib favors explicit instances, caller-owned inputs, immutable definitions, per-run snapshots, typed errors, and no ambient process state.

## Tasks / Subtasks

- [ ] Confirm the pre-implementation checkout state (AC: 1, 2)
  - [ ] Verify whether `go.mod`, `command/`, `flags/`, or `config/` already exist before editing. At story creation time they were not present; preserve any user-created files if the checkout changes before implementation.
  - [ ] Keep all implementation files small and capability-focused. Do not add broad shared helpers or `internal/` packages unless two concrete call sites prove the need.

- [ ] Initialize the Go module baseline (AC: 1)
  - [ ] Run or equivalent: `go mod init github.com/petabytecl/dib`.
  - [ ] Explicitly set the language version to `go 1.26` after initialization. Go 1.26's `go mod init` may create a lower default `go` line, so verify the final file instead of trusting the generated default.
  - [ ] Ensure `go.mod` contains no `toolchain`, `require`, or `replace` directives for this story unless an architecture update explicitly approves them.

- [ ] Create the public package boundary scaffold (AC: 2, 4)
  - [ ] Create `command/`, `flags/`, and `config/` as public capability packages with package comments in `doc.go` or minimal compilable package files.
  - [ ] Package comments must make the architecture direction explicit: explicit instances, caller-owned inputs, immutable definitions, per-run snapshots, typed errors, and no ambient process state.
  - [ ] Do not create a root public facade package, a `/cmd` app scaffold, package-level default command/flag/config instances, or package-global helper APIs.
  - [ ] Keep `flags/` independently usable without `command/` or `config/`; keep `command/` independent from `config/`.

- [ ] Add one minimal standard-library-only behavior proof in `command/` (AC: 3, 4)
  - [ ] Implement a non-executing command definition shape with name validation. Exact exported identifiers may settle during implementation, but the public behavior must let a caller construct a valid command definition, read back its stable name, and receive an inspectable error for an empty or whitespace-only name.
  - [ ] Keep the behavior narrow: no callback invocation, no command routing, no flag parsing, no config binding, no process args/env/stdout access, and no mutable package-global state.
  - [ ] Use returned values and errors. If a caller needs to inspect an error, expose an error value or type that works with `errors.Is` or `errors.As`; do not make string matching the programmatic contract.
  - [ ] Add at least one `testing` package test from the external-caller perspective, such as `package command_test`, and add a runnable `Example...` test where practical. The test/example must import only the Go standard library and local Dib packages.

- [ ] Verify the baseline locally (AC: 1, 3)
  - [ ] Run `go test ./...`.
  - [ ] Run `go vet ./...`.
  - [ ] Until `tools/depgate/` exists, run the temporary architecture-approved dependency check:
    ```bash
    go list -deps -f '{{if and (not .Standard) (not .Module.Main)}}{{.ImportPath}}{{end}}' ./... | sed '/^$/d'
    ```
    The command should produce no output for this story.
  - [ ] Record the exact verification commands and outcomes in the Dev Agent Record.

## Dev Notes

### Source Discovery

- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded PRD: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded readiness report: `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-11.md`.
- No UX design document was discovered in the configured planning artifacts.
- No `project-context.md` was found from the configured persistent facts glob.

### Current Repository State

- Branch at story creation: `main`.
- Recent commits are documentation/setup only: `15ef638 docs(bmad): add architecture decision document`, `99a53ef docs: initialize dib planning workspace`.
- At story creation time, `go.mod`, `command/`, `flags/`, and `config/` were absent. This story is expected to create new implementation files rather than update existing source.
- Existing untracked BMAD artifacts were present under `_bmad-output/`; avoid modifying unrelated planning artifacts.

### Architecture Guardrails

- Dib is a Go importable library/toolkit, not a CLI application starter. Do not add binary entry points, generated shell completion, man pages, Docker/Kubernetes config, service runtime, or deployment config.
- Use no external starter template and do not copy template structure, source, tests, comments, examples, fixtures, or file organization from inspiration projects.
- Runtime packages must import only the Go standard library. Tests should also stay standard-library-only unless a later architecture decision explicitly approves otherwise.
- Public package communication must happen through explicit values, returned snapshots, and typed errors. No package should communicate through global registries, process args, process env, stdout/stderr, hidden caches, or default singletons.
- Callback handling is deferred. This story must not add callback invocation behavior.
- Exact exported API identifiers are allowed to settle during implementation, but they must stay minimal and aligned with this story's narrow behavior proof. Do not create a wide API surface to anticipate later parser, router, or config stories.

### Package Structure Guidance

Expected new files are limited to the baseline module and first package scaffold, for example:

```text
go.mod
command/doc.go
command/definition.go
command/definition_test.go
flags/doc.go
config/doc.go
```

Equivalent names are acceptable if they remain cohesive, small, and standard-library-only, but the behavior proof belongs in `command/` for this story. Do not add `README`-driven behavior docs, clean-room policy docs, CI, release checklist, `tools/depgate/`, parser fuzz seeds, config precedence docs, or migration examples in this story unless they are strictly needed to satisfy the acceptance criteria; those are owned by later Epic 1 and Epic 5 stories.

### Latest Go Tooling Notes

- Official Go downloads list Go 1.26.4 as the current stable Go 1.26 release family as of this story creation date.
- Official Go module docs say `go mod init` creates a `go.mod` at the module root and its argument is the module path; the module path should be the repository location when possible.
- The `go` directive is the required minimum Go version for modules at Go 1.21 and newer, while the `toolchain` directive is a suggested toolchain and only takes effect for the main module when the default toolchain is older.
- Go 1.26 release notes state that `go mod init` defaults new modules to a lower `go` line, so implementation must explicitly set and verify `go 1.26`.

### Testing Standards

- Use the standard `testing` package only.
- Prefer table-driven tests for any validation behavior introduced here, even if the table is small.
- Assertions should inspect returned values/errors, not stdout/stderr or diagnostic strings.
- The first behavior proof should demonstrate explicit construction and no ambient state. It does not need to prove all cross-surface immutability, snapshot, typed-error, provenance, or redaction contracts; those acceptance hooks are expanded in Story 1.3 and later package stories.

### Security And Quality Checks

- No secrets, tokens, generated credentials, environment-specific paths, or network calls belong in runtime packages.
- Treat all future CLI args, env values, JSON config, readers, and lookup functions as untrusted boundary inputs; this story should establish comments and code direction consistent with that rule.
- Error text must not leak sensitive values. This story should not add sensitive-value handling, but package comments should avoid implying raw value dumps are acceptable.
- Preserve immutability expectations: do not expose mutable internals, do not mutate caller-owned inputs, and prefer returning new values over modifying existing values in place.

### References

- `_bmad-output/planning-artifacts/epics.md:205` - Story 1.1 source story and acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:86` - Additional implementation requirements for module bootstrap, package boundaries, no root facade, tests, and dependency gate.
- `_bmad-output/planning-artifacts/architecture.md:92` - Standard Go module bootstrap selected with no external starter template.
- `_bmad-output/planning-artifacts/architecture.md:124` - Initialization command and `go 1.26` / no `toolchain` guidance.
- `_bmad-output/planning-artifacts/architecture.md:134` - First implementation story gates.
- `_bmad-output/planning-artifacts/architecture.md:556` - Public package boundaries and deferred callback behavior.
- `_bmad-output/planning-artifacts/architecture.md:628` - File organization, test organization, and development workflow gates.
- `_bmad-output/planning-artifacts/architecture.md:797` - First implementation priority and temporary dependency-gate allowance.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:72` - Product constraints and guardrails.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:281` - FR20 behavior test evidence.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:290` - FR21 runtime dependency rule.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:308` - Cross-cutting NFRs.
- `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-11.md:310` - Greenfield check confirming Story 1.1 scope.
- Go downloads: https://go.dev/dl/
- Go managing dependencies: https://go.dev/doc/modules/managing-dependencies
- Go go.mod reference: https://go.dev/doc/modules/gomod-ref
- Go modules reference: https://go.dev/ref/mod
- Go 1.26 release notes: https://go.dev/doc/go1.26

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.

### File List
