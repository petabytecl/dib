# Testing

This document records local verification commands for Dib maintainers. Run
commands from the repository root with Go 1.26 or newer.

## Lint Gate

The lint gate runs [`golangci-lint`](https://golangci-lint.run) over the whole
repository using the strict ruleset in `.golangci.yml` (around sixty linters
derived from the community "golden" configuration). It supersedes the earlier
repository-local standard-library formatter: `golangci-lint` enforces
`gofumpt`/`gofmt` formatting plus correctness, complexity, and style analysis.

`golangci-lint` runs only as an external, pinned CI binary and is never imported
by the module, so the repository keeps its dependency-free posture: the gate adds
no third-party imports, no root `require`, no root `replace`, no `toolchain`
directive, and no go sum file. The `depguard` linter enforces that invariant by
allowing imports only from the Go standard library and the
`github.com/petabytecl/dib` module itself, so adding any external dependency must
be a deliberate, reviewed change to `.golangci.yml`.

The linter version is pinned: CI installs `golangci-lint` `v2.10.1` through the
maintainers' `golangci/golangci-lint-action@v9` action, and local runs use the
same version. The effective pin is that explicit version plus the Go version
selected from `go.mod`.

Rejected alternatives:

- Floating linter versions (a `latest` or `stable` release channel, or an
  unpinned `go install`): rejected because they make local and CI behavior depend
  on mutable remote state. The gate always records an exact `golangci-lint`
  version.

Local command:

```sh
golangci-lint run
```

CI step (pinned in `.github/workflows/ci.yml`):

```yaml
- name: Lint
  uses: golangci/golangci-lint-action@v9
  with:
    version: v2.10.1
```

## Coverage Gate

Story 6.2 selected a repository-local standard-library coverage tool instead of
an external coverage action or package. The short tooling review favored this
path because it adds deterministic package-aware coverage evidence for the public
runtime packages, runs with the same Go toolchain already pinned by `go.mod`,
and does not add third-party imports, a root `require`, a root `replace`, a
`toolchain` directive, or a go sum file. This approach is consistent with the
`tools/depgate` precedent already established in this repository.
Story 7.4 extended the gate to cover the `cli` package as the fourth public
runtime package.

The coverage tool lives under `tools/coverage`, imports only the Go standard
library, and is versioned with the repository. It invokes `go test -cover` via
`os/exec` (stdlib) and applies per-package thresholds to `command`, `config`,
`flags`, and `cli`. Each public runtime package that reports coverage below its
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
- `cli`: 85%

### Tooling Package Exceptions

Tooling packages (`tools/depgate`, `tools/coverage`) have a separate risk
profile from public runtime packages. Each exception names the critical-path
tests that preserve confidence:

- `tools/depgate`: exception granted; critical-path tests
  `TestDepgateFixtures`, `TestDepgateReportsEveryViolationDeterministically`,
  and `TestDepgateDisablesWorkspaceMode` in `tools/depgate/main_test.go`
  preserve confidence.
- `tools/coverage`: exception granted; critical-path tests
  `TestCoveragePassesPackagesMeetingThreshold`,
  `TestCoverageFailsPackagesBelowThreshold`, and
  `TestCoverageCommandRunsFromRepositoryRoot` in `tools/coverage/main_test.go`
  preserve confidence.

## Release Candidate Gates

The release checklist records exact command outcomes for the trust gates used by
release review:

```sh
golangci-lint run
GOCACHE=/tmp/dib-go-build go test ./...
GOCACHE=/tmp/dib-go-build go vet ./...
GOCACHE=/tmp/dib-go-build go run ./tools/coverage
GOCACHE=/tmp/dib-go-build go run ./tools/depgate
```

Additional race and fuzz commands are recorded in
`docs/release-checklist.md` when release-candidate evidence requires them.
