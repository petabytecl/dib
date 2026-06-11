---
baseline_commit: 6652d09614adafe322f9896f8e53763b37eacc13
created: "2026-06-11T14:42:41-04:00"
---

# Story 1.4: Enforce the Standard-Library Dependency Gate

Status: done

## Story

As a dependency-policy reviewer,
I want a repeatable repository gate that fails on non-standard-library imports,
so that Dib's zero-runtime-dependency claim is enforced before feature implementation scales.

## Requirements Trace

- FR21: Enforce the runtime dependency rule with a repository check that fails on non-standard-library imports.
- NFR1, NFR6, NFR9: runtime packages, tests, examples, and repository tooling stay standard-library-only; verification remains table-driven and uses Go 1.26+.
- UJ-4: Dependency-policy reviewers must be able to inspect the module graph and trust Dib before approving it for internal tools.

## Acceptance Criteria

1. Given Dib's runtime, tests, examples, and repository tooling must remain standard-library-only unless the architecture changes, when `tools/depgate/` is implemented, then `go run ./tools/depgate` inspects all non-tool packages included by `go test ./...`, including package tests and `examples/` packages, and it fails when any inspected package imports a non-standard-library package.
2. Given `tools/depgate/` is repository tooling rather than runtime library code, when the dependency gate inspects tool packages, then it also verifies tool packages use only the Go standard library unless the architecture is updated, and it does not create an import path from runtime packages into `tools/depgate/`.
3. Given dependency failures need to be actionable, when the gate finds a non-standard-library import, then the diagnostic identifies the package and offending import path, and it exits non-zero without hiding other detected dependency violations.
4. Given the first scaffold may temporarily use the architecture-approved `go list` command, when `tools/depgate/` exists, then documentation and CI guidance make `go run ./tools/depgate` the required dependency check, and the temporary `go list` command is no longer accepted as release-candidate evidence.
5. Given the gate protects FR21 and NFR1, when verification runs, then `go test ./...`, `go vet ./...`, and `go run ./tools/depgate` pass, and tests for `tools/depgate/` cover at least a passing stdlib-only fixture and a failing non-stdlib import fixture.

## Tasks / Subtasks

- [x] Confirm the live checkout before editing (AC: 1-5)
  - [x] Verify `tools/depgate/` does not already exist. If it exists, update it in place instead of replacing user work.
  - [x] Verify the repo is on `main` with Story 1.1, Story 1.2, and Story 1.3 completed in `sprint-status.yaml`.
  - [x] Preserve the existing runtime packages: `command/`, `flags/`, and `config/` must not import `tools/depgate/` or any new dependency.

- [x] Implement the isolated dependency gate command (AC: 1, 2, 3)
  - [x] Create `tools/depgate/main.go` as a package `main` command using only Go standard-library imports.
  - [x] Run the Go tool through `os/exec` with `go list -deps -test -e -json -buildvcs=false ./...` from the repository root so package tests and future `examples/` packages are included in the analyzed graph.
  - [x] Decode the JSON stream with `encoding/json.Decoder`; do not parse `go list` output with ad hoc line splitting.
  - [x] Treat packages as allowed only when `Standard == true` or `Module != nil && Module.Main == true`.
  - [x] Treat any package with `Standard == false` and not in the main module as a dependency violation, including unresolved external imports surfaced through `-e` package errors.
  - [x] Collect all violations, sort them deterministically, print every violation, and exit non-zero when any violation exists.
  - [x] Keep execution errors and JSON decode errors distinct from dependency violations in diagnostics.

- [x] Add focused depgate tests and fixtures (AC: 1, 2, 3, 5)
  - [x] Create `tools/depgate/main_test.go` with table-driven tests in package `main`.
  - [x] Create a passing stdlib-only fixture under `tools/depgate/testdata/stdlib-only/`.
  - [x] Create a failing non-stdlib fixture under `tools/depgate/testdata/non-stdlib-test-import/` where a package test imports a fake external path such as `example.com/external/notstdlib`.
  - [x] Run fixture checks with `-buildvcs=false` so temp copied fixtures do not fail because they are outside a VCS checkout.
  - [x] Assert the failing fixture exits non-zero and its output identifies both the package or test context and the offending import path.
  - [x] Assert the passing fixture exits successfully and does not report false violations for standard-library packages.

- [x] Update contributor dependency-gate guidance (AC: 4)
  - [x] Update `CONTRIBUTING.md` so local verification requires `go run ./tools/depgate` after `go test ./...` and `go vet ./...`.
  - [x] Remove or rewrite the old conditional wording that allowed the temporary ad hoc dependency check when no dedicated gate exists.
  - [x] Do not create `.github/workflows/ci.yml`; Story 1.5 owns CI wiring and must use `go run ./tools/depgate`.
  - [x] Do not create `docs/release-checklist.md`; Story 1.5 and Epic 5 own release-candidate evidence consolidation.

- [x] Verify the story output (AC: 1-5)
  - [x] Run `go test ./...`.
  - [x] Run `go vet ./...`.
  - [x] Run `go run ./tools/depgate`.
  - [x] Confirm `go.mod` still has no `require`, `replace`, or `toolchain` directives unless the architecture has changed.
  - [x] Confirm no runtime, test, fixture, example, or tool dependency was added.
  - [x] Confirm no root facade package, `/cmd`, CI workflow, release checklist, compatibility docs, config precedence docs, examples, or broad `internal/` package was created.
  - [x] Record exact commands and outcomes in the Dev Agent Record.

### Review Findings

- [x] [Review][Patch] Disable workspace mode while running `go list` [tools/depgate/main.go:61]
- [x] [Review][Patch] Process `DepsErrors` as dependency violations [tools/depgate/main.go:100]
- [x] [Review][Patch] Report resolved external imports with the main-module importer context [tools/depgate/main.go:100]

## Dev Notes

### ATDD Artifacts

- Checklist: `_bmad-output/test-artifacts/atdd-checklist-1-4-enforce-the-standard-library-dependency-gate.md`
- Backend command tests: `tools/depgate/main_test.go`
- Fixtures: `tools/depgate/testdata/stdlib-only/`, `tools/depgate/testdata/non-stdlib-test-import/`

### Source Discovery

- Loaded sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`.
- Loaded epics: `_bmad-output/planning-artifacts/epics.md`.
- Loaded PRD workspace: `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md`, `addendum.md`, `reconcile-brief.md`, and `review-rubric.md`.
- Loaded architecture: `_bmad-output/planning-artifacts/architecture.md`.
- Loaded readiness report: `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-11.md`.
- Loaded previous stories: `_bmad-output/implementation-artifacts/1-1-adopt-an-auditable-go-module-baseline.md`, `_bmad-output/implementation-artifacts/1-2-publish-the-clean-room-contribution-contract.md`, and `_bmad-output/implementation-artifacts/1-3-establish-cross-surface-behavior-contracts.md`.
- Loaded current source and docs: `go.mod`, `command/`, `flags/`, `config/`, `CONTRIBUTING.md`, `docs/behavior-matrices.md`, `docs/diagnostics-and-errors.md`, `docs/clean-room-policy.md`, and `docs/provenance-log.md`.
- No UX document, `project-context.md`, `CLAUDE.md`, or local `MEMORY.md` was discovered in the repo.

### Current Repository State

- Branch at story creation: `main`.
- Baseline commit at story creation: `6652d09614adafe322f9896f8e53763b37eacc13` (`feat: establish Go module baseline and contracts`).
- `main` is ahead of `origin/main` by one commit at story creation; do not assume the previous Story 1.1-1.3 work has been pushed.
- `sprint-status.yaml` marks Story 1.1, Story 1.2, and Story 1.3 as done; Story 1.4 is the first backlog story.
- Existing runtime package files:
  - `go.mod` declares module `github.com/petabytecl/dib` with `go 1.26` and no dependency directives.
  - `command/definition.go` exposes `Definition`, `NewDefinition`, `Name`, and `NameError`; use `NewDefinition` for validated definitions.
  - `command/definition_test.go` and `command/contract_test.go` use standard-library-only tests from `package command_test`.
  - `flags/doc.go` and `config/doc.go` are package documentation only.
- Existing docs:
  - `CONTRIBUTING.md` currently says to run `go test ./...` and `go vet ./...`, with conditional dependency-gate wording. This story must make `go run ./tools/depgate` the required local dependency check.
  - `docs/behavior-matrices.md` and `docs/diagnostics-and-errors.md` establish cross-surface contracts. Do not expand them for dependency tooling unless a verification issue proves it necessary.

### Architecture Guardrails

- `tools/depgate/` is repository tooling, not an importable library package. It must remain isolated from `command/`, `flags/`, and `config/`. [Source: `_bmad-output/planning-artifacts/architecture.md:550`, `_bmad-output/planning-artifacts/architecture.md:573`]
- The dependency gate must inspect all non-tool Go packages included by `go test ./...`, including package tests and `examples/` packages, and must also verify tool packages remain standard-library-only. [Source: `_bmad-output/planning-artifacts/architecture.md:647`, `_bmad-output/planning-artifacts/architecture.md:670`]
- Runtime code, tests, and runnable examples must stay standard-library-only unless the architecture is updated. [Source: `_bmad-output/planning-artifacts/architecture.md:416`]
- The dedicated dependency gate becomes the documented local/CI dependency check once it exists. Agents must not create alternate local-only dependency checks. [Source: `_bmad-output/planning-artifacts/architecture.md:425`, `_bmad-output/planning-artifacts/architecture.md:427`]
- The temporary Story 1.1 `go list` command was allowed only before `tools/depgate/` existed. It must not be used as release-candidate evidence after this story. [Source: `_bmad-output/planning-artifacts/architecture.md:814`, `_bmad-output/planning-artifacts/architecture.md:822`]
- Do not create `.github/workflows/ci.yml`; Story 1.5 owns CI execution and must call `go run ./tools/depgate`. [Source: `_bmad-output/planning-artifacts/epics.md:334`, `_bmad-output/planning-artifacts/architecture.md:632`]
- Do not create `docs/release-checklist.md`; Story 1.5 and Epic 5 own release-candidate evidence consolidation. [Source: `_bmad-output/planning-artifacts/epics.md:344`, `_bmad-output/planning-artifacts/architecture.md:672`]
- No root facade package, package-global helpers, `/cmd` scaffold, Docker/Kubernetes files, service runtime, or deployment config belongs in this story. [Source: `_bmad-output/planning-artifacts/architecture.md:564`, `_bmad-output/planning-artifacts/architecture.md:633`]

### Expected File Changes

Expected new files for this story:

```text
tools/depgate/main.go
tools/depgate/main_test.go
tools/depgate/testdata/stdlib-only/go.mod
tools/depgate/testdata/stdlib-only/pkg/pkg.go
tools/depgate/testdata/stdlib-only/pkg/pkg_test.go
tools/depgate/testdata/non-stdlib-test-import/go.mod
tools/depgate/testdata/non-stdlib-test-import/pkg/pkg.go
tools/depgate/testdata/non-stdlib-test-import/pkg/pkg_test.go
```

Expected updates:

```text
CONTRIBUTING.md
_bmad-output/implementation-artifacts/1-4-enforce-the-standard-library-dependency-gate.md
_bmad-output/implementation-artifacts/sprint-status.yaml
```

Do not create these in this story unless a failing verification gate proves they are strictly necessary:

```text
.github/workflows/ci.yml
docs/release-checklist.md
docs/compatibility.md
docs/config-precedence.md
docs/testing.md
examples/
internal/
cmd/
```

### Implementation Guidance

- Prefer a small `package main` with unexported helper functions so tests can exercise behavior without shelling out to the final binary for every assertion.
- Use `exec.CommandContext` to run `go list`. Set `cmd.Dir` explicitly in helper options so tests can point at `testdata` fixture modules.
- Use arguments equivalent to:

```text
go list -deps -test -e -json -buildvcs=false ./...
```

- `-deps` includes dependencies; `-test` includes test binaries and test-only imports; `-e` keeps package-error data in JSON so unresolved external imports can still be reported; `-buildvcs=false` avoids unrelated VCS stamping failures in copied temp fixtures.
- Decode the JSON stream into a minimal struct containing only fields the gate needs: `ImportPath`, `Standard`, `Module.Main`, `Error.ImportStack`, `Error.Pos`, `Error.Err`, and `DepsErrors`.
- For each decoded package, allow it when:
  - `Standard == true`, or
  - `Module != nil && Module.Main == true`.
- Flag it when:
  - `Standard == false` and it is not in the main module, including unresolved imports with no module metadata.
- Diagnostics should be deterministic and line-oriented. A good shape is:

```text
non-standard import: package=<source package or test context> import=<offending import path>
```

- If `Error.ImportStack` is present, use the first stack entry as the package/test context. If not, use the decoded package import path as context.
- Do not stop at the first violation. Collect, sort, and print all violations before returning an error exit code.
- Keep tool execution errors separate from dependency violations. A malformed JSON stream or a missing `go` binary is an execution failure, not a dependency-policy failure.
- Do not use `golang.org/x/tools/go/packages`, shell pipelines, or third-party assertion libraries; this story's tooling and tests must remain standard-library-only.

### Previous Story Intelligence

- Story 1.1 established the Go module, public package boundaries, minimal command definition behavior, and the temporary `go list` dependency check. That temporary check is no longer sufficient once `tools/depgate/` exists.
- Story 1.1 code review clarified that direct typed errors can satisfy `errors.As`; do not over-document unsupported wrapping behavior in depgate diagnostics.
- Story 1.2 established clean-room and provenance policy. Do not copy code, tests, examples, fixtures, names, or structure from external dependency-gate tools.
- Story 1.3 explicitly reserved `tools/depgate/` for Story 1.4 and prohibited dependency tooling in Story 1.3. Preserve that sequencing discipline.
- Story 1.3 verification confirmed no forbidden Story 1.4 paths existed before this story.

### Latest Technical Context

- Local Go version at story creation: `go version go1.26.4 linux/amd64`.
- Local `go help list` for Go 1.26.4 documents `Package.Standard`, `Package.Module`, `Package.Imports`, `Package.Deps`, `Package.TestImports`, `Package.XTestImports`, `Package.Error`, and `Package.DepsErrors` in `go list` JSON output.
- Local `go help list` documents that `-deps` lists dependencies and marks dependency-only packages with `DepOnly`.
- Local `go help list` documents that `-test` reports test binaries and the packages rebuilt for tests, which is why this story requires `-test` for package tests and future examples.
- A local probe confirmed that `go list -deps -test -e -json -buildvcs=false` can surface an unresolved fake external test import as a JSON package with `ImportPath: "example.com/external/notstdlib"` and package-error context, without adding the fake module to `go.mod`.
- Official Go source for `cmd/go/internal/list` documents the same `go list` package fields and `-deps` / `-test` behavior. Use it only as factual documentation; do not copy Go source code or examples. Source: https://go.dev/src/cmd/go/internal/list/list.go

### Testing Standards

- Use only the standard `testing` package and standard-library helpers.
- Keep depgate tests deterministic and table-driven where practical.
- Put fixtures under `tools/depgate/testdata/` so `go test ./...` does not treat them as repository packages.
- Fixture modules should use `go 1.26` and must not add dependencies to the root `go.mod`.
- Tests should assert exit status and stable diagnostic substrings, not exact full stderr/stdout formatting unless the output is intentionally part of the contract.
- Required verification for this story: `go test ./...`, `go vet ./...`, and `go run ./tools/depgate`.

### Security And Quality Checks

- Do not hardcode secrets, tokens, credentials, host-specific paths, or network-dependent test inputs.
- The fake failing import path must be a clearly non-real package path, such as `example.com/external/notstdlib`, and must remain inside `tools/depgate/testdata/`.
- Do not invoke `go get`, `go mod tidy`, or any network-fetching workflow to create the failing fixture.
- Do not mutate `go.mod` beyond what already exists. The expected root module remains only:

```text
module github.com/petabytecl/dib

go 1.26
```

- Keep files focused and small. If `tools/depgate/main.go` starts to grow, extract small unexported helpers in the same file or another focused file under `tools/depgate/`; do not create a broad shared `internal/` package for a single tool.
- Error messages should not leak environment details beyond the package/import path and any Go tool error needed to diagnose a failed dependency check.

### References

- `_bmad-output/planning-artifacts/epics.md:299` - Story 1.4 source story.
- `_bmad-output/planning-artifacts/epics.md:309` - Runtime, tests, examples, and tooling standard-library-only acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:314` - Tool-package isolation acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:319` - Actionable diagnostics acceptance criteria.
- `_bmad-output/planning-artifacts/epics.md:324` - Dedicated depgate replaces temporary `go list` evidence.
- `_bmad-output/planning-artifacts/epics.md:329` - Required verification and fixture coverage.
- `_bmad-output/planning-artifacts/epics.md:111` - Core PR gates include tests, vet, and dependency gate.
- `_bmad-output/planning-artifacts/architecture.md:416` - Standard-library-only runtime, tests, and runnable examples.
- `_bmad-output/planning-artifacts/architecture.md:427` - Dependency gate is the documented local/CI dependency check.
- `_bmad-output/planning-artifacts/architecture.md:550` - Expected `tools/depgate/` file location.
- `_bmad-output/planning-artifacts/architecture.md:573` - `tools/depgate/` is isolated repository tooling.
- `_bmad-output/planning-artifacts/architecture.md:647` - Dependency enforcement ownership.
- `_bmad-output/planning-artifacts/architecture.md:670` - `go run ./tools/depgate` build-process gate.
- `_bmad-output/planning-artifacts/architecture.md:822` - Temporary dependency check no longer accepted after depgate exists.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:290` - FR21 runtime dependency rule.
- `_bmad-output/planning-artifacts/prds/prd-dib-2026-06-10/prd.md:310` - NFR1 runtime dependency ceiling.
- `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-11.md:107` - FR21 testable scope.
- `_bmad-output/planning-artifacts/implementation-readiness-report-2026-06-11.md:319` - Dependency gate appears in Story 1.4.
- `_bmad-output/implementation-artifacts/1-3-establish-cross-surface-behavior-contracts.md:130` - Story 1.4 owns `tools/depgate/`.
- Go `cmd/go/internal/list` source documentation: https://go.dev/src/cmd/go/internal/list/list.go

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `go test ./tools/depgate -run TestDepgateAllowsStdlibOnlyFixture -count=1` - RED confirmed before implementation; failed because `tools/depgate` had no non-test Go files.
- `go test -count=1 ./...` - PASS after implementation.
- `go vet ./...` - PASS after implementation.
- `go run ./tools/depgate` - PASS with no dependency violations.
- `git diff --check` - PASS.
- `rg -n "require |replace |toolchain" go.mod tools/depgate/testdata/*/go.mod || true` - no dependency directives found.
- `rg -n "tools/depgate|github.com/petabytecl/dib/tools/depgate" command flags config go.mod || true` - no runtime imports of the depgate tool.
- `go test -count=1 ./tools/depgate` - PASS after code-review fixes.
- `go test -count=1 ./...` - PASS after code-review fixes.
- `go vet ./...` - PASS after code-review fixes.
- `go run ./tools/depgate` - PASS after code-review fixes.
- `git diff --check` - PASS after code-review fixes.
- Secret-pattern scan over source, docs, tools, module files, and BMAD artifacts - no secret-like matches found.
- `rg -n "require |replace |toolchain" go.mod tools/depgate/testdata/*/go.mod || true` - only deliberate local fixture `require`/`replace` entries found under `tools/depgate/testdata/resolved-nonstdlib-import/`.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Implemented `tools/depgate` as isolated standard-library-only repository tooling.
- The gate executes `go list -deps -test -e -json -buildvcs=false ./...`, decodes the JSON stream, allows standard-library and main-module packages, and reports every non-main non-stdlib package deterministically.
- Added active table-driven fixture tests for stdlib-only and non-stdlib test import modules, plus direct coverage for execution failure diagnostics.
- Updated `CONTRIBUTING.md` so `go run ./tools/depgate` is a required local verification command.
- Verified no root dependency directives, forbidden scaffold directories, runtime imports into `tools/depgate`, CI workflow, release checklist, compatibility docs, config precedence docs, examples, or broad `internal/` package were added.
- Addressed code-review findings by disabling Go workspace mode during dependency scans, processing `DepsErrors`, and reporting resolved non-standard imports with direct main-module importer context.
- Added resolved external import and workspace-mode regression coverage without changing the root module dependency graph.

### File List

- `CONTRIBUTING.md`
- `_bmad-output/implementation-artifacts/1-4-enforce-the-standard-library-dependency-gate.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/test-artifacts/atdd-checklist-1-4-enforce-the-standard-library-dependency-gate.md`
- `_bmad-output/test-artifacts/tmp/tea-atdd-api-tests-2026-06-11T14-54-34-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-e2e-tests-2026-06-11T14-54-34-0400.json`
- `_bmad-output/test-artifacts/tmp/tea-atdd-summary-2026-06-11T14-54-34-0400.json`
- `tools/depgate/main.go`
- `tools/depgate/main_test.go`
- `tools/depgate/testdata/non-stdlib-test-import/go.mod`
- `tools/depgate/testdata/non-stdlib-test-import/pkg/pkg.go`
- `tools/depgate/testdata/non-stdlib-test-import/pkg/pkg_test.go`
- `tools/depgate/testdata/resolved-external/go.mod`
- `tools/depgate/testdata/resolved-external/resolved.go`
- `tools/depgate/testdata/resolved-nonstdlib-import/go.mod`
- `tools/depgate/testdata/resolved-nonstdlib-import/pkg/pkg.go`
- `tools/depgate/testdata/resolved-nonstdlib-import/pkg/pkg_test.go`
- `tools/depgate/testdata/stdlib-only/go.mod`
- `tools/depgate/testdata/stdlib-only/pkg/pkg.go`
- `tools/depgate/testdata/stdlib-only/pkg/pkg_test.go`

### Change Log

- 2026-06-11: Implemented Story 1.4 dependency gate, fixture coverage, contributor guidance, and verification evidence.
- 2026-06-11: Addressed code-review findings and marked Story 1.4 done.
