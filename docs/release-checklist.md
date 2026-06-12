# Release Checklist

Use this checklist for each Dib Go module tag. It records release-candidate evidence; it is not a claim that a release is ready until every required item is filled in and reviewed.

## Release Identity

- Go module tag: `v0.1.0`
- Exact commit: `d5ce41ce693413b88df95e644eb4358702ae205e`
- Owner: Coto
- Date: 2026-06-12
- Reviewer: Release reviewer

## Go Version Alignment

Release is blocked if `go.mod`, `.github/workflows/ci.yml`, release guidance, or user-facing docs disagree about the supported Go version.

- `go.mod` version: `go 1.26`
- CI `actions/setup-go` source: `go-version-file: go.mod`
- Release guidance version: `docs/release-notes-v0.md` states Go 1.26+ and v0 experimental API status.
- Documentation version references: `go.mod`, `.github/workflows/ci.yml`, `docs/behavior-matrices.md`, and `docs/release-notes-v0.md` align on Go 1.26 or Go 1.26+.
- Drift review result: No drift found; local evidence used `go version go1.26.4 linux/amd64`.

## CI Trust Gates

CI failures block tagging. Record the exact command outcome for each required gate.

- `go test ./...`: PASS on 2026-06-12 with `GOCACHE=/tmp/dib-go-build go test ./...`.
- `go vet ./...`: PASS on 2026-06-12 with `GOCACHE=/tmp/dib-go-build go vet ./...`.
- `go run ./tools/depgate`: PASS on 2026-06-12 with `GOCACHE=/tmp/dib-go-build go run ./tools/depgate`; output was empty and exit code was 0.
- Workflow file: `.github/workflows/ci.yml`
- Runner image: `ubuntu-24.04`
- `actions/checkout` version: `actions/checkout@v6`
- `actions/setup-go` version: `actions/setup-go@v6` with `go-version-file: go.mod`

## Release-Candidate Gates

- `go test -race ./...`: PASS on 2026-06-12 with `GOCACHE=/tmp/dib-go-build go test -race ./...`.
- Parser fuzz evidence, when parser behavior changed:
  - `go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags`: PASS on 2026-06-12 with `GOCACHE=/tmp/dib-go-build go test -fuzz='^FuzzParse$' -fuzztime=5s ./flags`.
  - `go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags`: PASS on 2026-06-12 with `GOCACHE=/tmp/dib-go-build go test -fuzz='^FuzzParseBoundary$' -fuzztime=5s ./flags`.
  - `go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags`: PASS on 2026-06-12 with `GOCACHE=/tmp/dib-go-build go test -fuzz='^FuzzParseShortGroups$' -fuzztime=5s ./flags`.
- Docs/examples evidence input: `docs/behavior-matrices.md` records Story 5.4 release-candidate evidence, consolidates Story 5.3 adoption evidence, and links package tests plus Story 5.2 migration examples.
- Provenance evidence input: `docs/provenance-log.md`; Story 5.4 records final
  provenance review outcome: no new external reference material was used for
  release evidence or release notes; no provenance entry was required.
- Compatibility evidence input: `docs/compatibility.md` and
  `docs/behavior-matrices.md`; Story 5.4 compatibility review found Dib still
  framed as a clean-room native Go API, not source-compatible or drop-in.
- Migration evidence: Story 5.2 example pointers live in `examples/migration/`; `GOCACHE=/tmp/dib-go-build go test ./examples/migration` passed on 2026-06-12.
- V0 experimental API status: `docs/release-notes-v0.md` states that v0 may change APIs while preserving correctness, redaction, clean-room, dependency, and release-gate expectations.

## Standard-Library Dependency Evidence

- Root `go.mod` contains no `require`, `replace`, or `toolchain` directives: PASS; `rg -n "^(require|replace|toolchain)\b" go.mod` returned no output on 2026-06-12.
- Root `go.sum` absent: PASS; `test ! -e go.sum` exited 0 on 2026-06-12.
- Dependency gate reviewed: PASS; `go run ./tools/depgate` proves zero external imports for library, test, example, and tool packages in the root module.
- Any fixture-local dependency exceptions: Fixture-local external modules are isolated under `tools/depgate/testdata/` as intentional negative fixtures for `tools/depgate/main_test.go`.

## Waivers

Waivers require owner, reason, expiry, and impact tracking. Open-ended waivers block release.

| Item | Owner | Reason | Expiry | Impact |
| --- | --- | --- | --- | --- |
| None | Coto | No waivers requested; all required local gates were run. | Not applicable | No release impact. |

## Final Review

- All required evidence captured: Yes; exact commit, test, vet, dependency-gate, race-test, docs/examples, fuzz, provenance, compatibility, migration, CI runner/action, Go version, and dependency evidence are recorded above.
- All waivers approved with expiry: No waivers requested; any future waiver must include owner, reason, expiry, and impact before release review continues.
- Tagging decision: Evidence is captured for human release review of `v0.1.0`; this checklist does not approve or perform the tag action.
