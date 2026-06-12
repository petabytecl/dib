# Testing

This document records local verification commands for Dib maintainers. Run
commands from the repository root with Go 1.26 or newer.

## Lint Gate

Story 6.1 selected a repository-local standard-library lint tool instead of an
external linter action or binary. The short tooling review favored this path
because it adds deterministic maintainability coverage for Go formatting, runs
with the same Go toolchain already pinned by `go.mod`, and does not add
third-party imports, a root `require`, a root `replace`, a `toolchain`
directive, or a go sum file.

Rejected alternatives:

- `golangci-lint` as a GitHub Action or local binary: acceptable only with exact
  action and linter version pins, but unnecessary for the current gate and adds
  a development/CI supply-chain dependency.
- `go install` or shell installer based linting: rejected because it would make
  local and CI behavior depend on mutable remote state unless separately pinned
  and audited.

Local command:

```sh
GOCACHE=/tmp/dib-go-build go run ./tools/lint
```

CI command:

```sh
go run ./tools/lint
```

The lint tool lives under `tools/lint`, imports only the Go standard library,
and is versioned with the repository. Its effective pin is the checked-out Dib
commit plus the Go version selected from `go.mod`; there is no external linter
package or binary version to record. It walks repository Go source while
skipping repository metadata and BMAD/agent artifact directories.

## Coverage Gate

Story 6.2 selected a repository-local standard-library coverage tool instead of
an external coverage action or package. The short tooling review favored this
path because it adds deterministic package-aware coverage evidence for the three
public runtime packages (`command`, `config`, and `flags`), runs with the same
Go toolchain already pinned by `go.mod`, and does not add third-party imports, a
root `require`, a root `replace`, a `toolchain` directive, or a go sum file.
This approach is consistent with the `tools/lint` and `tools/depgate` precedent
already established in this repository.

The coverage tool lives under `tools/coverage`, imports only the Go standard
library, and is versioned with the repository. It invokes `go test -cover` via
`os/exec` (stdlib) and applies per-package thresholds to `command`, `config`,
and `flags`. Each public runtime package that reports coverage below its
threshold causes exit 1; an inability to extract coverage data causes exit 2.

Local command:

```sh
GOCACHE=/tmp/dib-go-build go run ./tools/coverage
```

CI command:

```sh
go run ./tools/coverage
```

Per-package thresholds (floor to nearest 5%, minimum 80%):

- `command`: 85%
- `config`: 85%
- `flags`: 85%

### Tooling Package Exceptions

Tooling packages (`tools/depgate`, `tools/lint`, `tools/coverage`) have a
separate risk profile from public runtime packages. Each exception names the
critical-path tests that preserve confidence:

- `tools/depgate`: exception granted; critical-path tests
  `TestDepgateFixtures`, `TestDepgateReportsEveryViolationDeterministically`,
  and `TestDepgateDisablesWorkspaceMode` in `tools/depgate/main_test.go`
  preserve confidence.
- `tools/lint`: exception granted; critical-path tests
  `TestLintPassesCleanFormattedGoFiles`,
  `TestLintReportsUnformattedFilesDeterministically`, and
  `TestLintCommandRunsFromRepositoryRoot` in `tools/lint/main_test.go` preserve
  confidence.
- `tools/coverage`: exception granted; critical-path tests
  `TestCoveragePassesPackagesMeetingThreshold`,
  `TestCoverageFailsPackagesBelowThreshold`, and
  `TestCoverageCommandRunsFromRepositoryRoot` in `tools/coverage/main_test.go`
  preserve confidence.

## Release Candidate Gates

The release checklist records exact command outcomes for the trust gates used by
release review:

```sh
GOCACHE=/tmp/dib-go-build go run ./tools/lint
GOCACHE=/tmp/dib-go-build go test ./...
GOCACHE=/tmp/dib-go-build go vet ./...
GOCACHE=/tmp/dib-go-build go run ./tools/coverage
GOCACHE=/tmp/dib-go-build go run ./tools/depgate
```

Additional race and fuzz commands are recorded in
`docs/release-checklist.md` when release-candidate evidence requires them.
